package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// Dish represents a menu item belonging to a restaurant.
type Dish struct {
	ID             any      `json:"id"`
	RestaurantSlug string   `json:"-"`
	Name           string   `json:"name"`
	Description    string   `json:"description,omitempty"`
	Category       string   `json:"category,omitempty"`
	Tags           []string `json:"tags,omitempty"`
	Price          float64  `json:"price"`
	ImageURL       string   `json:"imageUrl,omitempty"`
}

// SeedDefaultRestaurant resets and re-seeds "mamma-mia-mario" with sample menu items.
func (a *App) SeedDefaultRestaurant() error {
	ctx := context.Background()

	// Seed Mamma Mia Mario
	marioSlug := "mamma-mia-mario"
	_ = a.Restaurants.Delete(ctx, marioSlug)
	if err := a.Restaurants.Create(ctx, &Restaurant{
		Slug:         marioSlug,
		Name:         "Mamma mia Mario",
		Description:  "Um pedacinho de Napoli. Pizza no fogão a lenha, feita com amor atrás do armário.",
		Cuisine:      "Italiana",
		IsOpen:       true,
		OpenTime:     "07:00",
		ClosingTime:  "22:00",
		Rating:       4.9,
		ReviewsCount: 324,
		DeliveryFee:  1.99,
		CEP:          "15555-555",
		DeliveryTime: "20–30 min",
		Neighborhood: "Mission District",
		Phone:        "(81) 95552-0148",
		BannerURL:    "https://lh3.googleusercontent.com/aida-public/AB6AXuD8i7X0qDoCvkouDq-uUv7UEDtkKSTah3gKO68ElZOQTiDH0KjKi_fyaIcGAqaF7mZ_ZgqueJ6j_U-Z-tt30B1fuhs28n6b5ZhcHjaFmjEgmxSSCJChcNGK27S-R-GsleUn0DChdITkkqd3lz-H6R4XMb81BP7tfPtAvQa2d2wjv4dqHh-zPRTB2cfC81OAkENDitgwJzfayVROlNm2Y_cpClw84FTUl8AaFP--O1M",
		LogoURL:      "https://i.pinimg.com/736x/9e/f2/b9/9ef2b9c1ac760f54c9e7aeb5c1784460.jpg",
	}); err != nil {
		return fmt.Errorf("failed to insert mamma-mia-mario: %w", err)
	}

	sampleDishes := []NewDish{
		{
			RestaurantSlug: marioSlug,
			Name:           "Margherita",
			Description:    "San Marzano tomato, fior di latte mozzarella, fresh basil.",
			Category:       "Pizzas",
			Tags:           []string{"MOST LOVED", "MOST ORDERED"},
			Price:          16.00,
			ImageURL:       "https://lh3.googleusercontent.com/aida-public/AB6AXuCb-eU4XxPHWKrxNYHn7-TPGZZiDJ7J43_wyHW35GGvB4ZYni8tfx_WXwM5FpIXR5sPzrKxs7IsbzM8We9VCULm7jAj3tnmjxGjgDsOIjDnBTMIWNwM3hGed30aeoT2VhHS9PkWMmiaJALNAfdW7EcUR2jaBsZXviPHVg-BLQ_Yf-JaqHsdxRAb3o6FWddiPqt9vxwkUpiL3darOkhD4kKfXX9-ff_dhL3G24HQmLw",
		},
		{
			RestaurantSlug: marioSlug,
			Name:           "Spicy Diavola",
			Description:    "Spicy salami, mozzarella, tomato and a drizzle of chili honey.",
			Category:       "Pizzas",
			Tags:           []string{"MOST LOVED"},
			Price:          19.00,
			ImageURL:       "https://lh3.googleusercontent.com/aida-public/AB6AXuCTmOctI0RewSPwuz-BWjt1KjY06cnghRQd_5aD3yZ1mdACQ5v5UPrVcEaIeltyls3yjVEikrahYf-2fyZK4fwkm_VkaoGT7Rf2YHIePG4HfGnJJBV-j7xf-amYSPa3IDHVXFqKEhwbdbJomOR_fOEgiJJV31ru_FWjl8K31Wfai5rpjXVPa8DMj8HxJ_xQ4WWAwo0ZsO10gZXIHGP9fjCJZqNz6-vn7YPrM4fqcM8",
		},
		{
			RestaurantSlug: marioSlug,
			Name:           "Burrata & tomatoes",
			Description:    "Creamy burrata, heirloom tomatoes, basil and extra virgin olive oil.",
			Category:       "Acompanhamentos",
			Tags:           []string{},
			Price:          14.00,
			ImageURL:       "https://lh3.googleusercontent.com/aida-public/AB6AXuBFETatsb6xiRzdMPM-1wMApk2RxLzpPFaJRKZ7yhdMVwciJo0Y53nDdCJM2yXu6qiqF4MZhPHNMAMwt_xVzRCnp2QM2M1YnoXM0bZks8dQ9-rrvPxY2LrYt1WwMYvUKvwgdw4FMn-O9G0clEClprlRhqIKNsE95Xpbulp8jfHSz4g_3Snl6Isi9kNw_rLR0ZtjJf8FsS4esS6B-XfCxNztdT6WBZKjFfkPHG95JYM",
		},
	}

	for _, d := range sampleDishes {
		if _, err := a.Dishes.Create(ctx, d); err != nil {
			return fmt.Errorf("failed to insert sample dish: %w", err)
		}
	}

	return nil
}

