package dialect

import (
	"bytes"
	"strconv"
	"strings"
)

// Postgres implements the Dialect interface for PostgreSQL.
type Postgres struct{}

var _ Dialect = &Postgres{}

// Quote writes a single identifier part wrapped in double quotes into buf.
// Any existing double-quote characters within the identifier are escaped by doubling.
//
// Examples (PostgreSQL):
//
//	users      -> "users"
//	my"table   -> "my""table"
func (p *Postgres) Quote(buf *bytes.Buffer, s string) {
	if strings.IndexByte(s, '"') < 0 {
		buf.WriteByte('"')
		buf.WriteString(s)
		buf.WriteByte('"')
		return
	}
	WriteQuotedPart(buf, s, '"')
}

// QuoteIdentifier writes a full identifier (possibly dot-separated) into buf,
// quoting each part individually.
//
// Examples (PostgreSQL):
//
//	users          -> "users"
//	public.users   -> "public"."users"
//	db.sch.table   -> "db"."sch"."table"
func (p *Postgres) QuoteIdentifier(buf *bytes.Buffer, name string) {
	if strings.IndexByte(name, '.') < 0 {
		p.Quote(buf, name)
		return
	}
	QuoteIdentifierParts(buf, name, '"')
}

// PlaceholderPositional writes the positional bind parameter "?" into buf.
func (p *Postgres) PlaceholderPositional(buf *bytes.Buffer) {
	buf.WriteByte('?')
}

// PlaceholderIndexed writes the indexed bind parameter "$N" into buf.
// Index starts from 1.
//
// Examples:
//
//	1 -> $1, 2 -> $2
func (p *Postgres) PlaceholderIndexed(buf *bytes.Buffer, index int) {
	buf.WriteByte('$')
	buf.WriteString(strconv.Itoa(index))
}

// Bool writes the SQL boolean literal for b into buf.
//
// Examples:
//
//	true -> TRUE, false -> FALSE
func (p *Postgres) Bool(buf *bytes.Buffer, b bool) {
	if b {
		buf.WriteString("TRUE")
	} else {
		buf.WriteString("FALSE")
	}
}
