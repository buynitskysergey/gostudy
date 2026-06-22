package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	"golang.org/x/sync/errgroup"
)

func fetch(ctx context.Context, name string, delay time.Duration, fail bool) error {
	select {
	case <-time.After(delay):
		if fail {
			return fmt.Errorf("%s: unavailable", name)
		}
		fmt.Println(name, "ok")
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func main() {
	fmt.Println("=== All succeed ===")
	g, ctx := errgroup.WithContext(context.Background())
	g.Go(func() error { return fetch(ctx, "users", 50*time.Millisecond, false) })
	g.Go(func() error { return fetch(ctx, "orders", 80*time.Millisecond, false) })
	fmt.Println("Wait:", g.Wait())

	fmt.Println("\n=== First error cancels others ===")
	g, ctx = errgroup.WithContext(context.Background())
	g.Go(func() error { return fetch(ctx, "payments", 200*time.Millisecond, false) })
	g.Go(func() error { return fetch(ctx, "inventory", 30*time.Millisecond, true) })
	err := g.Wait()
	fmt.Println("Wait:", err)
	fmt.Println("ctx err:", errors.Is(ctx.Err(), context.Canceled))

	fmt.Println("\n=== SetLimit(2) ===")
	g, ctx = errgroup.WithContext(context.Background())
	g.SetLimit(2)
	for i := 1; i <= 6; i++ {
		i := i
		g.Go(func() error {
			fmt.Printf("job %d start\n", i)
			time.Sleep(30 * time.Millisecond)
			fmt.Printf("job %d done\n", i)
			return nil
		})
	}
	fmt.Println("Wait:", g.Wait())
}
