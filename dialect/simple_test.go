package dialect

import (
	"bytes"
	"testing"
)

func TestSimpleDialect_Quote(t *testing.T) {
	t.Parallel()
	d := &SimpleDialect{}
	var buf bytes.Buffer
	d.Quote(&buf, "users")
	if got, want := buf.String(), `"users"`; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestSimpleDialect_QuoteIdentifier(t *testing.T) {
	t.Parallel()
	d := &SimpleDialect{}
	var buf bytes.Buffer
	d.QuoteIdentifier(&buf, "created_at")
	if got, want := buf.String(), `"created_at"`; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestSimpleDialect_PlaceholderPositional(t *testing.T) {
	t.Parallel()
	d := &SimpleDialect{}
	var buf bytes.Buffer
	d.PlaceholderPositional(&buf)
	if got, want := buf.String(), "?"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestSimpleDialect_PlaceholderIndexed(t *testing.T) {
	t.Parallel()
	d := &SimpleDialect{}
	var buf bytes.Buffer
	d.PlaceholderIndexed(&buf, 2)
	if got, want := buf.String(), "?2"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestSimpleDialect_Bool(t *testing.T) {
	t.Parallel()
	d := &SimpleDialect{}
	var buf bytes.Buffer
	d.Bool(&buf, true)
	if got := buf.String(); got != "TRUE" {
		t.Errorf("Bool(true) = %q, want TRUE", got)
	}
	buf.Reset()
	d.Bool(&buf, false)
	if got := buf.String(); got != "FALSE" {
		t.Errorf("Bool(false) = %q, want FALSE", got)
	}
}
