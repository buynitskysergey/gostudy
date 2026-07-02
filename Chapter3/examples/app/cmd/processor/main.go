package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"golang.org/x/sync/errgroup"

	"study1/Chapter3/examples/app/internal/pipeline"
)

// runErrgroupDemo запускает 5 задач с лимитом 3.
// failTaskID > 0 — эта задача быстро возвращает ошибку; errgroup отменяет gctx для остальных.
func runErrgroupDemo(title string, failTaskID int) {
	pipeline.Log("main", title)
	g, gctx := errgroup.WithContext(context.Background())
	g.SetLimit(3)
	for i := 1; i <= 5; i++ {
		i := i
		g.Go(func() error {
			src := fmt.Sprintf("errgroup task %d", i)
			pipeline.Log(src, "старт")

			delay := 30 * time.Millisecond
			if i == failTaskID {
				delay = 10 * time.Millisecond // ошибка — быстро
			}

			select {
			case <-time.After(delay):
				if i == failTaskID {
					err := fmt.Errorf("simulated failure in task %d", i)
					pipeline.Log(src, fmt.Sprintf("возвращаю ошибку: %v", err))
					return err
				}
				pipeline.Log(src, "done")
				return nil
			case <-gctx.Done():
				pipeline.Log(src, fmt.Sprintf("отмена: %v", gctx.Err()))
				return gctx.Err()
			}
		})
	}
	if err := g.Wait(); err != nil {
		pipeline.Log("main", fmt.Sprintf("errgroup завершён с ошибкой: %v", err))
	} else {
		pipeline.Log("main", "errgroup: все задачи завершены")
	}
}

func main() {
	var numWorkers = 4
	if len(os.Args) > 1 {
		if n, err := strconv.Atoi(os.Args[1]); err == nil && n > 0 {
			numWorkers = n
		}
	}

	stopMemTrack := pipeline.StartPeakMemTracker()

	tasks := make([]pipeline.Task, 15)
	for i := range tasks {
		tasks[i] = pipeline.Task{ID: i + 1}
	}

	// ── Демо 1: полный pipeline (fan-out/fan-in + worker pool) ─────────────
	// Структура как в 09_fan_out_fan_in, но вынесена в пакет pipeline:
	// generator → jobs → N workers → results → fan-in.
	// Дополнительно: context.WithTimeout — pipeline останавливается по таймауту.
	pipeline.Log("main", fmt.Sprintf("=== Full pipeline (fan-out/fan-in + worker pool), workers=%d ===", numWorkers))
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	results, err := pipeline.Process(ctx, tasks, numWorkers)
	if err != nil {
		pipeline.Log("main", fmt.Sprintf("pipeline error: %v", err))
	}
	values := make([]string, len(results))
	for i, r := range results {
		values[i] = fmt.Sprintf("%d", r.Value)
	}
	pipeline.Log("main", fmt.Sprintf("processed %d tasks, values=[%s]",
		len(results), strings.Join(values, ", ")))

	fmt.Println("======================")
	pipeline.Log("main", "wait 1 second")
	time.Sleep(time.Second)
	fmt.Println("======================")

	// ── Демо 2: errgroup — параллельный запуск с лимитом и отменой по ошибке ─
	// SetLimit(3) ограничивает одновременно работающие goroutines.
	// WithContext: первая ошибка отменяет gctx для остальных задач.
	runErrgroupDemo("=== errgroup: успешное завершение (limit=3) ===", 0)

	fmt.Println("======================")
	pipeline.Log("main", "wait 1 second")
	time.Sleep(time.Second)
	fmt.Println("======================")

	// task 2 падает через 10ms → gctx отменяется → task 1, 3 прерываются,
	// task 4 и 5 либо не стартуют, либо сразу видят gctx.Done().
	runErrgroupDemo("=== errgroup: отмена по первой ошибке (limit=3, fail=task 2) ===", 4)

	peakHeap, peakSys := stopMemTrack()
	pipeline.Log("main", fmt.Sprintf("%d workers | пик памяти: heap=%s, sys=%s (всего от OS, включая стеки goroutines)",
		numWorkers, pipeline.FormatBytes(peakHeap), pipeline.FormatBytes(peakSys)))
}
