package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	_ "modernc.org/sqlite"
)

func main() {
	ctx := context.Background()
	db := open(ctx)
	defer db.Close()

	fmt.Println("Atomicity without tx:")
	reset(ctx, db)
	_ = debitOnlyThenFail(ctx, db, 1, 400)
	printSum(ctx, db)

	fmt.Println("\nAtomicity with tx (rollback on error):")
	reset(ctx, db)
	err := debitInTxThenFail(ctx, db, 1, 400)
	fmt.Println("  err:", err)
	printSum(ctx, db)

	fmt.Println("\nConsistency via CHECK-like WHERE:")
	reset(ctx, db)
	err = transfer(ctx, db, 1, 2, 5000)
	fmt.Println("  overdraw err:", err)
	printSum(ctx, db)
}

func open(ctx context.Context) *sql.DB {
	db, err := sql.Open("sqlite", "file:ch5_08.db?cache=shared&mode=rwc")
	if err != nil {
		log.Fatal(err)
	}
	db.SetMaxOpenConns(1)
	must(db.ExecContext(ctx, `
		DROP TABLE IF EXISTS accounts;
		CREATE TABLE accounts (
			id INTEGER PRIMARY KEY,
			balance_cents INTEGER NOT NULL CHECK (balance_cents >= 0)
		)`))
	return db
}

func reset(ctx context.Context, db *sql.DB) {
	must(db.ExecContext(ctx, `DELETE FROM accounts`))
	must(db.ExecContext(ctx, `INSERT INTO accounts(id, balance_cents) VALUES (1, 1000), (2, 1000)`))
}

func debitOnlyThenFail(ctx context.Context, db *sql.DB, id, amount int64) error {
	_, err := db.ExecContext(ctx, `UPDATE accounts SET balance_cents = balance_cents - ? WHERE id = ?`, amount, id)
	if err != nil {
		return err
	}
	return fmt.Errorf("crash before credit")
}

func debitInTxThenFail(ctx context.Context, db *sql.DB, id, amount int64) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, `UPDATE accounts SET balance_cents = balance_cents - ? WHERE id = ?`, amount, id); err != nil {
		return err
	}
	return fmt.Errorf("crash before commit") // Rollback via defer → Atomicity
}

func transfer(ctx context.Context, db *sql.DB, from, to, amount int64) error {
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
		return fmt.Errorf("insufficient funds")
	}
	if _, err := tx.ExecContext(ctx, `UPDATE accounts SET balance_cents = balance_cents + ? WHERE id = ?`, amount, to); err != nil {
		return err
	}
	return tx.Commit()
}

func printSum(ctx context.Context, db *sql.DB) {
	var sum int64
	_ = db.QueryRowContext(ctx, `SELECT COALESCE(SUM(balance_cents),0) FROM accounts`).Scan(&sum)
	fmt.Println("  total balance:", sum)
}

func must(_ sql.Result, err error) {
	if err != nil {
		log.Fatal(err)
	}
}
