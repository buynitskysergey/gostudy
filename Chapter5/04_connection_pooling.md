# Connection pooling

`*sql.DB` и `pgxpool.Pool` — это **пулы**. Каждый `Query` берёт соединение, возвращает его после завершения rows/exec. Понимание лимитов пула = стабильность под нагрузкой.

---

## Ключевые настройки database/sql

```go
db.SetMaxOpenConns(25)                 // потолок одновременных conn к БД
db.SetMaxIdleConns(10)                 // сколько держать тёплыми
db.SetConnMaxLifetime(30 * time.Minute) // ротация (failover, NAT, LB)
db.SetConnMaxIdleTime(5 * time.Minute) // не копить зомби idle
```

| Параметр | Слишком мало | Слишком много |
|----------|--------------|---------------|
| MaxOpen | очереди, latency | исчерпание `max_connections` Postgres |
| MaxIdle | холодный старт запросов | лишняя память на стороне БД |
| MaxLifetime | — | обрывы долгих операций, если слишком коротко |

Правило большого пальца: **MaxOpen на инстанс × число реплик < Postgres max_connections** (с запасом для админки и миграций).

---

## Симптомы неправильного пула

- `too many connections` в логах Postgres
- запросы висят, хотя CPU БД свободен (ждут conn в приложении)
- всплески latency после idle (мертвые conn за LB) — лечится `ConnMaxLifetime` + `Ping`

---

## Rows и утечки

Незакрытый `*sql.Rows` **держит** соединение:

```go
rows, err := db.QueryContext(ctx, q)
if err != nil {
    return err
}
defer rows.Close() // обязательно
```

То же для `Tx` — незавершённая транзакция занимает conn до Rollback/Commit.

---

## pgxpool

```go
cfg, _ := pgxpool.ParseConfig(url)
cfg.MaxConns = 25
cfg.MinConns = 2
cfg.MaxConnLifetime = 30 * time.Minute
cfg.MaxConnIdleTime = 5 * time.Minute
pool, _ := pgxpool.NewWithConfig(ctx, cfg)
```

Семантика та же: лимит на процесс, health через lifetime/idle.

---

## vs PHP/JS

| PHP-FPM / Node | Go |
|----------------|-----|
| часто 1 conn на worker | один пул на процесс, тысячи goroutines |
| «открыл в скрипте — закрыл» | conn возвращается в пул автоматически |
| RDS Proxy обязателен «всегда» | сначала посчитайте MaxOpen × pods |

---

## Anti-patterns

```go
// ❌ MaxOpenConns = 0 «без лимита» в проде на 50 подах
// ❌ Один глобальный *sql.Tx на всё приложение
// ❌ sql.Open в handler
```

---

## Пример

[examples/04_pooling/](./examples/04_pooling/)

```bash
go run ./examples/04_pooling/
```

8 воркеров × 2с «работы», но `MaxOpenConns=2`: conn держится через `db.Conn` на всё время работы → ~4 волны, ~8с wall time, растёт `WaitCount`.
