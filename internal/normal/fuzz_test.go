package normal

import (
	"math/big"
	"strconv"
	"testing"

	"github.com/altnovatrail/spec/types/formatv0"
)

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
			if w, ok := Notation(text, kind); ok && (len(text) > MaxNotation || w.Value == nil) {
				t.Fatalf("%q is in the notation of %s, with %d bytes", text, kind, len(text))
			}
		}
		_ = Number(text, []string{"cm", "rungs"})
		_ = Choice(Scripts(text))
	})
}

// read says that a text is in the notation of a kind, with a value, when it is short enough and the
// value is given, and out of it otherwise.
func read(t *testing.T, text string, kind formatv0.CheckKind, value *big.Rat, lowest bool) {
	t.Helper()
	w, ok := Notation(text, kind)
	if value == nil || len(text) > MaxNotation {
		if ok {
			t.Fatalf("%q is in the notation of %s", text, kind)
		}
		return
	}
	if !ok {
		t.Fatalf("%q is out of the notation of %s", text, kind)
	}
	if w.Value.Cmp(value) != 0 || w.Lowest != lowest {
		t.Fatalf("%q reads as %s, in lowest terms %t, want %s, %t", text, w.Value, w.Lowest, value, lowest)
	}
}
