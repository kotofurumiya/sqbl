# Architecture

sqbl is structured in four layers, each with a distinct responsibility:

- `sqblpg` / `sqblmysql` / `sqblsqlite` — DB-specific entry points; depend on `builder` and `dialect`
- `builder` — assembles and renders SQL statements; depends on `syntax` and `dialect`
- `syntax` — SQL fragment types; depends on `dialect`
- `dialect` — lexical differences (quoting, placeholders, booleans); no dependencies on other sqbl packages

---

## dialect

Absorbs DB-specific lexical differences. The dialect has no say in SQL structure — it only controls how individual tokens are rendered.

```go
type SqlDialect interface {
    Quote(buf *bytes.Buffer, s string)
    QuoteIdentifier(buf *bytes.Buffer, name string)
    PlaceholderPositional(buf *bytes.Buffer)
    PlaceholderIndexed(buf *bytes.Buffer, index int)
    Bool(buf *bytes.Buffer, b bool)
}
```

All methods write directly into a `*bytes.Buffer` to avoid intermediate string allocations. `Quote` handles a single segment; `QuoteIdentifier` splits on `.` and calls `Quote` on each part.

Built-in implementations: `PostgresDialect`, `MysqlDialect`, `SqliteDialect`, `SimpleDialect` (tests only).

---

## syntax

Defines the data types that represent pieces of SQL. Everything — columns, conditions, functions, parameters, subqueries — implements a single interface:

```go
type SqlFragment interface {
    AppendSQL(buf *bytes.Buffer, d dialect.SqlDialect)
}
```

Each fragment writes directly into the shared buffer instead of returning a string. There is no separate `Expression` type; condition types and structural types are both just `SqlFragment`.

Key types:

- `StringExpr` — a plain string; auto-quoted via `QuoteIdentifier` if it is a simple identifier, passed through as-is otherwise
- `Aliased` — `<source> AS <alias>`
- `SqlFn` — a function call such as `SUM("amount")` or `COUNT(*)`
- `WindowExpr` — a window function expression (`expr OVER (PARTITION BY ... ORDER BY ...)`)
- `Parameter` — a bind parameter placeholder (`?`, `$N`, `:name`, `@name`)
- Condition types: `ComparisonExpr`, `LogicalExpr`, `NotExpr`, `InExpr`, `NullExpr`, `BetweenExpr`

`ToFragment(v any)` converts arbitrary values to `SqlFragment`: `SqlFragment` values pass through, strings become `StringExpr`, anything else is formatted with `fmt.Sprint`.

---

## builder

Holds all query state and renders the final SQL string. All builders implement:

```go
type SqlBuilder interface {
    ToSql() string
    ToSqlWithDialect(d dialect.SqlDialect) string
}
```

`ToSql()` renders with a trailing semicolon using the builder's own dialect. `ToSqlWithDialect(d)` renders without a semicolon using the given dialect.

`SqlSelectBuilder` additionally implements `SqlFragment` via `AppendSQL`, which writes the query wrapped in parentheses. This is what makes it embeddable as a subquery in `From`, `Join`, and `In`.

State is accumulated independently of method call order and rendered at the end by a single `renderSQL` pass. `Select(...).From(...)` and `From(...).Select(...)` produce identical output.

Builders: `SqlSelectBuilder`, `SqlInsertBuilder`, `SqlUpdateBuilder`, `SqlDeleteBuilder`, `SqlCreateTableBuilder`, `SqlCreateIndexBuilder`, `SqlDropTableBuilder`, `SqlDropIndexBuilder`, `SqlAlterTableBuilder`.

---

## sqblpg / sqblmysql / sqblsqlite

Convenience packages. Each has two files:

- `sqblXXX.go` — entry point functions (`From`, `Select`, `InsertInto`, `Update`, `DeleteFrom`, `CreateTable`, etc.) that wire the appropriate dialect and return a typed builder
- `reexport.go` — re-exports `syntax` helpers (`Eq`, `And`, `In`, `Asc`, `P`, `Fn`, `Over`, etc.) so users only need to import one package; DB-specific functions are intentionally omitted where not applicable (e.g. `ILike` is only in `sqblpg`)
