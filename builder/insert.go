package builder

import (
	"bytes"
	"slices"

	"github.com/kotofurumiya/sqbl/dialect"
	"github.com/kotofurumiya/sqbl/internal/sqlbuf"
	"github.com/kotofurumiya/sqbl/syntax"
)

// InsertBuilder constructs a SQL INSERT statement using a fluent method chain.
// Use a dialect-specific constructor (e.g. sqblpg.InsertInto) to create an instance.
//
//	q := sqblpg.InsertInto("users").
//	    Columns("name", "email").
//	    Values(sqbl.P(1), sqbl.P(2))
//	sql := q.ToSQL()
type InsertBuilder struct {
	dialect          dialect.Dialect
	into             string
	orConflict       string // SQLite: "REPLACE", "IGNORE", "ABORT", "FAIL", "ROLLBACK"
	columns          []string
	rows             []syntax.ValueRow
	conflictTarget   []string             // columns for ON CONFLICT (col1, col2); nil means no conflict clause
	conflictWhere    syntax.SQLFragment   // index predicate after the conflict target (partial unique index)
	conflictAction   string               // "NOTHING" or "UPDATE"
	conflictSets     []syntax.SQLFragment // assignments for DO UPDATE SET
	updateWhere      syntax.SQLFragment   // condition after DO UPDATE SET
	duplicateKeySets []syntax.SQLFragment // assignments for ON DUPLICATE KEY UPDATE (MySQL)
	returning        []string
}

var _ Statement = InsertBuilder{}

// ToSQL renders the INSERT statement with a trailing semicolon.
func (b InsertBuilder) ToSQL() string {
	d := requireDialect(b.dialect)
	buf := sqlbuf.GetStringBuffer()
	b.renderSQL(buf, d)
	buf.WriteByte(';')
	s := buf.String()
	sqlbuf.PutStringBuffer(buf)
	return s
}

// ToSQLWithDialect renders the INSERT statement using the given dialect, without a trailing semicolon.
func (b InsertBuilder) ToSQLWithDialect(d dialect.Dialect) string {
	requireDialect(d)
	buf := sqlbuf.GetStringBuffer()
	b.renderSQL(buf, d)
	s := buf.String()
	sqlbuf.PutStringBuffer(buf)
	return s
}

func (b InsertBuilder) renderSQL(buf *bytes.Buffer, d dialect.Dialect) {
	// INSERT [OR conflict] INTO table
	buf.WriteString("INSERT")
	if b.orConflict != "" {
		buf.WriteString(" OR ")
		buf.WriteString(b.orConflict)
	}
	buf.WriteString(" INTO")
	if b.into != "" {
		buf.WriteByte(' ')
		d.QuoteIdentifier(buf, b.into)
	}

	// (col1, col2, ...)
	if len(b.columns) > 0 {
		buf.WriteString(" (")
		for i, col := range b.columns {
			if i > 0 {
				buf.WriteString(", ")
			}
			d.QuoteIdentifier(buf, col)
		}
		buf.WriteByte(')')
	}

	// VALUES (v1, v2), (v3, v4), ...
	if len(b.rows) > 0 {
		buf.WriteString(" VALUES ")
		for i, row := range b.rows {
			if i > 0 {
				buf.WriteString(", ")
			}
			buf.WriteByte('(')
			for j, v := range row {
				if j > 0 {
					buf.WriteString(", ")
				}
				syntax.AppendValue(buf, d, v)
			}
			buf.WriteByte(')')
		}
	}

	// ON CONFLICT ... and ON DUPLICATE KEY UPDATE ... are independent clauses;
	// each is rendered when its methods were called.
	switch b.conflictAction {
	case "NOTHING":
		b.renderConflictTarget(buf, d)
		buf.WriteString(" DO NOTHING")
	case "UPDATE":
		if len(b.conflictSets) > 0 {
			b.renderConflictTarget(buf, d)
			buf.WriteString(" DO UPDATE SET ")
			appendAssignments(buf, d, b.conflictSets)
			if b.updateWhere != nil {
				buf.WriteString(" WHERE ")
				b.updateWhere.AppendSQL(buf, d)
			}
		}
	}
	if len(b.duplicateKeySets) > 0 {
		buf.WriteString(" ON DUPLICATE KEY UPDATE ")
		appendAssignments(buf, d, b.duplicateKeySets)
	}

	// RETURNING col1, col2, ...
	if len(b.returning) > 0 {
		buf.WriteString(" RETURNING ")
		for i, col := range b.returning {
			if i > 0 {
				buf.WriteString(", ")
			}
			d.QuoteIdentifier(buf, col)
		}
	}
}

// appendAssignments writes comma-separated assignments into buf.
func appendAssignments(buf *bytes.Buffer, d dialect.Dialect, sets []syntax.SQLFragment) {
	for i, s := range sets {
		if i > 0 {
			buf.WriteString(", ")
		}
		s.AppendSQL(buf, d)
	}
}

