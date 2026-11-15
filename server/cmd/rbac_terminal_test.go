package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRBAC_And_TerminalLifecycle(t *testing.T) {
	app := NewApp(nil)
	mux := setupRouter(app)

	// 1. Create two user accounts: Alice and Bob
	alicePayload, _ := json.Marshal(map[string]string{
		"email":    "alice@pratello.com",
		"password": "alicePassword123",
	})
	wAlice := httptest.NewRecorder()
	mux.ServeHTTP(wAlice, httptest.NewRequest("POST", "/api/auth/register", bytes.NewReader(alicePayload)))
	var aliceAcct Account
	_ = json.NewDecoder(wAlice.Body).Decode(&aliceAcct)

	bobPayload, _ := json.Marshal(map[string]string{
		"email":    "bob@pratello.com",
		"password": "bobPassword123",
	})
	wBob := httptest.NewRecorder()
	mux.ServeHTTP(wBob, httptest.NewRequest("POST", "/api/auth/register", bytes.NewReader(bobPayload)))
	var bobAcct Account
	_ = json.NewDecoder(wBob.Body).Decode(&bobAcct)

	// Log in as Alice
	wLoginAlice := httptest.NewRecorder()
	mux.ServeHTTP(wLoginAlice, httptest.NewRequest("POST", "/api/auth/login", bytes.NewReader(alicePayload)))
	var aliceAuth struct {
		Token string `json:"token"`
	}
	_ = json.NewDecoder(wLoginAlice.Body).Decode(&aliceAuth)

	// Log in as Bob
	wLoginBob := httptest.NewRecorder()
	mux.ServeHTTP(wLoginBob, httptest.NewRequest("POST", "/api/auth/login", bytes.NewReader(bobPayload)))
	var bobAuth struct {
		Token string `json:"token"`
	}
	_ = json.NewDecoder(wLoginBob.Body).Decode(&bobAuth)

	// 2. Register a logical restaurant and make Alice an OWNER
	slug := "trattoria-da-alice"
	_ = app.Restaurants.Create(httptest.NewRequest("GET", "/", nil).Context(), &Restaurant{
		Slug: slug,
		Name: "Trattoria da Alice",
	})
	_, _ = app.Accounts.AddOwner(httptest.NewRequest("GET", "/", nil).Context(), slug, aliceAcct.ID)

	// 3. Test Owner Authority: Alice promotes Bob to Owner
	promotePayload, _ := json.Marshal(map[string]int{"account_id": bobAcct.ID})
	reqPromote := httptest.NewRequest("POST", fmt.Sprintf("/api/restaurants/%s/owners", slug), bytes.NewReader(promotePayload))
	reqPromote.Header.Set("Authorization", "Bearer "+aliceAuth.Token)
	wPromote := httptest.NewRecorder()
	mux.ServeHTTP(wPromote, reqPromote)
	if wPromote.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created for promote owner, got %d: %s", wPromote.Code, wPromote.Body.String())
	}

	// Verify both Alice and Bob are owners
	wListOwners := httptest.NewRecorder()
	mux.ServeHTTP(wListOwners, httptest.NewRequest("GET", fmt.Sprintf("/api/restaurants/%s/owners", slug), nil))
	var owners []Owner
	_ = json.NewDecoder(wListOwners.Body).Decode(&owners)
	if len(owners) != 2 {
		t.Fatalf("expected 2 owners, got %d", len(owners))
	}

	// 4. Test Entity-Level Authority: Alice removes Bob from Owner (allowed because 2 owners exist)
	reqRemoveBob := httptest.NewRequest("DELETE", fmt.Sprintf("/api/restaurants/%s/owners/%d", slug, bobAcct.ID), nil)
	reqRemoveBob.Header.Set("Authorization", "Bearer "+aliceAuth.Token)
	wRemoveBob := httptest.NewRecorder()
	mux.ServeHTTP(wRemoveBob, reqRemoveBob)
	if wRemoveBob.Code != http.StatusNoContent {
		t.Fatalf("expected 204 No Content for remove owner, got %d", wRemoveBob.Code)
	}

	// 5. Test Orphan Prevention Rule: Alice tries to remove HERSELF as the last remaining owner
	reqRemoveAlice := httptest.NewRequest("DELETE", fmt.Sprintf("/api/restaurants/%s/owners/%d", slug, aliceAcct.ID), nil)
	reqRemoveAlice.Header.Set("Authorization", "Bearer "+aliceAuth.Token)
	wRemoveAlice := httptest.NewRecorder()
	mux.ServeHTTP(wRemoveAlice, reqRemoveAlice)
	if wRemoveAlice.Code != http.StatusConflict {
		t.Fatalf("expected 409 Conflict when attempting to remove the last owner (orphan prevention), got %d: %s", wRemoveAlice.Code, wRemoveAlice.Body.String())
	}

	// 6. Test Manager Role: Make Bob a Manager
	_, _ = app.Accounts.AddManager(httptest.NewRequest("GET", "/", nil).Context(), slug, bobAcct.ID)

	// Bob cannot promote an owner (requires Owner authority)
	reqBobPromote := httptest.NewRequest("POST", fmt.Sprintf("/api/restaurants/%s/owners", slug), bytes.NewReader(promotePayload))
	reqBobPromote.Header.Set("Authorization", "Bearer "+bobAuth.Token)
	wBobPromote := httptest.NewRecorder()
	mux.ServeHTTP(wBobPromote, reqBobPromote)
	if wBobPromote.Code != http.StatusForbidden {
		t.Fatalf("expected 403 Forbidden for manager attempting owner promotion, got %d", wBobPromote.Code)
	}

	// 7. Manager creates a Physical Restaurant with CNPJ verification
	// First test CNPJ mismatch
	badBranchPayload, _ := json.Marshal(map[string]string{
		"name":       "Filial Jardins",
		"cnpj":       "99.888.777/0001-11",
		"owner_cnpj": "00.000.000/0000-00", // Mismatched CNPJ
	})
	reqBadBranch := httptest.NewRequest("POST", fmt.Sprintf("/api/restaurants/%s/branches", slug), bytes.NewReader(badBranchPayload))
	reqBadBranch.Header.Set("Authorization", "Bearer "+bobAuth.Token)
	wBadBranch := httptest.NewRecorder()
	mux.ServeHTTP(wBadBranch, reqBadBranch)
	if wBadBranch.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 Bad Request for mismatched owner CNPJ, got %d: %s", wBadBranch.Code, wBadBranch.Body.String())
	}

	// Valid branch creation
	validBranchPayload, _ := json.Marshal(map[string]string{
		"name":       "Filial Jardins",
		"cnpj":       "99.888.777/0001-11",
		"owner_cnpj": "12.345.678/0001-99", // Matches registered CNPJ
	})
	reqBranch := httptest.NewRequest("POST", fmt.Sprintf("/api/restaurants/%s/branches", slug), bytes.NewReader(validBranchPayload))
	reqBranch.Header.Set("Authorization", "Bearer "+bobAuth.Token)
	wBranch := httptest.NewRecorder()
	mux.ServeHTTP(wBranch, reqBranch)
	if wBranch.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created for branch creation, got %d: %s", wBranch.Code, wBranch.Body.String())
	}
	var branch PhysicalRestaurant
	_ = json.NewDecoder(wBranch.Body).Decode(&branch)

	// 8. Employees (HR records attached to physical restaurant, separate from user accounts)
	empPayload, _ := json.Marshal(map[string]string{
		"name":     "Carlos Silva",
		"cpf":      "123.456.789-00",
		"position": "Garçom",
	})
	reqEmp := httptest.NewRequest("POST", fmt.Sprintf("/api/restaurants/%s/branches/%d/employees", slug, branch.ID), bytes.NewReader(empPayload))
	reqEmp.Header.Set("Authorization", "Bearer "+bobAuth.Token)
	wEmp := httptest.NewRecorder()
	mux.ServeHTTP(wEmp, reqEmp)
	if wEmp.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created for employee record, got %d: %s", wEmp.Code, wEmp.Body.String())
	}
	var emp Employee
	_ = json.NewDecoder(wEmp.Body).Decode(&emp)
	if emp.AccountID != nil {
		t.Fatalf("expected employee to have nil account_id initially (no auto-generated account), got %v", emp.AccountID)
	}

	// Invite flow: Link employee record to Bob's account
	invitePayload, _ := json.Marshal(map[string]int{"account_id": bobAcct.ID})
	reqInvite := httptest.NewRequest("POST", fmt.Sprintf("/api/restaurants/%s/branches/%d/employees/%d/invite", slug, branch.ID, emp.ID), bytes.NewReader(invitePayload))
	reqInvite.Header.Set("Authorization", "Bearer "+aliceAuth.Token)
	wInvite := httptest.NewRecorder()
	mux.ServeHTTP(wInvite, reqInvite)
	if wInvite.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for employee invite, got %d: %s", wInvite.Code, wInvite.Body.String())
	}
	var linkedEmp Employee
	_ = json.NewDecoder(wInvite.Body).Decode(&linkedEmp)
	if linkedEmp.AccountID == nil || *linkedEmp.AccountID != bobAcct.ID {
		t.Fatalf("expected linked account ID %d, got %v", bobAcct.ID, linkedEmp.AccountID)
	}

	// 9. Shared 6-digit Terminal PIN & Shift Session
	// Test wrong PIN
	wrongPinPayload, _ := json.Marshal(map[string]string{
		"pin":       "999999",
		"device_id": "pos-tablet-01",
	})
	wWrongPin := httptest.NewRecorder()
	mux.ServeHTTP(wWrongPin, httptest.NewRequest("POST", fmt.Sprintf("/api/restaurants/%s/branches/%d/terminal/session", slug, branch.ID), bytes.NewReader(wrongPinPayload)))
	if wWrongPin.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 Unauthorized for wrong PIN, got %d", wWrongPin.Code)
	}

	// Fail 4 more times to trigger the 5-attempt lockout
	for i := 0; i < 4; i++ {
		wLockCheck := httptest.NewRecorder()
		mux.ServeHTTP(wLockCheck, httptest.NewRequest("POST", fmt.Sprintf("/api/restaurants/%s/branches/%d/terminal/session", slug, branch.ID), bytes.NewReader(wrongPinPayload)))
	}

	// 6th attempt must be locked out (429 Too Many Requests)
	wLocked := httptest.NewRecorder()
	mux.ServeHTTP(wLocked, httptest.NewRequest("POST", fmt.Sprintf("/api/restaurants/%s/branches/%d/terminal/session", slug, branch.ID), bytes.NewReader(wrongPinPayload)))
	if wLocked.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429 Too Many Requests on locked terminal, got %d: %s", wLocked.Code, wLocked.Body.String())
	}

	// Unlock by resetting PIN
	_ = app.Accounts.RotateTerminalPIN(httptest.NewRequest("GET", "/", nil).Context(), branch.ID, "654321")

	// Open valid shift session with new 6-digit PIN
	validPinPayload, _ := json.Marshal(map[string]string{
		"pin":       "654321",
		"device_id": "pos-tablet-01",
	})
	wSession := httptest.NewRecorder()
	mux.ServeHTTP(wSession, httptest.NewRequest("POST", fmt.Sprintf("/api/restaurants/%s/branches/%d/terminal/session", slug, branch.ID), bytes.NewReader(validPinPayload)))
	if wSession.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for valid terminal PIN session, got %d: %s", wSession.Code, wSession.Body.String())
	}

	var session TerminalSession
	_ = json.NewDecoder(wSession.Body).Decode(&session)
	if session.Token == "" || session.DeviceID != "pos-tablet-01" {
		t.Fatalf("unexpected terminal session: %+v", session)
	}

	// 10. Access low-privilege operational order viewer
	reqTermOrders := httptest.NewRequest("GET", fmt.Sprintf("/api/restaurants/%s/branches/%d/terminal/orders", slug, branch.ID), nil)
	reqTermOrders.Header.Set("X-Terminal-Session", session.Token)
	wTermOrders := httptest.NewRecorder()
	mux.ServeHTTP(wTermOrders, reqTermOrders)
	if wTermOrders.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for terminal orders view, got %d: %s", wTermOrders.Code, wTermOrders.Body.String())
	}
}
