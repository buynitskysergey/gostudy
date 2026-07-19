# ACID

ACID — контракт транзакции СУБД. В Go вы **не получаете ACID «от ORM»** — вы открываете транзакцию и пишете SQL внутри неё.

---

## Четыре свойства

| Буква | Смысл | Практика |
|-------|-------|----------|
| **A**tomicity | всё или ничего | `Begin`/`Commit`/`Rollback` |
| **C**onsistency | инварианты схемы и constraints | CHECK, FK, UNIQUE, логика в tx |
| **I**solation | параллельные tx не ломают друг друга | уровень изоляции + retry |
| **D**urability | после commit данные на диске | fsync СУБД; не путать с кэшем Redis |

---

## Consistency ≠ «бизнес всегда прав»

БД гарантирует **constraints** (NOT NULL, FK, UNIQUE, CHECK). Инвариант «баланс ≥ 0 при переводе» вы обеспечиваете SQL-условием или блокировкой **внутри** tx — иначе Consistency на уровне домена нарушена, хотя ACID формально «есть».

```sql
UPDATE accounts
SET balance_cents = balance_cents - 500
WHERE id = 1 AND balance_cents >= 500;
-- 0 rows → откат бизнес-операции
```

---

## Isolation: что реально ломается

Без правильной изоляции/блокировок:

- **Lost update** — два UPDATE перетирают друг друга
- **Write skew** — каждый tx видит валидное состояние, вместе — нет
- **Non-repeatable read** — повторный SELECT в той же tx даёт другое

Лечение: короткие tx, нужный isolation level, optimistic locking ([10](./10_optimistic_locking.md)), идемпотентность ([09](./09_idempotency.md)).

---

## Durability и кэш

`Commit` в Postgres ≠ ключ в Redis. После commit:

1. Инвалидировать кэш, или
2. Кэшировать с версией / коротким TTL

Иначе пользователь видит «старый» баланс — это не баг ACID БД, а рассинхрон кэша.

---

## vs PHP/JS

| Расхожее мнение | Реальность в Go |
|-----------------|-----------------|
| «Транзакция = ACID автоматически» | только то, что внутри одной tx одной БД |
| «Микросервисы + ACID на всё» | между сервисами — saga/outbox, не одна tx |
| Eloquent `transaction()` магически защищает баланс | нужен правильный SQL / lock |

---

## Anti-patterns

```text
❌ Две отдельные Exec без tx для перевода денег
❌ Полагаться на Redis как на durable ledger
❌ Долгая tx с внешним HTTP внутри
```

---

## Пример

[examples/08_acid/](./examples/08_acid/)

```bash
go run ./examples/08_acid/
```

Atomicity: падение mid-flight без tx vs с tx.
