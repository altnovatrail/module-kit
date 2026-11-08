package leak

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"sync"
	"testing"

	"github.com/altnovatrail/spec/types/formatv0"

	"github.com/altnovatrail/module-kit/internal/schema"
	"github.com/altnovatrail/module-kit/internal/yaml"
)

// taskOf reads a task as the check of a module reads one: by the YAML rules, against its schema,
// and strictly into its type of formatv0.
func taskOf(t testing.TB, text string) formatv0.Task {
	t.Helper()
	value, err := yaml.Read([]byte(text))
	if err != nil {
		t.Fatal(err)
	}
	if errs := schema.Validate("task.schema.json", value); len(errs) > 0 {
		t.Fatalf("the task is refused by its schema: %v", errs)
	}
	task, err := schema.Decode[formatv0.Task](value)
	if err != nil {
		t.Fatal(err)
	}
	return task
}

// A task on a handwriting sheet whose items show the texts that free a key: the alt of a picture of
// the statement shows 5, the text of the second item 12, and the options of the first, a choice by
// tapping, its key 4. The options of a choice written by hand are closed and free nothing, and two
// answer slots have keys of one magnitude.
const sheet = `id: own-texts
statement:
  - text: Answer each.
  - picture: pictures/dots.png
    alt: A row of 5 dots
    width: 100
    height: 20
canvas:
  kind: handwriting_sheet
items:
  - text: a) 3 or 4?
    answer_slots:
      - binding: choice
        options: ["3", "4"]
        check: {kind: choice, key: "4"}
  - text: b) eggs in 12 boxes
    answer_slots:
      - binding: zone
        zone: {left: 0, top: 0, width: 100, height: 100}
        check: {kind: number, key: "12"}
  - text: c)
    answer_slots:
      - binding: zone
        zone: {left: 0, top: 200, width: 100, height: 100}
        check: {kind: choice, options: [odd, even], key: odd}
  - text: d)
    answer_slots:
      - binding: zone
        zone: {left: 0, top: 400, width: 100, height: 100}
        check: {kind: number, key: "7"}
      - binding: zone
        zone: {left: 200, top: 400, width: 100, height: 100}
        check: {kind: number, key: "-7"}
  - text: e)
    answer_slots:
      - binding: zone
        zone: {left: 0, top: 600, width: 100, height: 100}
        check: {kind: number, key: "5"}
  - text: f)
    answer_slots:
      - binding: zone
        zone: {left: 0, top: 800, width: 100, height: 100}
        check: {kind: predicate, predicate: "value > 100"}
solution:
  - text: Each in turn.
leak_phrases: [seven, "4"]
skills:
  primary: [counting]
band: first
difficulty: 1
`

// A workbook whose first line shows its key, 3, in a text piece, and whose second shows only another
// number. The third holds a placeholder outside a template, which is not searched for.
const lines = `id: lines
statement:
  - text: Write the missing numbers.
canvas:
  kind: workbook
items:
  - line:
      - text: "3 +"
      - binding: box
        check: {kind: number, key: "3"}
      - text: "= 6"
  - line:
      - binding: box
        check: {kind: fraction, key: "1/2"}
      - text: of 8 is 4
  - line:
      - binding: box
        check: {kind: number, key: "{count}"}
      - text: apples
solution:
  - text: Each in turn.
skills:
  primary: [counting]
band: first
difficulty: 1
`

// short writes what a search finds: each key by its value and its place, and each leak phrase by
// its text and its place.
func short(found []Leak) string {
	said := make([]string, len(found))
	for i, l := range found {
		if l.Kind == Key {
			said[i] = fmt.Sprintf("key %s %d.%d", l.Written, l.Item, l.Slot)
		} else {
			said[i] = fmt.Sprintf("phrase %s %d", l.Written, l.Phrase)
		}
	}
	return strings.Join(said, ", ")
}

func TestIn(t *testing.T) {
	for _, c := range []struct {
		task, text, want string
	}{
		{sheet, "4", "phrase 4 1"},
		{sheet, "12 eggs, 5 dots", ""},
		{sheet, "It is odd.", "key odd 2.0"},
		{sheet, "7 or -7", "key 7 3.0, key -7 3.1"},
		{sheet, "Seven!", "phrase seven 0"},
		{sheet, "odd 7 seven 4", "key odd 2.0, key 7 3.0, key -7 3.1, phrase seven 0, phrase 4 1"},
		{sheet, "seven4 oddly", ""},
		{sheet, "oddly odd", "key odd 2.0"},
		{sheet, "101", ""},
		{sheet, "", ""},
		// A byte that is not UTF-8 is U+FFFD, no character of a word.
		{sheet, "7\xffodd", "key odd 2.0, key 7 3.0, key -7 3.1"},
		{lines, "3 and 6", ""},
		{lines, "2/4 or 0.5", "key 1/2 1.0"},
		{lines, "{count} apples", ""},
	} {
		search, err := For(taskOf(t, c.task))
		if err != nil {
			t.Fatal(err)
		}
		if got := short(search.In(c.text)); got != c.want {
			t.Errorf("%q in %s: %q, want %q", c.text, taskOf(t, c.task).ID, got, c.want)
		}
	}
}

// A template is searched through its instances, and a key that is no string is none.
func TestFor(t *testing.T) {
	template := taskOf(t, strings.Replace(lines, "id: lines\n", "id: lines\ntemplate:\n  parameters:\n    count: {from: 1, to: 3}\n", 1))
	if _, err := For(template); !errors.Is(err, ErrTemplate) {
		t.Errorf("a template gives %v, want ErrTemplate", err)
	}
	task := taskOf(t, lines)
	task.Items[0].Line[1].Check.Key = nil
	search, err := For(task)
	if err != nil {
		t.Fatal(err)
	}
	if got := short(search.In("3 1/2")); got != "key 1/2 1.0" {
		t.Errorf("with no key on the first line, 3 1/2 gives %q", got)
	}
}

