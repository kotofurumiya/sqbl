// Package syntax defines the SQL fragments that statement builders are made of:
// identifiers, raw SQL, parameters, function calls, conditions and aliases.
// Everything implements SQLFragment.
//
// Arguments typed any accept these values:
//   - a SQLFragment renders itself;
//   - a plain string is an identifier, quoted by the dialect (use Raw for SQL text);
//   - a bool is the dialect's boolean literal, and nil is NULL;
//   - integer and floating-point kinds, including named types, are numeric literals.
//
// Any other value, such as a named string type, time.Time or []byte, makes the
// statement fail to render with an error: pass it with P and bind it as a query
// argument. Values from users must always be bound this way, never written into SQL.
package syntax
