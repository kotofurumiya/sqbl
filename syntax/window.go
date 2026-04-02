package syntax

import (
	"strings"

	"github.com/kotofurumiya/sqbl/dialect"
)

// WindowExpr represents a window function expression: expr OVER (PARTITION BY ... ORDER BY ... frame).
// Use Over() to create one.
//
//	syntax.Over(syntax.Fn("ROW_NUMBER")).PartitionBy("dept").OrderBy("salary")
//	// ROW_NUMBER() OVER (PARTITION BY "dept" ORDER BY "salary")
type WindowExpr struct {
	expr        SqlFragment
	partitionBy []string
	orderBy     []any
	frame       string // e.g. "ROWS BETWEEN UNBOUNDED PRECEDING AND CURRENT ROW"
}

var _ SqlFragment = (*WindowExpr)(nil)

// Over wraps an expression (typically a Fn call) with an OVER clause.
//
//	syntax.Over(syntax.Fn("ROW_NUMBER"))
//	syntax.Over(syntax.Fn("SUM", "amount"))
func Over(expr any) *WindowExpr {
	return &WindowExpr{expr: ToFragment(expr)}
}

// PartitionBy sets the PARTITION BY columns.
//
//	w.PartitionBy("department", "region")
//	// PARTITION BY "department", "region"
func (w *WindowExpr) PartitionBy(cols ...string) *WindowExpr {
	w.partitionBy = cols
	return w
}

// OrderBy sets the ORDER BY columns for the window.
// Accepts column name strings or Order expressions (syntax.Asc / syntax.Desc).
//
//	w.OrderBy("salary")
//	w.OrderBy(syntax.Desc("salary"))
func (w *WindowExpr) OrderBy(cols ...any) *WindowExpr {
	w.orderBy = cols
	return w
}

// Rows sets the frame clause.
//
//	w.Rows("BETWEEN UNBOUNDED PRECEDING AND CURRENT ROW")
func (w *WindowExpr) Rows(frame string) *WindowExpr {
	w.frame = "ROWS " + frame
	return w
}

// Range sets the RANGE frame clause.
//
//	w.Range("BETWEEN UNBOUNDED PRECEDING AND CURRENT ROW")
func (w *WindowExpr) Range(frame string) *WindowExpr {
	w.frame = "RANGE " + frame
	return w
}

// ToSqlWithDialect implements SqlFragment.
func (w *WindowExpr) ToSqlWithDialect(d dialect.SqlDialect) string {
	var windowParts []string

	if len(w.partitionBy) > 0 {
		quoted := make([]string, 0, len(w.partitionBy))
		for _, col := range w.partitionBy {
			quoted = append(quoted, ToFragment(col).ToSqlWithDialect(d))
		}
		windowParts = append(windowParts, "PARTITION BY "+strings.Join(quoted, ", "))
	}

	if len(w.orderBy) > 0 {
		quoted := make([]string, 0, len(w.orderBy))
		for _, col := range w.orderBy {
			quoted = append(quoted, ToFragment(col).ToSqlWithDialect(d))
		}
		windowParts = append(windowParts, "ORDER BY "+strings.Join(quoted, ", "))
	}

	if w.frame != "" {
		windowParts = append(windowParts, w.frame)
	}

	return w.expr.ToSqlWithDialect(d) + " OVER (" + strings.Join(windowParts, " ") + ")"
}
