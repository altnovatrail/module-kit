package template

import (
	"maps"
	"slices"
	"strings"
	"testing"

	"github.com/altnovatrail/spec/types/formatv0"

	"github.com/altnovatrail/module-kit/internal/schema"
	"github.com/altnovatrail/module-kit/internal/yaml"
)

// taskOf reads a task, written in YAML, into its type.
func taskOf(t *testing.T, text string) formatv0.Task {
	t.Helper()
	value, err := yaml.Read([]byte(text))
	if err != nil {
		t.Fatal(err)
	}
	task, err := schema.Decode[formatv0.Task](value)
	if err != nil {
		t.Fatal(err)
	}
	return task
}

// times is a template of a line of the workbook, which the cases edit: 3 × 2 = 6 instances, the
// last parameter changing fastest.
const times = `id: times
template:
  parameters:
    b: {from: 2, to: 3}
    a: {from: 4, to: 6}
  derived:
    product: a * b
statement:
  - text: Write the missing number.
canvas:
  kind: workbook
items:
  - line:
      - text: "{a} ×"
      - binding: box
        check:
          kind: number
          key: "{b}"
          trap_values:
            - value: "{product}"
              trap: copied-the-product
      - text: "= {product}"
solution:
  - text: "{a} × {b} = {product}."
leak_phrases: ["{b} times"]
skills:
  primary: [times-tables]
band: grade-3
difficulty: 2
`

// edited gives times with each old text replaced by the new one after it, once.
func edited(t *testing.T, pairs ...string) formatv0.Task {
	t.Helper()
	text := times
	for i := 0; i < len(pairs); i += 2 {
		if !strings.Contains(text, pairs[i]) {
			t.Fatalf("times holds no %q", pairs[i])
		}
		text = strings.Replace(text, pairs[i], pairs[i+1], 1)
	}
	return taskOf(t, text)
}

const solver = "  solver:\n    program: |\n      def solve(numbers):\n          return {\"factor\": 1}\n    results: [factor]\n"

// The instances come in the order of the odometer: the parameters by their names, the first
// changing slowest, each through its domain upwards; each with every parameter and derived number.
func TestInstances(t *testing.T) {
	numbers, breaches := Instances(taskOf(t, times))
	want := []formatv0.Numbers{
		{"a": 4, "b": 2, "product": 8}, {"a": 4, "b": 3, "product": 12},
		{"a": 5, "b": 2, "product": 10}, {"a": 5, "b": 3, "product": 15},
		{"a": 6, "b": 2, "product": 12}, {"a": 6, "b": 3, "product": 18},
	}
	if len(breaches) > 0 || !slices.EqualFunc(numbers, want, maps.Equal) {
		t.Errorf("%v, %v; want %v", numbers, breaches, want)
	}
	// The constraint leaves combinations out, and reads the derived numbers; and and or evaluate their
	// right operand only when the left one does not decide.
	numbers, breaches = Instances(edited(t, "    product: a * b\n", "    product: a * b\n  constraint: product != 12 and (b == 3 or a // (b - 3) < 0)\n"))
	if want := []formatv0.Numbers{{"a": 4, "b": 2, "product": 8}, {"a": 5, "b": 2, "product": 10}, {"a": 5, "b": 3, "product": 15}, {"a": 6, "b": 3, "product": 18}}; len(breaches) > 0 || !slices.EqualFunc(numbers, want, maps.Equal) {
		t.Errorf("with a constraint: %v, %v; want %v", numbers, breaches, want)
	}
	// Floored division and its remainder take the sign of the divisor, and unary - binds tighter.
	numbers, breaches = Instances(edited(t, "    product: a * b\n", "    product: a * b\n    q: \"-a // b\"\n    r: a % -b\n"))
	if first := numbers[0]; len(breaches) > 0 || first["q"] != -2 || first["r"] != 0 || numbers[1]["q"] != -2 || numbers[1]["r"] != -2 {
		t.Errorf("floored: %v, %v", numbers, breaches)
	}
	// A template with a solver is counted as one without: its results come with the solvers.
	numbers, breaches = Instances(edited(t, "    product: a * b\n", "    product: a * b\n"+solver))
	if len(breaches) > 0 || len(numbers) != 6 {
		t.Errorf("with a solver: %v, %v", numbers, breaches)
	}
	if numbers, breaches := Instances(taskOf(t, strings.Replace(times, times[strings.Index(times, "template:"):strings.Index(times, "statement:")], "", 1))); numbers != nil || breaches != nil {
		t.Errorf("a task that is no template gives %v, %v", numbers, breaches)
	}
}

