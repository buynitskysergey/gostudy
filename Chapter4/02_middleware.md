# Middleware

В Go middleware — это функция, которая принимает `http.Handler` и возвращает новый `http.Handler`:

```go
type Middleware func(http.Handler) http.Handler
```

Нет глобального «kernel» как в Laravel — вы **явно** оборачиваете handler в `main` или router.

---

## Паттерн обёртки

```go
func Logging(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        next.ServeHTTP(w, r)
        log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start))
    })
}
```

Порядок важен: внешний middleware видит запрос первым и ответ последним.

```
Request  → Logging → Recover → Auth → handler
Response ← Logging ← Recover ← Auth ← handler
```

---

## Цепочка (chain)

```go
func Chain(h http.Handler, mws ...Middleware) http.Handler {
    for i := len(mws) - 1; i >= 0; i-- {
        h = mws[i](h)
    }
    return h
}

handler := Chain(mux, Logging, Recover, RequestID)
```

Как Express `app.use()`, но композиция чистых функций без скрытого состояния.

---

## Типовой набор

| Middleware | Задача |
|------------|--------|
| RequestID | `X-Request-Id` → context |
| Logging | method, path, status, duration |
| Recover | `recover()` → 500, не уронить процесс |
| CORS | только если нужен browser API |
| Auth | проверка токена / API key |

---

## ResponseWriter wrapper

Чтобы залогировать **status code**, оберните `ResponseWriter`:

```go
type statusWriter struct {
    http.ResponseWriter
    status int
}

func (w *statusWriter) WriteHeader(code int) {
    w.status = code
    w.ResponseWriter.WriteHeader(code)
}
```

Без этого `http.ResponseWriter` не отдаёт записанный статус наружу.

---

## Context values

```go
type ctxKey int
const keyRequestID ctxKey = 1

func RequestID(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        id := r.Header.Get("X-Request-Id")
        if id == "" {
            id = uuidOrRandom()
        }
        ctx := context.WithValue(r.Context(), keyRequestID, id)
        w.Header().Set("X-Request-Id", id)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

**Правило:** в context кладите request-scoped данные (id, user). Не кладите DB/logger «навсегда» — передавайте зависимостями (глава 2).

---

## Middleware vs PHP/JS

| Laravel / Express | Go |
|-------------------|-----|
| `$middleware` stack | `Chain(h, ...)` |
| `$request->attributes` | `context.WithValue` |
| exception handler | Recover middleware |
| `next()` | `next.ServeHTTP(w, r)` |

---

## Anti-patterns

```go
// ❌ Глобальные переменные вместо context / DI
var currentUser *User

// ❌ Middleware, который читает и исчерпывает Body до handler
// (если нужно — tee / LimitReader осознанно)

// ❌ Recover, который глотает ошибку без лога
```

---

## Пример

[examples/02_middleware/](./examples/02_middleware/)

```bash
go run ./examples/02_middleware/
```

Logging + Recover + RequestID в одной цепочке.
