package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestSessionRotationAndRevocation(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()
	a, err := store.CreateAccount(ctx, "", "user@example.com", "hash", "")
	if err != nil {
		t.Fatal(err)
	}
	svc := NewService(store)
	first, err := svc.IssueTokens(ctx, a.ID)
	if err != nil {
		t.Fatal(err)
	}
	second, err := svc.Refresh(ctx, first.RefreshToken)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Refresh(ctx, first.RefreshToken); !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("replayed refresh token: %v", err)
	}
	if _, err := svc.Authenticate(ctx, second.RefreshToken); !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("refresh accepted as access: %v", err)
	}
	if _, err := svc.Authenticate(ctx, first.AccessToken); err != nil {
		t.Fatal(err)
	}
	if err := svc.LogoutAll(ctx, a.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Authenticate(ctx, first.AccessToken); !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("old access token accepted: %v", err)
	}
	if _, err := svc.Refresh(ctx, second.RefreshToken); !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("revoked refresh accepted: %v", err)
	}
}

func TestHTTPRegistrationAndLogoutAll(t *testing.T) {
	svc := NewService(NewMemoryStore())
	svc.SetHashCost(bcrypt.MinCost)
	mux := http.NewServeMux()
	NewHandler(svc).Routes(mux)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, httptest.NewRequest("POST", "/auth/register", bytes.NewBufferString(`{"email":"user@example.com","password":"password123"}`)))
	if w.Code != http.StatusCreated {
		t.Fatalf("registration: %d %s", w.Code, w.Body.String())
	}
	var pair TokenPair
	if err := json.Unmarshal(w.Body.Bytes(), &pair); err != nil || pair.AccessToken == "" || pair.RefreshToken == "" {
		t.Fatalf("invalid token pair: %s (%v)", w.Body.String(), err)
	}
	req := httptest.NewRequest("POST", "/auth/logout-all", nil)
	req.Header.Set("Authorization", "Bearer "+pair.AccessToken)
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	if w.Code != http.StatusNoContent {
		t.Fatalf("logout-all: %d %s", w.Code, w.Body.String())
	}
	req = httptest.NewRequest("GET", "/auth/me", nil)
	req.Header.Set("Authorization", "Bearer "+pair.AccessToken)
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("old access token: %d", w.Code)
	}
}

func TestLogoutCannotRevokeAnotherAccount(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()
	svc := NewService(store)
	a, _ := store.CreateAccount(ctx, "", "a@example.com", "hash", "")
	b, _ := store.CreateAccount(ctx, "", "b@example.com", "hash", "")
	pair, err := svc.IssueTokens(ctx, a.ID)
	if err != nil {
		t.Fatal(err)
	}
	if err := svc.Logout(ctx, b.ID, pair.RefreshToken); !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("logout another account: %v", err)
	}
	if _, err := svc.Refresh(ctx, pair.RefreshToken); err != nil {
		t.Fatal(err)
	}
}
