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
  - Must split on `.` and write each segment quoted and escaped like `Quote`
  - Must write `.` between segments
  - `dialect.QuoteIdentifierParts` does all of this for a given quote character
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
The `dialect` package exports `WriteQuotedPart` and `QuoteIdentifierParts`, which quote and escape with a given quote character.

```go
package sqblduckdb

import (
	"bytes"
	"strconv"

	"github.com/kotofurumiya/sqbl/dialect"
)

// Dialect renders SQL for DuckDB.
type Dialect struct{}

var _ dialect.Dialect = &Dialect{}

func (d *Dialect) Quote(buf *bytes.Buffer, s string) {
	dialect.WriteQuotedPart(buf, s, '"')
}

func (d *Dialect) QuoteIdentifier(buf *bytes.Buffer, name string) {
	dialect.QuoteIdentifierParts(buf, name, '"')
}

func (d *Dialect) PlaceholderPositional(buf *bytes.Buffer) {
	buf.WriteByte('?')
}

func (d *Dialect) PlaceholderIndexed(buf *bytes.Buffer, index int) {
	buf.WriteByte('$')
	buf.WriteString(strconv.Itoa(index))
}

func (d *Dialect) Bool(buf *bytes.Buffer, b bool) {
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
Builders are values: each method returns a modified copy, so the entry points return builder values, not pointers.

A minimal package needs two files besides the dialect.

### sqblduckdb/sqblduckdb.go

```go
package sqblduckdb

import "github.com/kotofurumiya/sqbl/builder"

func From(table any) builder.SelectBuilder {
	var b builder.SelectBuilder
	return b.Dialect(&Dialect{}).From(table)
}

func Select(columns ...any) builder.SelectBuilder {
	var b builder.SelectBuilder
	return b.Dialect(&Dialect{}).Select(columns...)
}

func InsertInto(table string) builder.InsertBuilder {
	var b builder.InsertBuilder
	return b.Dialect(&Dialect{}).Into(table)
}

func Update(table string) builder.UpdateBuilder {
	var b builder.UpdateBuilder
	return b.Dialect(&Dialect{}).Table(table)
}

func DeleteFrom(table string) builder.DeleteBuilder {
	var b builder.DeleteBuilder
	return b.Dialect(&Dialect{}).From(table)
}

func CreateTable(table string) builder.CreateTableBuilder {
	var b builder.CreateTableBuilder
	return b.Dialect(&Dialect{}).Table(table)
}

func CreateIndex(name string) builder.CreateIndexBuilder {
	var b builder.CreateIndexBuilder
	return b.Dialect(&Dialect{}).Name(name)
}

func DropTable(table string) builder.DropTableBuilder {
	var b builder.DropTableBuilder
	return b.Dialect(&Dialect{}).Table(table)
}

func DropIndex(name string) builder.DropIndexBuilder {
	var b builder.DropIndexBuilder
	return b.Dialect(&Dialect{}).Name(name)
}

func AlterTable(table string) builder.AlterTableBuilder {
	var b builder.AlterTableBuilder
	return b.Dialect(&Dialect{}).Table(table)
}
```

### sqblduckdb/reexport.go

Re-export the syntax helpers that apply to your database, so users need only one import. Omit any function that your database does not support.

- Keep the parameter types of the `syntax` functions. For example, `Like` takes `pattern any`: a plain string there is an identifier, so a pattern is passed with `P` or `Raw`.
- Re-export `Ident` and `Raw` as type aliases. Without `Raw`, users of the package cannot write SQL literals.

```go
package sqblduckdb

import "github.com/kotofurumiya/sqbl/syntax"

type Ident = syntax.Ident
type Raw = syntax.Raw

