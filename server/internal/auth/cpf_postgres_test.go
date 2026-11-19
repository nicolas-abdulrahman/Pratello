package auth

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	db "pratello/server/internal/db"

	"github.com/lib/pq"
)

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
	var customerID string
	if err = database.QueryRow(`INSERT INTO account (cpf) VALUES ('12345678901') RETURNING id`).Scan(&shadowID); err != nil {
		t.Fatal(err)
	}
	if err = database.QueryRow(`INSERT INTO restaurant (slug,name) VALUES ('test','Test') RETURNING id`).Scan(&restaurantID); err != nil {
		t.Fatal(err)
	}
	if err = database.QueryRow(`INSERT INTO loyalty_customer (cpf,account_id) VALUES ('12345678901',$1) RETURNING id`, shadowID).Scan(&customerID); err != nil {
		t.Fatal(err)
	}
	if err = database.QueryRow(`INSERT INTO loyalty_wallet (loyalty_customer_id,restaurant_id,points_balance) VALUES ($1,$2,50) RETURNING id`, customerID, restaurantID).Scan(&walletID); err != nil {
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
	if err = database.QueryRow(`SELECT lc.account_id, lw.points_balance FROM loyalty_wallet lw JOIN loyalty_customer lc ON lc.id=lw.loyalty_customer_id WHERE lw.id=$1`, walletID).Scan(&owner, &balance); err != nil {
		t.Fatal(err)
	}
	if owner != a.ID || balance != 50 {
		t.Fatalf("wallet owner=%d balance=%d", owner, balance)
	}
	if err = store.RemoveCPF(ctx, a.ID); err != nil {
		t.Fatal(err)
	}
	if err = database.QueryRow(`SELECT account_id FROM loyalty_customer WHERE id=$1`, customerID).Scan(&owner); err != nil || owner != a.ID {
		t.Fatalf("wallet lost after unlink: %d %v", owner, err)
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
