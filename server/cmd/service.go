package main

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"os"
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrValidation                    = errors.New("validation failed")
	ErrAccountRequiredForRedemption = errors.New("account is required to redeem coins")
	ErrInsufficientCoins             = errors.New("insufficient loyalty coins")
	ErrInvalidCredentials           = errors.New("invalid email or password")
	ErrUnauthorized                 = errors.New("unauthorized")
	ErrForbidden                    = errors.New("forbidden: insufficient authority")
	ErrCNPJMismatch                 = errors.New("owner CNPJ does not match the logical restaurant registration")
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
		ImageURL:       input.ImageURL,
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

// ---------------- AUTH SERVICE ----------------

type RegisterInput struct {
	Name     string `json:"name,omitempty"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Phone    string `json:"phone,omitempty"`
}

type AccountClaims struct {
	AccountID int    `json:"account_id"`
	Email     string `json:"email"`
	jwt.RegisteredClaims
}

type AuthService struct {
	store     AccountStore
	jwtSecret []byte
}

func NewAuthService(store AccountStore) *AuthService {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "pratello-default-jwt-secret-key-change-in-prod"
	}
	return &AuthService{
		store:     store,
		jwtSecret: []byte(secret),
	}
}

func (s *AuthService) ValidateRegisterInput(input RegisterInput) error {
	email := strings.TrimSpace(input.Email)
	if email == "" {
		return errors.Join(ErrValidation, errors.New("e-mail é obrigatório"))
	}
	if _, err := mail.ParseAddress(email); err != nil {
		return errors.Join(ErrValidation, errors.New("formato de e-mail inválido"))
	}
	if len(input.Password) < 6 {
		return errors.Join(ErrValidation, errors.New("a senha deve ter no mínimo 6 caracteres"))
	}
	return nil
}

func (s *AuthService) Register(ctx context.Context, input RegisterInput) (*Account, error) {
	if err := s.ValidateRegisterInput(input); err != nil {
		return nil, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	return s.store.CreateAccount(ctx, strings.TrimSpace(input.Name), input.Email, string(hash), input.Phone)
}

func (s *AuthService) Login(ctx context.Context, email, password string) (string, *Account, error) {
	email = strings.TrimSpace(email)
	if email == "" || password == "" {
		return "", nil, ErrInvalidCredentials
	}

	acct, err := s.store.GetAccountByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return "", nil, ErrInvalidCredentials
		}
		return "", nil, fmt.Errorf("database error retrieving account: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(acct.Password), []byte(password)); err != nil {
		return "", nil, ErrInvalidCredentials
	}

	token, err := s.GenerateToken(acct)
	if err != nil {
		return "", nil, err
	}

	return token, acct, nil
}

func (s *AuthService) GenerateToken(account *Account) (string, error) {
	claims := AccountClaims{
		AccountID: account.ID,
		Email:     account.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "pratello-api",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}

func (s *AuthService) ParseToken(tokenStr string) (*AccountClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &AccountClaims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return s.jwtSecret, nil
	})
	if err != nil {
		return nil, ErrUnauthorized
	}

	claims, ok := token.Claims.(*AccountClaims)
	if !ok || !token.Valid {
		return nil, ErrUnauthorized
	}

	return claims, nil
}

func (s *AuthService) GetAccountByID(ctx context.Context, id int) (*Account, error) {
	return s.store.GetAccountByID(ctx, id)
}

func (s *AuthService) ListSysadmins(ctx context.Context) ([]Sysadmin, error) {
	return s.store.ListSysadmins(ctx)
}

func (s *AuthService) AddSysadmin(ctx context.Context, accountID int) (*Sysadmin, error) {
	if accountID <= 0 {
		return nil, errors.Join(ErrValidation, errors.New("invalid account_id"))
	}
	return s.store.AddSysadmin(ctx, accountID)
}

func (s *AuthService) ListManagers(ctx context.Context, restaurantSlug string) ([]Manager, error) {
	restaurantSlug = strings.ToLower(strings.TrimSpace(restaurantSlug))
	if restaurantSlug == "" {
		return nil, errors.Join(ErrValidation, errors.New("restaurant slug is required"))
	}
	return s.store.ListManagers(ctx, restaurantSlug)
}

func (s *AuthService) AddManager(ctx context.Context, restaurantSlug string, accountID int) (*Manager, error) {
	restaurantSlug = strings.ToLower(strings.TrimSpace(restaurantSlug))
	if restaurantSlug == "" {
		return nil, errors.Join(ErrValidation, errors.New("restaurant slug is required"))
	}
	if accountID <= 0 {
		return nil, errors.Join(ErrValidation, errors.New("invalid account_id"))
	}
	return s.store.AddManager(ctx, restaurantSlug, accountID)
}

func (s *AuthService) RemoveManager(ctx context.Context, restaurantSlug string, accountID int) error {
	restaurantSlug = strings.ToLower(strings.TrimSpace(restaurantSlug))
	if restaurantSlug == "" {
		return errors.Join(ErrValidation, errors.New("restaurant slug is required"))
	}
	if accountID <= 0 {
		return errors.Join(ErrValidation, errors.New("invalid account_id"))
	}
	return s.store.RemoveManager(ctx, restaurantSlug, accountID)
}

// ---------------- RBAC & AUTHORITY CHECKS ----------------

func (s *AuthService) GetAccountRoles(ctx context.Context, restaurantSlug string, accountID int) ([]string, error) {
	return s.store.GetAccountRoles(ctx, restaurantSlug, accountID)
}

func (s *AuthService) RequireOwner(ctx context.Context, restaurantSlug string, accountID int) error {
	roles, err := s.store.GetAccountRoles(ctx, restaurantSlug, accountID)
	if err != nil {
		return err
	}
	for _, r := range roles {
		if r == "owner" || r == "sysadmin" {
			return nil
		}
	}
	return ErrForbidden
}

func (s *AuthService) RequireManagerOrOwner(ctx context.Context, restaurantSlug string, accountID int) error {
	roles, err := s.store.GetAccountRoles(ctx, restaurantSlug, accountID)
	if err != nil {
		return err
	}
	for _, r := range roles {
		if r == "owner" || r == "manager" || r == "sysadmin" {
			return nil
		}
	}
	return ErrForbidden
}

// ---------------- OWNER ENTITY AUTHORITY ----------------

func (s *AuthService) ListOwners(ctx context.Context, restaurantSlug string) ([]Owner, error) {
	restaurantSlug = strings.ToLower(strings.TrimSpace(restaurantSlug))
	if restaurantSlug == "" {
		return nil, errors.Join(ErrValidation, errors.New("restaurant slug is required"))
	}
	return s.store.ListOwners(ctx, restaurantSlug)
}

func (s *AuthService) PromoteOwner(ctx context.Context, restaurantSlug string, requesterID int, targetAccountID int) (*Owner, error) {
	if err := s.RequireOwner(ctx, restaurantSlug, requesterID); err != nil {
		return nil, err
	}
	if targetAccountID <= 0 {
		return nil, errors.Join(ErrValidation, errors.New("invalid account_id to promote"))
	}
	return s.store.AddOwner(ctx, restaurantSlug, targetAccountID)
}

func (s *AuthService) RemoveOwner(ctx context.Context, restaurantSlug string, requesterID int, targetAccountID int) error {
	if err := s.RequireOwner(ctx, restaurantSlug, requesterID); err != nil {
		return err
	}
	if targetAccountID <= 0 {
		return errors.Join(ErrValidation, errors.New("invalid account_id"))
	}
	return s.store.RemoveOwner(ctx, restaurantSlug, targetAccountID)
}

// ---------------- PHYSICAL RESTAURANTS & CNPJ VERIFICATION ----------------

func (s *AuthService) CreatePhysicalRestaurant(ctx context.Context, restaurantSlug string, requesterID int, name, cnpj, ownerCnpj string) (*PhysicalRestaurant, error) {
	if err := s.RequireManagerOrOwner(ctx, restaurantSlug, requesterID); err != nil {
		return nil, err
	}

	name = strings.TrimSpace(name)
	cnpj = strings.TrimSpace(cnpj)
	ownerCnpj = strings.TrimSpace(ownerCnpj)
	if name == "" || cnpj == "" || ownerCnpj == "" {
		return nil, errors.Join(ErrValidation, errors.New("name, cnpj, and owner_cnpj are all required"))
	}

	registeredCNPJ, err := s.store.GetRestaurantCNPJ(ctx, restaurantSlug)
	if err == nil && registeredCNPJ != "" {
		cleanRegistered := strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(registeredCNPJ, ".", ""), "-", ""), "/", "")
		cleanOwner := strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(ownerCnpj, ".", ""), "-", ""), "/", "")
		if cleanRegistered != cleanOwner {
			return nil, ErrCNPJMismatch
		}
	}

	return s.store.CreatePhysicalRestaurant(ctx, restaurantSlug, name, cnpj, ownerCnpj)
}

func (s *AuthService) ListPhysicalRestaurants(ctx context.Context, restaurantSlug string) ([]PhysicalRestaurant, error) {
	restaurantSlug = strings.ToLower(strings.TrimSpace(restaurantSlug))
	return s.store.ListPhysicalRestaurants(ctx, restaurantSlug)
}

// ---------------- EMPLOYEES (HR RECORDS) ----------------

func (s *AuthService) ListEmployees(ctx context.Context, branchID int, requesterID int) ([]Employee, error) {
	pr, err := s.store.GetPhysicalRestaurant(ctx, branchID)
	if err != nil {
		return nil, err
	}
	if err := s.RequireManagerOrOwner(ctx, pr.RestaurantSlug, requesterID); err != nil {
		return nil, err
	}
	return s.store.ListEmployees(ctx, branchID)
}

func (s *AuthService) CreateEmployee(ctx context.Context, branchID int, requesterID int, name, cpf, position string) (*Employee, error) {
	pr, err := s.store.GetPhysicalRestaurant(ctx, branchID)
	if err != nil {
		return nil, err
	}
	if err := s.RequireManagerOrOwner(ctx, pr.RestaurantSlug, requesterID); err != nil {
		return nil, err
	}

	name = strings.TrimSpace(name)
	cpf = strings.TrimSpace(cpf)
	position = strings.TrimSpace(position)
	if name == "" || cpf == "" || position == "" {
		return nil, errors.Join(ErrValidation, errors.New("name, cpf and position are required"))
	}

	return s.store.CreateEmployee(ctx, branchID, name, cpf, position)
}

func (s *AuthService) InviteEmployee(ctx context.Context, branchID, employeeID int, requesterID int, targetAccountID int) (*Employee, error) {
	pr, err := s.store.GetPhysicalRestaurant(ctx, branchID)
	if err != nil {
		return nil, err
	}
	if err := s.RequireManagerOrOwner(ctx, pr.RestaurantSlug, requesterID); err != nil {
		return nil, err
	}
	return s.store.LinkEmployeeAccount(ctx, employeeID, targetAccountID)
}

// ---------------- SHARED 6-DIGIT TERMINAL SHIFT SESSION ----------------

func (s *AuthService) OpenTerminalSession(ctx context.Context, branchID int, pin string, deviceID string) (*TerminalSession, error) {
	pin = strings.TrimSpace(pin)
	deviceID = strings.TrimSpace(deviceID)
	if len(pin) != 6 {
		return nil, errors.Join(ErrValidation, errors.New("pin must be exactly 6 digits"))
	}
	if deviceID == "" {
		deviceID = "terminal-device-default"
	}
	return s.store.VerifyAndOpenTerminalSession(ctx, branchID, pin, deviceID)
}

func (s *AuthService) VerifyTerminalSession(ctx context.Context, token string, branchID int) (*TerminalSession, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return nil, ErrUnauthorized
	}
	session, err := s.store.GetTerminalSession(ctx, token)
	if err != nil {
		return nil, ErrUnauthorized
	}
	if session.PhysicalRestaurantID != branchID {
		return nil, ErrForbidden
	}
	return session, nil
}


