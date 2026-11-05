package template

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"math/big"
	"slices"
	"strconv"
	"strings"

	"go.starlark.net/starlark"

	"github.com/altnovatrail/spec/types/formatv0"

	"github.com/altnovatrail/module-kit/internal/sandbox"
	"github.com/altnovatrail/module-kit/internal/schema"
)

// The rules of templates that a template breaks by itself — by its names, its expressions, its
// placeholders, the combinations of its parameters and its solver — whatever its instances hold
// (format/README.md, sections 23 and 24). task.placeholders holds outside a template too, and the
// check of a module reads it there.
const (
	RuleNames        = "task.template-names"
	RuleExpressions  = "task.template-expressions"
	RulePlaceholders = "task.placeholders"
	RuleInstances    = "task.template-instances"
	RuleSolver       = "task.template-solver"
)

// A Breach is a reason that a template breaks a rule of templates: its Rule, one of RuleNames,
// RuleExpressions, RulePlaceholders, RuleInstances and RuleSolver; the value of the task that breaks
// it, as a JSON Pointer in URI fragment form, such as #/template/derived/b; and what breaks it.
type Breach struct {
	Rule, Pointer, Message string
}

// limit is how far a derived number may lie from 0, as an integer of YAML does: 2^53 − 1
// (format/README.md, sections 4 and 23).
const limit = 1<<53 - 1

// Instances counts the instances of a template and gives the numbers of each, in their order: every
// parameter, every derived number and every result of the solver of the instance, by its name
// (format/README.md, sections 23 and 24). The combinations of the parameters come in the order of an
// odometer — the parameters in the order of their names, the first changing slowest, each through
// its domain upwards —, and on each the derived numbers are computed and then the constraint, in the
// sandbox of Starlark; those for which the constraint holds, or all of them when there is none, are
// the instances. A solver then runs on them in their order, on one thread of its own, and its results
// join the numbers of each.
//
// A template that breaks a rule of templates gives its breaches and no numbers, in the order of
// spec's second implementation: every breach of its names, its expressions, its placeholders and the
// program of its solver; and only when none of them breaks, those of its combinations — every empty
// domain, or too many combinations, or the first combination on which an expression fails or a
// derived number lies past ±(2^53 − 1), where the count stops, or none that gives an instance; and
// only when it has instances, that of its solver on the first instance where it fails, or before the
// first. A task that is no template gives neither.
//
// The count looks at ctx before each combination, and the solver runs with it, on the caller's
// goroutine; how many templates the caller counts at once is its own to say (R11). The error is that
// of ctx when ctx is done before the count ends, and is no breach: there are then neither numbers nor
// breaches. Instances may be called from many goroutines at once, each on a template of its own or the
// same.
func Instances(ctx context.Context, task formatv0.Task) ([]formatv0.Numbers, []Breach, error) {
	if task.Template == nil {
		return nil, nil, nil
	}
	t := *task.Template
	parameters := slices.Sorted(maps.Keys(t.Parameters))
	derived := slices.Sorted(maps.Keys(t.Derived))
	breaches := slices.Concat(namesBreaches(t), expressionBreaches(t, parameters, derived), placeholderBreaches(task), programBreaches(t))
	if len(breaches) > 0 {
		return nil, breaches, nil
	}
	instances, breaches, err := count(ctx, t, parameters, derived)
	if err == nil && len(breaches) == 0 && t.Solver != nil {
		var breach *Breach
		instances, breach, err = solve(ctx, task, instances)
		if breach != nil {
			breaches = []Breach{*breach}
		}
	}
	if err != nil {
		return nil, nil, fmt.Errorf("template: the task %s: %w", task.ID, err)
	}
	return instances, breaches, nil
}

// results gives the names of the results of a template's solver: none when it has none.
func results(t formatv0.Template) []formatv0.TemplateName {
	if t.Solver == nil {
		return nil
	}
	return t.Solver.Results
}

