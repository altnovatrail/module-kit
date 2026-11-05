package verdict

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"testing"
	"time"
	"unicode"

	"github.com/altnovatrail/spec/types/formatv0"
	"golang.org/x/text/unicode/norm"

	"github.com/altnovatrail/module-kit/internal/sandbox"
)

// The answer slots of the cases: a check written by hand in a box, or a choice tapped.
func box(check formatv0.Check) AnswerSlot {
	return AnswerSlot{Binding: string(formatv0.LinePartBindingBox), Check: check}
}

func number(key string, units ...string) AnswerSlot {
	return box(formatv0.Check{Kind: formatv0.CheckKindNumber, Key: &key, Units: texts(units)})
}

func fraction(key string, form formatv0.CheckForm) AnswerSlot {
	return box(formatv0.Check{Kind: formatv0.CheckKindFraction, Key: &key, Form: &form})
}

func byHand(key string, options ...string) AnswerSlot {
	return box(formatv0.Check{Kind: formatv0.CheckKindChoice, Key: &key, Options: texts(options)})
}

func tapped(key string, options ...string) AnswerSlot {
	return AnswerSlot{Binding: string(formatv0.AnswerSlotBindingChoice), Check: formatv0.Check{Kind: formatv0.CheckKindChoice, Key: &key}, Options: texts(options)}
}

func predicate(expression string) AnswerSlot {
	return box(formatv0.Check{Kind: formatv0.CheckKindPredicate, Predicate: &expression})
}

// The predicates of the cases: one of the format's text, and one whose count of steps reaches the
// limit of a predicate on any value, that of the vector steps-reach-the-limit.yaml.
const (
	divisibleBy3 = "value % 3 == 0"
	tooManySteps = "value in [n * n for n in range(99998)]"
)

func rubric() AnswerSlot {
	return AnswerSlot{Binding: string(formatv0.AnswerSlotBindingSheet), Check: formatv0.Check{Kind: formatv0.CheckKindRubric, Criteria: []formatv0.Criterion{{Text: "Clear."}}}}
}

// trapped gives an answer slot with trap values, each a value and its trap.
func trapped(slot AnswerSlot, pairs ...string) AnswerSlot {
	for i := 0; i < len(pairs); i += 2 {
		slot.Check.TrapValues = append(slot.Check.TrapValues, formatv0.TrapValue{Value: pairs[i], Trap: formatv0.TrapID(pairs[i+1])})
	}
	return slot
}

func texts(list []string) []formatv0.Text {
	var t []formatv0.Text
	for _, s := range list {
		t = append(t, formatv0.Text(s))
	}
	return t
}

func ptr(s string) *string { return &s }

// What a case wants of an answer slot: the verdict, the reason and the trap by its ID, as a vector
// writes them.
func short(v SlotVerdict) string {
	text := string(v.Verdict)
	if v.Reason != "" {
		text += " " + string(v.Reason)
	}
	if v.Trap != nil {
		text += " " + string(v.Trap.Trap)
	}
	return text
}

