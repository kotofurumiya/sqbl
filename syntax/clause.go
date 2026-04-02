// Package syntax defines the low-level AST node types used by SqlSelectBuilder
// to represent SQL clauses, expressions, and fragments.
package syntax

// JoinClause represents a single JOIN entry.
// Kind is one of "INNER", "LEFT", "RIGHT", "FULL OUTER", or "CROSS".
// When Lateral is true, the LATERAL keyword is inserted between the join type and JOIN.
type JoinClause struct {
	Kind    string
	Table   SqlFragment
	On      SqlFragment
	Lateral bool
}
