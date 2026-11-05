package template

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"go.starlark.net/starlark"

	"github.com/altnovatrail/spec/types/formatv0"

	"github.com/altnovatrail/module-kit/internal/sandbox"
	"github.com/altnovatrail/module-kit/internal/schema"
	"github.com/altnovatrail/module-kit/internal/yaml"
)

// validTask reads a task, written in YAML, that the schema of a task takes, into its type.
func validTask(t *testing.T, text string) formatv0.Task {
	t.Helper()
	value, err := yaml.Read([]byte(text))
	if err != nil {
		t.Fatal(err)
	}
	if errs := schema.Validate("task.schema.json", value); len(errs) > 0 {
		t.Fatalf("the task is not valid against its schema: %v", errs)
	}
	return taskOf(t, text)
}

// sheet is a template of a handwriting sheet whose parameters each stand in one field of the task, by
// its name: a in the statement, m in the alt of its picture, b in the givens, c in the unknown, d in
// the text of an item, f in an option of an answer slot bound choice, and the others in closed fields
// alone — g in a key, h in a trap value, i in a trap's text, j in the solution, k in a leak phrase,
// l in a simpler case —, e only in doubled braces and z nowhere. The derived number shown stands in
// the statement, and hidden in the solution alone.
const sheet = `id: sheet
template:
  parameters:
    a: {from: 1, to: 1}
    b: {from: 2, to: 2}
    c: {from: 3, to: 3}
    d: {from: 4, to: 4}
    e: {from: 5, to: 5}
    f: {from: 6, to: 6}
    g: {from: 7, to: 7}
    h: {from: 8, to: 8}
    i: {from: 9, to: 9}
    j: {from: 10, to: 10}
    k: {from: 11, to: 11}
    l: {from: 12, to: 12}
    m: {from: 13, to: 13}
    z: {from: 14, to: 14}
  derived:
    shown: a + 1
    hidden: a + 2
  solver:
    program: |
      def solve(numbers):
          return {"found": len(numbers)}
    results: [found]
statement:
  - text: "Take {a}, {shown} and {{e}}."
  - picture: pictures/p.png
    alt: "A picture of {m}."
    width: 240
    height: 240
givens: ["{b} is given."]
unknown: "What {c} gives."
canvas:
  kind: handwriting_sheet
items:
  - text: "Item {d}."
    answer_slots:
      - binding: zone
        zone: {left: 0, top: 0, width: 10, height: 10}
        check:
          kind: number
          key: "{g}"
          trap_values:
            - value: "{h}"
              trap: t
              text: "Not {i}."
    simpler_case: "Try {l}."
  - answer_slots:
      - binding: choice
        options: ["{f}", "0"]
        check: {kind: choice, key: "0"}
solution:
  - text: "{j}, {hidden} and {found}."
leak_phrases: ["{k} times"]
skills:
  primary: [s]
band: b
difficulty: 1
`

// The solver gets the parameters and the derived numbers whose placeholders stand in the open fields
// of the template, in the order of their names: not those of closed fields alone, nor a result, nor a
// name in doubled braces.
func TestInputs(t *testing.T) {
	task := validTask(t, sheet)
	if got, want := Inputs(task), []formatv0.TemplateName{"a", "b", "c", "d", "f", "m", "shown"}; !slices.Equal(got, want) {
		t.Errorf("a sheet: %v, want %v", got, want)
	}
	// A text piece of a line is open, and a key, a trap value, the solution and a leak phrase are not.
	if got, want := Inputs(validTask(t, times)), []formatv0.TemplateName{"a", "product"}; !slices.Equal(got, want) {
		t.Errorf("a workbook: %v, want %v", got, want)
	}
	if got := Inputs(validTask(t, strings.Replace(times, `"{a} ×"`, `"×"`, 1))); !slices.Equal(got, []formatv0.TemplateName{"product"}) {
		t.Errorf("without a: %v", got)
	}
	if got := Inputs(taskOf(t, strings.Replace(times, times[strings.Index(times, "template:"):strings.Index(times, "statement:")], "", 1))); got != nil {
		t.Errorf("a task that is no template gets %v", got)
	}
	// A string that breaks task.placeholders gives no names of its own.
	if got := Inputs(taskOf(t, strings.Replace(times, `"{a} ×"`, `"{a} × }"`, 1))); !slices.Equal(got, []formatv0.TemplateName{"product"}) {
		t.Errorf("with a lone brace: %v", got)
	}
	numbers, breaches := instancesOf(t, task)
	if len(breaches) > 0 || len(numbers) != 1 || numbers[0]["found"] != 7 {
		t.Errorf("the sheet's solver finds %v, %v", numbers, breaches)
	}
}

