package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// testAddRestaurant is a helper function that sends a POST request with a Restaurant struct
func testAddRestaurant(t *testing.T, router http.Handler, r Restaurant) Restaurant {
	t.Helper()

	body, err := json.Marshal(r)
	if err != nil {
		t.Fatalf("Failed to marshal Restaurant struct: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/restaurants", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("Expected status 201 Created, got %d. Body: %s", rec.Code, rec.Body.String())
	}

	var created Restaurant
	if err := json.NewDecoder(rec.Body).Decode(&created); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	return created
}

func TestAddRestaurant(t *testing.T) {
	// Reset restaurants store before test
	restaurantsMu.Lock()
	restaurants = make(map[string]Restaurant)
	restaurantsMu.Unlock()

	router := setupRouter()

	// Pass Restaurant struct directly
	sample := Restaurant{
		Slug:        "deliciadeprazeres",
		Name:        "Delícia de Prazeres",
		Description: "Authentic comfort food, sweets, and savory specialties.",
	}

	created := testAddRestaurant(t, router, sample)

	if created.Slug != sample.Slug {
		t.Errorf("Expected slug '%s', got '%s'", sample.Slug, created.Slug)
	}
	if created.Name != sample.Name {
		t.Errorf("Expected name '%s', got '%s'", sample.Name, created.Name)
	}
	if created.Description != sample.Description {
		t.Errorf("Expected description '%s', got '%s'", sample.Description, created.Description)
	}

	// Verify retrieval via GET /api/restaurants/{slug}
	getReq := httptest.NewRequest(http.MethodGet, "/api/restaurants/"+sample.Slug, nil)
	getRec := httptest.NewRecorder()

	router.ServeHTTP(getRec, getReq)

	if getRec.Code != http.StatusOK {
		t.Fatalf("Expected status 200 OK, got %d. Body: %s", getRec.Code, getRec.Body.String())
	}

	var fetched Restaurant
	if err := json.NewDecoder(getRec.Body).Decode(&fetched); err != nil {
		t.Fatalf("Failed to decode GET response: %v", err)
	}

	if fetched.Slug != sample.Slug || fetched.Name != sample.Name {
		t.Errorf("Fetched restaurant mismatch: %+v", fetched)
	}
}

func TestAddRestaurantValidation(t *testing.T) {
	router := setupRouter()

	// Test missing name
	invalidPayload := map[string]string{
		"slug": "some-slug",
	}
	body, _ := json.Marshal(invalidPayload)
	req := httptest.NewRequest(http.MethodPost, "/api/restaurants", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected 400 Bad Request for missing name, got %d", rec.Code)
	}
}

func TestGetNonExistentRestaurant(t *testing.T) {
	// Reset restaurants store
	restaurantsMu.Lock()
	restaurants = make(map[string]Restaurant)
	restaurantsMu.Unlock()

	router := setupRouter()

	req := httptest.NewRequest(http.MethodGet, "/api/restaurants/nonexistent", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("Expected 404 Not Found, got %d", rec.Code)
	}
}

func TestListRestaurants(t *testing.T) {
	restaurantsMu.Lock()
	restaurants = make(map[string]Restaurant)
	restaurants["deliciadeprazeres"] = Restaurant{
		Slug:        "deliciadeprazeres",
		Name:        "Delícia de Prazeres",
		Description: "Authentic comfort food, sweets, and savory specialties.",
	}
	restaurantsMu.Unlock()

	router := setupRouter()

	req := httptest.NewRequest(http.MethodGet, "/api/restaurants", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK, got %d", rec.Code)
	}

	var list []Restaurant
	if err := json.NewDecoder(rec.Body).Decode(&list); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(list) != 1 {
		t.Fatalf("Expected 1 restaurant in list, got %d", len(list))
	}
	if list[0].Slug != "deliciadeprazeres" {
		t.Errorf("Expected slug deliciadeprazeres, got %s", list[0].Slug)
	}
}

func TestDishLifecycle(t *testing.T) {
	app := NewApp(nil)
	router := setupRouter(app)

	// 1. Create a restaurant first
	restBody, _ := json.Marshal(Restaurant{
		Slug: "pizzaria",
		Name: "Pizzaria da Nonna",
	})
	rReq := httptest.NewRequest(http.MethodPost, "/restaurants", bytes.NewReader(restBody))
	rReq.Header.Set("Content-Type", "application/json")
	rRec := httptest.NewRecorder()
	router.ServeHTTP(rRec, rReq)
	if rRec.Code != http.StatusCreated {
		t.Fatalf("Failed to create restaurant: %s", rRec.Body.String())
	}

	// 2. Add a dish to the restaurant
	dishPayload, _ := json.Marshal(CreateDishRequest{
		Name:  "Pizza Margherita",
		Price: 12.50,
	})
	dReq := httptest.NewRequest(http.MethodPost, "/restaurants/pizzaria/dishes", bytes.NewReader(dishPayload))
	dReq.Header.Set("Content-Type", "application/json")
	dRec := httptest.NewRecorder()
	router.ServeHTTP(dRec, dReq)
	if dRec.Code != http.StatusCreated {
		t.Fatalf("Expected 201 Created for dish, got %d: %s", dRec.Code, dRec.Body.String())
	}

	var created Dish
	_ = json.NewDecoder(dRec.Body).Decode(&created)
	if created.Name != "Pizza Margherita" || created.Price != 12.50 {
		t.Fatalf("Unexpected dish created: %+v", created)
	}

	// 3. List dishes for the restaurant
	listReq := httptest.NewRequest(http.MethodGet, "/restaurants/pizzaria/dishes", nil)
	listRec := httptest.NewRecorder()
	router.ServeHTTP(listRec, listReq)
	if listRec.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK for list dishes, got %d", listRec.Code)
	}

	// 4. Update the dish
	updatePayload, _ := json.Marshal(UpdateDishRequest{
		Name:  "Pizza Margherita Grande",
		Price: 15.00,
	})
	uReq := httptest.NewRequest(http.MethodPut, "/restaurants/pizzaria/dishes/1", bytes.NewReader(updatePayload))
	uReq.Header.Set("Content-Type", "application/json")
	uRec := httptest.NewRecorder()
	router.ServeHTTP(uRec, uReq)
	if uRec.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK for update dish, got %d", uRec.Code)
	}

	// 5. Delete the dish
	delReq := httptest.NewRequest(http.MethodDelete, "/restaurants/pizzaria/dishes/1", nil)
	delRec := httptest.NewRecorder()
	router.ServeHTTP(delRec, delReq)
	if delRec.Code != http.StatusNoContent {
		t.Fatalf("Expected 204 No Content for delete dish, got %d", delRec.Code)
	}
}

