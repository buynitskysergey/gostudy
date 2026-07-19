package main

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"

	_ "modernc.org/sqlite"
)

func main() {
	ctx := context.Background()
	db := open(ctx)
	defer db.Close()

	body := []byte(`{"from":1,"to":2,"amount":250}`)
	key := "transfer-demo-1"

	r1, err := transferIdempotent(ctx, db, key, body)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("1st: status=%d body=%s\n", r1.Status, r1.Body)

	r2, err := transferIdempotent(ctx, db, key, body)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("2nd (same key): status=%d body=%s\n", r2.Status, r2.Body)

	printBalances(ctx, db)

	_, err = transferIdempotent(ctx, db, key, []byte(`{"from":1,"to":2,"amount":999}`))
	fmt.Println("same key different body:", err)
}

type storedResp struct {
	Status int
	Body   string
}

func open(ctx context.Context) *sql.DB {
	db, err := sql.Open("sqlite", "file:ch5_09.db?cache=shared&mode=rwc")
	if err != nil {
		log.Fatal(err)
	}
	db.SetMaxOpenConns(1)
	must(db.ExecContext(ctx, `
		DROP TABLE IF EXISTS idempotency_keys;
		DROP TABLE IF EXISTS accounts;
		CREATE TABLE accounts (
			id INTEGER PRIMARY KEY,
			balance_cents INTEGER NOT NULL
		);
		CREATE TABLE idempotency_keys (
			key TEXT PRIMARY KEY,
			request_hash TEXT NOT NULL,
			status_code INTEGER NOT NULL,
			response TEXT NOT NULL
		);
		INSERT INTO accounts(id, balance_cents) VALUES (1, 1000), (2, 1000);
	`))
	return db
}

func transferIdempotent(ctx context.Context, db *sql.DB, key string, body []byte) (storedResp, error) {
	hash := hashBody(body)
	var req struct {
		From   int64 `json:"from"`
		To     int64 `json:"to"`
		Amount int64 `json:"amount"`
	}
	if err := json.Unmarshal(body, &req); err != nil {
		return storedResp{}, err
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return storedResp{}, err
	}
	defer tx.Rollback()

	var existingHash, resp string
	var status int
	err = tx.QueryRowContext(ctx,
		`SELECT request_hash, status_code, response FROM idempotency_keys WHERE key = ?`, key,
	).Scan(&existingHash, &status, &resp)
	if err == nil {
		if existingHash != hash {
			return storedResp{}, fmt.Errorf("idempotency key reuse with different body")
		}
		return storedResp{Status: status, Body: resp}, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return storedResp{}, err
	}

	res, err := tx.ExecContext(ctx,
		`UPDATE accounts SET balance_cents = balance_cents - ? WHERE id = ? AND balance_cents >= ?`,
		req.Amount, req.From, req.Amount)
	if err != nil {
		return storedResp{}, err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return storedResp{}, fmt.Errorf("insufficient funds")
	}
	if _, err := tx.ExecContext(ctx,
		`UPDATE accounts SET balance_cents = balance_cents + ? WHERE id = ?`, req.Amount, req.To); err != nil {
		return storedResp{}, err
	}

	out := fmt.Sprintf(`{"ok":true,"amount":%d}`, req.Amount)
	if _, err := tx.ExecContext(ctx,
		`INSERT INTO idempotency_keys(key, request_hash, status_code, response) VALUES (?,?,?,?)`,
		key, hash, 201, out); err != nil {
		return storedResp{}, err
	}
	if err := tx.Commit(); err != nil {
		return storedResp{}, err
	}
	return storedResp{Status: 201, Body: out}, nil
}

func hashBody(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

func printBalances(ctx context.Context, db *sql.DB) {
	rows, _ := db.QueryContext(ctx, `SELECT id, balance_cents FROM accounts ORDER BY id`)
	defer rows.Close()
	for rows.Next() {
		var id, bal int64
		_ = rows.Scan(&id, &bal)
		fmt.Printf("  account %d = %d\n", id, bal)
	}
}

func must(_ sql.Result, err error) {
	if err != nil {
		log.Fatal(err)
	}
}
