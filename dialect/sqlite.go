package dialect

import (
	"bytes"
	"strconv"
	"strings"
)

// SqliteDialect implements the Dialect interface for SQLite.
type SqliteDialect struct{}

var _ SqlDialect = &SqliteDialect{}

// Quote writes a single identifier part wrapped in double quotes into buf.
// Any existing double-quote characters within the identifier are escaped by doubling.
//
// Examples (SQLite):
//
//	users      -> "users"
//	my"table   -> "my""table"
func (s *SqliteDialect) Quote(buf *bytes.Buffer, str string) {
	if strings.IndexByte(str, '"') < 0 {
		buf.WriteByte('"')
		buf.WriteString(str)
		buf.WriteByte('"')
		return
	}
	WriteQuotedPart(buf, str, '"')
}

// QuoteIdentifier writes a full identifier (possibly dot-separated) into buf,
// quoting each part individually.
//
// Examples (SQLite):
//
//	users          -> "users"
//	main.users     -> "main"."users"
//	db.sch.table   -> "db"."sch"."table"
func (s *SqliteDialect) QuoteIdentifier(buf *bytes.Buffer, name string) {
	if strings.IndexByte(name, '.') < 0 {
		s.Quote(buf, name)
		return
	}
	QuoteIdentifierParts(buf, name, '"')
}

// PlaceholderPositional writes the positional bind parameter "?" into buf.
func (s *SqliteDialect) PlaceholderPositional(buf *bytes.Buffer) {
	buf.WriteByte('?')
}

// PlaceholderIndexed writes the numbered bind parameter "?N" into buf.
// Index starts from 1. The number binds the argument by position, so the same
// index can appear more than once and in any order.
//
// Examples:
//
//	1 -> ?1, 2 -> ?2
func (s *SqliteDialect) PlaceholderIndexed(buf *bytes.Buffer, index int) {
	buf.WriteByte('?')
	buf.WriteString(strconv.Itoa(index))
}

// Bool writes the SQL boolean literal for b into buf.
// SQLite has no native boolean type; 1 represents true and 0 represents false.
//
// Examples:
//
//	true -> 1, false -> 0
func (s *SqliteDialect) Bool(buf *bytes.Buffer, b bool) {
	if b {
		buf.WriteByte('1')
	} else {
		buf.WriteByte('0')
	}
}
