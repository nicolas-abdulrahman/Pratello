//go:build integration

package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	_ "github.com/lib/pq"
)

// TestPostgresAddAndReadRestaurant verifies adding a restaurant and reading its data
// against the real PostgreSQL database.
func TestPostgresAddAndReadRestaurant(t *testing.T) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://nick@localhost:5432/pratello?sslmode=disable"
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		t.Fatalf("Failed to open postgres connection: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		t.Skipf("⚠️ PostgreSQL is not reachable at %s (skipping integration test): %v", dbURL, err)
	}

	app := NewApp(db)
	if err := app.InitSchema(); err != nil {
		t.Fatalf("Failed to initialize database schema: %v", err)
	}

	router := setupRouter(app)

	// Clean up any leftover from prior runs
	slug := "bela-napoli-integration"
	_, _ = db.Exec("DELETE FROM restaurants WHERE slug = $1", slug)

	// 1. Add Restaurant (POST /restaurants)
	payload, _ := json.Marshal(map[string]string{
		"name":        "Bela Napoli Integration",
		"slug":        slug,
		"description": "Authentic wood-fired pizza",
	})
	postReq := httptest.NewRequest(http.MethodPost, "/restaurants", bytes.NewReader(payload))
	postReq.Header.Set("Content-Type", "application/json")
	postRec := httptest.NewRecorder()
	router.ServeHTTP(postRec, postReq)

	if postRec.Code != http.StatusCreated {
		t.Fatalf("Expected 201 Created for adding restaurant, got %d. Body: %s", postRec.Code, postRec.Body.String())
	}

	var created Restaurant
	if err := json.NewDecoder(postRec.Body).Decode(&created); err != nil {
		t.Fatalf("Failed to decode POST response: %v", err)
	}
	if created.Slug != slug || created.Name != "Bela Napoli Integration" {
		t.Errorf("Unexpected created restaurant data: %+v", created)
	}

	// 2. Read Restaurant Data (GET /restaurants/{slug})
	getReq := httptest.NewRequest(http.MethodGet, "/restaurants/"+slug, nil)
	getRec := httptest.NewRecorder()
	router.ServeHTTP(getRec, getReq)

	if getRec.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK for reading restaurant data, got %d. Body: %s", getRec.Code, getRec.Body.String())
	}

	var fetched Restaurant
	if err := json.NewDecoder(getRec.Body).Decode(&fetched); err != nil {
		t.Fatalf("Failed to decode GET response: %v", err)
	}
	if fetched.Slug != slug || fetched.Name != "Bela Napoli Integration" || fetched.Description != "Authentic wood-fired pizza" {
		t.Errorf("Unexpected fetched restaurant data from Postgres: %+v", fetched)
	}

	// Clean up after test
	_, _ = db.Exec("DELETE FROM restaurants WHERE slug = $1", slug)
}
