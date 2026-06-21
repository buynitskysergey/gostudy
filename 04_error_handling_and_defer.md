# Error Handling и Defer

## Error Handling

В Go **нет exceptions**. Ошибки — обычные значения типа `error` (интерфейс с одним методом `Error() string`).

```go
result, err := doSomething()
if err != nil {
    return fmt.Errorf("doSomething failed: %w", err)
}
// используем result
```

### Философия

1. **Явность** — каждая операция, которая может fail, возвращает `error`.
2. **Errors are values** — их можно сравнивать, оборачивать, типизировать.
3. **Нет try/catch** — поток управления линейный, видно все точки отказа.

Типичный паттерн в production-коде:

```go
func (s *Service) CreateUser(ctx context.Context, email string) (User, error) {
    if email == "" {
        return User{}, ErrInvalidEmail
    }
    u := User{Email: email}
    if err := s.repo.Save(ctx, u); err != nil {
        return User{}, fmt.Errorf("create user: %w", err)
    }
    return u, nil
}
```

### Sentinel errors

Предопределённые ошибки для сравнения:

```go
var ErrNotFound = errors.New("not found")
var ErrInvalidEmail = errors.New("invalid email")

if errors.Is(err, ErrNotFound) {
    // 404
}
```

### Wrapping (`%w`)

Go 1.13+: оборачивание сохраняет цепочку:

```go
return fmt.Errorf("fetch user %d: %w", id, err)

// Разбор:
if errors.Is(err, ErrNotFound) { ... }
var ve *ValidationError
if errors.As(err, &ve) { ... }  // извлечь typed error
```

### Custom error types

```go
type ValidationError struct {
    Field   string
    Message string
}

func (e *ValidationError) Error() string {
    return fmt.Sprintf("%s: %s", e.Field, e.Message)
}
```

Используйте когда вызывающему нужны **структурированные данные** об ошибке.

### Сравнение с PHP/JS

| Go | PHP | JS |
|----|-----|-----|
| `(T, error)` | throw / catch | throw / catch |
| `if err != nil` | try/catch блоки | try/catch |
| `errors.Is/As` | `instanceof` / код | `instanceof` |

**Anti-pattern:** игнорировать ошибки `_ = file.Close()`. В production — обрабатывать или явно документировать почему игнорируете.

### panic / recover

`panic` — для **программных багов**, не для бизнес-ошибок. В HTTP-сервисах — только на старте (fatal config) или middleware recover для 500.

---

## Defer

`defer` откладывает вызов функции до **выхода из текущей функции** (обычно для cleanup).

```go
func readFile(path string) ([]byte, error) {
    f, err := os.Open(path)
    if err != nil {
        return nil, err
    }
    defer f.Close()  // выполнится при return, даже при panic

    return io.ReadAll(f)
}
```

### Правила

1. Аргументы defer-вызова **вычисляются сразу**, выполнение — при return.
2. Несколько `defer` — выполняются в **обратном порядке** (LIFO, как stack).
3. Идиоматично для: `Close()`, `Unlock()`, rollback транзакций.

```go
mu.Lock()
defer mu.Unlock()
// ... critical section
```

### defer + named return (осторожно)

```go
func f() (result int, err error) {
    defer func() {
        if err != nil {
            result = 0  // может изменить named return
        }
    }()
    ...
}
```

В production чаще явный cleanup без named returns.

### defer в циклах — ловушка

```go
// Плохо — все Close() выполнятся только при выходе из функции
for _, path := range paths {
    f, _ := os.Open(path)
    defer f.Close()
}

// Хорошо — отдельная функция или Close() сразу
for _, path := range paths {
    if err := processFile(path); err != nil {
        return err
    }
}
```

---

## Пример

Смотрите: [examples/04_errors/](./examples/04_errors/)

Запуск:

```bash
go run ./examples/04_errors/
```

Демонстрирует: sentinel errors, wrapping, custom types, defer для cleanup, порядок defer.
