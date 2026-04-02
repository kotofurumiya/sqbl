package syntax

import "github.com/kotofurumiya/sqbl/dialect"

type paramMode int

const (
	paramPositional paramMode = iota // P()
	paramIndexed                     // P(1)
	paramNamed                       // P(":name")
)

// Parameter represents a SQL bind parameter placeholder.
// It implements SqlFragment and renders without quoting.
type Parameter struct {
	mode  paramMode
	index int    // indexed mode
	name  string // named mode (e.g. ":status", "@name")
}

var _ SqlFragment = Parameter{}

// ToSqlWithDialect renders the parameter as a placeholder string using the given dialect.
// positional mode: d.PlaceholderPositional() (e.g. ?)
// indexed mode: d.PlaceholderIndexed(p.index) (e.g. $1)
// named mode: p.name as-is (e.g. :status).
func (p Parameter) ToSqlWithDialect(d dialect.SqlDialect) string {
	switch p.mode {
	case paramPositional:
		return d.PlaceholderPositional()
	case paramIndexed:
		return d.PlaceholderIndexed(p.index)
	case paramNamed:
		return p.name
	default:
		return d.PlaceholderPositional()
	}
}

// P creates a Parameter placeholder.
//
//	P()          → positional: ? (all dialects)
//	P(1)         → indexed:    $1 (PostgreSQL), ? (MySQL/SQLite)
//	P(":status") → named:      :status (as-is, all dialects)
//	P("@name")   → named:      @name (as-is, all dialects)
func P(args ...any) Parameter {
	if len(args) == 0 {
		return Parameter{mode: paramPositional}
	}
	switch v := args[0].(type) {
	case int:
		return Parameter{mode: paramIndexed, index: v}
	case string:
		return Parameter{mode: paramNamed, name: v}
	default:
		return Parameter{mode: paramPositional}
	}
}
