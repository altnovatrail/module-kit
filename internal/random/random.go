package random

import (
	"errors"
	"fmt"
	"math"
	"math/bits"
	"strconv"
	"strings"
	"unicode/utf8"
)

// ErrDomain is what a function gives on an argument outside its domain, and a place on a name that
// is none: the function draws no integer, and no place is built (R10).
var ErrDomain = errors.New("outside the domain")

// The constants of the generator, as the format's text gives them (format/README.md, section 22):
// the multiplier M and the increment C of a step, each a high and a low word, and the multiplier of
// DXSM.
const (
	multiplierHigh = 2549297995355413924  // 0x2360ed051fc65da4
	multiplierLow  = 4865540595714422341  // 0x4385df649fccf645
	incrementHigh  = 6364136223846793005  // 0x5851f42d4c957f2d
	incrementLow   = 1442695040888963407  // 0x14057b7ef767814f
	dxsmMultiplier = 15750249268501108917 // 0xda942042e4dd58b5
)

// The constants of FNV-1a of 64 bits (RFC 9923, sections 2 and 5): where it starts, and the prime
// it multiplies by.
const (
	fnvOffset = 14695981039346656037 // 0xcbf29ce484222325
	fnvPrime  = 1099511628211        // 0x100000001b3
)

// Stream is the generator, PCG with 128 bits of state, its state the number high · 2^64 + low. It
// gives numbers only through its five functions, each drawing as many integers as the format's text
// says. A Stream is not for two goroutines at once: each use builds its own, and lives in one call.
type Stream struct {
	high, low uint64
}

// New gives the stream that starts from two words, taken as its state as they are.
func New(high, low uint64) *Stream {
	return &Stream{high, low}
}

// Words gives the two words that start the stream of a place of randomness: the high one FNV-1a of
// 64 bits of the UTF-8 text of the seed in decimal digits and then, for each name in order, | and
// the name; the low one the seed itself. A place has a name at least, and a name is UTF-8 that is
// not empty and holds no |; names compare as written, with no normalization. Any other list of
// names gives ErrDomain.
func Words(seed uint64, names []string) (high, low uint64, err error) {
	if len(names) == 0 {
		return 0, 0, fmt.Errorf("random: a place has a name at least: %w", ErrDomain)
	}
	for i, name := range names {
		if name == "" || strings.Contains(name, "|") || !utf8.ValidString(name) {
			return 0, 0, fmt.Errorf("random: name %d, %q, is no name of a place, UTF-8 that is not empty and holds no |: %w", i+1, name, ErrDomain)
		}
	}
	hash := uint64(fnvOffset)
	add := func(text string) {
		for i := range len(text) {
			hash ^= uint64(text[i])
			hash *= fnvPrime
		}
	}
	add(strconv.FormatUint(seed, 10))
	for _, name := range names {
		add("|")
		add(name)
	}
	return hash, seed, nil
}

// Place gives the stream of the place of randomness of a seed and its names, as Words starts it.
func Place(seed uint64, names ...string) (*Stream, error) {
	high, low, err := Words(seed, names)
	if err != nil {
		return nil, err
	}
	return New(high, low), nil
}

// Integer gives the next integer of the stream, from 0 to 2^64 − 1, and draws one: a step sets the
// state to state · M + C modulo 2^128, and DXSM draws the integer from the new state's words.
func (s *Stream) Integer() uint64 {
	high, low := bits.Mul64(s.low, multiplierLow)
	high += s.high*multiplierLow + s.low*multiplierHigh
	low, carry := bits.Add64(low, incrementLow, 0)
	high, _ = bits.Add64(high, incrementHigh, carry)
	s.high, s.low = high, low
	high ^= high >> 32
	high *= dxsmMultiplier
	high ^= high >> 48
	return high * (low | 1)
}

// Below gives an integer from 0 to b − 1, for a bound b from 1 to 2^64 − 1. A bound that is a power
// of two masks the next integer, which draws one; any other takes the high word of the product of
// the next integer and b, by Lemire's method, and takes the product again with the next integer
// while its low word is less than (2^64 − b) mod b, each product drawing one.
func (s *Stream) Below(b uint64) (uint64, error) {
	if b == 0 {
		return 0, fmt.Errorf("random: below(0): the bound is from 1 to 2^64 − 1: %w", ErrDomain)
	}
	return s.below(b), nil
}

// below is Below on a bound already in its domain.
func (s *Stream) below(b uint64) uint64 {
	if b&(b-1) == 0 {
		return s.Integer() & (b - 1)
	}
	high, low := bits.Mul64(s.Integer(), b)
	if low < b {
		// (2^64 − b) mod b, in the words of 64 bits that wrap: -b is 2^64 − b.
		threshold := -b % b
		for low < threshold {
			high, low = bits.Mul64(s.Integer(), b)
		}
	}
	return high
}

// Between gives a + below(c − a + 1), an integer from a to c, both included, for signed integers a
// ≤ c with c − a + 1 at most 2^63 − 1. It checks a ≤ c first, so that c − a, taken in the words of
// 64 bits that wrap, is the count less one and overflows nothing.
func (s *Stream) Between(a, c int64) (int64, error) {
	if a > c {
		return 0, fmt.Errorf("random: between(%d, %d): the first is greater than the second: %w", a, c, ErrDomain)
	}
	span := uint64(c) - uint64(a)
	if span > math.MaxInt64-1 {
		return 0, fmt.Errorf("random: between(%d, %d): it counts more than 2^63 − 1 integers: %w", a, c, ErrDomain)
	}
	return int64(uint64(a) + s.below(span+1)), nil
}

// Fraction gives a number from 0 up to 1, never 1: the low 53 bits of the next integer, divided by
// 2^53, which a double holds exactly. It draws one integer.
func (s *Stream) Fraction() float64 {
	return float64(s.Integer()&(1<<53-1)) / (1 << 53)
}

// MaxPermutation is the greatest count that Permutation orders, 2^31 − 1.
const MaxPermutation = math.MaxInt32

// Permutation gives an order of the integers from 0 to n − 1, for n from 0 to 2^31 − 1: from them
// in order, for i from n − 1 down to 1, it swaps the items at places i and below(i + 1). It draws
// nothing when n is 0 or 1.
func (s *Stream) Permutation(n int) ([]int, error) {
	if n < 0 || n > MaxPermutation {
		return nil, fmt.Errorf("random: permutation(%d): the count is from 0 to 2^31 − 1: %w", n, ErrDomain)
	}
	order := make([]int, n)
	for i := range order {
		order[i] = i
	}
	for i := n - 1; i > 0; i-- {
		j := s.below(uint64(i) + 1)
		order[i], order[j] = order[j], order[i]
	}
	return order, nil
}
