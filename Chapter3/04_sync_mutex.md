# sync.Mutex и sync.RWMutex

Когда несколько goroutines **разделяют mutable state**, channels не всегда удобны. **`sync.Mutex`** — mutual exclusion lock.

```go
var mu sync.Mutex
var counter int

mu.Lock()
counter++
mu.Unlock()
```

Идиома: **`defer mu.Unlock()`** сразу после Lock.

---

## sync.Mutex

- **Exclusive** lock — только одна goroutine в critical section.
- Zero value usable — `var mu sync.Mutex`, не нужен `make`.
- **Не reentrant** — повторный Lock в той же goroutine → deadlock.

```go
mu.Lock()
defer mu.Unlock()
// critical section
```

---

## sync.RWMutex

**Readers-writer** lock:

```go
var mu sync.RWMutex

// Много concurrent readers
mu.RLock()
_ = cache[key]
mu.RUnlock()

// Один writer — exclusive
mu.Lock()
cache[key] = value
mu.Unlock()
```

| Операция | Mutex | RWMutex |
|----------|-------|---------|
| Read-heavy shared cache | Lock на каждый read | `RLock` — параллельные reads |
| Write | Lock | Lock (blocks all R/W) |

Используйте RWMutex когда **reads >> writes** (кеш, config snapshot).

---

## Mutex vs Channels

| Mutex | Channel |
|-------|---------|
| Protect shared struct/map | Передать ownership данных |
| In-memory cache, counter | Pipeline, work distribution |
| «Share memory» | «Share by communicating» |

**Правило:** сначала подумайте channel; mutex — когда state genuinely shared и короткие critical sections.

---

## sync.Map

Для **concurrent map** — `sync.Map` или mutex + обычный map. Обычный `map` без sync — **data race**.

```go
// ❌ DATA RACE
m := map[string]int{}
go func() { m["a"] = 1 }()
go func() { _ = m["a"] }()
```

---

## Race detector

```bash
go run -race ./examples/04_mutex/
```

Всегда включайте `-race` в CI для concurrent кода.

---

## Связь с Chapter 2

В [Chapter2/examples/app/internal/storage/memory/repository.go](../Chapter2/examples/app/internal/storage/memory/repository.go) уже используется `RWMutex` для защиты `map` — типичный production case.

---

## Anti-patterns

```go
// ❌ Lock без Unlock (panic path без defer)
mu.Lock()
if err != nil { return err } // deadlock для других
mu.Unlock()

// ❌ Держать lock во время I/O
mu.Lock()
resp, _ := http.Get(url) // блокирует всех readers
mu.Unlock()

// ❌ RUnlock после Lock (как в баге chapter 2!)
mu.RLock()
defer mu.Unlock() // panic
```

---

## Пример

[examples/04_mutex/](./examples/04_mutex/)

```bash
go run ./examples/04_mutex/
go run -race ./examples/04_mutex/
```

Counter с Mutex, cache с RWMutex, демонстрация race без lock.