// ---------------- RESTAURANTS HANDLERS ----------------

// handleListRestaurants returns all registered restaurants.
func (a *App) handleListRestaurants(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	list, err := a.Restaurants.List(r.Context())
	if err != nil {
		http.Error(w, `{"error":"database error"}`, http.StatusInternalServerError)
		return
	}
	if list == nil {
		list = make([]Restaurant, 0)
	}
	_ = json.NewEncoder(w).Encode(list)
}

// handleCreateRestaurant routes through Layer 2 Business Service for validation & slug logic.
func (a *App) handleCreateRestaurant(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var body struct {
		Name        string `json:"name"`
		Slug        string `json:"slug"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid request body"})
		return
	}

	input := NewRestaurant{
		Name:        body.Name,
		Slug:        body.Slug,
		Description: body.Description,
	}

	restaurant, err := a.RestaurantService.CreateRestaurant(r.Context(), input)
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
	_ = json.NewEncoder(w).Encode(restaurant)
}

// handleGetRestaurant fetches a single restaurant by slug.
func (a *App) handleGetRestaurant(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	slug := strings.ToLower(r.PathValue("slug"))

	res, err := a.Restaurants.GetBySlug(r.Context(), slug)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "restaurant not found"})
		return
	}
	_ = json.NewEncoder(w).Encode(res)
}

func (a *App) handleUpdateRestaurant(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	slug := strings.ToLower(r.PathValue("slug"))

	var body struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid request body"})
		return
	}

	input := UpdateRestaurantInput{
		Name:        body.Name,
		Description: body.Description,
	}

	updated, err := a.Restaurants.Update(r.Context(), slug, input)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "restaurant not found"})
			return
		}
		http.Error(w, `{"error":"database error"}`, http.StatusInternalServerError)
		return
	}
	_ = json.NewEncoder(w).Encode(updated)
}

func (a *App) handleDeleteRestaurant(w http.ResponseWriter, r *http.Request) {
	slug := strings.ToLower(r.PathValue("slug"))

	err := a.Restaurants.Delete(r.Context(), slug)
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

// handleRestaurantMenu returns the consolidated profile, categories, and dishes for the restaurant page.
func (a *App) handleRestaurantMenu(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	slug := strings.ToLower(r.PathValue("slug"))

	restaurant, err := a.Restaurants.GetBySlug(r.Context(), slug)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "restaurant not found"})
			return
		}
		http.Error(w, `{"error":"database error"}`, http.StatusInternalServerError)
		return
	}

	dishes, err := a.Dishes.ListByRestaurant(r.Context(), slug)
	if err != nil && !errors.Is(err, ErrNotFound) {
		http.Error(w, `{"error":"database error"}`, http.StatusInternalServerError)
		return
	}
	if dishes == nil {
		dishes = make([]Dish, 0)
	}

	// Extract unique categories in order
	catSet := make(map[string]bool)
	var categories []string
	categories = append(categories, "Mais Pedidos")
	catSet["Mais Pedidos"] = true

	for _, d := range dishes {
		cat := d.Category
		if cat == "" {
			cat = "Pizzas"
		}
		if !catSet[cat] {
			catSet[cat] = true
			categories = append(categories, cat)
		}
	}

	_ = json.NewEncoder(w).Encode(map[string]any{
		"restaurant": restaurant,
		"categories": categories,
		"dishes":     dishes,
	})
}

// ---------------- DISHES HANDLERS ----------------

func (a *App) handleListDishes(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	slug := strings.ToLower(r.PathValue("slug"))

	dishes, err := a.Dishes.ListByRestaurant(r.Context(), slug)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "restaurant not found"})
			return
		}
		http.Error(w, `{"error":"database error"}`, http.StatusInternalServerError)
		return
	}
	if dishes == nil {
		dishes = make([]Dish, 0)
	}
	_ = json.NewEncoder(w).Encode(dishes)
}

// handleCreateDish routes through Layer 2 Business Service for price and restaurant validation.
func (a *App) handleCreateDish(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	slug := strings.ToLower(r.PathValue("slug"))

	var body struct {
		Name     string  `json:"name"`
		Price    float64 `json:"price"`
		ImageURL string  `json:"image_url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid request body"})
		return
	}

	input := NewDish{
		RestaurantSlug: slug,
		Name:           body.Name,
		Price:          body.Price,
		ImageURL:       body.ImageURL,
	}

	dish, err := a.DishService.AddDish(r.Context(), input)
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
	_ = json.NewEncoder(w).Encode(dish)
}

