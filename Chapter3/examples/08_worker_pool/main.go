package main

import (
	"context"
	"fmt"
	"math/rand/v2"
	"strings"
	"sync"
	"time"
)

var (
	programStart = time.Now()
	logMu        sync.Mutex
	minDelay     = 5 * time.Millisecond
	maxDelay     = 20 * time.Millisecond
)

// logger выводит сообщение с меткой времени (ms от старта) и источником.
func logger(source, message string) {
	logMu.Lock()
	defer logMu.Unlock()
	elapsed := time.Since(programStart).Milliseconds()
	fmt.Printf("%dms | [%s]: %s\n", elapsed, source, message)
}

// Job — задача для worker pool: у каждой задачи есть идентификатор.
type Job struct {
	ID int
}

// process имитирует «тяжёлую» работу (50 ms) и возвращает ID².
// Если контекст отменён — возвращает -1 (сигнал «результат не нужен»).
func process(ctx context.Context, job Job) int {
	logger("process", fmt.Sprintf("job %d: начало обработки", job.ID))
	delay := minDelay + rand.N(maxDelay-minDelay+time.Millisecond)
	select {
	case <-time.After(delay):
		result := job.ID * job.ID
		logger("process", fmt.Sprintf("job %d: готово через %dms, результат = %d", job.ID, delay.Milliseconds(), result))
		return result
	case <-ctx.Done():
		logger("process", fmt.Sprintf("job %d: отменено (ctx.Done)", job.ID))
		return -1
	}
}

func main() {
	const workers = 3   // фиксированное число goroutine-воркеров
	const jobCount = 12 // всего задач в очереди

	logger("main", fmt.Sprintf("старт: workers=%d, jobs=%d", workers, jobCount))

	// Контекст с отменой: producer вызовет cancel() после 8-й задачи
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Буферизованные каналы — producer не блокируется, пока буфер не заполнится.
	jobs := make(chan Job, jobCount)
	results := make(chan int, jobCount)

	var wg sync.WaitGroup

	// ── Блок 1: запуск worker pool ──────────────────────────────────────────
	// Демонстрирует: фиксированное число goroutines читает из jobs до close(jobs).
	// Полученную задачу всегда передаём в process — отмена обрабатывается там,
	// а не отбрасывает job, уже извлечённый из канала.
	for w := 1; w <= workers; w++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			src := fmt.Sprintf("worker %d", id)
			logger(src, "запущен, ждёт задачи")
			for job := range jobs {
				logger(src, fmt.Sprintf("взял job %d", job.ID))
				if r := process(ctx, job); r >= 0 {
					logger(src, fmt.Sprintf("отправил результат %d (job %d)", r, job.ID))
					results <- r
				} else {
					logger(src, fmt.Sprintf("пропустил job %d (отменён в process)", job.ID))
				}
			}
			logger(src, "выход: канал jobs закрыт")
		}(w)
	}

	// ── Блок 2: producer (отправитель задач) ───────────────────────────────
	// Демонстрирует: один goroutine кладёт задачи в jobs и закрывает канал.
	// После 8-й задачи — отложенный cancel(): часть задач успеет завершиться,
	// остальные прервутся внутри process() (mid-flight cancellation).
	go func() {
		logger("producer", "старт отправки задач")
		for i := 1; i <= jobCount; i++ {
			logger("producer", fmt.Sprintf("формирует job %d", i))
			time.Sleep(5 * time.Millisecond)
			jobs <- Job{ID: i}
			logger("producer", fmt.Sprintf("отправил job %d", i))
			if i == 8 {
				logger("producer", "cancelling after job 8 submitted (demo mid-flight cancellation)")
				cancel()
				break
			}
		}
		logger("producer", "все задачи отправлены, закрываю jobs")
		close(jobs)
	}()

	// ── Блок 3: closer (закрытие results) ──────────────────────────────────
	// Демонстрирует: отдельная goroutine ждёт всех воркеров и закрывает results,
	// чтобы main мог безопасно читать `for r := range results`.
	go func() {
		logger("closer", "жду завершения всех workers (wg.Wait)...")
		wg.Wait()
		logger("closer", "все workers завершены, закрываю results")
		close(results)
	}()

	// ── Блок 4: consumer (сбор результатов) ──────────────────────────────────
	// Демонстрирует: fan-in — main блокируется на range, пока closer не закроет канал.
	// Итоговая строка выводится только после завершения всех workers.
	var collected []int
	for r := range results {
		logger("main", fmt.Sprintf("получен результат: %d", r))
		collected = append(collected, r)
	}
	parts := make([]string, len(collected))
	for i, r := range collected {
		parts[i] = fmt.Sprintf("%d", r)
	}
	logger("main", "results: "+strings.Join(parts, " "))
	logger("main", "все workers завершены, программа завершена")
}
