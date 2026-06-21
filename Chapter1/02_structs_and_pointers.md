# Structs и Pointers

## Struct (структура)

**Struct** — составной тип данных. Ближайший аналог — объект без класса: данные + методы «привязаны» к типу, но нет наследования.

```go
type User struct {
    ID    int
    Email string
    // поля с маленькой буквы — unexported, видны только в пакете
    passwordHash string
}
```

### Создание

```go
// Литерал с именами полей (рекомендуется)
u := User{ID: 1, Email: "a@b.com"}

// Zero value — все поля получают значения по умолчанию
var u User  // ID=0, Email=""
```

### Методы

Метод — функция с **receiver** (получателем):

```go
func (u User) DisplayName() string {
    return u.Email
}

// Pointer receiver — может изменять struct и избегает копирования
func (u *User) SetEmail(email string) {
    u.Email = email
}
```

### Когда pointer receiver?

| Value receiver `(u User)` | Pointer receiver `(u *User)` |
|----------------------------|------------------------------|
| Маленькие, immutable типы | Нужно изменять struct |
| Не меняет данные | Большие struct (избежать копии) |
| | Consistency: если один метод pointer — все pointer |

**Правило для production:** если хоть один метод с pointer receiver — делайте все методы pointer receiver для этого типа.

---

## Pointers (указатели)

Указатель хранит **адрес** значения в памяти. Синтаксис: `*T` (тип указателя), `&x` (взять адрес), `*p` (разыменовать).

```go
x := 42
p := &x   // p имеет тип *int, указывает на x
*p = 100  // x теперь 100
```

### Зачем указатели в Go?

1. **Изменять значение** внутри функции (передача по ссылке).
2. **Избежать копирования** больших struct.
3. **Выразить отсутствие значения** — `nil` pointer = «нет объекта» (осторожно с nil!).

### Value vs Pointer semantics

```go
func incrementValue(n int)  { n++ }      // копия, оригинал не меняется
func incrementPointer(n *int) { *n++ }  // меняет оригинал

x := 5
incrementValue(x)   // x = 5
incrementPointer(&x) // x = 6
```

### nil

`nil` — zero value для указателей, slices, maps, interfaces, channels, functions.

```go
var u *User  // nil — указатель ни на что не указывает
// u.Email    // PANIC! разыменование nil
if u != nil {
    _ = u.Email  // безопасно
}
```

### Сравнение с PHP/JS

| Go | PHP | JS |
|----|-----|-----|
| Явные указатели | Объекты по ссылке | Объекты по ссылке |
| `*User`, `&u` | `$user` всегда ref-like | `const u = {}` — ref |
| Value types копируются | Scalars by value | Primitives by value |

В Go **нет** неявных ссылок на struct — передача `User` копирует struct. Для «как в PHP object» используйте `*User` или pointer receiver.

---

## Пример

Смотрите: [examples/02_structs/](./examples/02_structs/)

Запуск:

```bash
go run ./examples/02_structs/
```

Демонстрирует: struct literals, value vs pointer receivers, передачу в функции, nil safety.
