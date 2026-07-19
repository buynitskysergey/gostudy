package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

func main() {
	ctx := context.Background()
	db, err := sql.Open("sqlite", "file:ch5_05.db?cache=shared&mode=rwc")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	db.SetMaxOpenConns(1)

	dir := filepath.Join("migrations")
	if _, err := os.Stat(dir); err != nil {
		// go run из examples/05_migrations
		dir = filepath.Join("examples", "05_migrations", "migrations")
		if _, err := os.Stat(dir); err != nil {
			dir = "migrations"
		}
	}
	// resolve relative to this file's example dir
	if wd, err := os.Getwd(); err == nil {
		candidates := []string{
			filepath.Join(wd, "migrations"),
			filepath.Join(wd, "examples", "05_migrations", "migrations"),
			filepath.Join(wd, "Chapter5", "examples", "05_migrations", "migrations"),
		}
		for _, c := range candidates {
			if st, err := os.Stat(c); err == nil && st.IsDir() {
				dir = c
				break
			}
		}
	}

	if err := migrateUp(ctx, db, dir); err != nil {
		log.Fatal(err)
	}
	if err := migrateUp(ctx, db, dir); err != nil {
		log.Fatal(err)
	}
	fmt.Println("second up: no-op (already applied)")

	var n int
	_ = db.QueryRowContext(ctx, `SELECT COUNT(*) FROM schema_migrations`).Scan(&n)
	fmt.Println("versions applied:", n)

	var cols string
	_ = db.QueryRowContext(ctx, `SELECT sql FROM sqlite_master WHERE name='accounts'`).Scan(&cols)
	fmt.Println("accounts DDL:", cols)
}

func migrateUp(ctx context.Context, db *sql.DB, dir string) error {
	if _, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version TEXT PRIMARY KEY,
			applied_at TEXT NOT NULL
		)`); err != nil {
		return err
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	var ups []string
	for _, e := range entries {
		name := e.Name()
		if strings.HasSuffix(name, ".up.sql") {
			ups = append(ups, name)
		}
	}
	sort.Strings(ups)

	for _, name := range ups {
		version := strings.TrimSuffix(name, ".up.sql")
		var exists int
		if err := db.QueryRowContext(ctx,
			`SELECT COUNT(1) FROM schema_migrations WHERE version = ?`, version,
		).Scan(&exists); err != nil {
			return err
		}
		if exists > 0 {
			continue
		}

		body, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			return err
		}
		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, string(body)); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("%s: %w", name, err)
		}
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO schema_migrations(version, applied_at) VALUES (?, ?)`,
			version, time.Now().UTC().Format(time.RFC3339),
		); err != nil {
			_ = tx.Rollback()
			return err
		}
		if err := tx.Commit(); err != nil {
			return err
		}
		fmt.Println("applied", version)
	}
	return nil
}