// namesBreaches gives the breaches of task.template-names: no two of the parameters, the derived
// numbers and the results of the solver share a name. Two parameters and two derived numbers cannot,
// being the names of one object, and two results the schema keeps apart.
func namesBreaches(t formatv0.Template) []Breach {
	var found []Breach
	for _, name := range slices.Sorted(maps.Keys(t.Derived)) {
		if _, ok := t.Parameters[name]; ok {
			found = append(found, Breach{RuleNames, at("template", "derived", string(name)),
				fmt.Sprintf("%s is both a parameter and a derived number", name)})
		}
	}
	for k, name := range results(t) {
		var other string
		if _, ok := t.Parameters[name]; ok {
			other = "a parameter"
		} else if _, ok := t.Derived[name]; ok {
			other = "a derived number"
		}
		if other != "" {
			found = append(found, Breach{RuleNames, at("template", "solver", "results", strconv.Itoa(k)),
				fmt.Sprintf("%s is both %s and a result of the solver", name, other)})
		}
	}
	return found
}

// expressionBreaches gives the breaches of task.template-expressions: every derived number is an int
// expression of the subset that names only parameters, and the constraint a bool expression of it
// that names only parameters and derived numbers. The subset is of Starlark and runs in the dialect
// of the format, so an expression is also one that go.starlark.net at its version parses, as a
// predicate is (format/README.md, sections 15, 23 and 24).
func expressionBreaches(t formatv0.Template, parameters, derived []formatv0.TemplateName) []Breach {
	var found []Breach
	check := func(src formatv0.TemplateExpression, names []formatv0.TemplateName, want exprType, what string, place ...string) {
		got, err := typeOf(string(src), plain(names))
		switch {
		case err != nil:
			found = append(found, Breach{RuleExpressions, at(place...), fmt.Sprintf("%s is no expression of the subset of templates: %v", what, err)})
		case got != want:
			found = append(found, Breach{RuleExpressions, at(place...), fmt.Sprintf("%s is %s expression, and it has to be %s one", what, got, want)})
		default:
			if err := sandbox.CheckExpr(string(src), plain(names)...); err != nil {
				found = append(found, Breach{RuleExpressions, at(place...), fmt.Sprintf("%s is no expression of the dialect of the format, as go.starlark.net parses it: %v", what, err)})
			}
		}
	}
	for _, name := range derived {
		check(t.Derived[name], parameters, intType, "the derived number "+string(name), "template", "derived", string(name))
	}
	if t.Constraint != nil {
		check(*t.Constraint, slices.Concat(parameters, derived), boolType, "the constraint", "template", "constraint")
	}
	return found
}

// plain gives names as strings.
func plain(names []formatv0.TemplateName) []string {
	words := make([]string, len(names))
	for i, name := range names {
		words[i] = string(name)
	}
	return words
}

// placeholderBreaches gives the breaches of task.placeholders in a template: in every string of the
// task outside its template, every brace is doubled or part of a placeholder that names a parameter,
// a derived number or a result of the solver, and a placeholder of a result stands only in a key or
// a trap value of a check of a number or a fraction, in the solution or in a leak phrase
// (format/README.md, sections 23 and 24). A breach for each string that breaks it, at the string.
func placeholderBreaches(task formatv0.Task) []Breach {
	t := *task.Template
	known := func(name formatv0.TemplateName) bool {
		_, parameter := t.Parameters[name]
		_, derived := t.Derived[name]
		return parameter || derived || slices.Contains(results(t), name)
	}
	onlyResult := func(name formatv0.TemplateName) bool {
		_, parameter := t.Parameters[name]
		_, derived := t.Derived[name]
		return !parameter && !derived && slices.Contains(results(t), name)
	}
	var found []Breach
	for _, s := range stringsOf(task) {
		pieces, err := piecesOf(s.text)
		names := placeholdersOf(pieces)
		var message string
		if err != nil {
			message = err.Error()
		} else if i := slices.IndexFunc(names, func(n formatv0.TemplateName) bool { return !known(n) }); i >= 0 {
			message = fmt.Sprintf("{%s} names no parameter, derived number or result of the solver", names[i])
		} else if i := slices.IndexFunc(names, onlyResult); i >= 0 && !takesResults(task, s.tokens) {
			message = fmt.Sprintf("{%s} is a result of the solver, which stands only in a key or a trap value of a number or a fraction, in the solution and in a leak phrase", names[i])
		}
		if message != "" {
			found = append(found, Breach{RulePlaceholders, at(s.tokens...), message})
		}
	}
	return found
}

