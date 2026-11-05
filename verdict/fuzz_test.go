package verdict

import (
	"context"
	"slices"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/altnovatrail/spec/types/formatv0"
)

// The answer slots that FuzzOfSlot judges values in: every answer slot of the cases of TestOfSlot.
func fuzzSlots() []AnswerSlot {
	var slots []AnswerSlot
	for _, c := range slotCases {
		slots = append(slots, c.slot)
	}
	return slots
}

// FuzzOfSlot judges any value in the answer slots of the cases: a value that is UTF-8 gets a verdict
// and nothing else, and that verdict holds together with Same. The corpus is the values of the cases
// and a run of 31 combining marks, which NFKC breaks by U+034F (R27).
func FuzzOfSlot(f *testing.F) {
	for i, c := range slotCases {
		if c.value != nil {
			f.Add(uint8(i), *c.value)
		}
	}
	f.Add(uint8(0), "a"+strings.Repeat("\u0301", 31))
	slots := fuzzSlots()
	f.Fuzz(func(t *testing.T, which uint8, value string) {
		slot := slots[int(which)%len(slots)]
		got, err := OfSlot(context.Background(), slot, &value)
		if !utf8.ValidString(value) {
			if err == nil {
				t.Fatalf("a value that is not UTF-8 got %q", short(got))
			}
			return
		}
		kind := slot.Check.Kind
		if err != nil {
			t.Fatalf("%s: %v", kind, err)
		}
		switch got.Verdict {
		case Correct, NotYet:
			if got.Reason != "" {
				t.Fatalf("%s with the reason %s", got.Verdict, got.Reason)
			}
		case Unreadable:
			if got.Reason != Blank && got.Reason != Notation {
				t.Fatalf("unreadable with the reason %q", got.Reason)
			}
		case Unchecked:
			if kind != formatv0.CheckKindPredicate || got.Reason != "" || got.Trap != nil {
				t.Fatalf("%s gives %q", kind, short(got))
			}
		default:
			t.Fatalf("the verdict %q", got.Verdict)
		}
		if got.Trap != nil {
			if got.Verdict != NotYet || !slices.ContainsFunc(slot.Check.TrapValues, func(v formatv0.TrapValue) bool { return v.Value == got.Trap.Value }) ||
				!Same(slot, got.Trap.Value, value) {
				t.Fatalf("%s matches the trap value %+v", got.Verdict, *got.Trap)
			}
		}
		same := Same(slot, value, value)
		switch kind {
		case formatv0.CheckKindChoice:
			if !same {
				t.Fatal("a choice compares a value unequal to itself")
			}
		default:
			if same != (got.Verdict != Unreadable) {
				t.Fatalf("Same of the value with itself is %t, and the verdict %q", same, short(got))
			}
			if got.Verdict == Correct && kind != formatv0.CheckKindPredicate && !Same(slot, *slot.Check.Key, value) {
				t.Fatal("a right value that the check compares unequal to the key")
			}
		}
	})
}
