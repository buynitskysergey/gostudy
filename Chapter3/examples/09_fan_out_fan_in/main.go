package main

import (
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	programStart = time.Now()
	logMu        sync.Mutex
)

// formatBytes форматирует размер в человекочитаемый вид.
func formatBytes(b uint64) string {
	switch {
	case b >= 1<<20:
		return fmt.Sprintf("%.2f MiB", float64(b)/(1<<20))
	case b >= 1<<10:
		return fmt.Sprintf("%.2f KiB", float64(b)/(1<<10))
	default:
		return fmt.Sprintf("%d B", b)
	}
}

// startPeakMemTracker запускает goroutine, которая периодически снимает runtime.MemStats
// и запоминает максимум HeapInuse (куча) и Sys (всего от OS, включая стеки goroutines).
// Возвращает функцию stop: вызови её в конце main, чтобы получить зафиксированный пик.
func startPeakMemTracker() func() (heapPeak, sysPeak uint64) {
	stop := make(chan struct{})
	var wg sync.WaitGroup
	var mu sync.Mutex
	var peakHeap, peakSys uint64

	sample := func(ms *runtime.MemStats) {
		runtime.ReadMemStats(ms)
		mu.Lock()
		if ms.HeapInuse > peakHeap {
			peakHeap = ms.HeapInuse
		}
		if ms.Sys > peakSys {
			peakSys = ms.Sys
		}
		mu.Unlock()
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		var ms runtime.MemStats
		ticker := time.NewTicker(1 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				sample(&ms)
			case <-stop:
				sample(&ms)
				return
			}
		}
	}()

	return func() (uint64, uint64) {
		close(stop)
		wg.Wait()
		mu.Lock()
		defer mu.Unlock()
		return peakHeap, peakSys
	}
}

// logger выводит сообщение с меткой времени (ms от старта) и источником.
func logger(source, message string) {
	logMu.Lock()
	defer logMu.Unlock()
	elapsed := time.Since(programStart).Milliseconds()
	fmt.Printf("%dms | [%s]: %s\n", elapsed, source, message)
}

// Job — единица работы в pipeline: у каждой задачи есть идентификатор.
type Job struct {
	ID int
}

// Result — результат обработки задачи воркером.
type Result struct {
	JobID  int
	Output int
}

// generator (producer) кладёт numJobs задач в канал jobs и закрывает его.
// После close(jobs) все workers выйдут из `for job := range jobs`.
func generator(jobs chan<- Job, count int) {
	defer func() {
		logger("generator", "все задачи отправлены, закрываю jobs")
		close(jobs)
	}()
	logger("generator", fmt.Sprintf("старт: отправлю %d задач", count))
	for i := 1; i <= count; i++ {
		logger("generator", fmt.Sprintf("отправляет job %d", i))
		jobs <- Job{ID: i}
	}
}

// worker читает из общего канала jobs (fan-out: несколько воркеров конкурируют за задачи),
// обрабатывает каждую и пишет результат в общий канал results (fan-in).
func worker(id int, jobs <-chan Job, results chan<- Result, wg *sync.WaitGroup) {
	defer wg.Done()
	src := fmt.Sprintf("worker %d", id)
	logger(src, "запущен, ждёт задачи из jobs")
	for job := range jobs {
		logger(src, fmt.Sprintf("взял job %d, обрабатываю...", job.ID))
		time.Sleep(30 * time.Millisecond)
		result := Result{JobID: job.ID, Output: job.ID * 10}
		logger(src, fmt.Sprintf("job %d готово, output=%d, отправляю в results", job.ID, result.Output))
		results <- result
	}
	logger(src, "выход: канал jobs закрыт")
}

func main() {
    var numWorkers int = 1
    if len(os.Args) > 1 {
		iWorkers, _ := strconv.Atoi(os.Args[1])
		if iWorkers > 1 {
			numWorkers = iWorkers
		}
	}
	
	const numJobs = 15    // всего задач от generator

	stopMemTrack := startPeakMemTracker()

	logger("main", fmt.Sprintf("старт pipeline: workers=%d, jobs=%d", numWorkers, numJobs))

	// Буферизованные каналы — generator не блокируется, пока буфер не заполнится.
	jobs := make(chan Job, numJobs)
	results := make(chan Result, numJobs)

	// ── Блок 1: generator (producer) ───────────────────────────────────────
	// Демонстрирует: один goroutine кладёт задачи в jobs и закрывает канал.
	go generator(jobs, numJobs)

	// ── Блок 2: fan-out (pool workers) ─────────────────────────────────────
	// Демонстрирует: N goroutines читают из одного jobs channel.
	// Go runtime распределяет значения между workers — кто готов первым, тот и берёт.
	var wg sync.WaitGroup
	for w := 1; w <= numWorkers; w++ {
		wg.Add(1)
		go worker(w, jobs, results, &wg)
	}

	// ── Блок 3: closer (закрытие results) ──────────────────────────────────
	// Демонстрирует: отдельная goroutine ждёт всех workers и закрывает results,
	// чтобы main мог безопасно читать `for r := range results`.
	go func() {
		logger("closer", "жду завершения всех workers (wg.Wait)...")
		wg.Wait()
		logger("closer", "все workers завершены, закрываю results")
		close(results)
	}()

	// ── Блок 4: fan-in (aggregator) ────────────────────────────────────────
	// Демонстрирует: один consumer собирает результаты из общего results channel.
	// Итог выводится только после того, как closer закроет канал.
	var total int
	var outputs []string
	for r := range results {
		logger("main", fmt.Sprintf("получен результат: job=%d, output=%d", r.JobID, r.Output))
		total += r.Output
		outputs = append(outputs, fmt.Sprintf("%d", r.Output))
	}
	logger("main", fmt.Sprintf("fan-in complete: %d results, outputs=[%s], sum=%d",
		numJobs, strings.Join(outputs, ", "), total))
	logger("main", "pipeline завершён")

	peakHeap, peakSys := stopMemTrack()
	logger("main", fmt.Sprintf("%d workers | пик памяти: heap=%s, sys=%s (всего от OS, включая стеки goroutines)", numWorkers, formatBytes(peakHeap), formatBytes(peakSys)))
}
