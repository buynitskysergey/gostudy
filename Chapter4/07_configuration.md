# Configuration

Конфиг в Go — обычная структура, заполненная из **flags** и **env**. Без глобального Config Service и без обязательного Viper.

---

## 12-factor: env + явный struct

```go
type Config struct {
    Addr            string
    ShutdownTimeout time.Duration
    MaxBodyBytes    int64
}

func Load() (Config, error) {
    cfg := Config{
        Addr:            envOr("HTTP_ADDR", ":8080"),
        ShutdownTimeout: 10 * time.Second,
        MaxBodyBytes:    1 << 20,
    }

    if v := os.Getenv("SHUTDOWN_TIMEOUT"); v != "" {
        d, err := time.ParseDuration(v)
        if err != nil {
            return Config{}, fmt.Errorf("SHUTDOWN_TIMEOUT: %w", err)
        }
        cfg.ShutdownTimeout = d
    }
    return cfg, cfg.Validate()
}

func (c Config) Validate() error {
    if c.Addr == "" {
        return errors.New("HTTP_ADDR is required")
    }
    if c.MaxBodyBytes < 1024 {
        return errors.New("max body too small")
    }
    return nil
}
```

**Fail fast:** невалидный конфиг → `main` завершается до Listen.

---

## flag + env

Частый паттерн: flag перекрывает default, env — для контейнеров:

```go
addr := flag.String("addr", envOr("HTTP_ADDR", ":8080"), "listen address")
flag.Parse()
```

| Источник | Когда |
|----------|-------|
| default в коде | local dev |
| env | Docker / k8s |
| flag | one-off override (`-addr=:9090`) |

Приоритет зафиксируйте явно (обычно: flag > env > default).

---

## Чего избегать

```go
// ❌ Глобальный var Cfg Config — сложно тестировать
// ❌ Читать os.Getenv внутри каждого handler
// ❌ Молчаливые defaults для секретов (пустой API key = ok)
```

Секреты — только env/secret store, **не** в git и не в OpenAPI examples с реальными значениями.

---

## Связь с DI (глава 2)

```go
func main() {
    cfg, err := config.Load()
    if err != nil {
        log.Fatal(err)
    }
    store := memory.New()
    h := api.NewHandler(store, cfg.MaxBodyBytes)
    srv := &http.Server{Addr: cfg.Addr, Handler: h}
    // ...
}
```

Config загружается **один раз** в composition root и передаётся вниз.

---

## Configuration vs PHP/JS

| Laravel / Nest | Go |
|----------------|-----|
| `config/app.php` + `.env` | `Config` struct + env |
| `config('app.url')` везде | поля struct / DI |
| `env()` в рантайме сервисов | только в `Load()` |

---

## Пример

[examples/07_configuration/](./examples/07_configuration/)

```bash
go run ./examples/07_configuration/
HTTP_ADDR=:9090 go run ./examples/07_configuration/
go run ./examples/07_configuration/ -addr=:7070
```

Load + Validate + приоритет flag/env/default.
