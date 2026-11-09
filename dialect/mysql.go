package dialect

import (
	"bytes"
	"strings"
)

// MySQL implements the Dialect interface for MySQL.
type MySQL struct{}

var _ Dialect = &MySQL{}

// Quote writes a single identifier part wrapped in backticks into buf.
// Any existing backtick characters within the identifier are escaped by doubling.
//
// Examples (MySQL):
//
//	users      -> `users`
//	my`table   -> `my``table`
func (m *MySQL) Quote(buf *bytes.Buffer, s string) {
	if strings.IndexByte(s, '`') < 0 {
		buf.WriteByte('`')
		buf.WriteString(s)
		buf.WriteByte('`')
		return
	}
	WriteQuotedPart(buf, s, '`')
}

// QuoteIdentifier writes a full identifier (possibly dot-separated) into buf,
// quoting each part individually.
//
// Examples (MySQL):
//
//	users          -> `users`
//	public.users   -> `public`.`users`
//	db.sch.table   -> `db`.`sch`.`table`
func (m *MySQL) QuoteIdentifier(buf *bytes.Buffer, name string) {
	if strings.IndexByte(name, '.') < 0 {
		m.Quote(buf, name)
		return
	}
	QuoteIdentifierParts(buf, name, '`')
}

// PlaceholderPositional writes the positional bind parameter "?" into buf.
func (m *MySQL) PlaceholderPositional(buf *bytes.Buffer) {
	buf.WriteByte('?')
}

// PlaceholderIndexed writes the indexed bind parameter "?" into buf.
// MySQL uses ? for all indexed placeholders, ignoring the index.
func (m *MySQL) PlaceholderIndexed(buf *bytes.Buffer, _ int) {
	buf.WriteByte('?')
}

// Bool writes the SQL boolean literal for b into buf.
//
// Examples:
//
//	true -> TRUE, false -> FALSE
func (m *MySQL) Bool(buf *bytes.Buffer, b bool) {
	if b {
		buf.WriteString("TRUE")
	} else {
		buf.WriteString("FALSE")
	}
}
