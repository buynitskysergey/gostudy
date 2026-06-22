# Глава 3 — Concurrency

Здесь начинается **настоящий Go**. Goroutines, channels и context — то, ради чего язык выбирают для backend-сервисов. И то, что чаще всего misunderstood даже у опытных разработчиков.

## Философия concurrency в Go

> **Don't communicate by sharing memory; share memory by communicating.**

| PHP/JS привычка | Go |
|-----------------|-----|
| async/await, Promises | goroutines + channels |
| `setInterval`, event loop один | M:N scheduler, тысячи goroutines |
| `Mutex` редко | `sync` или channels — явный выбор |
| AbortController | `context.Context` + `select` |
| `Promise.all` | `errgroup` или `WaitGroup` |

**Главное:** concurrency в Go — не «магия runtime», а **явные приimitives**, которые вы комбинируете.

## Материалы главы

| Файл | Тема |
|------|------|
| [01_goroutines.md](./01_goroutines.md) | goroutine |
| [02_channels.md](./02_channels.md) | channel |
| [03_select.md](./03_select.md) | select |
| [04_sync_mutex.md](./04_sync_mutex.md) | sync.Mutex, sync.RWMutex |
| [05_waitgroup.md](./05_waitgroup.md) | sync.WaitGroup |
| [06_errgroup.md](./06_errgroup.md) | errgroup |
| [07_context_cancellation.md](./07_context_cancellation.md) | context cancellation |
| [08_worker_pool.md](./08_worker_pool.md) | worker pool |
| [09_fan_out_fan_in.md](./09_fan_out_fan_in.md) | fan-out / fan-in |

## Примеры

### Отдельные темы

```bash
cd c:\go\STUDY_1\Chapter3

go run ./examples/01_goroutines/
go run ./examples/02_channels/
go run ./examples/03_select/
go run ./examples/04_mutex/
go run ./examples/05_waitgroup/
go run ./examples/06_errgroup/
go run ./examples/07_context/
go run ./examples/08_worker_pool/
go run ./examples/09_fan_out_fan_in/
```

### Интеграция (все паттерны вместе)

Pipeline: fan-out → worker pool → fan-in, с `context` и `errgroup`:

```bash
go run ./examples/app/cmd/processor/
```

## Рекомендуемый порядок

1. **Goroutines** — что такое легковесный поток.
2. **Channels** — как goroutines общаются.
3. **Select** — мультиплексирование channel operations.
4. **Mutex / RWMutex** — когда channels не подходят (shared state).
5. **WaitGroup** — дождаться N goroutines.
6. **Context cancellation** — отмена дерева goroutines.
7. **errgroup** — WaitGroup + первая ошибка + cancel.
8. **Worker pool + fan-out/fan-in** — production-паттерны.
9. **app/processor** — собрать всё в pipeline.

## Предыдущие этапы

- [Глава 1 — Философия Go](../Chapter1/README.md)
- [Глава 2 — Идиоматичный Go](../Chapter2/README.md)
