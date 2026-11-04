// Package sandbox runs the Starlark of the format — the predicate of a check, the expressions of a
// template and its solver — in go.starlark.net at the version the format names, in its one
// dialect, with the format's limits of steps and module-kit's clocks, on the caller's goroutine and
// with its context (format/README.md, sections 15, 23 and 24; R88 spec; R11; architecture 3.4).
//
// It holds no code yet: the epic of code brings it (R16).
package sandbox