// The cases hold each step of the normalization, the notation of each kind and the choices, by the
// examples of the format's text where it gives them (format/README.md, section 15).
var slotCases = []struct {
	name  string
	slot  AnswerSlot
	value *string
	want  string
}{
	// A number, and steps 1 to 5 and 7.
	{"a whole number", number("7"), ptr("7"), "correct"},
	{"nothing read", number("7"), nil, "unreadable blank"},
	{"nothing written", number("7"), ptr(""), "unreadable blank"},
	{"white space only", number("7"), ptr(" \t\u00A0"), "unreadable blank"},
	{"an equals sign and a full stop only", number("7"), ptr("= ."), "unreadable blank"},
	{"an equals sign at the start", number("7"), ptr("= 7"), "correct"},
	{"two equals signs", number("7"), ptr("==7"), "unreadable notation"},
	{"full stops and marks at the end", number("7"), ptr("7.?!"), "correct"},
	{"a full-width digit", number("7"), ptr("７"), "correct"},
	{"white space around, of other kinds", number("7"), ptr("\u2003 7\u0085"), "correct"},
	{"another number", number("7"), ptr("8"), "not_yet"},
	{"a superscript digit after", number("72"), ptr("7²"), "unreadable notation"},
	{"a superscript minus", number("-3"), ptr("⁻3"), "unreadable notation"},
	{"a superscript letter", number("7"), ptr("7ⁿ"), "unreadable notation"},
	{"the minus sign", number("-3"), ptr("\u22123"), "correct"},
	{"a hyphen", number("-3"), ptr("\u20103"), "correct"},
	{"a non-breaking hyphen, which NFKC makes a hyphen", number("-3"), ptr("\u20113"), "correct"},
	{"an em dash", number("-3"), ptr("\u20143"), "correct"},
	{"a hyphen-minus that NFKC gives from a full-width one", number("-3"), ptr("\uFF0D3"), "correct"},
	{"a negative zero", number("0"), ptr("-0"), "correct"},
	// The notation of a number.
	{"a leading zero", number("7"), ptr("07"), "unreadable notation"},
	{"a plus sign", number("7"), ptr("+7"), "unreadable notation"},
	{"a decimal without its whole part", number("0.5"), ptr(".5"), "unreadable notation"},
	{"a decimal comma", number("1.5"), ptr("1,5"), "unreadable notation"},
	{"digits grouped by a space", number("1000"), ptr("1 000"), "unreadable notation"},
	{"digits grouped by a comma", number("1000"), ptr("1,000"), "unreadable notation"},
	{"a zero denominator", number("1"), ptr("1/0"), "unreadable notation"},
	{"a mixed number", number("3/2"), ptr("1 1/2"), "unreadable notation"},
	{"a vulgar fraction after a digit", number("3/2"), ptr("1½"), "unreadable notation"},
	{"two vulgar fractions", number("1"), ptr("½½"), "unreadable notation"},
	{"a vulgar fraction alone", number("1/2"), ptr("½"), "correct"},
	{"a vulgar fraction after a letter", number("1/2"), ptr("x½"), "unreadable notation"},
	{"an exponent", number("7"), ptr("7e0"), "unreadable notation"},
	{"digits of another script", number("7"), ptr("٧"), "unreadable notation"},
	{"a decimal, a fraction and a halved fraction of one value", number("1/2"), ptr("0.5"), "correct"},
	{"a fraction of the key's value", number("1/2"), ptr("3/6"), "correct"},
	{"a fraction with spaces", number("1/2"), ptr("1 / 2"), "correct"},
	{"the fraction slash", number("1/2"), ptr("1\u20442"), "correct"},
	{"the division slash", number("1/2"), ptr("1\u22152"), "correct"},
	{"an integer beyond a double's", number("9007199254740992"), ptr("9007199254740993"), "not_yet"},
	{"16 characters", number("1234567890.12345"), ptr("1234567890.12345"), "correct"},
	{"17 characters", number("1234567890.12345"), ptr("1234567890.123450"), "unreadable notation"},
	// Units, step 6.
	{"a unit after a space", number("7", "rungs", "rung"), ptr("7 rungs"), "correct"},
	{"a unit with no space", number("7", "rungs", "rung"), ptr("7rungs"), "correct"},
	{"a unit in capitals", number("7", "rungs", "rung"), ptr("7 RUNGS"), "correct"},
	{"the singular", number("7", "rungs", "rung"), ptr("7 rung"), "correct"},
	{"a unit alone", number("7", "rungs"), ptr("rungs"), "unreadable notation"},
	{"a unit after an equals sign alone", number("7", "rungs"), ptr("= rungs"), "unreadable notation"},
	{"a unit before a spaced full stop", number("7", "rungs"), ptr("7 rungs ."), "unreadable notation"},
	{"a unit that ends another", number("12", "m", "cm"), ptr("12cm"), "correct"},
	{"a unit that only the shorter one ends", number("12", "m"), ptr("12 cm"), "unreadable notation"},
	{"a unit with a superscript digit", number("12", "cm²"), ptr("12 cm²"), "correct"},
	{"a plain digit for the superscript one", number("12", "cm²"), ptr("12 cm2"), "unreadable notation"},
	{"a unit that ends in a full stop", number("40", "min."), ptr("40 min"), "correct"},
	{"a unit whose lowercase is fewer bytes", number("7", "İn"), ptr("7İN"), "correct"},
	{"a unit that normalizes to nothing", number("7", "."), ptr("7"), "correct"},
	// A fraction.
	{"a fraction in lowest terms", fraction("3/4", formatv0.CheckFormReduced), ptr("3/4"), "correct"},
	{"a vulgar fraction of it", fraction("3/4", formatv0.CheckFormReduced), ptr("¾"), "correct"},
	{"a fraction that can be reduced", fraction("3/4", formatv0.CheckFormReduced), ptr("6/8"), "not_yet"},
	{"a negative one that can be reduced", fraction("-3/4", formatv0.CheckFormReduced), ptr("-6/8"), "not_yet"},
	{"a negative one", fraction("-3/4", formatv0.CheckFormReduced), ptr("-3/4"), "correct"},
	{"any form", fraction("3/4", formatv0.CheckFormAny), ptr("6/8"), "correct"},
	{"a decimal for a fraction", fraction("3/4", formatv0.CheckFormAny), ptr("0.75"), "unreadable notation"},
	{"an integer for a fraction", fraction("4/4", formatv0.CheckFormAny), ptr("1"), "unreadable notation"},
	// Trap values.
	{"a trap value of the key's value", trapped(fraction("3/4", formatv0.CheckFormReduced), "6/8", "not-in-lowest-terms"), ptr("9/12"), "not_yet not-in-lowest-terms"},
	{"a value that matches no trap value", trapped(fraction("3/4", formatv0.CheckFormReduced), "6/8", "not-in-lowest-terms"), ptr("1/2"), "not_yet"},
	{"a trap value with the units", trapped(number("7", "rungs"), "6", "off-by-one"), ptr("6 rungs"), "not_yet off-by-one"},
	{"the right value matches no trap", trapped(number("7", "rungs"), "6", "off-by-one"), ptr("7"), "correct"},
	{"a trap value of a choice", trapped(byHand("odd", "even", "odd"), "even", "parity"), ptr("EVEN"), "not_yet parity"},
	// A choice.
	{"a sign", byHand("<", "<", ">", "="), ptr("<"), "correct"},
	{"a full-width sign", byHand("<", "<", ">", "="), ptr("＜"), "correct"},
	{"another sign", byHand("<", "<", ">", "="), ptr(">"), "not_yet"},
	{"an equals sign stays", byHand("<", "<", ">", "="), ptr("="), "not_yet"},
	{"none of the options", byHand("<", "<", ">", "="), ptr("7"), "unreadable notation"},
	{"a sign that no option is", byHand("<", "<", ">", "="), ptr("≤"), "unreadable notation"},
	{"white space alone", byHand("<", "<", ">", "="), ptr(" "), "unreadable blank"},
	{"nothing read in a choice", byHand("<", "<", ">", "="), nil, "unreadable blank"},
	{"a word in capitals", byHand("odd", "even", "odd"), ptr("ODD"), "correct"},
	{"a full stop stays", byHand("odd", "even", "odd"), ptr("Odd."), "unreadable notation"},
	{"a capital I with a dot", byHand("i", "i", "ii"), ptr("İ"), "correct"},
	{"white space of another kind", byHand("a b", "a b", "c"), ptr("a\u0085b"), "correct"},
	{"a zero-width no-break space is no white space", byHand("a b", "a b", "c"), ptr("a\uFEFFb"), "unreadable notation"},
	{"a choice tapped", tapped("Yes, it was.", "No.", "Yes, it was."), ptr("Yes, it was."), "correct"},
	{"another option tapped", tapped("Yes, it was.", "No.", "Yes, it was."), ptr("No."), "not_yet"},
	{"a key that no option is still gets its verdict", byHand("≤", "<", ">", "="), ptr("≤"), "correct"},
	// A predicate.
	{"a predicate that holds", predicate(divisibleBy3), ptr("12"), "correct"},
	{"a predicate that does not hold", predicate(divisibleBy3), ptr("13"), "not_yet"},
	{"a predicate on a value normalized", predicate(divisibleBy3), ptr("= ９."), "correct"},
	{"a predicate on a negative integer, by floored division", predicate("value % 12 == 0 and value < 0"), ptr("-24"), "correct"},
	{"a predicate on an integer beyond a double's", predicate("value == 9007199254740993"), ptr("9007199254740993"), "correct"},
	{"a predicate that holds for no integer of the double near it", predicate("value == 9007199254740993"), ptr("9007199254740992"), "not_yet"},
	{"a predicate with a comprehension and a set", predicate("value in set([n * n for n in range(10)])"), ptr("49"), "correct"},
	{"nothing read in a predicate", predicate(divisibleBy3), nil, "unreadable blank"},
	{"a predicate's value blank once normalized", predicate(divisibleBy3), ptr("="), "unreadable blank"},
	{"a decimal for a predicate", predicate(divisibleBy3), ptr("1.5"), "unreadable notation"},
	{"a decimal of an integer's value for a predicate", predicate(divisibleBy3), ptr("12.0"), "unreadable notation"},
	{"a fraction for a predicate", predicate(divisibleBy3), ptr("1/2"), "unreadable notation"},
	{"a predicate that fails on the value", predicate("12 % value == 0"), ptr("0"), "unchecked"},
	{"a predicate that gives no bool", predicate("value + 1"), ptr("3"), "unchecked"},
	{"a predicate that gives a truthy value", predicate("[value]"), ptr("3"), "unchecked"},
	{"a predicate that is no expression", predicate("value = 12"), ptr("12"), "unchecked"},
	{"a predicate that names what it may not", predicate("number % 12 == 0"), ptr("12"), "unchecked"},
	{"a predicate that calls a function recursively", predicate("(lambda f: f(f))(lambda f: f(f))"), ptr("3"), "unchecked"},
	{"a predicate that fails on purpose", predicate("fail('no')"), ptr("3"), "unchecked"},
	{"a predicate that prints", predicate("print(value) == None"), ptr("3"), "correct"},
}