// The numbers of a text in its normal form, each the longest run of the form of a number from a
// digit with no character of a word right outside it (format/README.md, section 20).
func TestNumbersIn(t *testing.T) {
	for _, c := range []struct {
		text string
		want []string
	}{
		{"7", []string{"7"}},
		{"007 and 0", []string{"7", "0"}},
		{"3/4, 3 /4, 3/ 4 and 3 / 4", []string{"3/4", "3/4", "3/4", "3/4"}},
		{"3  / 4", []string{"3", "4"}},
		{"6/8", []string{"3/4"}},
		{"1/0 and 1/00 and 1/05", []string{"1", "0", "1", "0", "1/5"}},
		{"0/5", []string{"0"}},
		{"2.50", []string{"5/2"}},
		{"2.", []string{"2"}},
		{".5", []string{"5"}},
		{"1.5.3", []string{"3/2", "3"}},
		{"1/2.5", []string{"1/2", "5"}},
		{"2.5/3", []string{"5/2", "3"}},
		{"x7.5 7x 7.5x", nil},
		{"(7)-[8]", []string{"7", "8"}},
		{"7\u0301", nil},
		{"\u0661 7", []string{"7"}},
		{"\u06617", nil},
		{"9007199254740993", []string{"9007199254740993"}},
	} {
		var got []string
		for _, n := range numbersIn([]rune(c.text)) {
			got = append(got, n.RatString())
		}
		if !slices.Equal(got, c.want) {
			t.Errorf("the numbers of %q are %q, want %q", c.text, got, c.want)
		}
	}
}

// The numbers of a text are written as they stand in its normal form, in its order: NFKC writes ½ and
// the digits of full width as ASCII, and each number of a time or of a sum is one of its own.
func TestNumbers(t *testing.T) {
	for _, c := range []struct {
		text string
		want []string
	}{
		{"Count on in minutes: an hour has 60 of them, not 100.", []string{"60", "100"}},
		{"Check parity on small numbers first, such as 1 + 3 and 3 + 5.", []string{"1", "3", "3", "5"}},
		{"Half is ½, or 2 / 4, at 07:30.", []string{"1/2", "2 / 4", "07", "30"}},
		{"１２ rungs", []string{"12"}},
		{"Seven, the 7th, x7 and 1,5", []string{"1", "5"}},
		{"Draw a clock face and count on in minutes.", nil},
		{"\xff7\xff", []string{"7"}},
		{"", nil},
	} {
		if got := Numbers(c.text); !slices.Equal(got, c.want) {
			t.Errorf("the numbers of %q are %q, want %q", c.text, got, c.want)
		}
	}
}

// A phrase is found where it stands, with no character of a word right outside an end that is one.
func TestHolds(t *testing.T) {
	for _, c := range []struct {
		text, phrase string
		want         bool
	}{
		{"odd", "odd", true},
		{"(odd)", "odd", true},
		{"odds", "odd", false},
		{"oddly odd", "odd", true},
		{"56<65", "<", true},
		{"a<b", "<", true},
		{"x<", "x<", true},
		{"ax<", "x<", false},
		{"x<b", "x<", true},
		{"aaa", "aa", false},
		{"a a a", "a a", true},
		{"odd", "", false},
		{"", "odd", false},
		{"caf\u00E9", "cafe", false},
		{"\u0915\u093F\u0924\u093E", "\u0915\u093F\u0924", false},
	} {
		if got := holds(c.text, c.phrase); got != c.want {
			t.Errorf("holds(%q, %q) = %t, want %t", c.text, c.phrase, got, c.want)
		}
	}
}

// One search finds the same in many texts at once: it holds nothing that changes.
func TestConcurrentIn(t *testing.T) {
	search, err := For(taskOf(t, sheet))
	if err != nil {
		t.Fatal(err)
	}
	want := short(search.In("odd 7 seven 4"))
	var wg sync.WaitGroup
	for range 8 {
		wg.Go(func() {
			for range 100 {
				if got := short(search.In("odd 7 seven 4")); got != want {
					t.Errorf("a search at once gives %q, want %q", got, want)
					return
				}
			}
		})
	}
	wg.Wait()
}

// The magnitude of a key, not its sign, is searched for.
func TestKeyOf(t *testing.T) {
	for _, c := range []struct {
		check formatv0.Check
		want  string
	}{
		{formatv0.Check{Kind: formatv0.CheckKindNumber, Key: new("-2.5")}, "5/2"},
		{formatv0.Check{Kind: formatv0.CheckKindFraction, Key: new("-3/6")}, "1/2"},
		{formatv0.Check{Kind: formatv0.CheckKindFraction, Key: new("7")}, ""},
		{formatv0.Check{Kind: formatv0.CheckKindNumber, Key: new("07")}, ""},
		{formatv0.Check{Kind: formatv0.CheckKindRubric}, ""},
		{formatv0.Check{Kind: formatv0.CheckKindPredicate, Key: new("7")}, ""},
	} {
		n, ok := keyOf(c.check)
		got := ""
		if ok {
			got = n.magnitude.RatString()
		}
		if got != c.want {
			t.Errorf("the key %v of %s is searched as %q, want %q", c.check.Key, c.check.Kind, got, c.want)
		}
	}
	if n, ok := keyOf(formatv0.Check{Kind: formatv0.CheckKindChoice, Key: new(" ＜ ")}); !ok || n.form != "<" || n.magnitude != nil {
		t.Errorf("a key of a choice is searched as %+v", n)
	}
}
