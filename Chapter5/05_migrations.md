# Миграции

Схема БД — часть релиза. Миграции: **версионированные SQL-файлы**, применяемые по порядку, с таблицей учёта версий.

---

## Минимальная модель

```
migrations/
  000001_init.up.sql
  000001_init.down.sql
  000002_accounts_version.up.sql
  000002_accounts_version.down.sql
```

Таблица `schema_migrations(version TEXT PRIMARY KEY, applied_at TIMESTAMP)`.

Алгоритм `up`:

1. Создать `schema_migrations`, если нет.
2. Прочитать файлы `*.up.sql`, отсортировать.
3. Для каждой версии, которой ещё нет — выполнить SQL в транзакции (если СУБД позволяет) и записать версию.

`down` — откат последней (осторожно в проде; чаще forward-only).

---

## Правила хороших миграций

| Правило | Почему |
|---------|--------|
| Только SQL (или код, генерирующий SQL) | воспроизводимо на CI/CD |
| Маленькие шаги | проще review и rollback |
| Expand/contract | zero-downtime: сначала добавить колонку, потом удалить старую |
| Не править уже применённые файлы | история иммутабельна; новая миграция |
| Отдельно от app boot (или явный флаг) | контроль, кто мигрирует |

---

## Expand / contract (zero-downtime)

1. **Expand:** `ADD COLUMN new_col` (nullable), писать в оба поля.
2. Заполнить / dual-write.
3. Читать из нового.
4. **Contract:** удалить старое поле в следующем релизе.

Никогда: `DROP COLUMN` + смена кода одним деплоем без совместимости.

---

## Инструменты

| Инструмент | Заметка |
|------------|---------|
| Свой мини-runner (как в примере) | прозрачно для учёбы |
| [golang-migrate](https://github.com/golang-migrate/migrate) | стандарт индустрии |
| [goose](https://github.com/pressly/goose) | up/down + Go-миграции |
| Atlas / dbmate | declarative / простой CLI |

В проде обычно CLI в CI или init-job в k8s, не «каждый pod мигрирует».

---

## Anti-patterns

```text
❌ Автомиграция на каждый старт 50 реплик без блокировки
❌ Хранить «текущую схему» только в ORM sync
❌ Редактировать 000001_*.sql после merge в main
```

---

## Пример

[examples/05_migrations/](./examples/05_migrations/)

```bash
go run ./examples/05_migrations/
```

Крошечный migrator + два SQL-шага на SQLite.
