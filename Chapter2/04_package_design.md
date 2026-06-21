# Package design

Пакет в Go — не «папка с классами», а **единство компиляции и API**. Идиоматичный дизайн пакетов делает код navigable и предотвращает циклические зависимости.

---

## Принципы

### 1. Один пакет — одна ответственность

```
internal/order/     — домен заказов: types, service, контракт Repository
internal/storage/memory/  — in-memory реализация хранилища
```

Не:

```
internal/utils/     — свалка функций
internal/helpers/   — «общее на будущее»
internal/common/    — антипаттерн в Go
```

### 2. Имена пакетов — короткие, lowercase, единственное число

```go
package order   // ✅
package orders  // ❌ множественное — не принято
package orderService // ❌ camelCase — не принято
```

Импорт читается как `order.NewService()` — пакет = префикс API.

### 3. Минимальный exported API

Экспортируйте только то, что нужно **вне** пакета. Остальное — unexported (`save`, `validate`).

### 4. Нет циклических импортов

Компилятор запрещает `A → B → A`. Решение: вынести интерфейс к потребителю или общие типы в третий пакет (редко).

---

## internal/ — encapsulation модуля

Код в `internal/` **нельзя импортировать** из других модулей:

```
study1/
└── Chapter2/examples/app/
    ├── cmd/api/           // main — composition root
    └── internal/order/    // только внутри этого модуля
```

Это сильнее, чем convention «не импортируйте» — **enforce на уровне компилятора**.

---

## pkg/ vs internal/

| Директория | Когда |
|------------|-------|
| `internal/` | Всё, что не для внешних потребителей (99% приложений) |
| `pkg/` | Библиотека, которую **намеренно** переиспользуют другие репозитории |

Для микросервиса `pkg/` часто **не нужен**. Не создавайте заранее.

---

## Где что лежит (внутри пакета)

Типичные файлы доменного пакета:

```
internal/order/
├── order.go       // типы домена
├── service.go     // бизнес-логика
├── repository.go  // интерфейс контракта (потребитель)
└── errors.go      // sentinel / typed errors
```

Не обязательно один файл — но **логическая группировка**, не «models.go на 2000 строк».

---

## init() — используйте редко

```go
func init() { ... }  // глобальные side effects — сложно тестировать
```

Идиоматично: явная инициализация в `main` или `NewXxx()` constructors.

---

## Сравнение с PHP

| PHP (Symfony) | Go |
|---------------|-----|
| `src/Entity/`, `src/Repository/` | `internal/order/` — домен + контракт вместе |
| PSR-4 namespace | import path = module + path |
| `App\Service\*` | `internal/*/service.go` |
| Autoconfigure | Explicit wiring в `main` |

---

## Anti-patterns

```
internal/models/    — анemic «мешок struct'ов» без поведения
internal/dto/       — часто лишний слой на старте
internal/interfaces/ — интерфейсы оторваны от потребителей
package util        — код без домена
```

---

## Пример

[examples/04_packages/](./examples/04_packages/) — два пакета, `internal/` boundary.

```bash
go run ./examples/04_packages/
```

Полный layout: [examples/app/](./examples/app/)
