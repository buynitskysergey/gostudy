# Channels

**Channel** — typed conduit для отправки и получения значений между goroutines. Основной способ **communication** в Go.

```go
ch := make(chan int)       // unbuffered
ch := make(chan int, 10)   // buffered, ёмкость 10
```

---

## Send / Receive

```go
ch <- 42      // send (блокирует, пока кто-то не receive)
v := <-ch     // receive
v, ok := <-ch // ok=false если channel closed
close(ch)     // закрывает channel (только sender!)
```

**Правило:** только **отправитель** закрывает channel. Receive на closed channel — zero value + `ok=false`.

---

## Unbuffered vs Buffered

| | Unbuffered | Buffered |
|---|------------|----------|
| Send блокирует | Пока receive | Пока buffer не полон |
| Синхронизация | Handoff «лицом к лицу» | Async до заполнения buffer |
| Использование | Сигнал, sync | Producer faster than consumer |

```go
ch := make(chan int)     // sync point
ch := make(chan int, 3)  // 3 send без блокировки
```

---

## Direction (направление)

Ограничение типа для API:

```go
func producer(out chan<- int)   { out <- 1 }  // только send
func consumer(in <-chan int)    { <-in }      // только receive
```

Компилятор не даст send в `<-chan`.

---

## Channel как сигнал

```go
done := make(chan struct{})
go func() {
    work()
    close(done)  // broadcast: все receive разблокируются
}()
<-done
```

`struct{}` — zero-size тип, идеален для сигналов без данных.

---

## Range over channel

```go
for v := range ch {
    fmt.Println(v)
}
// завершится когда channel closed и drained
```

---

## Nil channel — блокировка навсегда

```go
var ch chan int  // nil
<-ch             // deadlock в runtime
```

Используется в `select` для **отключения** case (см. глава 03).

---

## Anti-patterns

```go
// ❌ Send на closed channel — panic
close(ch)
ch <- 1

// ❌ Close со стороны receiver
// ❌ Leak: goroutine blocked on send, никто не receive
// ❌ Shared map без sync — race condition
```

---

## Channels vs shared memory

```go
// Идиоматично: передать ownership через channel
results := make(chan Result)
go func() { results <- compute() }()
r := <-results

// Менее идиоматично (но иногда OK): mutex + shared struct
```

---

## Пример

[examples/02_channels/](./examples/02_channels/)

```bash
go run ./examples/02_channels/
```

Unbuffered handoff, buffered pipeline, close + range.
