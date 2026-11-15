package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

func TestAuthAPI_Lifecycle(t *testing.T) {
	app := NewApp(nil) // Uses in-memory stores for isolated unit/transport testing
	mux := setupRouter(app)

	// 1. Register a new account
	regPayload := map[string]string{
		"email":    "manager@pratello.com",
		"password": "strongPassword123",
		"phone":    "+5511999998888",
	}
	body, _ := json.Marshal(regPayload)
	req := httptest.NewRequest("POST", "/api/auth/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created on register, got %d: %s", w.Code, w.Body.String())
	}

	var regResponse struct {
		ID        int    `json:"id"`
		Email     string `json:"email"`
		Phone     string `json:"phone"`
		CreatedAt string `json:"created_at"`
	}
	if err := json.NewDecoder(w.Body).Decode(&regResponse); err != nil {
		t.Fatalf("failed to decode register response: %v", err)
	}
	if regResponse.Email != "manager@pratello.com" || regResponse.ID == 0 {
		t.Fatalf("unexpected account registered: %+v", regResponse)
	}

	// 2. Duplicate registration returns 409 Conflict
	wDup := httptest.NewRecorder()
	mux.ServeHTTP(wDup, httptest.NewRequest("POST", "/api/auth/register", bytes.NewReader(body)))
	if wDup.Code != http.StatusConflict {
		t.Fatalf("expected 409 Conflict for duplicate email, got %d", wDup.Code)
	}

	// 3. Invalid registration returns 400 Bad Request
	badReg, _ := json.Marshal(map[string]string{"email": "not-an-email", "password": "123"})
	wBad := httptest.NewRecorder()
	mux.ServeHTTP(wBad, httptest.NewRequest("POST", "/api/auth/register", bytes.NewReader(badReg)))
	if wBad.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 Bad Request for invalid registration, got %d", wBad.Code)
	}

	// 4. Login with bad credentials returns 401 Unauthorized
	badLogin, _ := json.Marshal(map[string]string{"email": "manager@pratello.com", "password": "wrongpassword"})
	wBadLogin := httptest.NewRecorder()
	mux.ServeHTTP(wBadLogin, httptest.NewRequest("POST", "/api/auth/login", bytes.NewReader(badLogin)))
	if wBadLogin.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 Unauthorized for bad login, got %d", wBadLogin.Code)
	}

	// 5. Successful Login returns 200 and JWT Token
	loginPayload, _ := json.Marshal(map[string]string{
		"email":    "manager@pratello.com",
		"password": "strongPassword123",
	})
	wLogin := httptest.NewRecorder()
	mux.ServeHTTP(wLogin, httptest.NewRequest("POST", "/api/auth/login", bytes.NewReader(loginPayload)))
	if wLogin.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for login, got %d: %s", wLogin.Code, wLogin.Body.String())
	}

	var authRes struct {
		Token   string `json:"token"`
		Account struct {
			ID    int    `json:"id"`
			Email string `json:"email"`
		} `json:"account"`
	}
	if err := json.NewDecoder(wLogin.Body).Decode(&authRes); err != nil {
		t.Fatalf("failed to decode login response: %v", err)
	}
	if authRes.Token == "" || authRes.Account.ID != regResponse.ID {
		t.Fatalf("invalid login response: token=%s, id=%d", authRes.Token, authRes.Account.ID)
	}

	// 6. Access /auth/me without token -> 401
	wMeNoToken := httptest.NewRecorder()
	mux.ServeHTTP(wMeNoToken, httptest.NewRequest("GET", "/api/auth/me", nil))
	if wMeNoToken.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 Unauthorized without token, got %d", wMeNoToken.Code)
	}

	// 7. Access /auth/me with valid Bearer token -> 200
	reqMe := httptest.NewRequest("GET", "/api/auth/me", nil)
	reqMe.Header.Set("Authorization", "Bearer "+authRes.Token)
	wMe := httptest.NewRecorder()
	mux.ServeHTTP(wMe, reqMe)
	if wMe.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for /auth/me, got %d: %s", wMe.Code, wMe.Body.String())
	}

	var meRes struct {
		ID    int    `json:"id"`
		Email string `json:"email"`
	}
	if err := json.NewDecoder(wMe.Body).Decode(&meRes); err != nil {
		t.Fatalf("failed to decode /auth/me response: %v", err)
	}
	if meRes.Email != "manager@pratello.com" || meRes.ID != regResponse.ID {
		t.Fatalf("mismatched /auth/me data: %+v", meRes)
	}

	// 8. Sysadmin endpoints
	// Promote account to sysadmin
	saPayload, _ := json.Marshal(map[string]int{"account_id": regResponse.ID})
	wSaPost := httptest.NewRecorder()
	mux.ServeHTTP(wSaPost, httptest.NewRequest("POST", "/api/sysadmins", bytes.NewReader(saPayload)))
	if wSaPost.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created for /sysadmins, got %d: %s", wSaPost.Code, wSaPost.Body.String())
	}

	// List sysadmins
	wSaGet := httptest.NewRecorder()
	mux.ServeHTTP(wSaGet, httptest.NewRequest("GET", "/api/sysadmins", nil))
	if wSaGet.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for GET /sysadmins, got %d", wSaGet.Code)
	}

	var saList []struct {
		ID        int `json:"id"`
		AccountID int `json:"account_id"`
	}
	if err := json.NewDecoder(wSaGet.Body).Decode(&saList); err != nil {
		t.Fatalf("failed to decode sysadmins list: %v", err)
	}
	if len(saList) == 0 || saList[0].AccountID != regResponse.ID {
		t.Fatalf("unexpected sysadmin list: %+v", saList)
	}

	// 9. Restaurant Manager endpoints
	// Seed restaurant
	_ = app.Restaurants.Create(req.Context(), &Restaurant{
		Slug: "bella-italia",
		Name: "Bella Italia",
	})

	// Link manager
	mgrPayload, _ := json.Marshal(map[string]int{"account_id": regResponse.ID})
	wMgrPost := httptest.NewRecorder()
	mux.ServeHTTP(wMgrPost, httptest.NewRequest("POST", "/api/restaurants/bella-italia/managers", bytes.NewReader(mgrPayload)))
	if wMgrPost.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created for POST managers, got %d: %s", wMgrPost.Code, wMgrPost.Body.String())
	}

	// List managers
	wMgrGet := httptest.NewRecorder()
	mux.ServeHTTP(wMgrGet, httptest.NewRequest("GET", "/api/restaurants/bella-italia/managers", nil))
	if wMgrGet.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for GET managers, got %d", wMgrGet.Code)
	}

	var mgrList []struct {
		ID             int    `json:"id"`
		RestaurantSlug string `json:"restaurant_slug"`
		AccountID      int    `json:"account_id"`
	}
	if err := json.NewDecoder(wMgrGet.Body).Decode(&mgrList); err != nil {
		t.Fatalf("failed to decode manager list: %v", err)
	}
	if len(mgrList) != 1 || mgrList[0].AccountID != regResponse.ID {
		t.Fatalf("unexpected manager list: %+v", mgrList)
	}

	// Delete manager
	wMgrDel := httptest.NewRecorder()
	delURL := fmt.Sprintf("/api/restaurants/bella-italia/managers/%d", regResponse.ID)
	mux.ServeHTTP(wMgrDel, httptest.NewRequest("DELETE", delURL, nil))
	if wMgrDel.Code != http.StatusNoContent {
		t.Fatalf("expected 204 No Content for DELETE manager, got %d", wMgrDel.Code)
	}

	// Verify manager was removed
	wMgrGetAfter := httptest.NewRecorder()
	mux.ServeHTTP(wMgrGetAfter, httptest.NewRequest("GET", "/api/restaurants/bella-italia/managers", nil))
	var mgrListAfter []struct {
		ID int `json:"id"`
	}
	_ = json.NewDecoder(wMgrGetAfter.Body).Decode(&mgrListAfter)
	if len(mgrListAfter) != 0 {
		t.Fatalf("expected 0 managers after deletion, got %d", len(mgrListAfter))
	}
}

