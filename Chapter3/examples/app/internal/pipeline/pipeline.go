package pipeline

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Task — единица работы в pipeline: у каждой задачи есть идентификатор.
type Task struct {
	ID int
}

// Result — результат обработки задачи воркером.
type Result struct {
	TaskID int
	Value  int
}

// Process запускает fan-out/fan-in pipeline с пулом воркеров.
// Generator кладёт tasks в jobs; N workers читают из jobs (fan-out)
// и пишут в общий results (fan-in); closer закрывает results после wg.Wait.
// Останавливается при отмене ctx (context cancellation).
func Process(ctx context.Context, tasks []Task, numWorkers int) ([]Result, error) {
	Log("pipeline", fmt.Sprintf("старт: workers=%d, tasks=%d", numWorkers, len(tasks)))

	jobs := make(chan Task)
	results := make(chan Result)

	// ── Блок 1: fan-out (pool workers) ─────────────────────────────────────
	// N goroutines конкурируют за задачи из общего канала jobs.
	var wg sync.WaitGroup
	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			src := fmt.Sprintf("worker %d", workerID+1)
			Log(src, "запущен, ждёт задачи из jobs")
			for task := range jobs {
				select {
				case <-ctx.Done():
					Log(src, fmt.Sprintf("отмена ctx при job %d, выхожу", task.ID))
					return
				default:
					Log(src, fmt.Sprintf("взял task %d, обрабатываю...", task.ID))
					time.Sleep(200 * time.Millisecond)
					result := Result{TaskID: task.ID, Value: task.ID * task.ID}
					select {
					case results <- result:
						Log(src, fmt.Sprintf("task %d готово, value=%d, отправлено в results", task.ID, result.Value))
					case <-ctx.Done():
						Log(src, fmt.Sprintf("отмена ctx при отправке task %d, выхожу", task.ID))
						return
					}
				}
			}
			Log(src, "выход: канал jobs закрыт")
		}(w)
	}

	// ── Блок 2: generator (producer) ─────────────────────────────────────────
	// Кладёт tasks в jobs и закрывает канал — workers выйдут из range jobs.
	go func() {
		defer func() {
			Log("generator", "все задачи отправлены, закрываю jobs")
			close(jobs)
		}()
		Log("generator", fmt.Sprintf("старт: отправлю %d задач", len(tasks)))
		for _, t := range tasks {
			select {
			case jobs <- t:
				Log("generator", fmt.Sprintf("отправляет task %d", t.ID))
			case <-ctx.Done():
				Log("generator", "отмена ctx, прекращаю отправку")
				return
			}
		}
	}()

	// ── Блок 3: closer (закрытие results) ──────────────────────────────────
	// Ждёт всех workers и закрывает results, чтобы fan-in мог читать до конца.
	go func() {
		Log("closer", "жду завершения всех workers (wg.Wait)...")
		wg.Wait()
		Log("closer", "все workers завершены, закрываю results")
		close(results)
	}()

	// ── Блок 4: fan-in (aggregator) ──────────────────────────────────────────
	// Собирает результаты из results; выходит при закрытии канала или отмене ctx.
	var out []Result
	for {
		select {
		case r, ok := <-results:
			if !ok {
				if err := ctx.Err(); err != nil {
					Log("fan-in", fmt.Sprintf("results закрыт, ctx err: %v, собрано %d", err, len(out)))
					return out, fmt.Errorf("pipeline: %w", err)
				}
				Log("fan-in", fmt.Sprintf("complete: %d results", len(out)))
				return out, nil
			}
			Log("fan-in", fmt.Sprintf("получен результат: task=%d, value=%d", r.TaskID, r.Value))
			out = append(out, r)
		case <-ctx.Done():
			Log("fan-in", fmt.Sprintf("отмена ctx: %v, собрано %d", ctx.Err(), len(out)))
			return out, fmt.Errorf("pipeline: %w", ctx.Err())
		}
	}
}
