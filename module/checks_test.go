package module

import (
	"maps"
	"slices"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/altnovatrail/spec/types/formatv0"

	"github.com/altnovatrail/module-kit/verdict"
)

// The tasks of the examples that the cases of the rules of a check edit.
var (
	workbook = string(examplePart("task", "workbook-page").Data)
	parity   = string(examplePart("task", "even-or-odd").Data)
	pizza    = string(examplePart("task", "fraction-in-lowest-terms").Data)
	sentence = string(examplePart("task", "best-first-sentence").Data)
)

// The cases hold each rule that the check of an answer slot sees itself at each place it is broken,
// on the file of a task, as CheckFile reads it.
var slotRuleCases = []struct {
	name, data string
	refused    []string
}{
	{"a workbook page as it is", workbook, nil},
	{"a key that is none of the options written by hand", replaced(workbook, `key: "<"`, `key: "≤"`),
		[]string{"#/items/2/line/1/check/key task.choice-key"}},
	{"options alike but for width", replaced(workbook, `options: ["<", ">", "="]`, `options: ["<", ">", "＜", "="]`),
		[]string{"#/items/2/line/1/check/options/2 task.choice-options"}},
	{"a key that is none of the options tapped", replaced(sentence, "key: The door was open, and nobody had opened it.", "key: The door was open."),
		[]string{"#/items/0/answer_slots/0/check/key task.choice-key"}},
	{"options tapped alike but for letter case", replaced(sentence, "- I am going to write about a day.\n", "- I am going to write about a day.\n          - it was a day like any other.\n"),
		[]string{"#/items/0/answer_slots/0/options/3 task.choice-options"}},
	{"a key of a fraction in lowest terms that can be reduced", replaced(pizza, `key: "1/4"`, `key: "2/8"`, `- value: "2/8"`, `- value: "4/16"`),
		[]string{"#/items/0/line/1/check/key task.fraction-key"}},
	{"a trap value that is the key", replaced(workbook, `- value: "9"`, `- value: "8"`),
		[]string{"#/items/0/line/1/check/trap_values/0/value task.trap-values"}},
	{"a trap value of the key's value, with the units", replaced(workbook, `- value: "7"`, `- value: "12.0"`),
		[]string{"#/items/3/line/1/check/trap_values/0/value task.trap-values"}},
	{"a trap value of the key's value in any form", replaced(pizza, "          form: reduced\n", ""),
		[]string{"#/items/0/line/1/check/trap_values/0/value task.trap-values"}},
	{"two trap values alike", replaced(workbook, "              trap: times-table-slip\n              text: Count",
		"              trap: times-table-slip\n            - value: \"9.0\"\n              trap: times-table-slip\n              text: Count"),
		[]string{"#/items/0/line/1/check/trap_values/1/value task.trap-values"}},
	{"a trap value of a choice not as written", replaced(parity, "- value: even", "- value: EVEN"),
		[]string{"#/items/0/answer_slots/0/check/trap_values/0/value task.trap-values"}},
	{"a trap value of a choice that is the key", replaced(parity, "- value: even", "- value: Odd"),
		[]string{"#/items/0/answer_slots/0/check/trap_values/0/value task.trap-values"}},
	{"a trap value of a choice that is none of the options", replaced(parity, "- value: even", "- value: seven"),
		[]string{"#/items/0/answer_slots/0/check/trap_values/0/value task.trap-values"}},
	{"a placeholder outside a template is another rule's", replaced(workbook, `key: "8"`, `key: "{b}"`, `- value: "9"`, `- value: "{b}"`), nil},
	{"a template is read through its instances", string(examplePart("task", "fraction-template").Data), nil},
}

func TestCheckRules(t *testing.T) {
	for _, c := range slotRuleCases {
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

// Check runs them on each task of a module too, as a rule within a part.
func TestCheckRulesOfAModule(t *testing.T) {
	report, err := Check(module(fstest.MapFS{
		campDays: edit(campDays, `check: {kind: number, key: "9", units: [day, days]}`,
			`check: {kind: number, key: "9", units: [day, days], trap_values: [{value: "9.0", trap: off-by-one}]}`),
	}))
	if err != nil {
		t.Fatal(err)
	}
	want := []string{campDays + " #/items/0/answer_slots/0/check/trap_values/0/value task.trap-values"}
	if got := shortReport(report); !slices.Equal(got, want) {
		t.Errorf("refused by %q, want %q\n%s", got, want, report)
	}
}

// Every breach of a rule that the check of an answer slot sees itself says what breaks the rule.
func TestCheckRuleMessages(t *testing.T) {
	messages := map[string]bool{}
	for _, c := range slotRuleCases {
		report, err := CheckFile("task", []byte(c.data))
		if err != nil {
			t.Fatal(err)
		}
		for _, b := range report.Breaches {
			messages[b.Rule+": "+b.Message] = true
		}
	}
	for _, want := range []string{
		"task.choice-key: the key ≤ is none of the options as written",
		"task.choice-key: the key The door was open. is none of the options as written",
		"task.choice-options: the option ＜ is the option < before it, once normalized as a choice normalizes them",
		"task.choice-options: the option it was a day like any other. is the option It was a day like any other. before it, once normalized as a choice normalizes them",
		"task.fraction-key: the key 2/8 can be reduced, and the check takes only a fraction in lowest terms",
		"task.trap-values: the trap value 8 gets correct from the check, not not_yet: a trap value is a wrong value",
		"task.trap-values: the trap value 12.0 gets correct from the check, not not_yet: a trap value is a wrong value",
		"task.trap-values: the trap value 2/8 gets correct from the check, not not_yet: a trap value is a wrong value",
		"task.trap-values: the trap value 9.0 is equal to the trap value 9 before it, as the check compares them",
		"task.trap-values: the trap value EVEN is none of the options as written",
		"task.trap-values: the trap value Odd gets correct from the check, not not_yet: a trap value is a wrong value",
		"task.trap-values: the trap value seven gets unreadable (notation) from the check, not not_yet: a trap value is a wrong value",
	} {
		if !messages[want] {
			t.Errorf("no breach says %q; they say\n%s", want, strings.Join(slices.Sorted(maps.Keys(messages)), "\n"))
		}
	}
}

// A check that its schema accepts and that gives no verdict is a fault of module-kit.
func TestVerdictOfFault(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("a check that gives no verdict does not panic")
		}
	}()
	verdictOf(verdict.AnswerSlot{Binding: "box", Check: formatv0.Check{Kind: formatv0.CheckKindNumber}}, "7")
}
