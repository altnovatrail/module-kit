package module

import (
	"context"
	"maps"
	"slices"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/altnovatrail/spec/types/formatv0"

	"github.com/altnovatrail/module-kit/internal/yaml"
)

// taskValue reads a task that its schema accepts into its type.
func taskValue(t *testing.T, text string) formatv0.Task {
	t.Helper()
	value, err := yaml.Read([]byte(text))
	if err != nil {
		t.Fatal(err)
	}
	return typed[formatv0.Task](value)
}

// The texts of base that the cases put a leak in.
const (
	gapsGiven    = "It is 180 cm from the lowest rung to the highest."
	gapsUnknown  = "How many rungs there are."
	gapsTrapText = "How many rungs hold those gaps?"
	gapsSimpler  = "How many rungs has it?"
	offByOne     = "count each."
	ignored      = "check the answer against each."
	firstRung    = "Which step of your work gave you this answer?"
	fenceAlt     = "with one dot more than dashes."
	lampPath     = "A path is 20 m long"
	lampDemo     = "Draw the path, and mark a lamp at each end."
	gapsAfter    = "and then ask what stands at the ends."
)

// The cases hold task.no-leaks and lesson.no-leaks at each place they are broken, on a whole module.
var leakCases = []struct {
	name    string
	fsys    fstest.MapFS
	refused []string
}{
	{"the texts of a task", module(fstest.MapFS{
		gapsLadder: edit(gapsLadder, gapsGiven, "There are 7 rungs.", gapsUnknown, "How many rungs there are, 7 or Seven Rungs?",
			gapsTrapText, "Not 7 of them.", gapsSimpler, "How many rungs has it? Then 7."),
	}), []string{
		gapsLadder + " #/givens/1 task.no-leaks", gapsLadder + " #/items/0/answer_slots/0/check/trap_values/0/text task.no-leaks",
		gapsLadder + " #/items/0/simpler_case task.no-leaks", gapsLadder + " #/unknown task.no-leaks",
	}},
	{"the techniques and the traps that a task names, a trap named twice once", module(fstest.MapFS{
		techniquesPath: edit(techniquesPath, "title: Draw it", "title: Draw all 7", "spot the rule", "spot seven rungs"),
		trapsPath:      edit(trapsPath, offByOne, "count each, 7 in all.", ignored, "check the answer against each, seven rungs."),
	}), []string{
		gapsLadder + " #/items/0/answer_slots/0/check/trap_values/0/trap task.no-leaks", gapsLadder + " #/techniques/0 task.no-leaks",
		gapsLadder + " #/techniques/1 task.no-leaks", gapsLadder + " #/traps/0 task.no-leaks",
	}},
	{"a rung template, for each task with an item with a verdict that it gives away", module(fstest.MapFS{
		manifestPath: edit(manifestPath, firstRung, "Which step of your work gave you 9?"),
	}), []string{campDays + " # task.no-leaks"}},
	{"a number that no key of the task has", module(fstest.MapFS{
		gapsLadder:     edit(gapsLadder, gapsGiven, "There are 6 gaps."),
		techniquesPath: edit(techniquesPath, "title: Draw it", "title: Draw all 9"),
	}), nil},
	{"the texts of lessons, after the task too, and the texts of a lesson that does not name the task", module(fstest.MapFS{
		daysLesson: edit(daysLesson, "title: Days and dates", "title: Nine days and dates", "is counting posts.", "is counting posts, as 7 rungs."),
		gapsLesson: edit(gapsLesson, fenceAlt, "with 7 dots.", lampPath, "A path of seven rungs is 20 m long",
			lampDemo, lampDemo+" 7 is not the answer here.", gapsAfter, gapsAfter+" Seven rungs, 9 days."),
	}), []string{
		daysLesson + " #/title lesson.no-leaks", gapsLesson + " #/segments/0/explanation/1/alt lesson.no-leaks",
		gapsLesson + " #/segments/1/worked_example/demo/0/text lesson.no-leaks",
		gapsLesson + " #/segments/1/worked_example/statement/0/text lesson.no-leaks", gapsLesson + " #/segments/3/explanation/0/text lesson.no-leaks",
	}},
}

func TestLeakRules(t *testing.T) {
	for _, c := range leakCases {
		t.Run(c.name, func(t *testing.T) {
			report, err := Check(context.Background(), c.fsys)
			if err != nil {
				t.Fatal(err)
			}
			if got := shortReport(report); !slices.Equal(got, c.refused) {
				t.Errorf("refused by\n%s\nwant\n%s\nreport:\n%s", strings.Join(got, "\n"), strings.Join(c.refused, "\n"), report)
			}
		})
	}
}