func TestAuthAPI_PostgresRealDB(t *testing.T) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://nick@localhost:5432/pratello?sslmode=disable"
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		t.Fatalf("failed to open postgres: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		t.Skipf("skipping Postgres integration test, db not reachable: %v", err)
	}

	// Ensure schema columns are present
	_ = migrateSchema(db)

	app := NewApp(db)
	mux := setupRouter(app)

	testEmail := fmt.Sprintf("test_real_%d@pratello.com", time.Now().UnixNano())
	testPassword := "superSecret123"
	testName := "Maria Silva"
	testPhone := "+5511988887777"

	// 1. Register with real DB
	regPayload, _ := json.Marshal(map[string]string{
		"name":     testName,
		"email":    testEmail,
		"password": testPassword,
		"phone":    testPhone,
	})
	wReg := httptest.NewRecorder()
	mux.ServeHTTP(wReg, httptest.NewRequest("POST", "/api/auth/register", bytes.NewReader(regPayload)))
	if wReg.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created on real DB register, got %d: %s", wReg.Code, wReg.Body.String())
	}

	var regAcct struct {
		ID        int    `json:"id"`
		Name      string `json:"name"`
		Email     string `json:"email"`
		Phone     string `json:"phone"`
		CreatedAt string `json:"created_at"`
	}
	if err := json.NewDecoder(wReg.Body).Decode(&regAcct); err != nil {
		t.Fatalf("failed to decode register response: %v", err)
	}
	if regAcct.ID == 0 || regAcct.Email != testEmail {
		t.Fatalf("unexpected account returned: %+v", regAcct)
	}
	defer func() {
		_, _ = db.Exec(`DELETE FROM "account" WHERE id = $1`, regAcct.ID)
	}()

	// 2. Duplicate registration on real DB -> 409 Conflict with descriptive message
	wDup := httptest.NewRecorder()
	mux.ServeHTTP(wDup, httptest.NewRequest("POST", "/api/auth/register", bytes.NewReader(regPayload)))
	if wDup.Code != http.StatusConflict {
		t.Fatalf("expected 409 Conflict on duplicate email, got %d: %s", wDup.Code, wDup.Body.String())
	}

	// 3. Login with wrong password -> 401 Unauthorized with descriptive message
	badLoginPayload, _ := json.Marshal(map[string]string{
		"email":    testEmail,
		"password": "wrongPassword123",
	})
	wBadLogin := httptest.NewRecorder()
	mux.ServeHTTP(wBadLogin, httptest.NewRequest("POST", "/api/auth/login", bytes.NewReader(badLoginPayload)))
	if wBadLogin.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 Unauthorized for wrong password, got %d: %s", wBadLogin.Code, wBadLogin.Body.String())
	}

	// 4. Successful login on real DB -> 200 OK and JWT
	loginPayload, _ := json.Marshal(map[string]string{
		"email":    testEmail,
		"password": testPassword,
	})
	wLogin := httptest.NewRecorder()
	mux.ServeHTTP(wLogin, httptest.NewRequest("POST", "/api/auth/login", bytes.NewReader(loginPayload)))
	if wLogin.Code != http.StatusOK {
		t.Fatalf("expected 200 OK on real DB login, got %d: %s", wLogin.Code, wLogin.Body.String())
	}

	var loginRes struct {
		Token   string `json:"token"`
		Account struct {
			ID    int    `json:"id"`
			Name  string `json:"name"`
			Email string `json:"email"`
		} `json:"account"`
	}
	if err := json.NewDecoder(wLogin.Body).Decode(&loginRes); err != nil {
		t.Fatalf("failed to decode login response: %v", err)
	}
	if loginRes.Token == "" || loginRes.Account.ID != regAcct.ID {
		t.Fatalf("unexpected login payload: %+v", loginRes)
	}

	// 5. Query /api/auth/me with JWT on real DB -> 200 OK
	wMe := httptest.NewRecorder()
	reqMe := httptest.NewRequest("GET", "/api/auth/me", nil)
	reqMe.Header.Set("Authorization", "Bearer "+loginRes.Token)
	mux.ServeHTTP(wMe, reqMe)
	if wMe.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for /auth/me on real DB, got %d: %s", wMe.Code, wMe.Body.String())
	}
}
