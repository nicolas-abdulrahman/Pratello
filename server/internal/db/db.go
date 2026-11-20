package db

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"
	"time"

	"github.com/lib/pq"
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

// ResetDevelopmentData recreates the active development schema from the current
// application definition. It does not inspect or retain its previous layout.
func ResetDevelopmentData(ctx context.Context, database *sql.DB) error {
	tx, err := database.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var schema string
	if err := tx.QueryRowContext(ctx, `SELECT current_schema()`).Scan(&schema); err != nil {
		return fmt.Errorf("resolve development schema: %w", err)
	}
	quotedSchema := pq.QuoteIdentifier(schema)
	if _, err := tx.ExecContext(ctx, `SET LOCAL search_path TO `+quotedSchema); err != nil {
		return fmt.Errorf("select development schema: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `DROP SCHEMA `+quotedSchema+` CASCADE`); err != nil {
		return fmt.Errorf("drop development schema: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `CREATE SCHEMA `+quotedSchema); err != nil {
		return fmt.Errorf("recreate development schema: %w", err)
	}
	if err := applySchema(ctx, tx); err != nil {
		return err
	}
	return tx.Commit()
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
	ctx := context.Background()
	legacy, err := legacyLayout(ctx, tx)
	if err != nil {
		return err
	}
	if legacy {
		return fmt.Errorf("pre-DBML database layout detected: import data into the DBML baseline before starting the API")
	}
	if err := applySchema(ctx, tx); err != nil {
		return err
	}
	return tx.Commit()
}

func legacyLayout(ctx context.Context, tx *sql.Tx) (bool, error) {
	// CREATE TABLE IF NOT EXISTS does not upgrade old table layouts. Refuse
	// known incompatible baselines before applying partial DDL to their data.
	var legacy bool
	if err := tx.QueryRowContext(ctx, `SELECT EXISTS (
		SELECT 1 FROM information_schema.columns WHERE table_schema=current_schema() AND (
		(table_name='account' AND column_name='password_hash') OR
		(table_name='loyalty_wallet' AND column_name='loyalty_customer_id') OR
		(table_name='menu_item' AND column_name='price') OR
		(table_name='physical_restaurant' AND column_name='operational_address_id'))
	)`).Scan(&legacy); err != nil {
		return false, fmt.Errorf("inspect database baseline: %w", err)
	}
	return legacy, nil
}

func applySchema(ctx context.Context, tx *sql.Tx) error {
	if _, err := tx.ExecContext(ctx, SchemaSQL); err != nil {
		return fmt.Errorf("failed to execute schema.sql: %w", err)
	}
	return nil
}
