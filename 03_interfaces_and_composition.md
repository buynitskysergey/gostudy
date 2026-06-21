# Interfaces и Composition

## Interface (интерфейс)

**Интерфейс** — набор методов. Тип **неявно** реализует интерфейс, если у него есть все нужные методы. Ключевое слово `implements` отсутствует.

```go
type Reader interface {
    Read(p []byte) (n int, err error)
}
```

Любой тип с методом `Read([]byte) (int, error)` — это `Reader`. `*os.File`, `bytes.Buffer`, ваш тип — без объявления.

### Философия: «Accept interfaces, return structs»

```go
// Хорошо — функция принимает интерфейс (гибкость, тестируемость)
func Process(r io.Reader) error { ... }

// Хорошо — возвращает конкретный тип (ясность для вызывающего)
func NewUserService(repo *PostgresUserRepo) *UserService { ... }
```

### Маленькие интерфейсы

Стандартная библиотека полна **однометодных** интерфейсов:

```go
type Stringer interface { String() string }
type error  interface { Error() string }  // встроенный!
```

Интерфейс на 10 методов — code smell. Лучше несколько маленьких.

### Interface value

Интерфейс в runtime — пара `(type, value)`:

```go
var w io.Writer        // nil interface
var f *os.File         // nil pointer
w = f                  // w != nil! (type=*os.File, value=nil)
                       // частая ловушка для новичков
```

Проверяйте конкретный тип через **type assertion**:

```go
s, ok := w.(io.StringWriter)
if ok {
    s.WriteString("hello")
}
```

Или **type switch**:

```go
switch v := w.(type) {
case *bytes.Buffer:
    fmt.Println("buffer", v.Len())
case *os.File:
    fmt.Println("file")
default:
    fmt.Printf("unknown %T\n", v)
}
```

### Сравнение с PHP/JS/TS

| Go | TypeScript | PHP |
|----|------------|-----|
| Неявная реализация | `implements` явно | Interface + class implements |
| Duck typing на compile time | Structural typing | Runtime duck typing |
| Интерфейсы маленькие | Любого размера | Любого размера |

---

## Composition (композиция вместо наследования)

Go **не имеет** наследования, `extends`, `super`, virtual methods.

Вместо этого — **embedding** (встраивание) и **композиция**:

```go
type Logger struct{}

func (l Logger) Log(msg string) { fmt.Println(msg) }

type Service struct {
    Logger              // embedded — методы Logger «продвигаются» на Service
    repo UserRepository // обычное поле — композиция
}

func (s Service) DoWork() {
    s.Log("starting")   // вызов Logger.Log без явного поля
    s.repo.Save(...)
}
```

### Embedding vs Inheritance

| Наследование (PHP) | Embedding (Go) |
|-------------------|----------------|
| is-a: `Dog extends Animal` | has-a + promotion: `Service has Logger` |
| Override virtual methods | Нет override — явная делегация или новый метод |
| Иерархии классов | Плоские struct + интерфейсы |

**Production-паттерн:** struct содержит зависимости (repo, logger, clock), поведение описывается интерфейсами для тестов.

```go
type UserService struct {
    repo   UserRepository  // interface — mock в тестах
    logger Logger          // interface
}

type UserRepository interface {
    Save(ctx context.Context, u User) error
    FindByID(ctx context.Context, id int) (User, error)
}
```

Это **dependency injection** без фреймворка — просто передача интерфейсов в constructor:

```go
func NewUserService(repo UserRepository, logger Logger) *UserService {
    return &UserService{repo: repo, logger: logger}
}
```

---

## Пример

Смотрите: [examples/03_interfaces/](./examples/03_interfaces/)

Запуск:

```bash
go run ./examples/03_interfaces/
```

Демонстрирует: неявные интерфейсы, embedding, mock через interface, «accept interfaces, return structs».
