package auth

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"net/http/httptest"
	"os"
	db "pratello/server/internal/db"
	"strings"
	"sync"
	"testing"
	"time"
)

func idpSetup() (*memoryStore, *Service, *http.ServeMux) {
	store := NewMemoryStore()
	svc := NewService(store)
	svc.SetHashCost(bcrypt.MinCost)
	svc.verification = newTestVerification()
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
	requestTestEmail(t, mux, "user@example.com")
	w := idpRequest(mux, "POST", "/auth/register", `{"email":"user@example.com","password":"password123","code":"123456"}`, "")
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

func TestIDP028RedisVerification(t *testing.T) {
	url := os.Getenv("AUTH_TEST_REDIS_URL")
	if url == "" {
		t.Skip("set AUTH_TEST_REDIS_URL for Redis integration")
	}
	store, svc, mux := idpSetup()
	closeRedis, err := svc.EnableFixedCodeVerification(url)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = closeRedis() })
	ctx := t.Context()
	redis := svc.verification.(redisVerification).client
	email := fmt.Sprintf("verification-%d@example.test", time.Now().UnixNano())
	key := challengeKey("register-email", 0, 0, email)
	t.Cleanup(func() { _ = redis.Del(context.Background(), key).Err() })
	submit := func(code string, want int) {
		t.Helper()
		body, _ := json.Marshal(map[string]string{"email": email, "password": "password123", "code": code})
		w := idpRequest(mux, "POST", "/auth/register", string(body), "")
		if w.Code != want || (want == 400 && !strings.Contains(w.Body.String(), ErrWrongCode.Error())) {
			t.Fatalf("registration: %d %s", w.Code, w.Body)
		}
	}
	submit(fixedCode, 400)
	requestTestEmail(t, mux, email)
	if v, err := redis.Get(ctx, key).Result(); err != nil || v != fixedCode {
		t.Fatalf("Redis challenge: %v %v", v, err)
	}
	if ttl, err := redis.PTTL(ctx, key).Result(); err != nil || ttl <= 0 || ttl > verificationTTL {
		t.Fatalf("Redis TTL: %v %v", ttl, err)
	}
	submit("000000", 400)
	if err := svc.verifyCode(ctx, "change-email", 1, 0, email, fixedCode); !errors.Is(err, ErrWrongCode) {
		t.Fatalf("cross-purpose code accepted: %v", err)
	}
	if err := redis.PExpireAt(ctx, key, time.Now().Add(-time.Second)).Err(); err != nil {
		t.Fatal(err)
	}
	submit(fixedCode, 400)
	if len(store.accounts) != 0 {
		t.Fatal("invalid code created an account")
	}
	requestTestEmail(t, mux, email)
	submit(fixedCode, 201)
	if n, err := redis.Exists(ctx, key).Result(); err != nil || n != 0 {
		t.Fatalf("successful code not consumed: %d %v", n, err)
	}
	submit(fixedCode, 400)
}

func TestIDP029ContactChangesRequireAccountBoundCode(t *testing.T) {
	_, svc, mux := idpSetup()
	ctx := t.Context()
	makeAccount := func(email string) (int, string) {
		t.Helper()
		u, err := svc.Register(ctx, RegisterInput{Email: email, Password: "password123"})
		if err != nil {
			t.Fatal(err)
		}
		p, err := svc.IssueTokens(ctx, u.ID)
		if err != nil {
			t.Fatal(err)
		}
		return u.ID, p.AccessToken
	}
	id, token := makeAccount("one@example.test")
	_, other := makeAccount("two@example.test")
	call := func(method, path, body, bearer string, want int) {
		t.Helper()
		w := idpRequest(mux, method, path, body, bearer)
		if w.Code != want {
			t.Fatalf("%s: %d %s", path, w.Code, w.Body)
		}
	}
	call("PUT", "/me/email", `{"email":"new@example.test"}`, token, 202)
	call("POST", "/me/phone/request-code", `{"phone":"+5511999999999"}`, token, 202)
	call("POST", "/me/email/verify", `{"email":"new@example.test","password":"newpassword","current_password":"password123","code":"654321"}`, token, 400)
	call("POST", "/me/phone/verify", `{"phone":"+5511999999999","code":"123456"}`, other, 400)
	a, err := svc.Account(ctx, id)
	if err != nil || a.Email != "one@example.test" || a.Phone != "" {
		t.Fatalf("contact changed before verification: %+v %v", a, err)
	}
	call("POST", "/me/email/verify", `{"email":"new@example.test","password":"newpassword","current_password":"password123","code":"123456"}`, token, 200)
	call("POST", "/me/phone/verify", `{"phone":"+5511999999999","code":"123456"}`, token, 200)
	a, err = svc.Account(ctx, id)
	if err != nil || a.Email != "new@example.test" || a.Phone != "+5511999999999" || !a.PhoneVerified {
		t.Fatalf("contacts not updated: %+v %v", a, err)
	}
}

