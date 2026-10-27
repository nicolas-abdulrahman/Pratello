package main

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"unicode"
)

var (
	ErrValidation                    = errors.New("validation failed")
	ErrAccountRequiredForRedemption = errors.New("account is required to redeem coins")
	ErrInsufficientCoins             = errors.New("insufficient loyalty coins")
)

// RestaurantService handles business domain logic for restaurants.
type RestaurantService struct {
	store RestaurantStore
}

func NewRestaurantService(store RestaurantStore) *RestaurantService {
	return &RestaurantService{store: store}
}

// NormalizeSlug converts a name to a clean, URL-safe kebab-case slug.
func NormalizeSlug(input string) string {
	input = strings.TrimSpace(strings.ToLower(input))
	// Replace spaces and underscores with hyphens
	re := regexp.MustCompile(`[\s_]+`)
	input = re.ReplaceAllString(input, "-")
	// Remove all characters except a-z, 0-9, and hyphen
	clean := strings.Builder{}
	for _, r := range input {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '-' {
			clean.WriteRune(r)
		}
	}
	res := clean.String()
	// Strip repeated hyphens
	reHyphens := regexp.MustCompile(`-+`)
	return strings.Trim(reHyphens.ReplaceAllString(res, "-"), "-")
}

// CreateRestaurant handles validation, slug generation, and persistence.
func (s *RestaurantService) CreateRestaurant(ctx context.Context, input NewRestaurant) (*Restaurant, error) {
	name := strings.TrimSpace(input.Name)
	if len(name) < 2 {
		return nil, errors.Join(ErrValidation, errors.New("restaurant name must have at least 2 characters"))
	}
	if len(name) > 100 {
		return nil, errors.Join(ErrValidation, errors.New("restaurant name cannot exceed 100 characters"))
	}

	slug := strings.TrimSpace(input.Slug)
	if slug == "" {
		slug = NormalizeSlug(name)
	} else {
		slug = NormalizeSlug(slug)
	}

	if slug == "" {
		return nil, errors.Join(ErrValidation, errors.New("failed to generate valid slug from name"))
	}

	// Invariant: check if slug is already taken
	existing, err := s.store.GetBySlug(ctx, slug)
	if err == nil && existing != nil {
		return nil, ErrConflict
	}

	restaurant := &Restaurant{
		Slug:        slug,
		Name:        name,
		Description: strings.TrimSpace(input.Description),
	}

	if err := s.store.Create(ctx, restaurant); err != nil {
		return nil, err
	}
	return restaurant, nil
}

// DishService handles business logic for menu items and dishes.
type DishService struct {
	restaurants RestaurantStore
	dishes      DishStore
}

func NewDishService(restaurants RestaurantStore, dishes DishStore) *DishService {
	return &DishService{
		restaurants: restaurants,
		dishes:      dishes,
	}
}

// AddDish validates price, name, and restaurant existence before adding.
func (s *DishService) AddDish(ctx context.Context, input NewDish) (*Dish, error) {
	name := strings.TrimSpace(input.Name)
	if len(name) < 2 {
		return nil, errors.Join(ErrValidation, errors.New("dish name must have at least 2 characters"))
	}
	if input.Price <= 0 {
		return nil, errors.Join(ErrValidation, errors.New("dish price must be greater than zero"))
	}

	slug := strings.ToLower(strings.TrimSpace(input.RestaurantSlug))
	if slug == "" {
		return nil, errors.Join(ErrValidation, errors.New("restaurant slug is required"))
	}

	// Verify restaurant exists
	_, err := s.restaurants.GetBySlug(ctx, slug)
	if err != nil {
		return nil, ErrNotFound
	}

	return s.dishes.Create(ctx, NewDish{
		RestaurantSlug: slug,
		Name:           name,
		Price:          input.Price,
	})
}

// LoyaltyService implements loyalty coin rules:
// 1. Accrual: Anonymous/by CPF on purchase (1 coin per R$ 10).
// 2. Redemption: Strictly requires an active account (hasAccount == true).
type LoyaltyService struct {
	// In-memory or database balances
	balances map[string]int
}

func NewLoyaltyService() *LoyaltyService {
	return &LoyaltyService{
		balances: make(map[string]int),
	}
}

// CalculateEarnedCoins computes coin accrual for an order amount.
func (s *LoyaltyService) CalculateEarnedCoins(orderAmount float64) int {
	if orderAmount < 10.00 {
		return 0
	}
	return int(orderAmount / 10.00)
}

// AccrueCoins credits coins to a CPF.
func (s *LoyaltyService) AccrueCoins(cpf string, orderAmount float64) int {
	cpf = strings.TrimSpace(cpf)
	if len(cpf) != 11 {
		return 0
	}
	earned := s.CalculateEarnedCoins(orderAmount)
	if earned > 0 {
		s.balances[cpf] += earned
	}
	return s.balances[cpf]
}

// RedeemCoins spends coins. Invariant: strictly requires an active user account.
func (s *LoyaltyService) RedeemCoins(cpf string, coinsToRedeem int, hasAccount bool) (discountAmount float64, err error) {
	if !hasAccount {
		return 0, ErrAccountRequiredForRedemption
	}
	if coinsToRedeem <= 0 {
		return 0, errors.Join(ErrValidation, errors.New("coins to redeem must be greater than zero"))
	}
	currentBalance := s.balances[cpf]
	if currentBalance < coinsToRedeem {
		return 0, ErrInsufficientCoins
	}

	s.balances[cpf] -= coinsToRedeem
	// 1 coin = R$ 1.00 discount
	return float64(coinsToRedeem), nil
}
