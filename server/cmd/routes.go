package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
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

// App holds the stores for business logic and legacy fallback fields for backwards compatibility.
type App struct {
	Restaurants    RestaurantStore
	Dishes         DishStore
	DB             *sql.DB
	memMu          sync.RWMutex
	memRestaurants map[string]Restaurant
	memDishes      map[int]Dish
	dishIDCounter  int
}

func NewApp(db *sql.DB) *App {
	app := &App{
		DB:             db,
		memRestaurants: make(map[string]Restaurant),
		memDishes:      make(map[int]Dish),
	}
	if db != nil {
		app.Restaurants = NewSQLRestaurantStore(db)
		app.Dishes = NewSQLDishStore(db)
		_ = app.InitSchema()
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
	slug := "dogao_do_seu_joao"
	name := "Dogão do Seu João"
	desc := "Aqui é o dogão do seu joão"

	if a.DB != nil {
		// Delete if exists (cascades to dishes)
		_, _ = a.DB.Exec("DELETE FROM restaurants WHERE slug = $1", slug)

		// Insert restaurant
		_, err := a.DB.Exec("INSERT INTO restaurants (slug, name, description) VALUES ($1, $2, $3)", slug, name, desc)
		if err != nil {
			return fmt.Errorf("failed to insert default restaurant: %w", err)
		}

		// Insert initial dish "feijao"
		_, err = a.DB.Exec("INSERT INTO dishes (restaurant_slug, name, price) VALUES ($1, $2, $3)", slug, "feijao", 10.00)
		if err != nil {
			return fmt.Errorf("failed to insert default dish: %w", err)
		}
		return nil
	}

	a.memMu.Lock()
	defer a.memMu.Unlock()

	// Clear if existing
	delete(a.memRestaurants, slug)
	for id, d := range a.memDishes {
		if d.RestaurantSlug == slug {
			delete(a.memDishes, id)
		}
	}

	// Insert restaurant
	a.memRestaurants[slug] = Restaurant{
		Slug:        slug,
		Name:        name,
		Description: desc,
	}

	// Insert dish "feijao"
	a.dishIDCounter++
	a.memDishes[a.dishIDCounter] = Dish{
		ID:             a.dishIDCounter,
		RestaurantSlug: slug,
		Name:           "feijao",
		Price:          10.00,
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

func (a *App) handleListRestaurants(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if a.Restaurants != nil {
		list, err := a.Restaurants.List(r.Context())
		if err != nil {
			http.Error(w, `{"error":"database error"}`, http.StatusInternalServerError)
			return
		}
		if list == nil {
			list = make([]Restaurant, 0)
		}
		_ = json.NewEncoder(w).Encode(list)
		return
	}

	if a.DB != nil {
		rows, err := a.DB.QueryContext(r.Context(), "SELECT slug, name, COALESCE(description, '') FROM restaurants")
		if err != nil {
			http.Error(w, `{"error":"database error"}`, http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		list := make([]Restaurant, 0)
		for rows.Next() {
			var res Restaurant
			if err := rows.Scan(&res.Slug, &res.Name, &res.Description); err == nil {
				list = append(list, res)
			}
		}
		_ = json.NewEncoder(w).Encode(list)
		return
	}

	a.memMu.RLock()
	defer a.memMu.RUnlock()
	list := make([]Restaurant, 0, len(a.memRestaurants))
	for _, res := range a.memRestaurants {
		list = append(list, res)
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

	if a.Restaurants != nil {
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
		return
	}

	if a.DB != nil {
		var existing string
		err := a.DB.QueryRowContext(r.Context(), "SELECT slug FROM restaurants WHERE slug = $1", slug).Scan(&existing)
		if err == nil {
			w.WriteHeader(http.StatusConflict)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "restaurant with this slug already exists"})
			return
		}

		_, err = a.DB.ExecContext(r.Context(),
			"INSERT INTO restaurants (slug, name, description) VALUES ($1, $2, $3)",
			restaurant.Slug, restaurant.Name, restaurant.Description,
		)
		if err != nil {
			http.Error(w, `{"error":"failed to create restaurant"}`, http.StatusInternalServerError)
			return
		}
	} else {
		a.memMu.Lock()
		defer a.memMu.Unlock()
		if _, exists := a.memRestaurants[slug]; exists {
			w.WriteHeader(http.StatusConflict)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "restaurant with this slug already exists"})
			return
		}
		a.memRestaurants[slug] = restaurant
	}

	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(restaurant)
}

func (a *App) handleGetRestaurant(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	slug := strings.ToLower(r.PathValue("slug"))

	if a.Restaurants != nil {
		res, err := a.Restaurants.GetBySlug(r.Context(), slug)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "restaurant not found"})
			return
		}
		_ = json.NewEncoder(w).Encode(res)
		return
	}

	if a.DB != nil {
		var res Restaurant
		err := a.DB.QueryRowContext(r.Context(),
			"SELECT slug, name, COALESCE(description, '') FROM restaurants WHERE slug = $1", slug,
		).Scan(&res.Slug, &res.Name, &res.Description)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "restaurant not found"})
			return
		}
		_ = json.NewEncoder(w).Encode(res)
		return
	}

	a.memMu.RLock()
	defer a.memMu.RUnlock()
	res, exists := a.memRestaurants[slug]
	if !exists {
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

	if a.Restaurants != nil {
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
		return
	}

	if a.DB != nil {
		res, err := a.DB.ExecContext(r.Context(),
			"UPDATE restaurants SET name = $1, description = $2 WHERE slug = $3",
			req.Name, req.Description, slug,
		)
		if err != nil {
			http.Error(w, `{"error":"database error"}`, http.StatusInternalServerError)
			return
		}
		rows, _ := res.RowsAffected()
		if rows == 0 {
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "restaurant not found"})
			return
		}

		_ = json.NewEncoder(w).Encode(Restaurant{Slug: slug, Name: req.Name, Description: req.Description})
		return
	}

	a.memMu.Lock()
	defer a.memMu.Unlock()
	res, exists := a.memRestaurants[slug]
	if !exists {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "restaurant not found"})
		return
	}
	if req.Name != "" {
		res.Name = req.Name
	}
	res.Description = req.Description
	a.memRestaurants[slug] = res

	_ = json.NewEncoder(w).Encode(res)
}

