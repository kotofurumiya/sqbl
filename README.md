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
	"fmt"
	"log"

	sqbl "github.com/kotofurumiya/sqbl/sqblpg"
	// sqbl "github.com/kotofurumiya/sqbl/sqblmysql"
	// sqbl "github.com/kotofurumiya/sqbl/sqblsqlite"
)

func main() {
	// ToSQL returns an error when the query holds input with no safe SQL form,
	// such as a value of an unsupported Go type.
	sql, err := sqbl.From("users").
		Select("id", "name").
		Where(sqbl.Eq("active", true)).
		OrderBy(sqbl.Asc("name"), sqbl.Desc("created_at")).
		Limit(20).
		ToSQL()
	if err != nil {
		log.Fatal(err)
	}
	// SELECT "id", "name" FROM "users" WHERE "active" = TRUE ORDER BY "name" ASC, "created_at" DESC LIMIT 20;
	fmt.Println(sql)

	// MustSQL panics instead of returning an error. Use it for queries built from constants.
	// SELECT "u"."id", "u"."name", SUM("p"."amount") AS "total_spent"
	// FROM "users" AS "u"
	// INNER JOIN "purchases" AS "p" ON "u"."id" = "p"."user_id"
	// WHERE "p"."purchased_at" >= $1
	// GROUP BY "u"."id", "u"."name"
	// HAVING SUM("p"."amount") > 5000
	// ORDER BY "total_spent" DESC
	// LIMIT 10;
	fmt.Println(sqbl.From(sqbl.As("users", "u")).
		Select("u.id", "u.name", sqbl.As(sqbl.Fn("SUM", "p.amount"), "total_spent")).
		InnerJoin(sqbl.As("purchases", "p"), sqbl.Eq("u.id", "p.user_id")).
		Where(sqbl.Gte("p.purchased_at", sqbl.P(1))).
		GroupBy("u.id", "u.name").
		Having(sqbl.Gt(sqbl.Fn("SUM", "p.amount"), 5000)).
		OrderBy(sqbl.Desc("total_spent")).
		Limit(10).
		MustSQL())

	// SELECT "id", "payload" FROM "jobs" WHERE "status" = 'pending' ORDER BY "id" ASC LIMIT 1 FOR UPDATE SKIP LOCKED;
	fmt.Println(sqbl.From("jobs").
		Select("id", "payload").
		Where(sqbl.Eq("status", sqbl.Raw("'pending'"))).
		OrderBy(sqbl.Asc("id")).
		Limit(1).
		ForUpdate().SkipLocked().
		MustSQL())

	// INSERT INTO "users" ("name", "email") VALUES ($1, $2) RETURNING "id";
	fmt.Println(sqbl.InsertInto("users").
		Columns("name", "email").
		Values(sqbl.P(1), sqbl.P(2)).
		Returning("id").
		MustSQL())
}
```

- `ToSQL()` returns `(string, error)`. It fails instead of producing broken or injectable SQL, for example when a value has an unsupported Go type.
- `MustSQL()` returns only the string and panics on error. Use it for queries built from constants.
- Builders are immutable values. Every method returns a new builder, so a shared base query can be extended safely.

## Multiple Dialects

Build once with the root `sqbl` package, then render for each database with `ToSQLWithDialect`.

```go
import (
	"github.com/kotofurumiya/sqbl"
	"github.com/kotofurumiya/sqbl/dialect"
)

q := sqbl.From("users").Select("id").Where(sqbl.Eq("id", sqbl.P(1)))

