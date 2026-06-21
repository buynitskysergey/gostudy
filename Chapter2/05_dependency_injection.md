# Dependency Injection без контейнеров

В Go **нет** стандартного DI-контейнера — и это intentional. Зависимости передаются **явно** через constructors; сборка — в `main` (composition root).

---

## Constructor injection

```go
type Service struct {
    repo Repository
}

func NewService(repo Repository) *Service {
    return &Service{repo: repo}
}
```

- Зависимости видны в сигнатуре `NewService`
- `Service` не знает, откуда взялся `repo` — memory, postgres, mock
- Нет reflection, нет magic autowire

### Сравнение с PHP/Symfony

| Symfony | Go |
|---------|-----|
| `services.yaml`, autowire | `main.go`, ручная сборка |
| Constructor DI через контейнер | Constructor DI вручную |
| ServiceLocator anti-pattern | Явные параметры |

---

## Composition root — только main (или cmd/)

```go
func main() {
    repo := memory.New()
    svc := order.NewService(repo)
    // передать svc в HTTP handler, CLI, worker
}
```

**Правило:** wiring (кто от кого зависит) — в одном месте на entrypoint. Не размазывать `New()` по пакетам.

---

## Интерфейсы для подмены

```go
// production
repo := memory.New()
svc := order.NewService(repo)

// test
svc := order.NewService(&fakeRepo{...})
```

Fake — обычный struct с методами, без mock-фреймворка (на старте).

---

## Functional options (когда параметров много)

```go
type Option func(*Service)

func WithLogger(l Logger) Option {
    return func(s *Service) { s.logger = l }
}

func NewService(repo Repository, opts ...Option) *Service {
    s := &Service{repo: repo}
    for _, opt := range opts {
        opt(s)
    }
    return s
}
```

Используйте когда optional dependencies > 2–3. Не усложняйте простые сервисы.

---

## Чего избегать

```go
// ❌ Глобальная переменная
var DB *sql.DB

// ❌ Singleton getInstance()
func GetUserService() *Service { ... }

// ❌ Service locator
type Container struct {
    Get(name string) any
}

// ❌ DI-фреймворк на старте обучения — скрывает wiring
```

---

## Тестирование без контейнера

```go
func TestCreateOrder(t *testing.T) {
    repo := memory.New()
    svc := order.NewService(repo)

    ctx := context.Background()
    o, err := svc.Create(ctx, "ORD-1", 100)
    if err != nil {
        t.Fatal(err)
    }
    // ...
}
```

Тот же `NewService`, другая реализация `Repository`.

---

## Пример

Полная сборка в [examples/app/cmd/api/main.go](./examples/app/cmd/api/main.go):

```bash
go run ./examples/app/cmd/api/
```

Проследите цепочку: `memory.New()` → `order.NewService(repo)` → вызовы с `context`.
