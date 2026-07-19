package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"sync"

	_ "modernc.org/sqlite"
)

var ErrConflict = errors.New("version conflict")

func main() {
	ctx := context.Background()
	db, err := sql.Open("sqlite", "file:ch5_10.db?cache=shared&mode=rwc")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	db.SetMaxOpenConns(1)

	must(db.ExecContext(ctx, `
		DROP TABLE IF EXISTS accounts;
		CREATE TABLE accounts (
			id INTEGER PRIMARY KEY,
			balance_cents INTEGER NOT NULL,
			version INTEGER NOT NULL
		);
		INSERT INTO accounts(id, balance_cents, version) VALUES (1, 1000, 1);
	`))

	// Два «клиента» прочитали одну version, оба пытаются записать.
	bal, ver, err := read(ctx, db, 1)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("read: balance=%d version=%d\n", bal, ver)

	var wg sync.WaitGroup
	errs := make(chan error, 2)
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func(delta int64) {
			defer wg.Done()
			errs <- update(ctx, db, 1, bal+delta, ver)
		}(int64(100 + i))
	}
	wg.Wait()
	close(errs)

	var ok, conflict int
	for e := range errs {
		switch {
		case e == nil:
			ok++
			fmt.Println("update: ok")
		case errors.Is(e, ErrConflict):
			conflict++
			fmt.Println("update:", e)
		default:
			log.Fatal(e)
		}
	}
	bal, ver, _ = read(ctx, db, 1)
	fmt.Printf("final: balance=%d version=%d (ok=%d conflict=%d)\n", bal, ver, ok, conflict)
}

func read(ctx context.Context, db *sql.DB, id int64) (balance int64, version int64, err error) {
	err = db.QueryRowContext(ctx,
		`SELECT balance_cents, version FROM accounts WHERE id = ?`, id,
	).Scan(&balance, &version)
	return
}

func update(ctx context.Context, db *sql.DB, id, newBalance, expectedVersion int64) error {
	res, err := db.ExecContext(ctx, `
		UPDATE accounts
		SET balance_cents = ?, version = version + 1
		WHERE id = ? AND version = ?`,
		newBalance, id, expectedVersion)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrConflict
	}
	return nil
}

func must(_ sql.Result, err error) {
	if err != nil {
		log.Fatal(err)
	}
}
