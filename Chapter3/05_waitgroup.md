# sync.WaitGroup

**WaitGroup** — счётчик активных goroutines. **`Wait()`** блокируется, пока счётчик не станет 0.

```go
var wg sync.WaitGroup

for i := 0; i < 10; i++ {
    wg.Add(1)
    go func(id int) {
        defer wg.Done()  // Add(-1)
        work(id)
    }(i)
}

wg.Wait()  // дождаться всех
fmt.Println("all done")
```

---

## API

| Метод | Действие |
|-------|----------|
| `Add(n)` | Увеличить счётчик (вызывать **до** `go`, не из goroutine без care) |
| `Done()` | `Add(-1)` |
| `Wait()` | Блок до counter == 0 |

---

## Правила

1. **`Add` до запуска goroutine** (или документируйте иначе).
2. **`defer wg.Done()`** первой строкой в goroutine — даже при panic (Done всё равно вызовется... actually panic skips defer? defer runs on panic. Good.)
3. **Не переиспользуйте** WaitGroup, пока Wait не вернулся.
4. WaitGroup **не возвращает ошибки** — только «все завершились».

---

## WaitGroup vs errgroup

| WaitGroup | errgroup |
|-----------|----------|
| Ждать N goroutines | Ждать + **первая ошибка** |
| Ошибки — ваш problem | Автоматический cancel context |
| stdlib | `golang.org/x/sync/errgroup` |

Для «fire N workers, collect errors» — **errgroup** (см. [06_errgroup.md](./06_errgroup.md)).

---

## WaitGroup + channel

WaitGroup — «все done». Channel — «результат каждого»:

```go
results := make(chan int, 10)
var wg sync.WaitGroup

for i := 0; i < 10; i++ {
    wg.Add(1)
    go func(i int) {
        defer wg.Done()
        results <- i * 2
    }(i)
}

wg.Wait()
close(results)
for r := range results {
    fmt.Println(r)
}
```

Закрывайте results **после** Wait, не из goroutine без координации.

---

## Пример

[examples/05_waitgroup/](./examples/05_waitgroup/)

```bash
go run ./examples/05_waitgroup/
```
