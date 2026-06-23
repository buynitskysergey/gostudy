package main

import (
	"context"
	"fmt"
	"time"
)

var shouldCancel = true

func main() {
	fmt.Println("=== Timeout with select ===")
	ch := make(chan string)
	go func() {
		time.Sleep(200 * time.Millisecond)
		ch <- "slow result"
	}()

	select {
	case v := <-ch:
		fmt.Println("got:", v)
	case <-time.After(50 * time.Millisecond):
		fmt.Println("timeout — no result in 50ms")
	}

	fmt.Println("\n=== Context cancellation ===")
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		select {
		case <-time.After(500 * time.Millisecond):
			fmt.Println("work finished (won't happen — cancelled)")
		case <-ctx.Done():
			fmt.Println("cancelled:", ctx.Err())
		}
		close(done)
	}()
	if shouldCancel {
		cancel()
	}
	<-done

	fmt.Println("\n=== Disable case via nil channel ===")
	ch1 := make(chan int, 1)
	ch1 <- 1
	var ch2 chan int // nil — case disabled
	for ch1 != nil || ch2 != nil {
		select {
		case v, ok := <-ch1:
			if !ok {
				ch1 = nil
				continue
			}
			fmt.Println("ch1:", v)
			ch1 = nil
		case v, ok := <-ch2:
			if !ok {
				ch2 = nil
				continue
			}
			fmt.Println("ch2:", v)
		}
	}
}