func TestIDP002DuplicateEmailCaseInsensitive(t *testing.T) {
	store, _, mux := idpSetup()
	for i, email := range []string{"dup@pratello.com", "DUP@PRATELLO.COM"} {
		requestTestEmail(t, mux, email)
		body, _ := json.Marshal(map[string]string{"email": email, "password": "validPassword123", "code": fixedCode})
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

func TestIDP021OversizedRegistration(t *testing.T) {
	store, _, mux := idpSetup()
	body, err := json.Marshal(map[string]string{"email": "large@example.com", "password": "password123", "name": strings.Repeat("a", 1<<20)})
	if err != nil {
		t.Fatal(err)
	}
	w := idpRequest(mux, "POST", "/auth/register", string(body), "")
	if w.Code != http.StatusRequestEntityTooLarge || len(store.accounts) != 0 {
		t.Fatalf("oversized registration: %d accounts=%d", w.Code, len(store.accounts))
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

// This test uses an isolated, temporary schema in the explicitly supplied DB.
func TestPostgresCPFClaimPreservesWallet(t *testing.T) {
	url := os.Getenv("AUTH_TEST_DATABASE_URL")
	if url == "" {
		t.Skip("AUTH_TEST_DATABASE_URL is not set")
	}
	database, err := sql.Open("postgres", url)
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	database.SetMaxOpenConns(1)
	name := fmt.Sprintf("auth_test_%d", time.Now().UnixNano())
	quoted := pq.QuoteIdentifier(name)
	if _, err = database.Exec(`CREATE SCHEMA ` + quoted); err != nil {
		t.Fatal(err)
	}
	defer database.Exec(`DROP SCHEMA ` + quoted + ` CASCADE`)
	if _, err = database.Exec(`SET search_path TO ` + quoted); err != nil {
		t.Fatal(err)
	}
	if err = db.Migrate(database); err != nil {
		t.Fatalf("initialize schema: %v", err)
	}
	if err = db.Migrate(database); err != nil {
		t.Fatalf("repeat schema initialization: %v", err)
	}
	ctx := context.Background()
	store := NewSQLStore(database)
	var shadowID, restaurantID, walletID int
	if err = database.QueryRow(`INSERT INTO account (cpf) VALUES ('12345678901') RETURNING id`).Scan(&shadowID); err != nil {
		t.Fatal(err)
	}
	if err = database.QueryRow(`INSERT INTO restaurant (name_slug,name) VALUES ('test','Test') RETURNING id`).Scan(&restaurantID); err != nil {
		t.Fatal(err)
	}
	if err = database.QueryRow(`INSERT INTO loyalty_wallet (account_id,restaurant_id,coins_amount) VALUES ($1,$2,50) RETURNING id`, shadowID, restaurantID).Scan(&walletID); err != nil {
		t.Fatal(err)
	}
	a, err := store.CreateAccount(ctx, "", "person@example.com", "hash", "")
	if err != nil {
		t.Fatal(err)
	}
	if err = store.ClaimCPF(ctx, a.ID, "12345678901"); err != nil {
		t.Fatal(err)
	}
	var owner, balance int
	if err = database.QueryRow(`SELECT account_id, coins_amount FROM loyalty_wallet WHERE id=$1`, walletID).Scan(&owner, &balance); err != nil {
		t.Fatal(err)
	}
	if owner != a.ID || balance != 50 {
		t.Fatalf("wallet owner=%d balance=%d", owner, balance)
	}
	if err = store.RemoveCPF(ctx, a.ID); err != nil {
		t.Fatal(err)
	}
	if err = database.QueryRow(`SELECT account_id FROM loyalty_wallet WHERE id=$1`, walletID).Scan(&owner); err != nil || owner != a.ID {
		t.Fatalf("wallet lost after unlink: %d %v", owner, err)
	}
	var otherWallet, ledgerID int
	if err = database.QueryRow(`INSERT INTO account(cpf) VALUES ('33333333333') RETURNING id`).Scan(&shadowID); err != nil {
		t.Fatal(err)
	}
	if err = database.QueryRow(`INSERT INTO loyalty_wallet(account_id,restaurant_id,coins_amount) VALUES($1,$2,30) RETURNING id`, shadowID, restaurantID).Scan(&otherWallet); err != nil {
		t.Fatal(err)
	}
	if err = database.QueryRow(`INSERT INTO wallet_transaction(loyalty_wallet_id,points_delta,balance_after,reason) VALUES($1,30,30,'earned') RETURNING id`, otherWallet).Scan(&ledgerID); err != nil {
		t.Fatal(err)
	}
	if err = store.ClaimCPF(ctx, a.ID, "33333333333"); err != nil {
		t.Fatal(err)
	}
	var delta int
	if err = database.QueryRow(`SELECT l.points_delta,w.coins_amount FROM wallet_transaction l JOIN loyalty_wallet w ON w.id=l.loyalty_wallet_id WHERE l.id=$1 AND w.account_id=$2`, ledgerID, a.ID).Scan(&delta, &balance); err != nil || delta != 30 || balance != 80 {
		t.Fatalf("wallet consolidation: delta=%d balance=%d err=%v", delta, balance, err)
	}
	var graduateID int
	if err = database.QueryRow(`INSERT INTO account (cpf) VALUES ('22222222222') RETURNING id`).Scan(&graduateID); err != nil {
		t.Fatal(err)
	}
	graduated, err := store.RegisterWithCPF(ctx, "", "graduate@example.com", "hash", "", "22222222222", false)
	if err != nil || graduated.ID != graduateID || !graduated.EmailVerified || graduated.CPFClaimedAt == nil {
		t.Fatalf("graduation: %#v %v", graduated, err)
	}
	svc := NewService(store)
	pair, err := svc.IssueTokens(ctx, graduated.ID)
	if err != nil {
		t.Fatal(err)
	}
	rotated, err := svc.Refresh(ctx, pair.RefreshToken)
	if err != nil {
		t.Fatal(err)
	}
	if err = svc.LogoutAll(ctx, graduated.ID); err != nil {
		t.Fatal(err)
	}
	if _, err = svc.Authenticate(ctx, rotated.AccessToken); err == nil {
		t.Fatal("logout-all left an access token valid")
	}
	if _, err = svc.Refresh(ctx, rotated.RefreshToken); err == nil {
		t.Fatal("logout-all left a refresh token valid")
	}
	if err = svc.RemoveEmail(ctx, graduated.ID); err != ErrConflict {
		t.Fatalf("last email removal: %v", err)
	}
	if _, err = svc.AddPhone(ctx, graduated.ID, "+5511999999999"); err != nil {
		t.Fatal(err)
	}
	if err = svc.RemoveEmail(ctx, graduated.ID); err != nil {
		t.Fatal(err)
	}
	if err = svc.RemovePhone(ctx, graduated.ID); err != ErrConflict {
		t.Fatalf("last phone removal: %v", err)
	}
}

func TestIDP022UnavailableVerificationDoesNotCreateAccount(t *testing.T) {
	store, svc, mux := idpSetup()
	svc.verification = nil
	w := idpRequest(mux, "POST", "/auth/register/email/request-code", `{"email":"alice@example.com"}`, "")
	if w.Code != 503 {
		t.Fatalf("request: %d %s", w.Code, w.Body)
	}
	w = idpRequest(mux, "POST", "/auth/register", `{"email":"alice@example.com","password":"password123","code":"123456"}`, "")
	if w.Code != 503 {
		t.Fatalf("registration: %d %s", w.Code, w.Body)
	}
	if len(store.accounts) != 0 {
		t.Fatal("account created after notification failure")
	}
}

func TestService_Register_And_Login_Success(t *testing.T) {
	mock := NewMemoryStore()
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

func TestService_LoginRejectsUnverifiedEmail(t *testing.T) {
	store := NewMemoryStore()
	svc := NewService(store)
	svc.SetHashCost(bcrypt.MinCost)
	_, err := svc.Register(context.Background(), RegisterInput{Email: "unverified@example.com", Password: "password123"})
	if err != nil {
		t.Fatal(err)
	}
	store.byEmail["unverified@example.com"].EmailVerified = false
	if _, _, err := svc.Login(context.Background(), "unverified@example.com", "password123"); !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("unverified login: %v", err)
	}
}

func TestShadowGraduationAndClaim(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()
	svc := NewService(store)
	svc.SetHashCost(bcrypt.MinCost)
	store.accounts[1] = &Account{ID: 1, CPF: "12345678901", CreatedAt: time.Now()}
	store.nextAcctID = 2
	u, err := svc.Register(ctx, RegisterInput{Email: "a@example.com", Password: "password123", CPF: "12345678901"})
	if err != nil {
		t.Fatal(err)
	}
	if u.ID != 1 {
		t.Fatalf("shadow not graduated: %d", u.ID)
	}
	a, err := store.GetAccountByID(ctx, 1)
	if err != nil || a.CPFClaimedAt == nil || !a.EmailVerified {
		t.Fatalf("graduation: %#v %v", a, err)
	}
	if _, err := svc.Register(ctx, RegisterInput{Email: "b@example.com", Password: "password123", CPF: "12345678901"}); !errors.Is(err, ErrConflict) {
		t.Fatalf("real CPF reused: %v", err)
	}
	store.accounts[2] = &Account{ID: 2, CPF: "22222222222", CreatedAt: time.Now()}
	store.nextAcctID = 3
	b, err := svc.Register(ctx, RegisterInput{Email: "b@example.com", Password: "password123"})
	if err != nil {
		t.Fatal(err)
	}
	claimed, err := svc.ClaimCPF(ctx, b.ID, "22222222222")
	if err != nil || claimed.CPF != "22222222222" {
		t.Fatalf("claim: %#v %v", claimed, err)
	}
	if _, err := store.GetAccountByID(ctx, 2); !errors.Is(err, ErrNotFound) {
		t.Fatalf("shadow remains: %v", err)
	}
	if _, err := svc.ClaimCPF(ctx, b.ID, "12345678901"); !errors.Is(err, ErrConflict) {
		t.Fatalf("second claim: %v", err)
	}
	if err := svc.RemoveCPF(ctx, b.ID); err != nil {
		t.Fatal(err)
	}
	if a, _ := store.GetAccountByID(ctx, b.ID); a.CPF != "" {
		t.Fatal("CPF still attached")
	}
}

func TestConcurrentCPFRegistration(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()
	var wg sync.WaitGroup
	results := make(chan error, 5)
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			_, err := store.RegisterWithCPF(ctx, "", "", "", fmt.Sprintf("+551199999%04d", i), "12345678901", true)
			results <- err
		}(i)
	}
	wg.Wait()
	close(results)
	successes := 0
	for err := range results {
		if err == nil {
			successes++
		} else if !errors.Is(err, ErrConflict) {
			t.Fatal(err)
		}
	}
	if successes != 1 {
		t.Fatalf("registrations succeeded: %d", successes)
	}
}

func TestEmailChangeAndLastMethod(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()
	svc := NewService(store)
	svc.SetHashCost(bcrypt.MinCost)
	u, err := svc.Register(ctx, RegisterInput{Email: "a@example.com", Password: "password123"})
	if err != nil {
		t.Fatal(err)
	}
	if err = svc.RemoveEmail(ctx, u.ID); !errors.Is(err, ErrConflict) {
		t.Fatalf("removed last method: %v", err)
	}
	if _, err = svc.SetEmail(ctx, u.ID, "b@example.com", "newpassword", "wrong"); !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("wrong current password: %v", err)
	}
	if _, err = svc.SetEmail(ctx, u.ID, "b@example.com", "newpassword", "password123"); err != nil {
		t.Fatal(err)
	}
	if _, _, err = svc.Login(ctx, "b@example.com", "newpassword"); err != nil {
		t.Fatal(err)
	}
	if _, err = svc.AddPhone(ctx, u.ID, "+5511999999999"); err != nil {
		t.Fatal(err)
	}
	if err = svc.RemoveEmail(ctx, u.ID); err != nil {
		t.Fatal(err)
	}
	if _, _, err = svc.Login(ctx, "b@example.com", "newpassword"); !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("removed email login: %v", err)
	}
	if err = svc.RemovePhone(ctx, u.ID); !errors.Is(err, ErrConflict) {
		t.Fatalf("removed last phone: %v", err)
	}
}

