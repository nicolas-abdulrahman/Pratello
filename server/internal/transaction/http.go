package transaction

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
)

type Handler struct {
	svc      *Service
	identify func(context.Context, string) (int, error)
}

func NewHandler(s *Service, id func(context.Context, string) (int, error)) *Handler {
	return &Handler{s, id}
}

type route struct {
	method, path, op string
	status           int
}

func (h *Handler) Routes(mux *http.ServeMux) {
	const orders = "/orders/{order_id}/transactions"
	const payment = "/transactions/{id}"
	for _, prefix := range []string{"", "/api"} {
		for _, rt := range []route{
			{"GET", orders, "list", 200}, {"POST", orders, "start", 201}, {"GET", payment, "get", 200},
			{"GET", payment + "/pix", "pixGet", 200}, {"POST", payment + "/pix", "pixCreate", 201},
			{"POST", payment + "/pix/confirm", "confirm", 200}, {"POST", payment + "/refund", "refund", 200},
		} {
			mux.HandleFunc(rt.method+" "+prefix+rt.path, func(w http.ResponseWriter, r *http.Request) { h.handle(w, r, rt) })
		}
	}
}
func (h *Handler) handle(w http.ResponseWriter, r *http.Request, rt route) {
	w.Header().Set("Cache-Control", "no-store")
	x := request{op: rt.op, key: r.Header.Get("Idempotency-Key"), guest: r.Header.Get("X-Order-Token")}
	p := strings.Fields(r.Header.Get("Authorization"))
	if len(p) == 2 && strings.EqualFold(p[0], "Bearer") && h.identify != nil {
		id, err := h.identify(r.Context(), p[1])
		if err == nil {
			x.actor = id
		}
	}
	if x.actor <= 0 && (x.guest == "" || r.Header.Get("Authorization") != "") {
		w.Header().Set("WWW-Authenticate", "Bearer")
		respond(w, 401, map[string]string{"error": "authentication required"})
		return
	}
	err := parse(w, r, &x)
	if err == nil {
		var response any
		response, err = h.svc.execute(r.Context(), x)
		if err == nil {
			respond(w, rt.status, response)
			return
		}
	}
	status, message := 500, err.Error()
	var size *http.MaxBytesError
	switch {
	case errors.As(err, &size):
		status = 413
	case errors.Is(err, errInvalid):
		status = 400
	case errors.Is(err, errMissing):
		status = 404
	case errors.Is(err, errForbidden):
		status = 403
	case errors.Is(err, errConflict):
		status = 409
	case errors.Is(err, errUnavailable):
		status = 503
	default:
		slog.ErrorContext(r.Context(), "payment request failed", "error", err)
		message = "internal server error"
	}
	respond(w, status, map[string]string{"error": message})
}
func parse(w http.ResponseWriter, r *http.Request, x *request) error {
	for _, p := range []struct {
		name string
		dest *int
	}{{"order_id", &x.order}, {"id", &x.id}} {
		v := r.PathValue(p.name)
		if v == "" {
			continue
		}
		n, err := strconv.ParseInt(v, 10, 32)
		if err != nil || n < 1 {
			return errInvalid
		}
		*p.dest = int(n)
	}
	if len(r.Header.Values("Idempotency-Key")) > 1 {
		return errInvalid
	}
	var dest any
	switch x.op {
	case "start":
		dest = &x.start
	case "refund":
		dest = &x.refund
	}
	if dest == nil {
		return nil
	}
	d := json.NewDecoder(http.MaxBytesReader(w, r.Body, 16<<10))
	d.DisallowUnknownFields()
	if err := d.Decode(&dest); err != nil {
		return bodyError(err)
	}
	if dest == nil {
		return errInvalid
	}
	if err := d.Decode(new(any)); err != io.EOF {
		return bodyError(err)
	}
	return nil
}
func bodyError(err error) error {
	var size *http.MaxBytesError
	if errors.As(err, &size) {
		return err
	}
	return errInvalid
}
func respond(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
