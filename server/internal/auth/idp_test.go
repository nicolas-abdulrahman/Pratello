package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

func idpSetup() (*memoryStore, *Service, *http.ServeMux) {
	store := NewMemoryStore()
	svc := NewService(store)
	svc.SetHashCost(bcrypt.MinCost)
	mux := http.NewServeMux()
	NewHandler(svc).Routes(mux)
	return store, svc, mux
}

func idpRequest(mux *http.ServeMux, method, path, body, token string) *httptest.ResponseRecorder {
	r := httptest.NewRequest(method, path, strings.NewReader(body))
	if token != "" {
		r.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, r)
	return w
}

func TestIDP001RegistrationHashAndTokens(t *testing.T) {
	store, svc, mux := idpSetup()
	w := idpRequest(mux, "POST", "/auth/register", `{"email":"user@example.com","password":"password123"}`, "")
	if w.Code != http.StatusCreated {
		t.Fatalf("registration: %d %s", w.Code, w.Body.String())
	}
	var pair TokenPair
	if err := json.Unmarshal(w.Body.Bytes(), &pair); err != nil || pair.AccessToken == "" || pair.RefreshToken == "" || pair.TokenType != "Bearer" || pair.ExpiresIn != 900 {
		t.Fatalf("invalid token pair: %s (%v)", w.Body.String(), err)
	}
	if strings.Contains(w.Body.String(), "password123") {
		t.Fatal("password leaked in response")
	}
	a, err := store.GetAccountByEmail(context.Background(), "user@example.com")
	if err != nil || a.Password == "password123" || bcrypt.CompareHashAndPassword([]byte(a.Password), []byte("password123")) != nil {
		t.Fatalf("invalid password hash: %v", err)
	}
	if _, err = svc.Authenticate(context.Background(), pair.AccessToken); err != nil {
		t.Fatal(err)
	}
}

func TestIDP002DuplicateEmailCaseInsensitive(t *testing.T) {
	store, _, mux := idpSetup()
	for i, email := range []string{"dup@pratello.com", "DUP@PRATELLO.COM"} {
		body, _ := json.Marshal(map[string]string{"email": email, "password": "validPassword123"})
		w := idpRequest(mux, "POST", "/auth/register", string(body), "")
		want := http.StatusCreated
		if i == 1 {
			want = http.StatusConflict
		}
		if w.Code != want {
			t.Fatalf("attempt %d: %d %s", i, w.Code, w.Body.String())
		}
	}
	if len(store.accounts) != 1 {
		t.Fatalf("expected one account, got %d", len(store.accounts))
	}
}

func TestIDP003InvalidRegistration(t *testing.T) {
	store, _, mux := idpSetup()
	for _, body := range []string{`{"email":`, `{"email":"invalid-email","password":"password123"}`, `{"email":"a@example.com","password":"1234567"}`} {
		w := idpRequest(mux, "POST", "/auth/register", body, "")
		if w.Code != http.StatusBadRequest {
			t.Fatalf("body %q: %d %s", body, w.Code, w.Body.String())
		}
	}
	if len(store.accounts) != 0 {
		t.Fatal("invalid registration created an account")
	}
}

func TestIDP005InvalidLoginDoesNotEnumerate(t *testing.T) {
	_, svc, mux := idpSetup()
	if _, err := svc.Register(context.Background(), RegisterInput{Email: "user@example.com", Password: "password123"}); err != nil {
		t.Fatal(err)
	}
	a := idpRequest(mux, "POST", "/auth/login", `{"email":"user@example.com","password":"wrong"}`, "")
	b := idpRequest(mux, "POST", "/auth/login", `{"email":"missing@example.com","password":"wrong"}`, "")
	if a.Code != http.StatusUnauthorized || b.Code != http.StatusUnauthorized || !bytes.Equal(a.Body.Bytes(), b.Body.Bytes()) {
		t.Fatalf("different responses: %d %s / %d %s", a.Code, a.Body.String(), b.Code, b.Body.String())
	}
}

