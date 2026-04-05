package dialect

import (
	"bytes"
	"strings"
)

// MysqlDialect implements the Dialect interface for MySQL.
type MysqlDialect struct{}

var _ SqlDialect = &MysqlDialect{}

// Quote writes a single identifier part wrapped in backticks into buf.
// Any existing backtick characters within the identifier are escaped by doubling.
//
// Examples (MySQL):
//
//	users      -> `users`
//	my`table   -> `my``table`
func (m *MysqlDialect) Quote(buf *bytes.Buffer, s string) {
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
func (m *MysqlDialect) QuoteIdentifier(buf *bytes.Buffer, name string) {
	if strings.IndexByte(name, '.') < 0 {
		m.Quote(buf, name)
		return
	}
	QuoteIdentifierParts(buf, name, '`')
}

// PlaceholderPositional writes the positional bind parameter "?" into buf.
func (m *MysqlDialect) PlaceholderPositional(buf *bytes.Buffer) {
	buf.WriteByte('?')
}

// PlaceholderIndexed writes the indexed bind parameter "?" into buf.
// MySQL uses ? for all indexed placeholders, ignoring the index.
func (m *MysqlDialect) PlaceholderIndexed(buf *bytes.Buffer, _ int) {
	buf.WriteByte('?')
}

// Bool writes the SQL boolean literal for b into buf.
//
// Examples:
//
//	true -> TRUE, false -> FALSE
func (m *MysqlDialect) Bool(buf *bytes.Buffer, b bool) {
	if b {
		buf.WriteString("TRUE")
	} else {
		buf.WriteString("FALSE")
	}
}
