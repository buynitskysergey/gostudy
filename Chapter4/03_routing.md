# Routing

До Go 1.22 `ServeMux` умел только prefix-match. С **Go 1.22+** встроенный mux закрывает 80% нужд REST API: method, path patterns, path values.

---

## Patterns (Go 1.22+)

```go
mux := http.NewServeMux()

mux.HandleFunc("GET /healthz", health)
mux.HandleFunc("GET /api/v1/tasks", listTasks)
mux.HandleFunc("POST /api/v1/tasks", createTask)
mux.HandleFunc("GET /api/v1/tasks/{id}", getTask)
mux.HandleFunc("DELETE /api/v1/tasks/{id}", deleteTask)
```

| Pattern | Смысл |
|---------|--------|
| `GET /path` | только GET |
| `/path` | любой метод |
| `/tasks/{id}` | path parameter |
| `/files/{path...}` | wildcard (остаток пути) |

Конфликт более специфичного и общего pattern разрешает сам mux.

---

## Path values

```go
func getTask(w http.ResponseWriter, r *http.Request) {
    id := r.PathValue("id")
    // ...
}
```

Как `$request->route('id')` / `req.params.id`.

---

## Method not allowed

Если зарегистрирован `GET /x`, а пришёл `POST /x`, mux вернёт **405** (и `Allow` header) — не нужно руками.

404 — когда ни один pattern не совпал.

---

## Группы и префиксы

Stdlib не имеет `Group("/api/v1")`. Варианты:

1. **Явные полные paths** — прозрачно, хорошо для учебных/средних API.
2. **Под-mux + `StripPrefix`** — для монтажа:

```go
api := http.NewServeMux()
api.HandleFunc("GET /tasks", listTasks)

mux.Handle("/api/v1/", http.StripPrefix("/api/v1", api))
```

С Go 1.22 чаще проще писать полные patterns на одном mux.

---

## Когда внешнего router мало

| Нужно | Stdlib | chi / gorilla / echo |
|-------|--------|----------------------|
| Method + `{id}` | ✅ | ✅ |
| Middleware per-group | вручную | удобнее |
| Regex constraints | ❌ | часто есть |
| Huge route tree | ок | чуть эргономичнее |

Для курса и большинства сервисов: **начните с `ServeMux`**. Добавляйте chi, когда groups/middleware-per-route реально болят.

---

## Routing vs PHP/JS

| Laravel / Express | Go 1.22 ServeMux |
|-------------------|------------------|
| `Route::get('/tasks/{id}', ...)` | `mux.HandleFunc("GET /tasks/{id}", ...)` |
| route groups | полные paths или sub-mux |
| middleware groups | `Chain` на весь mux или на отдельный handler |

---

## Anti-patterns

```go
// ❌ Ручной switch по r.URL.Path в одном handler на всё
// ❌ Игнорировать method — принимать POST на GET-only resource
// ❌ Парсить id через strings.TrimPrefix вместо PathValue
```

---

## Пример

[examples/03_routing/](./examples/03_routing/)

```bash
go run ./examples/03_routing/
```

CRUD-like routes с `{id}` и демонстрацией 404/405.
