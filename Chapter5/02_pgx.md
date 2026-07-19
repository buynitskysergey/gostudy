# pgx

[pgx](https://github.com/jackc/pgx) — Postgres-native драйвер. Богаче `database/sql` на PostgreSQL: batch, listen/notify, правильные типы, `pgxpool`, копирование через `CopyFrom`.

Два режима использования:

| Режим | Пакет | Когда |
|-------|-------|-------|
| Native | `github.com/jackc/pgx/v5` | новый Postgres-сервис |
| Stdlib bridge | `pgx/stdlib` → `database/sql` | нужна совместимость с кодом на `*sql.DB` |

---

## Pool

```go
cfg, err := pgxpool.ParseConfig(os.Getenv("DATABASE_URL"))
if err != nil {
    return err
}
cfg.MaxConns = 20
cfg.MinConns = 2

pool, err := pgxpool.NewWithConfig(ctx, cfg)
defer pool.Close()

var balance int64
err = pool.QueryRow(ctx,
    `SELECT balance_cents FROM accounts WHERE id = $1`, id,
).Scan(&balance)
```

Плейсхолдеры Postgres — `$1`, `$2` (не `?`).

---

## Чем pgx удобнее stdlib на Postgres

| Возможность | Зачем |
|-------------|-------|
| `pgx.RowToStructByName` | меньше ручного Scan |
| `SendBatch` | несколько запросов за один round-trip |
| `CopyFrom` | быстрая bulk-загрузка |
| `pgtype` | NULL и Postgres-типы без сюрпризов |
| `BeginFunc` | транзакция с автоматическим rollback при error |

```go
err := pgx.BeginFunc(ctx, pool, func(tx pgx.Tx) error {
    _, err := tx.Exec(ctx, `UPDATE accounts SET balance_cents = balance_cents - $1 WHERE id = $2`, amount, fromID)
    if err != nil {
        return err
    }
    _, err = tx.Exec(ctx, `UPDATE accounts SET balance_cents = balance_cents + $1 WHERE id = $2`, amount, toID)
    return err
})
```

---

## Когда оставаться на database/sql

- Мульти-СУБД (SQLite в тестах, Postgres в проде) с одним кодом.
- Библиотеки принимают только `*sql.DB`.
- Команда уже стандартизировала stdlib.

Когда нужен pgx: Postgres-only сервис, производительность batch/copy, тонкий контроль типов и пула.

---

## Repository: интерфейс у потребителя

Как в главе 2 — домен не импортирует pgx:

```go
type AccountRepository interface {
    Get(ctx context.Context, id int64) (Account, error)
    Transfer(ctx context.Context, from, to int64, amount int64) error
}
```

`pgx` / `sql` — детали адаптера в `internal/storage/postgres`.

---

## Anti-patterns

```go
// ❌ Мешать ? и $1 в одном коде «на глаз»
// ❌ Держать *pgx.Conn вместо pool в HTTP-сервисе
// ❌ Тащить pgx.Tx в handlers — tx живёт в repository/service
```

---

## Пример

[examples/02_pgx/](./examples/02_pgx/)

```bash
# offline: тот же контракт репозитория без Postgres
go run ./examples/02_pgx/

# real Postgres
docker compose -f docker-compose.yml up -d
set DATABASE_URL=postgres://study:study@localhost:5432/study5?sslmode=disable
go run ./examples/02_pgx/
```
