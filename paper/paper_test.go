package paper

import (
	"context"
	"strings"
	"testing"
	"unicode/utf8"

	observationv1 "github.com/altnovatrail/spec/contracts/novatrail/observation/v1"
	paperv1 "github.com/altnovatrail/spec/contracts/novatrail/paper/v1"
	"google.golang.org/protobuf/proto"

	"github.com/altnovatrail/module-kit/observation"
	"github.com/altnovatrail/module-kit/verdict"
)

// The key is the solver's when it gives an answer, and otherwise AI's when it breaks no rule; with
// neither, it has no answer and no provenance.
func TestKey(t *testing.T) {
	solved, _ := Solve("37+48")
	if k := Key(solved, nil); k.GetProvenance() != paperv1.Provenance_PROVENANCE_SOLVER || k.GetAnswer().GetNumber().GetValue() != "85" {
		t.Errorf("the solver's key is %v", k)
	}
	ofAI := &paperv1.Answer{Kind: &paperv1.Answer_Choice{Choice: &paperv1.Choice{Options: []string{"A", "B"}, Key: "B"}}}
	if k := Key(nil, ofAI); k.GetProvenance() != paperv1.Provenance_PROVENANCE_AI || !proto.Equal(k.GetAnswer(), ofAI) {
		t.Errorf("AI's key is %v", k)
	}
	broken := &paperv1.Answer{Kind: &paperv1.Answer_Number{Number: &paperv1.Number{Value: "2,5"}}}
	for _, k := range []*paperv1.TaskKey{Key(nil, broken), Key(nil, nil), Key(nil, &paperv1.Answer{})} {
		if k.GetProvenance() != paperv1.Provenance_PROVENANCE_UNSPECIFIED || k.GetAnswer() != nil {
			t.Errorf("an answer that breaks a rule, or none, gives the key %v", k)
		}
	}
}

// A task with no key gives no item: no check is written of it, and the caller gives it unchecked.
func TestItemsOfNoKey(t *testing.T) {
	for _, answer := range []*paperv1.Answer{nil, {}, Key(nil, nil).GetAnswer()} {
		if items := Items(answer); items != nil {
			t.Errorf("an answer of no kind gives the items %v", items)
		}
	}
}

// Every unit is read for an empty one before two are compared, as the rules come in order.
func TestBreachOrder(t *testing.T) {
	for units, want := range map[string]Rule{"cm,cm,": UnitsEmpty, ",cm,cm": UnitsEmpty, "cm,CM": UnitsSame, "cm,m": ""} {
		answer := &paperv1.Answer{Kind: &paperv1.Answer_Number{Number: &paperv1.Number{Value: "12", Units: strings.Split(units, ",")}}}
		if got := Breach(answer); got != want {
			t.Errorf("the units %q break %q, not %q", units, got, want)
		}
	}
}

// The sides of an equation brought together are bounded as any part on the way, and spaces at
// either end of a formula are between no tokens and change nothing.
func TestSolveBounds(t *testing.T) {
	// c is of 100 digits, below 10^100, and 2c, the coefficient of x once the sides are brought
	// together, is of 101.
	c := "(999999999999999^6*9999999999)"
	if n, reason := Solve(c + "*x-" + c + "=" + c + "-" + c + "*x"); reason != TooLarge {
		t.Errorf("an equation whose sides brought together are too large gives %v %q", n, reason)
	}
	if n, reason := Solve(c + "*x-" + c + "=0"); reason != "" || n.GetValue() != "1" {
		t.Errorf("an equation of c alone gives %v %q", n, reason)
	}
	if n, reason := Solve(" 1+1 "); reason != "" || n.GetValue() != "2" {
		t.Errorf("%q gives %v %q", " 1+1 ", n, reason)
	}
}

