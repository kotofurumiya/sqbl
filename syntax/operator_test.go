package syntax

import (
	"fmt"
	"testing"

	"github.com/kotofurumiya/sqbl/dialect"
)

func TestOperator_Comparisons(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		fn         func(left any, right any) ComparisonExpr
		expectedOp string
	}{
		{
			name:       "Eq",
			fn:         Eq,
			expectedOp: "=",
		},
		{
			name:       "Ne",
			fn:         Ne,
			expectedOp: "<>",
		},
		{
			name:       "Lt",
			fn:         Lt,
			expectedOp: "<",
		},
		{
			name:       "Lte",
			fn:         Lte,
			expectedOp: "<=",
		},
		{
			name:       "Gt",
			fn:         Gt,
			expectedOp: ">",
		},
		{
			name:       "Gte",
			fn:         Gte,
			expectedOp: ">=",
		},
	}

	d := &dialect.SimpleDialect{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// t.Parallel()
			got := tt.fn("left", "right").ToSqlWithDialect(d)
			expected := fmt.Sprintf(`"left" %s "right"`, tt.expectedOp)
			if got != expected {
				t.Errorf("%s(%s, %s) = %q; want %q", tt.name, "left", "right", got, expected)
			}
		})
	}
}

func TestOperator_LogicalExpr(t *testing.T) {
	t.Parallel()
	d := &dialect.SimpleDialect{}

	tests := []struct {
		name     string
		expr     LogicalExpr
		expected string
	}{
		{
			name:     "And flat",
			expr:     And(Eq("a", 1), Eq("b", 2)),
			expected: `"a" = 1 AND "b" = 2`,
		},
		{
			name:     "Or flat",
			expr:     Or(Eq("a", 1), Eq("b", 2)),
			expected: `"a" = 1 OR "b" = 2`,
		},
		{
			name:     "Or nested in And",
			expr:     And(Eq("a", 1), Or(Eq("b", 2), Eq("c", 3))),
			expected: `"a" = 1 AND ("b" = 2 OR "c" = 3)`,
		},
		{
			name:     "And nested in Or",
			expr:     Or(And(Eq("a", 1), Eq("b", 2)), Eq("c", 3)),
			expected: `("a" = 1 AND "b" = 2) OR "c" = 3`,
		},
		{
			name:     "And nested in And",
			expr:     And(And(Eq("a", 1), Eq("b", 2)), Eq("c", 3)),
			expected: `("a" = 1 AND "b" = 2) AND "c" = 3`,
		},
		{
			name:     "Or nested in Or",
			expr:     Or(Or(Eq("a", 1), Eq("b", 2)), Eq("c", 3)),
			expected: `("a" = 1 OR "b" = 2) OR "c" = 3`,
		},
		{
			name:     "three levels deep",
			expr:     And(Eq("a", 1), Or(Eq("b", 2), And(Eq("c", 3), Eq("d", 4)))),
			expected: `"a" = 1 AND ("b" = 2 OR ("c" = 3 AND "d" = 4))`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// t.Parallel()
			got := tt.expr.ToSqlWithDialect(d)
			if got != tt.expected {
				t.Errorf("got %q; want %q", got, tt.expected)
			}
		})
	}
}

func TestOperator_Not(t *testing.T) {
	t.Parallel()
	d := &dialect.SimpleDialect{}

	tests := []struct {
		name     string
		expr     NotExpr
		expected string
	}{
		{
			name:     "Not Eq",
			expr:     Not(Eq("active", false)),
			expected: `NOT ("active" = FALSE)`,
		},
		{
			name:     "Not And",
			expr:     Not(And(Eq("a", 1), Eq("b", 2))),
			expected: `NOT ("a" = 1 AND "b" = 2)`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// t.Parallel()
			got := tt.expr.ToSqlWithDialect(d)
			if got != tt.expected {
				t.Errorf("got %q; want %q", got, tt.expected)
			}
		})
	}
}

