package sandbox

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"sync/atomic"
	"time"

	"go.starlark.net/resolve"
	"go.starlark.net/starlark"
	"go.starlark.net/syntax"
)

// options gives the options of the dialect of all the Starlark of the format, those that the
// solvers of mathtrail-standalone run with: Set, While, TopLevelControl and GlobalReassign on, and
// Recursion off, so that a function that calls itself fails when it does, through a lambda too
// (format/README.md, section 24). A value of its own each time, since the parser, the resolver and
// the compiler each take a pointer.
func options() *syntax.FileOptions {
	return &syntax.FileOptions{Set: true, While: true, TopLevelControl: true, GlobalReassign: true}
}

// builtins are the built-in names of the dialect, as the format lists them: the 33 names of
// starlark.Universe at the version of go.starlark.net that it names (format/README.md, section 24).
var builtins = []string{
	"None", "True", "False", "abs", "all", "any", "bool", "bytes", "chr", "dict", "dir", "enumerate", "fail", "float",
	"getattr", "hasattr", "hash", "int", "len", "list", "max", "min", "ord", "print", "range", "repr", "reversed", "set",
	"sorted", "str", "tuple", "type", "zip",
}

func builtin(name string) bool { return slices.Contains(builtins, name) }

// Limits are what one evaluation may take. Steps is the format's: the evaluation stops when the
// count of steps on its thread, as go.starlark.net at the version of the format counts them,
// reaches it. Clock is module-kit's: the wall time after which it stops the evaluation (R11). Zero
// is no limit.
type Limits struct {
	Steps uint64
	Clock time.Duration
}

// Predicate are the limits of a predicate: the format's 1 000 000 steps, and module-kit's clock of
// a second, which a simple predicate does not come near, since it takes 6 ms for the steps
// (format/README.md, section 15; R11).
var Predicate = Limits{Steps: 1_000_000, Clock: time.Second}

// Why an evaluation stopped without its result. A context that is done stops it too, and then the
// error is the context's, wrapped: that is the caller's stop, not the sandbox's (R11).
var (
	ErrSteps = errors.New("the count of steps reached the limit")
	ErrClock = errors.New("the clock stopped the evaluation")
	ErrPanic = errors.New("the interpreter panicked")
)

// CheckExpr says whether src is an expression of the dialect that names nothing but names, the
// names that it binds itself, such as n in [n * n for n in range(10)], and the built-in names of
// the dialect, as task.predicate-expression asks of a predicate on value (format/README.md,
// section 15). The error says why not, at the line and the column where go.starlark.net finds it.
func CheckExpr(src string, names ...string) error {
	expr, err := parse(src)
	if err != nil {
		return err
	}
	_, err = resolve.ExprOptions(options(), expr, func(name string) bool { return slices.Contains(names, name) }, builtin)
	return located(err)
}

// Eval evaluates the expression src, with env binding the names it is written on, on a thread of
// its own with the limits and ctx (format/README.md, sections 15 and 24; R11): a predicate, with
// value bound to the integer read. src is checked first as CheckExpr checks it, on the names of
// env, so that it names nothing else even in a process that changed starlark.Universe.
//
// The error is the context's, wrapped, when ctx is done before the evaluation or stops it;
// ErrSteps, ErrClock or ErrPanic, wrapped, when the evaluation stops so; and otherwise that of the
// expression, which is none of the dialect, names what it may not, or fails as it runs.
func Eval(ctx context.Context, limits Limits, src string, env starlark.StringDict) (starlark.Value, error) {
	return run(ctx, limits, func(thread *starlark.Thread) (starlark.Value, error) {
		expr, err := checked(src, env.Keys())
		if err != nil {
			return nil, err
		}
		return starlark.EvalExprOptions(options(), thread, expr, env)
	})
}

// A Func is an expression of the dialect compiled once on the names it is written on, so that it is
// evaluated on many bindings of them: an expression of a template, on each combination of the
// parameters (format/README.md, section 23). It holds nothing that changes, so one Func is called from
// many goroutines at once.
type Func struct {
	names  []string
	result string // the global that the program assigns the value of the expression to
	prog   *starlark.Program
}

