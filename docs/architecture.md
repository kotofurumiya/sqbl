# Architecture

sqbl is structured in five layers, each with a distinct responsibility:

- `sqbl` (root) — dialect-free entry points; depends on `builder` and `syntax`
- `sqblpg` / `sqblmysql` / `sqblsqlite` — DB-specific entry points; depend on `builder`, `syntax` and `dialect`
- `builder` — assembles and renders SQL statements; depends on `syntax` and `dialect`
- `syntax` — SQL fragment types; depends on `dialect`
- `dialect` — lexical differences (quoting, placeholders, booleans); no dependencies on other sqbl packages

---

## dialect

Absorbs DB-specific lexical differences. The dialect has no say in SQL structure — it only controls how individual tokens are rendered.

```go
type Dialect interface {
    Quote(buf *bytes.Buffer, s string)
    QuoteIdentifier(buf *bytes.Buffer, name string)
    PlaceholderPositional(buf *bytes.Buffer)
    PlaceholderIndexed(buf *bytes.Buffer, index int)
    Bool(buf *bytes.Buffer, b bool)
}
```

All methods write directly into a `*bytes.Buffer` to avoid intermediate string allocations. `Quote` handles a single segment and escapes the quote character. `QuoteIdentifier` handles a dotted name; the built-in dialects pass it to the `QuoteIdentifierParts` helper, which quotes and escapes each part.

Built-in implementations: `Postgres`, `MySQL`, `SQLite`.

---

## syntax

Defines the data types that represent pieces of SQL. Everything — columns, conditions, functions, parameters, subqueries — implements a single interface:

```go
type SQLFragment interface {
    AppendSQL(buf *bytes.Buffer, d dialect.Dialect)
}
```

Each fragment writes directly into the shared buffer instead of returning a string. There is no separate `Expression` type; condition types and structural types are both just `SQLFragment`.

Key types:

- `Ident` — an identifier, quoted via `QuoteIdentifier`; a plain string is treated the same way
- `Raw` — SQL text written as-is
- `Aliased` — `<source> AS <alias>`, optionally followed by column aliases (`AS ids (id)`)
- `FnExpr` — a function call such as `SUM("amount")` or `COUNT(*)`
- `CastExpr` — a type conversion such as `CAST($1 AS bigint[])`
- `WindowExpr` — a window function expression (`expr OVER (PARTITION BY ... ORDER BY ...)`)
- `Parameter` — a bind parameter placeholder (`?`, `$N`, `:name`, `@name`)
- Condition types: `ComparisonExpr`, `LogicalExpr`, `NotExpr`, `InExpr`, `NullExpr`, `BetweenExpr`
- `Subquery` — a `SQLFragment` that renders with its own parentheses, such as a SELECT builder; `IN` uses it as-is, so `IN (subquery)` gets a single pair of parentheses

Every argument that accepts an expression (`any`) follows one rule, implemented by `AppendValue` and `ToFragment`:

- `SQLFragment` values render themselves.
- Strings are identifiers. The content of a string is never inspected to decide whether it is an expression.
- `bool` uses the dialect's boolean literal, and `nil` is `NULL`.
- Integer and floating-point kinds, including named numeric types, are numeric literals.
- Anything else fails to render: a named string type, `time.Time`, `[]byte`, a pointer, NaN, and so on. Such values must be bound with `P`.

`AppendValue` writes directly into the buffer. `ToFragment` returns a `SQLFragment` for builders that store expressions before rendering; a value it cannot render fails when the statement is rendered, not when the builder method is called.

`AppendSQL` must not keep the buffer it receives, because builders take it from a pool.

---

## builder

Holds all query state and renders the final SQL string. All builders implement:

```go
type Statement interface {
    ToSQL() (string, error)
    ToSQLWithDialect(d dialect.Dialect) (string, error)
    MustSQL() string
    MustSQLWithDialect(d dialect.Dialect) string
}
```

`ToSQL()` renders with a trailing semicolon using the builder's own dialect. `ToSQLWithDialect(d)` renders without a semicolon using the given dialect. The `Must` variants panic instead of returning an error.

Rendering fails, instead of producing broken SQL, when there is no dialect or when the statement holds input with no safe SQL form. Examples are an unsupported value type, an invalid placeholder, a join without an `ON` condition, or an incomplete upsert.

- `SQLFragment.AppendSQL` has no error result, so a fragment that hits such input aborts the render with an internal panic (`internal/sqlerr`).
- The builder's render entry point recovers that panic and returns it as an error. Any other panic keeps propagating.
- Custom fragments therefore need no error handling of their own.

`SelectBuilder` and `CompoundBuilder` also implement `syntax.Subquery` via `AppendSQL`, which writes the query wrapped in parentheses. This is what makes them embeddable as a subquery in `From`, `Join`, and `In`. Together they form the `Query` interface, which CTEs, set operations and lateral joins accept.

- `SelectBuilder` set-operation methods (`Union`, `Intersect`, `Except` and their `All` forms) return a `CompoundBuilder`.
- `CompoundBuilder` has its own `OrderBy`, `Limit` and `Offset`, which apply to the combined rows. It parenthesizes operands where SQL would otherwise group them differently.

State is accumulated independently of method call order and rendered at the end by a single `renderSQL` pass. `Select(...).From(...)` and `From(...).Select(...)` produce identical output.

Builders are values. Every method returns a modified copy and never shares mutable state with the builder it came from, so a base query can be extended in several directions.

Builders: `SelectBuilder`, `CompoundBuilder`, `InsertBuilder`, `UpdateBuilder`, `DeleteBuilder`, `CreateTableBuilder`, `CreateIndexBuilder`, `DropTableBuilder`, `DropIndexBuilder`, `AlterTableBuilder`, `CreateDatabaseBuilder`, `DropDatabaseBuilder`.

---

## sqbl (root)

Dialect-free entry points for code that targets several databases.

- `sqbl.go` — the same entry point functions as the dialect packages, returning builders with no dialect. Render them with `ToSQLWithDialect`.
- `reexport.go` — re-exports every `syntax` helper, including DB-specific ones such as `ILike`, since the root package is not tied to a database.

---

## sqblpg / sqblmysql / sqblsqlite

Convenience packages. Each has two files:

- `sqblXXX.go` — entry point functions (`From`, `Select`, `InsertInto`, `Update`, `DeleteFrom`, `CreateTable`, etc.) that wire the appropriate dialect and return a typed builder
- `reexport.go` — re-exports `syntax` helpers (`Eq`, `And`, `In`, `Asc`, `P`, `Fn`, `Over`, etc.) so users only need to import one package; DB-specific functions are intentionally omitted where not applicable (e.g. `ILike` is only in `sqblpg`, and `sqblsqlite` has no `AsColumns`, `CreateDatabase`, or `DropDatabase`)
