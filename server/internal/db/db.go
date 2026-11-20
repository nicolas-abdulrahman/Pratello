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
	// CREATE TABLE IF NOT EXISTS does not upgrade old table layouts. Refuse
	// known incompatible baselines before applying partial DDL to their data.
	var legacy bool
	if err := tx.QueryRow(`SELECT EXISTS (
		SELECT 1 FROM information_schema.columns WHERE table_schema=current_schema() AND (
		(table_name='account' AND column_name='password_hash') OR
		(table_name='loyalty_wallet' AND column_name='loyalty_customer_id') OR
		(table_name='menu_item' AND column_name='price') OR
		(table_name='physical_restaurant' AND column_name='operational_address_id'))
	)`).Scan(&legacy); err != nil {
		return fmt.Errorf("inspect database baseline: %w", err)
	}
	if legacy {
		return fmt.Errorf("pre-DBML database layout detected: import data into the DBML baseline before starting the API")
	}
	if _, err := tx.Exec(SchemaSQL); err != nil {
		return fmt.Errorf("failed to execute schema.sql: %w", err)
	}

	return tx.Commit()
}
