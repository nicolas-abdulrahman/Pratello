package menu

import (
	"bytes"
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
	service interface {
		execute(context.Context, request) (any, error)
	}
	identify func(context.Context, string) (int, error)
}

func NewHandler(service *Service, identify func(context.Context, string) (int, error)) *Handler {
	return &Handler{service, identify}
}

type route struct {
	method, path string
	op           operation
	status       int
	body         func() any
}

func input[T any]() any { return new(T) }

func (h *Handler) Routes(mux *http.ServeMux) {
	const org = "/restaurants/{slug}"
	const items = org + "/items"
	const dish = items + "/{item_id}"
	const groups = dish + "/sides"
	const group = groups + "/{side_id}"
	const opts = group + "/options"
	const branchItem = org + "/branches/{branch_id}/items/{item_id}"
	routes := []route{
		{"GET", org + "/menu", getLayout, 200, nil}, {"PUT", org + "/menu", putLayout, 200, input[layoutInput]},
		{"GET", org + "/categories", listCategories, 200, nil}, {"POST", org + "/categories", addCategory, 201, input[categoryInput]},
		{"PUT", org + "/categories/{id}", putCategory, 200, input[categoryInput]}, {"DELETE", org + "/categories/{id}", dropCategory, 204, nil},
		{"GET", items, listItems, 200, nil}, {"POST", items, addItem, 201, input[itemInput]},
		{"GET", dish, getItem, 200, nil}, {"PUT", dish, putItem, 200, input[itemInput]},
		{"PATCH", dish, patchItem, 200, input[itemInput]}, {"DELETE", dish, archiveItem, 204, nil},
		{"PUT", branchItem + "/availability", stockItem, 200, input[stockInput]},
		{"GET", org + "/ingredients", listIngredients, 200, nil}, {"POST", org + "/ingredients", addIngredient, 201, input[ingredientInput]},
		{"PUT", org + "/ingredients/{id}", putIngredient, 200, input[ingredientInput]}, {"DELETE", org + "/ingredients/{id}", dropIngredient, 204, nil},
		{"GET", dish + "/ingredients", listRecipes, 200, nil}, {"POST", dish + "/ingredients", addRecipe, 201, input[recipeInput]},
		{"PUT", dish + "/ingredients/{id}", putRecipe, 200, input[recipeInput]}, {"DELETE", dish + "/ingredients/{id}", dropRecipe, 204, nil},
		{"GET", groups, listSides, 200, nil}, {"POST", groups, addSide, 201, input[sideInput]},
		{"PUT", groups + "/{id}", putSide, 200, input[sideInput]}, {"DELETE", groups + "/{id}", dropSide, 204, nil},
		{"POST", opts, addOption, 201, input[optionInput]}, {"PUT", opts + "/{id}", putOption, 200, input[optionInput]}, {"DELETE", opts + "/{id}", dropOption, 204, nil},
		{"PUT", branchItem + "/sides/{side_id}/options/{id}/availability", stockOption, 200, input[stockInput]},
		{"GET", dish + "/combo-items", listComponents, 200, nil}, {"POST", dish + "/combo-items", addComponent, 201, input[componentInput]},
		{"PUT", dish + "/combo-items/{id}", putComponent, 200, input[componentInput]}, {"DELETE", dish + "/combo-items/{id}", dropComponent, 204, nil},
		{"GET", dish + "/sales", listSales, 200, nil}, {"POST", dish + "/sales", addSale, 201, input[saleInput]}, {"DELETE", dish + "/sales/{id}", dropSale, 204, nil},
	}
	for _, prefix := range []string{"", "/api"} {
		for _, rt := range routes {
			mux.HandleFunc(rt.method+" "+prefix+rt.path, func(w http.ResponseWriter, r *http.Request) { h.handle(w, r, rt) })
		}
	}
}

const bodyLimit = 64 << 10

var errTooLarge = errors.New("menu request body too large")

func (n *nullable[T]) UnmarshalJSON(raw []byte) error {
	n.Set = true
	return json.Unmarshal(raw, &n.Value)
}

