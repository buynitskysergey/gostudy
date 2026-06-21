package order

import "errors"

var (
	ErrNotFound      = errors.New("order not found")
	ErrInvalidAmount = errors.New("invalid amount")
)
