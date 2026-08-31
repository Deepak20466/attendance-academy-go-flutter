package db

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

// RunMigrations applies every *.up.sql file in dir, in filename order, that
// hasn't already been recorded in schema_migrations. Each migration runs in
// its own transaction so a partial failure doesn't leave a half-applied file
// marked as done.
func RunMigrations(ctx context.Context, pool *pgxpool.Pool, dir string) error {
	if _, err := pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version VARCHAR(255) PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT now()
		)
	`); err != nil {
		return fmt.Errorf("ensure schema_migrations: %w", err)
	}

	files, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("read migrations dir %s: %w", dir, err)
	}

	var versions []string
	for _, f := range files {
		if f.IsDir() || !strings.HasSuffix(f.Name(), ".up.sql") {
			continue
		}
		versions = append(versions, strings.TrimSuffix(f.Name(), ".up.sql"))
	}
	sort.Strings(versions)

	applied := map[string]bool{}
	rows, err := pool.Query(ctx, `SELECT version FROM schema_migrations`)
	if err != nil {
		return fmt.Errorf("query applied migrations: %w", err)
	}
	for rows.Next() {
		var v string
		if err := rows.Scan(&v); err != nil {
			rows.Close()
			return err
		}
		applied[v] = true
	}
	rows.Close()

	for _, version := range versions {
		if applied[version] {
			continue
		}
		content, err := os.ReadFile(filepath.Join(dir, version+".up.sql"))
		if err != nil {
			return fmt.Errorf("read %s: %w", version, err)
		}

		tx, err := pool.Begin(ctx)
		if err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, string(content)); err != nil {
			tx.Rollback(ctx)
			return fmt.Errorf("apply migration %s: %w", version, err)
		}
		if _, err := tx.Exec(ctx, `INSERT INTO schema_migrations (version) VALUES ($1)`, version); err != nil {
			tx.Rollback(ctx)
			return err
		}
		if err := tx.Commit(ctx); err != nil {
			return err
		}
		log.Printf("migrate: applied %s", version)
	}

	return nil
}
