package random

import (
	"bytes"
	"encoding/binary"
	"errors"
	"hash/fnv"
	"io/fs"
	"math"
	"math/big"
	"math/rand/v2"
	"slices"
	"strconv"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/altnovatrail/spec"

	"github.com/altnovatrail/module-kit/internal/yaml"
)

// The reference is Go's: the generator PCG of math/rand/v2 with its methods Uint64, Uint64N,
// Int64N, Float64 and Perm, and FNV-1a of 64 bits of hash/fnv, which the format's text names beside
// its own (format/README.md, section 22). It lives in the tests alone (R10), and a failure here
// after a move to a newer Go means that Go changed its streams: a question for spec, while the code
// stays as the format's text has it.

// An input is what the generator is compared with the reference on: the words that start a stream,
// the arguments of below, between and permutation, and the seed and names of a place. The names are
// one string, joined by | as the text of their place joins them: a name holds no |, so every list of
// names is one such string, and a fuzzer reaches every name, a line break in it too. That a name
// with a | builds no place, TestWords holds.
type input struct {
	high, low uint64
	bound     uint64
	from, to  int64
	count     int
	seed      uint64
	names     string
}

// The most integers that the comparison orders: a longer permutation is only refused or skipped.
const maxCompared = 1024

// The bounds of below, from 0, which is refused: 1, every power of two, the integers on either side
// of each, and those of the vectors (format/README.md, section 22).
func bounds() []uint64 {
	b := []uint64{0, 3, 5, 6, 7, 10, 1000, 10000, math.MaxInt64, 1<<63 + 1, math.MaxUint64 - 1, math.MaxUint64}
	for k := range 64 {
		b = append(b, 1<<k, 1<<k+1)
		if k > 1 {
			b = append(b, 1<<k-1)
		}
	}
	slices.Sort(b)
	return slices.Compact(b)
}

// The arguments of between, those it refuses among them: one integer alone at either end, the
// widest counts, 2^63 − 1, and the counts past them, and the vectors' (format/README.md, section
// 22).
var betweens = [][2]int64{
	{-10, 10}, {-3, -3}, {0, 0}, {math.MinInt64, math.MinInt64}, {math.MaxInt64, math.MaxInt64},
	{math.MinInt64, -2}, {math.MinInt64, math.MinInt64 + math.MaxInt64 - 1}, {1, math.MaxInt64}, {0, math.MaxInt64 - 1},
	{-1, 1 << 62}, {math.MinInt64 / 2, math.MaxInt64 / 2},
	{math.MinInt64, -1}, {0, math.MaxInt64}, {math.MinInt64, math.MaxInt64}, {1, 0}, {math.MaxInt64, math.MinInt64},
}

// The counts of permutation, -1 refused: 0 and 1, which draw nothing, and orders up to 1024.
var counts = []int{0, 1, 2, 3, 16, 100, maxCompared, -1}

// The names of places beside those of the vectors: the place of a template, a name with a line break
// in it, and lists that build no place — an empty name first, between and last, and what is not
// UTF-8.
var placeNames = []string{"template|heads-and-legs-2|instance", "a", "a line\nand the next", "", "|a", "a||b", "a|", "\xff", "e-a3|g2/0|params"}

// vectorInputs gives the words, the seed and the names of each vector of the generator, read from
// spec.Format by the reader of YAML. A vector of words alone takes a seed and names of its own.
func vectorInputs(tb testing.TB) []input {
	tb.Helper()
	files, err := fs.Glob(spec.Format, "vectors/random/*.yaml")
	if err != nil || len(files) == 0 {
		tb.Fatalf("the vectors of the generator: %v", err)
	}
	var inputs []input
	for i, file := range files {
		data, err := fs.ReadFile(spec.Format, file)
		if err != nil {
			tb.Fatal(err)
		}
		value, err := yaml.Read(data)
		if err != nil {
			tb.Fatalf("%s: %v", file, err)
		}
		v := value.(map[string]any)
		state := v["state"].(map[string]any)
		in := input{high: parse(tb, state["high"]), low: parse(tb, state["low"]), seed: uint64(i), names: placeNames[i%len(placeNames)]}
		if seed, ok := v["seed"]; ok {
			in.seed = parse(tb, seed)
			var list []string
			for _, name := range v["names"].([]any) {
				list = append(list, name.(string))
			}
			in.names = strings.Join(list, "|")
		}
		inputs = append(inputs, in)
	}
	return inputs
}

