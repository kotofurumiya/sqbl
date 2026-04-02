package dialect

import "fmt"

// SimpleDialect is a minimal dialect that applies no escaping.
// It is intended for use in tests where dialect-specific behavior is not under test.
type SimpleDialect struct{}

var _ SqlDialect = &SimpleDialect{}

// Quote wraps the identifier in double quotes without any escaping.
func (s *SimpleDialect) Quote(str string) string {
	return "\"" + str + "\""
}

// QuoteIdentifier wraps the identifier in double quotes without any escaping.
func (s *SimpleDialect) QuoteIdentifier(str string) string {
	return "\"" + str + "\""
}

// PlaceholderPositional returns "?".
func (s *SimpleDialect) PlaceholderPositional() string {
	return "?"
}

// PlaceholderIndexed returns "?N" for the given index (e.g. ?1, ?2).
func (s *SimpleDialect) PlaceholderIndexed(index int) string {
	return fmt.Sprintf("?%d", index)
}

// Bool returns "TRUE" or "FALSE".
func (s *SimpleDialect) Bool(b bool) string {
	if b {
		return "TRUE"
	}
	return "FALSE"
}
