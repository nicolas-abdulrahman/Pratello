package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"

	_ "github.com/lib/pq"
)

type Restaurant struct {
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

var defaultApp = NewApp(nil)



func setupRouter(apps ...*App) *http.ServeMux {
	var app *App
	if len(apps) > 0 && apps[0] != nil {
		app = apps[0]
	} else {
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

func migrateSchema(db *sql.DB) error {
	if db == nil {
		return nil
	}
	queries := []string{
		`ALTER TABLE "account" ADD COLUMN IF NOT EXISTS "name" VARCHAR(500);`,
		`ALTER TABLE "account" ADD COLUMN IF NOT EXISTS "created_at" TIMESTAMPTZ NOT NULL DEFAULT NOW();`,
		`ALTER TABLE "physical_restaurant" ADD COLUMN IF NOT EXISTS "name" VARCHAR(500) NOT NULL DEFAULT '';`,
		`ALTER TABLE "physical_restaurant" ADD COLUMN IF NOT EXISTS "owner_cnpj" VARCHAR(500) NOT NULL DEFAULT '';`,
		`ALTER TABLE "physical_restaurant" ADD COLUMN IF NOT EXISTS "terminal_pin" VARCHAR(6) NOT NULL DEFAULT '123456';`,
		`ALTER TABLE "physical_restaurant" ADD COLUMN IF NOT EXISTS "pin_updated_at" TIMESTAMPTZ NOT NULL DEFAULT NOW();`,
		`ALTER TABLE "physical_restaurant" ADD COLUMN IF NOT EXISTS "failed_pin_attempts" INT NOT NULL DEFAULT 0;`,
		`ALTER TABLE "physical_restaurant" ADD COLUMN IF NOT EXISTS "locked_until" TIMESTAMPTZ;`,
	}
	for _, q := range queries {
		if _, err := db.Exec(q); err != nil {
			log.Printf("Migration notice: query %q failed: %v", q, err)
		}
	}
	return nil
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

	if db != nil {
		if err := migrateSchema(db); err != nil {
			log.Printf("Warning: schema migration had errors: %v\n", err)
		}
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
