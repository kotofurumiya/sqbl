package syntax

import (
	"bytes"
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
		{name: "Eq", fn: Eq, expectedOp: "="},
		{name: "Ne", fn: Ne, expectedOp: "<>"},
		{name: "Lt", fn: Lt, expectedOp: "<"},
		{name: "Lte", fn: Lte, expectedOp: "<="},
		{name: "Gt", fn: Gt, expectedOp: ">"},
		{name: "Gte", fn: Gte, expectedOp: ">="},
	}

	d := &dialect.SimpleDialect{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// t.Parallel()
			var buf bytes.Buffer
			tt.fn("left", "right").AppendSQL(&buf, d)
			expected := fmt.Sprintf(`"left" %s "right"`, tt.expectedOp)
			if got := buf.String(); got != expected {
				t.Errorf("%s(%s, %s) = %q; want %q", tt.name, "left", "right", got, expected)
			}
		})
	}
}

func TestOperator_ComparisonOperands(t *testing.T) {
	t.Parallel()
	d := &dialect.PostgresDialect{}

	tests := []struct {
		name     string
		expr     ComparisonExpr
		expected string
	}{
		{name: "column to column", expr: Eq("u.id", "o.user_id"), expected: `"u"."id" = "o"."user_id"`},
		// A plain string is never treated as a value, even if it looks like one.
		{name: "string is not a value", expr: Eq("status", "pending"), expected: `"status" = "pending"`},
		{name: "string is not an expression", expr: Eq("n", "n + 1"), expected: `"n" = "n + 1"`},
		{name: "parameter", expr: Eq("status", P(1)), expected: `"status" = $1`},
		{name: "raw literal", expr: Eq("status", Raw("'pending'")), expected: `"status" = 'pending'`},
		{name: "int", expr: Eq("age", 18), expected: `"age" = 18`},
		{name: "nil", expr: Eq("a", nil), expected: `"a" = NULL`},
		{name: "Fn on left", expr: Eq(Fn("LOWER", "email"), P(1)), expected: `LOWER("email") = $1`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			tt.expr.AppendSQL(&buf, d)
			if got := buf.String(); got != tt.expected {
				t.Errorf("got %q; want %q", got, tt.expected)
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
		{name: "And flat", expr: And(Eq("a", 1), Eq("b", 2)), expected: `"a" = 1 AND "b" = 2`},
		{name: "Or flat", expr: Or(Eq("a", 1), Eq("b", 2)), expected: `"a" = 1 OR "b" = 2`},
		{name: "Or nested in And", expr: And(Eq("a", 1), Or(Eq("b", 2), Eq("c", 3))), expected: `"a" = 1 AND ("b" = 2 OR "c" = 3)`},
		{name: "And nested in Or", expr: Or(And(Eq("a", 1), Eq("b", 2)), Eq("c", 3)), expected: `("a" = 1 AND "b" = 2) OR "c" = 3`},
		{name: "And nested in And", expr: And(And(Eq("a", 1), Eq("b", 2)), Eq("c", 3)), expected: `("a" = 1 AND "b" = 2) AND "c" = 3`},
		{name: "Or nested in Or", expr: Or(Or(Eq("a", 1), Eq("b", 2)), Eq("c", 3)), expected: `("a" = 1 OR "b" = 2) OR "c" = 3`},
		{name: "three levels deep", expr: And(Eq("a", 1), Or(Eq("b", 2), And(Eq("c", 3), Eq("d", 4)))), expected: `"a" = 1 AND ("b" = 2 OR ("c" = 3 AND "d" = 4))`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// t.Parallel()
			var buf bytes.Buffer
			tt.expr.AppendSQL(&buf, d)
			if got := buf.String(); got != tt.expected {
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
		{name: "Not Eq", expr: Not(Eq("active", false)), expected: `NOT ("active" = FALSE)`},
		{name: "Not And", expr: Not(And(Eq("a", 1), Eq("b", 2))), expected: `NOT ("a" = 1 AND "b" = 2)`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// t.Parallel()
			var buf bytes.Buffer
			tt.expr.AppendSQL(&buf, d)
			if got := buf.String(); got != tt.expected {
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
		{name: "In strings", expr: In("status", Raw("'active'"), Raw("'pending'")), expected: `"status" IN ('active', 'pending')`},
		{name: "In ints", expr: In("id", 1, 2, 3), expected: `"id" IN (1, 2, 3)`},
		{name: "NotIn", expr: NotIn("status", Raw("'deleted'"), Raw("'banned'")), expected: `"status" NOT IN ('deleted', 'banned')`},
		{name: "In bool", expr: In("active", true, false), expected: `"active" IN (TRUE, FALSE)`},
		{name: "In plain strings are identifiers", expr: In("a", "b", "c"), expected: `"a" IN ("b", "c")`},
		{name: "In with Fn left", expr: In(Fn("LOWER", "email"), P(1)), expected: `LOWER("email") IN (?1)`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// t.Parallel()
			var buf bytes.Buffer
			tt.expr.AppendSQL(&buf, d)
			if got := buf.String(); got != tt.expected {
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
		{name: "IsNull", expr: IsNull("deleted_at"), expected: `"deleted_at" IS NULL`},
		{name: "IsNotNull", expr: IsNotNull("email"), expected: `"email" IS NOT NULL`},
		{name: "IsNull with Fn", expr: IsNull(Fn("NULLIF", "score", 0)), expected: `NULLIF("score", 0) IS NULL`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// t.Parallel()
			var buf bytes.Buffer
			tt.expr.AppendSQL(&buf, d)
			if got := buf.String(); got != tt.expected {
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
		{name: "int bounds", expr: Between("age", 18, 65), expected: `"age" BETWEEN 18 AND 65`},
		{name: "string bounds", expr: Between("created_at", Raw("'2025-01-01'"), Raw("'2025-12-31'")), expected: `"created_at" BETWEEN '2025-01-01' AND '2025-12-31'`},
		{name: "bool bounds", expr: Between("flag", false, true), expected: `"flag" BETWEEN FALSE AND TRUE`},
		{name: "column bounds", expr: Between("x", "lo", "hi"), expected: `"x" BETWEEN "lo" AND "hi"`},
		{name: "parameter bounds", expr: Between("x", P(1), P(2)), expected: `"x" BETWEEN ?1 AND ?2`},
		{name: "Between with Fn left", expr: Between(Fn("ABS", "score"), 0, 100), expected: `ABS("score") BETWEEN 0 AND 100`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// t.Parallel()
			var buf bytes.Buffer
			tt.expr.AppendSQL(&buf, d)
			if got := buf.String(); got != tt.expected {
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
		{name: "Like", expr: Like("name", Raw("'%foo%'")), expected: `"name" LIKE '%foo%'`},
		{name: "ILike", expr: ILike("name", Raw("'%foo%'")), expected: `"name" ILIKE '%foo%'`},
		{name: "Like parameter", expr: Like("name", P(1)), expected: `"name" LIKE ?1`},
		{name: "Like with Fn left", expr: Like(Fn("LOWER", "name"), Raw("'%foo%'")), expected: `LOWER("name") LIKE '%foo%'`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// t.Parallel()
			var buf bytes.Buffer
			tt.expr.AppendSQL(&buf, d)
			if got := buf.String(); got != tt.expected {
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
		{name: "TRUE", value: true, expected: "TRUE"},
		{name: "FALSE", value: false, expected: "FALSE"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// t.Parallel()
			var buf bytes.Buffer
			d.Bool(&buf, tt.value)
			if got := buf.String(); got != tt.expected {
				t.Errorf("dialect.Bool(%v) = %q; want %q", tt.value, got, tt.expected)
			}
		})
	}
}
