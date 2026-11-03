package main

import (
	"context"
	"database/sql"
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

// App wires together Layer 1 (HTTP Transport), Layer 2 (Business Services), and Layer 3 (Persistence Stores).
type App struct {
	RestaurantService *RestaurantService
	DishService       *DishService
	Restaurants       RestaurantStore
	Dishes            DishStore
	DB                *sql.DB
}

func NewApp(db *sql.DB) *App {
	app := &App{
		DB:          db,
		Restaurants: NewSQLRestaurantStore(db),
		Dishes:      NewSQLDishStore(db),
	}
	if db != nil {
		_ = app.InitSchema()
	}
	app.RestaurantService = NewRestaurantService(app.Restaurants)
	app.DishService = NewDishService(app.Restaurants, app.Dishes)
	return app
}

func (a *App) InitSchema() error {
	if a.DB == nil {
		return nil
	}
	schema := `
	CREATE TABLE IF NOT EXISTS restaurants (
		slug TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		description TEXT,
		cuisine TEXT DEFAULT '',
		is_open BOOLEAN DEFAULT true,
		open_time TEXT DEFAULT '07:00',
		closing_time TEXT DEFAULT '22:00',
		rating NUMERIC(3, 2) DEFAULT 5.0,
		reviews_count INT DEFAULT 0,
		delivery_fee NUMERIC(10, 2) DEFAULT 0.00,
		cep TEXT DEFAULT '',
		delivery_time TEXT DEFAULT '20–30 min',
		neighborhood TEXT DEFAULT '',
		phone TEXT DEFAULT '',
		banner_url TEXT DEFAULT '',
		logo_url TEXT DEFAULT ''
	);

	CREATE TABLE IF NOT EXISTS dishes (
		id SERIAL PRIMARY KEY,
		restaurant_slug TEXT NOT NULL REFERENCES restaurants(slug) ON DELETE CASCADE,
		name TEXT NOT NULL,
		description TEXT DEFAULT '',
		category TEXT DEFAULT 'Pizzas',
		tags TEXT DEFAULT '',
		price NUMERIC(10, 2) NOT NULL,
		image_url TEXT DEFAULT ''
	);

	ALTER TABLE restaurants ADD COLUMN IF NOT EXISTS cuisine TEXT DEFAULT '';
	ALTER TABLE restaurants ADD COLUMN IF NOT EXISTS is_open BOOLEAN DEFAULT true;
	ALTER TABLE restaurants ADD COLUMN IF NOT EXISTS open_time TEXT DEFAULT '07:00';
	ALTER TABLE restaurants ADD COLUMN IF NOT EXISTS closing_time TEXT DEFAULT '22:00';
	ALTER TABLE restaurants ADD COLUMN IF NOT EXISTS rating NUMERIC(3, 2) DEFAULT 5.0;
	ALTER TABLE restaurants ADD COLUMN IF NOT EXISTS reviews_count INT DEFAULT 0;
	ALTER TABLE restaurants ADD COLUMN IF NOT EXISTS delivery_fee NUMERIC(10, 2) DEFAULT 0.00;
	ALTER TABLE restaurants ADD COLUMN IF NOT EXISTS cep TEXT DEFAULT '';
	ALTER TABLE restaurants ADD COLUMN IF NOT EXISTS delivery_time TEXT DEFAULT '20–30 min';
	ALTER TABLE restaurants ADD COLUMN IF NOT EXISTS neighborhood TEXT DEFAULT '';
	ALTER TABLE restaurants ADD COLUMN IF NOT EXISTS phone TEXT DEFAULT '';
	ALTER TABLE restaurants ADD COLUMN IF NOT EXISTS banner_url TEXT DEFAULT '';
	ALTER TABLE restaurants ADD COLUMN IF NOT EXISTS logo_url TEXT DEFAULT '';

	ALTER TABLE dishes ADD COLUMN IF NOT EXISTS description TEXT DEFAULT '';
	ALTER TABLE dishes ADD COLUMN IF NOT EXISTS category TEXT DEFAULT 'Pizzas';
	ALTER TABLE dishes ADD COLUMN IF NOT EXISTS tags TEXT DEFAULT '';
	ALTER TABLE dishes ADD COLUMN IF NOT EXISTS image_url TEXT DEFAULT '';
	`
	_, err := a.DB.Exec(schema)
	return err
}

// SeedDefaultRestaurant resets and re-seeds "mamma-mia-mario" and "dogao_do_seu_joao" with menu items.
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

// RegisterRoutes registers all routes for both "/" and "/api/" prefixes for maximum flexibility.
func (a *App) RegisterRoutes(mux *http.ServeMux) {
	prefixes := []string{"", "/api"}

	// Static File Server: serve public uploads from ./media/ directory
	_ = os.MkdirAll("media/dishes", 0755)
	mux.Handle("GET /media/", http.StripPrefix("/media/", http.FileServer(http.Dir("./media"))))

	for _, p := range prefixes {
		// Restaurants
		mux.HandleFunc(fmt.Sprintf("GET %s/restaurants", p), a.handleListRestaurants)
		mux.HandleFunc(fmt.Sprintf("POST %s/restaurants", p), a.handleCreateRestaurant)
		mux.HandleFunc(fmt.Sprintf("GET %s/restaurants/{slug}", p), a.handleGetRestaurant)
		mux.HandleFunc(fmt.Sprintf("PUT %s/restaurants/{slug}", p), a.handleUpdateRestaurant)
		mux.HandleFunc(fmt.Sprintf("DELETE %s/restaurants/{slug}", p), a.handleDeleteRestaurant)

		// Dishes
		mux.HandleFunc(fmt.Sprintf("GET %s/restaurants/{slug}/dishes", p), a.handleListDishes)
		mux.HandleFunc(fmt.Sprintf("POST %s/restaurants/{slug}/dishes", p), a.handleCreateDish)
		mux.HandleFunc(fmt.Sprintf("PUT %s/restaurants/{slug}/dishes/{id}", p), a.handleUpdateDish)
		mux.HandleFunc(fmt.Sprintf("DELETE %s/restaurants/{slug}/dishes/{id}", p), a.handleDeleteDish)

		// Full Menu / Page Endpoint (returns { restaurant, categories, dishes })
		mux.HandleFunc(fmt.Sprintf("GET %s/restaurants/{slug}/menu", p), a.handleRestaurantMenu)
		mux.HandleFunc(fmt.Sprintf("GET %s/r/{slug}", p), a.handleRestaurantMenu)

		// Images / Uploads
		mux.HandleFunc(fmt.Sprintf("POST %s/restaurants/{slug}/dishes/{id}/image", p), a.handleUploadDishImage)
		mux.HandleFunc(fmt.Sprintf("POST %s/upload/image", p), a.handleGenericImageUpload)
	}
}

// ---------------- RESTAURANTS HANDLERS ----------------

// handleListRestaurants: Direct read query to persistence layer (simple CRUD bypass).
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

// handleCreateRestaurant: Routes through Layer 2 Business Service for validation & slug logic.
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

// handleGetRestaurant: Direct read query to persistence layer.
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

// handleRestaurantMenu: Returns the consolidated profile, categories, and dishes for the restaurant page.
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

// handleCreateDish: Routes through Layer 2 Business Service for price and restaurant validation.
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
