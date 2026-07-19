# Глава 4 — Backend на stdlib

HTTP-сервис **без фреймворков**: `net/http`, свой middleware, роутинг Go 1.22+, JSON, валидация, OpenAPI и конфигурация. Всё, что нужно для production API, уже есть в стандартной библиотеке (и паре YAML/JSON-файлов для контракта).

## Философия HTTP в Go

> **The standard library is the framework.**

| PHP/JS привычка | Go (stdlib) |
|-----------------|-------------|
| Laravel / Express / Nest | `net/http` + тонкий wiring в `main` |
| global middleware stack | `func(http.Handler) http.Handler` |
| Router package обязателен | `ServeMux` (Go 1.22+: methods + `{id}`) |
| class-validator / Joi | явная валидация в коде |
| Swagger из декораторов | OpenAPI как **отдельный контракт** |
| dotenv / config-модуль | `flag` + env → struct |

**Главное:** HTTP в Go — это `http.Handler`. Middleware, routing и JSON — композиция вокруг одного интерфейса.

## Материалы главы

| Файл | Тема |
|------|------|
| [01_net_http.md](./01_net_http.md) | net/http: Handler, Server, graceful shutdown |
| [02_middleware.md](./02_middleware.md) | middleware |
| [03_routing.md](./03_routing.md) | routing (ServeMux Go 1.22+) |
| [04_json.md](./04_json.md) | JSON encode/decode |
| [05_validation.md](./05_validation.md) | validation |
| [06_openapi.md](./06_openapi.md) | OpenAPI |
| [07_configuration.md](./07_configuration.md) | configuration |

## Примеры

### Отдельные темы

```bash
cd c:\go\STUDY_1\Chapter4

go run ./examples/01_net_http/
go run ./examples/02_middleware/
go run ./examples/03_routing/
go run ./examples/04_json/
go run ./examples/05_validation/
go run ./examples/06_openapi/
go run ./examples/07_configuration/
```

### Интеграция (все паттерны вместе)

HTTP API: config → middleware → routes → JSON → validation → OpenAPI:

```bash
go run ./examples/app/cmd/api/
```

По умолчанию слушает `:8080`. Проверка:

```bash
curl http://localhost:8080/healthz
curl -X POST http://localhost:8080/api/v1/tasks -H "Content-Type: application/json" -d "{\"title\":\"learn Go\"}"
curl http://localhost:8080/openapi.yaml
```

Контракт: [`internal/openapi/openapi.yaml`](./examples/app/internal/openapi/openapi.yaml).

## Рекомендуемый порядок

1. **net/http** — Handler, Server, shutdown.
2. **Middleware** — цепочка вокруг Handler.
3. **Routing** — patterns Go 1.22+.
4. **JSON** — encode/decode и ошибки.
5. **Validation** — явные правила и 400.
6. **OpenAPI** — контракт рядом с кодом.
7. **Configuration** — flag + env → struct.
8. **app/api** — собрать всё в сервис.

## Предыдущие этапы

- [Глава 1 — Философия Go](../Chapter1/README.md)
- [Глава 2 — Идиоматичный Go](../Chapter2/README.md)
- [Глава 3 — Concurrency](../Chapter3/README.md)

## Следующий этап

- [Глава 5 — Базы данных](../Chapter5/README.md)
