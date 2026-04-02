package sqltesting

import "testing"

func TestCompact(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "multiline with indentation",
			input: "\n\t\tSELECT id\n\t\tFROM users\n\t",
			want:  "SELECT id FROM users",
		},
		{
			name:  "consecutive spaces",
			input: "SELECT  id   FROM   users",
			want:  "SELECT id FROM users",
		},
		{
			name:  "leading and trailing whitespace",
			input: "  SELECT 1  ",
			want:  "SELECT 1",
		},
		{
			name:  "already compact",
			input: "SELECT id FROM users",
			want:  "SELECT id FROM users",
		},
		{
			name:  "empty string",
			input: "",
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := Compact(tt.input); got != tt.want {
				t.Errorf("Compact(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
