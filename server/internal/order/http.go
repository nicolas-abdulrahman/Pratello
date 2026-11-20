package order

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"html/template"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	qrcode "github.com/skip2/go-qrcode"
	"pratello/server/internal/menu"
)

type Handler struct {
	svc      *Service
	identify func(context.Context, string) (int, error)
	base     string
}

func NewHandler(s *Service, id func(context.Context, string) (int, error), publicBase string) *Handler {
	return &Handler{s, id, strings.TrimRight(publicBase, "/")}
}

type route struct {
	method, path, op string
	status           int
	staff            bool
}

func (h *Handler) Routes(mux *http.ServeMux) {
	const branch = "/restaurants/{slug}/branches/{branch_id}"
	const order = "/orders/{order_id}"
	for _, prefix := range []string{"", "/api"} {
		for _, rt := range []route{
			{"GET", branch + "/orders", "list", 200, true}, {"POST", branch + "/orders/local", "local", 201, false}, {"POST", branch + "/orders/remote", "remote", 201, false},
			{"GET", order, "get", 200, false}, {"POST", order + "/items", "add", 201, false}, {"DELETE", order + "/items/{item_id}", "void", 204, false},
			{"POST", order + "/submit", "submit", 200, false}, {"PATCH", order + "/status", "status", 200, true},
			{"PUT", branch + "/delivery-fee", "fee", 200, true}, {"POST", branch + "/order-qr", "qr", 201, true},
			{"POST", "/order-qr/{code}/open", "scan", 201, false},
		} {
			mux.HandleFunc(rt.method+" "+prefix+rt.path, func(w http.ResponseWriter, r *http.Request) { h.handle(w, r, rt, prefix) })
		}
		mux.HandleFunc("GET "+prefix+"/order-qr/{code}", landing)
	}
}

