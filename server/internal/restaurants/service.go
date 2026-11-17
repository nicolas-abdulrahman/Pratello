package restaurants

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

// Domain Errors: Exported as vars so http.go or callers can match with errors.Is.
var (
	ErrNotFound                     = errors.New("not found")
	ErrConflict                     = errors.New("conflict: resource already exists")
	ErrValidation                   = errors.New("validation failed")
	ErrAccountRequiredForRedemption = errors.New("loyalty coins can only be redeemed by users with an active registered account")
)

// ---------------- DOMAIN TYPES (NO JSON TAGS) ----------------

// Restaurant represents a logical restaurant entity.
type Restaurant struct {
	Slug         string
	Name         string
	Description  string
	Cuisine      string
	IsOpen       bool
	OpenTime     string
	ClosingTime  string
	Rating       float64
	ReviewsCount int
	DeliveryFee  float64
	CEP          string
	DeliveryTime string
	Neighborhood string
	Phone        string
	BannerURL    string
	LogoURL      string
	MenuItems    []string
}

// Dish represents a menu item belonging to a restaurant.
type Dish struct {
	ID             any
	RestaurantSlug string
	Name           string
	Description    string
	Category       string
	Tags           []string
	Price          float64
	ImageURL       string
}

// NewRestaurant holds the business fields required to create a restaurant.
type NewRestaurant struct {
	Name        string
	Slug        string
	Description string
}

// UpdateRestaurantInput holds modifications for an existing restaurant.
type UpdateRestaurantInput struct {
	Name        string
	Description string
}

// NewDish holds the business fields required to create a dish.
type NewDish struct {
	RestaurantSlug string
	Name           string
	Description    string
	Category       string
	Tags           []string
	Price          float64
	ImageURL       string
}

// UpdateDishInput holds modifications for an existing dish.
type UpdateDishInput struct {
	Name        string
	Description string
	Category    string
	Tags        []string
	Price       float64
	ImageURL    string
}

// store defines the consumer-side interface for restaurant persistence.
type store interface {
	ListRestaurants(ctx context.Context) ([]Restaurant, error)
	GetRestaurantBySlug(ctx context.Context, slug string) (*Restaurant, error)
	CreateRestaurant(ctx context.Context, r *Restaurant) error
	UpdateRestaurant(ctx context.Context, slug string, input UpdateRestaurantInput) (*Restaurant, error)
	DeleteRestaurant(ctx context.Context, slug string) error

	ListDishes(ctx context.Context, restaurantSlug string) ([]Dish, error)
	CreateDish(ctx context.Context, input NewDish) (*Dish, error)
	UpdateDish(ctx context.Context, restaurantSlug string, dishID int, input UpdateDishInput) (*Dish, error)
	DeleteDish(ctx context.Context, restaurantSlug string, dishID int) error
	UpdateDishImage(ctx context.Context, restaurantSlug string, dishID int, imageURL string) (*Dish, error)
}

// Service orchestrates domain logic for restaurants, dishes, menus, and loyalty.
type Service struct {
	store   store
	loyalty *LoyaltyService
}

// NewService constructs a Service with the provided persistence implementation.
func NewService(s store) *Service {
	return &Service{
		store:   s,
		loyalty: NewLoyaltyService(),
	}
}

// NormalizeSlug converts a name to a clean, URL-safe kebab-case slug.
func NormalizeSlug(input string) string {
	input = strings.TrimSpace(strings.ToLower(input))
	re := regexp.MustCompile(`[\s_]+`)
	input = re.ReplaceAllString(input, "-")

	clean := strings.Builder{}
	for _, r := range input {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '-' {
			clean.WriteRune(r)
		}
	}
	res := clean.String()
	reHyphens := regexp.MustCompile(`-+`)
	return strings.Trim(reHyphens.ReplaceAllString(res, "-"), "-")
}

// ListRestaurants retrieves all restaurants.
func (s *Service) ListRestaurants(ctx context.Context) ([]Restaurant, error) {
	return s.store.ListRestaurants(ctx)
}

// GetRestaurant fetches a single restaurant by slug.
func (s *Service) GetRestaurant(ctx context.Context, slug string) (*Restaurant, error) {
	slug = strings.ToLower(strings.TrimSpace(slug))
	if slug == "" {
		return nil, errors.Join(ErrValidation, errors.New("slug is required"))
	}
	return s.store.GetRestaurantBySlug(ctx, slug)
}

// CreateRestaurant handles name validation, slug generation/normalization, conflict detection, and persistence.
func (s *Service) CreateRestaurant(ctx context.Context, input NewRestaurant) (*Restaurant, error) {
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

	existing, err := s.store.GetRestaurantBySlug(ctx, slug)
	if err == nil && existing != nil {
		return nil, ErrConflict
	}

	restaurant := &Restaurant{
		Slug:        slug,
		Name:        name,
		Description: strings.TrimSpace(input.Description),
	}

	if err := s.store.CreateRestaurant(ctx, restaurant); err != nil {
		return nil, err
	}
	return restaurant, nil
}