func (a *App) handleUpdateDish(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	slug := strings.ToLower(r.PathValue("slug"))
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid dish ID"})
		return
	}

	var body struct {
		Name     string  `json:"name"`
		Price    float64 `json:"price"`
		ImageURL string  `json:"image_url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid request body"})
		return
	}

	input := UpdateDishInput{
		Name:     body.Name,
		Price:    body.Price,
		ImageURL: body.ImageURL,
	}

	dish, err := a.Dishes.Update(r.Context(), slug, id, input)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "dish not found"})
			return
		}
		http.Error(w, `{"error":"database error"}`, http.StatusInternalServerError)
		return
	}
	_ = json.NewEncoder(w).Encode(dish)
}

func (a *App) handleDeleteDish(w http.ResponseWriter, r *http.Request) {
	slug := strings.ToLower(r.PathValue("slug"))
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid dish ID"})
		return
	}

	err = a.Dishes.Delete(r.Context(), slug, id)
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

// ---------------- IMAGE UPLOAD HANDLERS ----------------

// saveUploadedImage validates and persists a multipart file to disk under media/<subfolder>/.
func saveUploadedImage(r *http.Request, fieldName string, subfolder string) (publicURL string, err error) {
	// Max upload size: 5 MB
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

// handleUploadDishImage handles multipart upload for a specific dish, saves the file to disk under media/dishes/,
// and updates the dish record in PostgreSQL with the public image URL.
func (a *App) handleUploadDishImage(w http.ResponseWriter, r *http.Request) {
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

	updatedDish, err := a.Dishes.UpdateImage(r.Context(), slug, id, imageURL)
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
	_ = json.NewEncoder(w).Encode(updatedDish)
}

// handleGenericImageUpload accepts an image file upload and returns its static public URL under media/common/.
func (a *App) handleGenericImageUpload(w http.ResponseWriter, r *http.Request) {
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

// ---------------- AUTHENTICATION & ACCOUNTS HANDLERS ----------------

// extractBearerToken extracts the JWT token from the Authorization: Bearer <token> header.
func extractBearerToken(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return ""
	}
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) == 2 && strings.EqualFold(parts[0], "bearer") {
		return strings.TrimSpace(parts[1])
	}
	return ""
}

// handleAuthRegister processes account creation (Owner, Manager, User, Admin).
func (a *App) handleAuthRegister(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var body RegisterInput
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "corpo da requisição inválido (JSON esperado)"})
		return
	}

	account, err := a.AuthService.Register(r.Context(), body)
	if err != nil {
		if errors.Is(err, ErrConflict) {
			w.WriteHeader(http.StatusConflict)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "este e-mail já está cadastrado"})
			return
		}
		if errors.Is(err, ErrValidation) {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}
		log.Printf("[AuthRegister Error] Falha ao registrar %q: %v\n", body.Email, err)
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": fmt.Sprintf("falha interna ao criar conta no banco: %v", err)})
		return
	}

	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(account)
}

// handleAuthLogin verifies credentials and returns a JWT token.
func (a *App) handleAuthLogin(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "corpo da requisição inválido (JSON esperado)"})
		return
	}

	token, account, err := a.AuthService.Login(r.Context(), body.Email, body.Password)
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "e-mail ou senha inválidos"})
			return
		}
		log.Printf("[AuthLogin Error] Erro ao autenticar %q: %v\n", body.Email, err)
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": fmt.Sprintf("falha interna de autenticação no servidor: %v", err)})
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"token":   token,
		"account": account,
	})
}

// handleAuthMe returns profile info for the currently authenticated account.
func (a *App) handleAuthMe(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	tokenStr := extractBearerToken(r)
	if tokenStr == "" {
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "missing or invalid authorization token"})
		return
	}

	claims, err := a.AuthService.ParseToken(tokenStr)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid or expired token"})
		return
	}

	account, err := a.AuthService.GetAccountByID(r.Context(), claims.AccountID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "account not found"})
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "failed to fetch account"})
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(account)
}

// handleListSysadmins returns all global superadministrators.
func (a *App) handleListSysadmins(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	list, err := a.AuthService.ListSysadmins(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "database error"})
		return
	}
	_ = json.NewEncoder(w).Encode(list)
}

// handleCreateSysadmin promotes an account to sysadmin.
func (a *App) handleCreateSysadmin(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var body struct {
		AccountID int `json:"account_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid request body"})
		return
	}

	sa, err := a.AuthService.AddSysadmin(r.Context(), body.AccountID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "account not found"})
			return
		}
		if errors.Is(err, ErrValidation) {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "failed to promote sysadmin"})
		return
	}

	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(sa)
}

