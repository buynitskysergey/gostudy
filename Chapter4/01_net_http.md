# net/http

Пакет `net/http` — фундамент любого Go HTTP-сервиса. Один интерфейс правит всем:

```go
type Handler interface {
    ServeHTTP(ResponseWriter, *Request)
}
```

Всё остальное (middleware, router, JSON helpers) — композиция вокруг `Handler`.

---

## Handler и HandlerFunc

```go
mux := http.NewServeMux()
mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusOK)
    _, _ = w.Write([]byte("ok"))
})
```

`http.HandlerFunc` — адаптер: обычная функция становится `Handler`.

---

## http.Server

Не используйте голый `http.ListenAndServe` в production — нет таймаутов и graceful shutdown:

```go
srv := &http.Server{
    Addr:              ":8080",
    Handler:           mux,
    ReadHeaderTimeout: 5 * time.Second,
    ReadTimeout:       10 * time.Second,
    WriteTimeout:      10 * time.Second,
    IdleTimeout:       60 * time.Second,
}
```

| Поле | Зачем |
|------|-------|
| `ReadHeaderTimeout` | защита от Slowloris |
| `ReadTimeout` / `WriteTimeout` | лимиты на весь запрос/ответ |
| `IdleTimeout` | keep-alive соединения |

---

## Graceful shutdown

```go
go func() {
    if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
        log.Fatal(err)
    }
}()

stop := make(chan os.Signal, 1)
signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
<-stop

ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()
_ = srv.Shutdown(ctx) // дождаться активных запросов, не принимать новые
```

Как PHP-FPM «drain» или Nest `enableShutdownHooks` — но явно и в `main`.

---

## Request / ResponseWriter

| | PHP | Go |
|---|-----|-----|
| Тело | `$request->getContent()` | `r.Body` (`io.ReadCloser`) |
| Query | `$_GET` | `r.URL.Query()` |
| Headers | `$request->headers` | `r.Header.Get("X-Request-Id")` |
| Status | `http_response_code(404)` | `w.WriteHeader(404)` |
| Context | редко | `r.Context()` — отмена при disconnect |

**Важно:** `WriteHeader` вызывайте **один раз**, до первой записи в body. После `Write` статус уже 200.

---

## Context запроса

```go
ctx := r.Context()
// клиент закрыл соединение → ctx cancelled
select {
case <-ctx.Done():
    return
case result := <-work:
    ...
}
```

Связка с главой 3: каждый HTTP-запрос — дерево работы с общим `context`.

---

## Anti-patterns

```go
// ❌ ListenAndServe без таймаутов
http.ListenAndServe(":8080", mux)

// ❌ Игнорировать ошибку Write / Encode
json.NewEncoder(w).Encode(v)

// ❌ Паника в handler без recovery middleware
```

---

## Пример

[examples/01_net_http/](./examples/01_net_http/)

```bash
go run ./examples/01_net_http/
```

Минимальный сервер с таймаутами и graceful shutdown по Ctrl+C.
