package builder

import (
	"bytes"
	"fmt"

	"github.com/kotofurumiya/sqbl/dialect"
	"github.com/kotofurumiya/sqbl/internal/sqlbuf"
	"github.com/kotofurumiya/sqbl/syntax"
)

// SqlInsertBuilder constructs a SQL INSERT statement using a fluent method chain.
// Use a dialect-specific constructor (e.g. sqblpg.InsertInto) to create an instance.
//
//	q := sqblpg.InsertInto("users").
//	    Columns("name", "email").
//	    Values(sqbl.P(1), sqbl.P(2))
//	sql := q.ToSql()
type SqlInsertBuilder struct {
	dialect          dialect.SqlDialect
	into             string
	orConflict       string // SQLite: "REPLACE", "IGNORE", "ABORT", "FAIL", "ROLLBACK"
	columns          []string
	rows             []syntax.ValueRow
	conflictTarget   []string // columns for ON CONFLICT (col1, col2); nil means no conflict clause
	conflictAction   string   // "NOTHING" or "UPDATE"
	conflictSets     []string // raw SET fragments for DO UPDATE, e.g. "value = EXCLUDED.value"
	duplicateKeySets []string // raw SET fragments for ON DUPLICATE KEY UPDATE (MySQL)
	returning        []string
}

var _ SqlBuilder = SqlInsertBuilder{}

// ToSql renders the INSERT statement with a trailing semicolon.
func (b SqlInsertBuilder) ToSql() string {
	buf := sqlbuf.GetStringBuffer()
	b.renderSQL(buf, b.dialect)
	buf.WriteByte(';')
	s := buf.String()
	sqlbuf.PutStringBuffer(buf)
	return s
}

// ToSqlWithDialect renders the INSERT statement using the given dialect, without a trailing semicolon.
func (b SqlInsertBuilder) ToSqlWithDialect(d dialect.SqlDialect) string {
	buf := sqlbuf.GetStringBuffer()
	b.renderSQL(buf, d)
	s := buf.String()
	sqlbuf.PutStringBuffer(buf)
	return s
}