// CheckFile searches the texts of a task itself, and no text of another part: it has none.
func TestLeakRulesOfAFile(t *testing.T) {
	for _, c := range []struct {
		data    string
		refused []string
	}{
		{string(base[gapsLadder].Data), nil},
		{replaced(string(base[gapsLadder].Data), gapsGiven, "There are 7 rungs."), []string{"#/givens/1 task.no-leaks"}},
		{replaced(string(examplePart("task", "fraction-template").Data), "statement:\n", "givens: [\"3/8\"]\nstatement:\n"),
			[]string{"#/givens/0 task.no-leaks"}},
	} {
		report, err := CheckFile("task", []byte(c.data))
		if err != nil {
			t.Fatal(err)
		}
		if got := shortReport(report); !slices.Equal(got, c.refused) {
			t.Errorf("refused by %q, want %q\n%s", got, c.refused, report)
		}
	}
}

// Every breach of a rule of the leak check says what text holds what of the task, and where.
func TestLeakRuleMessages(t *testing.T) {
	messages := map[string]bool{}
	for _, c := range leakCases {
		report, err := Check(context.Background(), c.fsys)
		if err != nil {
			t.Fatal(err)
		}
		for _, b := range report.Breaches {
			messages[b.Rule+": "+b.Message] = true
		}
	}
	for _, want := range []string{
		`task.no-leaks: the text holds the key "7" of the answer slot at #/items/0/answer_slots/0, and the learner can see it before the task is solved`,
		`task.no-leaks: the text holds the key "7" of the answer slot at #/items/0/answer_slots/0 and the leak phrase "seven rungs" at #/leak_phrases/0, and the learner can see it before the task is solved`,
		`task.no-leaks: the text at techniques.yaml#/techniques/draw/title holds the key "7" of the answer slot at #/items/0/answer_slots/0, and the learner can see it before the task is solved`,
		`task.no-leaks: the text at techniques.yaml#/techniques/small-numbers/description holds the leak phrase "seven rungs" at #/leak_phrases/0, and the learner can see it before the task is solved`,
		`task.no-leaks: the text at traps.yaml#/traps/off-by-one/trap_advice holds the key "7" of the answer slot at #/items/0/answer_slots/0, and the learner can see it before the task is solved`,
		`task.no-leaks: the text at manifest.yaml#/rung_templates/0 holds the key "9" of the answer slot at #/items/0/answer_slots/0, and the learner can see it before the task is solved`,
		`lesson.no-leaks: the text holds the leak phrase "nine days" at tasks/camp-days-1.yaml#/leak_phrases/0, of the task that #/segments/1/task names, and the learner sees the whole lesson before the task is solved`,
		`lesson.no-leaks: the text holds the key "7" of the answer slot at tasks/gaps-ladder-1.yaml#/items/0/answer_slots/0, of the task that #/segments/2/task names, and the learner sees the whole lesson before the task is solved`,
	} {
		if !messages[want] {
			t.Errorf("no breach says %q; they say\n%s", want, strings.Join(slices.Sorted(maps.Keys(messages)), "\n"))
		}
	}
	ladder := taskValue(t, string(base[gapsLadder].Data))
	found := (&checker{}).searchOf(gapsLadder, ladder).In("7, seven rungs")
	if got := leaksSaid(ladder, "", append(found, found[1])); !strings.Contains(got, ", the leak phrase") || !strings.Contains(got, " and the leak phrase") {
		t.Errorf("three leaks are said %q", got)
	}
}

// A breach of lesson.no-leaks names the task's file by the path the lesson finds it at, even when the
// task's id is another, which breaks module.file-ids.
func TestLeakRuleNamesTheTasksFile(t *testing.T) {
	report, err := Check(context.Background(), module(fstest.MapFS{
		gapsLadder: edit(gapsLadder, "id: gaps-ladder-1", "id: gaps-ladder-2"),
		gapsLesson: edit(gapsLesson, gapsAfter, gapsAfter+" Seven rungs."),
	}))
	if err != nil {
		t.Fatal(err)
	}
	named := false
	for _, b := range report.Breaches {
		if b.Rule == RuleLessonNoLeaks {
			named = strings.Contains(b.Message, gapsLadder+"#/leak_phrases/0")
		}
	}
	if !named {
		t.Errorf("no breach of %s names %s:\n%s", RuleLessonNoLeaks, gapsLadder, report)
	}
}
