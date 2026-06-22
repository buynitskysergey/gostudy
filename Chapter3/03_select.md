# Select

**select** — switch для channel operations. Блокируется, пока **хотя бы один** case не готов.

```go
select {
case v := <-ch1:
    fmt.Println("from ch1", v)
case ch2 <- 42:
    fmt.Println("sent to ch2")
case <-time.After(time.Second):
    fmt.Println("timeout")
default:
    fmt.Println("no channel ready") // non-blocking, если default есть
}
```

---

## Семантика

- Если **несколько** case готовы — выбирается **случайный** (fairness).
- Если **ни один** не готов — блокировка (без `default`).
- `default` — non-blocking select.

---

## Типичные паттерны

### Timeout

```go
select {
case result := <-resultCh:
    return result, nil
case <-time.After(5 * time.Second):
    return nil, ErrTimeout
}
```

Production: `context.WithTimeout` вместо `time.After` (избегает leak при частых вызовах).

### Cancellation через context

```go
select {
case work := <-jobs:
    process(work)
case <-ctx.Done():
    return ctx.Err()
}
```

### Disable case через nil channel

```go
ch1, ch2 := make(chan int), make(chan int)
for ch1 != nil || ch2 != nil {
    select {
    case v, ok := <-ch1:
        if !ok { ch1 = nil; continue }
        handle(v)
    case v, ok := <-ch2:
        if !ok { ch2 = nil; continue }
        handle(v)
    }
}
```

Assign `nil` to channel — case навсегда блокируется и **исключается** из select.

---

## Select + for = event loop goroutine

```go
for {
    select {
    case msg := <-msgs:
        handle(msg)
    case <-ctx.Done():
        return
    }
}
```

Классический паттерн worker / server loop.

---

## Select vs PHP/JS

| Go select | JS |
|-----------|-----|
| Multiplex channels | `Promise.race([p1, p2])` |
| Blocking | async await one winner |
| + context | AbortSignal |

---

## Anti-patterns

```go
// ❌ Busy loop с default
for {
    select {
    default:
        // 100% CPU burn
    }
}

// ❌ time.After в цикле — timer leak до Go 1.23 fixes; используйте context
for {
    select {
    case <-time.After(time.Second): // плохо в tight loop
    }
}
```

---

## Пример

[examples/03_select/](./examples/03_select/)

```bash
go run ./examples/03_select/
```

Timeout, context cancel, nil channel disable, multi-case.
