// Composition root: единственное место сборки зависимостей.
package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	"study1/Chapter2/examples/app/internal/order"
	"study1/Chapter2/examples/app/internal/storage/memory"
)

func main() {
	repo := memory.New()
	svc := order.NewService(repo)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	created, err := svc.Create(ctx, "ORD-1", 1000)
	if err != nil {
		panic(err)
	}
	fmt.Println("created:", created)

	fetched, err := svc.Get(ctx, "ORD-1")
	if err != nil {
		panic(err)
	}
	fmt.Println("fetched:", fetched)

	_, err = svc.Get(ctx, "missing")
	if errors.Is(err, order.ErrNotFound) {
		fmt.Println("expected: order not found (errors.Is through wrap chain)")
	}

	// Context timeout propagates to repository
	expired, cancelExpired := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancelExpired()
	time.Sleep(time.Millisecond)

	_, err = svc.Get(expired, "ORD-1")
	if errors.Is(err, context.DeadlineExceeded) {
		fmt.Println("expected: context deadline exceeded")
	}
}