// Two readers who read x = 7 and 7 read the same value once both are brought to the key, so the
// join gives no ambiguity.
func TestJoinOfBroughtValues(t *testing.T) {
	answer := &paperv1.Answer{Kind: &paperv1.Answer_Number{Number: &paperv1.Number{Value: "7", Unknown: "x"}}}
	read := func(value string) *observationv1.Observation {
		return Brought(answer, &observationv1.Observation{AnswerSlots: []*observationv1.AnswerSlotObservation{{
			AnswerSlot: &observationv1.AnswerSlotPlace{Item: 1, Slot: 1},
			Read:       &observationv1.AnswerSlotObservation_Value{Value: value},
		}}})
	}
	items := Items(answer)
	joined, err := observation.Join(items, read("x = 7"), read("7"))
	if err != nil {
		t.Fatal(err)
	}
	got, err := observation.Verdicts(context.Background(), items, joined, nil)
	if err != nil {
		t.Fatal(err)
	}
	if got[0].Verdict != verdict.Correct {
		t.Errorf("the joined reading gives %v", got[0])
	}
}

// Brought copies the observation and leaves a slot with no value as it is.
func TestBroughtCopies(t *testing.T) {
	answer := &paperv1.Answer{Kind: &paperv1.Answer_Choice{Choice: &paperv1.Choice{Options: []string{"A", "B"}, Key: "A"}}}
	o := &observationv1.Observation{AnswerSlots: []*observationv1.AnswerSlotObservation{
		{AnswerSlot: &observationv1.AnswerSlotPlace{Item: 1, Slot: 1}, Read: &observationv1.AnswerSlotObservation_Value{Value: "(A)"}},
	}}
	b := Brought(answer, o)
	if o.GetAnswerSlots()[0].GetValue() != "(A)" || b.GetAnswerSlots()[0].GetValue() != "A" {
		t.Errorf("Brought gives %v of %v", b, o)
	}
	if got := BroughtValue(&paperv1.Answer{}, "x = 1"); got != "x = 1" {
		t.Errorf("an answer of no kind brings %q", got)
	}
}

// The parser takes what the grammar has and nothing else.
func TestParse(t *testing.T) {
	for _, formula := range []string{"1", "(1)", "-x=1", "2^10", "1.5*x=3", "x--1=0"} {
		if _, ok := parse(formula); !ok {
			t.Errorf("%q is of the grammar", formula)
		}
	}
	for _, formula := range []string{"", "1 +", "+1", "()", "2^", "2^x", "2^1.5", "1..2", "1.", "x=", "=1", "1\t+1", "é", strings.Repeat("1", 16)} {
		if _, ok := parse(formula); ok {
			t.Errorf("%q is not of the grammar", formula)
		}
	}
}

// The solver never panics, and an answer it gives is a number of at most 16 characters that is an
// answer of a key.
func FuzzSolve(f *testing.F) {
	for _, seed := range []string{"37+48", "3*x-5=16", "x/0=1", "(2^10)^10", "x*x=1", "-(-(-1))", "1/3+1/7"} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, formula string) {
		n, reason := Solve(formula)
		if (n == nil) == (reason == "") {
			t.Fatalf("%q gives %v and %q", formula, n, reason)
		}
		if n == nil {
			return
		}
		if utf8.RuneCountInString(n.GetValue()) > maxAnswer {
			t.Fatalf("%q gives %q, longer than 16 characters", formula, n.GetValue())
		}
		if rule := Breach(&paperv1.Answer{Kind: &paperv1.Answer_Number{Number: n}}); rule != "" {
			t.Fatalf("%q gives %v, which breaks %s", formula, n, rule)
		}
	})
}

// BroughtValue never panics, and gives a valid UTF-8 value of a valid one.
func FuzzBroughtValue(f *testing.F) {
	number := &paperv1.Answer{Kind: &paperv1.Answer_Number{Number: &paperv1.Number{Value: "4", Unknown: "x"}}}
	choice := &paperv1.Answer{Kind: &paperv1.Answer_Choice{Choice: &paperv1.Choice{Options: []string{"A", "B"}, Key: "A"}}}
	for _, seed := range []string{"x = 4", "X=4", "(A)", "[B]", "A)", "x²=4", "ｘ＝4"} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, value string) {
		if !utf8.ValidString(value) {
			return
		}
		for _, answer := range []*paperv1.Answer{number, choice} {
			if got := BroughtValue(answer, value); !utf8.ValidString(got) {
				t.Fatalf("%q gives %q", value, got)
			}
		}
	})
}
