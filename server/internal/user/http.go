package user

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
)

// Identify resolves an access token to a live account. Invalid credentials
// return zero; infrastructure errors are propagated separately.
type Identify func(context.Context, string) (int, error)

type Handler struct {
	svc      service
	identify Identify
}

func NewHandler(db *sql.DB, identify Identify) *Handler {
	return &Handler{service{postgres{db}}, identify}
}

func (h *Handler) Routes(mux *http.ServeMux) {
	type endpoint struct {
		method, path, action string
		role                 Role
		status               int
		body                 bool
	}
	const org = "/restaurants/{slug}"
	const branch = org + "/branches/{branch_id}"
	for _, prefix := range []string{"", "/api"} {
		for _, rt := range []endpoint{
			{"GET", "/administrators", "list", Administrator, 200, false}, {"POST", "/administrators", "grant", Administrator, 201, true}, {"DELETE", "/administrators/{target}", "revoke", Administrator, 204, false},
			{"GET", "/sysadmins", "list", Sysadmin, 200, false}, {"POST", "/sysadmins", "grant", Sysadmin, 201, true}, {"DELETE", "/sysadmins/{target}", "revoke", Sysadmin, 204, false},
			{"GET", org + "/owners", "list", Owner, 200, false}, {"POST", org + "/owners", "grant", Owner, 201, true}, {"DELETE", org + "/owners/{target}", "revoke", Owner, 204, false},
			{"GET", org + "/analysts", "list", Analyst, 200, false}, {"POST", org + "/analysts", "grant", Analyst, 201, true}, {"DELETE", org + "/analysts/{target}", "revoke", Analyst, 204, false},
			{"GET", branch + "/managers", "list", Manager, 200, false}, {"PUT", branch + "/managers/{target}", "grant", Manager, 204, false}, {"DELETE", branch + "/managers/{target}", "revoke", Manager, 204, false},
			{"GET", branch + "/employees", "list", Employee, 200, false}, {"POST", branch + "/employees", "grant", Employee, 201, true}, {"DELETE", branch + "/employees/{target}", "revoke", Employee, 204, false},
		} {
			mux.HandleFunc(rt.method+" "+prefix+rt.path, func(w http.ResponseWriter, r *http.Request) {
				c := change{slug: r.PathValue("slug"), role: rt.role, action: rt.action}
				parts := strings.Fields(r.Header.Get("Authorization"))
				if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || h.identify == nil {
					writeError(w, 401, "authentication required")
					return
				}
				var err error
				c.actor, err = h.identify(r.Context(), parts[1])
				if err != nil {
					writeError(w, 500, "internal server error")
					return
				}
				if c.actor == 0 {
					writeError(w, 401, "authentication required")
					return
				}
				for name, dst := range map[string]*int{"branch_id": &c.branch, "target": &c.target} {
					if raw := r.PathValue(name); raw != "" {
						id, e := strconv.ParseInt(raw, 10, 32)
						if e != nil || id <= 0 {
							writeError(w, 400, ErrInvalid.Error())
							return
						}
						*dst = int(id)
					}
				}
				if rt.body {
					var body struct {
						ID       int    `json:"account_id"`
						Position string `json:"position"`
					}
					d := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096))
					d.DisallowUnknownFields()
					if d.Decode(&body) != nil || d.Decode(new(any)) != io.EOF || body.ID <= 0 || body.ID > 2147483647 || (c.role != Employee && body.Position != "") {
						writeError(w, 400, ErrInvalid.Error())
						return
					}
					c.target, c.position = body.ID, body.Position
				}
				result, err := h.svc.run(r.Context(), c)
				if err != nil {
					roleError(w, err)
					return
				}
				write(w, rt.status, result)
			})
		}
	}
}

func write(w http.ResponseWriter, status int, value any) {
	if status != 204 {
		w.Header().Set("Content-Type", "application/json")
	}
	w.WriteHeader(status)
	if status != 204 {
		_ = json.NewEncoder(w).Encode(value)
	}
}
func writeError(w http.ResponseWriter, status int, msg string) {
	write(w, status, map[string]string{"error": msg})
}
func roleError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrInvalid):
		writeError(w, 400, err.Error())
	case errors.Is(err, ErrForbidden):
		writeError(w, 403, err.Error())
	case errors.Is(err, ErrMissing):
		writeError(w, 404, err.Error())
	case errors.Is(err, ErrConflict):
		writeError(w, 409, err.Error())
	case errors.Is(err, ErrUnavailable):
		writeError(w, 503, err.Error())
	default:
		writeError(w, 500, "internal server error")
	}
}

// WriteAccount renders the self-service profile without exposing credentials.
func WriteAccount(w http.ResponseWriter, a *Account) {
	var email, phone, name, cpf any
	if a.Email != "" {
		email = a.Email
	}
	if a.Phone != "" {
		phone = a.Phone
	}
	if a.Name != "" {
		name = a.Name
	}
	if a.CPF != "" {
		cpf = a.CPF
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"id": a.ID, "email": email, "phone": phone, "name": name, "cpf": cpf, "email_verified": a.EmailVerified, "phone_verified": a.PhoneVerified, "cpf_claimed_at": a.CPFClaimedAt, "is_shadow": a.Email == "" && a.Phone == "" && a.Password == "", "created_at": a.CreatedAt})
}
