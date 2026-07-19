# Optimistic locking

Оптимистичная блокировка: не держим row lock надолго, а проверяем, что строка **не изменилась** с момента чтения. Колонка `version` (или `updated_at`).

---

## Схема

```sql
CREATE TABLE accounts (
  id            INTEGER PRIMARY KEY,
  owner         TEXT NOT NULL,
  balance_cents INTEGER NOT NULL,
  version       INTEGER NOT NULL DEFAULT 1
);
```

---

## Update с проверкой версии

```go
res, err := tx.ExecContext(ctx, `
    UPDATE accounts
    SET balance_cents = ?, version = version + 1
    WHERE id = ? AND version = ?`,
    newBalance, id, expectedVersion,
)
n, _ := res.RowsAffected()
if n == 0 {
    return ErrConflict // кто-то успел раньше → 409 + retry
}
```

Клиент (или сервис) при конфликте: перечитать → пересчитать → повторить.

---

## Optimistic vs pessimistic

| | Optimistic (`version`) | Pessimistic (`SELECT FOR UPDATE`) |
|--|------------------------|-------------------------------------|
| Конфликты редки | ✅ | избыточно |
| Конфликты часты | много retry | ✅ |
| Длинная бизнес-логика | не держит lock | держит строку |
| HTTP API edit | классика | осторожно с latency |

Для перевода денег часто комбинируют: короткая tx + условие на баланс (+ иногда `version`).

---

## API-семантика

```http
HTTP/1.1 409 Conflict
{"error":"version conflict","hint":"reload and retry"}
```

Можно использовать `ETag` / `If-Match` на HTTP-уровне — та же идея.

---

## Связь с ACID и идемпотентностью

- Optimistic lock защищает от **lost update**.
- Идемпотентность защищает от **дубля запроса**.
- Это разные оси: нужны обе для платёжных API.

---

## Anti-patterns

```go
// ❌ Читать version, считать в Go, UPDATE без WHERE version = ?
// ❌ Глотать RowsAffected == 0 как успех
// ❌ Бесконечный retry без backoff / лимита
```

---

## Пример

[examples/10_optimistic_locking/](./examples/10_optimistic_locking/)

```bash
go run ./examples/10_optimistic_locking/
```

Два конкурентных обновления: один успех, второй `ErrConflict`.
