package dialect

import (
	"bytes"
	"strconv"
)

// SimpleDialect is a minimal dialect that applies no escaping.
// It is intended for use in tests where dialect-specific behavior is not under test.
type SimpleDialect struct{}

var _ SqlDialect = &SimpleDialect{}

// Quote writes the identifier wrapped in double quotes into buf, without any escaping.
func (s *SimpleDialect) Quote(buf *bytes.Buffer, str string) {
	buf.WriteByte('"')
	buf.WriteString(str)
	buf.WriteByte('"')
}

// QuoteIdentifier writes the identifier wrapped in double quotes into buf, without any escaping.
func (s *SimpleDialect) QuoteIdentifier(buf *bytes.Buffer, str string) {
	buf.WriteByte('"')
	buf.WriteString(str)
	buf.WriteByte('"')
}

// PlaceholderPositional writes "?" into buf.
func (s *SimpleDialect) PlaceholderPositional(buf *bytes.Buffer) {
	buf.WriteByte('?')
}

// PlaceholderIndexed writes "?N" for the given index (e.g. ?1, ?2) into buf.
func (s *SimpleDialect) PlaceholderIndexed(buf *bytes.Buffer, index int) {
	buf.WriteByte('?')
	buf.WriteString(strconv.Itoa(index))
}

// Bool writes "TRUE" or "FALSE" into buf.
func (s *SimpleDialect) Bool(buf *bytes.Buffer, b bool) {
	if b {
		buf.WriteString("TRUE")
	} else {
		buf.WriteString("FALSE")
	}
}
