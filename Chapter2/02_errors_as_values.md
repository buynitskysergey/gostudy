# Ошибки как значения

В идиоматичном Go ошибка — **обычное возвращаемое значение**, не control flow через exceptions.

```go
func FindByID(ctx context.Context, id string) (Order, error) {
    // ...
    return Order{}, ErrNotFound
}
```

Вызывающий **обязан** явно решить: вернуть дальше, обработать, залогировать.

---

## Философия для production

1. **Ошибка — часть API функции.** Сигнатура `(T, error)` документирует возможный сбой.
2. **Не panic для бизнес-логики.** `panic` — для программных багов и fatal на старте.
3. **Zero value + error.** При ошибке возвращайте zero value первого результата:

```go
return Order{}, err  // не return nil, err для struct
return "", err       // для string
return 0, err        // для int
```

4. **Именование:** `ErrNotFound`, `ErrInvalidInput` — sentinel errors. `ValidationError` — typed error.

---

## Sentinel errors

Предопределённые ошибки для сравнения через `errors.Is`:

```go
var (
    ErrNotFound      = errors.New("order not found")
    ErrInvalidAmount = errors.New("invalid amount")
)
```

Использование:

```go
if errors.Is(err, order.ErrNotFound) {
    // HTTP 404, fallback, retry — решение на уровне handler
}
```

**Когда sentinel:** фиксированный набор исходов, одинаковый смысл на всех слоях.

---

## Typed errors

Когда вызывающему нужны **поля**:

```go
type ValidationError struct {
    Field string
    Msg   string
}

func (e *ValidationError) Error() string {
    return fmt.Sprintf("%s: %s", e.Field, e.Msg)
}
```

Извлечение через `errors.As` (см. [03_wrapping_errors.md](./03_wrapping_errors.md)).

---

## Поток ошибок по слоям

```
Handler     → решает HTTP-код, сообщение клиенту
Service     → добавляет контекст операции (wrapping)
Repository  → возвращает ErrNotFound или infra error
```

| Слой | Ответственность |
|------|-----------------|
| Repository | «нет записи», «connection refused» |
| Service | «create order ORD-1: order not found» |
| Handler | `404`, `500`, логирование |

Не логируйте и не обрабатывайте одну ошибку на **каждом** слое — обычно wrap на service, log + map to HTTP на handler.

---

## Сравнение с PHP

| PHP | Go |
|-----|-----|
| `throw new NotFoundException()` | `return ErrNotFound` |
| `catch (NotFoundException $e)` | `errors.Is(err, ErrNotFound)` |
| Exception bubbling | Явный `return err` на каждом уровне |

Плюс Go: видны все точки отказа в коде. Минус: больше `if err != nil` — это норма.

---

## Anti-patterns

```go
// ❌ Игнорирование
result, _ := repo.FindByID(ctx, id)

// ❌ panic для «user not found»
if !found { panic("not found") }

// ❌ string comparison
if err.Error() == "not found" { ... }

// ❌ return nil, nil — двусмысленно
func Find(id string) (*Order, error) {
    if !found { return nil, nil } // caller не отличит «нет» от «ошибки»
}
```

---

## Пример

[examples/02_errors/](./examples/02_errors/)

```bash
go run ./examples/02_errors/
```

В `app/`: [examples/app/internal/order/errors.go](./examples/app/internal/order/errors.go)
