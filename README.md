# Go Study — от PHP/JS архитектора к production-ready Go

Учебный репозиторий для освоения Go без «пути новичка». Каждая глава — теория (`.md`) + runnable примеры.

## Структура

```
study1/
├── go.mod
├── README.md
├── Chapter1/          ← Этап 1: философия языка
├── Chapter2/          ← Этап 2: идиоматичный Go
├── Chapter3/          ← Этап 3: concurrency
└── Chapter4/          ← Этап 4: Backend на stdlib
    ├── README.md
    ├── 01_…07_*.md
    └── examples/
        ├── 01_net_http/ … 07_configuration/
        └── app/             ← HTTP API: middleware + JSON + OpenAPI
```

## Прогресс

| Глава | Этап | Статус |
|-------|------|--------|
| [Chapter 1](./Chapter1/README.md) | Философия Go: packages, structs, interfaces, errors, slices, generics | ✅ |
| [Chapter 2](./Chapter2/README.md) | Идиоматичный Go: контракты, errors, layout, DI, context | ✅ |
| [Chapter 3](./Chapter3/README.md) | Concurrency: goroutines, channels, select, sync, errgroup, worker pool | ✅ |
| [Chapter 4](./Chapter4/README.md) | Backend на stdlib: net/http, middleware, routing, JSON, validation, OpenAPI, config | 📖 текущий |

## Быстрый старт

```bash
cd c:\go\STUDY_1

# Глава 4 — HTTP API
go run ./Chapter4/examples/app/cmd/api/

# Глава 4 — отдельная тема
go run ./Chapter4/examples/02_middleware/

# Глава 3 — pipeline
go run ./Chapter3/examples/app/cmd/processor/
```

## Глава 4 — что внутри

| Тема | Документ |
|------|----------|
| net/http | [01_net_http.md](./Chapter4/01_net_http.md) |
| Middleware | [02_middleware.md](./Chapter4/02_middleware.md) |
| Routing | [03_routing.md](./Chapter4/03_routing.md) |
| JSON | [04_json.md](./Chapter4/04_json.md) |
| Validation | [05_validation.md](./Chapter4/05_validation.md) |
| OpenAPI | [06_openapi.md](./Chapter4/06_openapi.md) |
| Configuration | [07_configuration.md](./Chapter4/07_configuration.md) |

## Как учиться

1. Читайте `.md` в порядке номеров.
2. Запускайте соответствующий пример в `examples/`.
3. Завершите главу через `Chapter4/examples/app/cmd/api/`.
4. Сверяйте handlers с `internal/openapi/openapi.yaml`.

## Module

Единый Go module `study1` (см. [go.mod](./go.mod)). Import paths:

- `study1/Chapter1/examples/...`
- `study1/Chapter2/examples/...`
- `study1/Chapter3/examples/...`
- `study1/Chapter4/examples/...`

## Следующий этап (глава 5)

Тесты, structured logging, observability, graceful patterns в проде.
