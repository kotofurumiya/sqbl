package dialect

import (
	"bytes"
	"strings"
)

// QuoteIdentifierParts writes each dot-separated segment of a qualified identifier
// into buf, quoting each part with quoteChar.
//
// This avoids all intermediate string allocations that strings.Split + strings.Join
// would require, and also avoids the strings.Builder allocation from the previous
// implementation.
//
// quoteChar is the dialect-specific identifier quote byte ('"' for Postgres/SQLite,
// '`' for MySQL).
func QuoteIdentifierParts(buf *bytes.Buffer, name string, quoteChar byte) {
	rest := name
	for {
		dot := strings.IndexByte(rest, '.')
		if dot < 0 {
			// Last (or only) part — write and finish.
			WriteQuotedPart(buf, rest, quoteChar)
			break
		}
		WriteQuotedPart(buf, rest[:dot], quoteChar)
		buf.WriteByte('.')
		rest = rest[dot+1:]
	}
}

// WriteQuotedPart writes s to buf wrapped in quoteChar.
// Any occurrence of quoteChar within s is doubled to escape it.
func WriteQuotedPart(buf *bytes.Buffer, s string, quoteChar byte) {
	buf.WriteByte(quoteChar)
	if strings.IndexByte(s, quoteChar) < 0 {
		// No escaping needed — write s directly.
		buf.WriteString(s)
	} else {
		// Double each quoteChar occurrence.
		for {
			i := strings.IndexByte(s, quoteChar)
			if i < 0 {
				buf.WriteString(s)
				break
			}
			buf.WriteString(s[:i+1]) // include the quoteChar itself
			buf.WriteByte(quoteChar) // second copy (escape)
			s = s[i+1:]
		}
	}
	buf.WriteByte(quoteChar)
}
