# Custom Dialect

sqbl supports any SQL database by implementing the `SqlDialect` interface. This document walks through the interface contract and a complete implementation example.

---

## The SqlDialect interface

```go
// dialect/dialect.go
type SqlDialect interface {
    Quote(str string) string
    QuoteIdentifier(str string) string
    PlaceholderPositional() string
    PlaceholderIndexed(index int) string
    Bool(b bool) string
}
```

Each method has a narrow, well-defined contract:

- `Quote(str string) string`
  - Wraps a **single** identifier segment (no dots) in the dialect's quote character
  - Must escape any occurrence of the quote character inside the string
  - Must never split on dots — that is `QuoteIdentifier`'s job
  - Example: a DB using `[` / `]` for quoting would turn `my]col` into `[my]]col]`

- `QuoteIdentifier(name string) string`
  - Handles a dot-separated identifier such as `schema.table` or `db.schema.table`
  - Must split on `.` and call `Quote` on each segment independently
  - Must rejoin the segments with `.`
  - This ensures that `public.users` becomes `"public"."users"` rather than `"public.users"`

- `PlaceholderPositional() string`
  - Returns the positional bind parameter token (used by `P()` with no arguments)
  - For most databases this is `?`

- `PlaceholderIndexed(index int) string`
  - Returns the indexed bind parameter token for the given position (1-based)
  - For databases that use indexed placeholders (`$1`, `$2`): use the index
  - For databases that use positional placeholders (`?`): ignore the index

- `Bool(b bool) string`
  - Returns the SQL literal for a boolean value
  - Use `TRUE` / `FALSE` for databases with a native boolean type
  - Use `1` / `0` for databases without one (e.g. SQLite)

---

## Implementation example: DuckDB

DuckDB follows the SQL standard and uses double-quote identifiers and `$N` placeholders, similar to PostgreSQL.

```go
package dialect

import (
    "fmt"
    "strings"
)

type DuckDBDialect struct{}

var _ SqlDialect = &DuckDBDialect{}

func (d *DuckDBDialect) Quote(str string) string {
    escaped := strings.ReplaceAll(str, `"`, `""`)
    return `"` + escaped + `"`
}

func (d *DuckDBDialect) QuoteIdentifier(name string) string {
    parts := strings.Split(name, ".")
    for i, part := range parts {
        parts[i] = d.Quote(part)
    }
    return strings.Join(parts, ".")
}

func (d *DuckDBDialect) PlaceholderPositional() string {
    return "?"
}

func (d *DuckDBDialect) PlaceholderIndexed(index int) string {
    return fmt.Sprintf("$%d", index)
}

func (d *DuckDBDialect) Bool(b bool) string {
    if b {
        return "TRUE"
    }
    return "FALSE"
}
```

---

## Creating a convenience package

Once you have a dialect, create a convenience package that wires it automatically, following the same pattern as `sqblpg` or `sqblmysql`.

A minimal package needs two files.

### sqblduckdb/sqblduckdb.go

```go
package sqblduckdb

import (
    "github.com/kotofurumiya/sqbl/builder"
    "github.com/kotofurumiya/sqbl/dialect"
)

func newSelectBuilder() *builder.SqlSelectBuilder {
    b := &builder.SqlSelectBuilder{}
    return b.Dialect(&dialect.DuckDBDialect{})
}

func From(table any) *builder.SqlSelectBuilder {
    return newSelectBuilder().From(table)
}

func Select(columns ...any) *builder.SqlSelectBuilder {
    return newSelectBuilder().Select(columns...)
}

func InsertInto(table string) *builder.SqlInsertBuilder {
    b := &builder.SqlInsertBuilder{}
    return b.Dialect(&dialect.DuckDBDialect{}).Into(table)
}

func Update(table string) *builder.SqlUpdateBuilder {
    b := &builder.SqlUpdateBuilder{}
    return b.Dialect(&dialect.DuckDBDialect{}).Table(table)
}

func DeleteFrom(table string) *builder.SqlDeleteBuilder {
    b := &builder.SqlDeleteBuilder{}
    return b.Dialect(&dialect.DuckDBDialect{}).From(table)
}

func CreateTable(table string) *builder.SqlCreateTableBuilder {
    b := &builder.SqlCreateTableBuilder{}
    return b.Dialect(&dialect.DuckDBDialect{}).Table(table)
}