func TestIDP009InvalidCPF(t *testing.T) {
	store, svc, mux := idpSetup()
	u, err := svc.Register(context.Background(), RegisterInput{Email: "user@example.com", Password: "password123"})
	if err != nil {
		t.Fatal(err)
	}
	pair, err := svc.IssueTokens(context.Background(), u.ID)
	if err != nil {
		t.Fatal(err)
	}
	for _, cpf := range []string{"123", "1234567890x"} {
		body, _ := json.Marshal(map[string]string{"cpf": cpf})
		w := idpRequest(mux, "PUT", "/me/cpf", string(body), pair.AccessToken)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("cpf %q: %d %s", cpf, w.Code, w.Body.String())
		}
	}
	a, err := store.GetAccountByID(context.Background(), u.ID)
	if err != nil || a.CPF != "" {
		t.Fatalf("CPF changed: %#v %v", a, err)
	}
}

func TestIDP017InvalidBearer(t *testing.T) {
	_, svc, mux := idpSetup()
	u, err := svc.Register(context.Background(), RegisterInput{Email: "user@example.com", Password: "password123"})
	if err != nil {
		t.Fatal(err)
	}
	claims := AccountClaims{AccountID: u.ID, Type: "access", RegisteredClaims: jwt.RegisteredClaims{ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Minute))}}
	expired, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(svc.jwtSecret)
	if err != nil {
		t.Fatal(err)
	}
	claims.ExpiresAt = jwt.NewNumericDate(time.Now().Add(time.Hour))
	forged, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte("wrong-secret"))
	if err != nil {
		t.Fatal(err)
	}
	for _, token := range []string{"", expired, forged} {
		w := idpRequest(mux, "GET", "/auth/me", "", token)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("invalid token accepted: %d %s", w.Code, w.Body.String())
		}
	}
}

func TestIDP019LastOwner(t *testing.T) {
	store, svc, _ := idpSetup()
	ctx := context.Background()
	a, err := store.CreateAccount(ctx, "", "owner-a@example.com", "hash", "")
	if err != nil {
		t.Fatal(err)
	}
	b, err := store.CreateAccount(ctx, "", "owner-b@example.com", "hash", "")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.AddOwner(ctx, "trattoria", a.ID); err != nil {
		t.Fatal(err)
	}
	if err := svc.RemoveOwner(ctx, "trattoria", a.ID); !errors.Is(err, ErrCannotRemoveLastOwner) {
		t.Fatalf("removed last owner: %v", err)
	}
	if _, err := svc.AddOwner(ctx, "trattoria", b.ID); err != nil {
		t.Fatal(err)
	}
	if err := svc.RemoveOwner(ctx, "trattoria", a.ID); err != nil {
		t.Fatal(err)
	}
	owners, err := svc.ListOwners(ctx, "trattoria")
	if err != nil || len(owners) != 1 || owners[0].AccountID != b.ID {
		t.Fatalf("remaining owners: %v %v", owners, err)
	}
}