func TestPhoneRegistrationHTTPRequiresPIN(t *testing.T) {
	_, _, mux := idpSetup()
	const phone = `{"phone":"+5511888888888"}`
	const correct = `{"phone":"+5511888888888","code":"123456"}`
	w := idpRequest(mux, "POST", "/auth/register/phone/request-code", phone, "")
	if w.Code != 202 {
		t.Fatal(w.Code, w.Body)
	}
	w = idpRequest(mux, "POST", "/auth/register/phone/verify", phone, "")
	if w.Code != 400 {
		t.Fatal(w.Code, w.Body)
	}
	w = idpRequest(mux, "POST", "/auth/register/phone/verify", correct, "")
	if w.Code != http.StatusCreated {
		t.Fatalf("register phone: %d %s", w.Code, w.Body.String())
	}
	var pair TokenPair
	if err := json.Unmarshal(w.Body.Bytes(), &pair); err != nil || pair.RefreshToken == "" {
		t.Fatalf("invalid response: %s", w.Body.String())
	}
	w = idpRequest(mux, "POST", "/auth/login/phone/request-code", phone, "")
	if w.Code != 202 {
		t.Fatal(w.Code, w.Body)
	}
	w = idpRequest(mux, "POST", "/auth/login/phone/verify", `{"phone":"+5511888888888","code":"654321"}`, "")
	if w.Code != 400 || !strings.Contains(w.Body.String(), ErrWrongCode.Error()) {
		t.Fatal(w.Code, w.Body)
	}
	w = idpRequest(mux, "POST", "/auth/login/phone/verify", correct, "")
	if w.Code != http.StatusOK {
		t.Fatalf("phone login: %d %s", w.Code, w.Body.String())
	}
	if w = idpRequest(mux, "POST", "/auth/login/phone/verify", correct, ""); w.Code != 400 {
		t.Fatal(w.Code, w.Body)
	}
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, httptest.NewRequest("POST", "/auth/login/phone/request-code", bytes.NewBufferString(`{"phone":"+5511777777777"}`)))
	if w.Code != http.StatusAccepted {
		t.Fatalf("unknown phone request: %d", w.Code)
	}
}

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
	updated, err := svc.AddPhone(ctx, a.ID, a.Phone)
	if err != nil {
		t.Fatal(err)
	}
	if !updated.PhoneVerified {
		t.Fatalf("invalid phone state: %#v", updated)
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
	svc.verification = newTestVerification()
	mux := http.NewServeMux()
	NewHandler(svc).Routes(mux)
	w := httptest.NewRecorder()
	requestTestEmail(t, mux, "user@example.com")
	mux.ServeHTTP(w, httptest.NewRequest("POST", "/auth/register", bytes.NewBufferString(`{"email":"user@example.com","password":"password123","code":"123456"}`)))
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
