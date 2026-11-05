package random

import (
	"errors"
	"math"
	"reflect"
	"slices"
	"strings"
	"testing"
)

// The words of a place: the high one is FNV-1a of the text of the seed and the names, which the
// vectors give, and the low one the seed. A list with no name, and a name that is empty, holds | or
// is not UTF-8, builds no place, wherever it stands in the list.
func TestWords(t *testing.T) {
	for _, c := range []struct {
		seed      uint64
		names     []string
		high, low uint64
	}{
		{7, []string{"a"}, 3879510333719213901, 7},                             // one-name.yaml
		{42, []string{"шаг", "é"}, 4663844351231291435, 42},                    // non-ascii-name.yaml
		{0, []string{"zero"}, 17169638463767760023, 0},                         // seed-zero.yaml
		{math.MaxUint64, []string{"max"}, 2735668769848699190, math.MaxUint64}, // seed-max.yaml
	} {
		high, low, err := Words(c.seed, c.names)
		if err != nil || high != c.high || low != c.low {
			t.Errorf("Words(%d, %q) = %d, %d, %v, want %d, %d", c.seed, c.names, high, low, err, c.high, c.low)
		}
		place, err := Place(c.seed, c.names...)
		if err != nil || *place != (Stream{c.high, c.low}) {
			t.Errorf("Place(%d, %q) = %v, %v, want the stream of %d and %d", c.seed, c.names, place, err, c.high, c.low)
		}
	}
	for _, names := range [][]string{nil, {}, {""}, {"a|b"}, {"|"}, {"\xff"}, {"a", ""}, {"a", "b", "c|"}, {"template", "x", "\xc3"}} {
		high, low, err := Words(1, names)
		if !errors.Is(err, ErrDomain) || high != 0 || low != 0 {
			t.Errorf("Words(1, %q) = %d, %d, %v, want ErrDomain", names, high, low, err)
		}
		if place, err := Place(1, names...); !errors.Is(err, ErrDomain) || place != nil {
			t.Errorf("Place(1, %q) = %v, %v, want ErrDomain", names, place, err)
		}
	}
	// Names compare as written: no normalization makes é of one character and of two the same.
	composed, _, _ := Words(1, []string{"\u00e9"})
	decomposed, _, _ := Words(1, []string{"e\u0301"})
	if composed == decomposed {
		t.Error("é composed and decomposed give one place")
	}
}

// drawn counts the integers that a call draws from a stream: the steps from the state before it to
// the state after it, up to limit.
func drawn(t *testing.T, before Stream, call func(*Stream), limit int) int {
	t.Helper()
	after := before
	call(&after)
	for n := 0; n <= limit; n++ {
		if before == after {
			return n
		}
		before.Integer()
	}
	t.Fatalf("the call draws more than %d integers", limit)
	return 0
}

// Each function draws as many integers as the format's text says: one each for an integer, a bound
// that is a power of two, below(1) and between(a, a) among them, and a fraction; one more for each
// product that Lemire's method refuses; none for an order of 0 or 1 integers, and one for each swap
// of a longer one (format/README.md, section 22).
func TestDraws(t *testing.T) {
	below := func(b uint64) func(*Stream) { return func(s *Stream) { s.Below(b) } }
	between := func(a, c int64) func(*Stream) { return func(s *Stream) { s.Between(a, c) } }
	permutation := func(n int) func(*Stream) { return func(s *Stream) { s.Permutation(n) } }
	for _, c := range []struct {
		name      string
		high, low uint64
		call      func(*Stream)
		draws     int
	}{
		{"integer", 1, 2, func(s *Stream) { s.Integer() }, 1},
		{"below(1)", 1, 2, below(1), 1},
		{"below(2)", 1, 2, below(2), 1},
		{"below(2^63)", 1, 2, below(1 << 63), 1},
		{"below(3)", 1, 2, below(3), 1},
		{"below(2^64 − 1)", 1, 2, below(math.MaxUint64), 1},
		// below-refusals.yaml: from the words 0 and 2, the bound 2^63 + 1 refuses four products.
		{"below(2^63 + 1), refusing four products", 0, 2, below(1<<63 + 1), 5},
		{"between(a, a)", 1, 2, between(-3, -3), 1},
		{"between(-10, 10)", 1, 2, between(-10, 10), 1},
		{"between at the widest count", 1, 2, between(math.MinInt64, -2), 1},
		{"fraction", 1, 2, func(s *Stream) { s.Fraction() }, 1},
		{"permutation(0)", 1, 2, permutation(0), 0},
		{"permutation(1)", 1, 2, permutation(1), 0},
		{"permutation(2)", 1, 2, permutation(2), 1},
		// permutation.yaml: an order of 16 draws 15 at least, and none of its bounds is refused here.
		{"permutation(16)", 650905648525575843, 2026, permutation(16), 15},
	} {
		if n := drawn(t, Stream{c.high, c.low}, c.call, 100); n != c.draws {
			t.Errorf("%s draws %d integers, want %d", c.name, n, c.draws)
		}
	}
}

