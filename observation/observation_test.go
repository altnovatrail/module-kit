package observation

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	observationv1 "github.com/altnovatrail/spec/contracts/novatrail/observation/v1"
	"github.com/altnovatrail/spec/types/formatv0"
	"google.golang.org/protobuf/proto"

	"github.com/altnovatrail/module-kit/verdict"
)

// The tasks of the cases, by their items.

func text(s string) *formatv0.Text { t := formatv0.Text(s); return &t }

func texts(list ...string) []formatv0.Text {
	var t []formatv0.Text
	for _, s := range list {
		t = append(t, formatv0.Text(s))
	}
	return t
}

func numberCheck(key string, units ...string) formatv0.Check {
	return formatv0.Check{Kind: formatv0.CheckKindNumber, Key: &key, Units: texts(units...)}
}

func fractionCheck(key string) formatv0.Check {
	form := formatv0.CheckFormReduced
	return formatv0.Check{Kind: formatv0.CheckKindFraction, Key: &key, Form: &form}
}

func choiceCheck(key string, options ...string) formatv0.Check {
	return formatv0.Check{Kind: formatv0.CheckKindChoice, Key: &key, Options: texts(options...)}
}

func predicateCheck(expression string) formatv0.Check {
	return formatv0.Check{Kind: formatv0.CheckKindPredicate, Predicate: &expression}
}

func boxOf(c formatv0.Check) formatv0.LinePart {
	binding := formatv0.LinePartBindingBox
	return formatv0.LinePart{Binding: &binding, Check: &c}
}

func sheetSlot(binding formatv0.AnswerSlotBinding, c formatv0.Check) formatv0.AnswerSlot {
	s := formatv0.AnswerSlot{Binding: binding, Check: c}
	if binding == formatv0.AnswerSlotBindingZone {
		s.Zone = &formatv0.SheetRectangle{Left: 100, Top: 1000, Width: 400, Height: 200}
	}
	return s
}

func tappedSlot(key string, options ...string) formatv0.AnswerSlot {
	return formatv0.AnswerSlot{Binding: formatv0.AnswerSlotBindingChoice, Check: formatv0.Check{Kind: formatv0.CheckKindChoice, Key: &key}, Options: texts(options...)}
}

var (
	// A handwriting sheet whose one answer slot is bound anywhere.
	anywhere = []formatv0.Item{{AnswerSlots: []formatv0.AnswerSlot{sheetSlot(formatv0.AnswerSlotBindingAnywhere, numberCheck("7", "rungs"))}}}
	// Zones and answer slots bound choice: 1.1, 1.3 and 3.1 are read, 1.2 and 2.1 tapped.
	zones = []formatv0.Item{
		{Text: text("a)"), AnswerSlots: []formatv0.AnswerSlot{
			sheetSlot(formatv0.AnswerSlotBindingZone, numberCheck("15", "minutes")),
			func() formatv0.AnswerSlot {
				s := tappedSlot("after", "before", "after")
				s.Check.TrapValues = []formatv0.TrapValue{{Value: "before", Trap: "reversed-order"}}
				return s
			}(),
			sheetSlot(formatv0.AnswerSlotBindingZone, numberCheck("40")),
		}},
		{Text: text("b)"), AnswerSlots: []formatv0.AnswerSlot{tappedSlot("yes", "yes", "no")}},
		{Text: text("c)"), AnswerSlots: []formatv0.AnswerSlot{sheetSlot(formatv0.AnswerSlotBindingZone, numberCheck("55"))}},
	}
	// The boxes of a workbook: a number, a fraction in lowest terms and a sign, and a predicate.
	boxes = []formatv0.Item{
		{Line: []formatv0.LinePart{{Text: text("7 × 8 = ")}, boxOf(func() formatv0.Check {
			c := numberCheck("56")
			c.TrapValues = []formatv0.TrapValue{{Value: "54", Trap: "times-table"}}
			return c
		}())}},
		{Line: []formatv0.LinePart{boxOf(fractionCheck("3/4")), {Text: text(" and 1/2: ")}, boxOf(choiceCheck(">", "<", ">", "="))}},
		{Line: []formatv0.LinePart{{Text: text("A divisor of 12: ")}, boxOf(predicateCheck("12 % value == 0"))}},
	}
	// An essay: the one answer slot of its task is bound sheet, with a rubric.
	essay = []formatv0.Item{{AnswerSlots: []formatv0.AnswerSlot{sheetSlot(formatv0.AnswerSlotBindingSheet, formatv0.Check{Kind: formatv0.CheckKindRubric, Criteria: []formatv0.Criterion{{Text: "Clear."}}})}}}
	// Answer slots bound choice alone, which AI does not read.
	tapsOnly = []formatv0.Item{{AnswerSlots: []formatv0.AnswerSlot{tappedSlot("yes", "yes", "no")}}}
)

