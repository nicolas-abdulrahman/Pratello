package restaurants

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"pratello/server/internal/auth"
)

type Handler struct {
	svc     service
	auth    *auth.Service
	catalog Service
}

func NewHandler(db *sql.DB, identity *auth.Service) *Handler {
	return &Handler{svc: service{postgres{db}}, auth: identity, catalog: Service{store: &sqlStore{db}}}
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
			{"PUT", branch + "/managers/{target}", assignManager, 204, true}, {"DELETE", branch + "/managers/{target}", revokeManager, 204, true},
			{"GET", branch + "/phones", listPhones, 200, true}, {"POST", branch + "/phones", addPhone, 201, true}, {"DELETE", branch + "/phones/{target}", removePhone, 204, true},
			{"GET", branch + "/hours", getHours, 200, false}, {"PUT", branch + "/hours", setHours, 200, true},
			{"GET", branch + "/tables", listTables, 200, true}, {"POST", branch + "/tables", addTable, 201, true}, {"PUT", branch + "/tables/{target}", updateTable, 200, true}, {"DELETE", branch + "/tables/{target}", removeTable, 204, true},
			{"GET", org + "/owners", listOwners, 200, true}, {"POST", org + "/owners", addOwner, 201, true}, {"DELETE", org + "/owners/{target}", removeOwner, 204, true},
			{"GET", org + "/managers", listOrgManagers, 200, true}, {"POST", org + "/managers", addOrgManager, 201, true}, {"DELETE", org + "/managers/{target}", removeOrgManager, 204, true},
		} {
			mux.HandleFunc(rt.method+" "+p+rt.path, func(w http.ResponseWriter, r *http.Request) { h.handle(w, r, rt) })
		}
	}
	h.catalogRoutes(mux)
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
	case addOwner, addOrgManager:
		x.body = new(accountInput)
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

// guardCatalog requires a published branch for reads and organization authority for writes.
func (h *Handler) guardCatalog(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		x := request{slug: r.PathValue("slug")}
		write := r.Method != http.MethodGet && r.Method != http.MethodHead
		if write {
			var err error
			x.actor, err = h.actor(r)
			if err != nil {
				httpError(w, err)
				return
			}
		}
		if x.slug != "" {
			err := h.svc.repo.transact(r.Context(), false, func(tx transaction) error {
				a, err := tx.scope(r.Context(), x, false)
				if err != nil {
					return err
				}
				if write {
					if !a.managesOrg() {
						return errForbidden
					}
					return nil
				}
				_, err = tx.execute(r.Context(), request{op: publishedCatalog}, a)
				return err
			})
			if err != nil {
				httpError(w, err)
				return
			}
		}
		next(w, r)
	}
}

func (h *Handler) catalogRoutes(mux *http.ServeMux) {
	guard := h.guardCatalog
	for _, p := range []string{"", "/api"} {
		// Dishes
		mux.HandleFunc(fmt.Sprintf("GET %s/restaurants/{slug}/dishes", p), guard(h.listDishes))
		mux.HandleFunc(fmt.Sprintf("POST %s/restaurants/{slug}/dishes", p), guard(h.createDish))
		mux.HandleFunc(fmt.Sprintf("PUT %s/restaurants/{slug}/dishes/{id}", p), guard(h.updateDish))
		mux.HandleFunc(fmt.Sprintf("DELETE %s/restaurants/{slug}/dishes/{id}", p), guard(h.deleteDish))

		// Menu
		mux.HandleFunc(fmt.Sprintf("GET %s/restaurants/{slug}/menu", p), guard(h.restaurantMenu))

		// Image uploads
		mux.HandleFunc(fmt.Sprintf("POST %s/restaurants/{slug}/dishes/{id}/image", p), guard(h.uploadDishImage))
		mux.HandleFunc(fmt.Sprintf("POST %s/upload/image", p), guard(h.genericImageUpload))
	}
}

type restaurantResponse struct {
	Slug         string   `json:"slug"`
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	Cuisine      string   `json:"cuisine,omitempty"`
	IsOpen       bool     `json:"isOpen"`
	OpenTime     string   `json:"openTime,omitempty"`
	ClosingTime  string   `json:"closingTime,omitempty"`
	Rating       float64  `json:"rating"`
	ReviewsCount int      `json:"reviewsCount"`
	DeliveryFee  float64  `json:"deliveryFee"`
	CEP          string   `json:"CEP,omitempty"`
	DeliveryTime string   `json:"deliveryTime,omitempty"`
	Neighborhood string   `json:"neighborhood,omitempty"`
	Phone        string   `json:"phone,omitempty"`
	BannerURL    string   `json:"bannerUrl,omitempty"`
	LogoURL      string   `json:"logoUrl,omitempty"`
	MenuItems    []string `json:"menu_items,omitempty"`
}

func toRestaurantResponse(r *Restaurant) restaurantResponse {
	if r == nil {
		return restaurantResponse{}
	}
	return restaurantResponse{
		Slug:         r.Slug,
		Name:         r.Name,
		Description:  r.Description,
		Cuisine:      r.Cuisine,
		IsOpen:       r.IsOpen,
		OpenTime:     r.OpenTime,
		ClosingTime:  r.ClosingTime,
		Rating:       r.Rating,
		ReviewsCount: r.ReviewsCount,
		DeliveryFee:  r.DeliveryFee,
		CEP:          r.CEP,
		DeliveryTime: r.DeliveryTime,
		Neighborhood: r.Neighborhood,
		Phone:        r.Phone,
		BannerURL:    r.BannerURL,
		LogoURL:      r.LogoURL,
		MenuItems:    r.MenuItems,
	}
}

