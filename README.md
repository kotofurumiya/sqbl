# sqbl

Simple SQL builder for Go.

> [!WARNING]
> This is a personal project. It is not intended for production use, and issues and pull requests are not being accepted at this time.

## Features

- ✅ Just builds SQL strings - no magic, no surprises, no validation, no ORM
- ✅ Zero deps
- ✅ Supports major SQL dialects
  - ✅ PostgreSQL
  - ✅ MySQL
  - ✅ SQLite

## Install

```shell
# Add the module to your project
go get github.com/kotofurumiya/sqbl
```

## Basic Usage

```go
package main

import (
  sqbl "github.com/kotofurumiya/sqbl/sqblpg"
  // sqbl "github.com/kotofurumiya/sqbl/sqblmysql"
  // sqbl "github.com/kotofurumiya/sqbl/sqblsqlite"
)

func main() {
  // SELECT "id", "name" FROM "users";
  sql1 := sqbl.From("users").Select("id", "name").ToSql()

  // Method order is flexible
  sql2 := sqbl.Select("id", "name").From("users").ToSql()

  // SELECT "id", "name" FROM "users" WHERE "active" = TRUE ORDER BY "name" ASC, "created_at" DESC LIMIT 20;
  sql3 := sqbl.From("users").
    Select("id", "name").
    Where(sqbl.Eq("active", true)).
    OrderBy(sqbl.Asc("name"), sqbl.Desc("created_at")).
    Limit(20).
    ToSql()

  // SELECT
  //   "u"."id", "u"."name", SUM("p"."amount") AS "total_spent"
  // FROM
  //   "users" AS "u"
  // INNER JOIN
  //   "purchases" AS "p" ON "u"."id" = "p"."user_id"
  // WHERE
  //   "p"."purchased_at" >= '2025-01-01'
  // GROUP BY
  //   "u"."id", "u"."name"
  // HAVING
  //   SUM("p"."amount") > 5000
  // ORDER BY
  //   "total_spent" DESC
  // LIMIT 10;
  sql4 := sqbl.From(sqbl.As("users", "u")).
    Select("u.id", "u.name", sqbl.As(sqbl.Fn("SUM", "p.amount"), "total_spent")).
    InnerJoin(sqbl.As("purchases", "p"), sqbl.Eq("u.id", "p.user_id")).
    Where(sqbl.Gte("p.purchased_at", sqbl.Raw("'2025-01-01'"))).
    GroupBy("u.id", "u.name").
    Having(sqbl.Gt(sqbl.Fn("SUM", "p.amount"), 5000)).
    OrderBy(sqbl.Desc("total_spent")).
    Limit(10).
    ToSql()

  // Subquery in FROM
  sub := sqbl.From("orders").Select("user_id").Where(sqbl.Eq("status", sqbl.Raw("'paid'")))
  sql5 := sqbl.From(sqbl.As(sub, "sub")).Select("sub.user_id").ToSql()

  // SELECT "id", "payload" FROM "jobs" WHERE "status" = 'pending' ORDER BY "id" ASC LIMIT 1 FOR UPDATE SKIP LOCKED;
  sql6 := sqbl.From("jobs").
    Select("id", "payload").
    Where(sqbl.Eq("status", sqbl.Raw("'pending'"))).
    OrderBy(sqbl.Asc("id")).
    Limit(1).
    ForUpdate().SkipLocked().
    ToSql()

  // SELECT "id", "name" FROM "users" WHERE "role" = 'admin' OR "role" = 'moderator';
  sql7 := sqbl.From("users").
    Select("id", "name").
    Where(sqbl.Or(
      sqbl.Eq("role", sqbl.Raw("'admin'")),
      sqbl.Eq("role", sqbl.Raw("'moderator'")),
    )).
    ToSql()
}
```

## Multiple Dialects

Build once with the root `sqbl` package, then render for each database with `ToSqlWithDialect`.

