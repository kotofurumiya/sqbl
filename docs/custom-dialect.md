# Custom Dialect

sqbl supports any SQL database by implementing the `Dialect` interface. This document walks through the interface contract and a complete implementation example.

---

## The Dialect interface

```go
// dialect/dialect.go
type Dialect interface {
    Quote(buf *bytes.Buffer, s string)
    QuoteIdentifier(buf *bytes.Buffer, name string)
    PlaceholderPositional(buf *bytes.Buffer)
    PlaceholderIndexed(buf *bytes.Buffer, index int)
    Bool(buf *bytes.Buffer, b bool)
}
```

All methods write directly into a `*bytes.Buffer` instead of returning strings. This eliminates intermediate string allocations throughout the rendering pipeline.

Each method has a narrow, well-defined contract:

- `Quote(buf *bytes.Buffer, s string)`
  - Writes a **single** identifier segment (no dots) wrapped in the dialect's quote character into buf
  - Must escape any occurrence of the quote character inside the string
  - Must never split on dots — that is `QuoteIdentifier`'s job
  - Example: a DB using `[` / `]` for quoting would write `my]col` as `[my]]col]`

- `QuoteIdentifier(buf *bytes.Buffer, name string)`
  - Handles a dot-separated identifier such as `schema.table` or `db.schema.table`
  - Must split on `.` and write each segment quoted with `Quote`
  - Must write `.` between segments
  - This ensures that `public.users` becomes `"public"."users"` rather than `"public.users"`

- `PlaceholderPositional(buf *bytes.Buffer)`
  - Writes the positional bind parameter token (used by `P()` with no arguments) into buf
  - For most databases this is `?`

- `PlaceholderIndexed(buf *bytes.Buffer, index int)`
  - Writes the indexed bind parameter token for the given position (1-based) into buf
  - For databases that use indexed placeholders (`$1`, `$2`): use the index
  - For databases that use positional placeholders (`?`): ignore the index

- `Bool(buf *bytes.Buffer, b bool)`
  - Writes the SQL literal for a boolean value into buf
  - Use `TRUE` / `FALSE` for databases with a native boolean type
  - Use `1` / `0` for databases without one (e.g. SQLite)

---

## Implementation example: DuckDB

DuckDB follows the SQL standard and uses double-quote identifiers and `$N` placeholders, similar to PostgreSQL.

```go
package sqblduckdb

import (
    "bytes"
    "strconv"
    "strings"

    "github.com/kotofurumiya/sqbl/dialect"
)

type DuckDBDialect struct{}

var _ dialect.Dialect = &DuckDBDialect{}

func (d *DuckDBDialect) Quote(buf *bytes.Buffer, s string) {
    dialect.WriteQuotedPart(buf, s, '"')
}

func (d *DuckDBDialect) QuoteIdentifier(buf *bytes.Buffer, name string) {
    if strings.IndexByte(name, '.') < 0 {
        d.Quote(buf, name)
        return
    }
    dialect.QuoteIdentifierParts(buf, name, '"')
}

func (d *DuckDBDialect) PlaceholderPositional(buf *bytes.Buffer) {
    buf.WriteByte('?')
}

func (d *DuckDBDialect) PlaceholderIndexed(buf *bytes.Buffer, index int) {
    buf.WriteByte('$')
    buf.WriteString(strconv.Itoa(index))
}

func (d *DuckDBDialect) Bool(buf *bytes.Buffer, b bool) {
    if b {
        buf.WriteString("TRUE")
    } else {
        buf.WriteString("FALSE")
    }
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

func newSelectBuilder() *builder.SelectBuilder {
    b := &builder.SelectBuilder{}
    return b.Dialect(&DuckDBDialect{})
}

func From(table any) *builder.SelectBuilder {
    return newSelectBuilder().From(table)
}

func Select(columns ...any) *builder.SelectBuilder {
    return newSelectBuilder().Select(columns...)
}

func InsertInto(table string) *builder.InsertBuilder {
    b := &builder.InsertBuilder{}
    return b.Dialect(&DuckDBDialect{}).Into(table)
}

func Update(table string) *builder.UpdateBuilder {
    b := &builder.UpdateBuilder{}
    return b.Dialect(&DuckDBDialect{}).Table(table)
}

func DeleteFrom(table string) *builder.DeleteBuilder {
    b := &builder.DeleteBuilder{}
    return b.Dialect(&DuckDBDialect{}).From(table)
}

func CreateTable(table string) *builder.CreateTableBuilder {
    b := &builder.CreateTableBuilder{}
    return b.Dialect(&DuckDBDialect{}).Table(table)
}

func CreateIndex(name string) *builder.CreateIndexBuilder {
    b := &builder.CreateIndexBuilder{}
    return b.Dialect(&DuckDBDialect{}).Name(name)
}

func DropTable(table string) *builder.DropTableBuilder {
    b := &builder.DropTableBuilder{}
    return b.Dialect(&DuckDBDialect{}).Table(table)
}

func DropIndex(name string) *builder.DropIndexBuilder {
    b := &builder.DropIndexBuilder{}
    return b.Dialect(&DuckDBDialect{}).Name(name)
}

func AlterTable(table string) *builder.AlterTableBuilder {
    b := &builder.AlterTableBuilder{}
    return b.Dialect(&DuckDBDialect{}).Table(table)
}
```

