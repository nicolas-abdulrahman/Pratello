package main

import (
	"context"

	"github.com/stretchr/testify/mock"
)

// MockRestaurantStore is a mockery-compatible mock for RestaurantStore.
type MockRestaurantStore struct {
	mock.Mock
}

func NewMockRestaurantStore() *MockRestaurantStore {
	return &MockRestaurantStore{}
}

func (m *MockRestaurantStore) List(ctx context.Context) ([]Restaurant, error) {
	args := m.Called(ctx)
	if res := args.Get(0); res != nil {
		return res.([]Restaurant), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockRestaurantStore) GetBySlug(ctx context.Context, slug string) (*Restaurant, error) {
	args := m.Called(ctx, slug)
	if res := args.Get(0); res != nil {
		return res.(*Restaurant), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockRestaurantStore) Create(ctx context.Context, r *Restaurant) error {
	args := m.Called(ctx, r)
	return args.Error(0)
}

func (m *MockRestaurantStore) Update(ctx context.Context, slug string, req UpdateRestaurantRequest) (*Restaurant, error) {
	args := m.Called(ctx, slug, req)
	if res := args.Get(0); res != nil {
		return res.(*Restaurant), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockRestaurantStore) Delete(ctx context.Context, slug string) error {
	args := m.Called(ctx, slug)
	return args.Error(0)
}

// MockDishStore is a mockery-compatible mock for DishStore.
type MockDishStore struct {
	mock.Mock
}

func NewMockDishStore() *MockDishStore {
	return &MockDishStore{}
}

func (m *MockDishStore) ListByRestaurant(ctx context.Context, restaurantSlug string) ([]Dish, error) {
	args := m.Called(ctx, restaurantSlug)
	if res := args.Get(0); res != nil {
		return res.([]Dish), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockDishStore) Create(ctx context.Context, restaurantSlug string, req CreateDishRequest) (*Dish, error) {
	args := m.Called(ctx, restaurantSlug, req)
	if res := args.Get(0); res != nil {
		return res.(*Dish), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockDishStore) Update(ctx context.Context, restaurantSlug string, dishID int, req UpdateDishRequest) (*Dish, error) {
	args := m.Called(ctx, restaurantSlug, dishID, req)
	if res := args.Get(0); res != nil {
		return res.(*Dish), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockDishStore) Delete(ctx context.Context, restaurantSlug string, dishID int) error {
	args := m.Called(ctx, restaurantSlug, dishID)
	return args.Error(0)
}
