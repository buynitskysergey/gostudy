# errgroup

**errgroup** — `golang.org/x/sync/errgroup`: WaitGroup + **первая ошибка** + **cancel context** для всей группы.

```go
import "golang.org/x/sync/errgroup"

g, ctx := errgroup.WithContext(parentCtx)

g.Go(func() error {
    return callServiceA(ctx)
})

g.Go(func() error {
    return callServiceB(ctx)
})

if err := g.Wait(); err != nil {
    // первая ошибка; ctx отменён
}
```

---

## Зачем errgroup

| Проблема | WaitGroup | errgroup |
|----------|-----------|----------|
| Goroutine вернула error | Некуда положить | `return err` из `Go()` |
| Одна упала — остальные работают | Да (leak work) | `ctx` cancelled |
| Собрать все ошибки | Вручную | `Wait()` — только первая* |

*Для всех ошибок — `multierr` или свой collector.

---

## WithContext

```go
g, ctx := errgroup.WithContext(context.Background())
```

При **первой non-nil error** из любой `Go()`:
1. `Wait()` вернёт эту ошибку
2. `ctx` получит cancellation → другие goroutines должны проверять `ctx.Done()`

---

## Limit concurrency

```go
g, ctx := errgroup.WithContext(ctx)
g.SetLimit(5)  // max 5 goroutines одновременно (Go 1.20+)
```

Встроенный **worker pool limit** для errgroup — часто достаточно вместо ручного pool.

---

## Паттерн: parallel independent I/O

```go
g, ctx := errgroup.WithContext(ctx)

g.Go(func() error {
    user, err := userRepo.Get(ctx, id)
    if err != nil { return err }
    // store in closure variable or channel
    return nil
})

g.Go(func() error {
    orders, err := orderRepo.List(ctx, id)
    ...
})

if err := g.Wait(); err != nil {
    return fmt.Errorf("profile: %w", err)
}
```

Как `Promise.all`, но с cancel при первом reject.

---

## errgroup vs PHP/JS

| JS | errgroup |
|----|----------|
| `Promise.all([a, b])` | `g.Go` × N + `Wait` |
| `Promise.allSettled` | нет встроенного — свой код |
| AbortSignal | `ctx` from `WithContext` |

---

## Anti-patterns

```go
// ❌ Игнорировать ctx в goroutines — не остановятся при ошибке соседа
g.Go(func() error {
    return longWork(context.Background())
})

// ❌ Panic в Go() — не попадёт в Wait как error (crash)
// используйте recover или return error
```

---

## Пример

[examples/06_errgroup/](./examples/06_errgroup/)

```bash
go run ./examples/06_errgroup/
```

Parallel fetch с cancel при первой ошибке + `SetLimit`.
