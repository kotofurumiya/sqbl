package syntax

import (
	"testing"

	"github.com/kotofurumiya/sqbl/internal/sqltesting"
)

func TestPredicates(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		fragment SQLFragment
		want     string
	}{
		{"Eq", Eq("a", P(1)), `"a" = $1`},
		{"Ne", Ne("a", P(1)), `"a" <> $1`},
		{"Lt", Lt("a", P(1)), `"a" < $1`},
		{"Lte", Lte("a", P(1)), `"a" <= $1`},
		{"Gt", Gt("a", P(1)), `"a" > $1`},
		{"Gte", Gte("a", P(1)), `"a" >= $1`},
		{"Like", Like("name", Raw("'%foo%'")), `"name" LIKE '%foo%'`},
		{"ILike", ILike("name", P(1)), `"name" ILIKE $1`},

		// Operands accept every value kind, on either side.
		{"column to column", Eq("u.id", "o.user_id"), `"u"."id" = "o"."user_id"`},
		{"string operand is an identifier, not a value", Eq("status", "pending"), `"status" = "pending"`},
		{"literal operands", Eq(1, true), "1 = TRUE"},
		{"nil operand", Eq("a", nil), `"a" = NULL`},
		{"function operand", Like(Fn("LOWER", "email"), P(1)), `LOWER("email") LIKE $1`},

		{"In", In("id", P(1), 2, "other_id"), `"id" IN ($1, 2, "other_id")`},
		{"In single value", In("id", P(1)), `"id" IN ($1)`},
		{"In subquery-like fragment", In("id", Raw("SELECT 1")), `"id" IN (SELECT 1)`},
		// sqbl does not validate, so an empty list renders as-is.
		{"In without values", In("id"), `"id" IN ()`},
		{"NotIn", NotIn(Fn("LOWER", "role"), Raw("'guest'"), Raw("'bot'")), `LOWER("role") NOT IN ('guest', 'bot')`},
		{"IsNull", IsNull("deleted_at"), `"deleted_at" IS NULL`},
		{"IsNotNull on function", IsNotNull(Fn("NULLIF", "score", 0)), `NULLIF("score", 0) IS NOT NULL`},
		{"Between", Between("age", P(1), P(2)), `"age" BETWEEN $1 AND $2`},
		{"Between mixed bounds", Between(Fn("ABS", "x"), 0, "t.max"), `ABS("x") BETWEEN 0 AND "t"."max"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			sqltesting.AssertSQL(t, render(tt.fragment), tt.want)
		})
	}
}

// TestLogical checks how And, Or, and Not compose, especially where parentheses go.
func TestLogical(t *testing.T) {
	t.Parallel()

	a, b, c, d := Eq("a", 1), Eq("b", 2), Eq("c", 3), Eq("d", 4)
	tests := []struct {
		name     string
		fragment SQLFragment
		want     string
	}{
		{"And", And(a, b, c), `"a" = 1 AND "b" = 2 AND "c" = 3`},
		{"Or", Or(a, b), `"a" = 1 OR "b" = 2`},
		{"single operand has no operator", And(a), `"a" = 1`},
		{"no operands render nothing", Or(), ""},
		{"Or inside And is parenthesized", And(a, Or(b, c)), `"a" = 1 AND ("b" = 2 OR "c" = 3)`},
		{"And inside Or is parenthesized", Or(And(a, b), c), `("a" = 1 AND "b" = 2) OR "c" = 3`},
		// Same-operator nesting is parenthesized too; it is redundant but harmless.
		{"And inside And is parenthesized", And(And(a, b), c), `("a" = 1 AND "b" = 2) AND "c" = 3`},
		{"three levels", And(a, Or(b, And(c, d))), `"a" = 1 AND ("b" = 2 OR ("c" = 3 AND "d" = 4))`},
		{"Not", Not(Eq("active", false)), `NOT ("active" = FALSE)`},
		{"Not wraps a logical expression once", Not(And(a, b)), `NOT ("a" = 1 AND "b" = 2)`},
		// Not already carries its own parentheses, so And adds none.
		{"Not inside And", And(Not(a), b), `NOT ("a" = 1) AND "b" = 2`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			sqltesting.AssertSQL(t, render(tt.fragment), tt.want)
		})
	}
}