// Every rule of templates is broken at each place, and only when names, expressions and
// placeholders break none are the combinations counted.
func TestInstancesBreaches(t *testing.T) {
	for _, c := range []struct {
		name  string
		task  formatv0.Task
		found []string
	}{
		{"a parameter and a derived number alike", edited(t, "    product: a * b\n", "    product: a * b\n    a: b + 1\n"),
			[]string{"task.template-names #/template/derived/a a is both a parameter and a derived number"}},
		{"a result alike a parameter and a derived number", edited(t, "    product: a * b\n", "    product: a * b\n"+strings.Replace(solver, "[factor]", "[factor, b, product]", 1)),
			[]string{
				"task.template-names #/template/solver/results/1 b is both a parameter and a result of the solver",
				"task.template-names #/template/solver/results/2 product is both a derived number and a result of the solver",
			}},
		{"a derived number of a derived number, and a constraint of an int", edited(t, "    product: a * b\n", "    product: a * b\n    twice: product * 2\n  constraint: a + b\n"),
			[]string{
				"task.template-expressions #/template/derived/twice the derived number twice is no expression of the subset of templates: product at 1 is none of the names that it may name: a, b",
				"task.template-expressions #/template/constraint the constraint is an int expression, and it has to be a bool one",
			}},
		{"a derived number of a bool", edited(t, "    product: a * b\n", "    product: a < b\n"),
			[]string{"task.template-expressions #/template/derived/product the derived number product is a bool expression, and it has to be an int one"}},
		{"expressions of the subset that go.starlark.net does not parse", edited(t, "    product: a * b\n", "    product: \" a * b\"\n  constraint: \""+strings.Repeat("(", 76)+"a > 0"+strings.Repeat(")", 76)+"\"\n"),
			[]string{
				"task.template-expressions #/template/derived/product the derived number product is no expression of the dialect of the format, as go.starlark.net parses it: 1:2: got indent, want primary expression",
				"task.template-expressions #/template/constraint the constraint is no expression of the dialect of the format, as go.starlark.net parses it: 1:78: excessive nesting",
			}},
		{"the deepest that it parses", edited(t, "    product: a * b\n", "    product: \""+strings.Repeat("(", 75)+"a"+strings.Repeat(")", 75)+" + "+strings.Repeat("-", 980)[:900]+"b\"\n"), nil},
		{"a placeholder of nothing, and a lone brace", edited(t, `"= {product}"`, `"= {c}"`, "leak_phrases: [\"{b} times\"]", "leak_phrases: [\"{b} times }\"]"),
			[]string{
				"task.placeholders #/items/0/line/2/text {c} names no parameter, derived number or result of the solver",
				"task.placeholders #/leak_phrases/0 a } at 11 closes no placeholder: a brace of the text is written twice",
			}},
		{"a result in the key, a trap value, the solution and a leak phrase", edited(t, "    product: a * b\n", "    product: a * b\n"+solver,
			`key: "{b}"`, `key: "{factor}"`, `- value: "{product}"`, `- value: "{factor}"`, "{a} × {b} = {product}.", "{a} × {factor} = {product}.",
			"leak_phrases: [\"{b} times\"]", "leak_phrases: [\"{factor} times\"]"), nil},
		{"a result in the statement, a text of a line and a trap's text", edited(t, "    product: a * b\n", "    product: a * b\n"+solver,
			"Write the missing number.", "Write {factor}.", `"{a} ×"`, `"{factor} ×"`, "trap: copied-the-product\n", "trap: copied-the-product\n              text: Not {factor}.\n"),
			[]string{
				"task.placeholders #/items/0/line/0/text {factor} is a result of the solver, which stands only in a key or a trap value of a number or a fraction, in the solution and in a leak phrase",
				"task.placeholders #/items/0/line/1/check/trap_values/0/text {factor} is a result of the solver, which stands only in a key or a trap value of a number or a fraction, in the solution and in a leak phrase",
				"task.placeholders #/statement/0/text {factor} is a result of the solver, which stands only in a key or a trap value of a number or a fraction, in the solution and in a leak phrase",
			}},
		{"an empty domain", edited(t, "b: {from: 2, to: 3}", "b: {from: 3, to: 2}"),
			[]string{"task.template-instances #/template/parameters/b the domain of b is empty: its to, 2, is below its from, 3"}},
		{"too many combinations", edited(t, "b: {from: 2, to: 3}", "b: {from: -9007199254740991, to: 9007199254740991}"),
			[]string{"task.template-instances #/template/parameters its domains multiply to 54043195528445949 combinations, more than 10000"}},
		{"a derived number that divides by zero", edited(t, "    product: a * b\n", "    product: a * b\n    q: a % (b - 3)\n"),
			[]string{"task.template-instances #/template/derived/q on a = 4, b = 3, the derived number q fails: integer modulo by zero"}},
		{"a derived number past the integers of YAML", edited(t, "a: {from: 4, to: 6}", "a: {from: 4503599627370496, to: 4503599627370496}"),
			[]string{"task.template-instances #/template/derived/product on a = 4503599627370496, b = 2, the derived number product is 9007199254740992, past ±(2^53 − 1)"}},
		{"a derived number just within them", edited(t, "a: {from: 4, to: 6}", "a: {from: -4503599627370495, to: -4503599627370495}", "{from: 2, to: 3}", "{from: 2, to: 2}"), nil},
		{"a constraint that divides by zero", edited(t, "    product: a * b\n", "    product: a * b\n  constraint: a // (b - 2) > 0\n"),
			[]string{"task.template-instances #/template/constraint on a = 4, b = 2, the constraint fails: floored division by zero"}},
		{"no instance", edited(t, "    product: a * b\n", "    product: a * b\n  constraint: product > 18\n"),
			[]string{"task.template-instances #/template/constraint no combination of its parameters meets the constraint"}},
	} {
		numbers, breaches := Instances(c.task)
		var found []string
		for _, b := range breaches {
			found = append(found, b.Rule+" "+b.Pointer+" "+b.Message)
		}
		if !slices.Equal(found, c.found) || len(breaches) > 0 && numbers != nil {
			t.Errorf("%s: %q, %v; want %q", c.name, found, numbers, c.found)
		}
	}
}