type createDishRequest struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Category    string   `json:"category"`
	Tags        []string `json:"tags"`
	Price       float64  `json:"price"`
	ImageURL    string   `json:"image_url"`
}

type updateDishRequest struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Category    string   `json:"category"`
	Tags        []string `json:"tags"`
	Price       float64  `json:"price"`
	ImageURL    string   `json:"image_url"`
}

type dishResponse struct {
	ID             any      `json:"id"`
	RestaurantSlug string   `json:"restaurant_slug,omitempty"`
	Name           string   `json:"name"`
	Description    string   `json:"description,omitempty"`
	Category       string   `json:"category,omitempty"`
	Tags           []string `json:"tags,omitempty"`
	Price          float64  `json:"price"`
	ImageURL       string   `json:"imageUrl,omitempty"`
}

func toDishResponse(d *Dish) dishResponse {
	if d == nil {
		return dishResponse{}
	}
	return dishResponse{
		ID:             d.ID,
		RestaurantSlug: d.RestaurantSlug,
		Name:           d.Name,
		Description:    d.Description,
		Category:       d.Category,
		Tags:           d.Tags,
		Price:          d.Price,
		ImageURL:       d.ImageURL,
	}
}

func (h *Handler) listDishes(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	slug := strings.ToLower(r.PathValue("slug"))

	dishes, err := h.catalog.ListDishes(r.Context(), slug)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "restaurant not found"})
			return
		}
		http.Error(w, `{"error":"database error"}`, http.StatusInternalServerError)
		return
	}

	resp := make([]dishResponse, 0, len(dishes))
	for _, d := range dishes {
		resp = append(resp, toDishResponse(&d))
	}
	_ = json.NewEncoder(w).Encode(resp)
}

func (h *Handler) createDish(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	slug := strings.ToLower(r.PathValue("slug"))

	var req createDishRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid request body"})
		return
	}

	dish, err := h.catalog.AddDish(r.Context(), NewDish{
		RestaurantSlug: slug,
		Name:           req.Name,
		Description:    req.Description,
		Category:       req.Category,
		Tags:           req.Tags,
		Price:          req.Price,
		ImageURL:       req.ImageURL,
	})
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "restaurant not found"})
			return
		}
		if errors.Is(err, ErrValidation) {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}
		http.Error(w, `{"error":"failed to create dish"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(toDishResponse(dish))
}

func (h *Handler) updateDish(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	slug := strings.ToLower(r.PathValue("slug"))
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid dish ID"})
		return
	}

	var req updateDishRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid request body"})
		return
	}

	dish, err := h.catalog.UpdateDish(r.Context(), slug, id, UpdateDishInput{
		Name:        req.Name,
		Description: req.Description,
		Category:    req.Category,
		Tags:        req.Tags,
		Price:       req.Price,
		ImageURL:    req.ImageURL,
	})
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "dish not found"})
			return
		}
		http.Error(w, `{"error":"database error"}`, http.StatusInternalServerError)
		return
	}
	_ = json.NewEncoder(w).Encode(toDishResponse(dish))
}

func (h *Handler) deleteDish(w http.ResponseWriter, r *http.Request) {
	slug := strings.ToLower(r.PathValue("slug"))
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid dish ID"})
		return
	}

	err = h.catalog.DeleteDish(r.Context(), slug, id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "dish not found"})
			return
		}
		http.Error(w, `{"error":"database error"}`, http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) restaurantMenu(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	slug := strings.ToLower(r.PathValue("slug"))

	restaurant, categories, dishes, err := h.catalog.GetRestaurantMenu(r.Context(), slug)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "restaurant not found"})
			return
		}
		http.Error(w, `{"error":"database error"}`, http.StatusInternalServerError)
		return
	}

	dishResponses := make([]dishResponse, 0, len(dishes))
	for _, d := range dishes {
		dishResponses = append(dishResponses, toDishResponse(&d))
	}

	_ = json.NewEncoder(w).Encode(map[string]any{
		"restaurant": toRestaurantResponse(restaurant),
		"categories": categories,
		"dishes":     dishResponses,
	})
}

func saveUploadedImage(r *http.Request, fieldName string, subfolder string) (publicURL string, err error) {
	r.Body = http.MaxBytesReader(nil, r.Body, 5<<20)

	file, header, err := r.FormFile(fieldName)
	if err != nil {
		return "", errors.New("missing or invalid file in '" + fieldName + "' field")
	}
	defer file.Close()

	ext := strings.ToLower(filepath.Ext(header.Filename))
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" && ext != ".webp" && ext != ".gif" {
		return "", errors.New("unsupported image format (supported: jpg, jpeg, png, webp, gif)")
	}

	if subfolder == "" {
		subfolder = "common"
	}
	targetDir := filepath.Join("media", subfolder)
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create media directory: %w", err)
	}

	filename := fmt.Sprintf("%d-%s%s", time.Now().UnixNano(), strings.ReplaceAll(filepath.Base(header.Filename), ext, ""), ext)
	filePath := filepath.Join(targetDir, filename)

	dst, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		return "", fmt.Errorf("failed to write image content: %w", err)
	}

	return fmt.Sprintf("/media/%s/%s", subfolder, filename), nil
}

func (h *Handler) uploadDishImage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	slug := strings.ToLower(r.PathValue("slug"))
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid dish ID"})
		return
	}

	imageURL, err := saveUploadedImage(r, "image", "dishes")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	updatedDish, err := h.catalog.UpdateDishImage(r.Context(), slug, id, imageURL)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "dish not found"})
			return
		}
		http.Error(w, `{"error":"failed to update dish with image"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(toDishResponse(updatedDish))
}

func (h *Handler) genericImageUpload(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	imageURL, err := saveUploadedImage(r, "image", "common")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"image_url": imageURL,
	})
}
