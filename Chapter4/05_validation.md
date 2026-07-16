# Validation

В stdlib **нет** готового validator-а с тегами. Это фича: правила — обычный Go-код, который легко читать и тестировать.

---

## Явная функция Validate

```go
type CreateTaskRequest struct {
    Title string `json:"title"`
}

func (r CreateTaskRequest) Validate() error {
    title := strings.TrimSpace(r.Title)
    if title == "" {
        return validationError{fields: map[string]string{
            "title": "required",
        }}
    }
    if len(title) > 200 {
        return validationError{fields: map[string]string{
            "title": "max length is 200",
        }}
    }
    return nil
}
```

Handler:

```go
var req CreateTaskRequest
if err := decodeJSON(r, &req); err != nil {
    writeJSON(w, 400, ErrorBody{Error: "invalid json"})
    return
}
if err := req.Validate(); err != nil {
    writeJSON(w, 400, err) // field errors
    return
}
```

---

## Field errors (удобный JSON)

```go
type validationError struct {
    fields map[string]string
}

func (e validationError) Error() string { return "validation failed" }

func (e validationError) MarshalJSON() ([]byte, error) {
    return json.Marshal(map[string]any{
        "error":  "validation failed",
        "fields": e.fields,
    })
}
```

Ответ:

```json
{
  "error": "validation failed",
  "fields": { "title": "required" }
}
```

Клиент (фронт / мобилка) может подсветить поля — как у class-validator / Zod.

---

## Слои валидации

| Слой | Что проверяет |
|------|----------------|
| JSON decode | синтаксис, типы, unknown fields |
| Request.Validate() | формат, длины, enum, required |
| Domain / service | бизнес-правила (уникальность, статусы) |

Не смешивайте: «title пустой» — 400 validation; «task already archived» — 409/422 domain.

---

## Когда тег-валидаторы

`go-playground/validator` и подобные уместны на больших DTO. Для обучения и большинства handlers **явный код яснее**:

```go
// видно сразу
if req.Priority < 1 || req.Priority > 5 { ... }

// vs магия
`validate:"required,min=1,max=5"`
```

---

## Validation vs PHP/JS

| Laravel / Zod | Go stdlib |
|---------------|-----------|
| Form Request / schema.parse | `Validate() error` |
| `$errors->toArray()` | `fields` map в JSON |
| 422 ValidationException | `400` или `422` — выберите конвенцию и держитесь |

В этой главе используем **400** для невалидного ввода (простая HTTP-семантика).

---

## Anti-patterns

```go
// ❌ Валидировать только в SQL constraints — плохой UX
// ❌ Паниковать на bad input
// ❌ Возвращать разные форматы ошибок из разных handlers
```

---

## Пример

[examples/05_validation/](./examples/05_validation/)

```bash
go run ./examples/05_validation/
```

Validate + единый JSON с `fields`.