// withSolver gives times with a solver of the program and the results given, whose instances are
// a = 4, b = 2, then a = 4, b = 3, a = 5, b = 2, and so on, and a, b and product their numbers.
func withSolver(t *testing.T, program string, results string) formatv0.Task {
	t.Helper()
	indented := "      " + strings.ReplaceAll(strings.TrimSuffix(program, "\n"), "\n", "\n      ") + "\n"
	return edited(t, "    product: a * b\n", "    product: a * b\n  solver:\n    program: |\n"+indented+"    results: "+results+"\n",
		`key: "{b}"`, `key: "{factor}"`)
}

// solve gets a frozen dict of the numbers that the open fields show, inserted in the order of their
// names, each an int, and is called on every instance in their order; its results join the numbers.
func TestSolve(t *testing.T) {
	task := withSolver(t, `def solve(numbers):
    if list(numbers.keys()) != ["a", "product"] or type(numbers["a"]) != "int":
        fail("solve gets %r" % numbers)
    return {"factor": numbers["product"] // numbers["a"]}
`, "[factor]")
	numbers, breaches := instancesOf(t, task)
	if len(breaches) > 0 || len(numbers) != 6 {
		t.Fatalf("%v, %v", numbers, breaches)
	}
	for _, n := range numbers {
		if n["factor"] != n["b"] || len(n) != 4 {
			t.Errorf("the numbers %v", n)
		}
	}
	instance, err := Instance(task, numbers[1])
	if err != nil || *instance.Items[0].Line[1].Check.Key != "3" {
		t.Errorf("the instance of %v: %v, %v", numbers[1], instance.Items[0].Line[1].Check.Key, err)
	}
	// A template with many results gets each, and ints at the bounds.
	numbers, breaches = instancesOf(t, withSolver(t, `def solve(numbers):
    return {"most": 9007199254740991, "factor": numbers["product"] // numbers["a"], "least": -9007199254740991}
`, "[factor, least, most]"))
	if len(breaches) > 0 || numbers[0]["most"] != limit || numbers[0]["least"] != -limit || numbers[0]["factor"] != 2 {
		t.Errorf("at the bounds: %v, %v", numbers, breaches)
	}
	// A function of one parameter may give it a default, and may be a lambda.
	for _, program := range []string{
		"def solve(numbers = None):\n    return {\"factor\": 1}\n",
		"solve = lambda numbers: {\"factor\": 1}\n",
	} {
		if _, breaches := instancesOf(t, withSolver(t, program, "[factor]")); len(breaches) > 0 {
			t.Errorf("%q: %v", program, breaches)
		}
	}
}