// takesResults says whether a placeholder of a result of the solver may stand in the string of a task
// at the place tokens: in the key or the value of a trap value of a check of a number or a fraction,
// in the solution and in a leak phrase (format/README.md, section 24).
func takesResults(task formatv0.Task, tokens []string) bool {
	switch tokens[0] {
	case "solution", "leak_phrases":
		return true
	case "items":
	default:
		return false
	}
	// items/<i>/answer_slots/<j>/check/key, or .../check/trap_values/<k>/value, and so for a line.
	if len(tokens) < 6 || tokens[4] != "check" {
		return false
	}
	i, _ := strconv.Atoi(tokens[1])
	j, _ := strconv.Atoi(tokens[3])
	var check *formatv0.Check
	switch item := task.Items[i]; tokens[2] {
	case "answer_slots":
		check = &item.AnswerSlots[j].Check
	case "line":
		check = item.Line[j].Check
	}
	if check == nil || check.Kind != formatv0.CheckKindNumber && check.Kind != formatv0.CheckKindFraction {
		return false
	}
	return len(tokens) == 6 && tokens[5] == "key" || len(tokens) == 8 && tokens[5] == "trap_values" && tokens[7] == "value"
}

// at writes a place in a task as a JSON Pointer in URI fragment form.
func at(tokens ...string) string {
	return schema.Pointer(tokens)
}

// count counts the instances of a template whose names, expressions, placeholders and program break
// no rule, as Instances does, with ctx, whose error is the error when it is done.
func count(ctx context.Context, t formatv0.Template, parameters, derived []formatv0.TemplateName) ([]formatv0.Numbers, []Breach, error) {
	var empty []Breach
	combinations := big.NewInt(1)
	for _, name := range parameters {
		d := t.Parameters[name]
		if d.To < d.From {
			empty = append(empty, Breach{RuleInstances, at("template", "parameters", string(name)),
				fmt.Sprintf("the domain of %s is empty: its to, %d, is below its from, %d", name, d.To, d.From)})
			continue
		}
		size := new(big.Int).Sub(big.NewInt(d.To), big.NewInt(d.From))
		combinations.Mul(combinations, size.Add(size, big.NewInt(1)))
	}
	switch {
	case len(empty) > 0:
		return nil, empty, nil
	case combinations.Cmp(big.NewInt(MaxInstances)) > 0:
		return nil, []Breach{{RuleInstances, at("template", "parameters"),
			fmt.Sprintf("its domains multiply to %s combinations, more than %d", combinations, MaxInstances)}}, nil
	}
	expressions := compiled(t, parameters, derived)
	var instances []formatv0.Numbers
	values := make([]int64, len(parameters))
	for i, name := range parameters {
		values[i] = t.Parameters[name].From
	}
	for {
		// ctx is looked at once a combination: an expression of the subset is bounded by its length.
		if err := ctx.Err(); err != nil {
			return nil, nil, err
		}
		numbers, holds, breach := expressions.on(parameters, values)
		if breach != nil {
			return nil, []Breach{*breach}, nil
		}
		if holds {
			instances = append(instances, numbers)
		}
		// The odometer: the last parameter changes fastest.
		i := len(values) - 1
		for ; i >= 0 && values[i] == t.Parameters[parameters[i]].To; i-- {
			values[i] = t.Parameters[parameters[i]].From
		}
		if i < 0 {
			break
		}
		values[i]++
	}
	if len(instances) == 0 {
		return nil, []Breach{{RuleInstances, at("template", "constraint"), "no combination of its parameters meets the constraint"}}, nil
	}
	return instances, nil, nil
}

// The expressions of a template, each compiled once in the sandbox of Starlark: each derived number
// on the parameters, in the order of their names, and the constraint, when there is one, on the
// parameters and then the derived numbers.
type expressions struct {
	derived    []formatv0.TemplateName
	of         []*sandbox.Func
	constraint *sandbox.Func
}

// compiled compiles the expressions of a template that break no rule of templates. The sandbox takes
// every expression of the subset that go.starlark.net parses, so one that it does not is a fault of
// module-kit.
func compiled(t formatv0.Template, parameters, derived []formatv0.TemplateName) expressions {
	compile := func(src formatv0.TemplateExpression, names []formatv0.TemplateName) *sandbox.Func {
		f, err := sandbox.Compile(string(src), plain(names)...)
		if err != nil {
			panic(fmt.Sprintf("template: an expression of the subset that the sandbox does not take: %q: %v", src, err))
		}
		return f
	}
	e := expressions{derived: derived}
	for _, name := range derived {
		e.of = append(e.of, compile(t.Derived[name], parameters))
	}
	if t.Constraint != nil {
		e.constraint = compile(*t.Constraint, slices.Concat(parameters, derived))
	}
	return e
}

