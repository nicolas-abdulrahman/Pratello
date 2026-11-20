package main

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"time"

	"pratello/server/internal/db"
	"pratello/server/internal/integration"
)

func resetDemo(ctx context.Context, database *sql.DB, dir string) error {
	if database == nil {
		return fmt.Errorf("DATABASE_URL is required for the development reset")
	}
	media, err := writableDemoMedia(dir)
	if err != nil {
		return err
	}
	if err := db.ResetDevelopmentData(ctx, database); err != nil {
		return fmt.Errorf("wipe database: %w", err)
	}
	if err := resetDemoMedia(media); err != nil {
		return fmt.Errorf("reset media: %w", err)
	}
	return nil
}

func seedDemo(ctx context.Context, database *sql.DB, handler http.Handler) (integration.JoaoBurger, error) {
	ctx, cancel := context.WithTimeout(ctx, 45*time.Second)
	defer cancel()
	// Seed over the real API on a private listener before exposing the app port.
	seedServer := httptest.NewServer(handler)
	defer seedServer.Close()
	return integration.SeedJoaoBurgerFresh(ctx, seedServer.Client(), seedServer.URL, func(ctx context.Context, id int) error {
		var updated int
		return database.QueryRowContext(ctx, `UPDATE account SET is_administrator=true WHERE id=$1 RETURNING id`, id).Scan(&updated)
	})
}

func writableDemoMedia(dir string) (string, error) {
	absolute, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}
	if filepath.Base(absolute) != "media" {
		return "", fmt.Errorf("development reset needs a directory named media: %s", absolute)
	}
	if err := os.MkdirAll(absolute, 0755); err != nil {
		return "", err
	}
	resolved, err := filepath.EvalSymlinks(absolute)
	if err != nil {
		return "", err
	}
	source, err := filepath.Abs(filepath.Join("server", "internal", "integration", "media"))
	if err != nil {
		return "", err
	}
	if resolved != absolute || strings.HasSuffix(resolved, string(filepath.Separator)+"integration"+string(filepath.Separator)+"media") || resolved == source {
		return "", fmt.Errorf("refusing to reset fixture media or a symlink: %s", absolute)
	}
	return resolved, nil
}

func resetDemoMedia(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if err := os.RemoveAll(filepath.Join(dir, entry.Name())); err != nil {
			return err
		}
	}
	return os.CopyFS(dir, integration.Media())
}