q.MustSQLWithDialect(&dialect.Postgres{}) // SELECT "id" FROM "users" WHERE "id" = $1
q.MustSQLWithDialect(&dialect.MySQL{})    // SELECT `id` FROM `users` WHERE `id` = ?
q.MustSQLWithDialect(&dialect.SQLite{})   // SELECT "id" FROM "users" WHERE "id" = ?1
```

- Builders from the root package have no dialect, so `ToSQL()` returns an error on them.
- sqbl renders every clause you ask for. Branch in your own code for clauses that only some databases support, such as `ForUpdate()`.

## Identifiers, Values, and Raw SQL

A plain string passed where an expression is expected is always an identifier.
sqbl never guesses from the content of a string, so a value never leaks into the SQL as text.

| You pass | Rendered as | Example |
|---|---|---|
| `"users.id"` | quoted identifier | `"users"."id"` |
| `sqbl.P(1)` | placeholder | `$1` |
| `42`, `1.5`, `true`, `nil` | literal | `42`, `1.5`, `TRUE`, `NULL` |
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

- Only numbers (including named numeric types), bools and `nil` become literals.
  - Any other Go value fails to render: a named string type, `time.Time`, `[]byte`, a pointer, and so on. Pass it with `P` and bind it as a query argument.
  - NaN and infinities fail as well, since SQL has no literal for them.
- Comparing with `nil` fails, because `= NULL` never matches. Use `IsNull` or `IsNotNull`.
- `sqbl.Ident("...")` makes an identifier explicit. It behaves the same as a plain string.
- Each dot-separated part of an identifier is quoted separately. `*` and a trailing `.*` stay unquoted.
- Never build a `Raw` from untrusted input.
- Arguments typed as `string` that take a piece of SQL syntax are written as-is. Examples are function names, `Cast` types, column types and constraints in `CREATE TABLE`, `DoUpdate` assignments, and window frames.

## Parameters

`P` writes a placeholder. The value itself goes to the database driver as a query argument.

| Call | PostgreSQL | MySQL | SQLite |
|---|---|---|---|
| `P(1)` | `$1` | `?` | `?1` |
| `P()` | `?` | `?` | `?` |
| `P(":status")` | `:status` | `:status` | `:status` |

- An index must be 1 or more. A name must be `:`, `@` or `$` followed by letters, digits and underscores. Anything else fails to render.
- MySQL has no numbered placeholders, so the number in `P(n)` is ignored there.
  - Arguments bind in the order the placeholders appear in the SQL.
  - An index used twice needs its argument passed twice.
  - Prefer `P()` with MySQL.
- `P` takes a placeholder, not a value. Never pass user input to `P`.

## Conditions

- Calling `Where` or `Having` again adds a condition with `AND` instead of replacing the earlier one, so a shared base filter cannot be lost.
- `And()` with no conditions renders `TRUE`, and `Or()` renders `FALSE`. `In` with no values renders `FALSE`, and `NotIn` renders `TRUE`. A filter list built at run time can be empty.

```go
base := sqbl.From("orders").Where(sqbl.Eq("tenant_id", sqbl.P(1)))

// SELECT * FROM "orders" WHERE "tenant_id" = $1 AND "status" = $2;
base.Where(sqbl.Eq("status", sqbl.P(2)))

// SELECT * FROM "orders" WHERE TRUE;
sqbl.From("orders").Where(sqbl.And(filters...)) // with no filters
```

## Set Operations

`Union`, `UnionAll`, `Intersect`, `IntersectAll`, `Except` and `ExceptAll` return a `CompoundBuilder`.
Its `OrderBy`, `Limit` and `Offset` apply to the combined result.

```go
active := sqbl.From("users").Select("id", "name")
archived := sqbl.From("archived_users").Select("id", "name")

// SELECT "id", "name" FROM "users" UNION SELECT "id", "name" FROM "archived_users" ORDER BY "name" LIMIT 10;
active.Union(archived).OrderBy("name").Limit(10)

// SELECT "id" FROM "a" EXCEPT (SELECT "id" FROM "b" UNION SELECT "id" FROM "c");
a.Except(b.Union(c))
```

- A compound can be used anywhere a subquery can: in `From` with `As`, in `In`, or as a CTE.
- Operands are parenthesized when grouping requires it. SQLite rejects parenthesized operands, so keep set operations flat (`a.Union(b).Except(c)`) there.

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
