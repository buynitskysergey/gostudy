# Packages и Modules

## Package (пакет)

**Пакет** — единица компиляции и организации кода в Go. Каждый `.go` файл принадлежит ровно одному пакету.

```
myproject/
├── go.mod
├── cmd/
│   └── api/
│       └── main.go          // package main — точка входа
└── internal/
    └── user/
        ├── user.go          // package user
        └── repository.go    // package user (тот же пакет!)
```

### Правила

1. **`package main`** — единственный пакет, из которого собирается исполняемый файл (`go build` / `go run`).
2. **Имя пакета** обычно совпадает с именем директории (не обязательно, но так принято).
3. **Экспорт** определяется регистром первой буквы:
   - `User` — экспортируется (public в PHP/JS)
   - `user` — только внутри пакета (private)
4. **Один пакет = одна директория**. Несколько файлов в одной папке — один пакет, они видят друг друга без import.

### Сравнение с PHP/JS

| Go | PHP | JS |
|----|-----|-----|
| `package user` | `namespace App\User` | ES module file |
| Экспорт по регистру | `public` / без модификатора | `export` |
| `import "study1/internal/user"` | `use App\User\User` | `import { User } from './user'` |

### Философия

- Пакеты **маленькие и сфокусированные** — один пакет = одна ответственность.
- Имена пакетов **короткие, lowercase**: `http`, `user`, `repo` — не `userManagementService`.
- Циклические импорты **запрещены** компилятором — это заставляет проектировать слои явно.

---

## Module (модуль)

**Module** — единица версионирования и зависимостей (аналог `composer.json` / `package.json`).

Файл `go.mod` в корне проекта:

```go
module study1          // import path для вашего кода

go 1.22                // минимальная версия Go
```

### Import path

Когда вы пишете:

```go
import "study1/examples/01_packages/greeter"
```

Go ищет код относительно **module path** из `go.mod`. Полный путь = `module` + путь от корня.

### Зависимости

```bash
go get github.com/google/uuid@v1.6.0   # добавить зависимость
go mod tidy                           # убрать неиспользуемые, добавить недостающие
```

В `go.mod` появится:

```
require github.com/google/uuid v1.6.0
```

### Версионирование

Go использует **semver** через git tags. Мажорная версия ≥2 — суффикс в import path:

```go
import "github.com/user/repo/v2/pkg"
```

### `internal/` — важный паттерн

Код в директории `internal/` **нельзя импортировать** извне модуля. Это встроенное правило языка — аналог «package-private» на уровне репозитория.

```
study1/
├── go.mod
├── cmd/api/main.go           // может импортировать internal/
└── internal/user/user.go     // недоступен для других модулей
```

---

## Пример

Смотрите: [examples/01_packages/](./examples/01_packages/)

- `greeter/greeter.go` — пакет с экспортируемой функцией `Greet`
- `main.go` — точка входа, импортирует `greeter`

Запуск:

```bash
go run ./examples/01_packages/
```

Ожидаемый вывод:

```
Hello, Architect!
Package greeter version: 1.0
```
