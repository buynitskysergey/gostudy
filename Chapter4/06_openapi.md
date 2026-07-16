# OpenAPI

**OpenAPI** (ex-Swagger) — машиночитаемый контракт HTTP API. В Go-мире чаще **contract-first** или «spec рядом с кодом», а не магия декораторов как в Nest/Spring.

---

## Зачем контракт

| Без OpenAPI | С OpenAPI |
|-------------|-----------|
| Документация в README устаревает | Один YAML = docs + codegen клиентов |
| Фронт гадает по полям | `openapi-generator` / `oapi-codegen` |
| Ручной Postman | Импорт spec |

Для stdlib-бэкенда: **пишите `openapi.yaml` руками** (или генерируйте из тестов) и отдавайте его с сервера.

---

## Минимальный spec (OpenAPI 3)

```yaml
openapi: 3.0.3
info:
  title: Tasks API
  version: 1.0.0
paths:
  /api/v1/tasks:
    post:
      summary: Create task
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/CreateTask'
      responses:
        '201':
          description: Created
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Task'
        '400':
          description: Validation error
components:
  schemas:
    CreateTask:
      type: object
      required: [title]
      properties:
        title:
          type: string
          minLength: 1
          maxLength: 200
    Task:
      type: object
      properties:
        id: { type: string }
        title: { type: string }
        done: { type: boolean }
```

---

## Отдача spec из сервиса

```go
mux.HandleFunc("GET /openapi.yaml", func(w http.ResponseWriter, r *http.Request) {
    http.ServeFile(w, r, "openapi.yaml")
})
```

Или `embed`:

```go
//go:embed openapi.yaml
var openAPI []byte

mux.HandleFunc("GET /openapi.yaml", func(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/yaml")
    _, _ = w.Write(openAPI)
})
```

Swagger UI — HTML + CDN, url на `/openapi.yaml` (в примере: `GET /docs`).

---

## Contract-first vs code-first

| Подход | Суть |
|--------|------|
| **Contract-first** | YAML → `oapi-codegen` → interfaces/types → вы реализуете |
| **Code-first** | код → генерация spec (swaggo и т.п.) |
| **Spec рядом** | пишете handlers + YAML вручную, держите в sync |

Для обучения: **spec рядом** — видно оба артефакта. В команде часто contract-first.

---

## OpenAPI vs PHP/JS

| Nest / tsoa / Laravel Scramble | Go |
|--------------------------------|-----|
| Декораторы → Swagger | YAML/JSON файл |
| Автоиз типов TS | Явные schemas |
| UI из коробки | Serve file + внешний UI |

Go не прячет контракт в reflection — вы **владеете** документом.

---

## Anti-patterns

```
// ❌ Spec только в Confluence, не в репозитории
// ❌ Генерировать клиентов с устаревшего YAML
// ❌ Расхождение: код принимает field X, schema его нет
```

Дисциплина: PR меняет **и** handler, **и** `openapi.yaml`.

---

## Пример

[examples/06_openapi/](./examples/06_openapi/)

```bash
go run ./examples/06_openapi/
```

Сервер отдаёт embedded `openapi.yaml`, Swagger UI на `/docs` и реализует один endpoint по контракту.
