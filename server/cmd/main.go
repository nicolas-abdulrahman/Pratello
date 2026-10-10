package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"sync"

	_ "github.com/lib/pq"
)

type Restaurant struct {
	Slug        string   `json:"slug"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Cuisine     string   `json:"cuisine,omitempty"`
	MenuItems   []string `json:"menu_items,omitempty"`
}

// In-memory store for restaurants protected by a mutex (for backwards compatibility)
var (
	restaurantsMu sync.RWMutex
	restaurants   = make(map[string]Restaurant)
	defaultApp    *App
)

func init() {
	defaultApp = NewApp(nil)
	defaultApp.memRestaurants = restaurants
}

type CreateRestaurantRequest struct {
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
}

func setupRouter(apps ...*App) *http.ServeMux {
	var app *App
	if len(apps) > 0 && apps[0] != nil {
		app = apps[0]
	} else {
		// Sync with global restaurants map in case tests modified it directly
		restaurantsMu.RLock()
		defaultApp.memRestaurants = restaurants
		restaurantsMu.RUnlock()
		app = defaultApp
	}

	mux := http.NewServeMux()

	// Health check
	handleHealth := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok", "service": "pratello"})
	}
	mux.HandleFunc("GET /health", handleHealth)
	mux.HandleFunc("GET /api/health", handleHealth)

	// Register all REST routes from routes.go
	app.RegisterRoutes(mux)

	return mux
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	dbURL := os.Getenv("DATABASE_URL")
	var db *sql.DB
	if dbURL != "" {
		var err error
		db, err = sql.Open("postgres", dbURL)
		if err != nil {
			log.Printf("Warning: failed to open postgres connection: %v\n", err)
		} else if err := db.Ping(); err != nil {
			log.Printf("Warning: failed to ping postgres: %v (continuing with in-memory fallback)\n", err)
			db = nil
		} else {
			log.Println("Connected to PostgreSQL successfully")
		}
	} else {
		log.Println("DATABASE_URL not set; running with flexible in-memory store")
	}

	app := NewApp(db)
	if err := app.SeedDefaultRestaurant(); err != nil {
		log.Printf("Warning: failed to seed default restaurant: %v\n", err)
	} else {
		log.Println("Seeded default restaurant 'dogao_do_seu_joao' with dish 'feijao'")
	}

	mux := setupRouter(app)

	addr := ":" + port
	log.Printf("Pratello server running at http://localhost%s\n", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
