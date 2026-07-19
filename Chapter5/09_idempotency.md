# Идемпотентность

Идемпотентная операция при **повторе** даёт тот же эффект, что и при первом успешном выполнении. Критично для платежей, переводов, webhooks: клиент ретраит, сеть дублирует POST.

---

## Idempotency-Key

Клиент присылает ключ (UUID) в заголовке:

```http
POST /api/v1/transfers
Idempotency-Key: 8f3c-...
```

Сервер:

1. Если ключ уже есть и запрос эквивалентен → вернуть **сохранённый ответ** (тот же status/body).
2. Если ключ новый → выполнить, сохранить результат рядом с ключом.
3. Если ключ есть, но тело другое → **409 Conflict**.

```sql
CREATE TABLE idempotency_keys (
  key          TEXT PRIMARY KEY,
  request_hash TEXT NOT NULL,
  response     BLOB NOT NULL,
  status_code  INT NOT NULL,
  created_at   TIMESTAMP NOT NULL
);
```

Вставку ключа делайте **в той же транзакции**, что и бизнес-эффект (или используйте UNIQUE + «in progress» состояние).

---

## Паттерн в tx

```go
err := InTx(ctx, db, func(tx *sql.Tx) error {
    occupied, err := insertIdempotency(tx, key, hash)
    if err != nil {
        return err
    }
    if !occupied {
        // уже есть — прочитать сохранённый ответ (вне или внутри)
        return errIdempotentReplay
    }
    // ... transfer ...
    return saveIdempotencyResponse(tx, key, status, body)
})
```

`UNIQUE` на `key` защищает от гонки двух параллельных одинаковых запросов.

---

## Что делать идемпотентным

| Операция | Как |
|----------|-----|
| Создание платежа / перевода | Idempotency-Key |
| Webhook обработки | store `event_id` UNIQUE |
| PUT ресурса | натурально (тот же state) |
| POST без ключа | небезопасный retry |

GET/PUT/DELETE по смыслу часто идемпотентны; **POST с побочным эффектом — нет**, пока вы это не обеспечите.

---

## vs PHP/JS

| Stripe-like API | Go-сервис |
|-----------------|-----------|
| middleware сохраняет response | таблица + tx с доменной логикой |
| «просто UUID в redis» | Redis ок для ключа, но ответ надёжнее рядом с ledger в БД |

---

## Anti-patterns

```text
❌ Идемпотентность только in-memory map на одном поде
❌ Ключ без привязки к телу запроса (разные суммы под одним key)
❌ Выполнить перевод, потом отдельно записать key (окно дубля)
```

---

## Пример

[examples/09_idempotency/](./examples/09_idempotency/)

```bash
go run ./examples/09_idempotency/
```

Двойной перевод с одним ключом списывает деньги один раз.
