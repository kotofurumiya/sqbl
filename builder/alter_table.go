package builder

import (
	"bytes"

	"github.com/kotofurumiya/sqbl/dialect"
	"github.com/kotofurumiya/sqbl/internal/sqlbuf"
)

type alterOp struct {
	kind string // "ADD_COLUMN", "DROP_COLUMN", "RENAME_COLUMN", "RENAME_TABLE"
	from string
	to   string
	typ  string
}

// SqlAlterTableBuilder constructs an ALTER TABLE statement using a fluent method chain.
// Use a dialect-specific constructor (e.g. sqblpg.AlterTable) to create an instance.
//
//	sqblpg.AlterTable("users").AddColumn("bio", "TEXT")
//	sqblpg.AlterTable("users").RenameColumn("fullname", "name")
//	sqblpg.AlterTable("users").DropColumn("legacy_col")
//
// Note: SQLite supports ADD COLUMN since all versions, but DROP COLUMN and RENAME COLUMN
// require SQLite 3.35.0+ (2021-03-12) and 3.25.0+ (2018-09-15) respectively.
type SqlAlterTableBuilder struct {
	dialect dialect.SqlDialect
	table   string
	op      alterOp
}

var _ SqlBuilder = SqlAlterTableBuilder{}

// ToSql renders the ALTER TABLE statement with a trailing semicolon.
func (b SqlAlterTableBuilder) ToSql() string {
	buf := sqlbuf.GetStringBuffer()
	b.renderSQL(buf, b.dialect)
	buf.WriteByte(';')
	s := buf.String()
	sqlbuf.PutStringBuffer(buf)
	return s
}

// ToSqlWithDialect renders the ALTER TABLE statement using the given dialect, without a trailing semicolon.
func (b SqlAlterTableBuilder) ToSqlWithDialect(d dialect.SqlDialect) string {
	buf := sqlbuf.GetStringBuffer()
	b.renderSQL(buf, d)
	s := buf.String()
	sqlbuf.PutStringBuffer(buf)
	return s
}

func (b SqlAlterTableBuilder) renderSQL(buf *bytes.Buffer, d dialect.SqlDialect) {
	if b.table == "" {
		buf.WriteString("ALTER TABLE")
		return
	}

	buf.WriteString("ALTER TABLE ")
	d.QuoteIdentifier(buf, b.table)

	switch b.op.kind {
	case "ADD_COLUMN":
		buf.WriteString(" ADD COLUMN ")
		d.QuoteIdentifier(buf, b.op.from)
		if b.op.typ != "" {
			buf.WriteByte(' ')
			buf.WriteString(b.op.typ)
		}
	case "DROP_COLUMN":
		buf.WriteString(" DROP COLUMN ")
		d.QuoteIdentifier(buf, b.op.from)
	case "RENAME_COLUMN":
		buf.WriteString(" RENAME COLUMN ")
		d.QuoteIdentifier(buf, b.op.from)
		buf.WriteString(" TO ")
		d.QuoteIdentifier(buf, b.op.to)
	case "RENAME_TABLE":
		buf.WriteString(" RENAME TO ")
		d.QuoteIdentifier(buf, b.op.to)
	}
}

// Dialect sets the SQL dialect used when rendering the query.
func (b SqlAlterTableBuilder) Dialect(d dialect.SqlDialect) SqlAlterTableBuilder {
	b.dialect = d
	return b
}

// Table sets the target table name.
func (b SqlAlterTableBuilder) Table(table string) SqlAlterTableBuilder {
	b.table = table
	return b
}

// AddColumn adds a new column to the table.
//
//	sqblpg.AlterTable("users").AddColumn("bio", "TEXT")
//	// ALTER TABLE "users" ADD COLUMN "bio" TEXT
func (b SqlAlterTableBuilder) AddColumn(name, typ string) SqlAlterTableBuilder {
	b.op = alterOp{kind: "ADD_COLUMN", from: name, typ: typ}
	return b
}

// DropColumn removes a column from the table.
// Requires PostgreSQL or MySQL; SQLite 3.35.0+ only.
//
//	sqblpg.AlterTable("users").DropColumn("legacy_col")
//	// ALTER TABLE "users" DROP COLUMN "legacy_col"
func (b SqlAlterTableBuilder) DropColumn(name string) SqlAlterTableBuilder {
	b.op = alterOp{kind: "DROP_COLUMN", from: name}
	return b
}

// RenameColumn renames a column.
// Requires PostgreSQL 9.5+ or MySQL 8.0+; SQLite 3.25.0+ only.
//
//	sqblpg.AlterTable("users").RenameColumn("fullname", "name")
//	// ALTER TABLE "users" RENAME COLUMN "fullname" TO "name"
func (b SqlAlterTableBuilder) RenameColumn(from, to string) SqlAlterTableBuilder {
	b.op = alterOp{kind: "RENAME_COLUMN", from: from, to: to}
	return b
}

// RenameTable renames the table.
//
//	sqblpg.AlterTable("users").RenameTable("members")
//	// ALTER TABLE "users" RENAME TO "members"
func (b SqlAlterTableBuilder) RenameTable(to string) SqlAlterTableBuilder {
	b.op = alterOp{kind: "RENAME_TABLE", to: to}
	return b
}