// An instance is the template with its numbers in place, in a single pass, without its template.
func TestInstance(t *testing.T) {
	task := edited(t, "Write the missing number.", "Write {{the}} missing number, {{{b}}}.", "    product: a * b\n", "    product: a * b - 20\n")
	got, err := Instance(task, formatv0.Numbers{"a": 4, "b": 3, "product": -8})
	if err != nil {
		t.Fatal(err)
	}
	if got.Template != nil || got.ID != "times" || *got.Statement[0].Text != "Write {the} missing number, {3}." || *got.Items[0].Line[0].Text != "4 ×" ||
		*got.Items[0].Line[1].Check.Key != "3" || got.Items[0].Line[1].Check.TrapValues[0].Value != "-8" || *got.Solution[0].Text != "4 × 3 = -8." ||
		got.LeakPhrases[0] != "3 times" {
		t.Errorf("the instance is %+v", got)
	}
	// The template is as it was.
	if *task.Items[0].Line[1].Check.Key != "{b}" || task.Template == nil {
		t.Error("Instance changed the template")
	}
	for _, c := range []struct {
		name    string
		task    formatv0.Task
		numbers formatv0.Numbers
		err     string
	}{
		{"no template", taskOf(t, strings.Replace(times, times[strings.Index(times, "template:"):strings.Index(times, "statement:")], "", 1)), nil,
			"template: the task times is no template"},
		{"a number left out", taskOf(t, times), formatv0.Numbers{"a": 4, "b": 3},
			"template: the task times: #/items/0/line/1/check/trap_values/0/value: no number of the instance is {product}"},
		{"a lone brace", edited(t, `"= {product}"`, `"= {product"`), formatv0.Numbers{"a": 4, "b": 3, "product": 12},
			"template: the task times: #/items/0/line/2/text: a { at 3 opens no placeholder: a placeholder is {name}, and a brace of the text is written twice"},
	} {
		if _, err := Instance(c.task, c.numbers); err == nil || err.Error() != c.err {
			t.Errorf("%s: %v, want %q", c.name, err, c.err)
		}
	}
}

// A breach names an instance by the numbers of its parameters, in the order of their names.
func TestShown(t *testing.T) {
	tmpl := *taskOf(t, times).Template
	if got := Shown(tmpl, formatv0.Numbers{"product": 8, "b": 2, "a": 4}); got != "a = 4, b = 2" {
		t.Errorf("%q", got)
	}
	if got := Shown(tmpl, formatv0.Numbers{"b": -2}); got != "b = -2" {
		t.Errorf("%q", got)
	}
}

// A placeholder of a result may stand in a key and in a trap value of a number or a fraction alone,
// not of a choice, nor in the units of a number.
func TestTakesResults(t *testing.T) {
	task := taskOf(t, `id: c
statement:
  - text: Choose.
canvas:
  kind: handwriting_sheet
items:
  - answer_slots:
      - binding: choice
        options: ["1", "2"]
        check: {kind: choice, key: "1", trap_values: [{value: "2", trap: t}]}
  - answer_slots:
      - binding: zone
        zone: {left: 0, top: 0, width: 10, height: 10}
        check: {kind: number, key: "1", units: [u], trap_values: [{value: "2", trap: t}]}
solution:
  - text: "1"
skills:
  primary: [s]
band: b
difficulty: 1
`)
	for _, c := range []struct {
		tokens []string
		takes  bool
	}{
		{[]string{"items", "0", "answer_slots", "0", "check", "key"}, false},
		{[]string{"items", "0", "answer_slots", "0", "check", "trap_values", "0", "value"}, false},
		{[]string{"items", "0", "answer_slots", "0", "options", "0"}, false},
		{[]string{"items", "1", "answer_slots", "0", "check", "key"}, true},
		{[]string{"items", "1", "answer_slots", "0", "check", "trap_values", "0", "value"}, true},
		{[]string{"items", "1", "answer_slots", "0", "check", "trap_values", "0", "trap"}, false},
		{[]string{"items", "1", "answer_slots", "0", "check", "units", "0"}, false},
		{[]string{"items", "1", "answer_slots", "0", "binding"}, false},
		{[]string{"solution", "0", "text"}, true},
		{[]string{"leak_phrases", "0"}, true},
		{[]string{"statement", "0", "text"}, false},
	} {
		if got := takesResults(task, c.tokens); got != c.takes {
			t.Errorf("%v: %v, want %v", c.tokens, got, c.takes)
		}
	}
}
