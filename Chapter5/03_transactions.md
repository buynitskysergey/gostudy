# Транзакции

Транзакция — граница **атомарности**: либо все изменения видны, либо ни одно. В Go это явный `Begin` → работа → `Commit` / `Rollback`.

---

## database/sql

```go
tx, err := db.BeginTx(ctx, &sql.TxOptions{
    Isolation: sql.LevelReadCommitted, // default часто ок
    ReadOnly:  false,
})
if err != nil {
    return err
}
defer tx.Rollback() // no-op после Commit; всегда defer

if _, err := tx.ExecContext(ctx, `UPDATE accounts SET balance_cents = balance_cents - ? WHERE id = ? AND balance_cents >= ?`, amount, fromID, amount); err != nil {
    return err
}
if _, err := tx.ExecContext(ctx, `UPDATE accounts SET balance_cents = balance_cents + ? WHERE id = ?`, amount, toID); err != nil {
    return err
}
return tx.Commit()
```

Паттерн **defer Rollback + return Commit**: при ошибке или panic откат; при успехе Commit, Rollback становится пустым.

---

## Проверка «затронута ли строка»

```go
res, err := tx.ExecContext(ctx,
    `UPDATE accounts SET balance_cents = balance_cents - ? WHERE id = ? AND balance_cents >= ?`,
    amount, fromID, amount,
)
n, _ := res.RowsAffected()
if n == 0 {
    return ErrInsufficientFunds
}
```

Без проверки легко «успешно» списать с несуществующего счёта.

---

## Изоляция (кратко)

| Уровень | Типичный смысл |
|---------|----------------|
| Read Committed | default Postgres: не видите dirty reads |
| Repeatable Read | снимок на время tx (в PG — как snapshot) |
| Serializable | строже, больше конфликтов/retry |

Повышайте изоляцию **точечно**, не глобально «на всякий случай». См. [08_acid.md](./08_acid.md).

---

## Контекст и отмена

Если `ctx` отменён во время tx, драйвер прервёт запросы — **обязательно** Rollback (defer закрывает этот случай). Не переиспользуйте `tx` после ошибки.

---

## vs PHP/JS

| Laravel / Prisma | Go |
|------------------|-----|
| `DB::transaction(fn)` | `BeginTx` + defer Rollback + Commit |
| неявный commit в middleware | явная граница в service/repo |
| nested transactions (savepoints) | savepoints вручную, редко нужны |

---

## Anti-patterns

```go
// ❌ Длинная tx с HTTP/RPC внутри — держите блокировки БД коротко
// ❌ Забыть Rollback при ошибке
// ❌ Ловить error и всё равно Commit
// ❌ Передавать *sql.Tx вверх в HTTP layer без нужды
```

---

## Пример

[examples/03_transactions/](./examples/03_transactions/)

```bash
go run ./examples/03_transactions/
```

Перевод средств: с tx баланс сходится; без tx — нет.
