package sales

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"pratello/server/internal/menu"
)

type Handler struct {
	service  *Service
	identify func(context.Context, string) (int, error)
}

func NewHandler(s *Service, identify func(context.Context, string) (int, error)) *Handler {
	return &Handler{s, identify}
}
func (o *optional[T]) UnmarshalJSON(data []byte) error {
	o.set = true
	return json.Unmarshal(data, &o.value)
}

type route struct {
	method, path, op string
	private          bool
	status           int
}

func (h *Handler) Routes(mux *http.ServeMux) {
	const org = "/restaurants/{slug}"
	for _, prefix := range []string{"", "/api"} {
		for _, r := range []route{
			{"GET", org + "/coupons", "list", true, 200}, {"POST", org + "/coupons", "create", true, 201},
			{"GET", org + "/coupons/{code}", "validate", false, 200}, {"PUT", org + "/coupons/{code}", "update", true, 200}, {"DELETE", org + "/coupons/{code}", "deactivate", true, 204},
			{"GET", org + "/menu/banners", "banners", false, 200}, {"POST", org + "/menu/banners/{id}/claim", "claim", false, 200},
			{"GET", "/me" + org + "/coupons", "mine", true, 200}, {"POST", org + "/branches/{branch_id}/checkout/quote", "quote", false, 200},
		} {
			mux.HandleFunc(r.method+" "+prefix+r.path, func(w http.ResponseWriter, req *http.Request) { h.handle(w, req, r) })
		}
	}
}
func (h *Handler) handle(w http.ResponseWriter, r *http.Request, rt route) {
	w.Header().Set("Cache-Control", "no-store")
	x := request{op: rt.op, slug: r.PathValue("slug"), code: r.PathValue("code"), page: 1, size: 20}
	if rt.private || r.Header.Get("Authorization") != "" {
		p := strings.Fields(r.Header.Get("Authorization"))
		if len(p) == 2 && strings.EqualFold(p[0], "Bearer") && h.identify != nil {
			id, e := h.identify(r.Context(), p[1])
			if e == nil {
				x.actor = id
			}
		}
		if x.actor <= 0 {
			w.Header().Set("WWW-Authenticate", "Bearer")
			write(w, 401, map[string]string{"error": "authentication required"})
			return
		}
	}
	err := parse(w, r, &x)
	if err == nil {
		var v any
		v, err = h.service.execute(r.Context(), x)
		if err == nil {
			write(w, rt.status, v)
			return
		}
	}
	status := 500
	message := err.Error()
	var limit *http.MaxBytesError
	switch {
	case errors.As(err, &limit):
		status = 413
	case errors.Is(err, errInvalid), errors.Is(err, menu.ErrInvalid):
		status = 400
	case errors.Is(err, errMissing):
		status = 404
	case errors.Is(err, errForbidden):
		status = 403
	case errors.Is(err, errConflict), errors.Is(err, menu.ErrNotOrderable):
		status = 409
	case errors.Is(err, errUnavailable):
		status = 503
	default:
		slog.ErrorContext(r.Context(), "sales request failed", "error", err)
		message = "internal server error"
	}
	write(w, status, map[string]string{"error": message})
}
func parse(w http.ResponseWriter, r *http.Request, x *request) error {
	q := r.URL.Query()
	for _, p := range []struct {
		key, value string
		dest       *int
	}{{"branch_id", r.PathValue("branch_id"), &x.branch}, {"id", r.PathValue("id"), &x.id}, {"page", q.Get("page"), &x.page}, {"page_size", q.Get("page_size"), &x.size}} {
		v := p.value
		if p.key == "branch_id" && v == "" {
			v = q.Get(p.key)
		}
		if v == "" {
			continue
		}
		n, e := strconv.ParseInt(v, 10, 32)
		if e != nil || n < 1 {
			return errInvalid
		}
		*p.dest = int(n)
	}
	if x.size > 100 {
		return errInvalid
	}
	if x.code != "" && !codePattern.MatchString(normalize(x.code)) {
		return errInvalid
	}
	if x.op == "validate" {
		v, e := strconv.ParseInt(q.Get("subtotal_cents"), 10, 64)
		if e != nil || v < 0 {
			return errInvalid
		}
		x.subtotal = v
	}
	var dest any
	switch x.op {
	case "create", "update":
		dest = &x.edit
	case "quote":
		dest = &x.cart
	}
	if dest != nil {
		d := json.NewDecoder(http.MaxBytesReader(w, r.Body, 128<<10))
		d.DisallowUnknownFields()
		var obj = json.RawMessage{}
		if err := d.Decode(&obj); err != nil {
			return bodyError(err)
		}
		if len(obj) == 0 || obj[0] != '{' {
			return errInvalid
		}
		if err := d.Decode(new(any)); err != io.EOF {
			return bodyError(err)
		}
		// Decode the single object strictly, retaining omitted/null distinctions.
		strict := json.NewDecoder(strings.NewReader(string(obj)))
		strict.DisallowUnknownFields()
		if err := strict.Decode(dest); err != nil {
			return errInvalid
		}
	}
	return nil
}
func bodyError(err error) error {
	var limit *http.MaxBytesError
	if errors.As(err, &limit) {
		return err
	}
	return errInvalid
}
func write(w http.ResponseWriter, status int, v any) {
	if v != nil {
		w.Header().Set("Content-Type", "application/json")
	}
	w.WriteHeader(status)
	if v != nil {
		_ = json.NewEncoder(w).Encode(v)
	}
}
