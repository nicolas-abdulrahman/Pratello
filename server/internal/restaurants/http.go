package restaurants

import (
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"pratello/server/internal/auth"
)

type Handler struct {
	svc  service
	auth *auth.Service
}

func NewHandler(db *sql.DB, identity *auth.Service) *Handler {
	return &Handler{svc: service{postgres{db}}, auth: identity}
}

type route struct {
	method, path string
	op           operation
	status       int
	auth         bool
}

func (h *Handler) Routes(mux *http.ServeMux) {
	const org = "/restaurants/{slug}"
	const branches = org + "/branches"
	const branch = branches + "/{branch_id}"
	for _, p := range []string{"", "/api"} {
		for _, rt := range []route{
			{"GET", "/restaurants", listOrg, 200, false}, {"POST", "/restaurants", createOrg, 201, true}, {"GET", "/me/restaurants", myOrg, 200, true},
			{"GET", org, getOrg, 200, false}, {"PUT", org, updateOrg, 200, true}, {"DELETE", org, archiveOrg, 204, true},
			{"GET", branches, listBranch, 200, false}, {"POST", branches, createBranch, 201, true},
			{"GET", branch, getBranch, 200, false}, {"PUT", branch, updateBranch, 200, true}, {"DELETE", branch, archiveBranch, 204, true},
			{"PUT", branch + "/verification", verifyBranch, 200, true}, {"PUT", branch + "/address", setAddress, 200, true},
			{"GET", branch + "/phones", listPhones, 200, true}, {"POST", branch + "/phones", addPhone, 201, true}, {"DELETE", branch + "/phones/{target}", removePhone, 204, true},
			{"GET", branch + "/hours", getHours, 200, false}, {"PUT", branch + "/hours", setHours, 200, true},
			{"GET", branch + "/tables", listTables, 200, true}, {"POST", branch + "/tables", addTable, 201, true}, {"PUT", branch + "/tables/{target}", updateTable, 200, true}, {"DELETE", branch + "/tables/{target}", removeTable, 204, true},
		} {
			mux.HandleFunc(rt.method+" "+p+rt.path, func(w http.ResponseWriter, r *http.Request) { h.handle(w, r, rt) })
		}
	}
}

func (h *Handler) actor(r *http.Request) (int, error) {
	p := strings.Fields(r.Header.Get("Authorization"))
	if len(p) != 2 || !strings.EqualFold(p[0], "Bearer") || h.auth == nil {
		return 0, auth.ErrUnauthorized
	}
	a, err := h.auth.Authenticate(r.Context(), p[1])
	if err != nil {
		return 0, err
	}
	return a.AccountID, nil
}

func positive(s string) (int, error) {
	n, err := strconv.ParseInt(s, 10, 32)
	if err != nil || n < 1 {
		return 0, errInvalid
	}
	return int(n), nil
}

func parseRequest(w http.ResponseWriter, r *http.Request, op operation) (request, error) {
	x := request{op: op, slug: r.PathValue("slug"), page: 1, size: 20}
	var err error
	if v := r.PathValue("branch_id"); v != "" {
		x.branch, err = positive(v)
		if err != nil {
			return x, err
		}
	}
	if v := r.PathValue("target"); v != "" {
		x.target, err = positive(v)
		if err != nil {
			return x, err
		}
	}
	if op == listOrg || op == listBranch {
		if v := r.URL.Query().Get("page"); v != "" {
			x.page, err = positive(v)
			if err != nil {
				return x, err
			}
		}
		if v := r.URL.Query().Get("page_size"); v != "" {
			x.size, err = positive(v)
			if err != nil || x.size > 100 {
				return x, errInvalid
			}
		}
	}
	switch op {
	case createOrg:
		x.body = new(newOrg)
	case updateOrg, updateBranch:
		x.body = new(details)
	case createBranch:
		x.body = new(branchInput)
	case setAddress:
		x.body = new(address)
	case addPhone:
		x.body = new(phoneInput)
	case setHours:
		x.body = new([]hour)
	case addTable, updateTable:
		x.body = new(tableInput)
	}
	if x.body != nil {
		d := json.NewDecoder(http.MaxBytesReader(w, r.Body, 32<<10))
		d.DisallowUnknownFields()
		value := x.body
		if err = d.Decode(&value); err != nil {
			return x, decodeError(err)
		}
		if value == nil {
			return x, errInvalid
		}
		if err = d.Decode(new(any)); err != io.EOF {
			return x, decodeError(err)
		}
	}
	return x, nil
}

func decodeError(err error) error {
	var limit *http.MaxBytesError
	if errors.As(err, &limit) {
		return err
	}
	return errInvalid
}

func respond(w http.ResponseWriter, status int, body any) {
	if status != 204 {
		w.Header().Set("Content-Type", "application/json")
	}
	w.WriteHeader(status)
	if status != 204 {
		_ = json.NewEncoder(w).Encode(body)
	}
}

func httpError(w http.ResponseWriter, err error) {
	status, message := 500, "internal server error"
	var limit *http.MaxBytesError
	switch {
	case errors.As(err, &limit):
		status, message = 413, "request body too large"
	case errors.Is(err, auth.ErrUnauthorized):
		status, message = 401, "authentication required"
		w.Header().Set("WWW-Authenticate", "Bearer")
	case errors.Is(err, errInvalid):
		status, message = 400, errInvalid.Error()
	case errors.Is(err, errForbidden):
		status, message = 403, errForbidden.Error()
	case errors.Is(err, errMissing):
		status, message = 404, errMissing.Error()
	case errors.Is(err, errConflict), errors.Is(err, errLastOwner):
		status, message = 409, err.Error()
	case errors.Is(err, errUnavailable):
		status, message = 503, errUnavailable.Error()
	}
	if status == 500 {
		slog.Error("restaurant request failed", "error", err)
	}
	respond(w, status, map[string]string{"error": message})
}

func (h *Handler) handle(w http.ResponseWriter, r *http.Request, rt route) {
	var actor int
	if rt.auth || ((rt.op == getBranch || rt.op == getHours || rt.op == listBranch) && r.Header.Get("Authorization") != "") {
		var err error
		actor, err = h.actor(r)
		if err != nil {
			httpError(w, err)
			return
		}
	}
	x, err := parseRequest(w, r, rt.op)
	if err != nil {
		httpError(w, err)
		return
	}
	x.actor = actor
	result, err := h.svc.run(r.Context(), x)
	if err != nil {
		httpError(w, err)
		return
	}
	respond(w, rt.status, result)
}
