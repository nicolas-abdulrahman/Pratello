package review

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type Handler struct {
	service  *Service
	identify func(context.Context, string) (int, error)
}

func NewHandler(s *Service, identify func(context.Context, string) (int, error)) *Handler {
	return &Handler{s, identify}
}

func (h *Handler) Routes(mux *http.ServeMux) {
	for _, p := range []string{"", "/api"} {
		for _, r := range []struct {
			method, path string
			item, write  bool
		}{
			{"GET", "/restaurants/{slug}/items/{item_id}/reviews", true, false},
			{"GET", "/restaurants/{slug}/reviews", false, false},
			{"POST", "/orders/{order_id}/items/{order_item_id}/review", true, true},
			{"POST", "/orders/{order_id}/review", false, true},
		} {
			mux.HandleFunc(r.method+" "+p+r.path, func(w http.ResponseWriter, req *http.Request) { h.handle(w, req, r.item, r.write) })
		}
	}
}

func (h *Handler) handle(w http.ResponseWriter, r *http.Request, item, write bool) {
	w.Header().Set("Cache-Control", "no-store")
	p := request{item: item, slug: r.PathValue("slug"), page: 1, size: 20}
	if write {
		p.guest = r.Header.Get("X-Order-Token")
		a := r.Header.Get("Authorization")
		if a != "" {
			parts := strings.Fields(a)
			if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") && h.identify != nil {
				if id, err := h.identify(r.Context(), parts[1]); err == nil {
					p.actor = id
				}
			}
		}
		if p.actor <= 0 && (a != "" || p.guest == "") {
			w.Header().Set("WWW-Authenticate", "Bearer")
			respond(w, 401, map[string]string{"error": "purchase credentials required"})
			return
		}
	}
	err := parse(w, r, &p, write)
	if err == nil {
		var body any
		if write {
			body, err = h.service.write(r.Context(), p)
		} else {
			body, err = h.service.list(r.Context(), p)
		}
		if err == nil {
			respond(w, 200, body)
			return
		}
	}
	status, msg := 500, err.Error()
	var tooLarge *http.MaxBytesError
	switch {
	case errors.As(err, &tooLarge):
		status = 413
	case errors.Is(err, errInvalid):
		status = 400
	case errors.Is(err, errForbidden):
		status = 403
	case errors.Is(err, errMissing):
		status = 404
	case errors.Is(err, errConflict):
		status = 409
	case errors.Is(err, errUnavailable):
		status = 503
	default:
		slog.ErrorContext(r.Context(), "review request failed", "error", err)
		msg = "internal server error"
	}
	respond(w, status, map[string]string{"error": msg})
}

func parse(w http.ResponseWriter, r *http.Request, p *request, write bool) error {
	q, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		return errInvalid
	}
	for _, x := range []struct {
		value string
		dest  *int
	}{{r.PathValue("item_id"), &p.target}, {r.PathValue("order_id"), &p.order}, {r.PathValue("order_item_id"), &p.line}, {q.Get("page"), &p.page}, {q.Get("page_size"), &p.size}} {
		if x.value == "" {
			continue
		}
		n, e := strconv.ParseInt(x.value, 10, 32)
		if e != nil || n < 1 {
			return errInvalid
		}
		*x.dest = int(n)
	}
	if p.size > 100 {
		return errInvalid
	}
	if !write {
		return nil
	}
	d := json.NewDecoder(http.MaxBytesReader(w, r.Body, 16<<10))
	d.DisallowUnknownFields()
	if err := d.Decode(&p.input); err != nil {
		return bodyError(err)
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
func respond(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
