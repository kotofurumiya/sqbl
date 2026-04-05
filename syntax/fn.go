package syntax

import (
	"bytes"
	"fmt"

	"github.com/kotofurumiya/sqbl/dialect"
)

// SqlFn represents a SQL function call such as SUM(amount) or COUNT(*).
// It implements SqlFragment and can be used anywhere a column expression is accepted.
//
//	syntax.Fn("SUM", "amount")              // → SUM(amount)
//	syntax.Fn("COUNT", "*")                 // → COUNT(*)
//	syntax.Fn("COALESCE", "s.total", 0)     // → COALESCE(s.total, 0)
//	syntax.Fn("NOW")                        // → NOW()
type SqlFn struct {
	name string
	args []any
}

var _ SqlFragment = SqlFn{}

// Fn creates a SQL function call expression.
// The name is the function name (e.g. "SUM", "COUNT").
// Each argument is converted via ToFragment: strings are passed as-is (unquoted),
// numeric values are formatted with fmt.Sprint.
//
//	Fn("SUM", "amount")           → SUM(amount)
//	Fn("COUNT", "*")              → COUNT(*)
//	Fn("COALESCE", "x", 0)        → COALESCE(x, 0)
//	Fn("SUM", Fn("ABS", "val"))   → SUM(ABS(val))
func Fn(name string, args ...any) SqlFn {
	return SqlFn{name: name, args: args}
}

// AppendSQL implements SqlFragment, writing the function call into buf.
func (f SqlFn) AppendSQL(buf *bytes.Buffer, d dialect.SqlDialect) {
	buf.WriteString(f.name)
	buf.WriteByte('(')
	for i, arg := range f.args {
		if i > 0 {
			buf.WriteString(", ")
		}
		fnArgAppend(buf, d, arg)
	}
	buf.WriteByte(')')
}

// fnArgAppend writes a single Fn argument directly into buf.
// SqlFragment values (including nested Fn calls, Parameters, and conditions) are appended via AppendSQL.
// String values are treated as identifiers and quoted unless they are raw expressions.
// Numeric and other non-string values are rendered as raw SQL literals.
func fnArgAppend(buf *bytes.Buffer, d dialect.SqlDialect, v any) {
	switch val := v.(type) {
	case SqlFragment:
		val.AppendSQL(buf, d)
	case string:
		// Inline StringExpr logic to avoid ToFragment allocation.
		if isSimpleIdentifier(val) {
			d.QuoteIdentifier(buf, val)
		} else {
			buf.WriteString(val)
		}
	default:
		buf.WriteString(fmt.Sprint(val))
	}
}