func parse(tb testing.TB, text any) uint64 {
	tb.Helper()
	n, err := strconv.ParseUint(text.(string), 10, 64)
	if err != nil {
		tb.Fatal(err)
	}
	return n
}

// fixedInputs gives the words of every vector, and of 0 and of 2^64 − 1, each with every bound and
// with the arguments of between, the counts and the names in turn: the inputs of the ordinary test
// and the corpus of the fuzz test.
func fixedInputs(tb testing.TB) []input {
	tb.Helper()
	words := append(vectorInputs(tb), input{high: 0, low: 0, names: "zero"}, input{high: math.MaxUint64, low: math.MaxUint64, seed: math.MaxUint64, names: "max"})
	var inputs []input
	for _, w := range words {
		for i, b := range bounds() {
			in := w
			in.bound = b
			in.from, in.to = betweens[i%len(betweens)][0], betweens[i%len(betweens)][1]
			in.count = counts[i%len(counts)]
			if i > 0 {
				in.names = placeNames[i%len(placeNames)]
			}
			inputs = append(inputs, in)
		}
	}
	return inputs
}

// generatedInputs gives inputs that the generator itself draws, from a place of a constant seed:
// bounds and arguments of every magnitude, orders of up to 64 integers, and names of letters that
// UTF-8 writes in one byte and in more, a space and a line break among them, with a | that ends a
// name now and then, and so an empty name too.
func generatedInputs(tb testing.TB, n int) []input {
	tb.Helper()
	g, err := Place(2026, "reference")
	if err != nil {
		tb.Fatal(err)
	}
	letters := []string{"a", "z", "0", "-", "/", "_", " ", "\n", "é", "ш", "€", "😀", "|"}
	inputs := make([]input, n)
	for i := range inputs {
		in := input{high: g.Integer(), low: g.Integer(), seed: g.Integer()}
		in.bound = g.Integer() >> g.below(64)
		in.from = int64(g.Integer()) >> g.below(64)
		in.to = int64(g.Integer()) >> g.below(64)
		in.count = int(g.below(65))
		list := make([]string, 1+g.below(3))
		for j := range list {
			for range 1 + g.below(6) {
				list[j] += letters[g.below(uint64(len(letters)))]
			}
		}
		in.names = strings.Join(list, "|")
		inputs[i] = in
	}
	return inputs
}

// state gives the state of a stream as PCG's MarshalBinary writes its own: "pcg:", then the high
// word and the low one, big-endian.
func state(s *Stream) []byte {
	b := binary.BigEndian.AppendUint64([]byte("pcg:"), s.high)
	return binary.BigEndian.AppendUint64(b, s.low)
}

