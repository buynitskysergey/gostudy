# Go Study — от PHP/JS архитектора к production-ready Go

Учебный репозиторий для освоения Go без «пути новичка». Каждая глава — теория (`.md`) + runnable примеры.

## Структура

```
study1/
├── go.mod
├── README.md
├── Chapter1/          ← Этап 1: философия языка
│   ├── README.md
│   ├── 01_…06_*.md
│   └── examples/
└── Chapter2/          ← Этап 2: идиоматичный Go
    ├── README.md
    ├── 01_…07_*.md
    └── examples/
        ├── 01_contracts/ … 07_context/
        └── app/             ← мини-приложение (layout + DI + context)
```

## Прогресс

| Глава | Этап | Статус |
|-------|------|--------|
| [Chapter 1](./Chapter1/README.md) | Философия Go: packages, structs, interfaces, errors, slices, generics | ✅ |
| [Chapter 2](./Chapter2/README.md) | Идиоматичный Go: контракты, errors, layout, DI, context | 📖 текущий |

## Быстрый старт

```bash
cd c:\go\STUDY_1

# Глава 1 — любой пример
go run ./Chapter1/examples/03_interfaces/

# Глава 2 — мини-приложение (рекомендуется после чтения docs)
go run ./Chapter2/examples/app/cmd/api/
```

## Глава 2 — что внутри

| Тема | Документ |
|------|----------|
| Interfaces как контракты | [01_interfaces_as_contracts.md](./Chapter2/01_interfaces_as_contracts.md) |
| Ошибки как значения | [02_errors_as_values.md](./Chapter2/02_errors_as_values.md) |
| Wrapping errors | [03_wrapping_errors.md](./Chapter2/03_wrapping_errors.md) |
| Package design | [04_package_design.md](./Chapter2/04_package_design.md) |
| DI без контейнеров | [05_dependency_injection.md](./Chapter2/05_dependency_injection.md) |
| Project layout | [06_project_layout.md](./Chapter2/06_project_layout.md) |
| context.Context | [07_context.md](./Chapter2/07_context.md) |

## Как учиться

1. Читайте `.md` в порядке номеров.
2. Запускайте соответствующий пример в `examples/`.
3. В конце главы 2 пройдите `Chapter2/examples/app/` — все паттерны в одном месте.
4. Экспериментируйте: меняйте код, ломайте, чините.

## Module

Единый Go module `study1` (см. [go.mod](./go.mod)). Import path для кода глав:

- `study1/Chapter1/examples/...`
- `study1/Chapter2/examples/...`

## Следующий этап (глава 3)

HTTP-сервис: router, middleware, structured logging, тесты, graceful shutdown, конфигурация.
