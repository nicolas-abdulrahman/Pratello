Here is the complete Markdown file. It outlines your architecture and provides a strict, pragmatic testing strategy focused on catching real bugs, handling edge cases, and stressing your system's boundaries.

Save this as `ARCHITECTURE.md` or `CONTRIBUTING.md`.

***

# Backend Architecture & Testing Guidelines

## 1. Architecture Overview
This application follows a strict **Layered Architecture** with a **Domain-Driven** file structure. Dependencies flow in exactly one direction: `Interface -> Business -> Persistence`. 

Do not bleed responsibilities across these layers.

### Layer 1: Interface (`http.go`)
*   **Responsibility:** Handles all HTTP traffic.
*   **Role:** Parses incoming HTTP requests (JSON, Query Params), translates them into domain models, calls the Business layer, and translates the results (success or error) back into HTTP status codes and JSON.
*   **Rule:** This layer contains **zero** business logic and **zero** database queries.

### Layer 2: Business (`service.go`)
*   **Responsibility:** Core application logic and domain rules.
*   **Role:** Orchestrates calls to the Persistence layer. Enforces invariants (e.g., "cannot delete last auth method", "cannot redeem more coins than you have").
*   **Rule:** Completely agnostic to HTTP (no `net/http` imports). It only knows domain models and interfaces. Returns domain errors to the Interface layer.

### Layer 3: Persistence (`store.go`)
*   **Responsibility:** Database interactions.
*   **Role:** Executes SQL against PostgreSQL. Maps rows to Go structs and DB errors to domain errors.
*   **Rule:** Returns success or database errors. Contains **zero** business logic.

---

## 2. File Structure

Organized by domain, encapsulating all layers for that domain.

```text
server/
├── cmd/
│   └── api/
│       ├── main.go          # App entry point (wires dependencies)
│       ├── routes.go        # Route definitions
│       └── server.go        # HTTP server configuration and middleware
├── internal/
│   ├── auth/                # AUTH DOMAIN
│   │   ├── http.go          # Interface: Auth handlers & DTOs
│   │   ├── service.go       # Business: Auth logic (passwords, shadow accounts)
│   │   ├── store.go         # Persistence: SQL for accounts/tokens
│   │   └── auth_test.go     # Tests for Auth domain
│   ├── db/                  # INFRASTRUCTURE
│   │   ├── db.go            # DB connection pool setup
│   │   └── schema.sql       # Database schema / migrations
│   └── restaurants/         # RESTAURANTS DOMAIN
│       ├── http.go          
│       ├── service.go       
│       ├── store.go         
│       └── restaurants_test.go 
```

---

## 3. Testing Strategy: Stress & Boundaries

**MANDATORY RULE:** If you change *anything* in a domain, you must run and verify the tests for that domain. Do not merge code unless the tests for the modified domain pass.

We do not aim for 100% coverage by writing meaningless tests. We write tests that **stress potential failure points, validate bad input, and ensure business limits are strictly followed.**

### 3.1. The Interface Layer (`http_test.go`)
Go provides `net/http/httptest`, which creates in-memory fake requests and response recorders. It does not open a real network port. You pass the fake request directly to your handler and assert the response.

**Goal:** Ensure bad input is rejected gracefully and the API contract (JSON shapes, status codes) is maintained.

**What to Stress Test:**
*   **Malformed JSON:** Send `{ "email": ` and ensure it returns `400 Bad Request`.
*   **Missing Auth:** Hit an authenticated route without the `Authorization` header and assert `401`.
*   **Payload Limits:** Send a JSON body that exceeds the maximum allowed size.
*   **Domain Error Mapping:** Force the mocked Business layer to return `ErrNotFound` and assert the HTTP layer correctly translates it to a `404`.

### 3.2. The Business Layer (`service_test.go`)
**Goal:** Validate complex business logic, edge cases, and boundaries without hitting a real database.
**How:** Mock the Persistence layer interface.

