package db

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

func RunMigrations(ctx context.Context, pool *pgxpool.Pool, migrationsDir string) error {
	if err := createMigrationsTable(ctx, pool); err != nil {
		return fmt.Errorf("create migrations table: %w", err)
	}

	files, err := os.ReadDir(migrationsDir)
	if err != nil {
		return fmt.Errorf("read migrations dir: %w", err)
	}

	var sqlFiles []string
	for _, f := range files {
		if !f.IsDir() && strings.HasSuffix(f.Name(), ".sql") {
			sqlFiles = append(sqlFiles, f.Name())
		}
	}
	sort.Strings(sqlFiles)

	for _, file := range sqlFiles {
		applied, err := isMigrationApplied(ctx, pool, file)
		if err != nil {
			return fmt.Errorf("check migration %s: %w", file, err)
		}
		if applied {
			continue
		}

		content, err := os.ReadFile(filepath.Join(migrationsDir, file))
		if err != nil {
			return fmt.Errorf("read migration %s: %w", file, err)
		}

		tx, err := pool.Begin(ctx)
		if err != nil {
			return fmt.Errorf("begin tx: %w", err)
		}

		if _, err := tx.Exec(ctx, string(content)); err != nil {
			tx.Rollback(ctx)
			return fmt.Errorf("execute migration %s: %w", file, err)
		}

		if _, err := tx.Exec(ctx, "INSERT INTO _migrations (name) VALUES ($1)", file); err != nil {
			tx.Rollback(ctx)
			return fmt.Errorf("record migration %s: %w", file, err)
		}

		if err := tx.Commit(ctx); err != nil {
			return fmt.Errorf("commit migration %s: %w", file, err)
		}

		fmt.Printf("Applied migration: %s\n", file)
	}

	return nil
}

func createMigrationsTable(ctx context.Context, pool *pgxpool.Pool) error {
	_, err := pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS _migrations (
			id SERIAL PRIMARY KEY,
			name VARCHAR(255) NOT NULL UNIQUE,
			applied_at TIMESTAMPTZ DEFAULT NOW()
		)
	`)
	return err
}

func isMigrationApplied(ctx context.Context, pool *pgxpool.Pool, name string) (bool, error) {
	var count int
	err := pool.QueryRow(ctx, "SELECT COUNT(*) FROM _migrations WHERE name = $1", name).Scan(&count)
	return count > 0, err
}