// The observations of the cases.

func observation(slots ...*observationv1.AnswerSlotObservation) *observationv1.Observation {
	return &observationv1.Observation{AnswerSlots: slots}
}

func at(item, slot int32) *observationv1.AnswerSlotPlace {
	return &observationv1.AnswerSlotPlace{Item: item, Slot: slot}
}

func value(item, slot int32, v string) *observationv1.AnswerSlotObservation {
	return &observationv1.AnswerSlotObservation{AnswerSlot: at(item, slot), Read: &observationv1.AnswerSlotObservation_Value{Value: v}}
}

func reason(item, slot int32, r observationv1.Unreadable) *observationv1.AnswerSlotObservation {
	return &observationv1.AnswerSlotObservation{AnswerSlot: at(item, slot), Read: &observationv1.AnswerSlotObservation_Unreadable{Unreadable: r}}
}

func boxed(s *observationv1.AnswerSlotObservation, ymin, xmin, ymax, xmax int32) *observationv1.AnswerSlotObservation {
	s.CropBox = &observationv1.CropBox{Ymin: ymin, Xmin: xmin, Ymax: ymax, Xmax: xmax}
	return s
}

const (
	blank     = observationv1.Unreadable_UNREADABLE_BLANK
	illegible = observationv1.Unreadable_UNREADABLE_ILLEGIBLE
	ambiguous = observationv1.Unreadable_UNREADABLE_AMBIGUOUS
	noFinal   = observationv1.Unreadable_UNREADABLE_NO_FINAL_ANSWER
	several   = observationv1.Unreadable_UNREADABLE_SEVERAL_FINAL_ANSWERS
)

// boxesRead gives an observation of the boxes, each a value.
func boxesRead(a, b, c, d string) *observationv1.Observation {
	return observation(value(1, 1, a), value(2, 1, b), value(2, 2, c), value(3, 1, d))
}

// zonesRead gives an observation of the zones, each a value.
func zonesRead(a, b, c string) *observationv1.Observation {
	return observation(value(1, 1, a), value(1, 3, b), value(3, 1, c))
}

// written writes an observation shortly: each answer slot by its place, with its value or its
// reason and its crop_box.
func written(o *observationv1.Observation) string {
	var parts []string
	for _, s := range o.GetAnswerSlots() {
		part := placeOf(s.GetAnswerSlot())
		if read, ok := s.GetRead().(*observationv1.AnswerSlotObservation_Value); ok {
			part += fmt.Sprintf(" %q", read.Value)
		} else {
			part += " " + strings.TrimPrefix(s.GetUnreadable().String(), "UNREADABLE_")
		}
		if b := s.GetCropBox(); b != nil {
			part += fmt.Sprintf(" [%d %d %d %d]", b.Ymin, b.Xmin, b.Ymax, b.Xmax)
		}
		parts = append(parts, part)
	}
	return strings.Join(parts, "; ")
}