// on computes the derived numbers of a combination of the parameters, in the order of their names,
// and then the constraint, and gives the numbers, and whether the combination is an instance; or the
// breach of task.template-instances when an expression fails on it or a derived number lies past
// ±(2^53 − 1). An expression of the subset calls no function and holds no loop, so its steps are
// bounded by its length, and it runs with no limit of steps, no clock and no context: count looks at
// the caller's before each combination (format/README.md, section 23).
func (e expressions) on(parameters []formatv0.TemplateName, values []int64) (formatv0.Numbers, bool, *Breach) {
	numbers := formatv0.Numbers{}
	args := make(starlark.Tuple, len(values), len(values)+len(e.derived))
	for i, name := range parameters {
		numbers[name] = values[i]
		args[i] = starlark.MakeInt64(values[i])
	}
	shown := shownOf(parameters, values)
	for k, name := range e.derived {
		place := at("template", "derived", string(name))
		v, err := e.of[k].Call(context.Background(), sandbox.Limits{}, args[:len(parameters)]...)
		if err != nil {
			return nil, false, &Breach{RuleInstances, place, fmt.Sprintf("on %s, the derived number %s fails: %v", shown, name, failure(err))}
		}
		n, ok := v.(starlark.Int).Int64()
		if !ok || n > limit || n < -limit {
			return nil, false, &Breach{RuleInstances, place, fmt.Sprintf("on %s, the derived number %s is %s, past ±(2^53 − 1)", shown, name, v)}
		}
		numbers[name] = n
		args = append(args, v)
	}
	if e.constraint == nil {
		return numbers, true, nil
	}
	v, err := e.constraint.Call(context.Background(), sandbox.Limits{}, args...)
	if err != nil {
		return nil, false, &Breach{RuleInstances, at("template", "constraint"), fmt.Sprintf("on %s, the constraint fails: %v", shown, failure(err))}
	}
	return numbers, bool(v.(starlark.Bool)), nil
}

// failure gives the words of an error of an expression as Starlark says them, without the frames of
// the call that the sandbox makes of it.
func failure(err error) string {
	var evalErr *starlark.EvalError
	if errors.As(err, &evalErr) {
		return evalErr.Msg
	}
	return err.Error()
}

// shownOf writes the numbers of the parameters of a combination, in the order of their names, as a
// breach names an instance: a = 1, b = 4.
func shownOf(parameters []formatv0.TemplateName, values []int64) string {
	said := make([]string, len(parameters))
	for i, name := range parameters {
		said[i] = fmt.Sprintf("%s = %d", name, values[i])
	}
	return strings.Join(said, ", ")
}

// Shown writes the numbers of an instance of a template as a breach names the instance: its
// parameters, in the order of their names, each with its number, as in a = 1, b = 4. A name that
// numbers does not give is left out.
func Shown(t formatv0.Template, numbers formatv0.Numbers) string {
	var parameters []formatv0.TemplateName
	var values []int64
	for _, name := range slices.Sorted(maps.Keys(t.Parameters)) {
		if n, ok := numbers[name]; ok {
			parameters, values = append(parameters, name), append(values, n)
		}
	}
	return shownOf(parameters, values)
}

// Instance gives the instance of a template with the numbers of one of its instances: the task with
// every placeholder replaced by its number, in decimal digits with - before a negative one, and every
// doubled brace by one, in a single pass, without its template, and with the template's id
// (format/README.md, section 23). It evaluates no expression and runs no Starlark: api puts in the
// numbers that the package carries, picked by AttemptInstance, and the open view of the instance is
// openview.Task of it (section 25).
//
// The error is for a task that is no template, a string of it that breaks task.placeholders, and a
// placeholder whose number numbers does not give. The instance is not checked against the schema of
// a task: the check of a module holds every instance of a template to it, by task.instances.
func Instance(task formatv0.Task, numbers formatv0.Numbers) (formatv0.Task, error) {
	if task.Template == nil {
		return formatv0.Task{}, fmt.Errorf("template: the task %s is no template", task.ID)
	}
	value, err := fill(task, numbers)
	if err != nil {
		return formatv0.Task{}, fmt.Errorf("template: the task %s: %w", task.ID, err)
	}
	instance, err := schema.Decode[formatv0.Task](value)
	if err != nil {
		// The value is that of a task with only its strings changed.
		panic(fmt.Sprintf("template: an instance of a task: %v", err))
	}
	return instance, nil
}
