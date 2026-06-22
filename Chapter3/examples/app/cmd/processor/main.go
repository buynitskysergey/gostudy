package main

import (
	"context"
	"fmt"
	"time"

	"golang.org/x/sync/errgroup"

	"study1/Chapter3/examples/app/internal/pipeline"
)

func main() {
	tasks := make([]pipeline.Task, 15)
	for i := range tasks {
		tasks[i] = pipeline.Task{ID: i + 1}
	}

	fmt.Println("=== Full pipeline (fan-out/fan-in + worker pool) ===")
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	results, err := pipeline.Process(ctx, tasks, 4)
	if err != nil {
		fmt.Println("error:", err)
	}
	fmt.Printf("processed %d tasks\n", len(results))

	fmt.Println("\n=== errgroup parallel stage ===")
	g, gctx := errgroup.WithContext(context.Background())
	g.SetLimit(3)
	for i := 1; i <= 5; i++ {
		i := i
		g.Go(func() error {
			select {
			case <-time.After(30 * time.Millisecond):
				fmt.Printf("errgroup task %d done\n", i)
				return nil
			case <-gctx.Done():
				return gctx.Err()
			}
		})
	}
	fmt.Println("errgroup:", g.Wait())
}