// Each reason of the package is in the table, named as the value is, in lower case and without its
// prefix (observation.proto): a reason that a new tag of spec adds fails this test until the task
// that takes the tag names it.
func TestReasonsAreThePackages(t *testing.T) {
	values := observationv1.Unreadable(0).Descriptor().Values()
	for i := range values.Len() {
		v := values.Get(i)
		r := observationv1.Unreadable(v.Number())
		got, ok := reasons[r]
		if r == observationv1.Unreadable_UNREADABLE_UNSPECIFIED {
			if ok {
				t.Error("UNREADABLE_UNSPECIFIED is in the table, and no reason gives it")
			}
			continue
		}
		if want := verdict.Reason(strings.ToLower(strings.TrimPrefix(string(v.Name()), "UNREADABLE_"))); !ok || got != want {
			t.Errorf("%s gives %q in the table, want %q", v.Name(), got, want)
		}
	}
	if len(reasons) != values.Len()-1 {
		t.Errorf("the table names %d reasons, and the package has %d besides UNREADABLE_UNSPECIFIED", len(reasons), values.Len()-1)
	}
}

// Validate holds each condition of validity on its own.
func TestValidate(t *testing.T) {
	for _, c := range []struct {
		name  string
		items []formatv0.Item
		o     *observationv1.Observation
		why   string // what the error says; "" when the observation is valid
	}{
		{"every box a value", boxes, boxesRead("56", "3/4", ">", "6"), ""},
		{"reasons", boxes, observation(reason(1, 1, blank), reason(2, 1, illegible), reason(2, 2, ambiguous), value(3, 1, "6")), ""},
		{"U+FEFF is no white space", boxes, boxesRead("\uFEFF", "3/4", ">", "6"), ""},
		{"a box at 0 and 1000", anywhere, observation(boxed(value(1, 1, "7"), 0, 0, 1000, 1000)), ""},
		{"no final answer, and several", anywhere, observation(reason(1, 1, noFinal)), ""},
		{"zones", zones, zonesRead("15", "40", "55"), ""},
		{"an essay", essay, observation(value(1, 1, "An essay.")), ""},
		{"an essay that cannot be read", essay, observation(reason(1, 1, illegible)), ""},
		{"nothing to read", tapsOnly, observation(), ""},
		{"none", boxes, nil, "there is none"},
		{"a slot missing", boxes, observation(value(1, 1, "56"), value(2, 1, "3/4"), value(3, 1, "6")), "it holds the answer slots 1.1, 2.1, 3.1, and AI reads 1.1, 2.1, 2.2, 3.1"},
		{"a slot twice", boxes, observation(value(1, 1, "56"), value(2, 1, "3/4"), value(2, 1, "3/4"), value(2, 2, ">"), value(3, 1, "6")), "2.1, 2.1"},
		{"out of order", boxes, observation(value(1, 1, "56"), value(2, 2, ">"), value(2, 1, "3/4"), value(3, 1, "6")), "1.1, 2.2, 2.1, 3.1"},
		{"a place the task does not have", boxes, observation(value(1, 1, "56"), value(2, 1, "3/4"), value(2, 2, ">"), value(4, 1, "6")), "4.1"},
		{"no place", boxes, observation(value(1, 1, "56"), value(2, 1, "3/4"), value(2, 2, ">"), &observationv1.AnswerSlotObservation{Read: &observationv1.AnswerSlotObservation_Value{Value: "6"}}), "?.?"},
		{"a nil answer slot", boxes, observation(value(1, 1, "56"), value(2, 1, "3/4"), value(2, 2, ">"), nil), "?.?: it gives neither a value nor a reason"},
		{"a slot bound choice", zones, observation(value(1, 1, "15"), value(1, 2, "after"), value(1, 3, "40"), value(3, 1, "55")), "1.1, 1.2, 1.3, 3.1"},
		{"no observation of the task that AI does not read", tapsOnly, observation(value(1, 1, "yes")), "AI reads none"},
		{"neither a value nor a reason", boxes, observation(value(1, 1, "56"), value(2, 1, "3/4"), &observationv1.AnswerSlotObservation{AnswerSlot: at(2, 2)}, value(3, 1, "6")), "2.2: it gives neither a value nor a reason"},
		{"an empty value", boxes, boxesRead("", "3/4", ">", "6"), `1.1: its value "" is empty or only white space`},
		{"U+00A0 is white space", boxes, boxesRead("\u00A0", "3/4", ">", "6"), "empty or only white space"},
		{"U+0085 is white space", boxes, boxesRead("56", "\u0085", ">", "6"), "2.1: its value"},
		{"not UTF-8", boxes, boxesRead("5\xff", "3/4", ">", "6"), `1.1: its value "5\xff" is not UTF-8`},
		{"UNREADABLE_UNSPECIFIED", boxes, observation(value(1, 1, "56"), reason(2, 1, 0), value(2, 2, ">"), value(3, 1, "6")), "2.1: its reason is UNREADABLE_UNSPECIFIED"},
		{"a reason that this release does not know", boxes, observation(reason(1, 1, 1000), value(2, 1, "3/4"), value(2, 2, ">"), value(3, 1, "6")), "1.1: its reason 1000 is none that this release knows"},
		{"no final answer in a box", boxes, observation(value(1, 1, "56"), value(2, 1, "3/4"), reason(2, 2, noFinal), value(3, 1, "6")), "2.2: UNREADABLE_NO_FINAL_ANSWER only in a slot bound anywhere, and this one is bound box"},
		{"several in a zone", zones, observation(value(1, 1, "15"), reason(1, 3, several), value(3, 1, "55")), "1.3: UNREADABLE_SEVERAL_FINAL_ANSWERS only in a slot bound anywhere, and this one is bound zone"},
		{"a crop_box in a box", boxes, observation(boxed(value(1, 1, "56"), 1, 1, 2, 2), value(2, 1, "3/4"), value(2, 2, ">"), value(3, 1, "6")), "1.1: a crop_box only in a slot bound anywhere, and this one is bound box"},
		{"a crop_box beside a reason", anywhere, observation(boxed(reason(1, 1, noFinal), 1, 1, 2, 2)), "1.1: a crop_box stands only beside a value"},
		{"a box whose top is its bottom", anywhere, observation(boxed(value(1, 1, "7"), 500, 100, 500, 900)), "the crop_box [500, 100, 500, 900] is not"},
		{"a box whose left is right of its right", anywhere, observation(boxed(value(1, 1, "7"), 100, 900, 500, 100)), "is not"},
		{"a box past 1000", anywhere, observation(boxed(value(1, 1, "7"), 100, 100, 1001, 900)), "is not"},
		{"a box below 0", anywhere, observation(boxed(value(1, 1, "7"), -1, 100, 500, 900)), "is not"},
		{"a box of zeros", anywhere, observation(boxed(value(1, 1, "7"), 0, 0, 0, 0)), "is not"},
		{"ambiguous in an essay", essay, observation(reason(1, 1, ambiguous)), "1.1: a slot bound sheet gives a value, UNREADABLE_BLANK or UNREADABLE_ILLEGIBLE, not UNREADABLE_AMBIGUOUS"},
		{"no final answer in an essay", essay, observation(reason(1, 1, noFinal)), "only in a slot bound anywhere, and this one is bound sheet"},
		{"a crop_box in an essay", essay, observation(boxed(value(1, 1, "An essay."), 1, 1, 2, 2)), "1.1: a crop_box only in a slot bound anywhere, and this one is bound sheet"},
		{"every condition named", boxes, observation(value(1, 1, ""), reason(2, 1, 0)), "1.1: its value \"\" is empty or only white space; 2.1: its reason is UNREADABLE_UNSPECIFIED"},
	} {
		err := Validate(c.items, c.o)
		switch {
		case c.why == "" && err != nil:
			t.Errorf("%s: %v", c.name, err)
		case c.why != "" && (!errors.Is(err, ErrNotValid) || !strings.Contains(err.Error(), c.why)):
			t.Errorf("%s: %v, want ErrNotValid saying %q", c.name, err, c.why)
		}
	}
}