// The solver breaks task.template-solver before the first instance, or on the first instance where it
// fails, named by its numbers, and the count stops there: the program, which parses, holds no load and
// names only the built-in names, defines solve as a function of one parameter, and on every instance
// solve returns a dict of exactly its results, each an int within ±(2^53 − 1).
func TestSolverBreaches(t *testing.T) {
	before := "the solver fails before the first instance: "
	for _, c := range []struct {
		name, program, results, found string
	}{
		{"no program", "def solve(numbers)\n    return 1\n", "[factor]", before + "its program is no program of the dialect of the format: 2:1: got newline, want ':'"},
		{"a load", "load(\"f.star\", \"f\")\n", "[factor]", before + "its program is no program of the dialect of the format: 1:1: the dialect has no load"},
		{"a name of no built-in", "def solve(numbers):\n    return {\"factor\": gcd(1, 2)}\n", "[factor]",
			before + "its program is no program of the dialect of the format: 2:23: undefined: gcd"},
		{"a top level that fails", "x = 1 // 0\n", "[factor]", before + "its program fails as it runs: floored division by zero, at 1:7 in <toplevel>"},
		{"no solve", "def find(numbers):\n    return {}\n", "[factor]", before + "its program defines no solve"},
		{"a solve of no function", "solve = 3\n", "[factor]", before + "solve is a value of the type int, not a function of one parameter"},
		{"a built-in solve", "solve = dict\n", "[factor]", before + "solve is a value of the type builtin_function_or_method, not a function of one parameter"},
		{"a solve of two", "def solve(numbers, limit):\n    return {}\n", "[factor]", before + "solve has the parameters numbers, limit, and it has to be a function of one parameter"},
		{"a solve of none", "def solve():\n    return {}\n", "[factor]", before + "solve has no parameter, and it has to be a function of one parameter"},
		{"a solve of *args", "def solve(*numbers):\n    return {}\n", "[factor]", before + "solve has the parameters *numbers, and it has to be a function of one parameter"},
		{"a solve of **kwargs", "def solve(numbers, *more, **rest):\n    return {}\n", "[factor]",
			before + "solve has the parameters numbers, *more, **rest, and it has to be a function of one parameter"},
		{"a solve of a keyword alone", "def solve(numbers, *, limit = 1):\n    return {}\n", "[factor]",
			before + "solve has the parameters numbers, *, limit, and it has to be a function of one parameter"},
		{"a solve of a keyword-only parameter", "def solve(*, numbers):\n    return {}\n", "[factor]",
			before + "solve has the parameters *, numbers, and it has to be a function of one parameter"},
		{"a solve of *args before a keyword-only parameter", "def solve(a, *more, b, **rest):\n    return {}\n", "[factor]",
			before + "solve has the parameters a, *more, b, **rest, and it has to be a function of one parameter"},
		{"a failure on an instance", "def solve(numbers):\n    if numbers[\"a\"] == 5:\n        fail(\"no factor\")\n    return {\"factor\": 1}\n", "[factor]",
			"on a = 5, b = 2, solve fails: fail: no factor, at 3:13 in solve"},
		{"a closed number", "def solve(numbers):\n    return {\"factor\": numbers[\"b\"]}\n", "[factor]", `on a = 4, b = 2, solve fails: key "b" not in dict, at 2:30 in solve`},
		{"a frozen dict", "def solve(numbers):\n    numbers[\"factor\"] = 1\n    return numbers\n", "[factor]",
			"on a = 4, b = 2, solve fails: cannot insert into frozen hash table, at 2:12 in solve"},
		{"no dict", "def solve(numbers):\n    return None\n", "[factor]", "on a = 4, b = 2, solve returns a value of the type NoneType, not a dict of its results"},
		{"a key of no result", "def solve(numbers):\n    return {\"factor\": 1, \"a\": 1}\n", "[factor]", `on a = 4, b = 2, solve returns the key "a", which names no result of the solver`},
		{"a key of no string", "def solve(numbers):\n    return {1: 1, \"factor\": 1}\n", "[factor]", "on a = 4, b = 2, solve returns the key 1, which names no result of the solver"},
		{"a result missing", "def solve(numbers):\n    return {\"other\": 1}\n", "[factor, other]", "on a = 4, b = 2, solve returns no factor"},
		{"a bool", "def solve(numbers):\n    return {\"factor\": True}\n", "[factor]", "on a = 4, b = 2, the result factor is a value of the type bool, not an int"},
		{"a float", "def solve(numbers):\n    return {\"factor\": 2.0}\n", "[factor]", "on a = 4, b = 2, the result factor is a value of the type float, not an int"},
		{"past the bound", "def solve(numbers):\n    return {\"factor\": -9007199254740992}\n", "[factor]", "on a = 4, b = 2, the result factor is -9007199254740992, past ±(2^53 − 1)"},
		{"far past it", "def solve(numbers):\n    return {\"factor\": 1 << 70}\n", "[factor]", "on a = 4, b = 2, the result factor is 1180591620717411303424, past ±(2^53 − 1)"},
		{"recursion", "def solve(numbers):\n    return solve(numbers)\n", "[factor]", "on a = 4, b = 2, solve fails: function solve called recursively, at 2:12 in solve"},
	} {
		numbers, breaches := instancesOf(t, withSolver(t, c.program, c.results))
		var found []string
		for _, b := range breaches {
			if b.Rule != RuleSolver || b.Pointer != "#/template/solver/program" {
				t.Errorf("%s: %+v", c.name, b)
			}
			found = append(found, b.Message)
		}
		if !slices.Equal(found, []string{c.found}) || numbers != nil {
			t.Errorf("%s: %q, %v; want %q", c.name, found, numbers, c.found)
		}
	}
	// A program that is none of the dialect stands with the other breaches of the template's own, and
	// the combinations are not counted.
	_, breaches := instancesOf(t, edited(t, "    product: a * b\n", "    product: a * b\n  solver:\n    program: \"load('f.star', 'f')\"\n    results: [factor]\n",
		"b: {from: 2, to: 3}", "b: {from: 3, to: 2}", `"= {product}"`, `"= {c}"`))
	if len(breaches) != 2 || breaches[0].Rule != RulePlaceholders || breaches[1].Rule != RuleSolver {
		t.Errorf("with a breach of a placeholder: %v", breaches)
	}
}