### sqblduckdb/reexport.go

Re-export the syntax helpers that apply to your database. Omit any function that your database does not support (e.g. omit `ILike` for databases that lack case-insensitive LIKE).

```go
package sqblduckdb

import "github.com/kotofurumiya/sqbl/syntax"

func As(source any, alias string) syntax.Aliased               { return syntax.As(source, alias) }
func Eq(left any, right any) syntax.ComparisonExpr             { return syntax.Eq(left, right) }
func Ne(left any, right any) syntax.ComparisonExpr             { return syntax.Ne(left, right) }
func Lt(left any, right any) syntax.ComparisonExpr             { return syntax.Lt(left, right) }
func Lte(left any, right any) syntax.ComparisonExpr            { return syntax.Lte(left, right) }
func Gt(left any, right any) syntax.ComparisonExpr             { return syntax.Gt(left, right) }
func Gte(left any, right any) syntax.ComparisonExpr            { return syntax.Gte(left, right) }
func And(exprs ...syntax.SQLFragment) syntax.LogicalExpr       { return syntax.And(exprs...) }
func Or(exprs ...syntax.SQLFragment) syntax.LogicalExpr        { return syntax.Or(exprs...) }
func Not(expr syntax.SQLFragment) syntax.NotExpr               { return syntax.Not(expr) }
func In(left any, values ...any) syntax.InExpr                 { return syntax.In(left, values...) }
func NotIn(left any, values ...any) syntax.InExpr              { return syntax.NotIn(left, values...) }
func IsNull(col any) syntax.NullExpr                           { return syntax.IsNull(col) }
func IsNotNull(col any) syntax.NullExpr                        { return syntax.IsNotNull(col) }
func Between(col any, low, high any) syntax.BetweenExpr        { return syntax.Between(col, low, high) }
func Like(left any, pattern string) syntax.ComparisonExpr      { return syntax.Like(left, pattern) }
func Asc(col string) syntax.Order                              { return syntax.Asc(col) }
func Desc(col string) syntax.Order                             { return syntax.Desc(col) }
func P(args ...any) syntax.Parameter                           { return syntax.P(args...) }
func Fn(name string, args ...any) syntax.FnExpr                 { return syntax.Fn(name, args...) }
func Over(expr any) syntax.WindowExpr                          { return syntax.Over(expr) }
```

---

## Using a custom dialect without a convenience package

If you prefer not to create a separate package, you can set the dialect directly on the builder using `builder.SelectBuilder.Dialect`:

```go
import (
    "github.com/kotofurumiya/sqbl/builder"
    "github.com/kotofurumiya/sqbl/syntax"
)

b := (&builder.SelectBuilder{}).
    Dialect(&DuckDBDialect{}).
    From("users").
    Select("id", "name").
    Where(syntax.Eq("active", true))

sql := b.ToSQL()
```

---

## What the dialect does NOT control

The dialect is intentionally limited to lexical differences. The following are always handled by `builder.SelectBuilder` regardless of dialect:

- The order of SQL clauses (WITH → SELECT → FROM → JOIN → WHERE → ...)
- Keywords: `SELECT`, `FROM`, `WHERE`, `JOIN ON`, `GROUP BY`, `HAVING`, `ORDER BY`, `LIMIT`, `OFFSET`, `UNION`, `INTERSECT`, `EXCEPT`, `FOR UPDATE`, etc.
- Parenthesization of subqueries
- The trailing semicolon added by `ToSQL()`

If you need to generate SQL that differs structurally from what `SelectBuilder` produces, that is beyond the scope of the dialect interface and would require changes to `builder`.