func CreateIndex(name string) *builder.SqlCreateIndexBuilder {
    b := &builder.SqlCreateIndexBuilder{}
    return b.Dialect(&dialect.DuckDBDialect{}).Name(name)
}

func DropTable(table string) *builder.SqlDropTableBuilder {
    b := &builder.SqlDropTableBuilder{}
    return b.Dialect(&dialect.DuckDBDialect{}).Table(table)
}

func DropIndex(name string) *builder.SqlDropIndexBuilder {
    b := &builder.SqlDropIndexBuilder{}
    return b.Dialect(&dialect.DuckDBDialect{}).Name(name)
}

func AlterTable(table string) *builder.SqlAlterTableBuilder {
    b := &builder.SqlAlterTableBuilder{}
    return b.Dialect(&dialect.DuckDBDialect{}).Table(table)
}
```

### sqblduckdb/reexport.go

Re-export the syntax helpers that apply to your database. Omit any function that your database does not support (e.g. omit `ILike` for databases that lack case-insensitive LIKE).

```go
package sqblduckdb

import "github.com/kotofurumiya/sqbl/syntax"

func As(source any, alias string) *syntax.Aliased              { return syntax.As(source, alias) }
func Eq(left any, right any) syntax.ComparisonExpr             { return syntax.Eq(left, right) }
func Ne(left any, right any) syntax.ComparisonExpr             { return syntax.Ne(left, right) }
func Lt(left any, right any) syntax.ComparisonExpr             { return syntax.Lt(left, right) }
func Lte(left any, right any) syntax.ComparisonExpr            { return syntax.Lte(left, right) }
func Gt(left any, right any) syntax.ComparisonExpr             { return syntax.Gt(left, right) }
func Gte(left any, right any) syntax.ComparisonExpr            { return syntax.Gte(left, right) }
func And(exprs ...syntax.SqlFragment) syntax.LogicalExpr       { return syntax.And(exprs...) }
func Or(exprs ...syntax.SqlFragment) syntax.LogicalExpr        { return syntax.Or(exprs...) }
func Not(expr syntax.SqlFragment) syntax.NotExpr               { return syntax.Not(expr) }
func In(left any, values ...any) syntax.InExpr                 { return syntax.In(left, values...) }
func NotIn(left any, values ...any) syntax.InExpr              { return syntax.NotIn(left, values...) }
func IsNull(col any) syntax.NullExpr                           { return syntax.IsNull(col) }
func IsNotNull(col any) syntax.NullExpr                        { return syntax.IsNotNull(col) }
func Between(col any, low, high any) syntax.BetweenExpr        { return syntax.Between(col, low, high) }
func Like(left any, pattern string) syntax.ComparisonExpr      { return syntax.Like(left, pattern) }
func Asc(col string) syntax.Order                              { return syntax.Asc(col) }
func Desc(col string) syntax.Order                             { return syntax.Desc(col) }
func P(args ...any) syntax.Parameter                           { return syntax.P(args...) }
func Fn(name string, args ...any) *syntax.SqlFn                { return syntax.Fn(name, args...) }
func Over(expr any) *syntax.WindowExpr                         { return syntax.Over(expr) }
```

---

## Using a custom dialect without a convenience package

If you prefer not to create a separate package, you can set the dialect directly on the builder using `builder.SqlSelectBuilder.Dialect`:

```go
import (
    "github.com/kotofurumiya/sqbl/builder"
    "github.com/kotofurumiya/sqbl/syntax"
)

b := (&builder.SqlSelectBuilder{}).
    Dialect(&DuckDBDialect{}).
    From("users").
    Select("id", "name").
    Where(syntax.Eq("active", true))

sql := b.ToSql()
```

---

## What the dialect does NOT control

The dialect is intentionally limited to lexical differences. The following are always handled by `builder.SqlSelectBuilder` regardless of dialect:

- The order of SQL clauses (WITH → SELECT → FROM → JOIN → WHERE → ...)
- Keywords: `SELECT`, `FROM`, `WHERE`, `JOIN ON`, `GROUP BY`, `HAVING`, `ORDER BY`, `LIMIT`, `OFFSET`, `UNION`, `INTERSECT`, `EXCEPT`, `FOR UPDATE`, etc.
- Parenthesization of subqueries
- The trailing semicolon added by `ToSql()`

If you need to generate SQL that differs structurally from what `SqlSelectBuilder` produces, that is beyond the scope of the dialect interface and would require changes to `builder`.
