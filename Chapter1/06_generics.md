# Generics (обобщённое программирование)

Generics появились в **Go 1.18**. Философия Go: использовать **умеренно** — когда дублирование кода реально мешает, а не «generics ради generics».

## Type parameters

```go
func Max[T cmp.Ordered](a, b T) T {
    if a > b {
        return a
    }
    return b
}

m := Max(3, 7)        // T inferred as int
s := Max("a", "b")    // T inferred as string
```

### Constraints (ограничения)

Constraint — интерфейс, описывающий что type parameter должен уметь:

```go
// any — аналог interface{}, любой тип
func Print[T any](v T) {
    fmt.Println(v)
}

// Custom constraint
type Numeric interface {
    int | int64 | float64
}

func Sum[T Numeric](values []T) T {
    var total T
    for _, v := range values {
        total += v
    }
    return total
}
```

### Встроенные constraints (пакет `constraints` deprecated → используйте `cmp`, `slices`)

```go
import "cmp"

func Min[T cmp.Ordered](a, b T) T { ... }
```

`cmp.Ordered` — типы, поддерживающие `<`, `>`, `==`.

## Generic types

```go
type Stack[T any] struct {
    items []T
}

func (s *Stack[T]) Push(v T) {
    s.items = append(s.items, v)
}

func (s *Stack[T]) Pop() (T, bool) {
    var zero T
    if len(s.items) == 0 {
        return zero, false
    }
    v := s.items[len(s.items)-1]
    s.items = s.items[:len(s.items)-1]
    return v, true
}

// Использование
intStack := &Stack[int]{}
intStack.Push(42)
```

## Type inference

Компилятор выводит тип из аргументов:

```go
stack := &Stack[string{}}
stack.Push("hello")  // T = string
```

Явное указание когда inference не работает:

```go
var s Stack[int]  // zero value stack
```

## Generics vs interfaces

| Generics | Interfaces |
|----------|------------|
| Compile-time polymorphism | Runtime polymorphism |
| `Stack[int]`, `Stack[string]` — разные типы | `io.Reader` — любой Reader |
| Убирает дублирование алгоритмов | Скрывает реализацию |

**Идиома Go:** сначала interface. Generics — когда один алгоритм для многих **concrete types** и interface awkward (например, `Max`, `SliceContains`).

## Что НЕ делать

```go
// Плохо — generic repository на всё подряд без нужды
type Repository[T any] interface {
    Save(T) error
    Find(id int) (T, error)
}
```

Часто лучше явные типы или маленькие интерфейсы — проще читать и дебажить.

## Stdlib generics (Go 1.21+)

```go
import "slices"

slices.Contains([]int{1, 2, 3}, 2)  // true
slices.Sort(users)                   // если users []User with Ordered fields — custom Less
```

---

## Сравнение с TypeScript

| TS | Go |
|----|-----|
| `<T>` everywhere | Generics реже |
| Structural types | Constraints через interface |
| Erased at runtime (TS) | Monomorphization — отдельный код на каждый T |

---

## Пример

Смотрите: [examples/06_generics/](./examples/06_generics/)

Запуск:

```bash
go run ./examples/06_generics/
```

Демонстрирует: generic functions, constraints, generic types, inference.
