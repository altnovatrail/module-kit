package module

import (
	"fmt"
	"io/fs"
	"slices"
	"strings"
	"testing"

	"github.com/altnovatrail/spec"
	"github.com/altnovatrail/spec/types/formatv0"
)

// remainders is a task on a workbook whose line a worked example can solve: 17 ÷ 5 = 3 r 2.
const remainders = `id: remainders-1
statement:
  - text: Write the quotient and the remainder.
canvas:
  kind: workbook
items:
  - line:
      - text: "17 ÷ 5 ="
      - binding: box
        check: {kind: number, key: "3"}
      - text: r
      - binding: box
        check: {kind: number, key: "2"}
solution:
  - text: "17 ÷ 5 = 3 r 2."
leak_phrases: [three heaps]
skills:
  primary: [counting-gaps]
band: grades-3-4
difficulty: 1
`

// exampleLine gives a line of a worked example: a text piece for each string, and a box for each
// [1]string, holding its value.
func exampleLine(pieces ...any) formatv0.ExampleLine {
	var line formatv0.ExampleLine
	for _, p := range pieces {
		switch p := p.(type) {
		case string:
			line.Line = append(line.Line, formatv0.ExampleLinePart{Text: new(formatv0.Text(p))})
		case [1]string:
			line.Line = append(line.Line, formatv0.ExampleLinePart{Binding: new("box"), Value: new(formatv0.Text(p[0]))})
		}
	}
	return line
}

// A line of a worked example solves a line of a task when it has the same pieces in the same order,
// its texts are the task's in their normal form, and each box holds what its check takes for the key;
// otherwise it solves nothing (format/README.md, section 20).
func TestSolves(t *testing.T) {
	item := taskValue(t, remainders).Items[0]
	for _, c := range []struct {
		name string
		line formatv0.ExampleLine
		want bool
	}{
		{"the task's line", exampleLine("17 ÷ 5 =", [1]string{"3.0"}, "r", [1]string{"2"}), true},
		{"its texts in another form", exampleLine("17  ÷  5  =", [1]string{"3"}, "R", [1]string{"2"}), true},
		{"another value", exampleLine("17 ÷ 5 =", [1]string{"4"}, "r", [1]string{"2"}), false},
		{"another text", exampleLine("18 ÷ 5 =", [1]string{"3"}, "r", [1]string{"2"}), false},
		{"a box where a text stands", exampleLine([1]string{"17"}, [1]string{"3"}, "r", [1]string{"2"}), false},
		{"fewer pieces", exampleLine("17 ÷ 5 =", [1]string{"3"}), false},
		{"no pieces", formatv0.ExampleLine{}, false},
	} {
		if got := solves(c.line, item); got != c.want {
			t.Errorf("%s: solves = %v, want %v", c.name, got, c.want)
		}
	}
	predicate := item
	predicate.Line = slices.Clone(item.Line)
	predicate.Line[1] = formatv0.LinePart{Binding: item.Line[1].Binding, Check: &formatv0.WrittenCheck{Kind: formatv0.CheckKindPredicate, Predicate: new("value == 3")}}
	if solves(exampleLine("17 ÷ 5 =", [1]string{"3"}, "r", [1]string{"2"}), predicate) {
		t.Error("a box of a predicate, which has no key, is solved")
	}
}

// A text piece of a line of a worked example is searched for the leak phrases of a task, and for no
// key: its numbers are the example's own.
func TestExampleTextsOnlyForPhrases(t *testing.T) {
	task := taskValue(t, remainders)
	var c checker
	texts := []searchedText{
		{text: "Share 17 into three heaps, 3 and 2", place: "#/a", phrasesOnly: true},
		{text: "Share 17 into heaps of 3 and 2", place: "#/b", phrasesOnly: true},
		{text: "3 and 2", place: "#/c"},
	}
	c.lessonTaskLeaks("lessons/remainders.yaml", 1, "tasks/remainders-1.yaml", task, texts)
	var got []string
	for _, b := range c.breaches {
		got = append(got, b.Pointer)
	}
	if !slices.Equal(got, []string{"#/a", "#/c"}) {
		t.Errorf("the breaches are at %q, want #/a for the phrase and #/c for the keys", got)
	}
}

// A choice of notation: word has no option that its normalization leaves empty, and no two alike
// once their ends are dropped.
func TestWordChoiceOptions(t *testing.T) {
	task := strings.Replace(remainders, `check: {kind: number, key: "3"}`, `check: {kind: choice, notation: word, options: [Kim, "?", Lea], key: Kim}`, 1)
	report, err := CheckFile(partTask, []byte(task))
	if err != nil {
		t.Fatal(err)
	}
	if got := shortReport(report); !slices.Equal(got, []string{"#/items/0/line/1/check/options/1 task.choice-options"}) {
		t.Errorf("the breaches are %q", got)
	}
}

// A task on a workbook of 100 short items is refused at the first item of the row with no room, the
// 21st, long before most_items of page.yaml, and a value of a worked example wider than its box at the
// value: seven digits are 112 units wide, and a box 100.
func TestPageLimits(t *testing.T) {
	var items strings.Builder
	for range 100 {
		items.WriteString("  - line:\n      - binding: box\n        check: {kind: number, key: \"1\"}\n")
	}
	task := strings.Replace(remainders, remainders[strings.Index(remainders, "items:"):strings.Index(remainders, "solution:")], "items:\n"+items.String(), 1)
	report, err := CheckFile(partTask, []byte(task))
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Breaches) != 1 || report.Breaches[0].Rule != RuleTaskWorkbookPage || report.Breaches[0].Pointer != "#/items/20" ||
		!strings.Contains(report.Breaches[0].Message, "the row of item 21 has no room") {
		t.Errorf("100 items give %v", report)
	}

	lesson, err := fs.ReadFile(spec.Format, "examples/lesson/valid/remainders-on-a-workbook.yaml")
	if err != nil {
		t.Fatal(err)
	}
	wide := strings.Replace(string(lesson), `value: "5"`, `value: "1234567"`, 1)
	report, err = CheckFile(partLesson, []byte(wide))
	if err != nil {
		t.Fatal(err)
	}
	if got := shortReport(report); !slices.Equal(got, []string{"#/segments/0/worked_example/lines/0/line/1/value lesson.workbook-page"}) ||
		!strings.Contains(report.Breaches[0].Message, fmt.Sprintf("wider than its box of %d", 100)) {
		t.Errorf("a wide value gives %v", report)
	}
}
