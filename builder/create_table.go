package builder

import (
	"strings"

	"github.com/kotofurumiya/sqbl/dialect"
	"github.com/kotofurumiya/sqbl/syntax"
)

type columnDef struct {
	Name        string
	Typ         string
	Constraints []string // verbatim, e.g. "NOT NULL", "DEFAULT 0"
}

// constraintFunc renders a table-level constraint for the given dialect.
// Using a closure lets each constraint type handle identifier quoting at render time
// rather than at method-call time.
type constraintFunc func(d dialect.SqlDialect) string

// SqlCreateTableBuilder builds a CREATE TABLE statement.
type SqlCreateTableBuilder struct {
	dialect          dialect.SqlDialect
	table            string
	ifNotExists      bool
	columns          []columnDef
	tableConstraints []constraintFunc
}

var _ SqlBuilder = (*SqlCreateTableBuilder)(nil)

// renderSQL assembles the CREATE TABLE statement for the given dialect.
//
// The output follows this structure:
//
//	CREATE TABLE [IF NOT EXISTS] "table" (col_defs..., table_constraints...)
//
// Column definitions are rendered as:
//
//	"name" TYPE [CONSTRAINT ...]
//	e.g.  "id" BIGINT NOT NULL
//	      "score" INTEGER NOT NULL DEFAULT 0
//
// Table-level constraints are rendered after column definitions, separated by ", ":
//
//	PRIMARY KEY ("id")
//	UNIQUE ("email")
//	FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON DELETE CASCADE
//	CHECK (price > 0)
func (b *SqlCreateTableBuilder) renderSQL(d dialect.SqlDialect) string {
	// Header: "CREATE TABLE" or "CREATE TABLE IF NOT EXISTS", plus the quoted table name.
	prefix := "CREATE TABLE "
	if b.ifNotExists {
		prefix += "IF NOT EXISTS "
	}

	// Column defs: one entry per column, e.g. `"id" BIGINT NOT NULL`.
	// Each entry is the quoted column name, the type string, and any
	// verbatim constraint strings (NOT NULL, DEFAULT …) joined by spaces.
	colDefs := mapJoin(b.columns, ", ", func(col columnDef) string {
		parts := append([]string{d.QuoteIdentifier(col.Name), col.Typ}, col.Constraints...)
		return strings.Join(parts, " ")
	})

	// Table-level constraints: each constraintFunc is called with the dialect
	// to produce a rendered string such as `PRIMARY KEY ("id")`.
	constraintDefs := mapJoin(b.tableConstraints, ", ", func(fn constraintFunc) string {
		return fn(d)
	})

	// Body: column defs first, then table constraints.
	// mapJoin returns "" for empty slices, so skip empty strings when joining.
	var bodyParts []string
	if colDefs != "" {
		bodyParts = append(bodyParts, colDefs)
	}
	if constraintDefs != "" {
		bodyParts = append(bodyParts, constraintDefs)
	}

	return prefix + d.QuoteIdentifier(b.table) + " (" + strings.Join(bodyParts, ", ") + ")"
}

// ToSql renders the CREATE TABLE statement with a trailing semicolon.
func (b *SqlCreateTableBuilder) ToSql() string {
	return b.renderSQL(b.dialect) + ";"
}

// ToSqlWithDialect renders the CREATE TABLE statement without a trailing semicolon.
func (b *SqlCreateTableBuilder) ToSqlWithDialect(d dialect.SqlDialect) string {
	return b.renderSQL(d)
}

// Dialect sets the SQL dialect.
func (b *SqlCreateTableBuilder) Dialect(d dialect.SqlDialect) *SqlCreateTableBuilder {
	b2 := *b
	b2.dialect = d
	return &b2
}

// Table sets the target table name.
func (b *SqlCreateTableBuilder) Table(table string) *SqlCreateTableBuilder {
	b2 := *b
	b2.table = table
	return &b2
}