// On an argument outside its domain a function gives ErrDomain, its zero value, and draws nothing.
func TestRefusals(t *testing.T) {
	past := int64(MaxPermutation)
	past++ // at run time, so that a 32-bit int wraps it rather than the compiler refuse it
	for _, c := range []struct {
		name string
		call func(*Stream) (any, error)
		says string
	}{
		{"below(0)", func(s *Stream) (any, error) { return s.Below(0) }, "below(0)"},
		{"between(1, 0)", func(s *Stream) (any, error) { return s.Between(1, 0) }, "the first is greater"},
		{"between(max, min)", func(s *Stream) (any, error) { return s.Between(math.MaxInt64, math.MinInt64) }, "the first is greater"},
		{"between(min, max)", func(s *Stream) (any, error) { return s.Between(math.MinInt64, math.MaxInt64) }, "more than 2^63 − 1"},
		{"between(0, max), counting 2^63", func(s *Stream) (any, error) { return s.Between(0, math.MaxInt64) }, "more than 2^63 − 1"},
		{"between(min, -1), counting 2^63", func(s *Stream) (any, error) { return s.Between(math.MinInt64, -1) }, "more than 2^63 − 1"},
		{"permutation(-1)", func(s *Stream) (any, error) { return s.Permutation(-1) }, "permutation(-1)"},
		{"permutation(2^31)", func(s *Stream) (any, error) { return s.Permutation(int(past)) }, "the count is from 0"},
	} {
		s := Stream{1, 2}
		value, err := c.call(&s)
		if !errors.Is(err, ErrDomain) || !strings.Contains(err.Error(), c.says) {
			t.Errorf("%s: %v, want ErrDomain saying %q", c.name, err, c.says)
		}
		if !reflect.ValueOf(value).IsZero() {
			t.Errorf("%s gives %v", c.name, value)
		}
		if s != (Stream{1, 2}) {
			t.Errorf("%s draws an integer", c.name)
		}
	}
}

// At the edges of their domains the functions give what their domains hold: between(a, a) gives
// a at both ends of the integers, the widest counts reach their ends, and below(1) gives 0.
func TestEdges(t *testing.T) {
	s := New(1, 2)
	for _, c := range []struct{ a, c int64 }{
		{math.MinInt64, math.MinInt64},
		{math.MaxInt64, math.MaxInt64},
		{0, 0},
	} {
		if v, err := s.Between(c.a, c.c); err != nil || v != c.a {
			t.Errorf("between(%d, %d) = %d, %v", c.a, c.c, v, err)
		}
	}
	for _, c := range []struct{ a, c int64 }{
		{math.MinInt64, math.MinInt64 + math.MaxInt64 - 1},
		{1, math.MaxInt64},
		{0, math.MaxInt64 - 1},
	} {
		for range 100 {
			if v, err := s.Between(c.a, c.c); err != nil || v < c.a || v > c.c {
				t.Fatalf("between(%d, %d) = %d, %v", c.a, c.c, v, err)
			}
		}
	}
	for range 100 {
		if v, err := s.Below(1); err != nil || v != 0 {
			t.Fatalf("below(1) = %d, %v", v, err)
		}
		if f := s.Fraction(); f < 0 || f >= 1 || f*(1<<53) != math.Trunc(f*(1<<53)) {
			t.Fatalf("fraction gives %v, no multiple of 2^-53 below 1", f)
		}
	}
	order, err := s.Permutation(64)
	if err != nil || len(order) != 64 {
		t.Fatalf("permutation(64) = %v, %v", order, err)
	}
	for i, v := range slices.Sorted(slices.Values(order)) {
		if v != i {
			t.Fatalf("permutation(64) = %v, no order of 0 to 63", order)
		}
	}
}
