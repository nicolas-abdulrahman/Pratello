package restaurants

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// Handler handles HTTP requests for restaurants, menus, dishes, and image uploads.
type Handler struct {
	svc *Service
}

// NewHandler constructs an HTTP Handler for the restaurants domain.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// Routes registers all restaurant, dish, and upload routes on the provided mux.
func (h *Handler) Routes(mux *http.ServeMux) {
	for _, p := range []string{"", "/api"} {
		// Restaurants
		mux.HandleFunc(fmt.Sprintf("GET %s/restaurants", p), h.listRestaurants)
		mux.HandleFunc(fmt.Sprintf("POST %s/restaurants", p), h.createRestaurant)
		mux.HandleFunc(fmt.Sprintf("GET %s/restaurants/{slug}", p), h.getRestaurant)
		mux.HandleFunc(fmt.Sprintf("PUT %s/restaurants/{slug}", p), h.updateRestaurant)
		mux.HandleFunc(fmt.Sprintf("DELETE %s/restaurants/{slug}", p), h.deleteRestaurant)

		// Dishes
		mux.HandleFunc(fmt.Sprintf("GET %s/restaurants/{slug}/dishes", p), h.listDishes)
		mux.HandleFunc(fmt.Sprintf("POST %s/restaurants/{slug}/dishes", p), h.createDish)
		mux.HandleFunc(fmt.Sprintf("PUT %s/restaurants/{slug}/dishes/{id}", p), h.updateDish)
		mux.HandleFunc(fmt.Sprintf("DELETE %s/restaurants/{slug}/dishes/{id}", p), h.deleteDish)

		// Menu
		mux.HandleFunc(fmt.Sprintf("GET %s/restaurants/{slug}/menu", p), h.restaurantMenu)

		// Image uploads
		mux.HandleFunc(fmt.Sprintf("POST %s/restaurants/{slug}/dishes/{id}/image", p), h.uploadDishImage)
		mux.HandleFunc(fmt.Sprintf("POST %s/upload/image", p), h.genericImageUpload)
	}
}

// ---------------- DTO REQUEST & RESPONSE TYPES ----------------

type createRestaurantRequest struct {
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
}

type updateRestaurantRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
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

// ---------------- HANDLER IMPLEMENTATIONS ----------------

func (h *Handler) listRestaurants(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	list, err := h.svc.ListRestaurants(r.Context())
	if err != nil {
		http.Error(w, `{"error":"database error"}`, http.StatusInternalServerError)
		return
	}

	resp := make([]restaurantResponse, 0, len(list))
	for _, res := range list {
		resp = append(resp, toRestaurantResponse(&res))
	}
	_ = json.NewEncoder(w).Encode(resp)
}

func (h *Handler) createRestaurant(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req createRestaurantRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid request body"})
		return
	}

	res, err := h.svc.CreateRestaurant(r.Context(), NewRestaurant{
		Name:        req.Name,
		Slug:        req.Slug,
		Description: req.Description,
	})
	if err != nil {
		if errors.Is(err, ErrConflict) {
			w.WriteHeader(http.StatusConflict)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "restaurant with this slug already exists"})
			return
		}
		if errors.Is(err, ErrValidation) {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}
		http.Error(w, `{"error":"failed to create restaurant"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(toRestaurantResponse(res))
}

func (h *Handler) getRestaurant(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	slug := strings.ToLower(r.PathValue("slug"))

	res, err := h.svc.GetRestaurant(r.Context(), slug)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "restaurant not found"})
		return
	}
	_ = json.NewEncoder(w).Encode(toRestaurantResponse(res))
}

func (h *Handler) updateRestaurant(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	slug := strings.ToLower(r.PathValue("slug"))

	var req updateRestaurantRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid request body"})
		return
	}

	updated, err := h.svc.UpdateRestaurant(r.Context(), slug, UpdateRestaurantInput{
		Name:        req.Name,
		Description: req.Description,
	})
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "restaurant not found"})
			return
		}
		http.Error(w, `{"error":"database error"}`, http.StatusInternalServerError)
		return
	}
	_ = json.NewEncoder(w).Encode(toRestaurantResponse(updated))
}

func (h *Handler) deleteRestaurant(w http.ResponseWriter, r *http.Request) {
	slug := strings.ToLower(r.PathValue("slug"))

	err := h.svc.DeleteRestaurant(r.Context(), slug)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "restaurant not found"})
			return
		}
		http.Error(w, `{"error":"database error"}`, http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) listDishes(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	slug := strings.ToLower(r.PathValue("slug"))

	dishes, err := h.svc.ListDishes(r.Context(), slug)
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

	dish, err := h.svc.AddDish(r.Context(), NewDish{
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

	dish, err := h.svc.UpdateDish(r.Context(), slug, id, UpdateDishInput{
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

	err = h.svc.DeleteDish(r.Context(), slug, id)
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

	restaurant, categories, dishes, err := h.svc.GetRestaurantMenu(r.Context(), slug)
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

// ---------------- IMAGE UPLOADS ----------------

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

	updatedDish, err := h.svc.UpdateDishImage(r.Context(), slug, id, imageURL)
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