// IfNotExists adds IF NOT EXISTS to the CREATE TABLE statement.
func (b *SqlCreateTableBuilder) IfNotExists() *SqlCreateTableBuilder {
	b2 := *b
	b2.ifNotExists = true
	return &b2
}

// Column adds a column definition.
// constraints are verbatim strings appended after the type, e.g. "NOT NULL", "DEFAULT 0".
//
//	.Column("id",    "BIGINT",  "NOT NULL")
//	.Column("score", "INTEGER", "NOT NULL", "DEFAULT 0")
//	.Column("note",  "TEXT")
func (b *SqlCreateTableBuilder) Column(name, typ string, constraints ...string) *SqlCreateTableBuilder {
	b2 := *b
	b2.columns = append(append([]columnDef(nil), b.columns...), columnDef{Name: name, Typ: typ, Constraints: constraints})
	return &b2
}

// PrimaryKey adds a table-level PRIMARY KEY constraint.
//
//	.PrimaryKey("id")                  →  PRIMARY KEY ("id")
//	.PrimaryKey("order_id", "item_id") →  PRIMARY KEY ("order_id", "item_id")
func (b *SqlCreateTableBuilder) PrimaryKey(cols ...string) *SqlCreateTableBuilder {
	b2 := *b
	b2.tableConstraints = append(append([]constraintFunc(nil), b.tableConstraints...), func(d dialect.SqlDialect) string {
		return "PRIMARY KEY (" + quoteIdentifiers(d, cols) + ")"
	})
	return &b2
}

// Unique adds a table-level UNIQUE constraint.
//
//	.Unique("email")          →  UNIQUE ("email")
//	.Unique("first", "last")  →  UNIQUE ("first", "last")
func (b *SqlCreateTableBuilder) Unique(cols ...string) *SqlCreateTableBuilder {
	b2 := *b
	b2.tableConstraints = append(append([]constraintFunc(nil), b.tableConstraints...), func(d dialect.SqlDialect) string {
		return "UNIQUE (" + quoteIdentifiers(d, cols) + ")"
	})
	return &b2
}

// ForeignKey adds a table-level FOREIGN KEY ... REFERENCES constraint.
// onActions are optional verbatim referential-action clauses, e.g. "ON DELETE CASCADE".
//
//	.ForeignKey([]string{"user_id"}, "users", []string{"id"})
//	  →  FOREIGN KEY ("user_id") REFERENCES "users" ("id")
//
//	.ForeignKey([]string{"user_id"}, "users", []string{"id"}, "ON DELETE CASCADE", "ON UPDATE RESTRICT")
//	  →  FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON DELETE CASCADE ON UPDATE RESTRICT
func (b *SqlCreateTableBuilder) ForeignKey(cols []string, refTable string, refCols []string, onActions ...string) *SqlCreateTableBuilder {
	b2 := *b
	b2.tableConstraints = append(append([]constraintFunc(nil), b.tableConstraints...), func(d dialect.SqlDialect) string {
		fkCols := "FOREIGN KEY (" + quoteIdentifiers(d, cols) + ")"
		references := "REFERENCES " + d.QuoteIdentifier(refTable) + " (" + quoteIdentifiers(d, refCols) + ")"

		parts := []string{fkCols, references}
		parts = append(parts, onActions...)

		return strings.Join(parts, " ")
	})
	return &b2
}

// Check adds a table-level CHECK constraint.
// expr is a dialect-aware Expression, rendered the same way as a WHERE condition.
//
//	.Check(syntax.Gt("price", 0))
//	  →  CHECK ("price" > 0)
//
//	.Check(syntax.And(syntax.Gt("price", 0), syntax.IsNotNull("name")))
//	  →  CHECK ("price" > 0 AND "name" IS NOT NULL)
func (b *SqlCreateTableBuilder) Check(expr syntax.SqlFragment) *SqlCreateTableBuilder {
	b2 := *b
	b2.tableConstraints = append(append([]constraintFunc(nil), b.tableConstraints...), func(d dialect.SqlDialect) string {
		return "CHECK (" + expr.ToSqlWithDialect(d) + ")"
	})
	return &b2
}
