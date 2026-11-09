// Package sqlerr carries rendering failures from deep inside SQL fragments
// back to the statement builders, which return them as errors.
//
// Fragments render through SQLFragment.AppendSQL, which has no error result,
// so a failure is raised with Fail and recovered by Catch at the top of a render.
package sqlerr

import "fmt"

// Error is a rendering failure caused by input sqbl cannot turn into valid SQL.
type Error struct {
	msg string
}

// Error implements error.
func (e *Error) Error() string { return "sqbl: " + e.msg }

// Fail aborts the current render with an Error.
// It must only be called while rendering under Catch.
func Fail(format string, args ...any) {
	panic(&Error{msg: fmt.Sprintf(format, args...)})
}

// Catch converts a panic raised by Fail into *err.
// Other panics keep propagating. Use it as: defer sqlerr.Catch(&err).
func Catch(err *error) {
	p := recover()
	if p == nil {
		return
	}
	e, ok := p.(*Error)
	if !ok {
		panic(p)
	}
	*err = e
}
