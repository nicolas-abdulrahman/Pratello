package restaurants_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"pratello/server/internal/restaurants"
)

// ============================================================================
// SECTION 1: DOMAIN SERVICE UNIT TESTS
// ============================================================================

func TestRestaurantService_CreateRestaurant_AutoSlug(t *testing.T) {
	memStore := restaurants.NewMemoryStore()
	svc := restaurants.NewService(memStore)

	input := restaurants.NewRestaurant{
		Name:        "Pizzaria do Bairro & Cia",
		Description: "A melhor pizza da cidade",
	}

	res, err := svc.CreateRestaurant(context.Background(), input)
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}

	expectedSlug := "pizzaria-do-bairro-cia"
	if res.Slug != expectedSlug {
		t.Errorf("expected slug '%s', got '%s'", expectedSlug, res.Slug)
	}
}

func TestRestaurantService_CreateRestaurant_ValidationErrors(t *testing.T) {
	memStore := restaurants.NewMemoryStore()
	svc := restaurants.NewService(memStore)

	testCases := []struct {
		name        string
		input       restaurants.NewRestaurant
		errExpected string
	}{
		{
			name:        "Name too short",
			input:       restaurants.NewRestaurant{Name: "A"},
			errExpected: "at least 2 characters",
		},
		{
			name:        "Empty name",
			input:       restaurants.NewRestaurant{Name: "   "},
			errExpected: "at least 2 characters",
		},
		{
			name:        "Name too long",
			input:       restaurants.NewRestaurant{Name: strings.Repeat("X", 101)},
			errExpected: "cannot exceed 100 characters",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := svc.CreateRestaurant(context.Background(), tc.input)
			if err == nil {
				t.Fatalf("expected validation error, got nil")
			}
			if !errors.Is(err, restaurants.ErrValidation) {
				t.Errorf("expected ErrValidation in chain, got: %v", err)
			}
			if !strings.Contains(err.Error(), tc.errExpected) {
				t.Errorf("expected error message to contain '%s', got '%s'", tc.errExpected, err.Error())
			}
		})
	}
}

func TestRestaurantService_CreateRestaurant_DuplicateSlugConflict(t *testing.T) {
	memStore := restaurants.NewMemoryStore()
	svc := restaurants.NewService(memStore)

	_, err := svc.CreateRestaurant(context.Background(), restaurants.NewRestaurant{
		Name: "Luigi Pizza",
		Slug: "luigi-pizza",
	})
	if err != nil {
		t.Fatalf("unexpected error creating first restaurant: %v", err)
	}

	_, err = svc.CreateRestaurant(context.Background(), restaurants.NewRestaurant{
		Name: "Luigi Pizza Premium",
		Slug: "luigi-pizza",
	})
	if err == nil {
		t.Fatalf("expected conflict error, got nil")
	}
	if !errors.Is(err, restaurants.ErrConflict) {
		t.Errorf("expected ErrConflict, got %v", err)
	}
}

func TestDishService_AddDish_Validation(t *testing.T) {
	memStore := restaurants.NewMemoryStore()
	svc := restaurants.NewService(memStore)

	// Attempt adding dish to non-existent restaurant
	_, err := svc.AddDish(context.Background(), restaurants.NewDish{
		RestaurantSlug: "non-existent",
		Name:           "Calabresa",
		Price:          35.00,
	})
	if !errors.Is(err, restaurants.ErrNotFound) {
		t.Errorf("expected ErrNotFound for missing restaurant, got: %v", err)
	}

	// Create restaurant
	_, err = svc.CreateRestaurant(context.Background(), restaurants.NewRestaurant{
		Name: "Cantina da Nona",
		Slug: "cantina-da-nona",
	})
	if err != nil {
		t.Fatalf("failed to create restaurant: %v", err)
	}

	// Validation: Price <= 0
	_, err = svc.AddDish(context.Background(), restaurants.NewDish{
		RestaurantSlug: "cantina-da-nona",
		Name:           "Calabresa",
		Price:          0,
	})
	if !errors.Is(err, restaurants.ErrValidation) {
		t.Errorf("expected ErrValidation for zero price, got: %v", err)
	}

	// Validation: Name too short
	_, err = svc.AddDish(context.Background(), restaurants.NewDish{
		RestaurantSlug: "cantina-da-nona",
		Name:           "C",
		Price:          25.00,
	})
	if !errors.Is(err, restaurants.ErrValidation) {
		t.Errorf("expected ErrValidation for short dish name, got: %v", err)
	}

	// Successful creation
	dish, err := svc.AddDish(context.Background(), restaurants.NewDish{
		RestaurantSlug: "cantina-da-nona",
		Name:           "Pizza Margherita",
		Price:          45.00,
	})
	if err != nil {
		t.Fatalf("unexpected error adding valid dish: %v", err)
	}
	if dish.Name != "Pizza Margherita" {
		t.Errorf("expected dish name 'Pizza Margherita', got '%s'", dish.Name)
	}
}

