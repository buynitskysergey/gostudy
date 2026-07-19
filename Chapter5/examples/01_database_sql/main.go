package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	_ "modernc.org/sqlite"
)

type Account struct {
	ID           int64
	Owner        string
	BalanceCents int64
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	db, err := sql.Open("sqlite", "file:ch5_01.db?cache=shared&mode=rwc")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	db.SetMaxOpenConns(1) // SQLite: один writer

	if err := db.PingContext(ctx); err != nil {
		log.Fatal(err)
	}

	_, err = db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS accounts (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			owner TEXT NOT NULL,
			balance_cents INTEGER NOT NULL
		)`)
	if err != nil {
		log.Fatal(err)
	}

	id, err := createAccount(ctx, db, "alice", 10_000)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("created id:", id)

	a, err := getAccount(ctx, db, id)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("get: %+v\n", a)

	if err := updateBalance(ctx, db, id, 12_500); err != nil {
		log.Fatal(err)
	}

	list, err := listAccounts(ctx, db)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("list:")
	for _, x := range list {
		fmt.Printf("  %+v\n", x)
	}

	_, err = getAccount(ctx, db, 99999)
	fmt.Println("missing:", err) // sql.ErrNoRows wrapped
}

func createAccount(ctx context.Context, db *sql.DB, owner string, balance int64) (int64, error) {
	res, err := db.ExecContext(ctx,
		`INSERT INTO accounts(owner, balance_cents) VALUES (?, ?)`, owner, balance)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func getAccount(ctx context.Context, db *sql.DB, id int64) (Account, error) {
	var a Account
	err := db.QueryRowContext(ctx,
		`SELECT id, owner, balance_cents FROM accounts WHERE id = ?`, id,
	).Scan(&a.ID, &a.Owner, &a.BalanceCents)
	if errors.Is(err, sql.ErrNoRows) {
		return Account{}, fmt.Errorf("account %d: %w", id, err)
	}
	return a, err
}

func updateBalance(ctx context.Context, db *sql.DB, id, balance int64) error {
	_, err := db.ExecContext(ctx,
		`UPDATE accounts SET balance_cents = ? WHERE id = ?`, balance, id)
	return err
}

func listAccounts(ctx context.Context, db *sql.DB) ([]Account, error) {
	rows, err := db.QueryContext(ctx, `SELECT id, owner, balance_cents FROM accounts ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Account
	for rows.Next() {
		var a Account
		if err := rows.Scan(&a.ID, &a.Owner, &a.BalanceCents); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}
