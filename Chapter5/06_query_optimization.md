# Query optimization

Медленный API чаще упирается в SQL, чем в Go. Оптимизация = **измерить → индекс/запрос → снова измерить**, не гадать.

---

## EXPLAIN — первый инструмент

```sql
EXPLAIN QUERY PLAN
SELECT * FROM transfers WHERE from_id = 1 ORDER BY created_at DESC LIMIT 20;
```

В Postgres: `EXPLAIN (ANALYZE, BUFFERS)`. Ищите Seq Scan на больших таблицах там, где нужен Index Scan.

---

## Индексы под реальные запросы

```sql
CREATE INDEX transfers_from_created_idx
  ON transfers (from_id, created_at DESC);
```

Индекс окупается, если совпадает с `WHERE` / `JOIN` / `ORDER BY`. Лишние индексы замедляют INSERT/UPDATE.

| Антипаттерн | Почему больно |
|-------------|----------------|
| `WHERE lower(email) = ?` без functional index | не используется index по `email` |
| `SELECT *` + широкие строки | лишний I/O и сеть |
| N+1 в цикле | 1 + N round-trips |

---

## N+1

```go
// ❌ N+1
for _, t := range transfers {
    acc, _ := repo.GetAccount(ctx, t.FromID) // запрос на каждую строку
}

// ✅ один (или два) запроса
accounts, _ := repo.GetAccountsByIDs(ctx, ids)
```

В Go это особенно заметно: goroutines не лечат лишние round-trips к БД.

---

## Покрывающие приёмы

1. **SELECT нужных колонок** — не `*`, если таблица широкая.
2. **Пагинация** — `LIMIT/OFFSET` ок на малых offset; keyset (`WHERE id > ?`) стабильнее.
3. **Batch / IN** — собрать id, один `WHERE id IN (...)`.
4. **Денормализация / кэш** — когда read-модель читается в 100× чаще записи (см. Redis).
5. **Частичные индексы** (Postgres) — `WHERE status = 'open'`.

---

## Измерение в приложении

Логируйте `time.Since(start)` вокруг репозитория + `db.Stats()` под нагрузкой. Медленный запрос без EXPLAIN — гадание.

---

## vs PHP/JS

| Eloquent / Prisma | Go + SQL |
|-------------------|----------|
| скрытый lazy load → N+1 | N+1 только если вы сами так написали |
| `with('relation')` | JOIN или второй batch-запрос явно |
| query log в debug | middleware/обёртка репозитория |

---

## Anti-patterns

```go
// ❌ Оптимизировать до появления EXPLAIN и метрик
// ❌ Индекс «на все колонки» таблицы
// ❌ OFFSET 100000 на горячем path
```

---

## Пример

[examples/06_query_optimization/](./examples/06_query_optimization/)

```bash
go run ./examples/06_query_optimization/
```

N+1 vs batch + сравнение с индексом (`EXPLAIN QUERY PLAN` на SQLite).
