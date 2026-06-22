package pipeline

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type Task struct {
	ID int
}

type Result struct {
	TaskID int
	Value  int
}

// Process runs fan-out workers over jobs, fan-in to results. Stops on ctx cancel.
func Process(ctx context.Context, tasks []Task, numWorkers int) ([]Result, error) {
	jobs := make(chan Task)
	results := make(chan Result)

	var wg sync.WaitGroup
	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for task := range jobs {
				select {
				case <-ctx.Done():
					return
				default:
					time.Sleep(20 * time.Millisecond)
					results <- Result{TaskID: task.ID, Value: task.ID * task.ID}
				}
			}
		}(w)
	}

	go func() {
		defer close(jobs)
		for _, t := range tasks {
			select {
			case jobs <- t:
			case <-ctx.Done():
				return
			}
		}
	}()

	go func() {
		wg.Wait()
		close(results)
	}()

	var out []Result
	for {
		select {
		case r, ok := <-results:
			if !ok {
				if err := ctx.Err(); err != nil {
					return out, fmt.Errorf("pipeline: %w", err)
				}
				return out, nil
			}
			out = append(out, r)
		case <-ctx.Done():
			return out, fmt.Errorf("pipeline: %w", ctx.Err())
		}
	}
}
