package template

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"slices"
	"strconv"
	"strings"

	"go.starlark.net/starlark"

	"github.com/altnovatrail/spec/types/formatv0"

	"github.com/altnovatrail/module-kit/internal/sandbox"
	"github.com/altnovatrail/module-kit/internal/schema"
)

// solverLimits are the limits that the solver of a template runs with: sandbox.Solver, the format's
// budget and module-kit's clock (R11), which a test of the clock lowers.
var solverLimits = sandbox.Solver

// Inputs gives the names of the numbers that the solver of a template gets on each instance, in the
// order of their names, compared by code units: each parameter and derived number whose placeholder
// stands in an open field of the template — the statement, the givens, the unknown, the text of an
// item, a text piece of its line, an option of an answer slot bound choice (format/README.md,
// sections 6 and 24). A number that only a closed field shows — a key, a trap value, the solution —
// is not among them, so the solver solves the task that the learner sees and cannot read an answer
// that an expression already gives. The open view is that of the template as it is written, as
// spec's second implementation builds it.
//
// A task that is no template gives none, and a string that breaks task.placeholders gives none of
// its own.
func Inputs(task formatv0.Task) []formatv0.TemplateName {
	if task.Template == nil {
		return nil
	}
	t := *task.Template
	var names []formatv0.TemplateName
	for _, s := range stringsIn(schema.Open("task.schema.json", schema.ValueOf(task))) {
		pieces, _ := piecesOf(s.text)
		for _, name := range placeholdersOf(pieces) {
			_, parameter := t.Parameters[name]
			_, derived := t.Derived[name]
			if (parameter || derived) && !slices.Contains(names, name) {
				names = append(names, name)
			}
		}
	}
	slices.Sort(names)
	return names
}

// programBreaches gives the breach of task.template-solver of a solver whose program is no program
// of the dialect: one that does not parse, that holds load, or that names a name that it does not
// bind and that is no built-in name of the dialect (format/README.md, section 24). It fails before
// the first instance, whatever the instances are, so it stands with the breaches of names,
// expressions and placeholders.
func programBreaches(t formatv0.Template) []Breach {
	if t.Solver == nil {
		return nil
	}
	if err := sandbox.CheckProgram(t.Solver.Program); err != nil {
		return []Breach{beforeTheFirst("its program is no program of the dialect of the format: " + err.Error())}
	}
	return nil
}

// beforeTheFirst gives the breach of task.template-solver of a solver that fails before the first
// instance, for the reason given.
func beforeTheFirst(reason string) Breach {
	return Breach{RuleSolver, at("template", "solver", "program"), "the solver fails before the first instance: " + reason}
}

// solve runs the solver of a template on its instances, as section 24 of the format text runs it:
// the program once, on one thread with the budget of 100 000 000 steps and module-kit's clock of a
// minute, and then solve on every instance in their order, on the same thread, each with a frozen
// dict of the numbers that Inputs names, inserted in their order. The top level and every call count
// their steps together, and one clock runs for them all: the budget is the template's, not each
// instance's (R11).
//
// It gives the numbers of every instance with the results that solve returns among them; or the
// breach of task.template-solver where the solver first fails — before the first instance, or on an
// instance, named by its numbers —, since the steps and the clock that remain after a failure tell
// nothing of the instances after it; or the error of ctx when ctx is done before the solver ends,
// which is no breach (R11).
func solve(ctx context.Context, task formatv0.Task, instances []formatv0.Numbers) ([]formatv0.Numbers, *Breach, error) {
	solver := task.Template.Solver
	program, err := sandbox.Exec(ctx, solverLimits, solver.Program)
	if err != nil {
		if sandbox.ContextDone(err) {
			return nil, nil, err
		}
		breach := beforeTheFirst(stopped(err, "its program fails as it runs: "))
		return nil, &breach, nil
	}
	defer program.Close()
	fn, defined := program.Global("solve")
	if reason := notSolve(fn, defined); reason != "" {
		breach := beforeTheFirst(reason)
		return nil, &breach, nil
	}
	inputs := Inputs(task)
	solved := make([]formatv0.Numbers, len(instances))
	for k, numbers := range instances {
		given := starlark.NewDict(len(inputs))
		for _, name := range inputs {
			// A string key into a dict that grows: no error.
			_ = given.SetKey(starlark.String(name), starlark.MakeInt64(numbers[name]))
		}
		given.Freeze()
		failed := func(reason string) *Breach {
			return &Breach{RuleSolver, at("template", "solver", "program"), fmt.Sprintf("on %s, %s", Shown(*task.Template, numbers), reason)}
		}
		returned, err := program.Call(fn, given)
		if err != nil {
			if sandbox.ContextDone(err) {
				return nil, nil, err
			}
			return nil, failed(stopped(err, "solve fails: ")), nil
		}
		results, reason := resultsOf(returned, solver.Results)
		if reason != "" {
			return nil, failed(reason), nil
		}
		solved[k] = maps.Clone(numbers)
		maps.Copy(solved[k], results)
	}
	return solved, nil, nil
}

