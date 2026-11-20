package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestService_Register_ValidationErrors(t *testing.T) {
	mock := NewMemoryStore()
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

func TestHandler_Register_Success(t *testing.T) {
	mock := NewMemoryStore()
	svc := NewService(mock)
	svc.SetHashCost(bcrypt.MinCost)
	svc.verification = newTestVerification()
	h := NewHandler(svc)

	mux := http.NewServeMux()
	h.Routes(mux)

	payload, _ := json.Marshal(map[string]string{
		"name":     "Carlos Teste",
		"email":    "carlos@pratello.com",
		"password": "validPassword123",
		"code":     fixedCode,
	})

	requestTestEmail(t, mux, "carlos@pratello.com")
	req := httptest.NewRequest("POST", "/api/auth/register", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandler_Register_DuplicateConflict(t *testing.T) {
	mock := NewMemoryStore()
	svc := NewService(mock)
	svc.SetHashCost(bcrypt.MinCost)
	svc.verification = newTestVerification()
	h := NewHandler(svc)

	mux := http.NewServeMux()
	h.Routes(mux)

	payload, _ := json.Marshal(map[string]string{
		"email":    "dup@pratello.com",
		"password": "validPassword123",
		"code":     fixedCode,
	})

	// First register
	requestTestEmail(t, mux, "dup@pratello.com")
	w1 := httptest.NewRecorder()
	mux.ServeHTTP(w1, httptest.NewRequest("POST", "/api/auth/register", bytes.NewReader(payload)))
	if w1.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created on first, got %d", w1.Code)
	}

	// Duplicate register -> 409
	requestTestEmail(t, mux, "dup@pratello.com")
	w2 := httptest.NewRecorder()
	mux.ServeHTTP(w2, httptest.NewRequest("POST", "/api/auth/register", bytes.NewReader(payload)))
	if w2.Code != http.StatusConflict {
		t.Fatalf("expected 409 Conflict on duplicate, got %d", w2.Code)
	}
}

func TestHandler_Login_WrongPassword(t *testing.T) {
	mock := NewMemoryStore()
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

type testVerification map[string]string

func newTestVerification() testVerification { return make(testVerification) }
func (m testVerification) put(_ context.Context, k, v string, _ time.Duration) error {
	m[k] = v
	return nil
}
func (m testVerification) consume(_ context.Context, k, v string) (bool, error) {
	if m[k] != v {
		return false, nil
	}
	delete(m, k)
	return true, nil
}
func requestTestEmail(t *testing.T, mux *http.ServeMux, email string) {
	t.Helper()
	b, _ := json.Marshal(map[string]string{"email": email})
	if r := idpRequest(mux, "POST", "/auth/register/email/request-code", string(b), ""); r.Code != 202 {
		t.Fatalf("request verification: %d %s", r.Code, r.Body)
	}
}

func TestPhoneRegistrationIssuesTokensAndRejectsDuplicate(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()
	svc := NewService(store)
	const phone = "+5511888888888"
	svc.verification = newTestVerification()
	if err := svc.RequestPhoneRegistration(ctx, "+5511888888888"); err != nil {
		t.Fatal(err)
	}
	if err := svc.verifyCode(ctx, "register-phone", 0, 0, phone, fixedCode); err != nil {
		t.Fatal(err)
	}
	pair, err := svc.RegisterPhone(ctx, phone, "Customer", "")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Authenticate(ctx, pair.AccessToken); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.RegisterPhone(ctx, phone, "", ""); !errors.Is(err, ErrConflict) {
		t.Fatalf("duplicate registration: %v", err)
	}
	if _, err := svc.RegisterPhone(ctx, "+5511777777777", "", "invalid"); !errors.Is(err, ErrValidation) {
		t.Fatalf("invalid CPF accepted: %v", err)
	}
}
