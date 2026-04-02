package shared

import (
	"context"
	"fmt"
	"testing"
)

// mustExec calls conn.Exec and fatals the test on error.
// msg is included in the failure message for context, e.g. "create table", "insert".
func mustExec(t testing.TB, conn Conn, ctx context.Context, msg, sql string, args ...any) {
	t.Helper()
	if err := conn.Exec(ctx, sql, args...); err != nil {
		t.Fatalf("%s: %v", msg, err)
	}
}

// mustQuery calls conn.Query and fatals the test on error.
// The caller must not call rows.Close() directly; pass the returned Rows to
// collectRows, which closes the cursor after scanning.
func mustQuery(t testing.TB, conn Conn, ctx context.Context, msg, sql string, args ...any) Rows {
	t.Helper()
	rows, err := conn.Query(ctx, sql, args...)
	if err != nil {
		t.Fatalf("%s: %v", msg, err)
	}
	return rows
}

// collectRows scans every row using scan, closes the cursor, and returns the
// collected results. It fatals on scan errors and on rows.Err.
//
// Example:
//
//	type row struct{ name string; price int }
//	rows := mustQuery(t, conn, ctx, "select", sql)
//	got  := collectRows(t, rows, func(r Rows) (row, error) {
//	    var v row
//	    return v, r.Scan(&v.name, &v.price)
//	})
func collectRows[T any](t testing.TB, rows Rows, scan func(Rows) (T, error)) []T {
	t.Helper()
	defer rows.Close()
	var result []T
	for rows.Next() {
		v, err := scan(rows)
		if err != nil {
			t.Fatalf("scan: %v", err)
		}
		result = append(result, v)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("rows err: %v", err)
	}
	return result
}

// assertSlice compares got and want element by element. On a length mismatch
// it fatals immediately; on element mismatches it errors each differing index.
func assertSlice[T comparable](t testing.TB, got, want []T) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("got %d rows; want %d\n  got:  %v\n  want: %v",
			len(got), len(want), got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("[%d] got %v; want %v", i, got[i], want[i])
		}
	}
}

// dropIfExists executes DROP TABLE IF EXISTS for each table, silently ignoring
// errors. It is intended for pre-test cleanup where the table may not yet exist.
func dropIfExists(conn Conn, ctx context.Context, tables ...string) {
	for _, table := range tables {
		_ = conn.Exec(ctx, fmt.Sprintf("DROP TABLE IF EXISTS %s", table))
	}
}
