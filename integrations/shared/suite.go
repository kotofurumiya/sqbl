// Package shared provides dialect-agnostic integration test helpers.
// Each Run* function accepts a Conn and a Suite so the same test logic can
// be executed against any supported database by passing the appropriate
// adapter and builder constructors.
package shared

import (
	"context"
	"errors"

	"github.com/kotofurumiya/sqbl/builder"
)

// ErrNoRows is the canonical sentinel returned by Row.Scan when a query
// matched no rows. Each database adapter maps its driver-specific no-rows
// error to this value so shared tests can use errors.Is without importing
// driver packages.
var ErrNoRows = errors.New("no rows in result set")

// Row is a single-row scan abstraction that pgx.Row and database/sql.Row both satisfy.
type Row interface {
	Scan(dest ...any) error
}

// Rows is a multi-row scan abstraction that pgx.Rows and database/sql.Rows both satisfy.
// Close returns no error to match both libraries' signatures.
type Rows interface {
	Next() bool
	Scan(dest ...any) error
	Err() error
	Close()
}

// Conn is a database connection abstraction that can be backed by pgx.Conn,
// database/sql.DB, or any other driver.
type Conn interface {
	Exec(ctx context.Context, sql string, args ...any) error
	QueryRow(ctx context.Context, sql string, args ...any) Row
	Query(ctx context.Context, sql string, args ...any) (Rows, error)
}

// Suite holds the dialect-specific builder constructor functions.
// Populate each field with the corresponding sqblpg / sqblsqlite / sqblmysql function.
type Suite struct {
	From        func(table any) builder.SqlSelectBuilder
	InsertInto  func(table string) builder.SqlInsertBuilder
	Update      func(table string) builder.SqlUpdateBuilder
	DeleteFrom  func(table string) builder.SqlDeleteBuilder
	CreateTable func(table string) builder.SqlCreateTableBuilder
	CreateIndex func(name string) builder.SqlCreateIndexBuilder
	DropTable   func(table string) builder.SqlDropTableBuilder
	DropIndex   func(name string) builder.SqlDropIndexBuilder
	AlterTable  func(table string) builder.SqlAlterTableBuilder
}
