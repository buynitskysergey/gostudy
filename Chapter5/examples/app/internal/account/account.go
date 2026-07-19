package account

import (
	"errors"
	"time"
)

var (
	ErrNotFound           = errors.New("account not found")
	ErrInsufficientFunds  = errors.New("insufficient funds")
	ErrConflict           = errors.New("version conflict")
	ErrIdempotencyConflict = errors.New("idempotency key reuse with different body")
)

type Account struct {
	ID           int64     `json:"id"`
	Owner        string    `json:"owner"`
	BalanceCents int64     `json:"balance_cents"`
	Version      int64     `json:"version"`
	CreatedAt    time.Time `json:"created_at"`
}

type Transfer struct {
	ID          int64     `json:"id"`
	FromID      int64     `json:"from_id"`
	ToID        int64     `json:"to_id"`
	AmountCents int64     `json:"amount_cents"`
	CreatedAt   time.Time `json:"created_at"`
}