func TestOfSlot(t *testing.T) {
	for _, c := range slotCases {
		t.Run(c.name, func(t *testing.T) {
			got, err := OfSlot(context.Background(), c.slot, c.value)
			if err != nil {
				t.Fatal(err)
			}
			if short(got) != c.want {
				t.Errorf("OfSlot gives %q, want %q", short(got), c.want)
			}
		})
	}
}

// The trap of a wrong value is a copy of the trap value, with its text.
func TestTrapIsTheTrapValue(t *testing.T) {
	slot := trapped(number("7"), "6", "off-by-one")
	slot.Check.TrapValues[0].Text = new(formatv0.Text("Count the posts, not the gaps."))
	got, err := OfSlot(context.Background(), slot, ptr("6"))
	if err != nil {
		t.Fatal(err)
	}
	if got.Trap == nil || got.Trap.Text == nil || *got.Trap.Text != "Count the posts, not the gaps." {
		t.Fatalf("the trap is %+v", got.Trap)
	}
	got.Trap.Value = "5"
	if slot.Check.TrapValues[0].Value != "6" {
		t.Error("the trap is the check's own trap value, not a copy")
	}
}

// An error gives no verdict: the check is one that the format does not take, the value is not
// UTF-8, ctx is done, or the check is a rubric. A predicate stopped by ctx is TestPredicateStops'.
func TestOfSlotErrors(t *testing.T) {
	cancelled, cancel := context.WithCancel(context.Background())
	cancel()
	noKey := number("7")
	noKey.Check.Key = nil
	noChoiceKey := byHand("a", "a", "b")
	noChoiceKey.Check.Key = nil
	noExpression := predicate(divisibleBy3)
	noExpression.Check.Predicate = nil
	for _, c := range []struct {
		name  string
		ctx   context.Context
		slot  AnswerSlot
		value *string
		want  string
		is    error
	}{
		{"a kind that the format does not have", context.Background(), box(formatv0.Check{Kind: "text"}), ptr("7"), `verdict: the check kind "text" is none of the format's`, nil},
		{"a number with no key", context.Background(), noKey, ptr("7"), "verdict: the number check has no key", nil},
		{"a key out of the notation", context.Background(), number("seven"), ptr("7"), `verdict: the key "seven" is not in the notation of number`, nil},
		{"an integer key of a fraction", context.Background(), fraction("3", formatv0.CheckFormAny), ptr("3/1"), `verdict: the key "3" is not in the notation of fraction`, nil},
		{"a key with a placeholder", context.Background(), number("{b}"), ptr("7"), `verdict: the check holds the placeholder of a template, which an instance replaces with a number: "{b}"`, ErrPlaceholder},
		{"a trap value with a placeholder", context.Background(), trapped(fraction("3/4", formatv0.CheckFormAny), "{whole}/{part}", "upside-down"), ptr("3/4"),
			`verdict: the check holds the placeholder of a template, which an instance replaces with a number: "{whole}/{part}"`, ErrPlaceholder},
		{"a choice with no key", context.Background(), noChoiceKey, ptr("a"), "verdict: the choice check has no key", nil},
		{"a choice written with no options", context.Background(), byHand("a"), ptr("a"), "verdict: the choice check has no options", nil},
		{"a choice tapped with no options", context.Background(), tapped("a"), ptr("a"), "verdict: the choice check has no options", nil},
		{"a predicate with no expression", context.Background(), noExpression, ptr("3"), "verdict: the predicate check has no predicate", nil},
		{"a rubric", context.Background(), rubric(), ptr("An essay."), "verdict: a rubric gives no verdict", ErrNoVerdict},
		{"a value that is not UTF-8", context.Background(), number("7"), ptr("7\xff"), "verdict: the value read is not UTF-8", nil},
		{"a context that is done", cancelled, number("7"), ptr("7"), "verdict: context canceled", context.Canceled},
	} {
		t.Run(c.name, func(t *testing.T) {
			got, err := OfSlot(c.ctx, c.slot, c.value)
			if err == nil {
				t.Fatalf("no error, and the verdict %q", short(got))
			}
			if err.Error() != c.want {
				t.Errorf("the error says %q, want %q", err, c.want)
			}
			if c.is != nil && !errors.Is(err, c.is) {
				t.Errorf("the error %v is not %v", err, c.is)
			}
		})
	}
}

