// Package shared holds dialect-agnostic integration scenarios.
// Each database package adapts its driver to Conn, describes its builders and
// capabilities in a Suite, and calls RunAll; the same scenarios then prove
// that every supported database accepts the generated SQL and returns the
// expected rows.
package shared

import (
	"context"
	"errors"
	"testing"

	"github.com/kotofurumiya/sqbl/builder"
)

// ErrNoRows is returned by Row.Scan when a query matched no rows.
// Each adapter maps its driver-specific sentinel to this value.
var ErrNoRows = errors.New("no rows in result set")

// Row is a single-row result.
type Row interface {
	Scan(dest ...any) error
}

// Rows is a multi-row result cursor.
type Rows interface {
	Next() bool
	Scan(dest ...any) error
	Err() error
	Close()
}

// Conn is the minimal database connection the scenarios need.
type Conn interface {
	Exec(ctx context.Context, sql string, args ...any) error
	QueryRow(ctx context.Context, sql string, args ...any) Row
	Query(ctx context.Context, sql string, args ...any) (Rows, error)
}

// Suite holds a dialect's builder constructors and the database capabilities
// that decide which scenarios run. sqbl renders whatever it is asked for, so
// scenarios consult these flags instead of relying on the builder to refuse.
type Suite struct {
	From        func(table any) builder.SelectBuilder
	InsertInto  func(table string) builder.InsertBuilder
	Update      func(table string) builder.UpdateBuilder
	DeleteFrom  func(table string) builder.DeleteBuilder
	CreateTable func(table string) builder.CreateTableBuilder
	CreateIndex func(name string) builder.CreateIndexBuilder
	DropTable   func(table string) builder.DropTableBuilder
	DropIndex   func(name string) builder.DropIndexBuilder
	AlterTable  func(table string) builder.AlterTableBuilder

	// DropIndexOnTable is set when DROP INDEX must name its table (MySQL).
	DropIndexOnTable bool
	// PositionalParams is set when the driver accepts syntax.P() as "?".
	PositionalParams bool
	// IndexedParams is set when syntax.P(n) renders a numbered placeholder that may be
	// reused and given out of order.
	IndexedParams bool
	// NamedParams is set when the driver binds syntax.P(":name") to sql.Named args.
	NamedParams bool
	// FullJoin is set when FULL OUTER JOIN is supported.
	FullJoin bool
	// SetOpAll is set when INTERSECT ALL and EXCEPT ALL are supported.
	SetOpAll bool
	// Returning is set when INSERT, UPDATE and DELETE accept RETURNING.
	Returning bool
	// OnConflict is set when INSERT accepts ON CONFLICT clauses.
	OnConflict bool
	// PartialIndex is set when CREATE INDEX accepts a WHERE clause.
	PartialIndex bool
}

// scenarios lists every shared scenario; RunAll runs each as a subtest.
var scenarios = []struct {
	name string
	run  func(t *testing.T, conn Conn, s Suite)
}{
	{"CRUD", runCRUD},
	{"Params", runParams},
	{"Filtering", runFiltering},
	{"Joins", runJoins},
	{"Subqueries", runSubqueries},
	{"CTE", runCTE},
	{"RecursiveCTE", runRecursiveCTE},
	{"SetOperations", runSetOperations},
	{"WindowFunctions", runWindowFunctions},
	{"Identifiers", runIdentifiers},
	{"InjectionValues", runInjectionValues},
	{"InjectionIdentifiers", runInjectionIdentifiers},
	{"CreateTable", runCreateTable},
	{"CreateIndex", runCreateIndex},
	{"SchemaEvolution", runSchemaEvolution},
	{"Returning", runReturning},
	{"OnConflict", runOnConflict},
}

// RunAll runs every shared scenario as a parallel subtest.
// open must return a fresh connection per call, so that temporary tables and
// session state never leak between scenarios.
func RunAll(t *testing.T, open func(t *testing.T) Conn, s Suite) {
	for _, sc := range scenarios {
		t.Run(sc.name, func(t *testing.T) {
			t.Parallel()
			sc.run(t, open(t), s)
		})
	}
}