// rawAssignments wraps raw SQL assignment strings as fragments.
func rawAssignments(sets []string) []syntax.SQLFragment {
	frags := make([]syntax.SQLFragment, len(sets))
	for i, s := range sets {
		frags[i] = syntax.Raw(s)
	}
	return frags
}

// renderConflictTarget writes "ON CONFLICT [(cols) [WHERE pred]]" into buf.
func (b InsertBuilder) renderConflictTarget(buf *bytes.Buffer, d dialect.Dialect) {
	buf.WriteString(" ON CONFLICT")
	if len(b.conflictTarget) == 0 {
		return
	}
	buf.WriteString(" (")
	for i, col := range b.conflictTarget {
		if i > 0 {
			buf.WriteString(", ")
		}
		d.QuoteIdentifier(buf, col)
	}
	buf.WriteByte(')')
	if b.conflictWhere != nil {
		buf.WriteString(" WHERE ")
		b.conflictWhere.AppendSQL(buf, d)
	}
}

// Dialect sets the SQL dialect used when rendering the query.
func (b InsertBuilder) Dialect(d dialect.Dialect) InsertBuilder {
	b.dialect = d
	return b
}

// Into sets the target table for the INSERT statement.
//
//	b.Into("users")
func (b InsertBuilder) Into(table string) InsertBuilder {
	b.into = table
	return b
}

// Columns sets the column list for the INSERT statement.
// Calling Columns again replaces the previous column list.
//
//	b.Columns("name", "email")
func (b InsertBuilder) Columns(cols ...string) InsertBuilder {
	b.columns = slices.Clone(cols)
	return b
}

// Values adds a row of values to the INSERT statement.
// Call Values multiple times to insert multiple rows.
// Each value accepts the same values as syntax.AppendValue.
//
//	b.Values(sqbl.P(1), sqbl.P(2))
//	b.Values(sqbl.P(1), sqbl.P(2)).Values(sqbl.P(3), sqbl.P(4))
func (b InsertBuilder) Values(vals ...any) InsertBuilder {
	b.rows = append(append([]syntax.ValueRow(nil), b.rows...), syntax.ValueRow(slices.Clone(vals)))
	return b
}

// OnConflictDoNothing adds an ON CONFLICT DO NOTHING clause.
// Rows that violate a unique or exclusion constraint are silently discarded.
//
//	sqblpg.InsertInto("tags").Columns("name").Values(sqbl.P(1)).OnConflictDoNothing()
func (b InsertBuilder) OnConflictDoNothing() InsertBuilder {
	b.conflictAction = "NOTHING"
	return b
}

// OnConflict sets the conflict target columns and prepares the builder for a
// DO UPDATE clause. Chain DoUpdateSet or DoUpdate to complete the upsert.
//
//	sqblpg.InsertInto("settings").Columns("user_id", "key", "value").
//	    Values(sqbl.P(1), sqbl.P(2), sqbl.P(3)).
//	    OnConflict("user_id", "key").DoUpdateSet("value", "excluded.value")
func (b InsertBuilder) OnConflict(cols ...string) InsertBuilder {
	b.conflictTarget = slices.Clone(cols)
	return b
}

// DoUpdate adds raw SQL assignments to an ON CONFLICT … DO UPDATE SET clause.
// Each argument is written as-is, e.g. `"value" = excluded."value"`.
// Prefer DoUpdateSet unless the assignment cannot be expressed with it.
// Calling DoUpdate or DoUpdateSet again appends more assignments.
func (b InsertBuilder) DoUpdate(sets ...string) InsertBuilder {
	b.conflictAction = "UPDATE"
	b.conflictSets = append(append([]syntax.SQLFragment(nil), b.conflictSets...), rawAssignments(sets)...)
	return b
}

// DoUpdateSet adds an assignment to an ON CONFLICT … DO UPDATE SET clause.
// The value accepts the same values as syntax.AppendValue, so a plain string
// is an identifier. Call it multiple times to assign multiple columns.
//
// The row proposed for insertion is referred to as excluded. Write it in
// lowercase: the string is quoted as an identifier, and PostgreSQL does not
// match a quoted "EXCLUDED".
//
//	sqblpg.InsertInto("users").Columns("id", "name").Values(sqbl.P(1), sqbl.P(2)).
//	    OnConflict("id").
//	    DoUpdateSet("name", "excluded.name").
//	    DoUpdateSet("updated_at", sqblpg.Fn("NOW"))
//	// ... ON CONFLICT ("id") DO UPDATE SET "name" = "excluded"."name", "updated_at" = NOW()
func (b InsertBuilder) DoUpdateSet(col string, val any) InsertBuilder {
	b.conflictAction = "UPDATE"
	b.conflictSets = append(append([]syntax.SQLFragment(nil), b.conflictSets...), setClause{Col: col, Val: val})
	return b
}

