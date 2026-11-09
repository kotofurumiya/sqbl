package shared

import (
	"context"
	"fmt"
	"slices"
	"testing"
)

// MustExec executes sql and fails the test immediately on error.
func MustExec(t testing.TB, conn Conn, sql string, args ...any) {
	t.Helper()
	if err := conn.Exec(t.Context(), sql, args...); err != nil {
		t.Fatalf("exec %s: %v", sql, err)
	}
}

// MustFail executes sql and reports a test error unless the database rejects it.
// why describes the rule that should have rejected the statement.
func MustFail(t testing.TB, conn Conn, why, sql string, args ...any) {
	t.Helper()
	if err := conn.Exec(t.Context(), sql, args...); err == nil {
		t.Errorf("%s: statement succeeded; want an error\n  sql: %s", why, sql)
	}
}

// MustScan scans the single row of sql into dest and fails the test on error,
// including when no row matched.
func MustScan(t testing.TB, conn Conn, sql string, args []any, dest ...any) {
	t.Helper()
	if err := conn.QueryRow(t.Context(), sql, args...).Scan(dest...); err != nil {
		t.Fatalf("query row %s: %v", sql, err)
	}
}

// QueryRows runs sql and returns every row converted by scan.
func QueryRows[T any](t testing.TB, conn Conn, scan func(Rows) (T, error), sql string, args ...any) []T {
	t.Helper()
	rows, err := conn.Query(t.Context(), sql, args...)
	if err != nil {
		t.Fatalf("query %s: %v", sql, err)
	}
	defer rows.Close()
	var result []T
	for rows.Next() {
		v, err := scan(rows)
		if err != nil {
			t.Fatalf("scan %s: %v", sql, err)
		}
		result = append(result, v)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("rows %s: %v", sql, err)
	}
	return result
}

// QueryColumn runs a single-column query and returns its values in row order.
func QueryColumn[T any](t testing.TB, conn Conn, sql string, args ...any) []T {
	t.Helper()
	return QueryRows(t, conn, func(r Rows) (T, error) {
		var v T
		return v, r.Scan(&v)
	}, sql, args...)
}

// AssertSlice reports a test error unless got and want hold the same elements
// in the same order.
func AssertSlice[T comparable](t testing.TB, got, want []T) {
	t.Helper()
	if !slices.Equal(got, want) {
		t.Errorf("rows mismatch\n got: %v\nwant: %v", got, want)
	}
}

// AssertCount reports a test error unless table holds exactly want rows.
// It is used to prove that a statement neither dropped nor added rows.
func AssertCount(t testing.TB, conn Conn, table string, want int) {
	t.Helper()
	var got int
	MustScan(t, conn, fmt.Sprintf("SELECT COUNT(*) FROM %s", table), nil, &got)
	if got != want {
		t.Errorf("COUNT(*) FROM %s = %d; want %d", table, got, want)
	}
}

// UseTables drops the given persistent tables now and again when the test ends.
// Scenarios that cannot use TEMPORARY tables (for example, because they create
// indexes or foreign keys) rely on it to start clean even after a crashed run.
// List referencing tables before the tables they reference.
func UseTables(t testing.TB, conn Conn, tables ...string) {
	t.Helper()
	drop := func(ctx context.Context) {
		for _, table := range tables {
			_ = conn.Exec(ctx, fmt.Sprintf("DROP TABLE IF EXISTS %s", table))
		}
	}
	drop(t.Context())
	// t.Context() is already canceled when cleanups run.
	t.Cleanup(func() { drop(context.Background()) })
}
