package syntax

import (
	"fmt"
	"strings"

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

var _ SqlFragment = (*SqlFn)(nil)

// Fn creates a SQL function call expression.
// The name is the function name (e.g. "SUM", "COUNT").
// Each argument is converted via ToFragment: strings are passed as-is (unquoted),
// numeric values are formatted with fmt.Sprint.
//
//	Fn("SUM", "amount")           → SUM(amount)
//	Fn("COUNT", "*")              → COUNT(*)
//	Fn("COALESCE", "x", 0)        → COALESCE(x, 0)
//	Fn("SUM", Fn("ABS", "val"))   → SUM(ABS(val))
func Fn(name string, args ...any) *SqlFn {
	return &SqlFn{name: name, args: args}
}

// ToSqlWithDialect implements SqlFragment.
func (f *SqlFn) ToSqlWithDialect(d dialect.SqlDialect) string {
	argStrs := make([]string, 0, len(f.args))
	for _, arg := range f.args {
		argStrs = append(argStrs, fnArgToSql(d, arg))
	}
	return f.name + "(" + strings.Join(argStrs, ", ") + ")"
}

// fnArgToSql renders a single Fn argument.
// SqlFragment values (including nested Fn calls, Parameters, and conditions) are rendered via ToSqlWithDialect.
// String values are treated as identifiers and quoted unless they are raw expressions.
// Numeric and other non-string values are rendered as raw SQL literals.
func fnArgToSql(d dialect.SqlDialect, v any) string {
	switch val := v.(type) {
	case SqlFragment:
		return val.ToSqlWithDialect(d)
	case string:
		return ToFragment(val).ToSqlWithDialect(d)
	default:
		return fmt.Sprint(val)
	}
}
