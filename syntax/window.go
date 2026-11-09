package syntax

import (
	"bytes"
	"slices"

	"github.com/kotofurumiya/sqbl/dialect"
)

// WindowExpr represents a window function expression: expr OVER (PARTITION BY ... ORDER BY ... frame).
// Use Over() to create one.
//
//	syntax.Over(syntax.Fn("ROW_NUMBER")).PartitionBy("dept").OrderBy("salary")
//	// ROW_NUMBER() OVER (PARTITION BY "dept" ORDER BY "salary")
type WindowExpr struct {
	expr        SQLFragment
	partitionBy []string
	orderBy     []any
	frame       string // e.g. "ROWS BETWEEN UNBOUNDED PRECEDING AND CURRENT ROW"
}

var _ SQLFragment = WindowExpr{}

// Over wraps an expression (typically a Fn call) with an OVER clause.
//
//	syntax.Over(syntax.Fn("ROW_NUMBER"))
//	syntax.Over(syntax.Fn("SUM", "amount"))
func Over(expr any) WindowExpr {
	return WindowExpr{expr: ToFragment(expr)}
}

// PartitionBy sets the PARTITION BY columns.
//
//	w.PartitionBy("department", "region")
//	// PARTITION BY "department", "region"
func (w WindowExpr) PartitionBy(cols ...string) WindowExpr {
	w.partitionBy = slices.Clone(cols)
	return w
}

// OrderBy sets the ORDER BY columns for the window.
// Accepts column name strings or Order expressions (syntax.Asc / syntax.Desc).
//
//	w.OrderBy("salary")
//	w.OrderBy(syntax.Desc("salary"))
func (w WindowExpr) OrderBy(cols ...any) WindowExpr {
	w.orderBy = slices.Clone(cols)
	return w
}

// Rows sets the frame clause.
//
//	w.Rows("BETWEEN UNBOUNDED PRECEDING AND CURRENT ROW")
func (w WindowExpr) Rows(frame string) WindowExpr {
	w.frame = "ROWS " + frame
	return w
}

// Range sets the RANGE frame clause.
//
//	w.Range("BETWEEN UNBOUNDED PRECEDING AND CURRENT ROW")
func (w WindowExpr) Range(frame string) WindowExpr {
	w.frame = "RANGE " + frame
	return w
}

// AppendSQL implements SQLFragment, writing the window function expression into buf.
func (w WindowExpr) AppendSQL(buf *bytes.Buffer, d dialect.Dialect) {
	w.expr.AppendSQL(buf, d)
	buf.WriteString(" OVER (")
	needSpace := false
	if len(w.partitionBy) > 0 {
		buf.WriteString("PARTITION BY ")
		for i, col := range w.partitionBy {
			if i > 0 {
				buf.WriteString(", ")
			}
			AppendValue(buf, d, col)
		}
		needSpace = true
	}
	if len(w.orderBy) > 0 {
		if needSpace {
			buf.WriteByte(' ')
		}
		buf.WriteString("ORDER BY ")
		for i, col := range w.orderBy {
			if i > 0 {
				buf.WriteString(", ")
			}
			AppendValue(buf, d, col)
		}
		needSpace = true
	}
	if w.frame != "" {
		if needSpace {
			buf.WriteByte(' ')
		}
		buf.WriteString(w.frame)
	}
	buf.WriteByte(')')
}
