package builder

import (
	"bytes"
	"slices"

	"github.com/kotofurumiya/sqbl/dialect"
	"github.com/kotofurumiya/sqbl/syntax"
)

// InsertBuilder constructs a SQL INSERT statement using a fluent method chain.
// Use a dialect-specific constructor (e.g. sqblpg.InsertInto) to create an instance.
//
//	q := sqblpg.InsertInto("users").
//	    Columns("name", "email").
//	    Values(sqblpg.P(1), sqblpg.P(2))
//	sql, err := q.ToSQL()
//	// INSERT INTO "users" ("name", "email") VALUES ($1, $2);
type InsertBuilder struct {
	dialect          dialect.Dialect
	into             string
	orConflict       string // SQLite: "REPLACE", "IGNORE", "ABORT", "FAIL", "ROLLBACK"
	columns          []string
	rows             [][]any              // VALUES rows
	conflictTarget   []string             // columns for ON CONFLICT (col1, col2)
	conflictWhere    []syntax.SQLFragment // index predicate after the conflict target, ANDed
	doNothing        bool                 // ON CONFLICT ... DO NOTHING
	conflictSets     []syntax.SQLFragment // assignments for DO UPDATE SET
	updateWhere      []syntax.SQLFragment // condition after DO UPDATE SET, ANDed
	duplicateKeySets []syntax.SQLFragment // assignments for ON DUPLICATE KEY UPDATE (MySQL)
	returning        []any
}

var _ Statement = InsertBuilder{}

// ToSQL renders the statement with the builder's dialect and a trailing semicolon.
func (b InsertBuilder) ToSQL() (string, error) { return render(b, b.dialect, true) }

// ToSQLWithDialect renders the statement with d and no trailing semicolon.
func (b InsertBuilder) ToSQLWithDialect(d dialect.Dialect) (string, error) {
	return render(b, d, false)
}

// MustSQL is like ToSQL but panics on error.
func (b InsertBuilder) MustSQL() string { return must(b.ToSQL()) }

// MustSQLWithDialect is like ToSQLWithDialect but panics on error.
func (b InsertBuilder) MustSQLWithDialect(d dialect.Dialect) string {
	return must(b.ToSQLWithDialect(d))
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
			writeValues(buf, d, row)
			buf.WriteByte(')')
		}
	}

	b.renderOnConflict(buf, d)
	if len(b.duplicateKeySets) > 0 {
		buf.WriteString(" ON DUPLICATE KEY UPDATE ")
		appendAssignments(buf, d, b.duplicateKeySets)
	}

	if len(b.returning) > 0 {
		buf.WriteString(" RETURNING ")
		writeValues(buf, d, b.returning)
	}
}

