# database/sql

Пакет `database/sql` — стандартный слой доступа к SQL в Go. Это **не ORM**: вы пишете SQL, сканируете строки в переменные/structs.

```go
import (
    "database/sql"
    _ "modernc.org/sqlite" // side-effect: регистрация драйвера
)

db, err := sql.Open("sqlite", "file:app.db")
```

`sql.Open` **не обязан** сразу коннектиться — проверяйте `db.PingContext(ctx)`.

---

## Три базовых вызова

| Метод | Когда |
|-------|-------|
| `ExecContext` | INSERT/UPDATE/DELETE без строк результата |
| `QueryContext` | несколько строк |
| `QueryRowContext` | ровно одна строка (или `sql.ErrNoRows`) |

```go
res, err := db.ExecContext(ctx,
    `INSERT INTO accounts(owner, balance_cents) VALUES (?, ?)`,
    owner, balance,
)
id, _ := res.LastInsertId() // зависит от драйвера

row := db.QueryRowContext(ctx, `SELECT id, owner, balance_cents FROM accounts WHERE id = ?`, id)
var a Account
err = row.Scan(&a.ID, &a.Owner, &a.BalanceCents)
if errors.Is(err, sql.ErrNoRows) {
    return ErrNotFound
}
```

**Всегда** передавайте `context.Context` первым аргументом — таймауты и отмена HTTP-запроса доходят до драйвера.

---

## Scan и NULL

Go не мапит NULL в zero-value «молча»:

```go
var email sql.NullString
err := row.Scan(&id, &email)
if email.Valid {
    _ = email.String
}
```

Или указатели / `pgtype` (в pgx) — выберите один стиль в проекте.

---

## Prepared statements

`db.PrepareContext` полезен при **повторе одного SQL** в цикле. Для разовых запросов драйвер часто prepare'ит сам — не готовьте «на всякий случай» глобально на всё приложение.

```go
stmt, err := db.PrepareContext(ctx, `SELECT balance_cents FROM accounts WHERE id = ?`)
defer stmt.Close()
for _, id := range ids {
    var bal int64
    if err := stmt.QueryRowContext(ctx, id).Scan(&bal); err != nil {
        return err
    }
}
```

---

## sql.DB — это пул

`*sql.DB` безопасен для concurrent use и сам управляет соединениями. Не открывайте новый `sql.Open` на каждый запрос.

Подробнее: [04_connection_pooling.md](./04_connection_pooling.md).

---

## vs PHP/JS

| PDO / node-pg | database/sql |
|---------------|--------------|
| `$pdo->prepare` | `Query`/`Exec` (+ optional `Prepare`) |
| fetch into array | `Scan` в typed fields |
| один connection по умолчанию | пул внутри `*sql.DB` |
| Eloquent hydrate | вы сами собираете struct |

---

## Anti-patterns

```go
// ❌ Игнорировать sql.ErrNoRows — отличите «нет строки» от сбоя
// ❌ Склеивать SQL строками с user input — только плейсхолдеры
// ❌ sql.Open на каждый HTTP-запрос
// ❌ Забывать Close() у Rows (утечка соединений пула)
```

```go
rows, err := db.QueryContext(ctx, `SELECT id, owner FROM accounts`)
if err != nil {
    return err
}
defer rows.Close()

for rows.Next() {
    var a Account
    if err := rows.Scan(&a.ID, &a.Owner); err != nil {
        return err
    }
}
return rows.Err()
```

---

## Пример

[examples/01_database_sql/](./examples/01_database_sql/)

```bash
go run ./examples/01_database_sql/
```

CRUD аккаунтов на SQLite через `database/sql`.
