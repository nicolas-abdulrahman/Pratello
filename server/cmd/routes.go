package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

// Dish represents a menu item belonging to a restaurant.
type Dish struct {
	ID             int     `json:"id"`
	RestaurantSlug string  `json:"-"`
	Name           string  `json:"name"`
	Price          float64 `json:"price"`
}

type CreateDishRequest struct {
	Name  string  `json:"name"`
	Price float64 `json:"price"`
}

type UpdateDishRequest struct {
	Name  string  `json:"name"`
	Price float64 `json:"price"`
}

type UpdateRestaurantRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// App holds the database connection and repository stores.
type App struct {
	Restaurants RestaurantStore
	Dishes      DishStore
	DB          *sql.DB
}

func NewApp(db *sql.DB) *App {
	app := &App{DB: db}
	if db != nil {
		app.Restaurants = NewSQLRestaurantStore(db)
		app.Dishes = NewSQLDishStore(db)
		_ = app.InitSchema()
	} else {
		app.Restaurants = NewMemoryRestaurantStore()
		app.Dishes = NewMemoryDishStore()
	}
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
		description TEXT
	);

	CREATE TABLE IF NOT EXISTS dishes (
		id SERIAL PRIMARY KEY,
		restaurant_slug TEXT NOT NULL REFERENCES restaurants(slug) ON DELETE CASCADE,
		name TEXT NOT NULL,
		price NUMERIC(10, 2) NOT NULL
	);
	`
	_, err := a.DB.Exec(schema)
	return err
}

// SeedDefaultRestaurant resets and re-seeds dogao_do_seu_joao with 1 dish "feijao"
func (a *App) SeedDefaultRestaurant() error {
	ctx := context.Background()
	slug := "dogao_do_seu_joao"
	name := "Dogão do Seu João"
	desc := "Aqui é o dogão do seu joão"

	_ = a.Restaurants.Delete(ctx, slug)
	if err := a.Restaurants.Create(ctx, &Restaurant{
		Slug:        slug,
		Name:        name,
		Description: desc,
	}); err != nil {
		return fmt.Errorf("failed to insert default restaurant: %w", err)
	}

	if _, err := a.Dishes.Create(ctx, slug, CreateDishRequest{
		Name:  "feijao",
		Price: 10.00,
	}); err != nil {
		return fmt.Errorf("failed to insert default dish: %w", err)
	}

	return nil
}

// RegisterRoutes registers all routes for both "/" and "/api/" prefixes for maximum flexibility.
func (a *App) RegisterRoutes(mux *http.ServeMux) {
	// prefixes := []string{, "/api/r"}
	var p=  "/api/r"

	// Restaurants
	mux.HandleFunc(fmt.Sprintf("GET %s/", p), a.handleListRestaurants)
	mux.HandleFunc(fmt.Sprintf("POST %s/", p), a.handleCreateRestaurant)
	mux.HandleFunc(fmt.Sprintf("GET %s/{slug}", p), a.handleGetRestaurant)
	mux.HandleFunc(fmt.Sprintf("PUT %s/{slug}", p), a.handleUpdateRestaurant)
	mux.HandleFunc(fmt.Sprintf("DELETE %s/{slug}", p), a.handleDeleteRestaurant)

	// Dishes
	mux.HandleFunc(fmt.Sprintf("GET %s/{slug}/dishes", p), a.handleListDishes)
	mux.HandleFunc(fmt.Sprintf("POST %s/{slug}/dishes", p), a.handleCreateDish)
	mux.HandleFunc(fmt.Sprintf("PUT %s/{slug}/dishes/{id}", p), a.handleUpdateDish)
	mux.HandleFunc(fmt.Sprintf("DELETE %s/{slug}/dishes/{id}", p), a.handleDeleteDish)
}

// ---------------- RESTAURANTS HANDLERS ----------------

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

func (a *App) handleCreateRestaurant(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req CreateRestaurantRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid request body"})
		return
	}

	if strings.TrimSpace(req.Name) == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "restaurant name is required"})
		return
	}

	slug := strings.TrimSpace(req.Slug)
	if slug == "" {
		slug = strings.ToLower(strings.ReplaceAll(req.Name, " ", "-"))
	} else {
		slug = strings.ToLower(slug)
	}

	restaurant := Restaurant{
		Slug:        slug,
		Name:        req.Name,
		Description: req.Description,
	}

	if err := a.Restaurants.Create(r.Context(), &restaurant); err != nil {
		if errors.Is(err, ErrConflict) {
			w.WriteHeader(http.StatusConflict)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "restaurant with this slug already exists"})
			return
		}
		http.Error(w, `{"error":"failed to create restaurant"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(restaurant)
}

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

	var req UpdateRestaurantRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid request body"})
		return
	}

	updated, err := a.Restaurants.Update(r.Context(), slug, req)
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

func (a *App) handleCreateDish(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	slug := strings.ToLower(r.PathValue("slug"))

	var req CreateDishRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid request body"})
		return
	}

	if strings.TrimSpace(req.Name) == "" || req.Price < 0 {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "valid name and non-negative price are required"})
		return
	}

	dish, err := a.Dishes.Create(r.Context(), slug, req)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "restaurant not found"})
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

	var req UpdateDishRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid request body"})
		return
	}

	dish, err := a.Dishes.Update(r.Context(), slug, id, req)
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