// UpdateRestaurant modifies restaurant metadata.
func (s *Service) UpdateRestaurant(ctx context.Context, slug string, input UpdateRestaurantInput) (*Restaurant, error) {
	slug = strings.ToLower(strings.TrimSpace(slug))
	return s.store.UpdateRestaurant(ctx, slug, input)
}

// DeleteRestaurant deletes a restaurant by slug.
func (s *Service) DeleteRestaurant(ctx context.Context, slug string) error {
	slug = strings.ToLower(strings.TrimSpace(slug))
	return s.store.DeleteRestaurant(ctx, slug)
}

// ListDishes retrieves all dishes belonging to a restaurant.
func (s *Service) ListDishes(ctx context.Context, restaurantSlug string) ([]Dish, error) {
	restaurantSlug = strings.ToLower(strings.TrimSpace(restaurantSlug))
	return s.store.ListDishes(ctx, restaurantSlug)
}

// AddDish validates price, name, and restaurant existence before adding.
func (s *Service) AddDish(ctx context.Context, input NewDish) (*Dish, error) {
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
	_, err := s.store.GetRestaurantBySlug(ctx, slug)
	if err != nil {
		return nil, ErrNotFound
	}

	return s.store.CreateDish(ctx, NewDish{
		RestaurantSlug: slug,
		Name:           name,
		Description:    strings.TrimSpace(input.Description),
		Category:       strings.TrimSpace(input.Category),
		Tags:           input.Tags,
		Price:          input.Price,
		ImageURL:       strings.TrimSpace(input.ImageURL),
	})
}

// UpdateDish updates dish information.
func (s *Service) UpdateDish(ctx context.Context, slug string, id int, input UpdateDishInput) (*Dish, error) {
	slug = strings.ToLower(strings.TrimSpace(slug))
	return s.store.UpdateDish(ctx, slug, id, input)
}

// DeleteDish removes a dish from the menu.
func (s *Service) DeleteDish(ctx context.Context, slug string, id int) error {
	slug = strings.ToLower(strings.TrimSpace(slug))
	return s.store.DeleteDish(ctx, slug, id)
}

// UpdateDishImage updates the image URL for a dish.
func (s *Service) UpdateDishImage(ctx context.Context, slug string, id int, imageURL string) (*Dish, error) {
	slug = strings.ToLower(strings.TrimSpace(slug))
	return s.store.UpdateDishImage(ctx, slug, id, imageURL)
}

// GetRestaurantMenu returns the restaurant profile, ordered unique categories ("Mais Pedidos" first), and dishes.
func (s *Service) GetRestaurantMenu(ctx context.Context, slug string) (*Restaurant, []string, []Dish, error) {
	slug = strings.ToLower(strings.TrimSpace(slug))
	res, err := s.store.GetRestaurantBySlug(ctx, slug)
	if err != nil {
		return nil, nil, nil, err
	}

	dishes, err := s.store.ListDishes(ctx, slug)
	if err != nil && !errors.Is(err, ErrNotFound) {
		return nil, nil, nil, err
	}
	if dishes == nil {
		dishes = make([]Dish, 0)
	}

	catSet := make(map[string]bool)
	var categories []string
	categories = append(categories, "Mais Pedidos")
	catSet["Mais Pedidos"] = true

	for _, d := range dishes {
		cat := d.Category
		if cat == "" {
			cat = "Pizzas"
		}
		if !catSet[cat] {
			catSet[cat] = true
			categories = append(categories, cat)
		}
	}

	return res, categories, dishes, nil
}

