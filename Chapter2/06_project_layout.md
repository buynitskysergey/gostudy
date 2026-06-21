# Project layout

Стандартный layout Go-приложения — не framework convention, а **community standard**, проверенный годами.

---

## Базовая структура сервиса

```
app/
├── go.mod                    // отдельный module для примера (опционально)
├── cmd/
│   └── api/
│       └── main.go           // entrypoint: wiring, config, run
├── internal/
│   ├── order/                // домен
│   │   ├── order.go
│   │   ├── service.go
│   │   ├── repository.go
│   │   └── errors.go
│   └── storage/
│       └── memory/
│           └── repository.go
└── migrations/               // SQL (если есть) — часто в корне
```

### cmd/

Один подкаталог = один бинарник:

```
cmd/
├── api/main.go       // HTTP API
├── worker/main.go    // background jobs
└── migrate/main.go   // CLI migrations
```

`main` должен быть **тонким**: parse flags, wire deps, start server, graceful shutdown.

### internal/

Весь код приложения, который не экспортируется наружу. Компилятор блокирует внешние import.

### pkg/ (опционально)

Только если пишете **переиспользуемую библиотеку**. Для типичного REST-сервиса часто не нужен.

---

## Flat vs nested internal

```
// ✅ По домену/ bounded context
internal/order/
internal/user/
internal/payment/

// ❌ По техническому слою (PHP-style)
internal/models/
internal/services/
internal/repositories/
```

Go community предпочитает **vertical slices** (order, user), не horizontal layers.

---

## Где HTTP handlers?

```
internal/order/
├── service.go
└── handler.go      // если handlers простые

// или при росте:
internal/api/
└── handler/
    └── order.go    // HTTP layer, вызывает order.Service
```

Handler — тонкий слой: parse request → call service → map error to status.

---

## Config

```
// env + flags в main
// или internal/config/config.go — struct Config, Load() error
```

Не глобальный `config.Get()`. Передавайте `Config` в конструкторы или отдельные поля.

---

## Монорепо vs один module

Учебный проект `study1` — один module, главы в подпапках:

```
study1/
├── go.mod
├── Chapter1/
└── Chapter2/examples/app/   // layout внутри module
```

Production-репозиторий — обычно один module в корне с `cmd/`, `internal/`.

---

## Сравнение с PHP

| Laravel/Symfony | Go service |
|-----------------|------------|
| `public/index.php` | `cmd/api/main.go` |
| `src/Controller/` | `internal/api/handler/` |
| `src/Entity/` | `internal/order/order.go` |
| `config/services.yaml` | wiring в `main` |
| `vendor/` | module cache (не коммитят) |

---

## Эталон в этом репозитории

```
Chapter2/examples/app/
├── cmd/api/main.go
└── internal/
    ├── order/
    └── storage/memory/
```

Изучите файлы в этом порядке:

1. `internal/order/` — домен и контракты
2. `internal/storage/memory/` — инфраструктура
3. `cmd/api/main.go` — composition root

```bash
go run ./Chapter2/examples/app/cmd/api/
```

---

## Anti-patterns

```
src/           // не Go convention
app/Http/      // Laravel layout в Go repo
internal/internal/  // избыточная вложенность
```

---

## Дальше (этап 3)

HTTP router, middleware, structured logging, `testify` / stdlib tests, Docker, migrations — на базе этого layout.
