package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "modernc.org/sqlite"
)

func main() {
	ctx := context.Background()
	db := mustOpen(ctx)
	defer db.Close()
	seed(ctx, db)

	fmt.Println("--- without transaction (crash mid-flight) ---")
	_ = transferBroken(ctx, db, 1, 2, 300)
	printBalances(ctx, db)

	// reset
	_, _ = db.ExecContext(ctx, `UPDATE accounts SET balance_cents = 1000 WHERE id IN (1,2)`)

	fmt.Println("--- with transaction ---")
	if err := transferOK(ctx, db, 1, 2, 300); err != nil {
		log.Fatal(err)
	}
	printBalances(ctx, db)

	fmt.Println("--- insufficient funds (rolled back) ---")
	err := transferOK(ctx, db, 1, 2, 50_000)
	fmt.Println("error:", err)
	printBalances(ctx, db)
}

func mustOpen(ctx context.Context) *sql.DB {
	db, err := sql.Open("sqlite", "file:ch5_03.db?cache=shared&mode=rwc")
	if err != nil {
		log.Fatal(err)
	}
	db.SetMaxOpenConns(1)
	if err := db.PingContext(ctx); err != nil {
		log.Fatal(err)
	}
	_, err = db.ExecContext(ctx, `
		DROP TABLE IF EXISTS accounts;
		CREATE TABLE accounts (
			id INTEGER PRIMARY KEY,
			owner TEXT NOT NULL,
			balance_cents INTEGER NOT NULL
		)`)
	if err != nil {
		log.Fatal(err)
	}
	return db
}

func seed(ctx context.Context, db *sql.DB) {
	_, err := db.ExecContext(ctx, `
		INSERT INTO accounts(id, owner, balance_cents) VALUES
			(1, 'alice', 1000),
			(2, 'bob', 1000)`)
	if err != nil {
		log.Fatal(err)
	}
}

// transferBroken: списание прошло, зачисление «упало» — деньги исчезли.
func transferBroken(ctx context.Context, db *sql.DB, from, to, amount int64) error {
	_, err := db.ExecContext(ctx,
		`UPDATE accounts SET balance_cents = balance_cents - ? WHERE id = ?`, amount, from)
	if err != nil {
		return err
	}
	// имитация сбоя до второго UPDATE
	fmt.Println("  simulated crash after debit")
	_ = to
	return fmt.Errorf("boom")
}

func transferOK(ctx context.Context, db *sql.DB, from, to, amount int64) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	res, err := tx.ExecContext(ctx,
		`UPDATE accounts SET balance_cents = balance_cents - ? WHERE id = ? AND balance_cents >= ?`,
		amount, from, amount)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("insufficient funds or unknown from=%d", from)
	}
	if _, err := tx.ExecContext(ctx,
		`UPDATE accounts SET balance_cents = balance_cents + ? WHERE id = ?`, amount, to); err != nil {
		return err
	}
	return tx.Commit()
}

func printBalances(ctx context.Context, db *sql.DB) {
	rows, err := db.QueryContext(ctx, `SELECT id, owner, balance_cents FROM accounts ORDER BY id`)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	var sum int64
	for rows.Next() {
		var id, bal int64
		var owner string
		_ = rows.Scan(&id, &owner, &bal)
		sum += bal
		fmt.Printf("  %s(%d)=%d\n", owner, id, bal)
	}
	fmt.Println("  sum:", sum, "(ожидаем 2000 если инвариант жив)")
	time.Sleep(10 * time.Millisecond)
}