// stopped words why the solver stopped: the count of steps at the budget, module-kit's clock, a
// panic of the interpreter, or an error of the program, after failing, with the place of the
// innermost call of the program where it fails, as go.starlark.net's backtrace ends.
func stopped(err error, failing string) string {
	var evalErr *starlark.EvalError
	switch {
	case errors.Is(err, sandbox.ErrSteps):
		return fmt.Sprintf("the count of steps reaches %d, the budget of the template", solverLimits.Steps)
	case errors.Is(err, sandbox.ErrClock):
		return fmt.Sprintf("module-kit's clock stops it after %s s", strconv.FormatFloat(solverLimits.Clock.Seconds(), 'f', -1, 64))
	case errors.As(err, &evalErr):
		for i := range len(evalErr.CallStack) {
			// A built-in function, such as fail, stands at no place of the program.
			if frame := evalErr.CallStack.At(i); frame.Pos.Line > 0 {
				return fmt.Sprintf("%s%s, at %d:%d in %s", failing, evalErr.Msg, frame.Pos.Line, frame.Pos.Col, frame.Name)
			}
		}
		return failing + evalErr.Msg
	}
	// A panic of the interpreter, as the sandbox words it.
	return strings.TrimPrefix(err.Error(), "sandbox: ")
}

// notSolve says why the global solve of a program is no function of one parameter, as the solver
// calls it, or "" when it is one (format/README.md, section 24). A function of one parameter is one
// that the program defines, with def or lambda, with one parameter that the dict is passed to by its
// place, which may have a default: no other parameter, no keyword-only one, and neither *args nor
// **kwargs.
func notSolve(fn starlark.Value, defined bool) string {
	if !defined {
		return "its program defines no solve"
	}
	f, ok := fn.(*starlark.Function)
	if !ok {
		return fmt.Sprintf("solve is a value of the type %s, not a function of one parameter", fn.Type())
	}
	if f.NumParams() == 1 && f.NumKwonlyParams() == 0 && !f.HasVarargs() && !f.HasKwargs() {
		return ""
	}
	params := parametersOf(f)
	if len(params) == 0 {
		return "solve has no parameter, and it has to be a function of one parameter"
	}
	return fmt.Sprintf("solve has the parameters %s, and it has to be a function of one parameter", strings.Join(params, ", "))
}

// parametersOf writes the parameters of a function as its def does: those passed by their place, then
// *args, or a lone * before the keyword-only ones, then those, and then **kwargs. go.starlark.net
// numbers them in another order: those by place, the keyword-only ones, *args and **kwargs.
func parametersOf(f *starlark.Function) []string {
	name := func(i int) string {
		n, _ := f.Param(i)
		return n
	}
	byPlace := f.NumParams() - f.NumKwonlyParams()
	if f.HasVarargs() {
		byPlace--
	}
	if f.HasKwargs() {
		byPlace--
	}
	var params []string
	for i := range byPlace {
		params = append(params, name(i))
	}
	switch {
	case f.HasVarargs():
		params = append(params, "*"+name(byPlace+f.NumKwonlyParams()))
	case f.NumKwonlyParams() > 0:
		params = append(params, "*")
	}
	for i := range f.NumKwonlyParams() {
		params = append(params, name(byPlace+i))
	}
	if f.HasKwargs() {
		params = append(params, "**"+name(f.NumParams()-1))
	}
	return params
}

// resultsOf reads what solve returns on an instance: a dict whose keys are the names of the results,
// exactly these, and whose values are ints between −(2^53−1) and 2^53−1 (format/README.md, section
// 24). It gives the results by their names, or why the value is none such: the first key that names
// no result, in the order of the dict, and else the first result, in the order of results, that it
// lacks or whose value is no such int.
func resultsOf(v starlark.Value, names []formatv0.TemplateName) (formatv0.Numbers, string) {
	dict, ok := v.(*starlark.Dict)
	if !ok {
		return nil, fmt.Sprintf("solve returns a value of the type %s, not a dict of its results", v.Type())
	}
	for _, key := range dict.Keys() {
		if name, ok := key.(starlark.String); !ok || !slices.Contains(names, formatv0.TemplateName(name)) {
			return nil, fmt.Sprintf("solve returns the key %s, which names no result of the solver", key)
		}
	}
	results := formatv0.Numbers{}
	for _, name := range names {
		value, found, _ := dict.Get(starlark.String(name))
		if !found {
			return nil, fmt.Sprintf("solve returns no %s", name)
		}
		i, ok := value.(starlark.Int)
		if !ok {
			return nil, fmt.Sprintf("the result %s is a value of the type %s, not an int", name, value.Type())
		}
		n, ok := i.Int64()
		if !ok || n > limit || n < -limit {
			return nil, fmt.Sprintf("the result %s is %s, past ±(2^53 − 1)", name, i)
		}
		results[name] = n
	}
	return results, ""
}