**What to Stress Test (Boundaries & Failures):**
*   **The Shadow Account Race / Merge:** If a user registers with a CPF that exists on a shadow account, assert the service correctly calls the store to merge wallets and graduate the account.
*   **Last Auth Method Limit:** Try to call `DeletePhone()` on an account that has no email. Assert it returns `ErrCannotDeleteLastAuthMethod`.
*   **Financial Boundaries:** Attempt to issue a refund larger than the original transaction amount. Assert it fails before hitting the DB.
*   **Loyalty Limits:** Attempt to redeem coins when the user's wallet balance is lower than the requested amount.

### 3.3. The Persistence Layer (`store_test.go`)
**Goal:** Ensure SQL queries, constraints, and foreign keys actually work.
**How:** Integration tests using a real PostgreSQL database (via Docker/Testcontainers or a local test DB). **Never mock SQL.**

**What to Stress Test (Data Integrity & Concurrency):**
*   **Concurrency:** Fire 5 concurrent queries to insert the exact same `cpf` into the `account` table. Assert exactly one succeeds and the others return a unique constraint violation.
*   **Idempotency Locking:** (If using DB for locks) Ensure concurrent inserts with the same `idempotency_key` do not result in double charges.
*   **Soft Deletes:** Insert a record, set `archived_at`, and assert that a standard `GET` query does *not* return it.
*   **Foreign Keys:** Attempt to insert an `order` with a non-existent `physical_restaurant_id` and assert it fails with an FK violation.

---

## 4. Testing Examples (Go)

### Example 1: Interface Test (Bad Input)
```go
func TestCreateUser_InvalidJSON(t *testing.T) {
	reqBody := []byte(`{ "email": "test@test.com" `) // Malformed
	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBuffer(reqBody))
	rec := httptest.NewRecorder()

	handler := &Handler{service: nil} // Service won't be reached
	handler.Register(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}
```

### Example 2: Business Test (Mocking boundaries)
```go
func TestRedeemCoins_InsufficientBalance(t *testing.T) {
	// 1. Setup mock
	mockStore := new(MockStore)
	mockStore.On("GetWallet", ctx, userID).Return(&Wallet{Coins: 50}, nil)
	
	svc := NewService(mockStore)

	// 2. Attempt to redeem more than available
	err := svc.RedeemCoins(ctx, userID, 100)

	// 3. Assert business rule enforced
	if !errors.Is(err, ErrInsufficientFunds) {
		t.Errorf("expected ErrInsufficientFunds, got %v", err)
	}
}
```

### Example 3: Persistence Test (Concurrency)
```go
func TestAccountStore_ConcurrentShadowCreation(t *testing.T) {
	db := setupTestDB()
	store := NewStore(db)
	cpf := "12345678901"

	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// Upsert should not panic or throw an error on concurrent calls
			_, err := store.FindOrCreateShadow(context.Background(), cpf)
			if err != nil {
				t.Errorf("unexpected error in concurrent upsert: %v", err)
			}
		}()
	}
	wg.Wait()

	// Assert only 1 row was created
	var count int
	db.QueryRow("SELECT count(*) FROM account WHERE cpf = $1", cpf).Scan(&count)
	if count != 1 {
		t.Errorf("expected 1 account, got %d", count)
	}
}
```

---

## 5. Development Workflow

When adding a new feature:
1.  **Persistence (`store.go`):** Write the SQL queries. Add integration tests with edge cases.
2.  **Business (`service.go`):** Write the logic that calls the store. Add unit tests using the mocked store (focus on limits and bad logic).
3.  **Interface (`http.go`):** Write the HTTP handlers. Add `httptest` tests for JSON validation and status codes.
4.  **Run Tests:** `go test ./internal/<domain>/... -race` (Always use `-race` to catch concurrency bugs).



IMPORTANT:
Whenever you update casos de testes or openapi.yaml add a minor version to it 0.0.1
Every time you have a question on what should be done, dont assume the answer, ask me for the decision
If You change something in the db schema, ask me to procedd, if i allow you to proced then change the db schema and run the tests to make sure it works before continuing
at each domain i want you to have a http.go service.go store.go <domain>_test.go and <domain>_casos_de_testes.go no matter how long the files are, keep it on that structure