func (h *Handler) handle(w http.ResponseWriter, r *http.Request, rt route, prefix string) {
	w.Header().Set("Cache-Control", "no-store")
	x := command{op: rt.op, slug: r.PathValue("slug"), code: r.PathValue("code"), guest: r.Header.Get("X-Order-Token"), page: 1, size: 20, device: device(r.UserAgent())}
	if rt.staff || r.Header.Get("Authorization") != "" {
		parts := strings.Fields(r.Header.Get("Authorization"))
		if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") && h.identify != nil {
			v, e := h.identify(r.Context(), parts[1])
			if e == nil {
				x.actor = v
			}
		}
		if x.actor <= 0 {
			w.Header().Set("WWW-Authenticate", "Bearer")
			respond(w, 401, map[string]string{"error": "authentication required"})
			return
		}
	}
	base := ""
	err := parse(w, r, &x)
	if err == nil && rt.op == "qr" {
		base, err = h.origin(r)
	}
	if err == nil {
		v, e := h.svc.execute(r.Context(), x)
		err = e
		if err == nil {
			switch p := v.(type) {
			case opened:
				if p.Guest != "" {
					w.Header().Set("X-Order-Token", p.Guest)
				}
				v = p.Order
			case qrLink:
				link := base + prefix + "/order-qr/" + p.Token
				png, e := qrcode.Encode(link, qrcode.Medium, 256)
				if e != nil {
					writeError(w, r, e)
					return
				}
				v = struct {
					qrLink
					URL string `json:"url"`
					PNG string `json:"qr_code_base64"`
				}{p, link, base64.StdEncoding.EncodeToString(png)}
			}
			respond(w, rt.status, v)
			return
		}
	}
	writeError(w, r, err)
}
func (h *Handler) origin(r *http.Request) (string, error) {
	base := h.base
	if base == "" {
		scheme := "http"
		if r.TLS != nil {
			scheme = "https"
		}
		base = scheme + "://" + r.Host
	}
	u, err := url.Parse(base)
	if err != nil || len(base) > 500 || u.Hostname() == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" || u.Path != "" || u.Scheme != "https" && u.Scheme != "http" {
		return "", errInvalid
	}
	return u.String(), nil
}
func device(ua string) string {
	ua = strings.ToLower(ua)
	if strings.Contains(ua, "ipad") || strings.Contains(ua, "tablet") {
		return "tablet"
	}
	if strings.Contains(ua, "mobile") || strings.Contains(ua, "android") {
		return "mobile"
	}
	return "desktop"
}
func parse(w http.ResponseWriter, r *http.Request, x *command) error {
	q, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		return errInvalid
	}
	for _, p := range []struct {
		value string
		dest  *int
	}{{r.PathValue("branch_id"), &x.branch}, {r.PathValue("order_id"), &x.id}, {r.PathValue("item_id"), &x.item}, {q.Get("page"), &x.page}, {q.Get("page_size"), &x.size}} {
		if p.value == "" {
			continue
		}
		n, e := strconv.ParseInt(p.value, 10, 32)
		if e != nil || n < 1 {
			return errInvalid
		}
		*p.dest = int(n)
	}
	if x.size > 100 {
		return errInvalid
	}
	if x.op == "scan" && !validCode(x.code) {
		return errMissing
	}
	var dest any
	optional := false
	switch x.op {
	case "local", "remote":
		dest = &x.create
	case "qr":
		dest = &x.qr
		optional = true
	case "fee":
		dest = &x.fee
	case "submit":
		dest = &x.submit
		optional = true
	case "add":
		dest = &x.selection
	case "status":
		dest = &x.status
	case "list":
		x.status.Status = q.Get("status")
	case "void":
		x.reason = q.Get("reason")
	}
	if dest != nil {
		d := json.NewDecoder(http.MaxBytesReader(w, r.Body, 128<<10))
		d.DisallowUnknownFields()
		if err := d.Decode(&dest); err != nil {
			if err == io.EOF && optional {
				return nil
			}
			return bodyError(err)
		}
		if dest == nil {
			return errInvalid
		}
		if err := d.Decode(new(any)); err != io.EOF {
			return bodyError(err)
		}
	}
	if x.op == "status" || x.op == "list" && x.status.Status != "" {
		switch x.status.Status {
		case "pending", "preparing", "ready", "delivered", "cancelled":
		default:
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
func validCode(code string) bool {
	raw, err := base64.RawURLEncoding.Strict().DecodeString(code)
	return err == nil && len(raw) == 32
}
func respond(w http.ResponseWriter, status int, v any) {
	if v != nil {
		w.Header().Set("Content-Type", "application/json")
	}
	w.WriteHeader(status)
	if v != nil {
		_ = json.NewEncoder(w).Encode(v)
	}
}
func writeError(w http.ResponseWriter, r *http.Request, err error) {
	status, message := 500, err.Error()
	var tooBig *http.MaxBytesError
	switch {
	case errors.As(err, &tooBig):
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
		slog.ErrorContext(r.Context(), "order request failed", "error", err)
		message = "internal server error"
	}
	respond(w, status, map[string]string{"error": message})
}

var landingPage = template.Must(template.New("order").Parse(`<!doctype html><html lang="pt-BR"><meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1"><title>Pratello — Pedido</title><main><h1>Pratello</h1><p id="result" role="status">Abrindo seu pedido…</p><noscript>Ative JavaScript para abrir um pedido pelo QR.</noscript></main><script nonce="{{.}}">
(async () => {
  const result = document.getElementById('result');
  try {
    const res = await fetch(location.pathname + '/open', {method: 'POST'});
    if (!res.ok) throw new Error();
    const order = await res.json();
    const token = res.headers.get('X-Order-Token');
    if (token) try { sessionStorage.setItem('pratello.order.' + order.id, token); } catch (_) {}
    const table = order.local_info?.table_number;
    result.textContent = 'Pedido #' + order.id + ' aberto' + (table ? ' — Mesa ' + table : '') + '.';
  } catch (_) { result.textContent = 'Não foi possível abrir o pedido. Verifique o QR e tente novamente.'; }
})();</script></html>`))

func landing(w http.ResponseWriter, r *http.Request) {
	if !validCode(r.PathValue("code")) {
		http.NotFound(w, r)
		return
	}
	nonce, _, err := token()
	if err != nil {
		writeError(w, r, err)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Referrer-Policy", "no-referrer")
	w.Header().Set("Content-Security-Policy", "default-src 'none'; connect-src 'self'; script-src 'nonce-"+nonce+"'; base-uri 'none'; frame-ancestors 'none'")
	_ = landingPage.Execute(w, nonce)
}
