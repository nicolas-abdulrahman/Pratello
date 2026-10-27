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

	if _, err := a.Dishes.Create(ctx, NewDish{
		RestaurantSlug: slug,
		Name:           "feijao",
		Price:          10.00,
	}); err != nil {
		return fmt.Errorf("failed to insert default dish: %w", err)
	}

	return nil
}

// RegisterRoutes registers all routes for both "/" and "/api/" prefixes for maximum flexibility.
func (a *App) RegisterRoutes(mux *http.ServeMux) {
	prefixes := []string{"", "/api"}

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
		Name  string  `json:"name"`
		Price float64 `json:"price"`
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
		Name  string  `json:"name"`
		Price float64 `json:"price"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid request body"})
		return
	}

	input := UpdateDishInput{
		Name:  body.Name,
		Price: body.Price,
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
