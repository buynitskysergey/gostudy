package main

import (
	"fmt"
	"sync"
	"time"
)

func withMutex() {
	var mu sync.Mutex
	counter := 0
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			mu.Lock()
			counter++
			mu.Unlock()
		}()
	}
	wg.Wait()
	fmt.Println("counter with Mutex:", counter)
}

func withRWMutex() {
	cache := map[string]int{"go": 1}
	var mu sync.RWMutex
	var wg sync.WaitGroup

	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			mu.RLock()
			_ = cache["go"]
			mu.RUnlock()
		}()
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		time.Sleep(10 * time.Millisecond)
		mu.Lock()
		cache["go"] = 2
		mu.Unlock()
	}()
	wg.Wait()
	fmt.Println("cache after RWMutex:", cache)
}

func main() {
	withMutex()
	withRWMutex()
	fmt.Println("Run: go run -race ./examples/04_mutex/  to detect data races")
}
