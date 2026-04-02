package dialect

import (
	"testing"
)

func TestSimpleDialect_Quote(t *testing.T) {
	t.Parallel()
	d := &SimpleDialect{}
	if got, want := d.Quote("users"), `"users"`; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestSimpleDialect_QuoteIdentifier(t *testing.T) {
	t.Parallel()
	d := &SimpleDialect{}
	if got, want := d.QuoteIdentifier("created_at"), `"created_at"`; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestSimpleDialect_PlaceholderPositional(t *testing.T) {
	t.Parallel()
	d := &SimpleDialect{}
	if got, want := d.PlaceholderPositional(), "?"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestSimpleDialect_PlaceholderIndexed(t *testing.T) {
	t.Parallel()
	d := &SimpleDialect{}
	if got, want := d.PlaceholderIndexed(2), "?2"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestSimpleDialect_Bool(t *testing.T) {
	t.Parallel()
	d := &SimpleDialect{}
	if got := d.Bool(true); got != "TRUE" {
		t.Errorf("Bool(true) = %q, want TRUE", got)
	}
	if got := d.Bool(false); got != "FALSE" {
		t.Errorf("Bool(false) = %q, want FALSE", got)
	}
}
