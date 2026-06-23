package main

import (
	"fmt"
	"sync"
	"time"
)

func main() {
	fmt.Println("=== Block 1: async goroutines vs sync channel ===")

	// 1a: go не ждёт — main идёт дальше, порядок «inside» непредсказуем
	fmt.Println("--- 1a: launch in loop (async) ---")
	var wg sync.WaitGroup
	for i := 1; i <= 3; i++ {
		fmt.Printf("main: before go #%d\n", i)
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			time.Sleep(20 * time.Millisecond) // имитация работы
			fmt.Printf("goroutine #%d: inside (after sleep)\n", id)
		}(i)
		fmt.Printf("main: after go #%d  ← не ждёт goroutine\n", i)
	}
	wg.Wait()
	fmt.Println("main: all goroutines finished")

	// 1b: unbuffered channel — send/receive блокируют друг друга (sync handoff)
	fmt.Println("\n--- 1b: unbuffered channel (sync handoff) ---")
	ch := make(chan string)
	go func() {
		fmt.Println("goroutine: before send (blocks until receive)")
		ch <- "payload"
		fmt.Println("goroutine: after send")
	}()
	fmt.Println("main: before receive (blocks until send)")
	fmt.Println("main: received:", <-ch)
	fmt.Println("main: after receive")

	fmt.Println("\n=== Buffered: async until full ===")
	buf := make(chan int, 2)
	buf <- 1
	buf <- 2
	fmt.Println(<-buf, <-buf)

	fmt.Println("\n=== Close + range ===")
	nums := make(chan int)
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(nums)
		for i := 1; i <= 5; i++ {
			nums <- i
		}
	}()
	for n := range nums {
		fmt.Print(n, " ")
	}
	wg.Wait()
	fmt.Println()
}
