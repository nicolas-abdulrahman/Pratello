package db

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"
	"strings"
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

// ResetDevelopmentData empties every application table in the active schema.
// Truncating together handles foreign keys without cascading to other schemas.
func ResetDevelopmentData(ctx context.Context, database *sql.DB) error {
	tx, err := database.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	rows, err := tx.QueryContext(ctx, `SELECT schemaname, tablename FROM pg_tables WHERE schemaname=current_schema() ORDER BY tablename`)
	if err != nil {
		return err
	}
	defer rows.Close()
	var tables []string
	for rows.Next() {
		var schema, table string
		if err := rows.Scan(&schema, &table); err != nil {
			return err
		}
		tables = append(tables, "ONLY "+pq.QuoteIdentifier(schema)+"."+pq.QuoteIdentifier(table))
	}
	if err := rows.Err(); err != nil {
		return err
	}
	if len(tables) == 0 {
		return fmt.Errorf("no application tables found in active schema")
	}
	if _, err := tx.ExecContext(ctx, `TRUNCATE `+strings.Join(tables, ", ")+` RESTART IDENTITY`); err != nil {
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
