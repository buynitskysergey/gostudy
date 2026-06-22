# Fan-out / Fan-in

Классический pipeline-паттерн для **parallel map-reduce** over stream of work.

---

## Fan-out

**Один** producer → **несколько** goroutines читают из **одного** channel:

```go
jobs := make(chan Job)
for w := 0; w < numWorkers; w++ {
    go worker(jobs, results)
}
// все workers конкурируют за <-jobs
```

Go runtime распределяет values из channel между workers (whoever ready first).

---

## Fan-in

**Несколько** producers → **один** consumer через **один** results channel:

```go
results := make(chan Result)

// каждый worker пишет в тот же results
go worker(jobs, results)

// merger (optional explicit fan-in)
go func() {
    wg.Wait()
    close(results)
}()

for r := range results {
    aggregate(r)
}
```

---

## Полный pipeline

```
Input → [Generator] → jobs ──fan-out──► Workers (pool)
                                              │
                                         fan-in
                                              ▼
                                        results → [Aggregator] → Output
```

Stages соединяются **channels** — каждый stage — group of goroutines.

---

## Bounded parallelism

Fan-out **не** значит unlimited goroutines:

```go
const numWorkers = 5
for i := 0; i < numWorkers; i++ {
    go worker(ctx, jobs, results)
}
```

Fan-out = N workers на один jobs channel, не «одна goroutine на job».

---

## Context через весь pipeline

```go
ctx, cancel := context.WithCancel(context.Background())
defer cancel()

jobs := make(chan Job)
results := make(chan Result)

go generate(ctx, jobs)
for i := 0; i < n; i++ {
    go worker(ctx, jobs, results)
}
go func() {
    wg.Wait()
    close(results)
}()

for {
    select {
    case r, ok := <-results:
        if !ok { return }
        handle(r)
    case <-ctx.Done():
        return
    }
}
```

---

## Fan-out/fan-in vs PHP/JS

| | JS | Go |
|---|-----|-----|
| Parallel map | `Promise.all(items.map(fn))` | fan-out workers + jobs ch |
| Stream | async generators | goroutine + range channel |
| Merge | manual | fan-in to one channel |

---

## Когда использовать

✅ Batch processing (files, records, URLs)  
✅ ETL pipelines  
✅ Image/thumbnail generation  
❌ Простой CRUD — overkill  
❌ Strict ordering required — нужна другая architecture  

---

## Пример

[examples/09_fan_out_fan_in/](./examples/09_fan_out_fan_in/)

```bash
go run ./examples/09_fan_out_fan_in/
```

Generator → 3 workers → merge results.

**Полная интegrация:** [examples/app/cmd/processor/](./examples/app/cmd/processor/) — fan-out + pool + errgroup + context.