// SeedDefaultRestaurant resets and re-seeds "mamma-mia-mario" with sample menu items.
func (s *Service) SeedDefaultRestaurant(ctx context.Context) error {
	marioSlug := "mamma-mia-mario"
	_ = s.store.DeleteRestaurant(ctx, marioSlug)
	if err := s.store.CreateRestaurant(ctx, &Restaurant{
		Slug:         marioSlug,
		Name:         "Mamma mia Mario",
		Description:  "Um pedacinho de Napoli. Pizza no fogão a lenha, feita com amor atrás do armário.",
		Cuisine:      "Italiana",
		IsOpen:       true,
		OpenTime:     "07:00",
		ClosingTime:  "22:00",
		Rating:       4.9,
		ReviewsCount: 324,
		DeliveryFee:  1.99,
		CEP:          "15555-555",
		DeliveryTime: "20–30 min",
		Neighborhood: "Mission District",
		Phone:        "(81) 95552-0148",
		BannerURL:    "https://lh3.googleusercontent.com/aida-public/AB6AXuD8i7X0qDoCvkouDq-uUv7UEDtkKSTah3gKO68ElZOQTiDH0KjKi_fyaIcGAqaF7mZ_ZgqueJ6j_U-Z-tt30B1fuhs28n6b5ZhcHjaFmjEgmxSSCJChcNGK27S-R-GsleUn0DChdITkkqd3lz-H6R4XMb81BP7tfPtAvQa2d2wjv4dqHh-zPRTB2cfC81OAkENDitgwJzfayVROlNm2Y_cpClw84FTUl8AaFP--O1M",
		LogoURL:      "https://i.pinimg.com/736x/9e/f2/b9/9ef2b9c1ac760f54c9e7aeb5c1784460.jpg",
	}); err != nil {
		return fmt.Errorf("failed to insert mamma-mia-mario: %w", err)
	}

	sampleDishes := []NewDish{
		{
			RestaurantSlug: marioSlug,
			Name:           "Margherita",
			Description:    "San Marzano tomato, fior di latte mozzarella, fresh basil.",
			Category:       "Pizzas",
			Tags:           []string{"MOST LOVED", "MOST ORDERED"},
			Price:          16.00,
			ImageURL:       "https://lh3.googleusercontent.com/aida-public/AB6AXuCb-eU4XxPHWKrxNYHn7-TPGZZiDJ7J43_wyHW35GGvB4ZYni8tfx_WXwM5FpIXR5sPzrKxs7IsbzM8We9VCULm7jAj3tnmjxGjgDsOIjDnBTMIWNwM3hGed30aeoT2VhHS9PkWMmiaJALNAfdW7EcUR2jaBsZXviPHVg-BLQ_Yf-JaqHsdxRAb3o6FWddiPqt9vxwkUpiL3darOkhD4kKfXX9-ff_dhL3G24HQmLw",
		},
		{
			RestaurantSlug: marioSlug,
			Name:           "Spicy Diavola",
			Description:    "Spicy salami, mozzarella, tomato and a drizzle of chili honey.",
			Category:       "Pizzas",
			Tags:           []string{"MOST LOVED"},
			Price:          19.00,
			ImageURL:       "https://lh3.googleusercontent.com/aida-public/AB6AXuCTmOctI0RewSPwuz-BWjt1KjY06cnghRQd_5aD3yZ1mdACQ5v5UPrVcEaIeltyls3yjVEikrahYf-2fyZK4fwkm_VkaoGT7Rf2YHIePG4HfGnJJBV-j7xf-amYSPa3IDHVXFqKEhwbdbJomOR_fOEgiJJV31ru_FWjl8K31Wfai5rpjXVPa8DMj8HxJ_xQ4WWAwo0ZsO10gZXIHGP9fjCJZqNz6-vn7YPrM4fqcM8",
		},
		{
			RestaurantSlug: marioSlug,
			Name:           "Burrata & tomatoes",
			Description:    "Creamy burrata, heirloom tomatoes, basil and extra virgin olive oil.",
			Category:       "Acompanhamentos",
			Tags:           []string{},
			Price:          14.00,
			ImageURL:       "https://lh3.googleusercontent.com/aida-public/AB6AXuBFETatsb6xiRzdMPM-1wMApk2RxLzpPFaJRKZ7yhdMVwciJo0Y53nDdCJM2yXu6qiqF4MZhPHNMAMwt_xVzRCnp2QM2M1YnoXM0bZks8dQ9-rrvPxY2LrYt1WwMYvUKvwgdw4FMn-O9G0clEClprlRhqIKNsE95Xbulp8jfHSz4g_3Snl6Isi9kNw_rLR0ZtjJf8FsS4esS6B-XfCxNztdT6WBZKjFfkPHG95JYM",
		},
	}

	for _, d := range sampleDishes {
		if _, err := s.store.CreateDish(ctx, d); err != nil {
			return fmt.Errorf("failed to insert sample dish: %w", err)
		}
	}

	return nil
}

// ---------------- LOYALTY BUSINESS LOGIC ----------------

// LoyaltyService implements loyalty coin business rules:
// 1. Accrual: Anonymous/by CPF on purchase (1 coin per R$ 10).
// 2. Redemption: Strictly requires an active registered account (hasAccount == true).
type LoyaltyService struct {
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
		return 0, errors.Join(ErrValidation, fmt.Errorf("insufficient coins: balance %d, requested %d", currentBalance, coinsToRedeem))
	}

	s.balances[cpf] -= coinsToRedeem
	discountAmount = float64(coinsToRedeem) * 1.00
	return discountAmount, nil
}

func (s *Service) Loyalty() *LoyaltyService {
	return s.loyalty
}
