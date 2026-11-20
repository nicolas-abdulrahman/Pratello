package main

import (
	"encoding/json"
	"net/http"

	"pratello/server/internal/auth"
	"pratello/server/internal/restaurants"
)

// registerRoutes mounts all domain routes and static file handlers onto mux, wrapping with CORS middleware.
func registerRoutes(mux *http.ServeMux, authHandler *auth.Handler, restHandler *restaurants.Handler) http.Handler {
	// Health endpoints
	handleHealth := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok", "service": "pratello"})
	}
	mux.HandleFunc("GET /health", handleHealth)
	mux.HandleFunc("GET /api/health", handleHealth)

	// Serve local assets referenced by catalog photo URLs.
	mux.Handle("GET /media/", http.StripPrefix("/media/", http.FileServer(http.Dir("./media"))))

	// Mount domain routes
	if authHandler != nil {
		authHandler.Routes(mux)
	}
	if restHandler != nil {
		restHandler.Routes(mux)
	}

	// Wrap in CORS middleware
	return corsMiddleware(mux)
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Device-ID, Idempotency-Key, X-Order-Token")
		w.Header().Set("Access-Control-Expose-Headers", "X-Order-Token")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
