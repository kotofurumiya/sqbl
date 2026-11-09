package sqltesting

import "testing"

func TestCompact(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"multiline with indentation", "\n\t\tSELECT id\n\t\tFROM users\n\t", "SELECT id FROM users"},
		{"consecutive spaces", "SELECT  id   FROM   users", "SELECT id FROM users"},
		{"already compact", "SELECT id FROM users", "SELECT id FROM users"},
		{"empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := Compact(tt.input); got != tt.want {
				t.Errorf("Compact(%q) = %q; want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestAssertSQL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		got      string
		want     string
		wantFail bool
	}{
		{"multi-line want matches compact got", "SELECT 1 FROM t", "\n\tSELECT 1\n\tFROM t\n", false},
		// got is not compacted, so whitespace bugs in generated SQL are still reported.
		{"extra space in got fails", "SELECT  1", "SELECT 1", true},
		{"different SQL fails", "SELECT 1", "SELECT 2", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			rec := &recorder{T: t}
			AssertSQL(rec, tt.got, tt.want)
			if rec.failed != tt.wantFail {
				t.Errorf("failed = %v; want %v", rec.failed, tt.wantFail)
			}
		})
	}
}

// recorder captures Errorf calls so AssertSQL can be tested without failing the outer test.
type recorder struct {
	*testing.T
	failed bool
}

func (r *recorder) Errorf(string, ...any) { r.failed = true }