// handleListManagers returns managers linked to a specific restaurant.
func (a *App) handleListManagers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	slug := strings.ToLower(r.PathValue("slug"))

	list, err := a.AuthService.ListManagers(r.Context(), slug)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "restaurant not found"})
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "database error"})
		return
	}
	_ = json.NewEncoder(w).Encode(list)
}

// handleCreateManager associates an account as a manager for a restaurant.
func (a *App) handleCreateManager(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	slug := strings.ToLower(r.PathValue("slug"))

	var body struct {
		AccountID int `json:"account_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid request body"})
		return
	}

	manager, err := a.AuthService.AddManager(r.Context(), slug, body.AccountID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "restaurant or account not found"})
			return
		}
		if errors.Is(err, ErrValidation) {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "failed to link manager"})
		return
	}

	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(manager)
}

// handleDeleteManager unlinks a manager account from a restaurant.
func (a *App) handleDeleteManager(w http.ResponseWriter, r *http.Request) {
	slug := strings.ToLower(r.PathValue("slug"))
	idStr := r.PathValue("account_id")
	accountID, err := strconv.Atoi(idStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid account_id"})
		return
	}

	if err := a.AuthService.RemoveManager(r.Context(), slug, accountID); err != nil {
		if errors.Is(err, ErrNotFound) {
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "manager not found"})
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "database error"})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// requireAuth validates the Bearer JWT token on protected endpoints.
func (a *App) requireAuth(w http.ResponseWriter, r *http.Request) *AccountClaims {
	tokenStr := extractBearerToken(r)
	if tokenStr == "" {
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "missing or invalid authorization token"})
		return nil
	}
	claims, err := a.AuthService.ParseToken(tokenStr)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid or expired token"})
		return nil
	}
	return claims
}

// ---------------- OWNER ENTITY AUTHORITY HANDLERS ----------------

func (a *App) handleListOwners(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	slug := strings.ToLower(r.PathValue("slug"))

	list, err := a.AuthService.ListOwners(r.Context(), slug)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "restaurant not found"})
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "database error"})
		return
	}
	_ = json.NewEncoder(w).Encode(list)
}

func (a *App) handlePromoteOwner(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	slug := strings.ToLower(r.PathValue("slug"))

	claims := a.requireAuth(w, r)
	if claims == nil {
		return
	}

	var body struct {
		AccountID int `json:"account_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid request body"})
		return
	}

	owner, err := a.AuthService.PromoteOwner(r.Context(), slug, claims.AccountID, body.AccountID)
	if err != nil {
		if errors.Is(err, ErrForbidden) {
			w.WriteHeader(http.StatusForbidden)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "forbidden: only an owner can promote another owner"})
			return
		}
		if errors.Is(err, ErrNotFound) {
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "restaurant or target account not found"})
			return
		}
		if errors.Is(err, ErrValidation) {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "failed to promote owner"})
		return
	}

	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(owner)
}