// renderOnConflict writes the ON CONFLICT clause, or nothing when none was requested.
// A clause that was started but cannot be completed fails the render rather
// than being dropped, since silently ignoring it would change what the statement does.
func (b InsertBuilder) renderOnConflict(buf *bytes.Buffer, d dialect.Dialect) {
	doUpdate := len(b.conflictSets) > 0
	switch {
	case b.doNothing && doUpdate:
		failFragment("OnConflictDoNothing cannot be combined with DoUpdate or DoUpdateSet; keep one action").AppendSQL(buf, d)
	case len(b.conflictWhere) > 0 && len(b.conflictTarget) == 0:
		failFragment("OnConflictWhere needs conflict target columns; add OnConflict").AppendSQL(buf, d)
	case len(b.updateWhere) > 0 && !doUpdate:
		failFragment("DoUpdateWhere needs an update; add DoUpdate or DoUpdateSet").AppendSQL(buf, d)
	case !b.doNothing && !doUpdate:
		if len(b.conflictTarget) > 0 {
			failFragment("OnConflict needs an action; add OnConflictDoNothing, DoUpdate or DoUpdateSet").AppendSQL(buf, d)
		}
		return
	}

	buf.WriteString(" ON CONFLICT")
	if len(b.conflictTarget) > 0 {
		buf.WriteString(" (")
		quoteIdentifiers(buf, d, b.conflictTarget)
		buf.WriteByte(')')
		writeConditions(buf, d, " WHERE ", b.conflictWhere)
	}
	if b.doNothing {
		buf.WriteString(" DO NOTHING")
		return
	}
	buf.WriteString(" DO UPDATE SET ")
	appendAssignments(buf, d, b.conflictSets)
	writeConditions(buf, d, " WHERE ", b.updateWhere)
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

// Dialect sets the SQL dialect used when rendering the query.
func (b InsertBuilder) Dialect(d dialect.Dialect) InsertBuilder {
	b.dialect = d
	return b
}

// Into sets the target table, replacing any earlier one.
//
//	b.Into("users")
func (b InsertBuilder) Into(table string) InsertBuilder {
	b.into = table
	return b
}

// Columns sets the column list, replacing any earlier list.
//
//	b.Columns("name", "email")
//	// INSERT INTO "users" ("name", "email") ...
func (b InsertBuilder) Columns(cols ...string) InsertBuilder {
	b.columns = slices.Clone(cols)
	return b
}

// Values adds a row of values after any earlier rows.
// Each value follows the rules of syntax.AppendValue: a plain string is an
// identifier, so pass data with P.
//
//	b.Values(sqblpg.P(1), sqblpg.P(2)).Values(sqblpg.P(3), sqblpg.P(4))
//	// VALUES ($1, $2), ($3, $4)
func (b InsertBuilder) Values(vals ...any) InsertBuilder {
	b.rows = append(slices.Clone(b.rows), slices.Clone(vals))
	return b
}

// OnConflictDoNothing adds ON CONFLICT ... DO NOTHING. Supported by PostgreSQL and SQLite.
// Rows that violate a unique or exclusion constraint are discarded.
// Rendering fails when it is combined with DoUpdate or DoUpdateSet.
//
//	sqblpg.InsertInto("tags").Columns("name").Values(sqblpg.P(1)).OnConflictDoNothing()
//	// INSERT INTO "tags" ("name") VALUES ($1) ON CONFLICT DO NOTHING
func (b InsertBuilder) OnConflictDoNothing() InsertBuilder {
	b.doNothing = true
	return b
}

// OnConflict sets the conflict target columns, replacing any earlier list.
// Supported by PostgreSQL and SQLite. It needs an action: rendering fails unless
// OnConflictDoNothing, DoUpdate or DoUpdateSet is also called.
//
//	sqblpg.InsertInto("settings").Columns("user_id", "key", "value").
//	    Values(sqblpg.P(1), sqblpg.P(2), sqblpg.P(3)).
//	    OnConflict("user_id", "key").DoUpdateSet("value", "excluded.value")
//	// ... ON CONFLICT ("user_id", "key") DO UPDATE SET "value" = "excluded"."value"
func (b InsertBuilder) OnConflict(cols ...string) InsertBuilder {
	b.conflictTarget = slices.Clone(cols)
	return b
}

// DoUpdate adds raw SQL assignments to ON CONFLICT ... DO UPDATE SET, after any earlier ones.
// Each argument is written as-is, e.g. `"value" = excluded."value"`; never build
// it from untrusted input. Prefer DoUpdateSet unless the assignment cannot be expressed with it.
func (b InsertBuilder) DoUpdate(sets ...string) InsertBuilder {
	b.conflictSets = append(slices.Clone(b.conflictSets), rawAssignments(sets)...)
	return b
}

// DoUpdateSet adds an assignment to ON CONFLICT ... DO UPDATE SET, after any earlier ones.
// The value follows the rules of syntax.AppendValue, so a plain string is an identifier.
//
// The row proposed for insertion is referred to as excluded. Write it in
// lowercase: the string is quoted as an identifier, and PostgreSQL does not
// match a quoted "EXCLUDED".
//
//	sqblpg.InsertInto("users").Columns("id", "name").Values(sqblpg.P(1), sqblpg.P(2)).
//	    OnConflict("id").
//	    DoUpdateSet("name", "excluded.name").
//	    DoUpdateSet("updated_at", sqblpg.Fn("NOW"))
//	// ... ON CONFLICT ("id") DO UPDATE SET "name" = "excluded"."name", "updated_at" = NOW()
func (b InsertBuilder) DoUpdateSet(col string, val any) InsertBuilder {
	b.conflictSets = append(slices.Clone(b.conflictSets), setClause{Col: col, Val: val})
	return b
}

// OnConflictWhere adds an index predicate after the conflict target columns,
// so that a partial unique index can serve as the conflict target.
// Calling it again ANDs the new predicate with the earlier ones; nil is ignored.
// Rendering fails unless OnConflict gives target columns. Supported by PostgreSQL and SQLite.
//
//	sqblpg.InsertInto("users").Columns("email", "name").Values(sqblpg.P(1), sqblpg.P(2)).
//	    OnConflict("email").
//	    OnConflictWhere(sqblpg.Eq("active", true)).
//	    DoUpdateSet("name", "excluded.name")
//	// ... ON CONFLICT ("email") WHERE "active" = TRUE DO UPDATE SET "name" = "excluded"."name"
func (b InsertBuilder) OnConflictWhere(expr syntax.SQLFragment) InsertBuilder {
	b.conflictWhere = addCondition(b.conflictWhere, expr)
	return b
}

// DoUpdateWhere adds a condition to the DO UPDATE clause; conflicting rows that
// do not satisfy it are left unchanged. Calling it again ANDs the new condition
// with the earlier ones; nil is ignored. Rendering fails unless DoUpdate or
// DoUpdateSet is also called. Supported by PostgreSQL and SQLite.
//
//	sqblpg.InsertInto("users").Columns("id", "name", "version").
//	    Values(sqblpg.P(1), sqblpg.P(2), sqblpg.P(3)).
//	    OnConflict("id").
//	    DoUpdateSet("name", "excluded.name").
//	    DoUpdateWhere(sqblpg.Lt("users.version", "excluded.version"))
//	// ... DO UPDATE SET "name" = "excluded"."name" WHERE "users"."version" < "excluded"."version"
func (b InsertBuilder) DoUpdateWhere(expr syntax.SQLFragment) InsertBuilder {
	b.updateWhere = addCondition(b.updateWhere, expr)
	return b
}

// OnDuplicateKeyUpdate adds raw SQL assignments to ON DUPLICATE KEY UPDATE, after any earlier ones.
// Supported by MySQL and MariaDB. Each argument is written as-is, e.g. "`count` = VALUES(`count`)";
// never build it from untrusted input. Prefer OnDuplicateKeyUpdateSet unless the
// assignment cannot be expressed with it.
func (b InsertBuilder) OnDuplicateKeyUpdate(sets ...string) InsertBuilder {
	b.duplicateKeySets = append(slices.Clone(b.duplicateKeySets), rawAssignments(sets)...)
	return b
}

// OnDuplicateKeyUpdateSet adds an assignment to ON DUPLICATE KEY UPDATE, after any earlier ones.
// Supported by MySQL and MariaDB. The value follows the rules of syntax.AppendValue,
// so a plain string is an identifier.
//
//	sqblmysql.InsertInto("tags").Columns("name", "count").Values(sqblmysql.P(), sqblmysql.P()).
//	    OnDuplicateKeyUpdateSet("count", sqblmysql.Fn("VALUES", "count"))
//	// ... ON DUPLICATE KEY UPDATE `count` = VALUES(`count`)
func (b InsertBuilder) OnDuplicateKeyUpdateSet(col string, val any) InsertBuilder {
	b.duplicateKeySets = append(slices.Clone(b.duplicateKeySets), setClause{Col: col, Val: val})
	return b
}

// OrReplace sets the INSERT OR REPLACE conflict algorithm, replacing any earlier one. Supported by SQLite.
// Existing rows that conflict are deleted and replaced by the new row.
func (b InsertBuilder) OrReplace() InsertBuilder {
	b.orConflict = "REPLACE"
	return b
}

// OrIgnore sets the INSERT OR IGNORE conflict algorithm, replacing any earlier one. Supported by SQLite.
// Rows that violate a constraint are discarded.
func (b InsertBuilder) OrIgnore() InsertBuilder {
	b.orConflict = "IGNORE"
	return b
}

// OrAbort sets the INSERT OR ABORT conflict algorithm, replacing any earlier one. Supported by SQLite.
// The statement is aborted and its changes are rolled back.
func (b InsertBuilder) OrAbort() InsertBuilder {
	b.orConflict = "ABORT"
	return b
}

// OrFail sets the INSERT OR FAIL conflict algorithm, replacing any earlier one. Supported by SQLite.
// The statement fails, keeping the changes it made before the conflict.
func (b InsertBuilder) OrFail() InsertBuilder {
	b.orConflict = "FAIL"
	return b
}

// OrRollback sets the INSERT OR ROLLBACK conflict algorithm, replacing any earlier one. Supported by SQLite.
// The whole transaction is rolled back on conflict.
func (b InsertBuilder) OrRollback() InsertBuilder {
	b.orConflict = "ROLLBACK"
	return b
}

// Returning sets the RETURNING list, replacing any earlier list.
// Each item follows the rules of syntax.AppendValue, so "*" returns every column.
// Supported by PostgreSQL, SQLite 3.35+, and MariaDB 10.5+.
//
//	sqblpg.InsertInto("users").Columns("name").Values(sqblpg.P(1)).Returning("id", "created_at")
//	// INSERT INTO "users" ("name") VALUES ($1) RETURNING "id", "created_at"
func (b InsertBuilder) Returning(cols ...any) InsertBuilder {
	b.returning = slices.Clone(cols)
	return b
}