// Compile compiles src, an expression of the dialect that names nothing but names, the names that it
// binds itself and the built-in names of the dialect, as CheckExpr checks it, into a Func of names,
// in their order. The error is CheckExpr's.
//
// The Func is a program of one statement, which assigns the parsed src to a global that names do not
// hold: built from its tree rather than its text, so that it compiles whatever CheckExpr takes — a
// comment, a line break — and is as deep to the parser as src, which a lambda around it would not be;
// and with the steps of src, as Eval counts them.
func Compile(src string, names ...string) (*Func, error) {
	expr, err := checked(src, names)
	if err != nil {
		return nil, err
	}
	result := "_result"
	for slices.Contains(names, result) {
		result += "_"
	}
	at := syntax.Start(expr)
	file := &syntax.File{Options: options(), Stmts: []syntax.Stmt{
		&syntax.AssignStmt{OpPos: at, Op: syntax.EQ, LHS: &syntax.Ident{NamePos: at, Name: result}, RHS: expr},
	}}
	prog, err := starlark.FileProgram(file, func(name string) bool { return slices.Contains(names, name) })
	if err != nil {
		// CheckExpr resolved src on the same names, and the global is none of them.
		panic(fmt.Sprintf("sandbox: an expression that CheckExpr takes does not compile: %v", err))
	}
	return &Func{names: names, result: result, prog: prog}, nil
}

// Call evaluates the expression with args bound to its names, in order, on a thread of its own with
// the limits and ctx, as Eval does, and its error is as Eval's. Args that are not one for each name
// are an error.
func (f *Func) Call(ctx context.Context, limits Limits, args ...starlark.Value) (starlark.Value, error) {
	if len(args) != len(f.names) {
		return nil, fmt.Errorf("sandbox: %d values for the %d names %v", len(args), len(f.names), f.names)
	}
	env := make(starlark.StringDict, len(args))
	for i, name := range f.names {
		env[name] = args[i]
	}
	return run(ctx, limits, func(thread *starlark.Thread) (starlark.Value, error) {
		globals, err := f.prog.Init(thread, env)
		return globals[f.result], err
	})
}

// cancelled begins the error of go.starlark.net's interpreter when it sees the thread cancelled,
// which then names the reason (starlark/interp.go, checkStepsSlow).
const cancelled = "Starlark computation cancelled: "

// run runs an evaluation on a thread of its own, on the caller's goroutine: with the limits and
// ctx, a print that writes nothing and no load, and a panic of the interpreter recovered (R11). The
// limit of steps, the clock and ctx each cancel the thread with their own cause, and the first to
// come is the one that the error names, as the thread keeps the first reason it is cancelled with.
// The interpreter sees a cancel between its instructions, so a built-in function that it is in
// runs to its end first. An evaluation that ends before the interpreter sees a cancel keeps its own
// result or error, whatever cause comes as it ends.
func run(ctx context.Context, limits Limits, evaluate func(*starlark.Thread) (starlark.Value, error)) (result starlark.Value, err error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("sandbox: %w", err)
	}
	thread := &starlark.Thread{Print: func(*starlark.Thread, string) {}}
	var cause atomic.Pointer[error]
	stop := func(err error) {
		if cause.CompareAndSwap(nil, &err) {
			thread.Cancel(err.Error())
		}
	}
	thread.SetMaxExecutionSteps(limits.Steps)
	thread.OnMaxSteps = func(*starlark.Thread) { stop(ErrSteps) }
	if limits.Clock > 0 {
		clock := time.AfterFunc(limits.Clock, func() { stop(ErrClock) })
		defer clock.Stop()
	}
	defer context.AfterFunc(ctx, func() { stop(ctx.Err()) })()
	defer func() {
		if r := recover(); r != nil {
			result, err = nil, fmt.Errorf("sandbox: %w: %v", ErrPanic, r)
		}
	}()
	result, err = evaluate(thread)
	if err != nil {
		if c := cause.Load(); c != nil && strings.Contains(err.Error(), cancelled+(*c).Error()) {
			err = *c
		}
		return nil, fmt.Errorf("sandbox: %w", err)
	}
	return result, nil
}

// checked gives the expression src, parsed afresh, once CheckExpr has checked it on names: the
// resolver writes into the expression it resolves, and go.starlark.net resolves the one it
// evaluates again.
func checked(src string, names []string) (syntax.Expr, error) {
	if err := CheckExpr(src, names...); err != nil {
		return nil, err
	}
	return parse(src)
}

// parse parses src as an expression of the dialect.
func parse(src string) (syntax.Expr, error) {
	expr, err := options().ParseExpr("", src, 0)
	return expr, located(err)
}

// located writes an error of the parser or of the resolver at its line and column, as
// go.starlark.net finds them, without the name of a file, which an expression has none of.
func located(err error) error {
	var syntaxErr syntax.Error
	var resolveErr resolve.ErrorList
	switch {
	case errors.As(err, &syntaxErr):
		return fmt.Errorf("%d:%d: %s", syntaxErr.Pos.Line, syntaxErr.Pos.Col, syntaxErr.Msg)
	case errors.As(err, &resolveErr):
		return fmt.Errorf("%d:%d: %s", resolveErr[0].Pos.Line, resolveErr[0].Pos.Col, resolveErr[0].Msg)
	}
	return err
}
