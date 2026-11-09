package builder

import (
	"bytes"
	"errors"

	"github.com/kotofurumiya/sqbl/dialect"
	"github.com/kotofurumiya/sqbl/internal/sqlbuf"
	"github.com/kotofurumiya/sqbl/internal/sqlerr"
)

// Statement is implemented by all SQL statement builders.
//
// Rendering fails with an error, instead of producing broken SQL, when the
// statement holds input that has no safe SQL form, such as a value of an
// unsupported Go type or a missing dialect. The Must variants panic instead,
// for statements built from constants where a failure is a programming error.
type Statement interface {
	// ToSQL renders the statement with the builder's own dialect and a trailing semicolon.
	ToSQL() (string, error)
	// ToSQLWithDialect renders the statement with d and no trailing semicolon.
	ToSQLWithDialect(d dialect.Dialect) (string, error)
	// MustSQL is like ToSQL but panics on error.
	MustSQL() string
	// MustSQLWithDialect is like ToSQLWithDialect but panics on error.
	MustSQLWithDialect(d dialect.Dialect) string
}

// renderer is the rendering core every builder implements.
type renderer interface {
	renderSQL(buf *bytes.Buffer, d dialect.Dialect)
}

// render runs r.renderSQL with d and turns a failure raised inside any fragment into an error.
func render(r renderer, d dialect.Dialect, semicolon bool) (sql string, err error) {
	if d == nil {
		return "", errors.New("sqbl: no dialect to render with; use ToSQLWithDialect, " +
			"or start from a dialect package such as sqblpg")
	}
	buf := sqlbuf.GetStringBuffer()
	defer sqlbuf.PutStringBuffer(buf)
	defer sqlerr.Catch(&err)
	r.renderSQL(buf, d)
	if semicolon {
		buf.WriteByte(';')
	}
	return buf.String(), nil
}

// must returns sql, or panics with err.
func must(sql string, err error) string {
	if err != nil {
		panic(err)
	}
	return sql
}

// quoteIdentifiers writes each name quoted by the dialect into buf, separated by ", ".
// Used by table-constraint closures in create_table.go.
//
//	quoteIdentifiers(buf, d, []string{"id", "name"})  →  "id", "name"
//	quoteIdentifiers(buf, d, nil)                     →  (nothing written)
func quoteIdentifiers(buf *bytes.Buffer, d dialect.Dialect, names []string) {
	for i, name := range names {
		if i > 0 {
			buf.WriteString(", ")
		}
		d.QuoteIdentifier(buf, name)
	}
}
