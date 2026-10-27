package main

import (
	"context"
	"errors"
	"testing"
)

// In-memory fake store used solely to test Service Layer logic without any DB or mock framework.
type fakeRestaurantStore struct {
	data map[string]*Restaurant
}

func newFakeRestaurantStore() *fakeRestaurantStore {
	return &fakeRestaurantStore{data: make(map[string]*Restaurant)}
}

func (f *fakeRestaurantStore) List(ctx context.Context) ([]Restaurant, error) {
	var list []Restaurant
	for _, r := range f.data {
		list = append(list, *r)
	}
	return list, nil
}

func (f *fakeRestaurantStore) GetBySlug(ctx context.Context, slug string) (*Restaurant, error) {
	r, exists := f.data[slug]
	if !exists {
		return nil, ErrNotFound
	}
	return r, nil
}

func (f *fakeRestaurantStore) Create(ctx context.Context, r *Restaurant) error {
	if _, exists := f.data[r.Slug]; exists {
		return ErrConflict
	}
	f.data[r.Slug] = r
	return nil
}

func (f *fakeRestaurantStore) Update(ctx context.Context, slug string, input UpdateRestaurantInput) (*Restaurant, error) {
	r, exists := f.data[slug]
	if !exists {
		return nil, ErrNotFound
	}
	if input.Name != "" {
		r.Name = input.Name
	}
	r.Description = input.Description
	return r, nil
}

func (f *fakeRestaurantStore) Delete(ctx context.Context, slug string) error {
	if _, exists := f.data[slug]; !exists {
		return ErrNotFound
	}
	delete(f.data, slug)
	return nil
}

type fakeDishStore struct {
	dishes  map[int]*Dish
	counter int
}

func newFakeDishStore() *fakeDishStore {
	return &fakeDishStore{dishes: make(map[int]*Dish)}
}

func (f *fakeDishStore) ListByRestaurant(ctx context.Context, slug string) ([]Dish, error) {
	var list []Dish
	for _, d := range f.dishes {
		if d.RestaurantSlug == slug {
			list = append(list, *d)
		}
	}
	return list, nil
}

func (f *fakeDishStore) Create(ctx context.Context, input NewDish) (*Dish, error) {
	f.counter++
	d := &Dish{ID: f.counter, RestaurantSlug: input.RestaurantSlug, Name: input.Name, Price: input.Price}
	f.dishes[f.counter] = d
	return d, nil
}

func (f *fakeDishStore) Update(ctx context.Context, slug string, id int, input UpdateDishInput) (*Dish, error) {
	d, exists := f.dishes[id]
	if !exists || d.RestaurantSlug != slug {
		return nil, ErrNotFound
	}
	d.Name = input.Name
	d.Price = input.Price
	return d, nil
}

func (f *fakeDishStore) Delete(ctx context.Context, slug string, id int) error {
	delete(f.dishes, id)
	return nil
}

// ---------------- RESTAURANT SERVICE BUSINESS LOGIC TESTS ----------------

func TestRestaurantService_CreateRestaurant_AutoSlug(t *testing.T) {
	store := newFakeRestaurantStore()
	svc := NewRestaurantService(store)

	input := NewRestaurant{
		Name:        "Pizza & Burger Gourmet!",
		Description: "Craft pizzas",
	}

	res, err := svc.CreateRestaurant(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.Slug != "pizza-burger-gourmet" {
		t.Errorf("expected normalized slug 'pizza-burger-gourmet', got '%s'", res.Slug)
	}
	if res.Name != "Pizza & Burger Gourmet!" {
		t.Errorf("expected name preserved, got '%s'", res.Name)
	}
}

func TestRestaurantService_CreateRestaurant_ValidationErrors(t *testing.T) {
	store := newFakeRestaurantStore()
	svc := NewRestaurantService(store)

	tests := []struct {
		name    string
		input   NewRestaurant
		wantErr bool
	}{
		{
			name:    "Name too short",
			input:   NewRestaurant{Name: "A"},
			wantErr: true,
		},
		{
			name:    "Empty name",
			input:   NewRestaurant{Name: "   "},
			wantErr: true,
		},
		{
			name:    "Valid name",
			input:   NewRestaurant{Name: "Bela Napoli"},
			wantErr: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := svc.CreateRestaurant(context.Background(), tc.input)
			if (err != nil) != tc.wantErr {
				t.Fatalf("expected error: %v, got %v", tc.wantErr, err)
			}
		})
	}
}

