# Go Study — от PHP/JS архитектора к production-ready Go

Учебный репозиторий для освоения Go без «пути новичка». Каждая глава — теория (`.md`) + runnable примеры.

## Структура

```
study1/
├── go.mod
├── README.md
├── Chapter1/          ← Этап 1: философия языка
├── Chapter2/          ← Этап 2: идиоматичный Go
└── Chapter3/          ← Этап 3: concurrency
    ├── README.md
    ├── 01_…09_*.md
    └── examples/
        ├── 01_goroutines/ … 09_fan_out_fan_in/
        └── app/             ← pipeline: fan-out + pool + errgroup
```

## Прогресс

| Глава | Этап | Статус |
|-------|------|--------|
| [Chapter 1](./Chapter1/README.md) | Философия Go: packages, structs, interfaces, errors, slices, generics | ✅ |
| [Chapter 2](./Chapter2/README.md) | Идиоматичный Go: контракты, errors, layout, DI, context | ✅ |
| [Chapter 3](./Chapter3/README.md) | Concurrency: goroutines, channels, select, sync, errgroup, worker pool | 📖 текущий |

## Быстрый старт

```bash
cd c:\go\STUDY_1

# Глава 3 — интеграционный pipeline
go run ./Chapter3/examples/app/cmd/processor/

# Глава 3 — отдельная тема
go run ./Chapter3/examples/06_errgroup/
```

## Глава 3 — что внутри

| Тема | Документ |
|------|----------|
| Goroutines | [01_goroutines.md](./Chapter3/01_goroutines.md) |
| Channels | [02_channels.md](./Chapter3/02_channels.md) |
| Select | [03_select.md](./Chapter3/03_select.md) |
| Mutex / RWMutex | [04_sync_mutex.md](./Chapter3/04_sync_mutex.md) |
| WaitGroup | [05_waitgroup.md](./Chapter3/05_waitgroup.md) |
| errgroup | [06_errgroup.md](./Chapter3/06_errgroup.md) |
| Context cancellation | [07_context_cancellation.md](./Chapter3/07_context_cancellation.md) |
| Worker pool | [08_worker_pool.md](./Chapter3/08_worker_pool.md) |
| Fan-out / fan-in | [09_fan_out_fan_in.md](./Chapter3/09_fan_out_fan_in.md) |

## Как учиться

1. Читайте `.md` в порядке номеров.
2. Запускайте соответствующий пример в `examples/`.
3. Завершите главу через `Chapter3/examples/app/cmd/processor/`.
4. Для concurrent кода: `go run -race ./Chapter3/examples/04_mutex/`.

## Module

Единый Go module `study1` (см. [go.mod](./go.mod)). Import paths:

- `study1/Chapter1/examples/...`
- `study1/Chapter2/examples/...`
- `study1/Chapter3/examples/...`

## Следующий этап (глава 4)

HTTP-сервис: router, middleware, structured logging, тесты, graceful shutdown, конфигурация.