func TestLoyaltyService_AccrualAndRedemptionRules(t *testing.T) {
	svc := restaurants.NewService(restaurants.NewMemoryStore())
	loyalty := svc.Loyalty()
	testCPF := "12345678901"

	// Rule 1: Accrual is 1 coin per R$ 10.00
	earned := loyalty.CalculateEarnedCoins(9.99)
	if earned != 0 {
		t.Errorf("expected 0 coins for R$ 9.99, got %d", earned)
	}

	earned = loyalty.CalculateEarnedCoins(25.50)
	if earned != 2 {
		t.Errorf("expected 2 coins for R$ 25.50, got %d", earned)
	}

	bal := loyalty.AccrueCoins(testCPF, 105.00) // 10 coins
	if bal != 10 {
		t.Errorf("expected balance of 10 coins, got %d", bal)
	}

	// Rule 2: Invariant: Anonymous users (hasAccount == false) CANNOT redeem coins
	_, err := loyalty.RedeemCoins(testCPF, 5, false)
	if !errors.Is(err, restaurants.ErrAccountRequiredForRedemption) {
		t.Errorf("expected ErrAccountRequiredForRedemption for non-registered user, got: %v", err)
	}

	// Rule 3: Registered users with sufficient balance can redeem
	discount, err := loyalty.RedeemCoins(testCPF, 5, true)
	if err != nil {
		t.Fatalf("expected successful redemption, got: %v", err)
	}
	if discount != 5.00 {
		t.Errorf("expected R$ 5.00 discount, got %f", discount)
	}

	// Rule 4: Overspending balance should fail
	_, err = loyalty.RedeemCoins(testCPF, 10, true)
	if !errors.Is(err, restaurants.ErrValidation) {
		t.Errorf("expected ErrValidation when redeeming more coins than available, got: %v", err)
	}
}

// ============================================================================
// SECTION 2: STORE CONTRACT & PERSISTENCE UNIT TESTS
// ============================================================================

