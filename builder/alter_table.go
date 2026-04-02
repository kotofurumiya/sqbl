package builder

import (
	"strings"

	"github.com/kotofurumiya/sqbl/dialect"
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

var _ SqlBuilder = (*SqlAlterTableBuilder)(nil)

// ToSql renders the ALTER TABLE statement with a trailing semicolon.
func (b *SqlAlterTableBuilder) ToSql() string {
	return b.renderSQL(b.dialect) + ";"
}

// ToSqlWithDialect renders the ALTER TABLE statement using the given dialect, without a trailing semicolon.
func (b *SqlAlterTableBuilder) ToSqlWithDialect(d dialect.SqlDialect) string {
	return b.renderSQL(d)
}

func (b *SqlAlterTableBuilder) renderSQL(d dialect.SqlDialect) string {
	if b.table == "" {
		return "ALTER TABLE"
	}

	base := "ALTER TABLE " + d.QuoteIdentifier(b.table)

	var parts []string
	switch b.op.kind {
	case "ADD_COLUMN":
		col := d.QuoteIdentifier(b.op.from)
		if b.op.typ != "" {
			col += " " + b.op.typ
		}
		parts = append(parts, base+" ADD COLUMN "+col)
	case "DROP_COLUMN":
		parts = append(parts, base+" DROP COLUMN "+d.QuoteIdentifier(b.op.from))
	case "RENAME_COLUMN":
		parts = append(parts, base+" RENAME COLUMN "+d.QuoteIdentifier(b.op.from)+" TO "+d.QuoteIdentifier(b.op.to))
	case "RENAME_TABLE":
		parts = append(parts, base+" RENAME TO "+d.QuoteIdentifier(b.op.to))
	default:
		parts = append(parts, base)
	}

	return strings.Join(parts, " ")
}

// Dialect sets the SQL dialect used when rendering the query.
func (b *SqlAlterTableBuilder) Dialect(d dialect.SqlDialect) *SqlAlterTableBuilder {
	b2 := *b
	b2.dialect = d
	return &b2
}

// Table sets the target table name.
func (b *SqlAlterTableBuilder) Table(table string) *SqlAlterTableBuilder {
	b2 := *b
	b2.table = table
	return &b2
}

// AddColumn adds a new column to the table.
//
//	sqblpg.AlterTable("users").AddColumn("bio", "TEXT")
//	// ALTER TABLE "users" ADD COLUMN "bio" TEXT
func (b *SqlAlterTableBuilder) AddColumn(name, typ string) *SqlAlterTableBuilder {
	b2 := *b
	b2.op = alterOp{kind: "ADD_COLUMN", from: name, typ: typ}
	return &b2
}

// DropColumn removes a column from the table.
// Requires PostgreSQL or MySQL; SQLite 3.35.0+ only.
//
//	sqblpg.AlterTable("users").DropColumn("legacy_col")
//	// ALTER TABLE "users" DROP COLUMN "legacy_col"
func (b *SqlAlterTableBuilder) DropColumn(name string) *SqlAlterTableBuilder {
	b2 := *b
	b2.op = alterOp{kind: "DROP_COLUMN", from: name}
	return &b2
}

// RenameColumn renames a column.
// Requires PostgreSQL 9.5+ or MySQL 8.0+; SQLite 3.25.0+ only.
//
//	sqblpg.AlterTable("users").RenameColumn("fullname", "name")
//	// ALTER TABLE "users" RENAME COLUMN "fullname" TO "name"
func (b *SqlAlterTableBuilder) RenameColumn(from, to string) *SqlAlterTableBuilder {
	b2 := *b
	b2.op = alterOp{kind: "RENAME_COLUMN", from: from, to: to}
	return &b2
}

// RenameTable renames the table.
//
//	sqblpg.AlterTable("users").RenameTable("members")
//	// ALTER TABLE "users" RENAME TO "members"
func (b *SqlAlterTableBuilder) RenameTable(to string) *SqlAlterTableBuilder {
	b2 := *b
	b2.op = alterOp{kind: "RENAME_TABLE", to: to}
	return &b2
}