// A predicate that its limit of steps or module-kit's clock stops is unchecked, and one that ctx
// stops gives no verdict. The loop of the clock and of ctx keeps no list and has no end of its own:
// only they stop it.
func TestPredicateStops(t *testing.T) {
	endless := predicate("[n for n in range(1 << 62) if False] == []")
	for _, c := range []struct {
		name   string
		slot   AnswerSlot
		limits sandbox.Limits
	}{
		{"the limit of steps", predicate(tooManySteps), sandbox.Predicate},
		{"the clock", endless, sandbox.Limits{Clock: 20 * time.Millisecond}},
	} {
		t.Run(c.name, func(t *testing.T) {
			got, err := c.slot.predicateVerdict(context.Background(), c.limits, ptr("49"))
			if err != nil {
				t.Fatal(err)
			}
			if short(got) != "unchecked" {
				t.Errorf("the verdict is %q, want unchecked", short(got))
			}
		})
	}
	t.Run("ctx", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
		defer cancel()
		got, err := endless.predicateVerdict(ctx, sandbox.Limits{}, ptr("49"))
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Fatalf("the error is %v, and the verdict %q", err, short(got))
		}
		if want := "sandbox: context deadline exceeded"; err.Error() != want {
			t.Errorf("the error says %q, want %q", err, want)
		}
	})
	// Within the limits of a predicate: 40 000 sorts of a range of 1000 take 800 000 steps, 20 a turn,
	// and some 750 ms, so that ctx stops them first (R28).
	t.Run("ctx through OfItem", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		time.AfterFunc(20*time.Millisecond, cancel)
		slow := predicate("[n for n in range(40000) if sorted(range(1000)) == None] == []")
		item := formatv0.Item{AnswerSlots: []formatv0.AnswerSlot{{Binding: formatv0.AnswerSlotBindingAnywhere, Check: slow.Check}}}
		if _, err := OfItem(ctx, item, []*string{ptr("49")}); !errors.Is(err, context.Canceled) {
			t.Fatalf("the error is %v", err)
		}
	})
}