func (b SqlInsertBuilder) renderSQL(buf *bytes.Buffer, d dialect.SqlDialect) {
	// INSERT [OR conflict] INTO table
	_, isMysql := d.(*dialect.MysqlDialect)
	_, isSqlite := d.(*dialect.SqliteDialect)

	buf.WriteString("INSERT")
	if isSqlite && b.orConflict != "" {
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
				appendValue(buf, d, v)
			}
			buf.WriteByte(')')
		}
	}

	// ON CONFLICT (PostgreSQL/SQLite) or ON DUPLICATE KEY UPDATE (MySQL)
	if isMysql {
		if len(b.duplicateKeySets) > 0 {
			buf.WriteString(" ON DUPLICATE KEY UPDATE ")
			for i, s := range b.duplicateKeySets {
				if i > 0 {
					buf.WriteString(", ")
				}
				buf.WriteString(s)
			}
		}
	} else {
		switch b.conflictAction {
		case "NOTHING":
			if len(b.conflictTarget) > 0 {
				buf.WriteString(" ON CONFLICT (")
				for i, col := range b.conflictTarget {
					if i > 0 {
						buf.WriteString(", ")
					}
					d.QuoteIdentifier(buf, col)
				}
				buf.WriteString(") DO NOTHING")
			} else {
				buf.WriteString(" ON CONFLICT DO NOTHING")
			}
		case "UPDATE":
			if len(b.conflictSets) > 0 {
				buf.WriteString(" ON CONFLICT")
				if len(b.conflictTarget) > 0 {
					buf.WriteString(" (")
					for i, col := range b.conflictTarget {
						if i > 0 {
							buf.WriteString(", ")
						}
						d.QuoteIdentifier(buf, col)
					}
					buf.WriteByte(')')
				}
				buf.WriteString(" DO UPDATE SET ")
				for i, s := range b.conflictSets {
					if i > 0 {
						buf.WriteString(", ")
					}
					buf.WriteString(s)
				}
			}
		}
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

// Dialect sets the SQL dialect used when rendering the query.
func (b SqlInsertBuilder) Dialect(d dialect.SqlDialect) SqlInsertBuilder {
	b.dialect = d
	return b
}

// Into sets the target table for the INSERT statement.
//
//	b.Into("users")
func (b SqlInsertBuilder) Into(table string) SqlInsertBuilder {
	b.into = table
	return b
}

// Columns sets the column list for the INSERT statement.
// Calling Columns again replaces the previous column list.
//
//	b.Columns("name", "email")
func (b SqlInsertBuilder) Columns(cols ...string) SqlInsertBuilder {
	b.columns = cols
	return b
}

// Values adds a row of values to the INSERT statement.
// Call Values multiple times to insert multiple rows.
//
//	b.Values(sqbl.P(1), sqbl.P(2))
//	b.Values(sqbl.P(1), sqbl.P(2)).Values(sqbl.P(3), sqbl.P(4))
func (b SqlInsertBuilder) Values(vals ...any) SqlInsertBuilder {
	b.rows = append(append([]syntax.ValueRow(nil), b.rows...), syntax.ValueRow(vals))
	return b
}

// OnConflictDoNothing adds an ON CONFLICT DO NOTHING clause.
// Rows that violate a unique or exclusion constraint are silently discarded.
//
//	sqblpg.InsertInto("tags").Columns("name").Values(sqbl.P(1)).OnConflictDoNothing()
func (b SqlInsertBuilder) OnConflictDoNothing() SqlInsertBuilder {
	b.conflictAction = "NOTHING"
	return b
}

// OnConflict sets the conflict target columns and prepares the builder for a
// DO UPDATE clause. Chain DoUpdate immediately after to complete the upsert.
//
//	sqblpg.InsertInto("settings").Columns("user_id", "key", "value").
//	    Values(sqbl.P(1), sqbl.P(2), sqbl.P(3)).
//	    OnConflict("user_id", "key").DoUpdate("value = EXCLUDED.value")
func (b SqlInsertBuilder) OnConflict(cols ...string) SqlInsertBuilder {
	b.conflictTarget = cols
	return b
}

// DoUpdate completes an ON CONFLICT … DO UPDATE SET clause.
// Each argument is a raw SQL assignment fragment, e.g. "value = EXCLUDED.value".
// Must be called after OnConflict.
func (b SqlInsertBuilder) DoUpdate(sets ...string) SqlInsertBuilder {
	b.conflictAction = "UPDATE"
	b.conflictSets = sets
	return b
}

// OnDuplicateKeyUpdate adds an ON DUPLICATE KEY UPDATE clause (MySQL only).
// Each argument is a raw SQL assignment fragment, e.g. "value = VALUES(value)".
//
//	sqblmysql.InsertInto("tags").Columns("name", "count").Values(sqbl.P(1), sqbl.P(2)).
//	    OnDuplicateKeyUpdate("count = VALUES(count)")
func (b SqlInsertBuilder) OnDuplicateKeyUpdate(sets ...string) SqlInsertBuilder {
	b.duplicateKeySets = sets
	return b
}

// OrReplace sets the INSERT OR REPLACE conflict algorithm (SQLite only).
// Existing rows that conflict are deleted and replaced by the new row.
func (b SqlInsertBuilder) OrReplace() SqlInsertBuilder {
	b.orConflict = "REPLACE"
	return b
}

// OrIgnore sets the INSERT OR IGNORE conflict algorithm (SQLite only).
// Rows that violate a constraint are silently discarded.
func (b SqlInsertBuilder) OrIgnore() SqlInsertBuilder {
	b.orConflict = "IGNORE"
	return b
}

// OrAbort sets the INSERT OR ABORT conflict algorithm (SQLite only).
// The current INSERT statement is aborted and changes are rolled back.
func (b SqlInsertBuilder) OrAbort() SqlInsertBuilder {
	b.orConflict = "ABORT"
	return b
}

// OrFail sets the INSERT OR FAIL conflict algorithm (SQLite only).
// The INSERT fails with an error but prior changes in the same transaction are retained.
func (b SqlInsertBuilder) OrFail() SqlInsertBuilder {
	b.orConflict = "FAIL"
	return b
}

// OrRollback sets the INSERT OR ROLLBACK conflict algorithm (SQLite only).
// The entire transaction is rolled back on conflict.
func (b SqlInsertBuilder) OrRollback() SqlInsertBuilder {
	b.orConflict = "ROLLBACK"
	return b
}

// Returning adds a RETURNING clause to the INSERT statement.
// The specified columns are returned for each inserted row.
// Supported by PostgreSQL and SQLite 3.35+. Ignored for MySQL.
//
//	sqblpg.InsertInto("users").Columns("name").Values(sqbl.P(1)).Returning("id", "created_at")
func (b SqlInsertBuilder) Returning(cols ...string) SqlInsertBuilder {
	b.returning = cols
	return b
}

// appendValue writes a single value as SQL into buf using the given dialect.
// Handles Parameter, bool, string, and other types.
func appendValue(buf *bytes.Buffer, d dialect.SqlDialect, v any) {
	switch val := v.(type) {
	case syntax.Parameter:
		val.AppendSQL(buf, d)
	case bool:
		d.Bool(buf, val)
	case string:
		buf.WriteString(val)
	default:
		fmt.Fprintf(buf, "%v", val)
	}
}
