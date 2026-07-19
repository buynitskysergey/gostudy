// pgx: sqliteRepo реализует тот же AccountRepository, что и Postgres-адаптер.
// С DATABASE_URL — реальный pgxpool против Postgres (см. Chapter5/docker-compose.yml).
// Без DATABASE_URL (или с -force-sqlite) — SQLite ch5_01.db.
package main

import (
	"context"
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "modernc.org/sqlite"
)

type Account struct {
	ID           int64
	Owner        string
	BalanceCents int64
}

type AccountRepository interface {
	EnsureSchema(ctx context.Context) error
	Create(ctx context.Context, owner string, balance int64) (Account, error)
	Get(ctx context.Context, id int64) (Account, error)
	GetList(ctx context.Context, limit, offset int) ([]Account, error)
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var repo AccountRepository
	pgUrl := os.Getenv("DATABASE_URL")

	useSqlite := flag.Bool("force-sqlite", false, "disable postgres and use sqlite instead")
	flag.Parse()

	if *useSqlite {
		pgUrl = ""
	}

	if pgUrl != "" {
		pool, err := pgxpool.New(ctx, pgUrl)
		if err != nil {
			log.Fatalf("pgxpool: %v", err)
		}
		defer pool.Close()
		repo = &pgxRepo{pool: pool}
		fmt.Println("mode: pgx → Postgres")
		fmt.Println("hint: add -force-sqlite to use sqlite instead")
	} else {
		db, err := sql.Open("sqlite", "file:ch5_01.db?cache=shared&mode=rwc")
		if err != nil {
			log.Fatalf("sqlite: %v", err)
		}
		defer db.Close()
		db.SetMaxOpenConns(1)
		if err := db.PingContext(ctx); err != nil {
			log.Fatalf("sqlite ping: %v", err)
		}
		repo = &sqliteRepo{db: db}
		fmt.Println("mode: sqlite → ch5_01.db (set DATABASE_URL for real pgx)")
		fmt.Println("hint: docker compose -f Chapter5/docker-compose.yml up -d")
	}

	if err := repo.EnsureSchema(ctx); err != nil {
		log.Fatal(err)
	}

	// random user name & balance
	userName := fmt.Sprintf("user-%d", time.Now().UnixNano())
	balance := rand.Intn(1000000)
	a, err := repo.Create(ctx, userName, int64(balance))
	if err != nil {
		log.Fatal(err)
	}
	got, err := repo.Get(ctx, a.ID)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("created+get: %+v\n", got)

	list, err := repo.GetList(ctx, 10, 0)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("list:")
	for _, x := range list {
		fmt.Printf("  %+v\n", x)
	}
}

type pgxRepo struct {
	pool *pgxpool.Pool
}

func (r *pgxRepo) EnsureSchema(ctx context.Context) error {
	_, err := r.pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS accounts (
			id BIGSERIAL PRIMARY KEY,
			owner TEXT NOT NULL,
			balance_cents BIGINT NOT NULL
		)`)
	return err
}

func (r *pgxRepo) Create(ctx context.Context, owner string, balance int64) (Account, error) {
	var a Account
	err := r.pool.QueryRow(ctx,
		`INSERT INTO accounts(owner, balance_cents) VALUES ($1, $2)
		 RETURNING id, owner, balance_cents`, owner, balance,
	).Scan(&a.ID, &a.Owner, &a.BalanceCents)
	return a, err
}

func (r *pgxRepo) Get(ctx context.Context, id int64) (Account, error) {
	var a Account
	err := r.pool.QueryRow(ctx,
		`SELECT id, owner, balance_cents FROM accounts WHERE id = $1`, id,
	).Scan(&a.ID, &a.Owner, &a.BalanceCents)
	if err != nil {
		return Account{}, err
	}
	return a, nil
}

func (r *pgxRepo) GetList(ctx context.Context, limit, offset int) ([]Account, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, owner, balance_cents FROM accounts
		 ORDER BY id LIMIT $1 OFFSET $2`, limit, offset,
	)
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

// sqliteRepo — тот же контракт на SQLite (для go run без Docker/Postgres).
type sqliteRepo struct {
	db *sql.DB
}

func (r *sqliteRepo) EnsureSchema(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS accounts (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			owner TEXT NOT NULL,
			balance_cents INTEGER NOT NULL
		)`)
	return err
}

func (r *sqliteRepo) Create(ctx context.Context, owner string, balance int64) (Account, error) {
	res, err := r.db.ExecContext(ctx,
		`INSERT INTO accounts(owner, balance_cents) VALUES (?, ?)`, owner, balance)
	if err != nil {
		return Account{}, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return Account{}, err
	}
	return Account{ID: id, Owner: owner, BalanceCents: balance}, nil
}

func (r *sqliteRepo) Get(ctx context.Context, id int64) (Account, error) {
	var a Account
	err := r.db.QueryRowContext(ctx,
		`SELECT id, owner, balance_cents FROM accounts WHERE id = ?`, id,
	).Scan(&a.ID, &a.Owner, &a.BalanceCents)
	if errors.Is(err, sql.ErrNoRows) {
		return Account{}, fmt.Errorf("account %d: %w", id, err)
	}
	return a, err
}

func (r *sqliteRepo) GetList(ctx context.Context, limit, offset int) ([]Account, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, owner, balance_cents FROM accounts
		 ORDER BY id LIMIT ? OFFSET ?`, limit, offset,
	)
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
