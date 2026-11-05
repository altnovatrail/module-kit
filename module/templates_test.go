package module

import (
	"maps"
	"slices"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/altnovatrail/module-kit/internal/schema"
)

// The templates of the examples that the cases of the rules of templates edit.
var (
	missingFactor = string(examplePart("task", "missing-factor").Data)
	thoughtOf     = string(examplePart("task", "number-thought-of-template").Data)
)

// The module with-templates, whose task missing-factor is a template and whose lesson names it, and
// the paths of the two.
var (
	withTemplates    = example("with-templates")
	templatePath     = "tasks/missing-factor.yaml"
	templateLesson   = "lessons/missing-factors.yaml"
	templateTraps    = "traps.yaml"
	sayTheTimesTable = "Say the times table aloud, one step at a time, and stop where the product is."
)

// The cases hold the rules of templates and the rules of a task on its instances, each at a place it
// is broken, on the file of a task, as CheckFile reads it.
var templateRuleCases = []struct {
	name, data string
	refused    []string
}{
	{"a template as it is", missingFactor, nil},
	{"a template with a solver as it is", thoughtOf, nil},
	{"its own rules", replaced(missingFactor, "    one_more: second + 1\n", "    one_more: second / 1\n", `"{first} ×"`, `"{first} × {"`),
		[]string{"#/items/0/line/0/text task.placeholders", "#/template/derived/one_more task.template-expressions"}},
	{"its own rules, and no instance read", replaced(missingFactor, "    one_more: second + 1\n", "    one_more: second + 1\n  constraint: first > 9\n",
		`- value: "{one_more}"`, `- value: "{second}"`), []string{"#/template/constraint task.template-instances"}},
	{"an instance that its schema does not take", replaced(missingFactor, "second: {from: 2, to: 9}", "second: {from: -1000000000000000, to: -999999999999999}"),
		[]string{"#/items/0/line/1/check/key task.instances"}},
	{"a rule of an answer slot on every instance, at one place", replaced(missingFactor, `- value: "{one_more}"`, `- value: "{second}"`),
		[]string{"#/items/0/line/1/check/trap_values/0/value task.trap-values"}},
	{"a rule of an answer slot on some instances", replaced(missingFactor, `- value: "{one_more}"`, `- value: "{first}"`),
		[]string{"#/items/0/line/1/check/trap_values/0/value task.trap-values"}},
	{"a text of the task that holds the key of an instance", replaced(missingFactor, "              text: Count up", "              text: Not 7. Count up"),
		[]string{"#/items/0/line/1/check/trap_values/0/text task.no-leaks"}},
	{"a template with a solver that breaks its own rules", replaced(thoughtOf, "    forgot_ben: \"2 * left\"\n", "    forgot_ben: \"2 * left\"\n    left: to_ben\n"),
		[]string{"#/template/derived/left task.template-names"}},
	{"a solver that fails on an instance", replaced(thoughtOf, "range(1001)", "range(20)"),
		[]string{"#/template/solver/program task.template-solver"}},
	{"a rule of an answer slot on the results of a solver", replaced(thoughtOf, `- value: "{before_ben}"`, `- value: "{sweets}"`),
		[]string{"#/items/0/answer_slots/0/check/trap_values/0/value task.trap-values"}},
	{"a text of the task that holds a key that a solver finds", replaced(thoughtOf, "Suppose Masha gave no sweets to Ben.", "Suppose Masha gave 10 sweets to Ben."),
		[]string{"#/items/0/simpler_case task.no-leaks"}},
	{"a placeholder outside a template", replaced(workbook, `key: "8"`, `key: "{b}"`, `- value: "7"`, `- value: "{b}"`),
		[]string{"#/items/0/line/1/check task.placeholders", "#/items/3/line/1/check task.placeholders"}},
}

func TestTemplateRules(t *testing.T) {
	for _, c := range templateRuleCases {
		t.Run(c.name, func(t *testing.T) {
			report, err := CheckFile("task", []byte(c.data))
			if err != nil {
				t.Fatal(err)
			}
			if got := shortReport(report); !slices.Equal(got, c.refused) {
				t.Errorf("refused by %q, want %q\n%s", got, c.refused, report)
			}
		})
	}
}

// Check reads the instances of a template for the rules between parts too: a text of another part
// and a lesson that names the template are searched for the keys of every instance. A rule that reads
// no string of a task reads the template as it is written, once.
func TestTemplateRulesOfAModule(t *testing.T) {
	edit := func(name, old, new string) *fstest.MapFile {
		return file(replaced(string(withTemplates[name].Data), old, new))
	}
	for _, c := range []struct {
		name    string
		fsys    fstest.MapFS
		refused []string
	}{
		{"with-templates as it is", withTemplates, nil},
		{"a trap's advice that holds the key of an instance", with(withTemplates, fstest.MapFS{
			templateTraps: edit(templateTraps, sayTheTimesTable, "Say the times table aloud, as far as 7 times."),
		}), []string{templatePath + " #/items/0/line/1/check/trap_values/0/trap task.no-leaks"}},
		{"a lesson that holds the key of an instance", with(withTemplates, fstest.MapFS{
			templateLesson: edit(templateLesson, "and stop at the product the line gives.", "and stop at 8."),
		}), []string{templateLesson + " #/segments/0/explanation/0/text lesson.no-leaks"}},
		{"a rule between parts, once", with(withTemplates, fstest.MapFS{
			templatePath: edit(templatePath, "band: grade-3", "band: grade-9"),
		}), []string{templatePath + " #/band task.known-band"}},
	} {
		t.Run(c.name, func(t *testing.T) {
			report, err := Check(c.fsys)
			if err != nil {
				t.Fatal(err)
			}
			if got := shortReport(report); !slices.Equal(got, c.refused) {
				t.Errorf("refused by %q, want %q\n%s", got, c.refused, report)
			}
		})
	}
}