func TestStore_RestaurantAndDish_CRUD(t *testing.T) {
	ctx := context.Background()
	store := restaurants.NewMemoryStore()

	// 1. Create Restaurant
	r := &restaurants.Restaurant{
		Slug:        "trattoria-roma",
		Name:        "Trattoria Roma",
		Description: "Autêntica comida italiana",
	}
	if err := store.CreateRestaurant(ctx, r); err != nil {
		t.Fatalf("failed to create restaurant: %v", err)
	}

	// 2. Duplicate Create returns ErrConflict
	if err := store.CreateRestaurant(ctx, r); !errors.Is(err, restaurants.ErrConflict) {
		t.Fatalf("expected ErrConflict on duplicate create, got: %v", err)
	}

	// 3. GetBySlug
	fetched, err := store.GetRestaurantBySlug(ctx, "trattoria-roma")
	if err != nil {
		t.Fatalf("failed to get restaurant: %v", err)
	}
	if fetched.Name != "Trattoria Roma" {
		t.Errorf("expected name 'Trattoria Roma', got '%s'", fetched.Name)
	}

	// 4. Update
	updated, err := store.UpdateRestaurant(ctx, "trattoria-roma", restaurants.UpdateRestaurantInput{
		Name:        "Trattoria Roma Nova",
		Description: "Nova descrição",
	})
	if err != nil {
		t.Fatalf("failed to update restaurant: %v", err)
	}
	if updated.Name != "Trattoria Roma Nova" {
		t.Errorf("expected updated name 'Trattoria Roma Nova', got '%s'", updated.Name)
	}

	// 5. Create Dish
	dish, err := store.CreateDish(ctx, restaurants.NewDish{
		RestaurantSlug: "trattoria-roma",
		Name:           "Carbonara",
		Price:          42.50,
	})
	if err != nil {
		t.Fatalf("failed to create dish: %v", err)
	}
	if dish.ID == nil {
		t.Fatalf("expected dish ID to be set")
	}

	// 6. List Dishes
	dishes, err := store.ListDishes(ctx, "trattoria-roma")
	if err != nil {
		t.Fatalf("failed to list dishes: %v", err)
	}
	if len(dishes) != 1 {
		t.Fatalf("expected 1 dish, got %d", len(dishes))
	}

	// 7. Delete Restaurant
	if err := store.DeleteRestaurant(ctx, "trattoria-roma"); err != nil {
		t.Fatalf("failed to delete restaurant: %v", err)
	}

	// 8. Get non-existent returns ErrNotFound
	_, err = store.GetRestaurantBySlug(ctx, "trattoria-roma")
	if !errors.Is(err, restaurants.ErrNotFound) {
		t.Errorf("expected ErrNotFound after deletion, got: %v", err)
	}
}

// ============================================================================
// SECTION 3: HTTP TRANSPORT UNIT TESTS
// ============================================================================

func TestHTTP_Restaurants_And_Menu(t *testing.T) {
	store := restaurants.NewMemoryStore()
	svc := restaurants.NewService(store)
	h := restaurants.NewHandler(svc)

	mux := http.NewServeMux()
	h.Routes(mux)

	// 1. POST /restaurants
	body := `{"name":"Pizzaria Bella","slug":"bella-napoli","description":"Forno a lenha"}`
	req := httptest.NewRequest(http.MethodPost, "/restaurants", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created, got %d: %s", w.Code, w.Body.String())
	}

	// 2. GET /restaurants/bella-napoli
	req = httptest.NewRequest(http.MethodGet, "/restaurants/bella-napoli", nil)
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d: %s", w.Code, w.Body.String())
	}

	// 3. POST /restaurants/bella-napoli/dishes
	dishBody := `{"name":"Pizza Margherita","price":39.90,"category":"Pizzas"}`
	req = httptest.NewRequest(http.MethodPost, "/restaurants/bella-napoli/dishes", strings.NewReader(dishBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created for dish, got %d: %s", w.Code, w.Body.String())
	}

	// 4. GET /restaurants/bella-napoli/menu
	req = httptest.NewRequest(http.MethodGet, "/restaurants/bella-napoli/menu", nil)
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for menu, got %d: %s", w.Code, w.Body.String())
	}

	var menuResp struct {
		Restaurant map[string]any `json:"restaurant"`
		Categories []string       `json:"categories"`
		Dishes     []map[string]any `json:"dishes"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &menuResp); err != nil {
		t.Fatalf("failed to decode menu response: %v", err)
	}
	if len(menuResp.Categories) == 0 || menuResp.Categories[0] != "Mais Pedidos" {
		t.Errorf("expected 'Mais Pedidos' as first category, got: %v", menuResp.Categories)
	}
	if len(menuResp.Dishes) != 1 {
		t.Errorf("expected 1 dish in menu response, got %d", len(menuResp.Dishes))
	}

	// 5. Duplicate create -> 409 Conflict
	req = httptest.NewRequest(http.MethodPost, "/restaurants", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Errorf("expected 409 Conflict, got %d: %s", w.Code, w.Body.String())
	}
}
