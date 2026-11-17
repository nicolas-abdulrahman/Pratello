package auth

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// -------------------------------------------------------------
// Domain Errors
// -------------------------------------------------------------

var (
	ErrInvalidCredentials    = errors.New("e-mail ou senha inválidos")
	ErrNotFound              = errors.New("recurso não encontrado")
	ErrConflict              = errors.New("conflito de dados (registro já existente)")
	ErrValidation            = errors.New("erro de validação dos dados")
	ErrUnauthorized          = errors.New("não autorizado")
	ErrForbidden             = errors.New("acesso proibido para este nível de privilégio")
	ErrTerminalLocked        = errors.New("terminal bloqueado temporariamente por excesso de tentativas")
	ErrInvalidPIN            = errors.New("PIN incorreto")
	ErrCannotRemoveLastOwner = errors.New("não é possível remover o último proprietário (proteção contra orfandade)")
)

// -------------------------------------------------------------
// Consumer-defined Store Interface
// -------------------------------------------------------------

type authStore interface {
	CreateAccount(ctx context.Context, name, email, passwordHash, phone string) (*Account, error)
	GetAccountByEmail(ctx context.Context, email string) (*Account, error)
	GetAccountByID(ctx context.Context, id int) (*Account, error)

	ListSysadmins(ctx context.Context) ([]Sysadmin, error)
	AddSysadmin(ctx context.Context, accountID int) (*Sysadmin, error)

	ListManagers(ctx context.Context, restaurantSlug string) ([]Manager, error)
	AddManager(ctx context.Context, restaurantSlug string, accountID int) (*Manager, error)
	RemoveManager(ctx context.Context, restaurantSlug string, accountID int) error

	ListOwners(ctx context.Context, restaurantSlug string) ([]Owner, error)
	AddOwner(ctx context.Context, restaurantSlug string, accountID int) (*Owner, error)
	RemoveOwner(ctx context.Context, restaurantSlug string, accountID int) error
	GetAccountRoles(ctx context.Context, restaurantSlug string, accountID int) ([]string, error)

	CreatePhysicalRestaurant(ctx context.Context, restaurantSlug, name, cnpj, ownerCnpj string) (*PhysicalRestaurant, error)
	ListPhysicalRestaurants(ctx context.Context, restaurantSlug string) ([]PhysicalRestaurant, error)
	GetPhysicalRestaurant(ctx context.Context, branchID int) (*PhysicalRestaurant, error)
	GetRestaurantCNPJ(ctx context.Context, restaurantSlug string) (string, error)

	ListEmployees(ctx context.Context, branchID int) ([]Employee, error)
	CreateEmployee(ctx context.Context, branchID int, name, cpf, position string) (*Employee, error)
	LinkEmployeeAccount(ctx context.Context, employeeID int, accountID int) (*Employee, error)

	VerifyAndOpenTerminalSession(ctx context.Context, branchID int, pin string, deviceID string) (*TerminalSession, error)
	GetTerminalSession(ctx context.Context, token string) (*TerminalSession, error)
	RotateTerminalPIN(ctx context.Context, branchID int, newPIN string) error
	GetTerminalOrders(ctx context.Context, branchID int) ([]TerminalOrder, error)
}

// -------------------------------------------------------------
// Domain Types & Claims
// -------------------------------------------------------------

type AccountClaims struct {
	AccountID int
	Email     string
	jwt.RegisteredClaims
}

type User struct {
	ID        int
	Name      string
	Email     string
	Phone     string
	CreatedAt time.Time
}

type RegisterInput struct {
	Name     string
	Email    string
	Password string
	Phone    string
}

// -------------------------------------------------------------
// Service
// -------------------------------------------------------------

type Service struct {
	store     authStore
	jwtSecret []byte
	hashCost  int
}

func NewService(store authStore) *Service {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "pratello-default-jwt-secret-key-change-in-prod"
	}
	return &Service{
		store:     store,
		jwtSecret: []byte(secret),
		hashCost:  bcrypt.DefaultCost,
	}
}

// SetHashCost allows tests to use bcrypt.MinCost for maximum execution speed.
func (s *Service) SetHashCost(cost int) {
	s.hashCost = cost
}

func (s *Service) ValidateRegisterInput(input RegisterInput) error {
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

func (s *Service) Register(ctx context.Context, input RegisterInput) (*User, error) {
	if err := s.ValidateRegisterInput(input); err != nil {
		return nil, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), s.hashCost)
	if err != nil {
		return nil, fmt.Errorf("falha ao criptografar senha: %w", err)
	}

	acct, err := s.store.CreateAccount(ctx, strings.TrimSpace(input.Name), input.Email, string(hash), input.Phone)
	if err != nil {
		return nil, err
	}

	return &User{
		ID:        acct.ID,
		Name:      acct.Name,
		Email:     acct.Email,
		Phone:     acct.Phone,
		CreatedAt: acct.CreatedAt,
	}, nil
}

func (s *Service) Login(ctx context.Context, email, password string) (string, *User, error) {
	email = strings.TrimSpace(email)
	if email == "" || password == "" {
		return "", nil, ErrInvalidCredentials
	}

	acct, err := s.store.GetAccountByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return "", nil, ErrInvalidCredentials
		}
		return "", nil, fmt.Errorf("erro no banco ao buscar conta: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(acct.Password), []byte(password)); err != nil {
		return "", nil, ErrInvalidCredentials
	}

	token, err := s.GenerateToken(acct.ID, acct.Email)
	if err != nil {
		return "", nil, fmt.Errorf("falha ao gerar token de acesso: %w", err)
	}

	user := &User{
		ID:        acct.ID,
		Name:      acct.Name,
		Email:     acct.Email,
		Phone:     acct.Phone,
		CreatedAt: acct.CreatedAt,
	}

	return token, user, nil
}

