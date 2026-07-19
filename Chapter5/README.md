# Глава 5 — Базы данных

От in-memory store главы 4 к **production persistence**: `database/sql`, pgx, транзакции, пулы, миграции, кэш и паттерны целостности (ACID, идемпотентность, optimistic locking).

## Философия данных в Go

> **The database is not an ORM — you own the SQL.**

| PHP/JS привычка | Go |
|-----------------|-----|
| Eloquent / Prisma / TypeORM | `database/sql` или pgx + явный SQL |
| «магические» модели | struct + scan в поля |
| скрытые N+1 | вы сами пишете JOIN / batch |
| connection «где-то в контейнере» | `*sql.DB` / `*pgxpool.Pool` в DI |
| migrate через artisan | SQL-файлы + версионирование |
| Redis «для всего» | кэш с явным TTL и invalidation |

**Главное:** драйвер даёт соединения и запросы; целостность, идемпотентность и производительность — ваша ответственность в коде и схеме.

## Материалы главы

| Файл | Тема |
|------|------|
| [01_database_sql.md](./01_database_sql.md) | database/sql |
| [02_pgx.md](./02_pgx.md) | pgx |
| [03_transactions.md](./03_transactions.md) | транзакции |
| [04_connection_pooling.md](./04_connection_pooling.md) | connection pooling |
| [05_migrations.md](./05_migrations.md) | миграции |
| [06_query_optimization.md](./06_query_optimization.md) | query optimization |
| [07_redis_cache.md](./07_redis_cache.md) | Redis cache |
| [08_acid.md](./08_acid.md) | ACID |
| [09_idempotency.md](./09_idempotency.md) | идемпотентность |
| [10_optimistic_locking.md](./10_optimistic_locking.md) | optimistic locking |

## Примеры

### Отдельные темы

Примеры на **SQLite** (pure Go) и **miniredis** — запускаются без Docker:

```bash
cd c:\go\STUDY_1\Chapter5

go run ./examples/01_database_sql/
go run ./examples/02_pgx/
go run ./examples/03_transactions/
go run ./examples/04_pooling/
go run ./examples/05_migrations/
go run ./examples/06_query_optimization/
go run ./examples/07_redis_cache/
go run ./examples/08_acid/
go run ./examples/09_idempotency/
go run ./examples/10_optimistic_locking/
```

`02_pgx`: по умолчанию offline-demo интерфейса; для реального Postgres:

```bash
docker compose -f docker-compose.yml up -d
$env:DATABASE_URL="postgres://study:study@localhost:5432/study5?sslmode=disable"
go run ./examples/02_pgx/
```

### Интеграция (все паттерны вместе)

Ledger API: миграции → пул → транзакции → optimistic lock → idempotency key → Redis cache:

```bash
go run ./examples/app/cmd/api/
```

По умолчанию SQLite (`./data/ledger.db`) + in-process Redis (miniredis). Проверка:

```bash
curl http://localhost:8080/healthz
curl -X POST http://localhost:8080/api/v1/accounts -H "Content-Type: application/json" -d "{\"owner\":\"alice\",\"balance_cents\":10000}"
curl -X POST http://localhost:8080/api/v1/transfers -H "Content-Type: application/json" -H "Idempotency-Key: t-1" -d "{\"from_id\":1,\"to_id\":2,\"amount_cents\":500}"
```

Опционально Postgres + Redis через [docker-compose.yml](./docker-compose.yml).

## Рекомендуемый порядок

1. **database/sql** — `DB`, `Query`/`Exec`, `Scan`, `context`.
2. **Connection pooling** — `SetMaxOpenConns` и жизнь пула.
3. **ACID** — что гарантирует БД (и чего нет).
4. **Transactions** — `BeginTx`, commit/rollback, ошибки.
5. **pgx** — Postgres-native API и когда уходить со stdlib.
6. **Migrations** — версионирование схемы.
7. **Query optimization** — индексы, EXPLAIN, N+1.
8. **Optimistic locking** — `version` и конфликт 409.
9. **Idempotency** — повтор безопасных мутаций.
10. **Redis cache** — cache-aside, TTL, invalidation.
11. **app/api** — собрать ledger-сервис.

## Предыдущие этапы

- [Глава 1 — Философия Go](../Chapter1/README.md)
- [Глава 2 — Идиоматичный Go](../Chapter2/README.md)
- [Глава 3 — Concurrency](../Chapter3/README.md)
- [Глава 4 — Backend на stdlib](../Chapter4/README.md)