func (a *App) handleRemoveOwner(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	slug := strings.ToLower(r.PathValue("slug"))
	idStr := r.PathValue("account_id")
	accountID, err := strconv.Atoi(idStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid account_id"})
		return
	}

	claims := a.requireAuth(w, r)
	if claims == nil {
		return
	}

	err = a.AuthService.RemoveOwner(r.Context(), slug, claims.AccountID, accountID)
	if err != nil {
		if errors.Is(err, ErrForbidden) {
			w.WriteHeader(http.StatusForbidden)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "forbidden: only an owner can remove an owner"})
			return
		}
		if errors.Is(err, ErrOrphanEntity) {
			w.WriteHeader(http.StatusConflict)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "cannot remove the last owner: entity cannot be orphaned"})
			return
		}
		if errors.Is(err, ErrNotFound) {
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "owner not found"})
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "database error"})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ---------------- PHYSICAL RESTAURANTS (BRANCHES) ----------------

func (a *App) handleCreateBranch(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	slug := strings.ToLower(r.PathValue("slug"))

	claims := a.requireAuth(w, r)
	if claims == nil {
		return
	}

	var body struct {
		Name      string `json:"name"`
		CNPJ      string `json:"cnpj"`
		OwnerCNPJ string `json:"owner_cnpj"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid request body"})
		return
	}

	pr, err := a.AuthService.CreatePhysicalRestaurant(r.Context(), slug, claims.AccountID, body.Name, body.CNPJ, body.OwnerCNPJ)
	if err != nil {
		if errors.Is(err, ErrForbidden) {
			w.WriteHeader(http.StatusForbidden)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "forbidden: requires manager or owner role"})
			return
		}
		if errors.Is(err, ErrCNPJMismatch) {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "owner CNPJ verification failed: does not match restaurant registration"})
			return
		}
		if errors.Is(err, ErrValidation) {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}
		if errors.Is(err, ErrNotFound) {
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "restaurant not found"})
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "failed to create physical restaurant"})
		return
	}

	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(pr)
}

func (a *App) handleListBranches(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	slug := strings.ToLower(r.PathValue("slug"))

	branches, err := a.AuthService.ListPhysicalRestaurants(r.Context(), slug)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "failed to list branches"})
		return
	}
	_ = json.NewEncoder(w).Encode(branches)
}

// ---------------- EMPLOYEES (HR RECORDS) ----------------

func (a *App) handleListEmployees(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	branchIDStr := r.PathValue("branch_id")
	branchID, err := strconv.Atoi(branchIDStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid branch_id"})
		return
	}

	claims := a.requireAuth(w, r)
	if claims == nil {
		return
	}

	employees, err := a.AuthService.ListEmployees(r.Context(), branchID, claims.AccountID)
	if err != nil {
		if errors.Is(err, ErrForbidden) {
			w.WriteHeader(http.StatusForbidden)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "forbidden: requires manager or owner role"})
			return
		}
		if errors.Is(err, ErrNotFound) {
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "physical restaurant not found"})
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "database error"})
		return
	}
	_ = json.NewEncoder(w).Encode(employees)
}

func (a *App) handleCreateEmployee(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	branchIDStr := r.PathValue("branch_id")
	branchID, err := strconv.Atoi(branchIDStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid branch_id"})
		return
	}

	claims := a.requireAuth(w, r)
	if claims == nil {
		return
	}

	var body struct {
		Name     string `json:"name"`
		CPF      string `json:"cpf"`
		Position string `json:"position"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid request body"})
		return
	}

	emp, err := a.AuthService.CreateEmployee(r.Context(), branchID, claims.AccountID, body.Name, body.CPF, body.Position)
	if err != nil {
		if errors.Is(err, ErrForbidden) {
			w.WriteHeader(http.StatusForbidden)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "forbidden: requires manager or owner role"})
			return
		}
		if errors.Is(err, ErrValidation) {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}
		if errors.Is(err, ErrNotFound) {
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "physical restaurant not found"})
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "failed to create employee record"})
		return
	}

	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(emp)
}