// Join joins two readings answer slot by answer slot, by the rules of contracts/README.md, section
// 10.
func TestJoin(t *testing.T) {
	for _, c := range []struct {
		name          string
		items         []formatv0.Item
		first, second *observationv1.Observation
		want          string
	}{
		{"alike", boxes, boxesRead("56", "3/4", ">", "6"), boxesRead("56", "3/4", ">", "6"), `1.1 "56"; 2.1 "3/4"; 2.2 ">"; 3.1 "6"`},
		{"written otherwise, the first's", boxes, boxesRead("56", "3 / 4", "＞", "6"), boxesRead("= 56.", "3/4", ">", "6"), `1.1 "56"; 2.1 "3 / 4"; 2.2 "＞"; 3.1 "6"`},
		{"a trap read alike", boxes, boxesRead("54", "3/4", ">", "6"), boxesRead("= 54", "3/4", ">", "6"), `1.1 "54"; 2.1 "3/4"; 2.2 ">"; 3.1 "6"`},
		{"different values", boxes, boxesRead("56", "3/4", ">", "6"), boxesRead("54", "1/2", "<", "4"), "1.1 AMBIGUOUS; 2.1 AMBIGUOUS; 2.2 AMBIGUOUS; 3.1 AMBIGUOUS"},
		{"one value, two verdicts: 6/8 and 3/4 under form reduced", boxes, boxesRead("56", "6/8", ">", "6"), boxesRead("56", "3/4", ">", "6"), `1.1 "56"; 2.1 AMBIGUOUS; 2.2 ">"; 3.1 "6"`},
		{"one value and one verdict: 2/4 and 1/2, both not yet right", boxes, boxesRead("56", "2/4", ">", "6"), boxesRead("56", "1 / 2", ">", "6"), `1.1 "56"; 2.1 "2/4"; 2.2 ">"; 3.1 "6"`},
		{"a predicate compares the integers read", boxes, boxesRead("56", "3/4", ">", "-0"), boxesRead("56", "3/4", ">", "0"), `1.1 "56"; 2.1 "3/4"; 2.2 ">"; 3.1 "-0"`},
		{"two integers that the predicate holds for", boxes, boxesRead("56", "3/4", ">", "6"), boxesRead("56", "3/4", ">", "4"), `1.1 "56"; 2.1 "3/4"; 2.2 ">"; 3.1 AMBIGUOUS`},
		{"out of notation, the same once normalized", boxes, boxesRead("7²", "0.75", "≥", "012"), boxesRead("7³", "0.75", "≥", "012"), `1.1 "7²"; 2.1 "0.75"; 2.2 "≥"; 3.1 "012"`},
		{"in notation and out of it", boxes, boxesRead("56", "3/4", ">", "12"), boxesRead("56²", "0.75", "≥", "012"), "1.1 AMBIGUOUS; 2.1 AMBIGUOUS; 2.2 AMBIGUOUS; 3.1 AMBIGUOUS"},
		{"blank before the join", boxes, boxesRead("=", "=", "56", "!"), observation(reason(1, 1, blank), value(2, 1, "."), value(2, 2, ">"), value(3, 1, "6")), `1.1 BLANK; 2.1 BLANK; 2.2 AMBIGUOUS; 3.1 AMBIGUOUS`},
		{"= is a sign, not blank", boxes, boxesRead("56", "3/4", "=", "6"), boxesRead("56", "3/4", " = ", "6"), `1.1 "56"; 2.1 "3/4"; 2.2 "="; 3.1 "6"`},
		{"reasons", boxes, observation(reason(1, 1, illegible), reason(2, 1, illegible), value(2, 2, ">"), reason(3, 1, blank)), observation(reason(1, 1, illegible), reason(2, 1, blank), reason(2, 2, blank), value(3, 1, "6")), "1.1 ILLEGIBLE; 2.1 AMBIGUOUS; 2.2 AMBIGUOUS; 3.1 AMBIGUOUS"},
		{"the first's box", anywhere, observation(boxed(value(1, 1, "7"), 1, 2, 3, 4)), observation(boxed(value(1, 1, "7 rungs"), 5, 6, 7, 8)), `1.1 "7" [1 2 3 4]`},
		{"the second's box", anywhere, observation(value(1, 1, "7")), observation(boxed(value(1, 1, "7 rungs"), 5, 6, 7, 8)), `1.1 "7" [5 6 7 8]`},
		{"no box", anywhere, observation(value(1, 1, "7")), observation(value(1, 1, "7")), `1.1 "7"`},
		{"a word that is no unit", anywhere, observation(boxed(value(1, 1, "7 steps"), 1, 2, 3, 4)), observation(value(1, 1, "7  steps.")), `1.1 "7 steps" [1 2 3 4]`},
		{"ambiguous has no box", anywhere, observation(boxed(value(1, 1, "7"), 1, 2, 3, 4)), observation(boxed(value(1, 1, "8"), 1, 2, 3, 4)), "1.1 AMBIGUOUS"},
		{"no final answer and several", anywhere, observation(reason(1, 1, noFinal)), observation(reason(1, 1, several)), "1.1 AMBIGUOUS"},
		{"several and several", anywhere, observation(reason(1, 1, several)), observation(reason(1, 1, several)), "1.1 SEVERAL_FINAL_ANSWERS"},
		{"zones, and nothing of the slots tapped", zones, zonesRead("15", "40", "55"), zonesRead("15 minutes", "41", "55"), `1.1 "15"; 1.3 AMBIGUOUS; 3.1 "55"`},
		{"nothing to read", tapsOnly, observation(), observation(), ""},
	} {
		got, err := Join(c.items, c.first, c.second)
		if err != nil {
			t.Errorf("%s: %v", c.name, err)
			continue
		}
		if w := written(got); w != c.want {
			t.Errorf("%s: joins into %s, want %s", c.name, w, c.want)
		}
		if err := Validate(c.items, got); err != nil {
			t.Errorf("%s: the joined observation: %v", c.name, err)
		}
	}
}

