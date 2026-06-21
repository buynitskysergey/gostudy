# context.Context

`context.Context` — сквозной **контекст выполнения**: отмена, таймауты, request-scoped values. В идиоматичном Go — **первый параметр** функций, которые делают I/O или могут быть отменены.

```go
func (s *Service) Get(ctx context.Context, id string) (Order, error)
func (r *Repository) FindByID(ctx context.Context, id string) (Order, error)
```

Не `ctx` последним, не в struct (кроме HTTP handler, где request.Context()).

---

## Зачем context

| Возможность | Пример |
|-------------|--------|
| **Cancellation** | клиент закрыл соединение → отменить запрос к БД |
| **Timeout** | «этот запрос не дольше 2s» |
| **Deadline** | «закончить до 15:00:00» |
| **Values** | request ID, auth (осторожно!) |

---

## Создание context

```go
// Корень — только в main, tests, request handler
ctx := context.Background()
ctx := context.TODO()  // placeholder при рефакторинге

// Производные
ctx, cancel := context.WithTimeout(parent, 2*time.Second)
defer cancel()  // обязательно — освобождает ресурсы timer

ctx, cancel := context.WithCancel(parent)
ctx, cancel := context.WithDeadline(parent, time.Now().Add(time.Hour))
```

**Всегда** вызывайте `cancel()` в defer — идиома Go.

---

## Проверка отмены

```go
select {
case <-ctx.Done():
    return zero, ctx.Err()  // context.Canceled или context.DeadlineExceeded
default:
    // продолжаем работу
}
```

В long-running loops — проверяйте `ctx.Done()` на каждой итерации.

Для DB/http клиентов передавайте `ctx` в API библиотеки — они сами отменят запрос.

---

## Проброс по слоям

```
HTTP Request (r.Context())
    → Handler.GetOrder(ctx, id)
    → Service.Get(ctx, id)
    → Repository.FindByID(ctx, id)
    → sql.QueryContext(ctx, ...)
```

**Не храните** context в struct Service:

```go
// ❌
type Service struct {
    ctx context.Context
}

// ✅ ctx — параметр каждого метода
func (s *Service) Get(ctx context.Context, id string) (Order, error)
```

---

## Values — использовать умеренно

```go
ctx = context.WithValue(ctx, requestIDKey, "abc-123")
```

- Только request-scoped metadata (request ID, trace ID)
- Ключи — unexported typed constants (не string)
- **Не** передавайте бизнес-объекты через context — это anti-pattern

---

## context vs PHP/JS

| Go | PHP | Node |
|----|-----|------|
| `context.Context` | нет прямого аналога | `AbortController` (отмена) |
| Первый аргумент | — | часто options object |
| Propagation явная | — | middleware chain |

---

## Anti-patterns

```go
// ❌ context.Background() внутри service — теряется timeout клиента
func (s *Service) Get(id string) (Order, error) {
    ctx := context.Background()
}

// ❌ nil context
repo.FindByID(nil, id)

// ❌ WithValue для DI
ctx = context.WithValue(ctx, "db", db)  // используйте constructor injection
```

---

## Пример

[examples/07_context/](./examples/07_context/)

```bash
go run ./examples/07_context/
```

В `app/`: timeout и cancellation в [examples/app/cmd/api/main.go](./examples/app/cmd/api/main.go) и repo проверяет `ctx.Done()`.