// The budget of steps and the clock are the template's: the top level and every call of solve count
// together, and the run stops when the count reaches the budget or the clock stops it, before the
// first instance or on the instance where it is. They are lowered here, so that a test does not take
// 100 000 000 steps or a minute; the vectors hold the budget itself.
func TestSolverStops(t *testing.T) {
	defer func(l sandbox.Limits) { solverLimits = l }(solverLimits)
	solverLimits = sandbox.Limits{Steps: 1000}
	// 10 steps for the top level, and 320 for each call of 30 turns: 330, 650, 970, and the fourth
	// instance reaches 1000.
	turns := "def solve(numbers):\n    for turn in range(30):\n        pass\n    return {\"factor\": 1}\n"
	for _, c := range []struct{ name, program, found string }{
		{"the budget of the template", turns, "on a = 5, b = 3, the count of steps reaches 1000, the budget of the template"},
		{"the budget of an instance", "def solve(numbers):\n    while True:\n        pass\n", "on a = 4, b = 2, the count of steps reaches 1000, the budget of the template"},
		{"the budget of the top level", "while True:\n    pass\n", "the solver fails before the first instance: the count of steps reaches 1000, the budget of the template"},
	} {
		_, breaches := instancesOf(t, withSolver(t, c.program, "[factor]"))
		if len(breaches) != 1 || breaches[0].Message != c.found {
			t.Errorf("%s: %v, want %q", c.name, breaches, c.found)
		}
	}
	// Four instances take 1290 steps: 10 below a budget of 1300 they give their results, as
	// within-the-budget does, and at a budget of 1290 the count reaches it, as reaches-the-budget does.
	four := edited(t, "a: {from: 4, to: 6}", "a: {from: 4, to: 5}", "    product: a * b\n", "    product: a * b\n  solver:\n    program: |\n"+
		"      def solve(numbers):\n          for turn in range(30):\n              pass\n          return {\"factor\": 1}\n    results: [factor]\n")
	solverLimits.Steps = 1300
	if numbers, breaches := instancesOf(t, four); len(numbers) != 4 || breaches != nil {
		t.Errorf("four instances within the budget: %v, %v", numbers, breaches)
	}
	solverLimits.Steps = 1290
	if _, breaches := instancesOf(t, four); len(breaches) != 1 || breaches[0].Message != "on a = 5, b = 3, the count of steps reaches 1290, the budget of the template" {
		t.Errorf("four instances at the budget: %v", breaches)
	}
	solverLimits = sandbox.Limits{Clock: 20 * time.Millisecond}
	_, breaches := instancesOf(t, withSolver(t, "def solve(numbers):\n    while True:\n        pass\n", "[factor]"))
	if want := "on a = 4, b = 2, module-kit's clock stops it after 0.02 s"; len(breaches) != 1 || breaches[0].Message != want {
		t.Errorf("the clock: %v, want %q", breaches, want)
	}
	if got := stopped(fmt.Errorf("sandbox: %w: boom", sandbox.ErrPanic), "solve fails: "); got != "the interpreter panicked: boom" {
		t.Errorf("a panic: %q", got)
	}
	if got := stopped(fmt.Errorf("sandbox: %w", &starlark.EvalError{Msg: "no place"}), "solve fails: "); got != "solve fails: no place" {
		t.Errorf("an error at no place of the program: %q", got)
	}
}

