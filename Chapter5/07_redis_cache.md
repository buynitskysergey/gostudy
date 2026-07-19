# Redis cache

Redis — не «вторая БД», а **кэш и координация**: горячие чтения, rate limit, distributed lock, pub/sub. Источник истины остаётся в Postgres/SQLite.

---

## Cache-aside (самый частый паттерн)

```go
func (s *Service) GetAccount(ctx context.Context, id int64) (Account, error) {
    key := fmt.Sprintf("account:%d", id)
    if raw, err := s.rdb.Get(ctx, key).Bytes(); err == nil {
        var a Account
        if json.Unmarshal(raw, &a) == nil {
            return a, nil
        }
    }

    a, err := s.repo.Get(ctx, id) // source of truth
    if err != nil {
        return Account{}, err
    }
    b, _ := json.Marshal(a)
    _ = s.rdb.Set(ctx, key, b, 30*time.Second).Err()
    return a, nil
}
```

При записи — **инвалидация**:

```go
_ = s.rdb.Del(ctx, fmt.Sprintf("account:%d", id)).Err()
```

---

## TTL и согласованность

| Стратегия | Смысл |
|-----------|-------|
| TTL only | просто; возможна краткая устаревшесть |
| TTL + invalidate on write | обычно достаточно для CRUD |
| Write-through | писать в кэш вместе с БД — сложнее |

Кэш **может врать** короткое время — заложите это в продукт (или не кэшируйте деньги без версии).

---

## Клиент в Go

```go
rdb := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
```

Всегда `context`, таймауты, обработка `redis.Nil` (промах кэша ≠ ошибка сервиса).

Пример сначала пробует Redis из [docker-compose.yml](./docker-compose.yml) (`localhost:6379`, или `REDIS_ADDR`).
Если недоступен — fallback на [miniredis](https://github.com/alicebob/miniredis): in-process Redis, тот же клиент `go-redis`.

---

## Чего не класть в Redis «по умолчанию»

- Единственную копию критичных данных без persistence-плана
- Огромные ключи без TTL (утечка памяти)
- Сессии/токены без продуманного expiry

---

## vs PHP/JS

| Laravel Cache / ioredis | Go |
|-------------------------|-----|
| `Cache::remember` | явный Get → DB → Set |
| facade | `*redis.Client` в DI |
| «кэш сам инвалидируется» | вы зовёте `Del` / короткий TTL |

---

## Anti-patterns

```go
// ❌ Игнорировать ошибку Redis на read path — лучше деградировать в БД
// ❌ Кэшировать без TTL «навсегда»
// ❌ Считать Del достаточным при multi-key связанных данных без плана
```

---

## Пример

[examples/07_redis_cache/](./examples/07_redis_cache/)

```bash
# опционально: docker compose up -d redis
go run ./examples/07_redis_cache/
```

Cache-aside: Redis из compose (или miniredis fallback) — miss → hit → invalidate.
