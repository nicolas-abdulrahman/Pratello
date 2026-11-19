package db

import (
	"database/sql"
	_ "embed"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

//go:embed schema.sql
var SchemaSQL string

//go:embed legacy.sql
var legacySQL string

// Open opens a database connection with standard connection pooling settings.
func Open(connStr string) (*sql.DB, error) {
	database, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	database.SetMaxOpenConns(25)
	database.SetMaxIdleConns(10)
	database.SetConnMaxLifetime(5 * time.Minute)

	if err := database.Ping(); err != nil {
		_ = database.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return database, nil
}

// Migrate executes the embedded schema.sql against the database in an idempotent manner.
func Migrate(database *sql.DB) error {
	if database == nil {
		return fmt.Errorf("cannot migrate nil database")
	}

	tx, err := database.Begin()
	if err != nil {
		return fmt.Errorf("failed to start migration: %w", err)
	}
	defer tx.Rollback()
	if _, err := tx.Exec(legacySQL); err != nil {
		return fmt.Errorf("failed to migrate legacy schema: %w", err)
	}
	if _, err := tx.Exec(SchemaSQL); err != nil {
		return fmt.Errorf("failed to execute schema.sql: %w", err)
	}

	// Idempotent column migrations for tables that may have been created in legacy iterations
	columnPatches := []string{
		`ALTER TABLE "account" ADD COLUMN IF NOT EXISTS "name" VARCHAR(500);`,
		`ALTER TABLE "account" ADD COLUMN IF NOT EXISTS "created_at" TIMESTAMPTZ NOT NULL DEFAULT NOW();`,
		`ALTER TABLE "physical_restaurant" ADD COLUMN IF NOT EXISTS "name" VARCHAR(500) NOT NULL DEFAULT '';`,
		`ALTER TABLE "physical_restaurant" ADD COLUMN IF NOT EXISTS "owner_cnpj" VARCHAR(500) NOT NULL DEFAULT '';`,
		`ALTER TABLE "physical_restaurant" ADD COLUMN IF NOT EXISTS "terminal_pin" VARCHAR(6) NOT NULL DEFAULT '123456';`,
		`ALTER TABLE "physical_restaurant" ADD COLUMN IF NOT EXISTS "pin_updated_at" TIMESTAMPTZ NOT NULL DEFAULT NOW();`,
		`ALTER TABLE "physical_restaurant" ADD COLUMN IF NOT EXISTS "failed_pin_attempts" INT NOT NULL DEFAULT 0;`,
		`ALTER TABLE "physical_restaurant" ADD COLUMN IF NOT EXISTS "locked_until" TIMESTAMPTZ;`,
	}

	for _, patch := range columnPatches {
		if _, err := tx.Exec(patch); err != nil {
			return fmt.Errorf("failed to apply compatibility migration: %w", err)
		}
	}

	return tx.Commit()
}
