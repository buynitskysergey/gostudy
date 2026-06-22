package main

import (
	"fmt"
	"sync"
	"time"
)

func main() {
	const n = 5
	results := make(chan int, n)
	var wg sync.WaitGroup

	for i := 1; i <= n; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			time.Sleep(time.Duration(id*20) * time.Millisecond)
			results <- id * 10
		}(i)
	}

	wg.Wait()
	close(results)

	fmt.Print("results: ")
	for r := range results {
		fmt.Print(r, " ")
	}
	fmt.Println()
}
