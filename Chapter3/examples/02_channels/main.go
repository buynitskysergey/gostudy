package main

import (
	"fmt"
	"sync"
)

func main() {
	fmt.Println("=== Unbuffered: sync handoff ===")
	ch := make(chan string)
	go func() {
		ch <- "payload"
	}()
	fmt.Println("received:", <-ch)

	fmt.Println("\n=== Buffered: async until full ===")
	buf := make(chan int, 2)
	buf <- 1
	buf <- 2
	fmt.Println(<-buf, <-buf)

	fmt.Println("\n=== Close + range ===")
	nums := make(chan int)
	var wg sync.WaitGroup
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
