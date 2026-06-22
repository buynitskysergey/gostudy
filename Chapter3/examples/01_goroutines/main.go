package main

import (
	"fmt"
	"sync"
	"time"
)

func main() {
	fmt.Println("=== With WaitGroup (correct) ===")
	var wg sync.WaitGroup
	for i := 1; i <= 3; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			time.Sleep(50 * time.Millisecond)
			fmt.Printf("worker %d done\n", id)
		}(i)
	}
	wg.Wait()
	fmt.Println("all workers finished")

	fmt.Println("\n=== Without coordination (may print nothing — race with main exit) ===")
	for i := 1; i <= 3; i++ {
		go func(id int) {
			time.Sleep(100 * time.Millisecond)
			fmt.Printf("orphan worker %d (main may already exit)\n", id)
		}(i)
	}
	time.Sleep(150 * time.Millisecond) // demo-only wait; use WaitGroup in production
}
