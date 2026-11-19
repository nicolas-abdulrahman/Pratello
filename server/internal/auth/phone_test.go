package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
)

type recordingNotifier struct{ phone, code string }

func (n *recordingNotifier) SendOTP(phone, code string) error {
	n.phone, n.code = phone, code
	return nil
}

func TestPhoneRegistrationHTTPWithoutPIN(t *testing.T) {
	mux := http.NewServeMux()
	NewHandler(NewService(NewMemoryStore())).Routes(mux)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, httptest.NewRequest("POST", "/auth/register/phone/verify", bytes.NewBufferString(`{"phone":"+5511888888888"}`)))
	if w.Code != http.StatusCreated {
		t.Fatalf("register phone: %d %s", w.Code, w.Body.String())
	}
	var pair TokenPair
	if err := json.Unmarshal(w.Body.Bytes(), &pair); err != nil || pair.RefreshToken == "" {
		t.Fatalf("invalid response: %s", w.Body.String())
	}
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, httptest.NewRequest("POST", "/auth/login/phone/verify", bytes.NewBufferString(`{"phone":"+5511888888888"}`)))
	if w.Code != http.StatusOK {
		t.Fatalf("phone login: %d %s", w.Code, w.Body.String())
	}
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, httptest.NewRequest("POST", "/auth/login/phone/request-code", bytes.NewBufferString(`{"phone":"+5511777777777"}`)))
	if w.Code != http.StatusAccepted {
		t.Fatalf("unknown phone request: %d", w.Code)
	}
}
func (*recordingNotifier) SendEmail(string, string) error { return nil }

func TestPhoneLoginRequiresVerifiedPhone(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()
	svc := NewService(store)
	a, err := store.CreateAccount(ctx, "", "user@example.com", "hash", "+5511999999999")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.LoginPhone(ctx, a.Phone); !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("unverified phone login: %v", err)
	}
	n := &recordingNotifier{}
	svc.SetNotifier(n)
	updated, err := svc.AddPhone(ctx, a.ID, a.Phone)
	if err != nil {
		t.Fatal(err)
	}
	if !updated.PhoneVerified || n.phone != a.Phone || !regexp.MustCompile(`^[0-9]{6}$`).MatchString(n.code) {
		t.Fatalf("invalid notification or state: %#v %#v", updated, n)
	}
	pair, err := svc.LoginPhone(ctx, a.Phone)
	if err != nil || pair.AccessToken == "" {
		t.Fatalf("verified phone login: %v", err)
	}
	if err := svc.RemovePhone(ctx, a.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.LoginPhone(ctx, a.Phone); !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("removed phone login: %v", err)
	}
}

func TestPhoneConflictsAndLastMethod(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()
	svc := NewService(store)
	a, _ := store.CreateAccount(ctx, "", "a@example.com", "hash", "")
	b, _ := store.CreateAccount(ctx, "", "b@example.com", "hash", "")
	if _, err := svc.AddPhone(ctx, a.ID, "+5511999999999"); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.AddPhone(ctx, b.ID, "+5511999999999"); !errors.Is(err, ErrConflict) {
		t.Fatalf("duplicate phone: %v", err)
	}
	store.mu.Lock()
	store.accounts[a.ID].EmailVerified = false
	store.mu.Unlock()
	if err := svc.RemovePhone(ctx, a.ID); !errors.Is(err, ErrConflict) {
		t.Fatalf("removed last method: %v", err)
	}
	unchanged, err := store.GetAccountByID(ctx, a.ID)
	if err != nil || unchanged.Phone != "+5511999999999" || !unchanged.PhoneVerified {
		t.Fatalf("phone changed after conflict: %#v %v", unchanged, err)
	}
}

func TestPhoneRegistrationIssuesTokensAndRejectsDuplicate(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()
	svc := NewService(store)
	n := &recordingNotifier{}
	svc.SetNotifier(n)
	if err := svc.RequestPhoneRegistration(ctx, "+5511888888888"); err != nil {
		t.Fatal(err)
	}
	if n.phone != "+5511888888888" || !regexp.MustCompile(`^[0-9]{6}$`).MatchString(n.code) {
		t.Fatalf("invalid notification: %#v", n)
	}
	pair, err := svc.RegisterPhone(ctx, n.phone, "Customer", "")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Authenticate(ctx, pair.AccessToken); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.RegisterPhone(ctx, n.phone, "", ""); !errors.Is(err, ErrConflict) {
		t.Fatalf("duplicate registration: %v", err)
	}
	if _, err := svc.RegisterPhone(ctx, "+5511777777777", "", "invalid"); !errors.Is(err, ErrValidation) {
		t.Fatalf("invalid CPF accepted: %v", err)
	}
}