// A context that is done stops the count, in the expressions or in the solver, with its error, and is
// no breach.
func TestInstancesContext(t *testing.T) {
	done, cancel := context.WithCancel(context.Background())
	cancel()
	// ctx done before stops the count at its first combination, whether the template has expressions
	// and a solver or neither; and a deadline stops the solver, in its top level or in solve.
	bare := taskOf(t, "id: times\ntemplate:\n  parameters:\n    a: {from: 1, to: 2}\nstatement:\n  - text: Write the number.\n"+
		"canvas:\n  kind: workbook\nitems:\n  - line:\n      - binding: box\n        check: {kind: number, key: \"{a}\"}\nsolution:\n  - text: \"{a}\"\n"+
		"skills:\n  primary: [s]\nband: b\ndifficulty: 1\n")
	for _, task := range []formatv0.Task{taskOf(t, times), bare, withSolver(t, "def solve(numbers):\n    return {\"factor\": 1}\n", "[factor]")} {
		if numbers, breaches, err := Instances(done, task); !errors.Is(err, context.Canceled) || numbers != nil || breaches != nil ||
			err.Error() != "template: the task times: context canceled" {
			t.Errorf("a context done before: %v, %v, %v", numbers, breaches, err)
		}
	}
	for _, program := range []string{"def solve(numbers):\n    while True:\n        pass\n", "while True:\n    pass\n"} {
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
		numbers, breaches, err := Instances(ctx, withSolver(t, program, "[factor]"))
		cancel()
		if !errors.Is(err, context.DeadlineExceeded) || numbers != nil || breaches != nil {
			t.Errorf("a deadline in the solver %q: %v, %v, %v", program, numbers, breaches, err)
		}
	}
}

// Instances is called from many goroutines at once, each solver on a thread of its own.
func TestConcurrentInstances(t *testing.T) {
	task := withSolver(t, "def solve(numbers):\n    return {\"factor\": numbers[\"product\"] // numbers[\"a\"]}\n", "[factor]")
	want, _ := instancesOf(t, task)
	var wg sync.WaitGroup
	for range 16 {
		wg.Go(func() {
			if got, breaches, err := Instances(context.Background(), task); !slices.EqualFunc(got, want, maps.Equal) || breaches != nil || err != nil {
				t.Errorf("%v, %v, %v", got, breaches, err)
			}
		})
	}
	wg.Wait()
}
