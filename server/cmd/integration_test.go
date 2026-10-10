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

// TestPostgresSmokeLifecycle runs a fast end-to-end smoke test against a real PostgreSQL database:
// 1. Connects to PostgreSQL and verifies schema
// 2. Creates a Restaurant
// 3. Adds a Dish (testing foreign key association)
// 4. Verifies the Dish was stored
// 5. Deletes the Restaurant (testing ON DELETE CASCADE)
// 6. Confirms clean deletion
func TestPostgresSmokeLifecycle(t *testing.T) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable"
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		t.Fatalf("Failed to open postgres connection: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		t.Skipf("⚠️ PostgreSQL is not reachable at %s (skipping integration test).\n👉 Start PostgreSQL or run via Docker to execute.\nError: %v", dbURL, err)
	}

	app := NewApp(db)
	if err := app.InitSchema(); err != nil {
		t.Fatalf("Failed to initialize database schema: %v", err)
	}

	router := setupRouter(app)

	// Clean up any leftovers from prior runs
	_, _ = db.Exec("DELETE FROM restaurants WHERE slug = 'smoke-burger'")

	// 1. Create a restaurant (POST /restaurants)
	restPayload, _ := json.Marshal(Restaurant{
		Slug:        "smoke-burger",
		Name:        "Smokehouse Burger",
		Description: "Craft burgers with real smoke flavor",
	})
	req := httptest.NewRequest(http.MethodPost, "/restaurants", bytes.NewReader(restPayload))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("Expected 201 Created for restaurant, got %d. Body: %s", rec.Code, rec.Body.String())
	}

	// 2. Add a dish (POST /restaurants/{slug}/dishes)
	dishPayload, _ := json.Marshal(CreateDishRequest{
		Name:  "Smoked Bacon Cheeseburger",
		Price: 14.50,
	})
	dishReq := httptest.NewRequest(http.MethodPost, "/restaurants/smoke-burger/dishes", bytes.NewReader(dishPayload))
	dishReq.Header.Set("Content-Type", "application/json")
	dishRec := httptest.NewRecorder()
	router.ServeHTTP(dishRec, dishReq)

	if dishRec.Code != http.StatusCreated {
		t.Fatalf("Expected 201 Created for dish, got %d. Body: %s", dishRec.Code, dishRec.Body.String())
	}

	var createdDish Dish
	if err := json.NewDecoder(dishRec.Body).Decode(&createdDish); err != nil {
		t.Fatalf("Failed to parse created dish: %v", err)
	}

	// 3. Verify dish is queryable (GET /restaurants/{slug}/dishes)
	listReq := httptest.NewRequest(http.MethodGet, "/restaurants/smoke-burger/dishes", nil)
	listRec := httptest.NewRecorder()
	router.ServeHTTP(listRec, listReq)

	if listRec.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK for list dishes, got %d", listRec.Code)
	}

	var dishes []Dish
	_ = json.NewDecoder(listRec.Body).Decode(&dishes)
	if len(dishes) != 1 || dishes[0].Name != "Smoked Bacon Cheeseburger" {
		t.Fatalf("Expected 1 dish in postgres, got: %+v", dishes)
	}

	// 4. Delete restaurant (DELETE /restaurants/{slug}) - tests cascade delete
	delReq := httptest.NewRequest(http.MethodDelete, "/restaurants/smoke-burger", nil)
	delRec := httptest.NewRecorder()
	router.ServeHTTP(delRec, delReq)

	if delRec.Code != http.StatusNoContent {
		t.Fatalf("Expected 204 No Content for delete restaurant, got %d", delRec.Code)
	}

	// 5. Verify restaurant is gone (GET /restaurants/{slug} -> 404)
	getReq := httptest.NewRequest(http.MethodGet, "/restaurants/smoke-burger", nil)
	getRec := httptest.NewRecorder()
	router.ServeHTTP(getRec, getReq)

	if getRec.Code != http.StatusNotFound {
		t.Fatalf("Expected 404 Not Found after deletion, got %d", getRec.Code)
	}

	// 6. Verify dishes were cascade deleted from database directly
	var dishCount int
	_ = db.QueryRow("SELECT COUNT(*) FROM dishes WHERE restaurant_slug = 'smoke-burger'").Scan(&dishCount)
	if dishCount != 0 {
		t.Fatalf("Expected 0 dishes in DB after cascade delete, found %d", dishCount)
	}

	t.Log("🎉 Postgres integration smoke test passed: Restaurant created, Dish added, and Cascade Delete verified.")
}