// compare holds the generator to the reference on one input: the words of the place, and then, on
// one stream, an integer and each of the four other functions, with the stream's state the
// reference's after each, so that each function draws as many integers as Go's. On an argument
// outside its domain a function gives ErrDomain and draws nothing, where Go would panic.
func compare(t testing.TB, in input) {
	t.Helper()
	list := strings.Split(in.names, "|")
	high, low, err := Words(in.seed, list)
	if slices.ContainsFunc(list, func(n string) bool { return n == "" || !utf8.ValidString(n) }) {
		if !errors.Is(err, ErrDomain) {
			t.Fatalf("the names %q build a place: %d, %d, %v", list, high, low, err)
		}
	} else {
		h := fnv.New64a()
		h.Write([]byte(strconv.FormatUint(in.seed, 10)))
		for _, name := range list {
			h.Write([]byte("|" + name))
		}
		if err != nil || high != h.Sum64() || low != in.seed {
			t.Fatalf("the seed %d and the names %q: %d, %d, %v, and hash/fnv gives %d, %d", in.seed, list, high, low, err, h.Sum64(), in.seed)
		}
	}

	s := New(in.high, in.low)
	pcg := rand.NewPCG(in.high, in.low)
	r := rand.New(pcg)
	same := func(what string) {
		t.Helper()
		if want, _ := pcg.MarshalBinary(); !bytes.Equal(state(s), want) {
			t.Fatalf("from the words %d and %d, after %s the state is %x, and PCG's is %x", in.high, in.low, what, state(s), want)
		}
	}
	same("nothing")
	if got, want := s.Integer(), r.Uint64(); got != want {
		t.Fatalf("from the words %d and %d, integer gives %d, and Uint64 %d", in.high, in.low, got, want)
	}
	same("integer")

	got, err := s.Below(in.bound)
	if in.bound == 0 {
		if !errors.Is(err, ErrDomain) || got != 0 {
			t.Fatalf("below(0) gives %d, %v", got, err)
		}
	} else if want := r.Uint64N(in.bound); err != nil || got != want {
		t.Fatalf("from the words %d and %d, below(%d) gives %d, %v, and Uint64N %d", in.high, in.low, in.bound, got, err, want)
	}
	same("below")

	// between takes a ≤ c with c − a + 1 at most 2^63 − 1, counted here without words that wrap.
	count := new(big.Int).Sub(big.NewInt(in.to), big.NewInt(in.from))
	count.Add(count, big.NewInt(1))
	value, err := s.Between(in.from, in.to)
	if count.Sign() <= 0 || count.Cmp(big.NewInt(math.MaxInt64)) > 0 {
		if !errors.Is(err, ErrDomain) || value != 0 {
			t.Fatalf("between(%d, %d) gives %d, %v", in.from, in.to, value, err)
		}
	} else if want := in.from + r.Int64N(in.to-in.from+1); err != nil || value != want {
		t.Fatalf("from the words %d and %d, between(%d, %d) gives %d, %v, and Int64N %d", in.high, in.low, in.from, in.to, value, err, want)
	}
	same("between")

	if got, want := s.Fraction(), r.Float64(); got != want {
		t.Fatalf("from the words %d and %d, fraction gives %v, and Float64 %v", in.high, in.low, got, want)
	}
	same("fraction")

	switch {
	case in.count < 0 || int64(in.count) > MaxPermutation:
		if order, err := s.Permutation(in.count); !errors.Is(err, ErrDomain) || order != nil {
			t.Fatalf("permutation(%d) gives %v, %v", in.count, order, err)
		}
	case in.count <= maxCompared:
		if order, err := s.Permutation(in.count); err != nil || !slices.Equal(order, r.Perm(in.count)) {
			t.Fatalf("from the words %d and %d, permutation(%d) gives %v, %v, and Perm another order", in.high, in.low, in.count, order, err)
		}
	}
	same("permutation")
}

// PCG's MarshalBinary writes its state as state does: else the comparison of states compares
// nothing.
func TestStateAsPCGWritesIt(t *testing.T) {
	want, _ := rand.NewPCG(1, 2).MarshalBinary()
	if got := state(New(1, 2)); !bytes.Equal(got, want) {
		t.Fatalf("the state of the words 1 and 2 is %x, and PCG writes %x", got, want)
	}
}

// The generator, the five functions and the words of a place are Go's, on every go test: on the
// words of every vector, on the bounds and the edges of each domain, and on several thousand inputs
// that the generator itself draws from a constant seed (R10).
func TestReference(t *testing.T) {
	for _, in := range fixedInputs(t) {
		compare(t, in)
	}
	for _, in := range generatedInputs(t, 5000) {
		compare(t, in)
	}
}

// FuzzReference holds the generator to the reference on any input, its corpus the inputs of
// TestReference that do not depend on the generator: the words of the vectors with the bounds and
// the edges of each domain (R10; R33 spec).
func FuzzReference(f *testing.F) {
	for _, in := range fixedInputs(f) {
		f.Add(in.high, in.low, in.bound, in.from, in.to, in.count, in.seed, in.names)
	}
	f.Fuzz(func(t *testing.T, high, low, bound uint64, from, to int64, count int, seed uint64, names string) {
		compare(t, input{high, low, bound, from, to, count, seed, names})
	})
}
