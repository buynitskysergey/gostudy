# Goroutines

**Goroutine** — легковесный поток выполнения, управляемый Go runtime. Запуск:

```go
go doWork()
```

Функция выполняется **конкурентно** (concurrently) с вызывающим кодом.

---

## Goroutine vs OS thread

| | OS thread | Goroutine |
|---|-----------|-----------|
| Создание | ~1–2 MB stack, дорого | ~2–8 KB stack, дёшево |
| Количество | сотни | миллионы (теоретически) |
| Планировщик | OS kernel | Go runtime (M:N) |

Go runtime распределяет goroutines по небольшому пулу OS threads.

---

## Синтаксис

```go
go func() {
    fmt.Println("async work")
}()

go processItem(item)  // с аргументами
```

Goroutine **не возвращает** результат напрямую — используйте channels, shared state (с sync) или callback.

---

## Главная ловушка: main завершается — goroutines умирают

```go
func main() {
    go fmt.Println("hello")
    // main exit → программа завершилась, goroutine не успела
}
```

**Решения:**
- `sync.WaitGroup` — дождаться завершения
- channel — получить сигнал
- `time.Sleep` — только для демо, **не в production**

---

## Closure и loop variable (Go 1.22+)

```go
for i := 0; i < 3; i++ {
    go func() {
        fmt.Println(i) // Go 1.22+: каждая goroutine видит своё i
    }()
}
```

До Go 1.22 нужно было `go func(i int) { ... }(i)`.

---

## Goroutine vs PHP/JS

| | JS (Node) | Go |
|---|-----------|-----|
| Модель | Single-threaded event loop | M:N goroutines |
| Async | `async/await` | `go` + channels |
| Параллелизм CPU | Worker threads (ограничено) | `GOMAXPROCS` goroutines на CPU cores |

Go goroutine — **не** Promise: нет `.then()`, нет встроенного cancellation без context/channels.

---

## Когда запускать goroutine

✅ I/O-bound work (HTTP, DB, файлы)  
✅ Независимые задачи с координацией через channels/context  
❌ CPU-bound без `GOMAXPROCS` awareness — может быть overhead  
❌ «На всякий случай async» — каждая goroutine должна иметь **exit condition**

---

## Пример

[examples/01_goroutines/](./examples/01_goroutines/)

```bash
go run ./examples/01_goroutines/
```

Демонстрирует: запуск, WaitGroup для ожидания, проблему раннего exit без координации.