// The joined observation is a new message: changing either observation after the join changes
// nothing of it.
func TestJoinSharesNothing(t *testing.T) {
	first := observation(boxed(value(1, 1, "7"), 1, 2, 3, 4))
	joined, err := Join(anywhere, first, observation(value(1, 1, "7")))
	if err != nil {
		t.Fatal(err)
	}
	first.AnswerSlots[0].CropBox.Ymin = 0
	first.AnswerSlots[0].AnswerSlot.Item = 9
	if w := written(joined); w != `1.1 "7" [1 2 3 4]` {
		t.Errorf("the joined observation is %s once the first changed", w)
	}
}

// Join gives no joined observation when either is not valid, of an essay, or when the verdict
// refuses a check.
func TestJoinErrors(t *testing.T) {
	placeholder := []formatv0.Item{{Line: []formatv0.LinePart{boxOf(numberCheck("{sum}"))}}}
	for _, c := range []struct {
		name          string
		items         []formatv0.Item
		first, second *observationv1.Observation
		want          error
		says          string
	}{
		{"the first not valid", boxes, boxesRead("", "3/4", ">", "6"), boxesRead("56", "3/4", ">", "6"), ErrNotValid, "the first observation is not valid"},
		{"the second not valid", boxes, boxesRead("56", "3/4", ">", "6"), nil, ErrNotValid, "the second observation is not valid: there is none"},
		{"an essay", essay, observation(value(1, 1, "An essay.")), observation(value(1, 1, "An essay.")), nil, "an essay has one reader"},
		{"a placeholder", placeholder, observation(value(1, 1, "5")), observation(value(1, 1, "5")), verdict.ErrPlaceholder, "answer slot 1.1"},
		{"a placeholder, read out of notation", placeholder, observation(value(1, 1, "x")), observation(value(1, 1, "x")), verdict.ErrPlaceholder, "answer slot 1.1"},
		{"a placeholder, read as reasons", placeholder, observation(reason(1, 1, blank)), observation(reason(1, 1, blank)), verdict.ErrPlaceholder, "answer slot 1.1"},
	} {
		got, err := Join(c.items, c.first, c.second)
		if got != nil || err == nil || c.want != nil && !errors.Is(err, c.want) || !strings.Contains(err.Error(), c.says) {
			t.Errorf("%s: %s, %v, want %v saying %q", c.name, written(got), err, c.want, c.says)
		}
	}
}

