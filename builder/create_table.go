package builder

import (
	"bytes"

	"github.com/kotofurumiya/sqbl/dialect"
	"github.com/kotofurumiya/sqbl/internal/sqlbuf"
	"github.com/kotofurumiya/sqbl/syntax"
)

type columnDef struct {
	Name        string
	Typ         string
	Constraints []string // verbatim, e.g. "NOT NULL", "DEFAULT 0"
}

// constraintFunc writes a table-level constraint into buf for the given dialect.
// Using a closure lets each constraint type handle identifier quoting at render time
// rather than at method-call time.
type constraintFunc func(buf *bytes.Buffer, d dialect.SqlDialect)

// SqlCreateTableBuilder builds a CREATE TABLE statement.
type SqlCreateTableBuilder struct {
	dialect          dialect.SqlDialect
	table            string
	ifNotExists      bool
	columns          []columnDef
	tableConstraints []constraintFunc
}

var _ SqlBuilder = SqlCreateTableBuilder{}

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
//
// ToSql renders the CREATE TABLE statement with a trailing semicolon.
func (b SqlCreateTableBuilder) ToSql() string {
	buf := sqlbuf.GetStringBuffer()
	b.renderSQL(buf, b.dialect)
	buf.WriteByte(';')
	s := buf.String()
	sqlbuf.PutStringBuffer(buf)
	return s
}

// ToSqlWithDialect renders the CREATE TABLE statement without a trailing semicolon.
func (b SqlCreateTableBuilder) ToSqlWithDialect(d dialect.SqlDialect) string {
	buf := sqlbuf.GetStringBuffer()
	b.renderSQL(buf, d)
	s := buf.String()
	sqlbuf.PutStringBuffer(buf)
	return s
}

func (b SqlCreateTableBuilder) renderSQL(buf *bytes.Buffer, d dialect.SqlDialect) {
	// CREATE TABLE [IF NOT EXISTS] "table" (
	buf.WriteString("CREATE TABLE ")
	if b.ifNotExists {
		buf.WriteString("IF NOT EXISTS ")
	}
	d.QuoteIdentifier(buf, b.table)
	buf.WriteString(" (")

	// Column defs: "name" TYPE [constraint ...], ...
	// Each entry is the quoted column name, the type, and any verbatim
	// constraint strings (NOT NULL, DEFAULT …) separated by spaces.
	first := true
	for _, col := range b.columns {
		if !first {
			buf.WriteString(", ")
		}
		first = false
		d.QuoteIdentifier(buf, col.Name)
		buf.WriteByte(' ')
		buf.WriteString(col.Typ)
		for _, c := range col.Constraints {
			buf.WriteByte(' ')
			buf.WriteString(c)
		}
	}

	// Table-level constraints: PRIMARY KEY (...), UNIQUE (...), FOREIGN KEY (...), CHECK (...)
	// Each constraintFunc is a closure that captures the column names and renders
	// them with the dialect at render time.
	for _, fn := range b.tableConstraints {
		if !first {
			buf.WriteString(", ")
		}
		first = false
		fn(buf, d)
	}

	buf.WriteByte(')')
}

// Dialect sets the SQL dialect.
func (b SqlCreateTableBuilder) Dialect(d dialect.SqlDialect) SqlCreateTableBuilder {
	b.dialect = d
	return b
}

// Table sets the target table name.
func (b SqlCreateTableBuilder) Table(table string) SqlCreateTableBuilder {
	b.table = table
	return b
}

// IfNotExists adds IF NOT EXISTS to the CREATE TABLE statement.
func (b SqlCreateTableBuilder) IfNotExists() SqlCreateTableBuilder {
	b.ifNotExists = true
	return b
}

// Column adds a column definition.
// constraints are verbatim strings appended after the type, e.g. "NOT NULL", "DEFAULT 0".
//
//	.Column("id",    "BIGINT",  "NOT NULL")
//	.Column("score", "INTEGER", "NOT NULL", "DEFAULT 0")
//	.Column("note",  "TEXT")
func (b SqlCreateTableBuilder) Column(name, typ string, constraints ...string) SqlCreateTableBuilder {
	b.columns = append(append([]columnDef(nil), b.columns...), columnDef{Name: name, Typ: typ, Constraints: constraints})
	return b
}

// PrimaryKey adds a table-level PRIMARY KEY constraint.
//
//	.PrimaryKey("id")                  →  PRIMARY KEY ("id")
//	.PrimaryKey("order_id", "item_id") →  PRIMARY KEY ("order_id", "item_id")
func (b SqlCreateTableBuilder) PrimaryKey(cols ...string) SqlCreateTableBuilder {
	b.tableConstraints = append(append([]constraintFunc(nil), b.tableConstraints...), func(buf *bytes.Buffer, d dialect.SqlDialect) {
		buf.WriteString("PRIMARY KEY (")
		quoteIdentifiers(buf, d, cols)
		buf.WriteByte(')')
	})
	return b
}

// Unique adds a table-level UNIQUE constraint.
//
//	.Unique("email")          →  UNIQUE ("email")
//	.Unique("first", "last")  →  UNIQUE ("first", "last")
func (b SqlCreateTableBuilder) Unique(cols ...string) SqlCreateTableBuilder {
	b.tableConstraints = append(append([]constraintFunc(nil), b.tableConstraints...), func(buf *bytes.Buffer, d dialect.SqlDialect) {
		buf.WriteString("UNIQUE (")
		quoteIdentifiers(buf, d, cols)
		buf.WriteByte(')')
	})
	return b
}

// ForeignKey adds a table-level FOREIGN KEY ... REFERENCES constraint.
// onActions are optional verbatim referential-action clauses, e.g. "ON DELETE CASCADE".
//
//	.ForeignKey([]string{"user_id"}, "users", []string{"id"})
//	  →  FOREIGN KEY ("user_id") REFERENCES "users" ("id")
//
//	.ForeignKey([]string{"user_id"}, "users", []string{"id"}, "ON DELETE CASCADE", "ON UPDATE RESTRICT")
//	  →  FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON DELETE CASCADE ON UPDATE RESTRICT
func (b SqlCreateTableBuilder) ForeignKey(cols []string, refTable string, refCols []string, onActions ...string) SqlCreateTableBuilder {
	b.tableConstraints = append(append([]constraintFunc(nil), b.tableConstraints...), func(buf *bytes.Buffer, d dialect.SqlDialect) {
		buf.WriteString("FOREIGN KEY (")
		quoteIdentifiers(buf, d, cols)
		buf.WriteString(") REFERENCES ")
		d.QuoteIdentifier(buf, refTable)
		buf.WriteString(" (")
		quoteIdentifiers(buf, d, refCols)
		buf.WriteByte(')')
		for _, a := range onActions {
			buf.WriteByte(' ')
			buf.WriteString(a)
		}
	})
	return b
}

// Check adds a table-level CHECK constraint.
// expr is a dialect-aware Expression, rendered the same way as a WHERE condition.
//
//	.Check(syntax.Gt("price", 0))
//	  →  CHECK ("price" > 0)
//
//	.Check(syntax.And(syntax.Gt("price", 0), syntax.IsNotNull("name")))
//	  →  CHECK ("price" > 0 AND "name" IS NOT NULL)
func (b SqlCreateTableBuilder) Check(expr syntax.SqlFragment) SqlCreateTableBuilder {
	b.tableConstraints = append(append([]constraintFunc(nil), b.tableConstraints...), func(buf *bytes.Buffer, d dialect.SqlDialect) {
		buf.WriteString("CHECK (")
		expr.AppendSQL(buf, d)
		buf.WriteByte(')')
	})
	return b
}
