# Worker pool

**Worker pool** — фиксированное число goroutines обрабатывает задачи из очереди (channel). Ограничивает параллелизм и ресурсы.

```
        jobs channel
              │
    ┌─────────┼─────────┐
    ▼         ▼         ▼
 Worker1   Worker2   Worker3
    │         │         │
    └─────────┼─────────┘
              ▼
        results channel
```

---

## Базовая реализация

```go
func workerPool(ctx context.Context, jobs <-chan Job, results chan<- Result, n int) {
    var wg sync.WaitGroup
    wg.Add(n)

    for i := 0; i < n; i++ {
        go func(workerID int) {
            defer wg.Done()
            for job := range jobs {
                select {
                case <-ctx.Done():
                    return
                default:
                    results <- process(job)
                }
            }
        }(i)
    }

    wg.Wait()
    close(results)
}
```

---

## Зачем pool

| Без pool | С pool |
|----------|--------|
| 10 000 jobs → 10 000 goroutines | 10 000 jobs → N workers |
| DB connection exhaustion | Bounded concurrency |
| Memory spike | Predictable load |

---

## errgroup.SetLimit как pool

Для independent tasks без shared jobs channel:

```go
g, ctx := errgroup.WithContext(ctx)
g.SetLimit(10)
for _, job := range jobs {
    job := job
    g.Go(func() error {
        return process(ctx, job)
    })
}
return g.Wait()
```

Проще ручного pool — когда нет fan-in результатов через один channel.

---

## Jobs channel lifecycle

```go
jobs := make(chan Job, 100)

// producer
go func() {
    defer close(jobs)
    for _, j := range allJobs {
        select {
        case jobs <- j:
        case <-ctx.Done():
            return
        }
    }
}()

// consumers (workers) — range jobs until closed
```

**Закрывает jobs только producer.** Workers exit on `range jobs` + ctx.

---

## Production tips

1. **Buffer size** jobs channel — сглаживает bursts (не unbounded без причины).
2. **Backpressure** — unbuffered jobs = sync producer/consumer.
3. **Metrics** — queue depth, worker utilization.
4. **ctx** в каждом `process()` — отмена долгих задач.

---

## Пример

[examples/08_worker_pool/](./examples/08_worker_pool/)

```bash
go run ./examples/08_worker_pool/
```

5 workers, 20 jobs, context cancel mid-flight.

Интеграция: [examples/app/internal/pool/](./examples/app/internal/pool/)