// The rule task.predicate-expression: an expression of the dialect on value, the names it binds and
// the built-in names alone.
func TestCheckPredicate(t *testing.T) {
	for _, c := range []struct{ expression, want string }{
		{divisibleBy3, ""},
		{tooManySteps, ""},
		{"(lambda x: x > 0)(value) and chr(ord('a')) == 'a'", ""},
		{"value = 12", "1:8: got '=' after expression, want EOF"},
		{"number % 12 == 0", "1:1: undefined: number"},
		{"json.encode(value) == '12'", "1:1: undefined: json"},
		{"permutations([value])", "1:1: undefined: permutations"},
	} {
		err := CheckPredicate(c.expression)
		if got := fmt.Sprint(err); err == nil && c.want != "" || err != nil && got != c.want {
			t.Errorf("CheckPredicate(%q) gives %v, want %q", c.expression, err, c.want)
		}
	}
}

// The answer slots of an item: on a handwriting sheet its answer_slots, with the options of a choice
// tapped only, and on a workbook the boxes of its line.
func TestAnswerSlots(t *testing.T) {
	key := "7"
	check := formatv0.Check{Kind: formatv0.CheckKindNumber, Key: &key}
	boxBinding := formatv0.LinePartBindingBox
	line := formatv0.Item{Line: []formatv0.LinePart{
		{Text: new(formatv0.Text("3 +")), Binding: nil}, {Binding: &boxBinding, Check: &check}, {Text: new(formatv0.Text("= 10"))},
	}}
	if got := AnswerSlots(line); len(got) != 1 || got[0].Binding != "box" || *got[0].Check.Key != "7" || got[0].Options != nil {
		t.Errorf("the answer slots of a line are %+v", got)
	}
	sheet := formatv0.Item{AnswerSlots: []formatv0.AnswerSlot{
		{Binding: formatv0.AnswerSlotBindingZone, Check: check, Options: texts([]string{"x"})},
		{Binding: formatv0.AnswerSlotBindingChoice, Check: formatv0.Check{Kind: formatv0.CheckKindChoice, Key: &key}, Options: texts([]string{"7", "8"})},
	}}
	got := AnswerSlots(sheet)
	if len(got) != 2 || got[0].Binding != "zone" || got[0].Options != nil || got[1].Binding != "choice" || !slices.Equal(got[1].ChoiceOptions(), texts([]string{"7", "8"})) {
		t.Errorf("the answer slots of a sheet are %+v", got)
	}
	if slot := byHand("<", "<", ">"); !slices.Equal(slot.ChoiceOptions(), texts([]string{"<", ">"})) {
		t.Errorf("the options of a choice written by hand are %q", slot.ChoiceOptions())
	}
}