func (s *Service) GenerateToken(accountID int, email string) (string, error) {
	claims := AccountClaims{
		AccountID: accountID,
		Email:     email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}

func (s *Service) ParseToken(tokenStr string) (*AccountClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &AccountClaims{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("método de assinatura inesperado: %v", token.Header["alg"])
		}
		return s.jwtSecret, nil
	})
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*AccountClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, ErrUnauthorized
}

func (s *Service) GetAccountByID(ctx context.Context, id int) (*User, error) {
	acct, err := s.store.GetAccountByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return &User{
		ID:        acct.ID,
		Name:      acct.Name,
		Email:     acct.Email,
		Phone:     acct.Phone,
		CreatedAt: acct.CreatedAt,
	}, nil
}

func (s *Service) ListSysadmins(ctx context.Context) ([]Sysadmin, error) {
	return s.store.ListSysadmins(ctx)
}

func (s *Service) AddSysadmin(ctx context.Context, accountID int) (*Sysadmin, error) {
	return s.store.AddSysadmin(ctx, accountID)
}

func (s *Service) ListManagers(ctx context.Context, restaurantSlug string) ([]Manager, error) {
	return s.store.ListManagers(ctx, restaurantSlug)
}

func (s *Service) AddManager(ctx context.Context, restaurantSlug string, accountID int) (*Manager, error) {
	return s.store.AddManager(ctx, restaurantSlug, accountID)
}

func (s *Service) RemoveManager(ctx context.Context, restaurantSlug string, accountID int) error {
	return s.store.RemoveManager(ctx, restaurantSlug, accountID)
}

func (s *Service) ListOwners(ctx context.Context, restaurantSlug string) ([]Owner, error) {
	return s.store.ListOwners(ctx, restaurantSlug)
}

func (s *Service) AddOwner(ctx context.Context, restaurantSlug string, accountID int) (*Owner, error) {
	return s.store.AddOwner(ctx, restaurantSlug, accountID)
}

func (s *Service) RemoveOwner(ctx context.Context, restaurantSlug string, accountID int) error {
	return s.store.RemoveOwner(ctx, restaurantSlug, accountID)
}

func (s *Service) CheckRole(ctx context.Context, restaurantSlug string, accountID int, requiredRole string) (bool, error) {
	roles, err := s.store.GetAccountRoles(ctx, restaurantSlug, accountID)
	if err != nil {
		return false, err
	}
	for _, r := range roles {
		if strings.EqualFold(r, requiredRole) || strings.EqualFold(r, "sysadmin") {
			return true, nil
		}
	}
	return false, nil
}

func (s *Service) CheckOwnerOrManager(ctx context.Context, restaurantSlug string, accountID int) (bool, error) {
	roles, err := s.store.GetAccountRoles(ctx, restaurantSlug, accountID)
	if err != nil {
		return false, err
	}
	for _, r := range roles {
		rLower := strings.ToLower(r)
		if rLower == "owner" || rLower == "manager" || rLower == "sysadmin" {
			return true, nil
		}
	}
	return false, nil
}

func (s *Service) ListEmployees(ctx context.Context, branchID int) ([]Employee, error) {
	return s.store.ListEmployees(ctx, branchID)
}

func (s *Service) CreateEmployee(ctx context.Context, branchID int, name, cpf, position string) (*Employee, error) {
	if strings.TrimSpace(name) == "" || strings.TrimSpace(cpf) == "" {
		return nil, errors.Join(ErrValidation, errors.New("nome e CPF do colaborador são obrigatórios"))
	}
	return s.store.CreateEmployee(ctx, branchID, strings.TrimSpace(name), strings.TrimSpace(cpf), strings.TrimSpace(position))
}

func (s *Service) LinkEmployeeAccount(ctx context.Context, employeeID, accountID int) (*Employee, error) {
	if _, err := s.store.GetAccountByID(ctx, accountID); err != nil {
		return nil, err
	}
	return s.store.LinkEmployeeAccount(ctx, employeeID, accountID)
}

func (s *Service) OpenTerminalSession(ctx context.Context, branchID int, pin, deviceID string) (*TerminalSession, error) {
	if len(pin) != 6 {
		return nil, errors.Join(ErrValidation, errors.New("o PIN deve conter exatamente 6 dígitos"))
	}
	if strings.TrimSpace(deviceID) == "" {
		return nil, errors.Join(ErrValidation, errors.New("identificador do dispositivo é obrigatório"))
	}
	return s.store.VerifyAndOpenTerminalSession(ctx, branchID, pin, deviceID)
}

func (s *Service) ValidateTerminalToken(ctx context.Context, token, deviceID string) (*TerminalSession, error) {
	ts, err := s.store.GetTerminalSession(ctx, token)
	if err != nil {
		return nil, ErrUnauthorized
	}
	if ts.DeviceID != deviceID {
		return nil, ErrForbidden
	}
	return ts, nil
}

func (s *Service) GetTerminalOrders(ctx context.Context, branchID int) ([]TerminalOrder, error) {
	return s.store.GetTerminalOrders(ctx, branchID)
}