// OnConflictWhere adds an index predicate after the conflict target columns,
// so that a partial unique index can be used as the conflict target.
// It is rendered only when OnConflict has been given at least one column.
// Supported by PostgreSQL and SQLite.
//
//	sqblpg.InsertInto("users").Columns("email", "name").Values(sqbl.P(1), sqbl.P(2)).
//	    OnConflict("email").
//	    OnConflictWhere(sqblpg.Eq("active", 1)).
//	    DoUpdate(`"name" = excluded."name"`)
func (b InsertBuilder) OnConflictWhere(expr syntax.SQLFragment) InsertBuilder {
	b.conflictWhere = expr
	return b
}

// DoUpdateWhere adds a WHERE condition to the DO UPDATE clause.
// Conflicting rows that do not satisfy the condition are left unchanged.
// Supported by PostgreSQL and SQLite.
//
// The row proposed for insertion is referred to as excluded. Write it in
// lowercase when passing it as a plain string: the string is quoted as an
// identifier, and PostgreSQL does not match a quoted "EXCLUDED".
//
//	sqblpg.InsertInto("users").Columns("id", "name", "version").
//	    Values(sqbl.P(1), sqbl.P(2), sqbl.P(3)).
//	    OnConflict("id").
//	    DoUpdate(`"name" = excluded."name"`, `"version" = excluded."version"`).
//	    DoUpdateWhere(sqblpg.Lt("users.version", "excluded.version"))
func (b InsertBuilder) DoUpdateWhere(expr syntax.SQLFragment) InsertBuilder {
	b.updateWhere = expr
	return b
}

// OnDuplicateKeyUpdate adds raw SQL assignments to an ON DUPLICATE KEY UPDATE clause.
// Supported by MySQL and MariaDB.
// Each argument is written as-is, e.g. "`count` = VALUES(`count`)".
// Prefer OnDuplicateKeyUpdateSet unless the assignment cannot be expressed with it.
// Calling OnDuplicateKeyUpdate or OnDuplicateKeyUpdateSet again appends more assignments.
//
//	sqblmysql.InsertInto("tags").Columns("name", "count").Values(sqbl.P(1), sqbl.P(2)).
//	    OnDuplicateKeyUpdate("`count` = VALUES(`count`)")
func (b InsertBuilder) OnDuplicateKeyUpdate(sets ...string) InsertBuilder {
	b.duplicateKeySets = append(append([]syntax.SQLFragment(nil), b.duplicateKeySets...), rawAssignments(sets)...)
	return b
}

// OnDuplicateKeyUpdateSet adds an assignment to an ON DUPLICATE KEY UPDATE clause.
// Supported by MySQL and MariaDB.
// The value accepts the same values as syntax.AppendValue, so a plain string
// is an identifier. Call it multiple times to assign multiple columns.
//
//	sqblmysql.InsertInto("tags").Columns("name", "count").Values(sqbl.P(1), sqbl.P(2)).
//	    OnDuplicateKeyUpdateSet("count", sqblmysql.Fn("VALUES", "count"))
//	// ... ON DUPLICATE KEY UPDATE `count` = VALUES(`count`)
func (b InsertBuilder) OnDuplicateKeyUpdateSet(col string, val any) InsertBuilder {
	b.duplicateKeySets = append(append([]syntax.SQLFragment(nil), b.duplicateKeySets...), setClause{Col: col, Val: val})
	return b
}

// OrReplace sets the INSERT OR REPLACE conflict algorithm. Supported by SQLite.
// Existing rows that conflict are deleted and replaced by the new row.
func (b InsertBuilder) OrReplace() InsertBuilder {
	b.orConflict = "REPLACE"
	return b
}

// OrIgnore sets the INSERT OR IGNORE conflict algorithm. Supported by SQLite.
// Rows that violate a constraint are silently discarded.
func (b InsertBuilder) OrIgnore() InsertBuilder {
	b.orConflict = "IGNORE"
	return b
}

// OrAbort sets the INSERT OR ABORT conflict algorithm. Supported by SQLite.
// The current INSERT statement is aborted and changes are rolled back.
func (b InsertBuilder) OrAbort() InsertBuilder {
	b.orConflict = "ABORT"
	return b
}

// OrFail sets the INSERT OR FAIL conflict algorithm. Supported by SQLite.
// The INSERT fails with an error but prior changes in the same transaction are retained.
func (b InsertBuilder) OrFail() InsertBuilder {
	b.orConflict = "FAIL"
	return b
}

// OrRollback sets the INSERT OR ROLLBACK conflict algorithm. Supported by SQLite.
// The entire transaction is rolled back on conflict.
func (b InsertBuilder) OrRollback() InsertBuilder {
	b.orConflict = "ROLLBACK"
	return b
}

// Returning adds a RETURNING clause to the INSERT statement.
// The specified columns are returned for each inserted row.
// Supported by PostgreSQL, SQLite 3.35+, and MariaDB 10.5+.
//
//	sqblpg.InsertInto("users").Columns("name").Values(sqbl.P(1)).Returning("id", "created_at")
func (b InsertBuilder) Returning(cols ...string) InsertBuilder {
	b.returning = slices.Clone(cols)
	return b
}
