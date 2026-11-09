package syntax

import (
	"bytes"
	"fmt"
	"reflect"
	"regexp"

	"github.com/kotofurumiya/sqbl/dialect"
	"github.com/kotofurumiya/sqbl/internal/sqlerr"
)

type paramMode int

const (
	paramPositional paramMode = iota // P()
	paramIndexed                     // P(1)
	paramNamed                       // P(":name")
	paramInvalid                     // rejected arguments; fails when rendered
)

// namedParamPattern limits named placeholders to a prefix and a plain name,
// so a name can never smuggle SQL into the statement.
var namedParamPattern = regexp.MustCompile(`^[:@$][A-Za-z_][A-Za-z0-9_]*$`)

// Parameter represents a SQL bind parameter placeholder.
// Create one with P.
type Parameter struct {
	mode    paramMode
	index   int    // indexed mode
	name    string // named mode (e.g. ":status", "@name")
	invalid string // invalid mode: why P rejected its arguments
}

var _ SQLFragment = Parameter{}

// AppendSQL implements SQLFragment, writing the placeholder into buf.
func (p Parameter) AppendSQL(buf *bytes.Buffer, d dialect.Dialect) {
	switch p.mode {
	case paramPositional:
		d.PlaceholderPositional(buf)
	case paramIndexed:
		d.PlaceholderIndexed(buf, p.index)
	case paramNamed:
		buf.WriteString(p.name)
	default:
		sqlerr.Fail("%s", p.invalid)
	}
}

// P creates a bind parameter placeholder. The value itself is passed to the
// database driver as a query argument, never written into the SQL.
//
//	P()          → positional: ? in every dialect
//	P(1)         → indexed:    $1 (PostgreSQL), ?1 (SQLite), ? (MySQL)
//	P(":status") → named:      :status in every dialect
//	P("@name")   → named:      @name in every dialect
//
// An index must be an integer of 1 or more. A name must be one of the prefixes
// :, @ or $ followed by letters, digits and underscores. Any other argument
// makes the statement fail to render.
//
// MySQL has no numbered placeholders, so P(n) renders a plain ? there and the
// number is ignored: arguments bind in the order the placeholders appear in the
// SQL, and an index used twice needs its argument passed twice.
//
// P takes a placeholder name, not a value: never pass user input to P.
func P(args ...any) Parameter {
	switch len(args) {
	case 0:
		return Parameter{mode: paramPositional}
	case 1:
	default:
		return invalidParam("P takes at most one argument, got %d", len(args))
	}
	if name, ok := args[0].(string); ok {
		if !namedParamPattern.MatchString(name) {
			return invalidParam("invalid parameter name %q; use a prefix :, @ or $ followed by a plain name, "+
				"and pass values as query arguments", name)
		}
		return Parameter{mode: paramNamed, name: name}
	}
	rv := reflect.ValueOf(args[0])
	var index int64
	switch rv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		index = rv.Int()
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		if rv.Uint() > uint64(maxIndex) {
			return invalidParam("parameter index %d is too large", rv.Uint())
		}
		index = int64(rv.Uint())
	default:
		return invalidParam("P takes an integer index or a placeholder name, got %T", args[0])
	}
	if index < 1 || index > maxIndex {
		return invalidParam("parameter index must be between 1 and %d, got %d", maxIndex, index)
	}
	return Parameter{mode: paramIndexed, index: int(index)}
}

// maxIndex keeps indexes within int on every platform.
const maxIndex = 1<<31 - 1

func invalidParam(format string, args ...any) Parameter {
	return Parameter{mode: paramInvalid, invalid: fmt.Sprintf(format, args...)}
}