// The verdict of an item is made of its answer slots' verdicts.
func TestOfItem(t *testing.T) {
	keyA, keyB := "7", "8"
	boxBinding := formatv0.LinePartBindingBox
	item := formatv0.Item{Line: []formatv0.LinePart{
		{Binding: &boxBinding, Check: &formatv0.Check{Kind: formatv0.CheckKindNumber, Key: &keyA}},
		{Text: new(formatv0.Text("and"))},
		{Binding: &boxBinding, Check: &formatv0.Check{Kind: formatv0.CheckKindNumber, Key: &keyB}},
	}}
	for _, c := range []struct {
		values []*string
		want   string
	}{
		{[]*string{ptr("7"), ptr("8")}, "correct: correct, correct"},
		{[]*string{ptr("7"), ptr("9")}, "not_yet: correct, not_yet"},
		{[]*string{nil, ptr("9")}, "not_yet: unreadable blank, not_yet"},
		{[]*string{ptr("seven"), ptr("8")}, "unreadable: unreadable notation, correct"},
	} {
		got, err := OfItem(context.Background(), item, c.values)
		if err != nil {
			t.Fatal(err)
		}
		var slots []string
		for _, s := range got.Slots {
			slots = append(slots, short(s))
		}
		if text := string(got.Verdict) + ": " + strings.Join(slots, ", "); text != c.want {
			t.Errorf("OfItem gives %q, want %q", text, c.want)
		}
	}
	cancelled, cancel := context.WithCancel(context.Background())
	cancel()
	essay := formatv0.Item{AnswerSlots: []formatv0.AnswerSlot{{Binding: formatv0.AnswerSlotBindingSheet, Check: rubric().Check}}}
	for _, c := range []struct {
		ctx    context.Context
		item   formatv0.Item
		values []*string
		want   string
		is     error
	}{
		{context.Background(), item, []*string{ptr("7")}, "verdict: 1 values for the 2 answer slots of the item", nil},
		{context.Background(), formatv0.Item{}, nil, "verdict: the item has no answer slots", nil},
		{context.Background(), essay, []*string{ptr("An essay.")}, "verdict: answer slot 1: a rubric gives no verdict", ErrNoVerdict},
		{cancelled, item, []*string{ptr("7"), ptr("8")}, "verdict: context canceled", context.Canceled},
	} {
		_, err := OfItem(c.ctx, c.item, c.values)
		if err == nil || err.Error() != c.want {
			t.Errorf("the error is %v, want %q", err, c.want)
		}
		if c.is != nil && !errors.Is(err, c.is) {
			t.Errorf("the error %v is not %v", err, c.is)
		}
	}
}

