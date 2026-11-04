// Package sandbox runs the Starlark of the format — the predicate of a check, the expressions of a
// template and its solver — in go.starlark.net at the version the format names, in its one
// dialect, with the format's limits of steps and module-kit's clocks, on the caller's goroutine and
// with its context (format/README.md, sections 15, 23 and 24; R88 spec; R11; architecture 3.4).
//
// It holds so far what a predicate takes: CheckExpr checks an expression of the dialect and the
// names it may take, and Eval evaluates one on a thread of its own under Limits, of which Predicate
// are a predicate's, stopping it with ErrSteps, ErrClock or ErrPanic, or with the error of a
// context that is done (R28). The expressions of a template and the solver take the same thread
// with their own tasks of the epic of code (R16).
//
// The sandbox shares nothing between evaluations and leaves starlark.Universe as it is: the
// dialect's built-in names are its 33 names at the version the format names, and an expression
// names no other even in a process that adds to it.
package sandbox