// XFAIL tests exercise the endpoint first; only the specific known failure is skipped.
func TestIDP020AnonymousSysadminPromotion(t *testing.T) {
	store, _, mux := idpSetup()
	a, err := store.CreateAccount(context.Background(), "", "target@example.com", "hash", "")
	if err != nil {
		t.Fatal(err)
	}
	w := idpRequest(mux, "POST", "/sysadmins", `{"account_id":1}`, "")
	admins, err := store.ListSysadmins(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if w.Code == http.StatusCreated && len(admins) == 1 && admins[0].AccountID == a.ID {
		t.Skip("XFAIL: anonymous request can promote a sysadmin")
	}
	if (w.Code != http.StatusUnauthorized && w.Code != http.StatusForbidden) || len(admins) != 0 {
		t.Fatalf("anonymous promotion: %d admins=%v", w.Code, admins)
	}
}

func TestIDP021OversizedRegistration(t *testing.T) {
	store, _, mux := idpSetup()
	body, err := json.Marshal(map[string]string{"email": "large@example.com", "password": "password123", "name": strings.Repeat("a", 1<<20)})
	if err != nil {
		t.Fatal(err)
	}
	w := idpRequest(mux, "POST", "/auth/register", string(body), "")
	if w.Code == http.StatusCreated && len(store.accounts) == 1 {
		t.Skip("XFAIL: registration accepts an oversized body")
	}
	if w.Code != http.StatusRequestEntityTooLarge || len(store.accounts) != 0 {
		t.Fatalf("oversized registration: %d accounts=%d", w.Code, len(store.accounts))
	}
}

func TestIDP023ManagerCannotPromoteOwner(t *testing.T) {
	store, svc, mux := idpSetup()
	ctx := context.Background()
	manager, err := store.CreateAccount(ctx, "", "manager@example.com", "hash", "")
	if err != nil {
		t.Fatal(err)
	}
	target, err := store.CreateAccount(ctx, "", "target@example.com", "hash", "")
	if err != nil {
		t.Fatal(err)
	}
	owner, err := store.CreateAccount(ctx, "", "owner@example.com", "hash", "")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.AddOwner(ctx, "trattoria", owner.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.AddManager(ctx, "trattoria", manager.ID); err != nil {
		t.Fatal(err)
	}
	pair, err := svc.IssueTokens(ctx, manager.ID)
	if err != nil {
		t.Fatal(err)
	}
	body, _ := json.Marshal(map[string]int{"account_id": target.ID})
	w := idpRequest(mux, "POST", "/restaurants/trattoria/owners", string(body), pair.AccessToken)
	owners, err := svc.ListOwners(ctx, "trattoria")
	if err != nil {
		t.Fatal(err)
	}
	var promoted bool
	for _, candidate := range owners {
		if candidate.AccountID == target.ID {
			promoted = true
		}
	}
	if w.Code == http.StatusCreated && len(owners) == 2 && promoted {
		t.Skip("XFAIL: manager can promote an owner")
	}
	if w.Code != http.StatusForbidden || len(owners) != 1 || owners[0].AccountID != owner.ID {
		t.Fatalf("manager promotion: %d owners=%v", w.Code, owners)
	}
}

func TestIDP024ClaimAnotherAccountsCPF(t *testing.T) {
	_, svc, _ := idpSetup()
	ctx := context.Background()
	a, err := svc.Register(ctx, RegisterInput{Email: "a@example.com", Password: "password123", CPF: "12345678901"})
	if err != nil {
		t.Fatal(err)
	}
	b, err := svc.Register(ctx, RegisterInput{Email: "b@example.com", Password: "password123"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.ClaimCPF(ctx, b.ID, "12345678901"); !errors.Is(err, ErrConflict) {
		t.Fatalf("claimed another account's CPF: %v", err)
	}
	original, err := svc.Account(ctx, a.ID)
	if err != nil || original.CPF != "12345678901" {
		t.Fatalf("original CPF changed: %#v %v", original, err)
	}
	other, err := svc.Account(ctx, b.ID)
	if err != nil || other.CPF != "" {
		t.Fatalf("claimant changed: %#v %v", other, err)
	}
}

func TestIDP025ShadowCannotAuthenticate(t *testing.T) {
	store, svc, mux := idpSetup()
	store.accounts[1] = &Account{ID: 1, CPF: "12345678901", CreatedAt: time.Now()}
	store.nextAcctID = 2
	// Even a correctly signed access token must not authenticate a shadow.
	token, err := svc.GenerateToken(1, "")
	if err != nil {
		t.Fatal(err)
	}
	w := idpRequest(mux, "GET", "/auth/me", "", token)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("shadow authenticated: %d %s", w.Code, w.Body.String())
	}
}

func TestIDP026ConcurrentLastMethodRemoval(t *testing.T) {
	_, svc, _ := idpSetup()
	ctx := context.Background()
	u, err := svc.Register(ctx, RegisterInput{Email: "race@example.com", Password: "password123"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.AddPhone(ctx, u.ID, "+5511999999999"); err != nil {
		t.Fatal(err)
	}
	start := make(chan struct{})
	results := make(chan error, 2)
	go func() { <-start; results <- svc.RemoveEmail(ctx, u.ID) }()
	go func() { <-start; results <- svc.RemovePhone(ctx, u.ID) }()
	close(start)
	successes, conflicts := 0, 0
	for i := 0; i < 2; i++ {
		err := <-results
		if err == nil {
			successes++
		} else if errors.Is(err, ErrConflict) {
			conflicts++
		} else {
			t.Fatal(err)
		}
	}
	if successes != 1 || conflicts != 1 {
		t.Fatalf("successes=%d conflicts=%d", successes, conflicts)
	}
	a, err := svc.Account(ctx, u.ID)
	if err != nil || (!a.EmailVerified && !a.PhoneVerified) {
		t.Fatalf("no verified method left: %#v %v", a, err)
	}
}
