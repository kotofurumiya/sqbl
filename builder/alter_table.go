package builder

import (
	"bytes"
	"slices"

	"github.com/kotofurumiya/sqbl/dialect"
)

type alterOp struct {
	kind string // "ADD_COLUMN", "DROP_COLUMN", "RENAME_COLUMN", "RENAME_TABLE"
	from string
	to   string
	typ  string
}

// AlterTableBuilder constructs an ALTER TABLE statement using a fluent method chain.
// Use a dialect-specific constructor (e.g. sqblpg.AlterTable) to create an instance.
//
// Each operation method adds an operation after any earlier ones, and the
// operations are rendered comma-separated in call order. PostgreSQL and MySQL
// accept several operations in one statement; SQLite accepts one, and neither
// allows RENAME TO next to other operations.
//
//	sqblpg.AlterTable("users").AddColumn("bio", "TEXT").DropColumn("legacy_col")
//	// ALTER TABLE "users" ADD COLUMN "bio" TEXT, DROP COLUMN "legacy_col"
//
// Note: SQLite supports ADD COLUMN since all versions, but DROP COLUMN and RENAME COLUMN
// require SQLite 3.35.0+ (2021-03-12) and 3.25.0+ (2018-09-15) respectively.
type AlterTableBuilder struct {
	dialect dialect.Dialect
	table   string
	ops     []alterOp
}

var _ Statement = AlterTableBuilder{}

// ToSQL renders the statement with the builder's dialect and a trailing semicolon.
func (b AlterTableBuilder) ToSQL() (string, error) { return render(b, b.dialect, true) }

// ToSQLWithDialect renders the statement with d and no trailing semicolon.
func (b AlterTableBuilder) ToSQLWithDialect(d dialect.Dialect) (string, error) {
	return render(b, d, false)
}

// MustSQL is like ToSQL but panics on error.
func (b AlterTableBuilder) MustSQL() string { return must(b.ToSQL()) }

// MustSQLWithDialect is like ToSQLWithDialect but panics on error.
func (b AlterTableBuilder) MustSQLWithDialect(d dialect.Dialect) string {
	return must(b.ToSQLWithDialect(d))
}

func (b AlterTableBuilder) renderSQL(buf *bytes.Buffer, d dialect.Dialect) {
	if b.table == "" {
		buf.WriteString("ALTER TABLE")
		return
	}

	buf.WriteString("ALTER TABLE ")
	d.QuoteIdentifier(buf, b.table)

	for i, op := range b.ops {
		if i > 0 {
			buf.WriteByte(',')
		}
		switch op.kind {
		case "ADD_COLUMN":
			buf.WriteString(" ADD COLUMN ")
			d.QuoteIdentifier(buf, op.from)
			if op.typ != "" {
				buf.WriteByte(' ')
				buf.WriteString(op.typ)
			}
		case "DROP_COLUMN":
			buf.WriteString(" DROP COLUMN ")
			d.QuoteIdentifier(buf, op.from)
		case "RENAME_COLUMN":
			buf.WriteString(" RENAME COLUMN ")
			d.QuoteIdentifier(buf, op.from)
			buf.WriteString(" TO ")
			d.QuoteIdentifier(buf, op.to)
		case "RENAME_TABLE":
			buf.WriteString(" RENAME TO ")
			d.QuoteIdentifier(buf, op.to)
		}
	}
}

// addOp returns b with op appended to its operations.
func (b AlterTableBuilder) addOp(op alterOp) AlterTableBuilder {
	b.ops = append(slices.Clone(b.ops), op)
	return b
}

// Dialect sets the SQL dialect used when rendering the query.
func (b AlterTableBuilder) Dialect(d dialect.Dialect) AlterTableBuilder {
	b.dialect = d
	return b
}

// Table sets the target table name.
func (b AlterTableBuilder) Table(table string) AlterTableBuilder {
	b.table = table
	return b
}

// AddColumn adds an ADD COLUMN operation. typ is written verbatim and may carry constraints.
//
//	sqblpg.AlterTable("users").AddColumn("bio", "TEXT")
//	// ALTER TABLE "users" ADD COLUMN "bio" TEXT
func (b AlterTableBuilder) AddColumn(name, typ string) AlterTableBuilder {
	return b.addOp(alterOp{kind: "ADD_COLUMN", from: name, typ: typ})
}

// DropColumn adds a DROP COLUMN operation.
// Requires PostgreSQL or MySQL; SQLite 3.35.0+ only.
//
//	sqblpg.AlterTable("users").DropColumn("legacy_col")
//	// ALTER TABLE "users" DROP COLUMN "legacy_col"
func (b AlterTableBuilder) DropColumn(name string) AlterTableBuilder {
	return b.addOp(alterOp{kind: "DROP_COLUMN", from: name})
}

// RenameColumn adds a RENAME COLUMN operation.
// Requires PostgreSQL 9.5+ or MySQL 8.0+; SQLite 3.25.0+ only.
//
//	sqblpg.AlterTable("users").RenameColumn("fullname", "name")
//	// ALTER TABLE "users" RENAME COLUMN "fullname" TO "name"
func (b AlterTableBuilder) RenameColumn(from, to string) AlterTableBuilder {
	return b.addOp(alterOp{kind: "RENAME_COLUMN", from: from, to: to})
}

// RenameTable adds a RENAME TO operation.
//
//	sqblpg.AlterTable("users").RenameTable("members")
//	// ALTER TABLE "users" RENAME TO "members"
func (b AlterTableBuilder) RenameTable(to string) AlterTableBuilder {
	return b.addOp(alterOp{kind: "RENAME_TABLE", to: to})
}
