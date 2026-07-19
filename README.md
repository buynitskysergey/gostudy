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
├── Chapter4/          ← Этап 4: Backend на stdlib
└── Chapter5/          ← Этап 5: Базы данных
    ├── README.md
    ├── 01_…10_*.md
    ├── docker-compose.yml
    └── examples/
        ├── 01_database_sql/ … 10_optimistic_locking/
        └── app/             ← Ledger API: SQL + tx + cache
```

## Прогресс

| Глава | Этап | Статус |
|-------|------|--------|
| [Chapter 1](./Chapter1/README.md) | Философия Go: packages, structs, interfaces, errors, slices, generics | ✅ |
| [Chapter 2](./Chapter2/README.md) | Идиоматичный Go: контракты, errors, layout, DI, context | ✅ |
| [Chapter 3](./Chapter3/README.md) | Concurrency: goroutines, channels, select, sync, errgroup, worker pool | ✅ |
| [Chapter 4](./Chapter4/README.md) | Backend на stdlib: net/http, middleware, routing, JSON, validation, OpenAPI, config | ✅ |
| [Chapter 5](./Chapter5/README.md) | Базы данных: database/sql, pgx, tx, pool, migrations, Redis, ACID, idempotency, optimistic lock | 📖 текущий |

## Быстрый старт

```bash
cd c:\go\STUDY_1

# Глава 5 — Ledger API (SQLite + miniredis)
go run ./Chapter5/examples/app/cmd/api/

# Глава 5 — отдельная тема
go run ./Chapter5/examples/03_transactions/

# Глава 4 — HTTP API
go run ./Chapter4/examples/app/cmd/api/
```

## Глава 5 — что внутри

| Тема | Документ |
|------|----------|
| database/sql | [01_database_sql.md](./Chapter5/01_database_sql.md) |
| pgx | [02_pgx.md](./Chapter5/02_pgx.md) |
| Транзакции | [03_transactions.md](./Chapter5/03_transactions.md) |
| Connection pooling | [04_connection_pooling.md](./Chapter5/04_connection_pooling.md) |
| Миграции | [05_migrations.md](./Chapter5/05_migrations.md) |
| Query optimization | [06_query_optimization.md](./Chapter5/06_query_optimization.md) |
| Redis cache | [07_redis_cache.md](./Chapter5/07_redis_cache.md) |
| ACID | [08_acid.md](./Chapter5/08_acid.md) |
| Идемпотентность | [09_idempotency.md](./Chapter5/09_idempotency.md) |
| Optimistic locking | [10_optimistic_locking.md](./Chapter5/10_optimistic_locking.md) |

## Как учиться

1. Читайте `.md` в порядке из [Chapter5/README.md](./Chapter5/README.md).
2. Запускайте соответствующий пример в `examples/`.
3. Завершите главу через `Chapter5/examples/app/cmd/api/`.
4. Опционально поднимите Postgres/Redis: `docker compose -f Chapter5/docker-compose.yml up -d`.

## Module

Единый Go module `study1` (см. [go.mod](./go.mod)). Import paths:

- `study1/Chapter1/examples/...`
- `study1/Chapter2/examples/...`
- `study1/Chapter3/examples/...`
- `study1/Chapter4/examples/...`
- `study1/Chapter5/examples/...`

## Следующий этап (глава 6)

Тесты, structured logging, observability, graceful patterns в проде.
