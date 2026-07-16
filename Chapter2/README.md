# Глава 2 — Идиоматичный Go

Самый важный этап. Писать синтаксически корректный Go может почти каждый. **Идиоматичный** Go — когда код читается как стандартная библиотека: явный, предсказуемый, тестируемый.

## Чем идиоматичный Go отличается от «просто Go»

| «Просто Go» (PHP/JS-мышление) | Идиоматичный Go |
|-------------------------------|-----------------|
| Большие интерфейсы «на будущее» | Маленькие интерфейсы у **потребителя** |
| `panic` / exceptions для бизнес-ошибок | `error` как значение, явная обработка |
| DI-контейнер, service locator | **Constructor injection** в `main` |
| `utils`, `helpers`, `common` | Пакеты по **домену** и ответственности |
| Глобальные переменные для DB/config | Зависимости передаются явно |
| Context «где-нибудь потом» | `context.Context` — **первый** аргумент в I/O |

## Материалы главы

| Файл | Тема |
|------|------|
| [01_interfaces_as_contracts.md](./01_interfaces_as_contracts.md) | interfaces как контракты |
| [02_errors_as_values.md](./02_errors_as_values.md) | ошибки как значения |
| [03_wrapping_errors.md](./03_wrapping_errors.md) | wrapping errors |
| [04_package_design.md](./04_package_design.md) | package design |
| [05_dependency_injection.md](./05_dependency_injection.md) | DI без контейнеров |
| [06_project_layout.md](./06_project_layout.md) | project layout |
| [07_context.md](./07_context.md) | context.Context |

## Примеры

### Отдельные темы

```bash
cd c:\go\STUDY_1\Chapter2

go run ./examples/01_contracts/
go run ./examples/02_errors/
go run ./examples/03_wrapping/
go run ./examples/04_packages/
go run ./examples/07_context/
```

### Мини-приложение (все паттерны вместе)

Идиоматичный layout + DI + errors + context:

```bash
go run ./examples/app/cmd/api/
```

Структура приложения — эталон для [06_project_layout.md](./06_project_layout.md).

## Рекомендуемый порядок

1. **Interfaces as contracts** — фундамент для всего остального.
2. **Errors + wrapping** — как строить цепочки ошибок в сервисах.
3. **Package design + layout** — куда класть код.
4. **DI** — как собирать зависимости в `main`.
5. **Context** — сквозная отмена и таймауты.
6. **app/** — пройдитесь по файлам и сопоставьте с документацией.

## Предыдущий этап

[Глава 1 — Философия Go](../Chapter1/README.md)

## Следующие этапы

- [Глава 3 — Concurrency](../Chapter3/README.md)
- [Глава 4 — Backend на stdlib](../Chapter4/README.md)