// short writes the verdicts of the items as a vector writes them.
func short(items []verdict.ItemVerdict) string {
	var parts []string
	for _, item := range items {
		var slots []string
		for _, s := range item.Slots {
			text := string(s.Verdict)
			if s.Reason != "" {
				text += " " + string(s.Reason)
			}
			if s.Trap != nil {
				text += " " + string(s.Trap.Trap)
			}
			slots = append(slots, text)
		}
		parts = append(parts, fmt.Sprintf("%s (%s)", item.Verdict, strings.Join(slots, ", ")))
	}
	return strings.Join(parts, "; ")
}

func tap(item, slot int32, v string) Tapped { return Tapped{Place: at(item, slot), Value: v} }

// Verdicts gives each item's verdict on the observation and the options tapped.
func TestVerdicts(t *testing.T) {
	for _, c := range []struct {
		name   string
		items  []formatv0.Item
		o      *observationv1.Observation
		tapped []Tapped
		want   string
	}{
		{"right", zones, zonesRead("15 minutes", "40", "55"), []Tapped{tap(1, 2, "after"), tap(2, 1, "yes")}, "correct (correct, correct, correct); correct (correct); correct (correct)"},
		{"a trap tapped, and none", zones, zonesRead("15", "40", "55"), []Tapped{tap(1, 2, "before")}, "not_yet (correct, not_yet reversed-order, correct); unreadable (unreadable blank); correct (correct)"},
		{"an option as none of them", zones, zonesRead("15", "40", "55"), []Tapped{tap(2, 1, "maybe")}, "unreadable (correct, unreadable blank, correct); unreadable (unreadable notation); correct (correct)"},
		{"no observation", zones, nil, []Tapped{tap(1, 2, "before"), tap(2, 1, "yes")}, "not_yet (unchecked, not_yet reversed-order, unchecked); correct (correct); unchecked (unchecked)"},
		{"reasons", boxes, observation(reason(1, 1, blank), reason(2, 1, illegible), reason(2, 2, ambiguous), value(3, 1, "=")), nil, "unreadable (unreadable blank); unreadable (unreadable illegible, unreadable ambiguous); unreadable (unreadable blank)"},
		{"reasons of a final answer", anywhere, observation(reason(1, 1, several)), nil, "unreadable (unreadable several_final_answers)"},
		{"no final answer", anywhere, observation(reason(1, 1, noFinal)), nil, "unreadable (unreadable no_final_answer)"},
		{"values", boxes, boxesRead("54", "6/8", "<", "5"), nil, "not_yet (not_yet times-table); not_yet (not_yet, not_yet); not_yet (not_yet)"},
		{"the predicate's own unchecked", boxes, boxesRead("56", "3/4", ">", "0"), nil, "correct (correct); correct (correct, correct); unchecked (unchecked)"},
		{"out of notation", boxes, boxesRead("56²", "0.75", "≥", "012"), nil, "unreadable (unreadable notation); unreadable (unreadable notation, unreadable notation); unreadable (unreadable notation)"},
		{"nothing to read", tapsOnly, observation(), []Tapped{tap(1, 1, "no")}, "not_yet (not_yet)"},
	} {
		got, err := Verdicts(context.Background(), c.items, c.o, c.tapped)
		if err != nil {
			t.Errorf("%s: %v", c.name, err)
			continue
		}
		if s := short(got); s != c.want {
			t.Errorf("%s: %s, want %s", c.name, s, c.want)
		}
	}
}