func (a *App) handleDeleteRestaurant(w http.ResponseWriter, r *http.Request) {
	slug := strings.ToLower(r.PathValue("slug"))

	if a.Restaurants != nil {
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
		return
	}

	if a.DB != nil {
		res, err := a.DB.ExecContext(r.Context(), "DELETE FROM restaurants WHERE slug = $1", slug)
		if err != nil {
			http.Error(w, `{"error":"database error"}`, http.StatusInternalServerError)
			return
		}
		rows, _ := res.RowsAffected()
		if rows == 0 {
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "restaurant not found"})
			return
		}
		w.WriteHeader(http.StatusNoContent)
		return
	}

	a.memMu.Lock()
	defer a.memMu.Unlock()
	if _, exists := a.memRestaurants[slug]; !exists {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "restaurant not found"})
		return
	}

	delete(a.memRestaurants, slug)
	for id, dish := range a.memDishes {
		if dish.RestaurantSlug == slug {
			delete(a.memDishes, id)
		}
	}
	w.WriteHeader(http.StatusNoContent)
}

// ---------------- DISHES HANDLERS ----------------

func (a *App) handleListDishes(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	slug := strings.ToLower(r.PathValue("slug"))

	if a.Dishes != nil {
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
		return
	}

	if a.DB != nil {
		// Verify restaurant exists
		var dummy string
		if err := a.DB.QueryRowContext(r.Context(), "SELECT slug FROM restaurants WHERE slug = $1", slug).Scan(&dummy); err != nil {
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "restaurant not found"})
			return
		}

		rows, err := a.DB.QueryContext(r.Context(),
			"SELECT id, name, price FROM dishes WHERE restaurant_slug = $1 ORDER BY id ASC", slug,
		)
		if err != nil {
			http.Error(w, `{"error":"database error"}`, http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		dishes := make([]Dish, 0)
		for rows.Next() {
			var d Dish
			if err := rows.Scan(&d.ID, &d.Name, &d.Price); err == nil {
				d.RestaurantSlug = slug
				dishes = append(dishes, d)
			}
		}
		_ = json.NewEncoder(w).Encode(dishes)
		return
	}

	a.memMu.RLock()
	defer a.memMu.RUnlock()
	if _, exists := a.memRestaurants[slug]; !exists {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "restaurant not found"})
		return
	}

	dishes := make([]Dish, 0)
	for _, d := range a.memDishes {
		if d.RestaurantSlug == slug {
			dishes = append(dishes, d)
		}
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

	if a.Dishes != nil {
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
		return
	}

	if a.DB != nil {
		var dummy string
		if err := a.DB.QueryRowContext(r.Context(), "SELECT slug FROM restaurants WHERE slug = $1", slug).Scan(&dummy); err != nil {
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "restaurant not found"})
			return
		}

		var newID int
		err := a.DB.QueryRowContext(r.Context(),
			"INSERT INTO dishes (restaurant_slug, name, price) VALUES ($1, $2, $3) RETURNING id",
			slug, req.Name, req.Price,
		).Scan(&newID)
		if err != nil {
			http.Error(w, `{"error":"failed to create dish"}`, http.StatusInternalServerError)
			return
		}

		dish := Dish{ID: newID, RestaurantSlug: slug, Name: req.Name, Price: req.Price}
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(dish)
		return
	}

	a.memMu.Lock()
	defer a.memMu.Unlock()
	if _, exists := a.memRestaurants[slug]; !exists {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "restaurant not found"})
		return
	}

	a.dishIDCounter++
	dish := Dish{
		ID:             a.dishIDCounter,
		RestaurantSlug: slug,
		Name:           req.Name,
		Price:          req.Price,
	}
	a.memDishes[dish.ID] = dish

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

	if a.Dishes != nil {
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
		return
	}

	if a.DB != nil {
		res, err := a.DB.ExecContext(r.Context(),
			"UPDATE dishes SET name = $1, price = $2 WHERE id = $3 AND restaurant_slug = $4",
			req.Name, req.Price, id, slug,
		)
		if err != nil {
			http.Error(w, `{"error":"database error"}`, http.StatusInternalServerError)
			return
		}
		rows, _ := res.RowsAffected()
		if rows == 0 {
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "dish not found"})
			return
		}

		_ = json.NewEncoder(w).Encode(Dish{ID: id, RestaurantSlug: slug, Name: req.Name, Price: req.Price})
		return
	}

	a.memMu.Lock()
	defer a.memMu.Unlock()
	dish, exists := a.memDishes[id]
	if !exists || dish.RestaurantSlug != slug {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "dish not found"})
		return
	}

	if req.Name != "" {
		dish.Name = req.Name
	}
	if req.Price >= 0 {
		dish.Price = req.Price
	}
	a.memDishes[id] = dish

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

	if a.Dishes != nil {
		err := a.Dishes.Delete(r.Context(), slug, id)
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
		return
	}

	if a.DB != nil {
		res, err := a.DB.ExecContext(r.Context(),
			"DELETE FROM dishes WHERE id = $1 AND restaurant_slug = $2", id, slug,
		)
		if err != nil {
			http.Error(w, `{"error":"database error"}`, http.StatusInternalServerError)
			return
		}
		rows, _ := res.RowsAffected()
		if rows == 0 {
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "dish not found"})
			return
		}
		w.WriteHeader(http.StatusNoContent)
		return
	}

	a.memMu.Lock()
	defer a.memMu.Unlock()
	dish, exists := a.memDishes[id]
	if !exists || dish.RestaurantSlug != slug {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "dish not found"})
		return
	}

	delete(a.memDishes, id)
	w.WriteHeader(http.StatusNoContent)
}
