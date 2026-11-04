package verdict

import (
	"context"
	"math/big"
	"slices"
	"strconv"
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

// FuzzNotation writes numbers in the notation of each kind and reads them back: an integer, a
// simple fraction with spaces or none, and a decimal, of at most 16 characters, read as the same
// rational, in lowest terms when the numerator and the denominator as written have no common
// divisor; longer, they are out of the notation. Text of any other form is read without a panic.
func FuzzNotation(f *testing.F) {
	f.Add(int64(7), int64(4), "7")
	f.Add(int64(-6), int64(8), "-6/8")
	f.Add(int64(0), int64(1), "0.5")
	f.Add(int64(-9007199254740993), int64(1), "1234567890.123450")
	f.Fuzz(func(t *testing.T, n, d int64, text string) {
		integer := strconv.FormatInt(n, 10)
		read(t, integer, formatv0.CheckKindNumber, new(big.Rat).SetInt64(n), true)
		read(t, integer, formatv0.CheckKindPredicate, new(big.Rat).SetInt64(n), true)
		read(t, integer, formatv0.CheckKindFraction, nil, false)
		if d > 0 {
			value := new(big.Rat).SetFrac64(n, d)
			lowest := new(big.Int).GCD(nil, nil, new(big.Int).Abs(big.NewInt(n)), big.NewInt(d)).Cmp(one) == 0
			denominator := strconv.FormatInt(d, 10)
			for _, fraction := range []string{integer + "/" + denominator, integer + " / " + denominator, integer + " /" + denominator} {
				read(t, fraction, formatv0.CheckKindNumber, value, lowest)
				read(t, fraction, formatv0.CheckKindFraction, value, lowest)
				read(t, fraction, formatv0.CheckKindPredicate, nil, false)
			}
		}
		if d >= 0 {
			digits := strconv.FormatInt(d, 10)
			part := new(big.Rat).SetFrac(big.NewInt(d), new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(len(digits))), nil))
			whole := new(big.Rat).SetInt64(n)
			value := new(big.Rat).Add(whole, part)
			if n < 0 {
				value.Sub(whole, part)
			}
			decimal := integer + "." + digits
			read(t, decimal, formatv0.CheckKindNumber, value, true)
			read(t, decimal, formatv0.CheckKindFraction, nil, false)
		}
		for _, kind := range []formatv0.CheckKind{formatv0.CheckKindNumber, formatv0.CheckKindFraction, formatv0.CheckKindPredicate} {
			if w, ok := notation(text, kind); ok && (len(text) > maxNotation || w.value == nil) {
				t.Fatalf("%q is in the notation of %s, with %d bytes", text, kind, len(text))
			}
		}
		_ = numberText(text, []string{"cm", "rungs"})
	})
}

// read says that a text is in the notation of a kind, with a value, when it is short enough and the
// value is given, and out of it otherwise.
func read(t *testing.T, text string, kind formatv0.CheckKind, value *big.Rat, lowest bool) {
	t.Helper()
	w, ok := notation(text, kind)
	if value == nil || len(text) > maxNotation {
		if ok {
			t.Fatalf("%q is in the notation of %s", text, kind)
		}
		return
	}
	if !ok {
		t.Fatalf("%q is out of the notation of %s", text, kind)
	}
	if w.value.Cmp(value) != 0 || w.lowest != lowest {
		t.Fatalf("%q reads as %s, in lowest terms %t, want %s, %t", text, w.value, w.lowest, value, lowest)
	}
}
