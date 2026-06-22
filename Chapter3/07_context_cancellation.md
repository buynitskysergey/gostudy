# Context cancellation (concurrency)

В [Chapter 2](../Chapter2/07_context.md) context — первый аргумент I/O. В **concurrency** context — механизм **координированной отмены** дерева goroutines.

---

## Cancellation tree

```go
ctx, cancel := context.WithCancel(context.Background())
defer cancel()

go worker(ctx, 1)
go worker(ctx, 2)

// cancel() → все worker'ы получают ctx.Done()
```

`cancel()` **идempotent** — можно вызывать несколько раз.

---

## Goroutine должна уважать context

```go
func worker(ctx context.Context, id int) error {
    for {
        select {
        case <-ctx.Done():
            return ctx.Err()  // context.Canceled
        default:
            if err := doUnit(ctx); err != nil {
                return err
            }
        }
    }
}
```

Без `select` + `ctx.Done()` goroutine **не остановится** при cancel.

---

## Связь с errgroup

```go
g, ctx := errgroup.WithContext(parent)
g.Go(func() error { return worker(ctx) })
g.Go(func() error { return worker(ctx) })
_ = g.Wait() // при error одной — ctx cancelled, вторая должна выйти
```

---

## Timeout в concurrent pipeline

```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

results, err := runPipeline(ctx, jobs)
if errors.Is(err, context.DeadlineExceeded) {
    // partial results? cleanup goroutines
}
```

Pipeline **обязан** пробрасывать `ctx` в каждый stage и worker.

---

## Graceful shutdown (preview этапа 4)

HTTP server shutdown:

```go
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()
srv.Shutdown(ctx)  // ждёт in-flight requests
```

Тот же context cancellation, другой domain.

---

## Anti-patterns

```go
// ❌ Detached context — теряется cancel от request
go func() {
    process(context.Background())
}()

// ❌ Goroutine leak — нет exit on ctx.Done()
go func() {
    for { work() }
}()

// ❌ Забыли defer cancel() — timer leak
ctx, _ := context.WithTimeout(parent, time.Second)
```

---

## Пример

[examples/07_context/](./examples/07_context/)

```bash
go run ./examples/07_context/
```

Workers, parent cancel, timeout, интеграция с select.

Полный pipeline: [examples/app/](./examples/app/)
