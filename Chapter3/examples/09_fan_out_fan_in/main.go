package main

import (
	"fmt"
	"sync"
	"time"
)

type Job struct {
	ID int
}

type Result struct {
	JobID  int
	Output int
}

func generator(jobs chan<- Job, count int) {
	defer close(jobs)
	for i := 1; i <= count; i++ {
		jobs <- Job{ID: i}
	}
}

func worker(id int, jobs <-chan Job, results chan<- Result, wg *sync.WaitGroup) {
	defer wg.Done()
	for job := range jobs {
		time.Sleep(30 * time.Millisecond)
		results <- Result{JobID: job.ID, Output: job.ID * 10}
		fmt.Printf("worker %d processed job %d\n", id, job.ID)
	}
}

func main() {
	const numWorkers = 3
	const numJobs = 9

	jobs := make(chan Job, numJobs)
	results := make(chan Result, numJobs)

	go generator(jobs, numJobs)

	var wg sync.WaitGroup
	for w := 1; w <= numWorkers; w++ {
		wg.Add(1)
		go worker(w, jobs, results, &wg)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	var total int
	for r := range results {
		total += r.Output
	}
	fmt.Printf("fan-in complete: %d results, sum=%d\n", numJobs, total)
}
