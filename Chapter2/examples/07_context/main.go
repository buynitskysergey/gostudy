package main

import (
	"context"
	"errors"
	"fmt"
	"time"
)

func slowOperation(ctx context.Context) error {
	select {
	case <-time.After(100 * time.Millisecond):
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func main() {
	// Timeout — типичный паттерн для HTTP handler / service call
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	err := slowOperation(ctx)
	if errors.Is(err, context.DeadlineExceeded) {
		fmt.Println("operation cancelled: deadline exceeded")
	}

	// Cancel вручную
	ctx2, cancel2 := context.WithCancel(context.Background())
	cancel2() // немедленная отмена

	err = slowOperation(ctx2)
	if errors.Is(err, context.Canceled) {
		fmt.Println("operation cancelled: manual cancel")
	}

	// Request-scoped value (request ID) — умеренно
	type ctxKey struct{}
	ctx3 := context.WithValue(context.Background(), ctxKey{}, "req-abc-123")
	if id, ok := ctx3.Value(ctxKey{}).(string); ok {
		fmt.Println("request id:", id)
	}
}
