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

// Solver are the limits of the solver of a template, for the top level of its program and every
// call of solve together: the format's budget of 100 000 000 steps, and module-kit's clock of a
// minute, which a solver of simple steps does not come near, since a loop takes 0.10–0.16 s for the
// budget (format/README.md, section 24; R11, R33).
var Solver = Limits{Steps: 100_000_000, Clock: time.Minute}

// Why an evaluation stopped without its result. A context that is done stops it too, and then the
// error is the context's, wrapped: that is the caller's stop, not the sandbox's (R11).
var (
	ErrSteps = errors.New("the count of steps reached the limit")
	ErrClock = errors.New("the clock stopped the evaluation")
	ErrPanic = errors.New("the interpreter panicked")
)

// ContextDone says whether an error of the sandbox is that of a context that is done: the caller's
// stop, which gives no verdict and no breach (R11).
func ContextDone(err error) bool {
	return errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded)
}

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

// CheckProgram says whether src is a program of the dialect: one that its parser parses, that holds
// no load, and that names nothing but the names that it binds itself and the built-in names of the
// dialect, as task.template-solver asks of the program of a solver (format/README.md, section 24).
// The error says why not, at the line and the column where go.starlark.net finds it, or where the
// load stands.
func CheckProgram(src string) error {
	file, err := parseFile(src)
	if err != nil {
		return err
	}
	var load *syntax.LoadStmt
	syntax.Walk(file, func(n syntax.Node) bool {
		if l, ok := n.(*syntax.LoadStmt); ok && load == nil {
			load = l
		}
		return load == nil
	})
	if load != nil {
		return fmt.Errorf("%d:%d: the dialect has no load", load.Load.Line, load.Load.Col)
	}
	return located(resolve.File(file, none, builtin))
}

// none is the predicate of a program's predeclared names: it has none.
func none(string) bool { return false }

// A Program is a program of the dialect whose top level has run on a thread of its own, and whose
// functions Call calls on the same thread: the solver of a template, whose solve is called on each
// of its instances in turn (format/README.md, section 24). The limits are the program's, not each
// call's: the steps of the top level and of every call count together against one limit, and one
// clock runs from Exec to Close, as does the watch of ctx.
//
// Its globals are frozen once the top level has run, as Starlark freezes those of a module
// ("Immediately after execution of a Starlark module, all values in its top-level environment are
// frozen", the spec of Starlark of go.starlark.net, Freezing a value), so a call changes nothing that
// the next one reads. A Program is for one goroutine at a time.
type Program struct {
	thread   *thread
	globals  starlark.StringDict
	panicked error // the panic of a call, after which the thread is no longer called
	closed   bool  // whether Close has ended its clock and its watch of ctx
}

// Exec runs the top level of the program src on a thread of its own with the limits and ctx, on the
// caller's goroutine, and gives the Program, whose Close the caller calls once it has called its
// functions. src is checked first as CheckProgram checks it, so that it names nothing else even in a
// process that changed starlark.Universe.
//
// The error is the context's, wrapped, when ctx is done before the program runs or stops it;
// CheckProgram's; ErrSteps, ErrClock or ErrPanic, wrapped, when the top level stops so; and otherwise
// that of the top level, which fails as it runs. With an error there is no Program.
func Exec(ctx context.Context, limits Limits, src string) (*Program, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("sandbox: %w", err)
	}
	if err := CheckProgram(src); err != nil {
		return nil, err
	}
	// Parsed afresh: the resolver wrote into the one that CheckProgram checked, and FileProgram
	// resolves the one it compiles again.
	file, _ := parseFile(src)
	prog, err := starlark.FileProgram(file, none)
	if err != nil {
		// CheckProgram resolved src on fewer built-in names than starlark.Universe holds.
		panic(fmt.Sprintf("sandbox: a program that CheckProgram takes does not compile: %v", err))
	}
	p := &Program{thread: start(ctx, limits)}
	if _, err := p.thread.do(func(thread *starlark.Thread) (starlark.Value, error) {
		globals, err := prog.Init(thread, nil)
		globals.Freeze()
		p.globals = globals
		return starlark.None, err
	}); err != nil {
		p.thread.end()
		return nil, err
	}
	return p, nil
}

// Global gives the value of a global of the program, which its top level bound, and whether it bound
// it.
func (p *Program) Global(name string) (starlark.Value, bool) {
	v, ok := p.globals[name]
	return v, ok
}