// Verdicts gives no verdict on an observation that is not valid, an essay, options tapped out of
// place, an item with no answer slots, a check the verdict refuses and a context done.
func TestVerdictsErrors(t *testing.T) {
	done, cancel := context.WithCancel(context.Background())
	cancel()
	placeholder := []formatv0.Item{{Line: []formatv0.LinePart{boxOf(numberCheck("{sum}"))}}}
	for _, c := range []struct {
		name   string
		ctx    context.Context
		items  []formatv0.Item
		o      *observationv1.Observation
		tapped []Tapped
		want   error
		says   string
	}{
		{"not valid", context.Background(), boxes, boxesRead("", "3/4", ">", "6"), nil, ErrNotValid, "the observation is not valid"},
		{"an essay", context.Background(), essay, observation(value(1, 1, "An essay.")), nil, verdict.ErrNoVerdict, "answer slot 1.1"},
		{"tapped where AI reads", context.Background(), zones, zonesRead("15", "40", "55"), []Tapped{tap(1, 1, "15")}, nil, "option tapped 1: the answer slot 1.1"},
		{"tapped out of order", context.Background(), zones, zonesRead("15", "40", "55"), []Tapped{tap(2, 1, "yes"), tap(1, 2, "after")}, nil, "option tapped 2: the answer slot 1.2"},
		{"tapped twice", context.Background(), zones, zonesRead("15", "40", "55"), []Tapped{tap(1, 2, "after"), tap(1, 2, "after")}, nil, "option tapped 2"},
		{"tapped with no place", context.Background(), zones, zonesRead("15", "40", "55"), []Tapped{{Value: "after"}}, nil, "the answer slot ?.?"},
		{"tapped in a place the task does not have", context.Background(), zones, zonesRead("15", "40", "55"), []Tapped{tap(5, 1, "yes")}, nil, "5.1"},
		{"an item with no answer slots", context.Background(), append([]formatv0.Item{{Text: text("?")}}, anywhere...), nil, nil, nil, "item 1 has no answer slots"},
		{"a placeholder", context.Background(), placeholder, observation(value(1, 1, "5")), nil, verdict.ErrPlaceholder, "answer slot 1.1"},
		{"a placeholder, and no observation", context.Background(), placeholder, nil, nil, verdict.ErrPlaceholder, "answer slot 1.1"},
		{"a predicate with no expression", context.Background(), []formatv0.Item{{Line: []formatv0.LinePart{boxOf(formatv0.Check{Kind: formatv0.CheckKindPredicate})}}}, observation(reason(1, 1, illegible)), nil, nil, "the predicate check has no predicate"},
		{"a tapped value not UTF-8", context.Background(), zones, zonesRead("15", "40", "55"), []Tapped{tap(2, 1, "y\xff")}, nil, "answer slot 2.1"},
		{"a context done", done, boxes, boxesRead("56", "3/4", ">", "6"), nil, context.Canceled, "observation: context canceled"},
	} {
		got, err := Verdicts(c.ctx, c.items, c.o, c.tapped)
		if got != nil || err == nil || c.want != nil && !errors.Is(err, c.want) || !strings.Contains(err.Error(), c.says) {
			t.Errorf("%s: %s, %v, want %v saying %q", c.name, short(got), err, c.want, c.says)
		}
	}
}

