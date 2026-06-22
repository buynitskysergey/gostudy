package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

func worker(ctx context.Context, id int, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		select {
		case <-ctx.Done():
			fmt.Printf("worker %d stopped: %v\n", id, ctx.Err())
			return
		default:
			time.Sleep(30 * time.Millisecond)
			fmt.Printf("worker %d tick\n", id)
		}
	}
}

func main() {
	fmt.Println("=== Parent cancel stops workers ===")
	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	for i := 1; i <= 3; i++ {
		wg.Add(1)
		go worker(ctx, i, &wg)
	}
	time.Sleep(100 * time.Millisecond)
	cancel()
	wg.Wait()

	fmt.Println("\n=== Timeout ===")
	ctx, cancel = context.WithTimeout(context.Background(), 60*time.Millisecond)
	defer cancel()
	done := make(chan struct{})
	go func() {
		select {
		case <-time.After(200 * time.Millisecond):
		case <-ctx.Done():
			fmt.Println("timeout:", ctx.Err())
		}
		close(done)
	}()
	<-done
}