// Call calls fn with args on the thread of the program, within the steps and the clock that the top
// level and the calls before have left, and its error is as Exec's: the context's, ErrSteps, ErrClock
// or ErrPanic, wrapped, or that of the call. After a panic, Call calls nothing and gives the same
// error again; after Close, which ended its clock and its watch of ctx, it calls nothing and gives
// ErrClosed.
func (p *Program) Call(fn starlark.Value, args ...starlark.Value) (starlark.Value, error) {
	switch {
	case p.closed:
		return nil, fmt.Errorf("sandbox: %w", ErrClosed)
	case p.panicked != nil:
		return nil, p.panicked
	}
	v, err := p.thread.do(func(thread *starlark.Thread) (starlark.Value, error) {
		return starlark.Call(thread, fn, args, nil)
	})
	if errors.Is(err, ErrPanic) {
		p.panicked = err
	}
	return v, err
}

// Close stops the clock of the program and the watch of its ctx, so that Call calls nothing after
// it. It is called once the program's functions have been called, and may be called again.
func (p *Program) Close() {
	p.closed = true
	p.thread.end()
}

// ErrClosed is what Call gives after Close: the program has no clock and no watch of ctx left.
var ErrClosed = errors.New("the program is closed")

// cancelled begins the error of go.starlark.net's interpreter when it sees the thread cancelled,
// which then names the reason (starlark/interp.go, checkStepsSlow).
const cancelled = "Starlark computation cancelled: "

// run runs an evaluation on a thread of its own, on the caller's goroutine, with the limits and ctx
// (R11).
func run(ctx context.Context, limits Limits, evaluate func(*starlark.Thread) (starlark.Value, error)) (starlark.Value, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("sandbox: %w", err)
	}
	t := start(ctx, limits)
	defer t.end()
	return t.do(evaluate)
}

// A thread is a thread of Starlark with limits and a context, for one evaluation, or for a program
// and every call of its functions: with a print that writes nothing and no load (R11). The limit of
// steps, the clock and ctx each cancel it with their own cause, and the first to come is the one
// that an error names, as the thread keeps the first reason it is cancelled with. The interpreter
// sees a cancel between its instructions and on the entry to each call, so a built-in function that
// it is in runs to its end first.
type thread struct {
	starlark *starlark.Thread
	cause    atomic.Pointer[error]
	clock    *time.Timer // nil with no clock
	unwatch  func() bool // ends the watch of ctx
}

// start gives a thread with the limits and ctx, whose clock runs from now.
func start(ctx context.Context, limits Limits) *thread {
	t := &thread{starlark: &starlark.Thread{Print: func(*starlark.Thread, string) {}}}
	t.starlark.SetMaxExecutionSteps(limits.Steps)
	t.starlark.OnMaxSteps = func(*starlark.Thread) { t.stop(ErrSteps) }
	if limits.Clock > 0 {
		t.clock = time.AfterFunc(limits.Clock, func() { t.stop(ErrClock) })
	}
	t.unwatch = context.AfterFunc(ctx, func() { t.stop(ctx.Err()) })
	return t
}

// stop cancels the thread with its cause, unless another cause came first.
func (t *thread) stop(cause error) {
	if t.cause.CompareAndSwap(nil, &cause) {
		t.starlark.Cancel(cause.Error())
	}
}

// end stops the clock and the watch of ctx.
func (t *thread) end() {
	if t.clock != nil {
		t.clock.Stop()
	}
	t.unwatch()
}

// do runs an evaluation on the thread, on the caller's goroutine, with a panic of the interpreter
// recovered. An evaluation that ends before the interpreter sees a cancel keeps its own result or
// error, whatever cause comes as it ends.
func (t *thread) do(evaluate func(*starlark.Thread) (starlark.Value, error)) (result starlark.Value, err error) {
	defer func() {
		if r := recover(); r != nil {
			result, err = nil, fmt.Errorf("sandbox: %w: %v", ErrPanic, r)
		}
	}()
	result, err = evaluate(t.starlark)
	if err != nil {
		if c := t.cause.Load(); c != nil && strings.Contains(err.Error(), cancelled+(*c).Error()) {
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

// parseFile parses src as a program of the dialect.
func parseFile(src string) (*syntax.File, error) {
	file, err := options().Parse("", src, 0)
	return file, located(err)
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
