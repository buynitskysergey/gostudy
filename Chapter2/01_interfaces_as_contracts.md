# Interfaces как контракты

В идиоматичном Go интерфейс — не «абстрактный класс на будущее», а **контракт между пакетами**: «мне нужен кто-то, кто умеет X».

## Главное правило: define interfaces where they are used

Интерфейс объявляет **потребитель**, не реализация.

```go
// internal/order/service.go — order знает, что ему нужно от хранилища
type Repository interface {
    Save(ctx context.Context, o Order) error
    FindByID(ctx context.Context, id string) (Order, error)
}

type Service struct {
    repo Repository // зависимость от контракта, не от postgres/memory
}
```

```go
// internal/storage/memory/repository.go — memory просто реализует методы
// Никакого "implements Repository" — компилятор проверит неявно
type Repository struct { ... }
func (r *Repository) Save(ctx context.Context, o order.Order) error { ... }
```

### Почему не в `storage`?

```go
// Плохо: storage определяет интерфейс, который все должны реализовать
package storage
type OrderRepository interface { ... } // «god interface»
```

Потребитель (`order`) не должен зависеть от деталей инфраструктуры. Интерфейс описывает **минимум**, нужный сервису.

### Сравнение с PHP/JS

| PHP/Symfony | Go |
|-------------|-----|
| Interface в `Contract/` на уровне домена | Interface рядом с **использованием** |
| DI autowire по type-hint | Ручная сборка в `main`, mock в тестах |
| Fat interface `UserRepositoryInterface` с 20 методами | 1–3 метода на интерфейс |

---

## Маленькие интерфейсы

Стандартная библиотека — эталон:

```go
type Reader interface { Read(p []byte) (n int, err error) }
type Writer interface { Write(p []byte) (n int, err error) }
```

Сервису заказов не нужен «полный репозиторий» — только Save и FindByID.

**Code smell:** интерфейс с 10+ методами или суффикс `Interface` в имени (`UserRepositoryInterface`).

---

## Accept interfaces, return structs

```go
// ✅ Принимаем интерфейс — гибкость и тестируемость
func NewService(repo Repository) *Service

// ✅ Возвращаем конкретный тип — вызывающий знает, что получил
func NewMemoryRepository() *memory.Repository
```

В `main` вы передаёте `*memory.Repository` туда, где ожидается `order.Repository` — это работает автоматически.

---

## Compile-time проверка реализации

```go
var _ order.Repository = (*Repository)(nil)
```

Если сигнатуры разойдутся — ошибка компиляции, не runtime.

Полезно в пакете **реализации**, особенно при рефакторинге.

---

## Интерфейсы на границах

| Граница | Интерфейс |
|---------|-----------|
| Service → Storage | `order.Repository` |
| Handler → Service | часто конкретный `*order.Service` или узкий `OrderCreator` |
| Test → Service | mock `Repository` (ручной struct с методами) |

Не интерфейсify всё подряд. Интерфейс там, где нужна **подмена реализации** (тест, другой backend).

---

## Anti-patterns

```go
// ❌ Пустой интерфейс — any, теряется типизация
func Process(v interface{}) { ... }

// ❌ Интерфейс в каждом struct «для тестов»
type UserService struct {
    repo UserRepository // OK
}
type userRepository interface { ... } // в том же файле — OK
// type Everything interface { ... 50 methods } — плохо

// ❌ Type assertion по всему коду вместо правильного интерфейса
if pg, ok := repo.(*postgres.Repo); ok { ... }
```

---

## Пример

[examples/01_contracts/](./examples/01_contracts/)

```bash
go run ./examples/01_contracts/
```

Мини-проект `app/` — полный контракт `order.Repository`: [examples/app/internal/order/repository.go](./examples/app/internal/order/repository.go)
