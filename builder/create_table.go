package builder

import (
	"bytes"
	"slices"

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
type constraintFunc func(buf *bytes.Buffer, d dialect.Dialect)

// CreateTableBuilder builds a CREATE TABLE statement.
type CreateTableBuilder struct {
	dialect          dialect.Dialect
	table            string
	ifNotExists      bool
	columns          []columnDef
	tableConstraints []constraintFunc
	options          []string
}

var _ Statement = CreateTableBuilder{}

// renderSQL assembles the CREATE TABLE statement for the given dialect.
//
// The output follows this structure:
//
//	CREATE TABLE [IF NOT EXISTS] "table" (col_defs..., table_constraints...) [options...]
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

// ToSQL renders the CREATE TABLE statement with a trailing semicolon.
func (b CreateTableBuilder) ToSQL() string {
	d := requireDialect(b.dialect)
	buf := sqlbuf.GetStringBuffer()
	b.renderSQL(buf, d)
	buf.WriteByte(';')
	s := buf.String()
	sqlbuf.PutStringBuffer(buf)
	return s
}

// ToSQLWithDialect renders the CREATE TABLE statement without a trailing semicolon.
func (b CreateTableBuilder) ToSQLWithDialect(d dialect.Dialect) string {
	requireDialect(d)
	buf := sqlbuf.GetStringBuffer()
	b.renderSQL(buf, d)
	s := buf.String()
	sqlbuf.PutStringBuffer(buf)
	return s
}

func (b CreateTableBuilder) renderSQL(buf *bytes.Buffer, d dialect.Dialect) {
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

	for _, opt := range b.options {
		buf.WriteByte(' ')
		buf.WriteString(opt)
	}
}

// Dialect sets the SQL dialect.
func (b CreateTableBuilder) Dialect(d dialect.Dialect) CreateTableBuilder {
	b.dialect = d
	return b
}

// Table sets the target table name.
func (b CreateTableBuilder) Table(table string) CreateTableBuilder {
	b.table = table
	return b
}

// IfNotExists adds IF NOT EXISTS to the CREATE TABLE statement.
func (b CreateTableBuilder) IfNotExists() CreateTableBuilder {
	b.ifNotExists = true
	return b
}

// Options appends table options after the closing parenthesis.
// Each option is written as-is, separated by a space. Calling Options again appends more options.
// SQLite requires commas between options, so pass them as one string there.
//
//	.Options("STRICT, WITHOUT ROWID")  // SQLite:     ...) STRICT, WITHOUT ROWID
//	.Options("ENGINE=InnoDB")          // MySQL:      ...) ENGINE=InnoDB
//	.Options("WITH (fillfactor = 70)") // PostgreSQL: ...) WITH (fillfactor = 70)
func (b CreateTableBuilder) Options(opts ...string) CreateTableBuilder {
	b.options = append(append([]string(nil), b.options...), opts...)
	return b
}

// Column adds a column definition.
// constraints are verbatim strings appended after the type, e.g. "NOT NULL", "DEFAULT 0".
//
//	.Column("id",    "BIGINT",  "NOT NULL")
//	.Column("score", "INTEGER", "NOT NULL", "DEFAULT 0")
//	.Column("note",  "TEXT")
func (b CreateTableBuilder) Column(name, typ string, constraints ...string) CreateTableBuilder {
	b.columns = append(append([]columnDef(nil), b.columns...), columnDef{Name: name, Typ: typ, Constraints: slices.Clone(constraints)})
	return b
}

// PrimaryKey adds a table-level PRIMARY KEY constraint.
//
//	.PrimaryKey("id")                  →  PRIMARY KEY ("id")
//	.PrimaryKey("order_id", "item_id") →  PRIMARY KEY ("order_id", "item_id")
func (b CreateTableBuilder) PrimaryKey(cols ...string) CreateTableBuilder {
	cols = slices.Clone(cols)
	b.tableConstraints = append(append([]constraintFunc(nil), b.tableConstraints...), func(buf *bytes.Buffer, d dialect.Dialect) {
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
func (b CreateTableBuilder) Unique(cols ...string) CreateTableBuilder {
	cols = slices.Clone(cols)
	b.tableConstraints = append(append([]constraintFunc(nil), b.tableConstraints...), func(buf *bytes.Buffer, d dialect.Dialect) {
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
func (b CreateTableBuilder) ForeignKey(cols []string, refTable string, refCols []string, onActions ...string) CreateTableBuilder {
	cols, refCols, onActions = slices.Clone(cols), slices.Clone(refCols), slices.Clone(onActions)
	b.tableConstraints = append(append([]constraintFunc(nil), b.tableConstraints...), func(buf *bytes.Buffer, d dialect.Dialect) {
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
func (b CreateTableBuilder) Check(expr syntax.SQLFragment) CreateTableBuilder {
	b.tableConstraints = append(append([]constraintFunc(nil), b.tableConstraints...), func(buf *bytes.Buffer, d dialect.Dialect) {
		buf.WriteString("CHECK (")
		expr.AppendSQL(buf, d)
		buf.WriteByte(')')
	})
	return b
}
