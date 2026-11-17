package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// -------------------------------------------------------------
// SECTION A: SERVICE TESTS (Local mockStore, fast, business rules)
// -------------------------------------------------------------

type mockStore struct {
	accounts map[string]*Account
	byID     map[int]*Account
	managers map[string][]Manager
	owners   map[string][]Owner
}

func newMockStore() *mockStore {
	return &mockStore{
		accounts: make(map[string]*Account),
		byID:     make(map[int]*Account),
		managers: make(map[string][]Manager),
		owners:   make(map[string][]Owner),
	}
}

func (m *mockStore) CreateAccount(ctx context.Context, name, email, passwordHash, phone string) (*Account, error) {
	if _, ok := m.accounts[email]; ok {
		return nil, ErrConflict
	}
	a := &Account{
		ID:        len(m.accounts) + 1,
		Name:      name,
		Email:     email,
		Password:  passwordHash,
		Phone:     phone,
		CreatedAt: time.Now(),
	}
	m.accounts[email] = a
	m.byID[a.ID] = a
	return a, nil
}

func (m *mockStore) GetAccountByEmail(ctx context.Context, email string) (*Account, error) {
	a, ok := m.accounts[email]
	if !ok {
		return nil, ErrNotFound
	}
	return a, nil
}

func (m *mockStore) GetAccountByID(ctx context.Context, id int) (*Account, error) {
	a, ok := m.byID[id]
	if !ok {
		return nil, ErrNotFound
	}
	return a, nil
}

func (m *mockStore) ListSysadmins(ctx context.Context) ([]Sysadmin, error) {
	return []Sysadmin{}, nil
}

func (m *mockStore) AddSysadmin(ctx context.Context, accountID int) (*Sysadmin, error) {
	return &Sysadmin{ID: 1, AccountID: accountID, CreatedAt: time.Now()}, nil
}

func (m *mockStore) ListManagers(ctx context.Context, restaurantSlug string) ([]Manager, error) {
	return m.managers[restaurantSlug], nil
}

func (m *mockStore) AddManager(ctx context.Context, restaurantSlug string, accountID int) (*Manager, error) {
	mgr := Manager{ID: len(m.managers[restaurantSlug]) + 1, RestaurantSlug: restaurantSlug, AccountID: accountID}
	m.managers[restaurantSlug] = append(m.managers[restaurantSlug], mgr)
	return &mgr, nil
}

func (m *mockStore) RemoveManager(ctx context.Context, restaurantSlug string, accountID int) error {
	list := m.managers[restaurantSlug]
	for i, mgr := range list {
		if mgr.AccountID == accountID {
			m.managers[restaurantSlug] = append(list[:i], list[i+1:]...)
			return nil
		}
	}
	return ErrNotFound
}

func (m *mockStore) ListOwners(ctx context.Context, restaurantSlug string) ([]Owner, error) {
	return m.owners[restaurantSlug], nil
}

func (m *mockStore) AddOwner(ctx context.Context, restaurantSlug string, accountID int) (*Owner, error) {
	own := Owner{ID: len(m.owners[restaurantSlug]) + 1, RestaurantSlug: restaurantSlug, AccountID: accountID}
	m.owners[restaurantSlug] = append(m.owners[restaurantSlug], own)
	return &own, nil
}

func (m *mockStore) RemoveOwner(ctx context.Context, restaurantSlug string, accountID int) error {
	list := m.owners[restaurantSlug]
	if len(list) <= 1 {
		return ErrCannotRemoveLastOwner
	}
	for i, own := range list {
		if own.AccountID == accountID {
			m.owners[restaurantSlug] = append(list[:i], list[i+1:]...)
			return nil
		}
	}
	return ErrNotFound
}

func (m *mockStore) GetAccountRoles(ctx context.Context, restaurantSlug string, accountID int) ([]string, error) {
	var roles []string
	for _, o := range m.owners[restaurantSlug] {
		if o.AccountID == accountID {
			roles = append(roles, "owner")
		}
	}
	for _, mgr := range m.managers[restaurantSlug] {
		if mgr.AccountID == accountID {
			roles = append(roles, "manager")
		}
	}
	return roles, nil
}

func (m *mockStore) CreatePhysicalRestaurant(ctx context.Context, restaurantSlug, name, cnpj, ownerCnpj string) (*PhysicalRestaurant, error) {
	return &PhysicalRestaurant{ID: 1, RestaurantSlug: restaurantSlug, Name: name, CNPJ: cnpj}, nil
}

func (m *mockStore) ListPhysicalRestaurants(ctx context.Context, restaurantSlug string) ([]PhysicalRestaurant, error) {
	return []PhysicalRestaurant{}, nil
}