func TestOperator_In(t *testing.T) {
	t.Parallel()
	d := &dialect.SimpleDialect{}

	tests := []struct {
		name     string
		expr     InExpr
		expected string
	}{
		{
			name:     "In strings",
			expr:     In("status", "'active'", "'pending'"),
			expected: `"status" IN ('active', 'pending')`,
		},
		{
			name:     "In ints",
			expr:     In("id", 1, 2, 3),
			expected: `"id" IN (1, 2, 3)`,
		},
		{
			name:     "NotIn",
			expr:     NotIn("status", "'deleted'", "'banned'"),
			expected: `"status" NOT IN ('deleted', 'banned')`,
		},
		{
			name:     "In bool",
			expr:     In("active", true, false),
			expected: `"active" IN (TRUE, FALSE)`,
		},
		{
			name:     "In with Fn left",
			expr:     In(Fn("LOWER", "email"), "'alice@example.com'"),
			expected: `LOWER("email") IN ('alice@example.com')`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// t.Parallel()
			got := tt.expr.ToSqlWithDialect(d)
			if got != tt.expected {
				t.Errorf("got %q; want %q", got, tt.expected)
			}
		})
	}
}

func TestOperator_Null(t *testing.T) {
	t.Parallel()
	d := &dialect.SimpleDialect{}

	tests := []struct {
		name     string
		expr     NullExpr
		expected string
	}{
		{
			name:     "IsNull",
			expr:     IsNull("deleted_at"),
			expected: `"deleted_at" IS NULL`,
		},
		{
			name:     "IsNotNull",
			expr:     IsNotNull("email"),
			expected: `"email" IS NOT NULL`,
		},
		{
			name:     "IsNull with Fn",
			expr:     IsNull(Fn("NULLIF", "score", 0)),
			expected: `NULLIF("score", 0) IS NULL`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// t.Parallel()
			got := tt.expr.ToSqlWithDialect(d)
			if got != tt.expected {
				t.Errorf("got %q; want %q", got, tt.expected)
			}
		})
	}
}

func TestOperator_Between(t *testing.T) {
	t.Parallel()
	d := &dialect.SimpleDialect{}

	tests := []struct {
		name     string
		expr     BetweenExpr
		expected string
	}{
		{
			name:     "int bounds",
			expr:     Between("age", 18, 65),
			expected: `"age" BETWEEN 18 AND 65`,
		},
		{
			name:     "string bounds",
			expr:     Between("created_at", "'2025-01-01'", "'2025-12-31'"),
			expected: `"created_at" BETWEEN '2025-01-01' AND '2025-12-31'`,
		},
		{
			name:     "bool bounds",
			expr:     Between("flag", false, true),
			expected: `"flag" BETWEEN FALSE AND TRUE`,
		},
		{
			name:     "Between with Fn left",
			expr:     Between(Fn("ABS", "score"), 0, 100),
			expected: `ABS("score") BETWEEN 0 AND 100`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// t.Parallel()
			got := tt.expr.ToSqlWithDialect(d)
			if got != tt.expected {
				t.Errorf("got %q; want %q", got, tt.expected)
			}
		})
	}
}

func TestOperator_LikeILike(t *testing.T) {
	t.Parallel()
	d := &dialect.SimpleDialect{}

	tests := []struct {
		name     string
		expr     ComparisonExpr
		expected string
	}{
		{
			name:     "Like",
			expr:     Like("name", "'%foo%'"),
			expected: `"name" LIKE '%foo%'`,
		},
		{
			name:     "ILike",
			expr:     ILike("name", "'%foo%'"),
			expected: `"name" ILIKE '%foo%'`,
		},
		{
			name:     "Like with Fn left",
			expr:     Like(Fn("LOWER", "name"), "'%foo%'"),
			expected: `LOWER("name") LIKE '%foo%'`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// t.Parallel()
			got := tt.expr.ToSqlWithDialect(d)
			if got != tt.expected {
				t.Errorf("got %q; want %q", got, tt.expected)
			}
		})
	}
}

func TestOperator_ComparisonsBool(t *testing.T) {
	t.Parallel()
	d := &dialect.SimpleDialect{}

	tests := []struct {
		name     string
		value    bool
		expected string
	}{
		{
			name:     "TRUE",
			value:    true,
			expected: "TRUE",
		},
		{
			name:     "FALSE",
			value:    false,
			expected: "FALSE",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// t.Parallel()
			got := d.Bool(tt.value)
			if got != tt.expected {
				t.Errorf("dialect.Bool(%v) = %q; want %q", tt.value, got, tt.expected)
			}
		})
	}
}
