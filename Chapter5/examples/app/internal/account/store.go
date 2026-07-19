package account

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

func (s *Store) Create(ctx context.Context, owner string, balance int64) (Account, error) {
	now := time.Now().UTC()
	res, err := s.db.ExecContext(ctx, `
		INSERT INTO accounts(owner, balance_cents, version, created_at)
		VALUES (?, ?, 1, ?)`, owner, balance, now.Format(time.RFC3339Nano))
	if err != nil {
		return Account{}, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return Account{}, err
	}
	return Account{ID: id, Owner: owner, BalanceCents: balance, Version: 1, CreatedAt: now}, nil
}

func (s *Store) Get(ctx context.Context, id int64) (Account, error) {
	var a Account
	var created string
	err := s.db.QueryRowContext(ctx, `
		SELECT id, owner, balance_cents, version, created_at
		FROM accounts WHERE id = ?`, id,
	).Scan(&a.ID, &a.Owner, &a.BalanceCents, &a.Version, &created)
	if errors.Is(err, sql.ErrNoRows) {
		return Account{}, ErrNotFound
	}
	if err != nil {
		return Account{}, err
	}
	a.CreatedAt, _ = time.Parse(time.RFC3339Nano, created)
	return a, nil
}

type TransferRequest struct {
	FromID      int64 `json:"from_id"`
	ToID        int64 `json:"to_id"`
	AmountCents int64 `json:"amount_cents"`
}

type IdempotentResult struct {
	Replay     bool
	StatusCode int
	Body       []byte
}

// Transfer выполняет перевод в одной tx с optimistic locking и idempotency key.
func (s *Store) Transfer(ctx context.Context, key string, req TransferRequest) (IdempotentResult, error) {
	raw, _ := json.Marshal(req)
	hash := hashBytes(raw)

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return IdempotentResult{}, err
	}
	defer tx.Rollback()

	var existingHash string
	var status int
	var resp string
	err = tx.QueryRowContext(ctx, `
		SELECT request_hash, status_code, response FROM idempotency_keys WHERE key = ?`, key,
	).Scan(&existingHash, &status, &resp)
	switch {
	case err == nil:
		if existingHash != hash {
			return IdempotentResult{}, ErrIdempotencyConflict
		}
		return IdempotentResult{Replay: true, StatusCode: status, Body: []byte(resp)}, nil
	case !errors.Is(err, sql.ErrNoRows):
		return IdempotentResult{}, err
	}

	if req.FromID == req.ToID {
		return IdempotentResult{}, fmt.Errorf("from_id and to_id must differ")
	}
	if req.AmountCents <= 0 {
		return IdempotentResult{}, fmt.Errorf("amount must be > 0")
	}

	from, err := getForUpdate(ctx, tx, req.FromID)
	if err != nil {
		return IdempotentResult{}, err
	}
	to, err := getForUpdate(ctx, tx, req.ToID)
	if err != nil {
		return IdempotentResult{}, err
	}
	if from.BalanceCents < req.AmountCents {
		return IdempotentResult{}, ErrInsufficientFunds
	}

	if err := bumpBalance(ctx, tx, from.ID, from.BalanceCents-req.AmountCents, from.Version); err != nil {
		return IdempotentResult{}, err
	}
	if err := bumpBalance(ctx, tx, to.ID, to.BalanceCents+req.AmountCents, to.Version); err != nil {
		return IdempotentResult{}, err
	}

	now := time.Now().UTC()
	res, err := tx.ExecContext(ctx, `
		INSERT INTO transfers(from_id, to_id, amount_cents, created_at)
		VALUES (?, ?, ?, ?)`, req.FromID, req.ToID, req.AmountCents, now.Format(time.RFC3339Nano))
	if err != nil {
		return IdempotentResult{}, err
	}
	tid, _ := res.LastInsertId()
	tr := Transfer{
		ID: tid, FromID: req.FromID, ToID: req.ToID,
		AmountCents: req.AmountCents, CreatedAt: now,
	}
	body, _ := json.Marshal(tr)

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO idempotency_keys(key, request_hash, status_code, response, created_at)
		VALUES (?, ?, ?, ?, ?)`,
		key, hash, 201, string(body), now.Format(time.RFC3339Nano)); err != nil {
		return IdempotentResult{}, err
	}
	if err := tx.Commit(); err != nil {
		return IdempotentResult{}, err
	}
	return IdempotentResult{StatusCode: 201, Body: body}, nil
}

func getForUpdate(ctx context.Context, tx *sql.Tx, id int64) (Account, error) {
	var a Account
	var created string
	err := tx.QueryRowContext(ctx, `
		SELECT id, owner, balance_cents, version, created_at
		FROM accounts WHERE id = ?`, id,
	).Scan(&a.ID, &a.Owner, &a.BalanceCents, &a.Version, &created)
	if errors.Is(err, sql.ErrNoRows) {
		return Account{}, ErrNotFound
	}
	if err != nil {
		return Account{}, err
	}
	a.CreatedAt, _ = time.Parse(time.RFC3339Nano, created)
	return a, nil
}

func bumpBalance(ctx context.Context, tx *sql.Tx, id, newBal, expectedVersion int64) error {
	res, err := tx.ExecContext(ctx, `
		UPDATE accounts
		SET balance_cents = ?, version = version + 1
		WHERE id = ? AND version = ?`, newBal, id, expectedVersion)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrConflict
	}
	return nil
}

func hashBytes(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}
