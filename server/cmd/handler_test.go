package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// ---------------- RESTAURANT HANDLER UNIT TESTS ----------------

func TestUnit_HandleListRestaurants_Success(t *testing.T) {
	mockStore := NewMockRestaurantStore()
	expected := []Restaurant{
		{Slug: "pizzaria", Name: "Pizzaria Pratello", Description: "Wood-fired pizza"},
		{Slug: "trattoria", Name: "Trattoria", Description: "Pasta specialists"},
	}
	mockStore.On("List", mock.Anything).Return(expected, nil)

	app := &App{Restaurants: mockStore}
	mux := setupRouter(app)

	req := httptest.NewRequest(http.MethodGet, "/restaurants", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

	var body []Restaurant
	err := json.NewDecoder(rec.Body).Decode(&body)
	assert.NoError(t, err)
	assert.Equal(t, expected, body)
	mockStore.AssertExpectations(t)
}

func TestUnit_HandleListRestaurants_DatabaseError(t *testing.T) {
	mockStore := NewMockRestaurantStore()
	mockStore.On("List", mock.Anything).Return(nil, errors.New("db connection lost"))

	app := &App{Restaurants: mockStore}
	mux := setupRouter(app)

	req := httptest.NewRequest(http.MethodGet, "/restaurants", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	mockStore.AssertExpectations(t)
}

func TestUnit_HandleGetRestaurant_Success(t *testing.T) {
	mockStore := NewMockRestaurantStore()
	expected := &Restaurant{Slug: "pizzaria", Name: "Pizzaria Pratello", Description: "Wood-fired pizza"}
	mockStore.On("GetBySlug", mock.Anything, "pizzaria").Return(expected, nil)

	app := &App{Restaurants: mockStore}
	mux := setupRouter(app)

	req := httptest.NewRequest(http.MethodGet, "/restaurants/pizzaria", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var body Restaurant
	err := json.NewDecoder(rec.Body).Decode(&body)
	assert.NoError(t, err)
	assert.Equal(t, *expected, body)
	mockStore.AssertExpectations(t)
}

func TestUnit_HandleGetRestaurant_NotFound(t *testing.T) {
	mockStore := NewMockRestaurantStore()
	mockStore.On("GetBySlug", mock.Anything, "nonexistent").Return(nil, ErrNotFound)

	app := &App{Restaurants: mockStore}
	mux := setupRouter(app)

	req := httptest.NewRequest(http.MethodGet, "/restaurants/nonexistent", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
	mockStore.AssertExpectations(t)
}

func TestUnit_HandleCreateRestaurant_Success(t *testing.T) {
	mockStore := NewMockRestaurantStore()
	mockStore.On("Create", mock.Anything, mock.MatchedBy(func(r *Restaurant) bool {
		return r.Name == "New Place" && r.Slug == "new-place"
	})).Return(nil)

	app := &App{Restaurants: mockStore}
	mux := setupRouter(app)

	payload := map[string]string{
		"name":        "New Place",
		"description": "A great new spot",
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/restaurants", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)
	var created Restaurant
	err := json.NewDecoder(rec.Body).Decode(&created)
	assert.NoError(t, err)
	assert.Equal(t, "new-place", created.Slug)
	assert.Equal(t, "New Place", created.Name)
	mockStore.AssertExpectations(t)
}

func TestUnit_HandleCreateRestaurant_Conflict(t *testing.T) {
	mockStore := NewMockRestaurantStore()
	mockStore.On("Create", mock.Anything, mock.Anything).Return(ErrConflict)

	app := &App{Restaurants: mockStore}
	mux := setupRouter(app)

	payload := map[string]string{"name": "Existing Place", "slug": "existing"}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/restaurants", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusConflict, rec.Code)
	mockStore.AssertExpectations(t)
}

func TestUnit_HandleUpdateRestaurant_Success(t *testing.T) {
	mockStore := NewMockRestaurantStore()
	updateReq := UpdateRestaurantRequest{Name: "Updated Name", Description: "New desc"}
	expected := &Restaurant{Slug: "pizzaria", Name: "Updated Name", Description: "New desc"}
	mockStore.On("Update", mock.Anything, "pizzaria", updateReq).Return(expected, nil)

	app := &App{Restaurants: mockStore}
	mux := setupRouter(app)

	body, _ := json.Marshal(updateReq)
	req := httptest.NewRequest(http.MethodPut, "/restaurants/pizzaria", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var updated Restaurant
	err := json.NewDecoder(rec.Body).Decode(&updated)
	assert.NoError(t, err)
	assert.Equal(t, "Updated Name", updated.Name)
	mockStore.AssertExpectations(t)
}

func TestUnit_HandleDeleteRestaurant_Success(t *testing.T) {
	mockStore := NewMockRestaurantStore()
	mockStore.On("Delete", mock.Anything, "pizzaria").Return(nil)

	app := &App{Restaurants: mockStore}
	mux := setupRouter(app)

	req := httptest.NewRequest(http.MethodDelete, "/restaurants/pizzaria", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNoContent, rec.Code)
	mockStore.AssertExpectations(t)
}

// ---------------- DISH HANDLER UNIT TESTS ----------------

func TestUnit_HandleListDishes_Success(t *testing.T) {
	mockDishStore := NewMockDishStore()
	expected := []Dish{
		{ID: 1, RestaurantSlug: "pizzaria", Name: "Margherita", Price: 12.50},
		{ID: 2, RestaurantSlug: "pizzaria", Name: "Diavola", Price: 14.00},
	}
	mockDishStore.On("ListByRestaurant", mock.Anything, "pizzaria").Return(expected, nil)

	app := &App{Dishes: mockDishStore}
	mux := setupRouter(app)

	req := httptest.NewRequest(http.MethodGet, "/restaurants/pizzaria/dishes", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var body []Dish
	err := json.NewDecoder(rec.Body).Decode(&body)
	assert.NoError(t, err)
	assert.Equal(t, expected, body)
	mockDishStore.AssertExpectations(t)
}

func TestUnit_HandleCreateDish_Success(t *testing.T) {
	mockDishStore := NewMockDishStore()
	createReq := CreateDishRequest{Name: "Calzone", Price: 15.00}
	expected := &Dish{ID: 42, RestaurantSlug: "pizzaria", Name: "Calzone", Price: 15.00}
	mockDishStore.On("Create", mock.Anything, "pizzaria", createReq).Return(expected, nil)

	app := &App{Dishes: mockDishStore}
	mux := setupRouter(app)

	body, _ := json.Marshal(createReq)
	req := httptest.NewRequest(http.MethodPost, "/restaurants/pizzaria/dishes", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)
	var dish Dish
	err := json.NewDecoder(rec.Body).Decode(&dish)
	assert.NoError(t, err)
	assert.Equal(t, 42, dish.ID)
	assert.Equal(t, "Calzone", dish.Name)
	mockDishStore.AssertExpectations(t)
}

func TestUnit_HandleDeleteDish_Success(t *testing.T) {
	mockDishStore := NewMockDishStore()
	mockDishStore.On("Delete", mock.Anything, "pizzaria", 42).Return(nil)

	app := &App{Dishes: mockDishStore}
	mux := setupRouter(app)

	req := httptest.NewRequest(http.MethodDelete, "/restaurants/pizzaria/dishes/42", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNoContent, rec.Code)
	mockDishStore.AssertExpectations(t)
}
