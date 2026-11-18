// Package sandbox runs the Starlark of the format — the predicate of a check, the expressions of a
// template and its solver — in go.starlark.net at the version the format names, in its one
// dialect, with the format's limits of steps and module-kit's clocks, on the caller's goroutine and
// with its context (R88 spec; R11).
//
// CheckExpr and CheckProgram check an expression and a program of the dialect. Eval evaluates an
// expression on a thread of its own under Limits, stopping it with ErrSteps, ErrClock or ErrPanic,
// or with the error of a context that is done (R28). Compile compiles an expression once into a
// Func, which Call evaluates on each binding of its names under the same Limits and stops: the
// expressions of a template, on each combination of its parameters (R32). Exec runs the top level
// of a program into a Program whose Call calls its functions on the same thread, within the same
// steps and clock and with the same stops: the solver of a template, whose solve is called on each
// of its instances (R33).
//
// The sandbox shares nothing between evaluations or programs and leaves starlark.Universe as it is:
// an expression or a program names no built-in name but the dialect's 33, even in a process that
// adds to it.
package sandbox
