package db

import (
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/lib/pq"
)

// AUTH_TEST_DATABASE_URL must point at a database containing the legacy public
// schema. Tables are copied without data into a disposable schema.
func TestLegacyMigrationIsTransactionalAndPreservesPasswords(t *testing.T) {
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
	name := fmt.Sprintf("legacy_test_%d", time.Now().UnixNano())
	quoted := pq.QuoteIdentifier(name)
	if _, err = database.Exec(`CREATE SCHEMA ` + quoted); err != nil {
		t.Fatal(err)
	}
	defer database.Exec(`DROP SCHEMA ` + quoted + ` CASCADE`)
	_, err = database.Exec(`DO $$ DECLARE r record; BEGIN
		FOR r IN SELECT tablename FROM pg_catalog.pg_tables WHERE schemaname='public' LOOP
			EXECUTE format('CREATE TABLE %I.%I (LIKE public.%I INCLUDING ALL)', ` + pq.QuoteLiteral(name) + `, r.tablename, r.tablename);
		END LOOP;
	END $$`)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = database.Exec(`SET search_path TO ` + quoted); err != nil {
		t.Fatal(err)
	}
	if _, err = database.Exec(`DROP TABLE account; CREATE TABLE account (
		id SERIAL PRIMARY KEY, name VARCHAR(500), email VARCHAR(255) NOT NULL UNIQUE,
		password VARCHAR(255) NOT NULL, phone VARCHAR(500), created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	)`); err != nil {
		t.Fatal(err)
	}
	if _, err = database.Exec(`INSERT INTO account (id,email,password) VALUES (99999,'legacy@example.com','legacy-hash')`); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 2; i++ {
		if err = Migrate(database); err != nil {
			t.Fatalf("migration %d: %v", i+1, err)
		}
	}
	var hash string
	var verified bool
	if err = database.QueryRow(`SELECT password_hash,email_verified FROM account WHERE id=99999`).Scan(&hash, &verified); err != nil || hash != "legacy-hash" || !verified {
		t.Fatalf("legacy account: %q %t %v", hash, verified, err)
	}
	if _, err = database.Exec(`UPDATE account SET password_hash='new-hash' WHERE id=99999`); err != nil {
		t.Fatal(err)
	}
	if err = Migrate(database); err != nil {
		t.Fatal(err)
	}
	if err = database.QueryRow(`SELECT password_hash FROM account WHERE id=99999`).Scan(&hash); err != nil || hash != "new-hash" {
		t.Fatalf("password overwritten: %q %v", hash, err)
	}
	if _, err = database.Exec(`UPDATE account SET password_hash=NULL WHERE id=99999`); err != nil {
		t.Fatal(err)
	}
	if err = Migrate(database); err != nil {
		t.Fatal(err)
	}
	var removed bool
	if err = database.QueryRow(`SELECT password_hash IS NULL AND password IS NULL FROM account WHERE id=99999`).Scan(&removed); err != nil || !removed {
		t.Fatalf("removed password restored: removed=%t err=%v", removed, err)
	}
}