func (m *mockStore) GetPhysicalRestaurant(ctx context.Context, branchID int) (*PhysicalRestaurant, error) {
	return &PhysicalRestaurant{ID: branchID, TerminalPIN: "123456"}, nil
}

func (m *mockStore) GetRestaurantCNPJ(ctx context.Context, restaurantSlug string) (string, error) {
	return "12.345.678/0001-90", nil
}

func (m *mockStore) ListEmployees(ctx context.Context, branchID int) ([]Employee, error) {
	return []Employee{}, nil
}

func (m *mockStore) CreateEmployee(ctx context.Context, branchID int, name, cpf, position string) (*Employee, error) {
	return &Employee{ID: 1, PhysicalRestaurantID: branchID, Name: name, CPF: cpf, Position: position}, nil
}

func (m *mockStore) LinkEmployeeAccount(ctx context.Context, employeeID int, accountID int) (*Employee, error) {
	return &Employee{ID: employeeID, AccountID: &accountID}, nil
}

func (m *mockStore) VerifyAndOpenTerminalSession(ctx context.Context, branchID int, pin string, deviceID string) (*TerminalSession, error) {
	if pin != "123456" {
		return nil, ErrInvalidPIN
	}
	return &TerminalSession{ID: 1, PhysicalRestaurantID: branchID, Token: "token_123", DeviceID: deviceID}, nil
}

func (m *mockStore) GetTerminalSession(ctx context.Context, token string) (*TerminalSession, error) {
	if token == "token_123" {
		return &TerminalSession{ID: 1, Token: token, DeviceID: "device_abc", ExpiresAt: time.Now().Add(time.Hour)}, nil
	}
	return nil, ErrNotFound
}

func (m *mockStore) RotateTerminalPIN(ctx context.Context, branchID int, newPIN string) error {
	return nil
}

func (m *mockStore) GetTerminalOrders(ctx context.Context, branchID int) ([]TerminalOrder, error) {
	return []TerminalOrder{}, nil
}

func TestService_Register_ValidationErrors(t *testing.T) {
	mock := newMockStore()
	svc := NewService(mock)
	svc.SetHashCost(bcrypt.MinCost)

	// Short password
	_, err := svc.Register(context.Background(), RegisterInput{
		Email:    "test@pratello.com",
		Password: "123",
	})
	if !errors.Is(err, ErrValidation) {
		t.Fatalf("expected ErrValidation for short password, got %v", err)
	}

	// Invalid email
	_, err = svc.Register(context.Background(), RegisterInput{
		Email:    "invalid-email",
		Password: "password123",
	})
	if !errors.Is(err, ErrValidation) {
		t.Fatalf("expected ErrValidation for invalid email, got %v", err)
	}
}

func TestService_Register_And_Login_Success(t *testing.T) {
	mock := newMockStore()
	svc := NewService(mock)
	svc.SetHashCost(bcrypt.MinCost)

	user, err := svc.Register(context.Background(), RegisterInput{
		Name:     "Alice Silva",
		Email:    "alice@pratello.com",
		Password: "password123",
		Phone:    "11988887777",
	})
	if err != nil {
		t.Fatalf("unexpected register error: %v", err)
	}
	if user.Email != "alice@pratello.com" {
		t.Fatalf("unexpected user email: %s", user.Email)
	}

	// Login with wrong password
	_, _, err = svc.Login(context.Background(), "alice@pratello.com", "wrongpass")
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("expected ErrInvalidCredentials for wrong password, got %v", err)
	}

	// Login with correct password
	token, loginUser, err := svc.Login(context.Background(), "alice@pratello.com", "password123")
	if err != nil {
		t.Fatalf("unexpected login error: %v", err)
	}
	if token == "" || loginUser.ID != user.ID {
		t.Fatalf("invalid login response: token=%s, id=%d", token, loginUser.ID)
	}
}

func TestService_Owners_OrphanPrevention(t *testing.T) {
	mock := newMockStore()
	svc := NewService(mock)
	svc.SetHashCost(bcrypt.MinCost)

	// Add 1 owner
	_, _ = svc.AddOwner(context.Background(), "trattoria", 10)

	// Remove single owner -> must fail with ErrCannotRemoveLastOwner
	err := svc.RemoveOwner(context.Background(), "trattoria", 10)
	if !errors.Is(err, ErrCannotRemoveLastOwner) {
		t.Fatalf("expected ErrCannotRemoveLastOwner, got %v", err)
	}

	// Add second owner -> removing one should succeed
	_, _ = svc.AddOwner(context.Background(), "trattoria", 20)
	if err := svc.RemoveOwner(context.Background(), "trattoria", 10); err != nil {
		t.Fatalf("expected remove owner to succeed when multiple exist, got %v", err)
	}
}

// -------------------------------------------------------------
// SECTION B: STORE TESTS (Covering memory store & interface)
// -------------------------------------------------------------