func As(source any, alias string) syntax.Aliased { return syntax.As(source, alias) }
func AsColumns(source any, alias string, columns ...string) syntax.Aliased {
	return syntax.AsColumns(source, alias, columns...)
}
func Eq(left any, right any) syntax.ComparisonExpr       { return syntax.Eq(left, right) }
func Ne(left any, right any) syntax.ComparisonExpr       { return syntax.Ne(left, right) }
func Lt(left any, right any) syntax.ComparisonExpr       { return syntax.Lt(left, right) }
func Lte(left any, right any) syntax.ComparisonExpr      { return syntax.Lte(left, right) }
func Gt(left any, right any) syntax.ComparisonExpr       { return syntax.Gt(left, right) }
func Gte(left any, right any) syntax.ComparisonExpr      { return syntax.Gte(left, right) }
func And(exprs ...syntax.SQLFragment) syntax.LogicalExpr { return syntax.And(exprs...) }
func Or(exprs ...syntax.SQLFragment) syntax.LogicalExpr  { return syntax.Or(exprs...) }
func Not(expr syntax.SQLFragment) syntax.NotExpr         { return syntax.Not(expr) }
func In(left any, values ...any) syntax.InExpr           { return syntax.In(left, values...) }
func NotIn(left any, values ...any) syntax.InExpr        { return syntax.NotIn(left, values...) }
func IsNull(col any) syntax.NullExpr                     { return syntax.IsNull(col) }
func IsNotNull(col any) syntax.NullExpr                  { return syntax.IsNotNull(col) }
func Between(col any, low, high any) syntax.BetweenExpr  { return syntax.Between(col, low, high) }
func Like(left any, pattern any) syntax.ComparisonExpr   { return syntax.Like(left, pattern) }
func ILike(left any, pattern any) syntax.ComparisonExpr  { return syntax.ILike(left, pattern) }
func Asc(expr any) syntax.Order                          { return syntax.Asc(expr) }
func Desc(expr any) syntax.Order                         { return syntax.Desc(expr) }
func P(args ...any) syntax.Parameter                     { return syntax.P(args...) }
func Fn(name string, args ...any) syntax.FnExpr          { return syntax.Fn(name, args...) }
func Cast(expr any, typ string) syntax.CastExpr          { return syntax.Cast(expr, typ) }
func Over(expr any) syntax.WindowExpr                    { return syntax.Over(expr) }
```

Usage:

```go
// SELECT "id", "we""ird" FROM "public"."users" WHERE "active" = TRUE AND "name" ILIKE $1;
sqblduckdb.From("public.users").
	Select("id", "we\"ird").
	Where(sqblduckdb.And(sqblduckdb.Eq("active", true), sqblduckdb.ILike("name", sqblduckdb.P(1)))).
	MustSQL()
```

---

## Using a custom dialect without a convenience package

If you prefer not to create a separate package, set the dialect directly on a builder with its `Dialect` method:

```go
import (
	"github.com/kotofurumiya/sqbl/builder"
	"github.com/kotofurumiya/sqbl/syntax"
)

var b builder.SelectBuilder
sql, err := b.Dialect(&sqblduckdb.Dialect{}).
	From("users").
	Select("id", "name").
	Where(syntax.Eq("active", true)).
	ToSQL()
// SELECT "id", "name" FROM "users" WHERE "active" = TRUE;
```

---

## What the dialect does NOT control

The dialect is intentionally limited to lexical differences. The following are always handled by the builders regardless of dialect:

- The order of SQL clauses (WITH → SELECT → FROM → JOIN → WHERE → ...)
- Keywords: `SELECT`, `FROM`, `WHERE`, `JOIN ON`, `GROUP BY`, `HAVING`, `ORDER BY`, `LIMIT`, `OFFSET`, `UNION`, `INTERSECT`, `EXCEPT`, `FOR UPDATE`, etc.
- Parenthesization of subqueries and set-operation operands
- Rendering errors for input with no safe SQL form, such as an unsupported value type
- The trailing semicolon added by `ToSQL()`

If you need to generate SQL that differs structurally from what the builders produce, that is beyond the scope of the dialect interface and would require changes to `builder`.
