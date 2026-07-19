package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

func main() {
	ctx := context.Background()
	db, err := sql.Open("sqlite", "file:ch5_06.db?cache=shared&mode=rwc")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	// >1: N+1 держит Rows открытыми и параллельно делает QueryRow
	db.SetMaxOpenConns(4)

	mustExec(ctx, db, `DROP TABLE IF EXISTS transfers`)
	mustExec(ctx, db, `DROP TABLE IF EXISTS accounts`)
	mustExec(ctx, db, `
		CREATE TABLE accounts (
			id INTEGER PRIMARY KEY,
			owner TEXT NOT NULL
		)`)
	mustExec(ctx, db, `
		CREATE TABLE transfers (
			id INTEGER PRIMARY KEY,
			from_id INTEGER NOT NULL,
			amount_cents INTEGER NOT NULL,
			created_at TEXT NOT NULL
		)`)

	for i := 1; i <= 50; i++ {
		mustExec(ctx, db, `INSERT INTO accounts(id, owner) VALUES (?, ?)`, i, fmt.Sprintf("u%d", i))
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	for i := 1; i <= 200; i++ {
		mustExec(ctx, db, `INSERT INTO transfers(id, from_id, amount_cents, created_at) VALUES (?, ?, ?, ?)`,
			i, (i%50)+1, int64(i*10), now)
	}

	fmt.Println("=== N+1 ===")
	t0 := time.Now()
	n1 := sumOwnersN1(ctx, db)
	fmt.Printf("queries-ish loop, owners=%d, took %s\n", n1, time.Since(t0))

	fmt.Println("=== batch IN ===")
	t1 := time.Now()
	n2 := sumOwnersBatch(ctx, db)
	fmt.Printf("batch, owners=%d, took %s\n", n2, time.Since(t1))

	fmt.Println("=== EXPLAIN without index ===")
	explain(ctx, db, `SELECT * FROM transfers WHERE from_id = 1 ORDER BY created_at DESC LIMIT 20`)

	mustExec(ctx, db, `CREATE INDEX transfers_from_created_idx ON transfers(from_id, created_at)`)
	fmt.Println("=== EXPLAIN with index ===")
	explain(ctx, db, `SELECT * FROM transfers WHERE from_id = 1 ORDER BY created_at DESC LIMIT 20`)
}

func sumOwnersN1(ctx context.Context, db *sql.DB) int {
	ids := loadFromIDs(ctx, db, 40)
	seen := map[string]struct{}{}
	for _, fromID := range ids {
		var owner string
		// отдельный запрос на каждый id — классический N+1
		if err := db.QueryRowContext(ctx, `SELECT owner FROM accounts WHERE id = ?`, fromID).Scan(&owner); err != nil {
			log.Fatal(err)
		}
		seen[owner] = struct{}{}
	}
	return len(seen)
}

func sumOwnersBatch(ctx context.Context, db *sql.DB) int {
	fromIDs := loadFromIDs(ctx, db, 40)
	uniq := map[int64]struct{}{}
	args := make([]any, 0, len(fromIDs))
	for _, id := range fromIDs {
		if _, ok := uniq[id]; ok {
			continue
		}
		uniq[id] = struct{}{}
		args = append(args, id)
	}
	if len(args) == 0 {
		return 0
	}
	placeholders := strings.TrimRight(strings.Repeat("?,", len(args)), ",")
	q := `SELECT owner FROM accounts WHERE id IN (` + placeholders + `)`
	arows, err := db.QueryContext(ctx, q, args...)
	if err != nil {
		log.Fatal(err)
	}
	defer arows.Close()
	n := 0
	for arows.Next() {
		var owner string
		_ = arows.Scan(&owner)
		n++
	}
	return n
}

func loadFromIDs(ctx context.Context, db *sql.DB, limit int) []int64 {
	rows, err := db.QueryContext(ctx, `SELECT from_id FROM transfers LIMIT ?`, limit)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	var ids []int64
	for rows.Next() {
		var fromID int64
		_ = rows.Scan(&fromID)
		ids = append(ids, fromID)
	}
	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}
	return ids
}

func explain(ctx context.Context, db *sql.DB, q string) {
	rows, err := db.QueryContext(ctx, `EXPLAIN QUERY PLAN `+q)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		var id, parent, notused int
		var detail string
		_ = rows.Scan(&id, &parent, &notused, &detail)
		fmt.Println(" ", detail)
	}
}

func mustExec(ctx context.Context, db *sql.DB, q string, args ...any) {
	if _, err := db.ExecContext(ctx, q, args...); err != nil {
		log.Fatal(err)
	}
}
