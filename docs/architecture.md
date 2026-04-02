# Architecture

sqbl is structured in four layers, each with a distinct responsibility:

- `sqblpg` / `sqblmysql` / `sqblsqlite` - DB-specific entry points (convenience packages)
  - depend on `builder` and `dialect`
- `builder` - assembles SQL statements (SELECT / INSERT / UPDATE / DELETE / CREATE / DROP / ALTER) and renders them as SQL strings
  - depends on `syntax` and `dialect`
- `syntax` - defines data types that represent fragments of SQL
  - depends on `dialect`
- `dialect` - abstracts DB-specific lexical differences (quoting, placeholders, booleans)
  - no dependencies on other sqbl packages

---

## dialect package

Defines the `SqlDialect` interface and provides built-in implementations.

Responsibility: absorb DB-specific lexical differences. The dialect has no say in SQL structure (clause order, keywords like SELECT/FROM/WHERE). It only controls how individual tokens are rendered.

### SqlDialect interface

- `Quote(str string) string`
  - Wraps a single identifier segment in the dialect's quote character
  - Escapes any existing quote characters within the segment
  - Example (PostgreSQL): `users` → `"users"`, `my"table` → `"my""table"`
  - Example (MySQL): `users` → `` `users` ``, `` my`table `` → `` `my``table` ``
- `QuoteIdentifier(name string) string`
  - Splits a dot-separated identifier and quotes each segment with `Quote`
  - Example (PostgreSQL): `public.users` → `"public"."users"`
  - Example (MySQL): `mydb.users` → `` `mydb`.`users` ``
- `PlaceholderPositional() string`
  - Returns the positional bind parameter token (no index)
  - All built-in dialects return `?`
- `PlaceholderIndexed(index int) string`
  - Returns the indexed bind parameter token for the given 1-based index
  - PostgreSQL: `$1`, `$2`, ...
  - MySQL / SQLite: always `?` (index is ignored)
- `Bool(b bool) string`
  - Returns the SQL representation of a boolean value
  - PostgreSQL / MySQL: `TRUE` / `FALSE`
  - SQLite: `1` / `0`

### Built-in implementations

- `PostgresDialect` — double-quote identifiers, `$N` indexed placeholders, `TRUE`/`FALSE`
- `MysqlDialect` — backtick identifiers, `?` placeholders, `TRUE`/`FALSE`
- `SqliteDialect` — double-quote identifiers, `?` placeholders, `1`/`0`
- `SimpleDialect` — used directly in test files; no quoting, `?` positional / `?N` indexed placeholders, `TRUE`/`FALSE`

---

## syntax package

Defines the data types that represent pieces of SQL.

### SqlFragment interface

Everything in sqbl — columns, conditions, functions, parameters, subqueries — implements a single interface:

```go
type SqlFragment interface {
    ToSqlWithDialect(d dialect.SqlDialect) string
}
```

There is no separate `Expression` interface. Condition types (`ComparisonExpr`, `LogicalExpr`, etc.) implement `SqlFragment` directly, just like structural fragments (`StringSource`, `Aliased`, etc.). The distinction is only semantic: condition types are typically passed to `Where()` and `Having()`, while structural types are typically passed to `Select()` and `From()`. Both are just `SqlFragment`.

### Fragment types

- `StringSource` — wraps a raw string
  - If the string is a simple identifier (no `(`, `'`, `*`, or spaces), it is auto-quoted via `QuoteIdentifier`
  - Otherwise it is returned as-is (e.g. `SUM(amount)`, `'2025-01-01'`)
- `*Aliased` — renders as `<source> AS <alias>` where the alias is always quoted with `Quote`
- `Order` — renders as `<column> ASC` or `<column> DESC`; the column is always quoted
- `*SqlSelectBuilder` — renders itself as `(<subquery>)` with parentheses, enabling subquery embedding
- `Parameter` — renders a bind parameter placeholder (see [Parameter and P()](#parameter-and-p) below)

`ToFragment(v any)` is the central conversion helper used throughout `builder`:
- If `v` implements `SqlFragment`, it is used directly
- If `v` is a `string`, it is wrapped in `StringSource`
- Otherwise, `fmt.Sprint(v)` is wrapped in `StringSource`

### Condition types

Condition types implement `SqlFragment` and are used in `Where()`, `Having()`, and JOIN `ON` clauses.

- `ComparisonExpr` — binary comparison: `col = value`, `col >= value`, etc.
  - The left side accepts `string` (auto-quoted if simple identifier) or any `SqlFragment` (e.g. `*SqlFn`)
  - The right side is handled by type:
    - `SqlFragment` (including `Parameter`) → rendered via `ToSqlWithDialect`
    - `bool` → rendered via `dialect.Bool`
    - `string` containing `'` or `"` → used as-is (SQL literal)
    - plain `string` (simple identifier) → quoted via `QuoteIdentifier`
    - other types → formatted with `fmt.Sprintf("%v", v)`
- `LogicalExpr` — joins multiple `SqlFragment` values with `AND` or `OR`
  - Nested `LogicalExpr` are automatically wrapped in parentheses to preserve evaluation order
  - Example: `And(Eq("a", 1), Or(Eq("b", 2), Eq("c", 3)))` → `"a" = 1 AND ("b" = 2 OR "c" = 3)`
- `NotExpr` — wraps a `SqlFragment` in `NOT (...)`
- `InExpr` — renders `col IN (...)` or `col NOT IN (...)`
  - Values support `SqlFragment` (including `Parameter`), `bool`, `string`, and other types via `fmt.Sprintf`
- `NullExpr` — renders `col IS NULL` or `col IS NOT NULL`
- `BetweenExpr` — renders `col BETWEEN low AND high`
  - Bounds support `SqlFragment` (including `Parameter`), `bool`, `string`, and other types via `fmt.Sprintf`

### SqlFn

`*SqlFn` represents a SQL function call such as `SUM("amount")` or `COUNT(*)`. It implements `SqlFragment` and can be used anywhere a column expression or condition operand is accepted.

```go
syntax.Fn("SUM", "amount")           // → SUM("amount")
syntax.Fn("COUNT", "*")              // → COUNT(*)
syntax.Fn("COALESCE", "x", 0)        // → COALESCE("x", 0)
syntax.Fn("NOW")                     // → NOW()
syntax.Fn("SUM", Fn("ABS", "val"))   // → SUM(ABS("val"))
```

Arguments follow the same rendering rules as `ComparisonExpr.Right`:
- `SqlFragment` → `ToSqlWithDialect`
- `string` → quoted if simple identifier, otherwise as-is
- other types → `fmt.Sprintf`

Because `*SqlFn` is a `SqlFragment`, it can be used on the left side of comparisons:

```go
syntax.Gt(syntax.Fn("SUM", "amount"), 5000)  // → SUM("amount") > 5000
syntax.As(syntax.Fn("SUM", "amount"), "total") // → SUM("amount") AS "total"
```

### WindowExpr

`*WindowExpr` represents a window function expression. It implements `SqlFragment`.

```go
syntax.Over(syntax.Fn("ROW_NUMBER")).
    PartitionBy("dept").
    OrderBy("salary")
// → ROW_NUMBER() OVER (PARTITION BY "dept" ORDER BY "salary")

syntax.Over(syntax.Fn("SUM", "amount")).
    PartitionBy("user_id").
    OrderBy("created_at").
    Rows("BETWEEN UNBOUNDED PRECEDING AND CURRENT ROW")
// → SUM("amount") OVER (PARTITION BY "user_id" ORDER BY "created_at" ROWS BETWEEN UNBOUNDED PRECEDING AND CURRENT ROW)
```

### Parameter and P()

`Parameter` represents a SQL bind parameter placeholder. It implements `SqlFragment` and renders without quoting.

- `P()` — positional: renders via `dialect.PlaceholderPositional()` → `?` (all dialects)
- `P(1)` — indexed: renders via `dialect.PlaceholderIndexed(1)` → `$1` (PostgreSQL), `?` (MySQL/SQLite)
- `P(":status")` — named: renders the string as-is → `:status` (all dialects)
- `P("@name")` — named: renders the string as-is → `@name` (all dialects)

### JoinClause struct

Holds the data for a single JOIN clause:
- `Kind string` — one of `"INNER"`, `"LEFT"`, `"RIGHT"`, `"FULL OUTER"`, `"CROSS"`
- `Table SqlFragment` — the joined table or subquery
- `On SqlFragment` — the ON condition (`nil` for CROSS JOIN)
- `Lateral bool` — when true, inserts the `LATERAL` keyword (PostgreSQL/MySQL)

---

## builder package

Holds all query state and renders the final SQL string.

### SqlBuilder interface

All statement builders implement `SqlBuilder`:

```go
type SqlBuilder interface {
    ToSql() string
    ToSqlWithDialect(d dialect.SqlDialect) string
}
```

- `ToSql()` — renders the statement with a trailing semicolon using the builder's own dialect
- `ToSqlWithDialect(d)` — renders the statement using the given dialect, without a trailing semicolon

### SqlSelectBuilder struct

Fields:
- `dialect` — the `SqlDialect` used for rendering
- `ctes []cteClause` — WITH clauses (each has a name, a `*SqlSelectBuilder`, and a `Recursive bool` flag)
- `distinct bool` — whether SELECT DISTINCT is set
- `distinctOn []string` — DISTINCT ON columns (PostgreSQL only; ignored for other dialects)
- `columns []SqlFragment` — SELECT columns
- `from SqlFragment` — the FROM clause target
- `joins []JoinClause` — JOIN clauses
- `where SqlFragment` — WHERE condition
- `groups []SqlFragment` — GROUP BY columns
- `having SqlFragment` — HAVING condition
- `orders []SqlFragment` — ORDER BY expressions
- `limit *int` — LIMIT value; `nil` means omitted (using `*int` to distinguish 0 from unset)
- `offset *int` — OFFSET value; `nil` means omitted
- `lock string` — row-level lock clause (`FOR UPDATE`, `FOR SHARE`, etc.)
- `lockOption string` — lock modifier (`NOWAIT` or `SKIP LOCKED`)
- `setOps []setOpClause` — UNION / INTERSECT / EXCEPT clauses

### ToSql vs ToSqlWithDialect

`SqlSelectBuilder` implements `syntax.SqlFragment`, which allows it to be embedded as a subquery.

- `ToSql() string`
  - Renders the full query using the builder's own dialect
  - Appends a trailing semicolon
  - No surrounding parentheses
  - Used for top-level query output
- `ToSqlWithDialect(d dialect.SqlDialect) string`
  - Renders the query using the provided dialect (inherited from the parent builder)
  - No trailing semicolon
  - Wrapped in parentheses
  - Called automatically when the builder is passed as a subquery to `From`, `Join`, or `In`

### Clause rendering order

`renderSQL` assembles clauses in this fixed order regardless of method call order:

- `WITH [RECURSIVE] cte AS (...), ...`
- `SELECT [DISTINCT [ON (col, ...)]] col, ...` (defaults to `*` if no columns are set)
- `FROM table`
- `[INNER|LEFT|RIGHT|FULL OUTER|CROSS] [LATERAL] JOIN table ON condition`
- `WHERE condition`
- `GROUP BY col, ...`
- `HAVING condition`
- `ORDER BY col [ASC|DESC], ...`
- `LIMIT n`
- `OFFSET n`
- `FOR [UPDATE|SHARE|NO KEY UPDATE|KEY SHARE] [NOWAIT|SKIP LOCKED]`
- `[UNION|INTERSECT|EXCEPT [ALL]] SELECT ...`

Because state is accumulated independently and rendered at the end, method call order does not matter. `Select(...).From(...)` and `From(...).Select(...)` produce identical output.

### SqlInsertBuilder struct

Fields:
- `dialect` — the `SqlDialect` used for rendering
- `into string` — the target table name
- `columns []string` — column list (omitted from SQL when empty)
- `rows []syntax.ValueRow` — one entry per `Values(...)` call
- `orConflict string` — SQLite `INSERT OR` conflict action (`"REPLACE"`, `"IGNORE"`, etc.)
- `conflictAction string` — `"NOTHING"` or `"UPDATE"` for `ON CONFLICT` (PostgreSQL/SQLite)
- `conflictTarget []string` — conflict target columns for `ON CONFLICT (...)`
- `conflictSets []string` — `SET` assignments for `ON CONFLICT ... DO UPDATE SET`
- `duplicateKeySets []string` — `SET` assignments for MySQL `ON DUPLICATE KEY UPDATE`
- `returning []string` — RETURNING columns

Clause rendering order:
- `INSERT [OR conflict] INTO table` — `OR` clause is rendered for SQLite when `orConflict` is set
- `(col1, col2, ...)` — omitted if no columns are set
- `VALUES (row1), (row2), ...` — omitted if no rows are added
- `ON CONFLICT ... DO NOTHING / DO UPDATE SET ...` — PostgreSQL/SQLite
- `ON DUPLICATE KEY UPDATE ...` — MySQL only (rendered via `OnDuplicateKeyUpdate`)
- `RETURNING col, ...`

Value rendering in rows:
- `Parameter` → rendered via `Parameter.ToSqlWithDialect`
- `bool` → rendered via `dialect.Bool`
- `string` → used as-is (SQL literal)
- other types → `fmt.Sprintf("%v", v)`

### SqlUpdateBuilder struct

Fields:
- `dialect` — the `SqlDialect` used for rendering
- `table string` — the target table name
- `sets []setClause` — one entry per `Set(col, val)` call
- `where SqlFragment` — WHERE condition
- `returning []string` — RETURNING columns

Clause rendering order:
- `UPDATE table`
- `SET col = val, ...`
- `WHERE condition`
- `RETURNING col, ...`

### SqlDeleteBuilder struct

Fields:
- `dialect` — the `SqlDialect` used for rendering
- `from string` — the target table name
- `where SqlFragment` — WHERE condition
- `returning []string` — RETURNING columns

Clause rendering order:
- `DELETE FROM table`
- `WHERE condition`
- `RETURNING col, ...`

### SqlCreateTableBuilder struct

Builds `CREATE TABLE` statements. Columns are added via `Column(name, type)`, and constraints via `PrimaryKey`, `Unique`, `Check`, `ForeignKey`, `NotNull`, `Default`, etc.

### SqlCreateIndexBuilder struct

Builds `CREATE [UNIQUE] INDEX` statements. Supports `IfNotExists()`, `On(table)`, `Columns(cols...)`, `Using(method)`, and `Where(predicate)` for partial indexes.

### SqlDropTableBuilder struct

Builds `DROP TABLE` statements. Supports `IfExists()` and `Cascade()` (PostgreSQL only; silently ignored for other dialects).

### SqlDropIndexBuilder struct

Builds `DROP INDEX` statements. Supports `IfExists()` and `On(table)`.

Dialect differences:
- PostgreSQL / SQLite: `DROP INDEX [IF EXISTS] "idx_name"`
- MySQL: `DROP INDEX "idx_name" ON "table_name"` (`IfExists` is ignored; `On()` is required)

### SqlAlterTableBuilder struct

Builds `ALTER TABLE` statements. Multiple operations can be chained; each call appends one operation:

- `AddColumn(name, type string)` — `ADD COLUMN "name" TYPE`
- `DropColumn(name string)` — `DROP COLUMN "name"`
- `RenameColumn(from, to string)` — `RENAME COLUMN "from" TO "to"`
- `RenameTable(to string)` — `RENAME TO "to"`

Note: SQLite support for `DROP COLUMN` and `RENAME COLUMN` requires version 3.35.0+.

---

## sqblpg / sqblmysql / sqblsqlite packages

Each DB-specific package has two files:

- `sqblXXX.go`
  - Entry points that wire the appropriate dialect and return a typed builder:
    - `From(table any) *SqlSelectBuilder` — starts a SELECT query
    - `Select(columns ...any) *SqlSelectBuilder` — starts a column-first SELECT query
    - `InsertInto(table string) *SqlInsertBuilder` — starts an INSERT query
    - `Update(table string) *SqlUpdateBuilder` — starts an UPDATE query
    - `DeleteFrom(table string) *SqlDeleteBuilder` — starts a DELETE query
    - `CreateTable(table string) *SqlCreateTableBuilder` — starts a CREATE TABLE statement
    - `CreateIndex(name string) *SqlCreateIndexBuilder` — starts a CREATE INDEX statement
    - `DropTable(table string) *SqlDropTableBuilder` — starts a DROP TABLE statement
    - `DropIndex(name string) *SqlDropIndexBuilder` — starts a DROP INDEX statement
    - `AlterTable(table string) *SqlAlterTableBuilder` — starts an ALTER TABLE statement
- `reexport.go`
  - Re-exports helper functions from the `syntax` package (`Eq`, `And`, `In`, `Asc`, `P`, `Fn`, `Over`, etc.)
  - Allows users to import only `sqblpg` (or `sqblmysql` / `sqblsqlite`) and write `sqblpg.Eq(...)` instead of also importing `syntax`
  - DB-specific functions that do not apply are intentionally omitted (e.g. `ILike` is only in `sqblpg` since it is a PostgreSQL extension)
  - `syntax.Values` is not re-exported; rows are added directly via `SqlInsertBuilder.Values(vals ...any)`
