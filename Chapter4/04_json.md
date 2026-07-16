# JSON

Пакет `encoding/json` — стандарт для API. Без «магии» сериализаторов: struct tags и явный encode/decode.

---

## Struct tags

```go
type Task struct {
    ID        string    `json:"id"`
    Title     string    `json:"title"`
    Done      bool      `json:"done"`
    CreatedAt time.Time `json:"created_at"`
}
```

| Tag | Эффект |
|-----|--------|
| `json:"name"` | имя поля в JSON |
| `json:"-"` | не сериализовать |
| `json:"name,omitempty"` | пропустить zero value |

---

## Decode из request

```go
func decodeJSON(r *http.Request, dst any) error {
    defer r.Body.Close()
    dec := json.NewDecoder(io.LimitReader(r.Body, 1<<20)) // 1 MiB
    dec.DisallowUnknownFields()
    if err := dec.Decode(dst); err != nil {
        return err
    }
    // второй Decode должен дать EOF — иначе в body был лишний JSON
    if err := dec.Decode(&struct{}{}); err != io.EOF {
        return fmt.Errorf("request body must contain a single JSON object")
    }
    return nil
}
```

| Приём | Зачем |
|-------|-------|
| `LimitReader` | защита от huge body |
| `DisallowUnknownFields` | жёсткий контракт |
| проверка на один object | нет `[{...},{...}]` сюрпризов |

---

## Encode в response

```go
func writeJSON(w http.ResponseWriter, status int, v any) {
    w.Header().Set("Content-Type", "application/json; charset=utf-8")
    w.WriteHeader(status)
    if err := json.NewEncoder(w).Encode(v); err != nil {
        // headers уже ушли — только лог
        log.Printf("write json: %v", err)
    }
}
```

Всегда выставляйте `Content-Type` **до** `WriteHeader`.

---

## Ошибки API

Единый envelope упрощает клиентов:

```go
type ErrorBody struct {
    Error string `json:"error"`
}

writeJSON(w, http.StatusBadRequest, ErrorBody{Error: "invalid json"})
```

Для валидации полей — см. [05_validation.md](./05_validation.md).

---

## Pointer vs value для optional

```go
type UpdateTask struct {
    Title *string `json:"title"` // nil = поле не прислали
    Done  *bool   `json:"done"`
}
```

Без указателя `false` и «не передали» неразличимы. Как optional в TypeScript.

---

## JSON vs PHP/JS

| | PHP / JS | Go |
|---|----------|-----|
| Decode | `json_decode` / `JSON.parse` | `json.Decoder` |
| Types | часто `array`/`object` | строгий struct |
| Unknown fields | часто игнор | `DisallowUnknownFields` по желанию |
| Dates | string / Carbon | `time.Time` (RFC3339) |

---

## Anti-patterns

```go
// ❌ json.Unmarshal всего Body в []byte без лимита размера
// ❌ map[string]any везде — теряете контракт
// ❌ Encode без Content-Type
```

---

## Пример

[examples/04_json/](./examples/04_json/)

```bash
go run ./examples/04_json/
```

Encode/decode, unknown fields, limit body.
