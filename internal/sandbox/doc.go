// Package sandbox runs the Starlark of the format — the predicate of a check, the expressions of a
// template and its solver — in go.starlark.net at the version the format names, in its one
// dialect, with the format's limits of steps and module-kit's clocks, on the caller's goroutine and
// with its context (format/README.md, sections 15, 23 and 24; R88 spec; R11; architecture 3.4).
//
// CheckExpr checks an expression of the dialect and the names it may take, and Eval evaluates one on
// a thread of its own under Limits, of which Predicate are a predicate's, stopping it with ErrSteps,
// ErrClock or ErrPanic, or with the error of a context that is done (R28). Compile compiles an
// expression once into a Func of the names it is written on, which Call evaluates on each binding of
// them under the same Limits and stops: the expressions of a template, on each combination of its
// parameters (R32). CheckProgram checks a program of the dialect, and Exec runs its top level on a
// thread of its own under Limits, of which Solver are a solver's, into a Program whose Call calls its
// functions on the same thread, within the same steps and clock and with the same stops: the solver
// of a template, whose solve is called on each of its instances (R33).
//
// The sandbox shares nothing between evaluations or programs and leaves starlark.Universe as it is:
// the dialect's built-in names are its 33 names at the version the format names, and an expression
// or a program names no other even in a process that adds to it.
package sandbox