```go
import (
  "github.com/kotofurumiya/sqbl"
  "github.com/kotofurumiya/sqbl/dialect"
)

q := sqbl.From("users").Select("id").Where(sqbl.Eq("id", sqbl.P(1)))

q.ToSqlWithDialect(&dialect.PostgresDialect{}) // SELECT "id" FROM "users" WHERE "id" = $1
q.ToSqlWithDialect(&dialect.SqliteDialect{})   // SELECT "id" FROM "users" WHERE "id" = ?
```

- Builders from the root package have no dialect, so `ToSql()` panics on them.
- sqbl renders every clause you ask for. Branch in your own code for clauses that only some databases support, such as `ForUpdate()`.

## Identifiers, Values, and Raw SQL

A plain string passed where an expression is expected is always an identifier.
sqbl never guesses from the content of a string, so a value can never leak into the SQL by accident.

| You pass | Rendered as | Example |
|---|---|---|
| `"users.id"` | quoted identifier | `"users"."id"` |
| `sqbl.P(1)` | placeholder | `$1` |
| `42`, `true`, `nil` | literal | `42`, `TRUE`, `NULL` |
| `sqbl.Fn("NOW")` | function call | `NOW()` |
| `sqbl.Raw("n + 1")` | written as-is | `n + 1` |

```go
// "status" = $1 — pass values as parameters
sqbl.Eq("status", sqbl.P(1))

// "u"."id" = "o"."user_id" — a plain string is a column
sqbl.Eq("u.id", "o.user_id")

// "status" = 'pending' — write SQL text explicitly with Raw
sqbl.Eq("status", sqbl.Raw("'pending'"))
```

- `sqbl.Ident("...")` makes an identifier explicit. It behaves the same as a plain string.
- Each dot-separated part of an identifier is quoted separately. `*` and a trailing `.*` stay unquoted.
- Never build a `Raw` from untrusted input.
- Arguments typed as `string` that take a piece of SQL syntax are written as-is. Examples are column types and constraints in `CREATE TABLE`, `DoUpdate` assignments, and window frames.

## Docs

- Manual
  - [Examples](./examples/README.md)
    - Basics
      - [Fundamentals](./examples/basic_fundamentals/basic_fundamentals.go) — SELECT, WHERE, ORDER BY, LIMIT, GROUP BY, JOIN, subquery, parameters
      - [SELECT](./examples/basic_select/basic_select.go) — basic queries, WHERE, OR, ORDER BY, LIMIT
      - [INSERT](./examples/basic_insert/basic_insert.go) — VALUES, ON CONFLICT, RETURNING
      - [UPDATE](./examples/basic_update/basic_update.go) — SET, WHERE, RETURNING
      - [DELETE](./examples/basic_delete/basic_delete.go) — WHERE, RETURNING
      - [Subquery](./examples/basic_subquery/basic_subquery.go) — FROM subquery, WHERE subquery, IN, UNION
      - [Parameters](./examples/basic_params/basic_params.go) — indexed (`$1`), named (`:name`), positional (`?`)
      - [Window Functions](./examples/basic_window/basic_window.go) — ROW_NUMBER, RANK, DENSE_RANK, PARTITION BY
      - [CTE](./examples/basic_cte/basic_cte.go) — simple, chained, recursive
      - [DDL](./examples/basic_ddl/basic_ddl.go) — CREATE TABLE/INDEX, ALTER TABLE, DROP
    - Pro
      - [Job Queue](./examples/pro_job_queue/pro_job_queue.go) — FOR UPDATE SKIP LOCKED
      - [Audit Log](./examples/pro_audit_log/pro_audit_log.go) — upsert with ON CONFLICT DO UPDATE
      - [Ranking](./examples/pro_ranking/pro_ranking.go) — CTE + RANK() OVER window function
- Internal Docs
  - [Architecture](./docs/architecture.md)
  - [Custom Dialect](./docs/custom-dialect.md)

## Development

### Lint

```shell
golangci-lint run
```

### Test

```shell
go test ./...
```

### Benchmark

```shell
go test ./benchmark/ -bench=. -benchmem
```

## License

MIT License. See [LICENSE.md](LICENSE.md).
