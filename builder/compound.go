package builder

import (
	"bytes"
	"slices"
	"strings"

	"github.com/kotofurumiya/sqbl/dialect"
	"github.com/kotofurumiya/sqbl/syntax"
)

// setOp is one set operation and its right operand.
// op is "UNION", "INTERSECT" or "EXCEPT", optionally followed by " ALL".
type setOp struct {
	op    string
	query Query
}

// CompoundBuilder combines SELECT queries with UNION, INTERSECT and EXCEPT.
// It is created by the set-operation methods of SelectBuilder.
//
// Operations apply left to right, as the method chain reads. OrderBy, Limit
// and Offset apply to the combined rows. Operands are parenthesized where SQL
// would otherwise read them differently: an operand with its own ORDER BY,
// LIMIT, OFFSET, locking or WITH, a compound used as a right operand, and a
// compound continued with INTERSECT, which binds tighter than UNION and EXCEPT.
// SQLite rejects parenthesized operands, so keep such operands out of queries
// meant for SQLite.
//
//	sqblpg.From("users").Select("id").
//	    Union(sqblpg.From("archived_users").Select("id")).
//	    OrderBy("id").
//	    Limit(10)
//	// SELECT "id" FROM "users" UNION SELECT "id" FROM "archived_users" ORDER BY "id" LIMIT 10
type CompoundBuilder struct {
	first     Query
	wrapFirst bool // first is a compound that must keep its own grouping
	ops       []setOp
	orders    []syntax.SQLFragment
	limit     *int
	offset    *int
}

var _ Query = CompoundBuilder{}

// newCompound starts a compound with first as the left operand.
func newCompound(first Query, op string, q Query) CompoundBuilder {
	return CompoundBuilder{first: first, ops: []setOp{{op: op, query: q}}}
}

// ToSQL renders the statement with the leftmost query's dialect and a trailing semicolon.
func (c CompoundBuilder) ToSQL() (string, error) { return render(c, c.dialectOf(), true) }

// ToSQLWithDialect renders the statement with d and no trailing semicolon.
func (c CompoundBuilder) ToSQLWithDialect(d dialect.Dialect) (string, error) {
	return render(c, d, false)
}

// MustSQL is like ToSQL but panics on error.
func (c CompoundBuilder) MustSQL() string { return must(c.ToSQL()) }

// MustSQLWithDialect is like ToSQLWithDialect but panics on error.
func (c CompoundBuilder) MustSQLWithDialect(d dialect.Dialect) string {
	return must(c.ToSQLWithDialect(d))
}

// AppendSQL implements syntax.SQLFragment, writing the compound in parentheses
// so that it can be embedded as a subquery.
func (c CompoundBuilder) AppendSQL(buf *bytes.Buffer, d dialect.Dialect) {
	buf.WriteByte('(')
	c.renderSQL(buf, d)
	buf.WriteByte(')')
}

// IsSubquery implements syntax.Subquery, so IN (compound) renders single parentheses.
func (c CompoundBuilder) IsSubquery() {}

func (c CompoundBuilder) dialectOf() dialect.Dialect {
	if c.first == nil {
		return nil
	}
	return c.first.dialectOf()
}

func (c CompoundBuilder) writeOperand(buf *bytes.Buffer, d dialect.Dialect, right bool) {
	if right || c.hasTrailingClauses() {
		c.AppendSQL(buf, d)
		return
	}
	c.renderSQL(buf, d)
}

func (c CompoundBuilder) renderSQL(buf *bytes.Buffer, d dialect.Dialect) {
	if c.first == nil {
		failFragment("a compound query needs a left operand; start from a SelectBuilder such as sqblpg.From(...).Union(...)").
			AppendSQL(buf, d)
	}
	if c.wrapFirst {
		c.first.AppendSQL(buf, d)
	} else {
		c.first.writeOperand(buf, d, false)
	}
	for _, o := range c.ops {
		buf.WriteByte(' ')
		buf.WriteString(o.op)
		buf.WriteByte(' ')
		o.query.writeOperand(buf, d, true)
	}
	writeOrderLimitOffset(buf, d, c.orders, c.limit, c.offset)
}

func (c CompoundBuilder) hasTrailingClauses() bool {
	return len(c.orders) > 0 || c.limit != nil || c.offset != nil
}

// combine appends a set operation, or nests c as the left operand when appending
// flat would change its meaning.
func (c CompoundBuilder) combine(op string, q Query) CompoundBuilder {
	looser := slices.ContainsFunc(c.ops, func(o setOp) bool { return !strings.HasPrefix(o.op, "INTERSECT") })
	if c.hasTrailingClauses() || (strings.HasPrefix(op, "INTERSECT") && looser) {
		return CompoundBuilder{first: c, wrapFirst: true, ops: []setOp{{op: op, query: q}}}
	}
	c.ops = append(slices.Clone(c.ops), setOp{op: op, query: q})
	return c
}

// Union combines the compound with q by UNION, removing duplicate rows.
func (c CompoundBuilder) Union(q Query) CompoundBuilder { return c.combine("UNION", q) }

// UnionAll combines the compound with q by UNION ALL, keeping duplicate rows.
func (c CompoundBuilder) UnionAll(q Query) CompoundBuilder { return c.combine("UNION ALL", q) }

// Intersect combines the compound with q by INTERSECT, keeping rows present in both.
func (c CompoundBuilder) Intersect(q Query) CompoundBuilder { return c.combine("INTERSECT", q) }

// IntersectAll combines the compound with q by INTERSECT ALL, keeping duplicates.
// Not supported by SQLite or MySQL.
func (c CompoundBuilder) IntersectAll(q Query) CompoundBuilder {
	return c.combine("INTERSECT ALL", q)
}

// Except combines the compound with q by EXCEPT, keeping rows absent from q.
func (c CompoundBuilder) Except(q Query) CompoundBuilder { return c.combine("EXCEPT", q) }

// ExceptAll combines the compound with q by EXCEPT ALL, keeping duplicates.
// Not supported by SQLite or MySQL.
func (c CompoundBuilder) ExceptAll(q Query) CompoundBuilder { return c.combine("EXCEPT ALL", q) }

// OrderBy sets the ORDER BY items of the combined rows, replacing any earlier list.
// Each item follows the rules of syntax.AppendValue; use syntax.Asc or syntax.Desc for a direction.
//
//	c.OrderBy(sqblpg.Desc("id"))
//	// ... UNION SELECT ... ORDER BY "id" DESC
func (c CompoundBuilder) OrderBy(cols ...any) CompoundBuilder {
	c.orders = toFragments(cols)
	return c
}

// Limit sets the maximum number of combined rows to return, replacing any earlier limit.
func (c CompoundBuilder) Limit(n int) CompoundBuilder {
	c.limit = &n
	return c
}

// Offset sets the number of combined rows to skip, replacing any earlier offset.
// OFFSET without LIMIT is accepted by PostgreSQL only.
func (c CompoundBuilder) Offset(n int) CompoundBuilder {
	c.offset = &n
	return c
}