// Essay says how the submission of an essay ends.
func TestEssay(t *testing.T) {
	for _, c := range []struct {
		name string
		o    *observationv1.Observation
		want Submission
	}{
		{"read", observation(value(1, 1, "An essay.")), SubmissionAssessment},
		{"a rubric normalizes nothing", observation(value(1, 1, "=")), SubmissionAssessment},
		{"blank", observation(reason(1, 1, blank)), SubmissionUnreadable},
		{"illegible", observation(reason(1, 1, illegible)), SubmissionUnreadable},
		{"no observation", nil, SubmissionUnchecked},
	} {
		if got, err := Essay(essay, c.o); err != nil || got != c.want {
			t.Errorf("%s: %q, %v, want %q", c.name, got, err, c.want)
		}
	}
	if got, err := Essay(essay, observation(reason(1, 1, ambiguous))); got != "" || !errors.Is(err, ErrNotValid) {
		t.Errorf("an essay read as ambiguous: %q, %v, want ErrNotValid", got, err)
	}
	for name, items := range map[string][]formatv0.Item{"boxes": boxes, "anywhere": anywhere, "an essay beside a slot bound choice": append(append([]formatv0.Item{}, essay...), tapsOnly...)} {
		if got, err := Essay(items, observation()); got != "" || err == nil || !strings.Contains(err.Error(), "no essay") {
			t.Errorf("%s: %q, %v, want an error that it is no essay", name, got, err)
		}
	}
}

// A value that is not UTF-8 is not valid since no message of the package holds one: protobuf-go
// refuses to write it, as proto3 has a string hold UTF-8 alone (R29).
func TestNotUTF8IsNoMessage(t *testing.T) {
	if _, err := proto.Marshal(boxesRead("5\xff", "3/4", ">", "6")); err == nil || !strings.Contains(err.Error(), "invalid UTF-8") {
		t.Errorf("proto.Marshal writes a value that is not UTF-8: %v", err)
	}
}