func (a *App) handleInviteEmployee(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	branchID, _ := strconv.Atoi(r.PathValue("branch_id"))
	empID, _ := strconv.Atoi(r.PathValue("employee_id"))

	claims := a.requireAuth(w, r)
	if claims == nil {
		return
	}

	var body struct {
		AccountID int `json:"account_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid request body"})
		return
	}

	emp, err := a.AuthService.InviteEmployee(r.Context(), branchID, empID, claims.AccountID, body.AccountID)
	if err != nil {
		if errors.Is(err, ErrForbidden) {
			w.WriteHeader(http.StatusForbidden)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "forbidden: requires manager or owner role"})
			return
		}
		if errors.Is(err, ErrNotFound) {
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "employee or account not found"})
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "failed to link employee account"})
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(emp)
}

// ---------------- TERMINAL PIN & SHIFT SESSION ----------------

func (a *App) handleOpenTerminalSession(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	branchIDStr := r.PathValue("branch_id")
	branchID, err := strconv.Atoi(branchIDStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid branch_id"})
		return
	}

	var body struct {
		PIN      string `json:"pin"`
		DeviceID string `json:"device_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid request body"})
		return
	}

	session, err := a.AuthService.OpenTerminalSession(r.Context(), branchID, body.PIN, body.DeviceID)
	if err != nil {
		if errors.Is(err, ErrTerminalLocked) {
			w.WriteHeader(http.StatusTooManyRequests)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "terminal is locked due to too many failed attempts (try again in 15 minutes)"})
			return
		}
		if errors.Is(err, ErrInvalidPIN) {
			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid terminal PIN"})
			return
		}
		if errors.Is(err, ErrValidation) {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}
		if errors.Is(err, ErrNotFound) {
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "physical restaurant not found"})
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "failed to open terminal session"})
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(session)
}

func (a *App) handleTerminalOrders(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	branchIDStr := r.PathValue("branch_id")
	branchID, err := strconv.Atoi(branchIDStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid branch_id"})
		return
	}

	termToken := r.Header.Get("X-Terminal-Session")
	if termToken == "" {
		termToken = extractBearerToken(r)
	}

	session, err := a.AuthService.VerifyTerminalSession(r.Context(), termToken, branchID)
	if err != nil {
		if errors.Is(err, ErrForbidden) {
			w.WriteHeader(http.StatusForbidden)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "terminal session is not scoped to this branch"})
			return
		}
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid or expired terminal shift session"})
		return
	}

	// Operational access only: return shift orders for this branch
	_ = json.NewEncoder(w).Encode(map[string]any{
		"session_id":             session.ID,
		"physical_restaurant_id": session.PhysicalRestaurantID,
		"device_id":              session.DeviceID,
		"scope":                  "operational_orders_view_only",
		"orders":                 []any{},
	})
}