func decode(w http.ResponseWriter, r *http.Request, dst any) error {
	data, err := io.ReadAll(http.MaxBytesReader(w, r.Body, bodyLimit))
	if err != nil {
		var size *http.MaxBytesError
		if errors.As(err, &size) {
			return errTooLarge
		}
		return ErrInvalid
	}
	data = bytes.TrimSpace(data)
	if len(data) == 0 || data[0] != '{' || !json.Valid(data) {
		return ErrInvalid
	}
	d := json.NewDecoder(bytes.NewReader(data))
	d.DisallowUnknownFields()
	if d.Decode(dst) != nil {
		return ErrInvalid
	}
	return nil
}

func parseID(value string) (int, error) {
	v, err := strconv.ParseInt(value, 10, 32)
	if err != nil || !validID(int(v)) {
		return 0, ErrInvalid
	}
	return int(v), nil
}

func parseQueryID(q url.Values, key string, dst *int) error {
	xs, ok := q[key]
	if !ok {
		return nil
	}
	if len(xs) != 1 {
		return ErrInvalid
	}
	v, err := parseID(xs[0])
	if err == nil {
		*dst = v
	}
	return err
}

func parse(w http.ResponseWriter, r *http.Request, rt route) (request, error) {
	x := request{op: rt.op, slug: r.PathValue("slug"), write: rt.method != "GET", page: 1, size: 20}
	for _, p := range []struct {
		key  string
		dest *int
	}{{"item_id", &x.item}, {"branch_id", &x.branch}, {"side_id", &x.side}, {"id", &x.id}} {
		if value := r.PathValue(p.key); value != "" {
			n, err := parseID(value)
			if err != nil {
				return x, err
			}
			*p.dest = n
		}
	}
	q, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		return x, ErrInvalid
	}
	if !x.write || rt.op == dropSale {
		if err := parseQueryID(q, "branch_id", &x.branch); err != nil {
			return x, err
		}
	}
	if rt.op == listItems {
		if parseQueryID(q, "page", &x.page) != nil || parseQueryID(q, "page_size", &x.size) != nil || parseQueryID(q, "category_id", &x.category) != nil || x.size > 100 {
			return x, ErrInvalid
		}
	}
	if rt.op == dropCategory {
		if force, ok := q["force"]; ok {
			if len(force) != 1 || (force[0] != "true" && force[0] != "false") {
				return x, ErrInvalid
			}
			x.force = force[0] == "true"
		}
	}
	if rt.body != nil {
		x.body = rt.body()
		if err := decode(w, r, x.body); err != nil {
			return x, err
		}
	}
	return x, nil
}

func (h *Handler) handle(w http.ResponseWriter, r *http.Request, rt route) {
	w.Header().Set("Cache-Control", "no-store")
	actor := 0
	if rt.method != "GET" || len(r.Header.Values("Authorization")) > 0 {
		parts := strings.Fields(r.Header.Get("Authorization"))
		if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") && h.identify != nil {
			id, err := h.identify(r.Context(), parts[1])
			if err == nil {
				actor = id
			}
		}
		if actor <= 0 {
			w.Header().Set("WWW-Authenticate", "Bearer")
			writeError(w, 401, "unauthorized", "Authentication required")
			return
		}
	}
	x, err := parse(w, r, rt)
	x.actor = actor
	var result any
	if err == nil {
		result, err = h.service.execute(r.Context(), x)
	}
	if err != nil {
		switch {
		case errors.Is(err, errTooLarge):
			writeError(w, 413, "payload_too_large", err.Error())
		case errors.Is(err, ErrInvalid):
			writeError(w, 400, "validation_error", err.Error())
		case errors.Is(err, ErrForbidden):
			writeError(w, 403, "forbidden", err.Error())
		case errors.Is(err, ErrMissing):
			writeError(w, 404, "not_found", err.Error())
		case errors.Is(err, ErrConflict):
			writeError(w, 409, "conflict", err.Error())
		case errors.Is(err, ErrUnavailable):
			writeError(w, 503, "unavailable", err.Error())
		default:
			slog.ErrorContext(r.Context(), "menu request failed", "operation", rt.op, "error", err)
			writeError(w, 500, "internal_error", "Internal server error")
		}
		return
	}
	if result == nil {
		w.WriteHeader(rt.status)
		return
	}
	writeJSON(w, rt.status, result)
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, map[string]string{"code": code, "error": message})
}

func writeJSON(w http.ResponseWriter, status int, result any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(result)
}