func TestRestaurantService_CreateRestaurant_DuplicateSlugConflict(t *testing.T) {
	store := newFakeRestaurantStore()
	svc := NewRestaurantService(store)

	// First creation succeeds
	_, err := svc.CreateRestaurant(context.Background(), NewRestaurant{
		Name: "Pizzaria Bela",
		Slug: "bela",
	})
	if err != nil {
		t.Fatalf("first creation failed: %v", err)
	}

	// Second creation with same slug must fail with ErrConflict
	_, err = svc.CreateRestaurant(context.Background(), NewRestaurant{
		Name: "Outra Bela",
		Slug: "bela",
	})
	if !errors.Is(err, ErrConflict) {
		t.Fatalf("expected ErrConflict, got: %v", err)
	}
}

// ---------------- DISH SERVICE BUSINESS LOGIC TESTS ----------------

func TestDishService_AddDish_Validation(t *testing.T) {
	rStore := newFakeRestaurantStore()
	dStore := newFakeDishStore()
	svc := NewDishService(rStore, dStore)

	// Insert parent restaurant
	_ = rStore.Create(context.Background(), &Restaurant{Slug: "pizzaria", Name: "Pizzaria"})

	// Negative price rejected
	_, err := svc.AddDish(context.Background(), NewDish{RestaurantSlug: "pizzaria", Name: "Pizza", Price: -5.00})
	if !errors.Is(err, ErrValidation) {
		t.Errorf("expected ErrValidation for negative price, got %v", err)
	}

	// Zero price rejected
	_, err = svc.AddDish(context.Background(), NewDish{RestaurantSlug: "pizzaria", Name: "Pizza", Price: 0})
	if !errors.Is(err, ErrValidation) {
		t.Errorf("expected ErrValidation for zero price, got %v", err)
	}

	// Name too short rejected
	_, err = svc.AddDish(context.Background(), NewDish{RestaurantSlug: "pizzaria", Name: "P", Price: 10.00})
	if !errors.Is(err, ErrValidation) {
		t.Errorf("expected ErrValidation for short name, got %v", err)
	}

	// Non-existent restaurant rejected
	_, err = svc.AddDish(context.Background(), NewDish{RestaurantSlug: "ghost-restaurant", Name: "Pizza", Price: 10.00})
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound for missing restaurant, got %v", err)
	}

	// Valid dish succeeds
	dish, err := svc.AddDish(context.Background(), NewDish{RestaurantSlug: "pizzaria", Name: "Calzone", Price: 15.00})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dish.Name != "Calzone" || dish.Price != 15.00 {
		t.Errorf("unexpected dish data: %+v", dish)
	}
}

// ---------------- LOYALTY SERVICE BUSINESS LOGIC TESTS ----------------

func TestLoyaltyService_AccrualAndRedemptionRules(t *testing.T) {
	svc := NewLoyaltyService()
	cpf := "52998224725"

	// 1. Accrue coins: R$ 95 order -> 9 coins (R$ 10 = 1 coin)
	balance := svc.AccrueCoins(cpf, 95.00)
	if balance != 9 {
		t.Fatalf("expected 9 coins accrued for 95.00, got %d", balance)
	}

	// Small order < R$ 10 earns 0 coins
	balance = svc.AccrueCoins(cpf, 8.50)
	if balance != 9 {
		t.Fatalf("expected balance still 9 for small purchase, got %d", balance)
	}

	// 2. Redemption without account MUST FAIL (Core business rule!)
	_, err := svc.RedeemCoins(cpf, 5, false /* hasAccount = false */)
	if !errors.Is(err, ErrAccountRequiredForRedemption) {
		t.Fatalf("expected ErrAccountRequiredForRedemption when hasAccount=false, got: %v", err)
	}

	// 3. Redemption with account but exceeding balance MUST FAIL
	_, err = svc.RedeemCoins(cpf, 20, true /* hasAccount = true */)
	if !errors.Is(err, ErrInsufficientCoins) {
		t.Fatalf("expected ErrInsufficientCoins, got: %v", err)
	}

	// 4. Valid redemption with account succeeds
	discount, err := svc.RedeemCoins(cpf, 5, true /* hasAccount = true */)
	if err != nil {
		t.Fatalf("unexpected error on valid redemption: %v", err)
	}
	if discount != 5.00 {
		t.Errorf("expected R$ 5.00 discount, got %f", discount)
	}

	// Remaining balance must be 4
	if svc.balances[cpf] != 4 {
		t.Errorf("expected remaining balance 4, got %d", svc.balances[cpf])
	}
}
