# Wrapping errors

**Wrapping** — добавление контекста к ошибке при движении вверх по стеку вызовов, **без потери** исходной причины.

```go
if err := s.repo.Save(ctx, o); err != nil {
    return Order{}, fmt.Errorf("create order %s: %w", id, err)
}
```

`%w` — ключевое слово Go 1.13+. Без него цепочка рвётся.

---

## errors.Is — проверка sentinel в цепочке

```go
err := svc.Get(ctx, "missing")
if errors.Is(err, order.ErrNotFound) {
    // сработает, даже если service обернул:
    // "get order missing: order not found"
}
```

`errors.Is` обходит всю цепочку wrap'ов.

---

## errors.As — извлечение typed error

```go
var ve *ValidationError
if errors.As(err, &ve) {
    fmt.Println(ve.Field)
}
```

Полезно на границе HTTP: mapped status code по типу ошибки.

---

## Правила wrapping в production

1. **Wrap на границе слоя** — service wrap'ит repo, handler обычно не wrap'ит (логирует as-is).
2. **Сообщение wrap'а — что делали**, не дублируйте текст исходной ошибки:

```go
// ✅
return fmt.Errorf("save order %s: %w", id, err)

// ❌ избыточно
return fmt.Errorf("save order failed: %s: %w", err.Error(), err)
```

3. **Не wrap'ите всё подряд** — `ErrNotFound` часто пробрасывают без изменений или с одним wrap.
4. **Внешние API** — не утекайте внутренние детали клиенту; в лог — полная цепочка.

---

## fmt.Errorf vs errors.Join

```go
// Один контекст — %w
fmt.Errorf("fetch user: %w", err)

// Несколько независимых ошибок (Go 1.20+)
errors.Join(err1, err2)
```

---

## Unwrap для своих типов

```go
type OpError struct {
    Op  string
    Err error
}

func (e *OpError) Error() string { return e.Op + ": " + e.Err.Error() }
func (e *OpError) Unwrap() error { return e.Err }
```

Тогда `errors.Is` / `As` работают через ваш тип.

---

## Сравнение с PHP

| PHP | Go |
|-----|-----|
| `$e->getPrevious()` | `errors.Unwrap(err)` / цепочка `%w` |
| `instanceof` | `errors.As` |
| Exception message concatenation | `fmt.Errorf("context: %w", err)` |

---

## Anti-patterns

```go
// ❌ %v вместо %w — теряется цепочка
fmt.Errorf("failed: %v", err)

// ❌ string matching
strings.Contains(err.Error(), "not found")

// ❌ wrap на каждой строке
if err != nil { return fmt.Errorf("step1: %w", err) }
if err != nil { return fmt.Errorf("step2: %w", err) } // один wrap на операцию достаточно
```

---

## Пример

[examples/03_wrapping/](./examples/03_wrapping/)

```bash
go run ./examples/03_wrapping/
```

В `app/`: service wrap'ит ошибки repo — [examples/app/internal/order/service.go](./examples/app/internal/order/service.go)
