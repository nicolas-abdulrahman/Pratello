package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type Restaurant struct {
	Slug        string   `json:"slug"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Cuisine     string   `json:"cuisine"`
	MenuItems   []string `json:"menu_items"`
}

// Sample in-memory store for restaurants
var sampleRestaurants = map[string]Restaurant{
	"bobpizzeria": {
		Slug:        "bobpizzeria",
		Name:        "Bob's Pizzeria",
		Description: "The crispiest wood-fired artisanal pizza in town.",
		Cuisine:     "Italian / Pizza",
		MenuItems:   []string{"Margherita DOP", "Spicy Pepperoni & Hot Honey", "Truffle Mushroom", "Garlic Focaccia"},
	},
	"deliciadeprazeres": {
		Slug:        "deliciadeprazeres",
		Name:        "Delícia de Prazeres",
		Description: "Authentic comfort food, sweets, and savory specialties.",
		Cuisine:     "Bakery & Delicatessen",
		MenuItems:   []string{"Pastel de Nata", "Coxinha Gourmet", "Pão de Queijo Recheado", "Bolo de Cenoura com Chocolate"},
	},
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	mux := http.NewServeMux()

	// 1. Health check
	mux.HandleFunc("GET /api/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok", "service": "pratello"})
	})

	// 2. Restaurant details API endpoint using Go 1.22 path matching
	mux.HandleFunc("GET /api/restaurants/{restaurant}", func(w http.ResponseWriter, r *http.Request) {
		slug := strings.ToLower(r.PathValue("restaurant"))
		restaurant, exists := sampleRestaurants[slug]
		if !exists {
			// If not found in sample db, generate a placeholder
			restaurant = Restaurant{
				Slug:        slug,
				Name:        strings.Title(slug),
				Description: fmt.Sprintf("Welcome to %s on Pratello!", strings.Title(slug)),
				Cuisine:     "Restaurant & Bar",
				MenuItems:   []string{"Special of the Day", "Chef's Tasting Menu", "Seasonal Dessert"},
			}
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(restaurant)
	})

	// 3. Static frontend file server (Vite build output in frontend/dist)
	frontendDist := filepath.Join("frontend", "dist")
	if _, err := os.Stat(frontendDist); os.IsNotExist(err) {
		if _, err := os.Stat(filepath.Join("..", "frontend", "dist")); err == nil {
			frontendDist = filepath.Join("..", "frontend", "dist")
		}
	}
	setupFrontendRoutes(mux, frontendDist)

	addr := ":" + port
	log.Printf("Pratello server running at http://localhost%s\n", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

// setupFrontendRoutes serves static files or returns an informative message if frontend isn't built yet
func setupFrontendRoutes(mux *http.ServeMux, distDir string) {
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		// Do not intercept /api routes
		if strings.HasPrefix(r.URL.Path, "/api/") {
			http.NotFound(w, r)
			return
		}

		// Check if frontend/dist exists
		if _, err := os.Stat(distDir); os.IsNotExist(err) {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			fmt.Fprintf(w, `
<!DOCTYPE html>
<html>
<head><title>Pratello - Setup Required</title><script src="https://cdn.tailwindcss.com"></script></head>
<body class="bg-slate-900 text-white min-h-screen flex items-center justify-center p-6">
  <div class="max-w-md w-full bg-slate-800 rounded-xl p-8 shadow-xl border border-slate-700">
    <h1 class="text-2xl font-bold text-amber-400 mb-2">🍽️ Pratello Backend Ready!</h1>
    <p class="text-slate-300 mb-4">Go server is up on port 8080, but the Svelte frontend hasn't been built yet.</p>
    <div class="bg-slate-950 p-4 rounded text-sm font-mono text-emerald-400 mb-4">
      cd frontend<br/>
      npm install<br/>
      npm run build
    </div>
    <p class="text-xs text-slate-400">Or run <span class="text-emerald-300">npm run dev</span> inside <code>frontend/</code> to run the Vite dev server with HMR.</p>
  </div>
</body>
</html>`)
			return
		}

		// Try to serve requested static file (e.g. /assets/index.js)
		requestedPath := filepath.Clean(filepath.Join(distDir, r.URL.Path))
		info, err := os.Stat(requestedPath)
		if err == nil && !info.IsDir() {
			http.ServeFile(w, r, requestedPath)
			return
		}

		// Fallback to index.html for SPA routing (e.g. /bobpizzeria or /deliciadeprazeres)
		http.ServeFile(w, r, filepath.Join(distDir, "index.html"))
	})
}