// Every combination of the answer slots' verdicts gives the item's.
func TestItemOf(t *testing.T) {
	for _, c := range []struct {
		slots []Verdict
		want  Verdict
	}{
		{[]Verdict{Correct}, Correct},
		{[]Verdict{Correct, NotYet}, NotYet},
		{[]Verdict{Unreadable, NotYet, Unchecked}, NotYet},
		{[]Verdict{Correct, Unreadable}, Unreadable},
		{[]Verdict{Unchecked, Unreadable}, Unreadable},
		{[]Verdict{Correct, Unchecked}, Unchecked},
		{[]Verdict{Unchecked}, Unchecked},
		{nil, Unchecked},
	} {
		var slots []SlotVerdict
		for _, v := range c.slots {
			slots = append(slots, SlotVerdict{Verdict: v})
		}
		if got := ItemOf(slots); got != c.want {
			t.Errorf("ItemOf(%v) = %s, want %s", c.slots, got, c.want)
		}
	}
}

// Same compares two values as the check does.
func TestSame(t *testing.T) {
	for _, c := range []struct {
		slot AnswerSlot
		a, b string
		want bool
	}{
		{number("7", "rungs"), "6 rungs", "6", true},
		{number("7"), "0.5", "1/2", true},
		{number("7"), "seven", "seven", false},
		{number("7"), "", "", false},
		{fraction("3/4", formatv0.CheckFormReduced), "6/8", "3/4", true},
		{fraction("3/4", formatv0.CheckFormAny), "3", "3", false},
		{byHand("odd", "even", "odd"), "ODD", "odd", true},
		{byHand("odd", "even", "odd"), "seven", "SEVEN", true},
		{byHand("odd", "even", "odd"), "", " ", true},
		{byHand("odd", "even", "odd"), "odd", "odd.", false},
		{predicate(divisibleBy3), "7", "7.", true},
		{predicate(divisibleBy3), "7", "07", false},
		{rubric(), "An essay.", "An essay.", false},
		{byHand("odd", "even", "odd"), "odd\xff", "odd\xff", false},
	} {
		if got := Same(c.slot, c.a, c.b); got != c.want {
			t.Errorf("Same(%s, %q, %q) = %t, want %t", c.slot.Check.Kind, c.a, c.b, got, c.want)
		}
	}
}

// NormalForm gives a value as the kind normalizes it, and whether it is in the kind's notation.
func TestNormalForm(t *testing.T) {
	brokenKey := byHand("none", "even", "odd")
	for _, c := range []struct {
		slot     AnswerSlot
		value    string
		form     string
		notation bool
	}{
		{number("7", "rungs"), "= 7 Rungs.", "7", true},
		{number("7", "rungs"), "rungs", "rungs", false},
		{number("7"), "7²", "7�", false},
		{number("7"), "=", "", false},
		{fraction("3/4", formatv0.CheckFormReduced), "6 / 8", "6 / 8", true},
		{fraction("3/4", formatv0.CheckFormAny), "0.75", "0.75", false},
		{predicate(divisibleBy3), "-0", "-0", true},
		{predicate(divisibleBy3), "012", "012", false},
		{byHand("odd", "even", "odd"), " EVEN ", "even", true},
		{byHand("odd", "even", "odd"), "seven", "seven", false},
		{byHand("odd", "even", "odd"), " ", "", false},
		{byHand("=", "<", ">", "="), "=", "=", true},
		{tapped("yes", "yes", "no"), "No", "no", true},
		{brokenKey, "NONE", "none", true},
		{rubric(), " An essay. ", " An essay. ", false},
		{byHand("odd", "even", "odd\uFFFD"), "odd\xff", "odd\xff", false},
	} {
		form, notation := NormalForm(c.slot, c.value)
		if form != c.form || notation != c.notation {
			t.Errorf("NormalForm(%s, %q) = %q, %t, want %q, %t", c.slot.Check.Kind, c.value, form, notation, c.form, c.notation)
		}
	}
}

// The version that a release names is that of the tables of NFKC and of the package unicode alike
// (format/README.md, section 15; R27).
func TestUnicodeVersion(t *testing.T) {
	for name, version := range map[string]string{"UnicodeVersion": UnicodeVersion, "norm.Version": norm.Version, "unicode.Version": unicode.Version} {
		if version != "17.0.0" {
			t.Errorf("%s is %s: the normalization follows one version, 17.0.0, until a task raises it", name, version)
		}
	}
}
