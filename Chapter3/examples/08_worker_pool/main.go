package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type Job struct {
	ID int
}

func process(ctx context.Context, job Job) int {
	select {
	case <-time.After(50 * time.Millisecond):
		return job.ID * job.ID
	case <-ctx.Done():
		return -1
	}
}

func main() {
	const workers = 3
	const jobCount = 12

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	jobs := make(chan Job, jobCount)
	results := make(chan int, jobCount)

	var wg sync.WaitGroup
	for w := 1; w <= workers; w++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for job := range jobs {
				select {
				case <-ctx.Done():
					return
				default:
					if r := process(ctx, job); r >= 0 {
						results <- r
					}
				}
			}
		}(w)
	}

	go func() {
		for i := 1; i <= jobCount; i++ {
			select {
			case jobs <- Job{ID: i}:
			case <-ctx.Done():
				close(jobs)
				return
			}
			if i == 8 {
				fmt.Println("cancelling after job 8 submitted")
				cancel()
			}
		}
		close(jobs)
	}()

	go func() {
		wg.Wait()
		close(results)
	}()

	fmt.Print("results: ")
	for r := range results {
		fmt.Print(r, " ")
	}
	fmt.Println()
}
