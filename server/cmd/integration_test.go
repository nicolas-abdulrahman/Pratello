//go:build integration

package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
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

	// 3. Add Dish for the restaurant
	dishPayload, _ := json.Marshal(map[string]any{
		"name":  "Pizza Margherita",
		"price": 45.00,
	})
	dishPostReq := httptest.NewRequest(http.MethodPost, "/restaurants/"+slug+"/dishes", bytes.NewReader(dishPayload))
	dishPostReq.Header.Set("Content-Type", "application/json")
	dishPostRec := httptest.NewRecorder()
	router.ServeHTTP(dishPostRec, dishPostReq)

	if dishPostRec.Code != http.StatusCreated {
		t.Fatalf("Expected 201 Created for adding dish, got %d. Body: %s", dishPostRec.Code, dishPostRec.Body.String())
	}
	var createdDish Dish
	if err := json.NewDecoder(dishPostRec.Body).Decode(&createdDish); err != nil {
		t.Fatalf("Failed to decode dish response: %v", err)
	}

	// 4. Upload Image for Dish (POST /restaurants/{slug}/dishes/{id}/image)
	var bodyBuf bytes.Buffer
	mw := multipart.NewWriter(&bodyBuf)
	part, err := mw.CreateFormFile("image", "margherita.png")
	if err != nil {
		t.Fatalf("Failed to create multipart form file: %v", err)
	}
	fakeImageData := []byte("\x89PNG\r\n\x1a\nfake-png-binary-content")
	_, _ = part.Write(fakeImageData)
	_ = mw.Close()

	uploadReq := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/restaurants/%s/dishes/%v/image", slug, createdDish.ID), &bodyBuf)
	uploadReq.Header.Set("Content-Type", mw.FormDataContentType())
	uploadRec := httptest.NewRecorder()
	router.ServeHTTP(uploadRec, uploadReq)

	if uploadRec.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK for image upload, got %d. Body: %s", uploadRec.Code, uploadRec.Body.String())
	}
	var dishWithImage Dish
	if err := json.NewDecoder(uploadRec.Body).Decode(&dishWithImage); err != nil {
		t.Fatalf("Failed to decode uploaded dish response: %v", err)
	}
	if !strings.HasPrefix(dishWithImage.ImageURL, "/media/dishes/") {
		t.Fatalf("Expected image_url starting with /media/dishes/, got: %s", dishWithImage.ImageURL)
	}

	// 5. Fetch image statically (GET /media/dishes/...)
	staticReq := httptest.NewRequest(http.MethodGet, dishWithImage.ImageURL, nil)
	staticRec := httptest.NewRecorder()
	router.ServeHTTP(staticRec, staticReq)

	if staticRec.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK for statically served image, got %d", staticRec.Code)
	}
	if !bytes.Equal(staticRec.Body.Bytes(), fakeImageData) {
		t.Fatalf("Static file content mismatch: expected %q, got %q", fakeImageData, staticRec.Body.Bytes())
	}

	// 6. Test Restaurant Menu / Page Endpoint (GET /r/{slug})
	menuReq := httptest.NewRequest(http.MethodGet, "/r/"+slug, nil)
	menuRec := httptest.NewRecorder()
	router.ServeHTTP(menuRec, menuReq)

	if menuRec.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK for GET /r/%s, got %d. Body: %s", slug, menuRec.Code, menuRec.Body.String())
	}
	var menuData struct {
		Restaurant Restaurant `json:"restaurant"`
		Categories []string   `json:"categories"`
		Dishes     []Dish     `json:"dishes"`
	}
	if err := json.NewDecoder(menuRec.Body).Decode(&menuData); err != nil {
		t.Fatalf("Failed to decode /r/%s response: %v", slug, err)
	}
	if menuData.Restaurant.Slug != slug {
		t.Errorf("Expected restaurant slug %s, got %s", slug, menuData.Restaurant.Slug)
	}
	if len(menuData.Categories) == 0 || menuData.Categories[0] != "Mais Pedidos" {
		t.Errorf("Expected categories to include 'Mais Pedidos', got %v", menuData.Categories)
	}
	if len(menuData.Dishes) == 0 {
		t.Errorf("Expected at least 1 dish in menu response, got %d", len(menuData.Dishes))
	}

	// Clean up after test
	_ = os.Remove("." + dishWithImage.ImageURL)
	_, _ = db.Exec("DELETE FROM restaurants WHERE slug = $1", slug)
}
