# Go: философия языка (этап 1)

Материалы для разработчика с опытом PHP/JS, который хочет сразу думать «по-Go», а не проходить путь новичка.

## Ключевые идеи Go

| Идея | Что это значит на практике |
|------|---------------------------|
| **Простота** | Мало синтаксиса, мало «магии». Явность важнее краткости. |
| **Явность** | Ошибки возвращаются явно, не через exceptions. Типы видны сразу. |
| **Композиция** | Нет классов и наследования. Поведение собирается из маленьких типов и интерфейсов. |
| **Интерфейсы неявные** | Тип реализует интерфейс автоматически — достаточно иметь нужные методы. |
| **Маленькие пакеты** | Код организуется по **пакетам** (единицам компиляции), а не по классам. |
| **Один способ сделать вещь** | Обычно один идиоматичный путь: `if err != nil`, `range`, `defer`. |

### Сравнение с PHP/JS (кратко)

| Концепция | PHP/JS | Go |
|-----------|--------|-----|
| Зависимости | Composer / npm | **Modules** (`go.mod`) |
| ООП | Классы, extends | **Structs** + **composition** |
| Полиморфизм | Interfaces (TS), abstract classes | **Interfaces** (неявные) |
| Ошибки | try/catch, exceptions | `(result, error)` — значение второго класса |
| Массивы | Array + mutable | **Slice** — view над массивом, ссылочная семантика |
| null | null / undefined | **nil** только для указателей, slices, maps, interfaces, channels |
| Generics | TS generics, PHP без них (до 8+) | Generics с Go 1.18+, используют умеренно |

## Структура материалов

| Файл | Темы |
|------|------|
| [01_packages_and_modules.md](./01_packages_and_modules.md) | packages, modules |
| [02_structs_and_pointers.md](./02_structs_and_pointers.md) | structs, pointers |
| [03_interfaces_and_composition.md](./03_interfaces_and_composition.md) | interfaces, composition |
| [04_error_handling_and_defer.md](./04_error_handling_and_defer.md) | error handling, defer |
| [05_slices_and_maps.md](./05_slices_and_maps.md) | slices, maps |
| [06_generics.md](./06_generics.md) | generics |

## Запуск примеров

```bash
cd c:\go\STUDY_1

# Каждый пример — отдельная программа
go run ./examples/01_packages/
go run ./examples/02_structs/
go run ./examples/03_interfaces/
go run ./examples/04_errors/
go run ./examples/05_collections/
go run ./examples/06_generics/
```

## Как читать

1. Прочитайте `.md` файл — там объяснение и «почему так в Go».
2. Откройте соответствующий каталог в `examples/` и запустите код.
3. Меняйте примеры — Go компилируется быстро, экспериментируйте.

## Следующий этап (когда будете готовы)

После философии — production-ready паттерны: `context`, HTTP-серверы, тестирование, логирование, конфигурация, graceful shutdown, layout проекта (`cmd/`, `internal/`).
