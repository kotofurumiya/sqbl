package syntax

import (
	"bytes"

	"github.com/kotofurumiya/sqbl/dialect"
)

// SqlFn represents a SQL function call such as SUM(amount) or COUNT(*).
// It implements SqlFragment and can be used anywhere a column expression is accepted.
type SqlFn struct {
	name string
	args []any
}

var _ SqlFragment = SqlFn{}

// Fn creates a SQL function call expression.
// The name is written as-is. Each argument accepts the same values as
// AppendValue, so a plain string is an identifier.
//
//	Fn("SUM", "amount")                  → SUM("amount")
//	Fn("COUNT", "*")                     → COUNT(*)
//	Fn("COALESCE", "x", 0)               → COALESCE("x", 0)
//	Fn("SUM", Fn("ABS", "val"))          → SUM(ABS("val"))
//	Fn("DATE_TRUNC", Raw("'day'"), "ts") → DATE_TRUNC('day', "ts")
//	Fn("NOW")                            → NOW()
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
		AppendValue(buf, d, arg)
	}
	buf.WriteByte(')')
}