// Every breach of a rule of templates says what breaks it, and a breach on an instance names the
// numbers of the first instance that breaks the rule at its place.
func TestTemplateRuleMessages(t *testing.T) {
	messages := map[string]bool{}
	for _, c := range templateRuleCases {
		report, err := CheckFile("task", []byte(c.data))
		if err != nil {
			t.Fatal(err)
		}
		for _, b := range report.Breaches {
			messages[b.Rule+": "+b.Message] = true
		}
	}
	report, err := Check(with(withTemplates, fstest.MapFS{
		templateTraps:  file(replaced(string(withTemplates[templateTraps].Data), sayTheTimesTable, "Say the times table aloud, as far as 7 times.")),
		templateLesson: file(replaced(string(withTemplates[templateLesson].Data), "and stop at the product the line gives.", "and stop at 8.")),
	}))
	if err != nil {
		t.Fatal(err)
	}
	for _, b := range report.Breaches {
		messages[b.Rule+": "+b.Message] = true
	}
	for _, want := range []string{
		"task.placeholders: a { at 11 opens no placeholder: a placeholder is {name}, and a brace of the text is written twice",
		"task.template-expressions: the derived number one_more is no expression of the subset of templates: '/' at 8 is no part of the expressions of templates",
		"task.template-instances: no combination of its parameters meets the constraint",
		"task.instances: on the instance first = 2, second = -1000000000000000: the instance is not valid against task.schema.json: maxLength: maxLength: got 17, want 16",
		"task.trap-values: on the instance first = 2, second = 2: the trap value 2 gets correct from the check, not not_yet: a trap value is a wrong value",
		`task.no-leaks: on the instance first = 2, second = 7: the text holds the key "7" of the answer slot at #/items/0/line/1, and the learner can see it before the task is solved`,
		`task.no-leaks: on the instance first = 2, second = 7: the text at traps.yaml#/traps/times-table-slip/trap_advice holds the key "7" of the answer slot at #/items/0/line/1, and the learner can see it before the task is solved`,
		`lesson.no-leaks: on the instance first = 2, second = 8: the text holds the key "8" of the answer slot at tasks/missing-factor.yaml#/items/0/line/1, of the task that #/segments/1/task names, and the learner sees the whole lesson before the task is solved`,
		"task.template-names: left is both a parameter and a derived number",
		"task.template-solver: on left = 4, to_ben = 6, solve fails: fail: 0 whole numbers fit the story, at 11:13 in solve",
		"task.trap-values: on the instance left = 2, to_ben = 3: the trap value 10 gets correct from the check, not not_yet: a trap value is a wrong value",
		`task.no-leaks: on the instance left = 2, to_ben = 3: the text holds the key "10" of the answer slot at #/items/0/answer_slots/0, and the learner can see it before the task is solved`,
		"task.placeholders: the key or a trap value holds the placeholder of a template, and the task is no template: no instance puts a number in its place",
	} {
		if !messages[want] {
			t.Errorf("no breach says %q; they say\n%s", want, strings.Join(slices.Sorted(maps.Keys(messages)), "\n"))
		}
	}
}

// task.instances names each keyword at each place by the first instance that fails it there: an
// instance that fails another keyword at the same place is a mistake of its own.
func TestInstancesByKeyword(t *testing.T) {
	at := []string{"items", "0", "line", "1", "check", "key"}
	c := checker{}
	c.templateRules("tasks/t.yaml", &expansion{instances: []instance{
		{shown: "a = 1", invalid: []schema.Error{{Keyword: "pattern", Location: at, Message: "no match"}}},
		{shown: "a = 2", invalid: []schema.Error{{Keyword: "pattern", Location: at, Message: "no match"}}},
		{shown: "a = 3", invalid: []schema.Error{{Keyword: "maxLength", Location: at, Message: "too long"}}},
	}})
	var got []string
	for _, b := range c.report().Breaches {
		got = append(got, b.Rule+": "+b.Message)
	}
	want := []string{
		"task.instances: on the instance a = 1: the instance is not valid against task.schema.json: pattern: no match",
		"task.instances: on the instance a = 3: the instance is not valid against task.schema.json: maxLength: too long",
	}
	if !slices.Equal(got, want) {
		t.Errorf("%q, want %q", got, want)
	}
}
