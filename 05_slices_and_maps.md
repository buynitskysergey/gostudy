# Slices и Maps

## Slice (срез)

**Slice** — динамический view над массивом. Это **основная** коллекция в Go (не `Array`, не `List`).

```go
// Slice literal
nums := []int{1, 2, 3}

// make с length и capacity
buf := make([]byte, 0, 256)  // len=0, cap=256
```

### Внутреннее устройство

Slice — struct из трёх полей: **pointer** (на backing array), **len**, **cap**.

```
nums := []int{1, 2, 3}
//  ptr → [1|2|3|?|?]  len=3, cap=3 (или больше после append)
```

### append

```go
nums = append(nums, 4, 5)  // может переаллоцировать backing array
```

Если `cap` исчерпан — новый массив в 2 раза больше (амортизированно O(1)).

### Slicing

```go
a := []int{0, 1, 2, 3, 4}
b := a[1:4]   // [1, 2, 3] — shared backing array с a!
b[0] = 99     // a[1] тоже 99
```

**Production:** будьте осторожны с subslices — неожиданные мутации. Копируйте явно:

```go
copy(dst, src)
// или
slices.Clone(a)  // Go 1.21+
```

### range

```go
for i, v := range nums {
    fmt.Println(i, v)
}

for _, v := range nums {  // индекс не нужен
    _ = v
}
```

`v` — **копия** элемента. Для struct в цикле используйте index или pointer:

```go
for i := range users {
    users[i].Active = true  // правильно
}
```

### nil slice vs empty slice

```go
var s []int       // nil — len=0, cap=0, == nil true
s = make([]int, 0) // empty — len=0, != nil (иногда важно для JSON: null vs [])
```

JSON: `nil` slice → `null`, empty slice → `[]`.

### Сравнение с PHP/JS

| Go slice | PHP array | JS Array |
|----------|-----------|----------|
| Typed | Mixed types | Mixed |
| Reference to array | Copy-on-write / ref | Reference |
| `append` returns new slice header | `$a[] = x` | `push` |

---

## Map (ассоциативный массив)

```go
counts := map[string]int{
    "go":   3,
    "php":  10,
}

counts["js"] = 5
v, ok := counts["rust"]  // v=0, ok=false — ключ не найден
if ok {
    _ = v
}
```

### Zero value и nil map

```go
var m map[string]int  // nil — нельзя писать! panic
m = make(map[string]int)  // теперь можно
m["key"] = 1
```

**Чтение** из nil map безопасно (вернёт zero value). **Запись** — panic.

### delete

```go
delete(counts, "php")
```

### Итерация

```go
for key, value := range counts {
    fmt.Println(key, value)
}
// Порядок случайный! Не полагайтесь на порядок ключей.
```

### Concurrent access

Map **не thread-safe**. Для goroutines: `sync.Map` или mutex + обычный map.

### Сравнение с PHP/JS

| Go | PHP | JS |
|----|-----|-----|
| `map[K]V` | associative array | `Map` / object |
| Нет `.keys()` chain | `array_keys()` | `Object.keys()` |
| Typed keys/values | Mixed | Mixed |

---

## Production tips

1. **Preallocate** slices когда знаете размер: `make([]T, 0, n)`.
2. **Не** передавайте большие slices по value без нужды — slice header маленький, но backing array shared.
3. Maps — для **lookup**, не для ordered data. Для порядка — slice of structs.
4. Используйте `slices` и `maps` пакеты (stdlib Go 1.21+) для Contains, Sort, Clone.

---

## Пример

Смотрите: [examples/05_collections/](./examples/05_collections/)

Запуск:

```bash
go run ./examples/05_collections/
```

Демонстрирует: append, slicing traps, range, maps, nil vs empty.