func TestStore_CreateAccount_Conflict(t *testing.T) {
	store := NewMemoryStore()

	_, err := store.CreateAccount(context.Background(), "Bob", "bob@pratello.com", "hash", "11999999999")
	if err != nil {
		t.Fatalf("unexpected error creating account: %v", err)
	}

	// Duplicate
	_, err = store.CreateAccount(context.Background(), "Bob", "bob@pratello.com", "hash", "11999999999")
	if !errors.Is(err, ErrConflict) {
		t.Fatalf("expected ErrConflict for duplicate email, got %v", err)
	}
}

func TestStore_GetAccountByEmail_NotFound(t *testing.T) {
	store := NewMemoryStore()

	_, err := store.GetAccountByEmail(context.Background(), "unknown@pratello.com")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

// -------------------------------------------------------------
// SECTION C: HTTP TESTS (Covering Handler endpoints and error mapping)
// -------------------------------------------------------------

func TestHandler_Register_Success(t *testing.T) {
	mock := newMockStore()
	svc := NewService(mock)
	svc.SetHashCost(bcrypt.MinCost)
	h := NewHandler(svc)

	mux := http.NewServeMux()
	h.Routes(mux)

	payload, _ := json.Marshal(map[string]string{
		"name":     "Carlos Teste",
		"email":    "carlos@pratello.com",
		"password": "validPassword123",
	})

	req := httptest.NewRequest("POST", "/api/auth/register", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandler_Register_DuplicateConflict(t *testing.T) {
	mock := newMockStore()
	svc := NewService(mock)
	svc.SetHashCost(bcrypt.MinCost)
	h := NewHandler(svc)

	mux := http.NewServeMux()
	h.Routes(mux)

	payload, _ := json.Marshal(map[string]string{
		"email":    "dup@pratello.com",
		"password": "validPassword123",
	})

	// First register
	w1 := httptest.NewRecorder()
	mux.ServeHTTP(w1, httptest.NewRequest("POST", "/api/auth/register", bytes.NewReader(payload)))
	if w1.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created on first, got %d", w1.Code)
	}

	// Duplicate register -> 409
	w2 := httptest.NewRecorder()
	mux.ServeHTTP(w2, httptest.NewRequest("POST", "/api/auth/register", bytes.NewReader(payload)))
	if w2.Code != http.StatusConflict {
		t.Fatalf("expected 409 Conflict on duplicate, got %d", w2.Code)
	}
}

func TestHandler_Login_WrongPassword(t *testing.T) {
	mock := newMockStore()
	svc := NewService(mock)
	svc.SetHashCost(bcrypt.MinCost)
	h := NewHandler(svc)

	mux := http.NewServeMux()
	h.Routes(mux)

	_, _ = svc.Register(context.Background(), RegisterInput{
		Email:    "login_test@pratello.com",
		Password: "password123",
	})

	badLogin, _ := json.Marshal(map[string]string{
		"email":    "login_test@pratello.com",
		"password": "wrongpassword",
	})

	w := httptest.NewRecorder()
	mux.ServeHTTP(w, httptest.NewRequest("POST", "/api/auth/login", bytes.NewReader(badLogin)))
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 Unauthorized for bad login, got %d", w.Code)
	}
}

func TestHandler_Terminal_Session_And_Orders(t *testing.T) {
	mock := newMockStore()
	svc := NewService(mock)
	svc.SetHashCost(bcrypt.MinCost)
	h := NewHandler(svc)

	mux := http.NewServeMux()
	h.Routes(mux)

	// Open session with valid PIN
	sessionReq, _ := json.Marshal(map[string]string{
		"pin":       "123456",
		"device_id": "device_abc",
	})
	wSession := httptest.NewRecorder()
	mux.ServeHTTP(wSession, httptest.NewRequest("POST", "/api/restaurants/trattoria/branches/1/terminal/session", bytes.NewReader(sessionReq)))
	if wSession.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for terminal session, got %d: %s", wSession.Code, wSession.Body.String())
	}

	var sessionRes terminalSessionResponse
	_ = json.NewDecoder(wSession.Body).Decode(&sessionRes)
	if sessionRes.Token == "" {
		t.Fatalf("expected non-empty token")
	}

	// Read terminal orders
	reqOrders := httptest.NewRequest("GET", "/api/restaurants/trattoria/branches/1/terminal/orders", nil)
	reqOrders.Header.Set("Authorization", "Bearer token_123")
	reqOrders.Header.Set("X-Device-Id", "device_abc")
	wOrders := httptest.NewRecorder()
	mux.ServeHTTP(wOrders, reqOrders)
	if wOrders.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for terminal orders, got %d: %s", wOrders.Code, wOrders.Body.String())
	}
}
