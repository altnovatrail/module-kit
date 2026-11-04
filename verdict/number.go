package verdict

import (
	"math/big"
	"regexp"
	"strings"

	"github.com/altnovatrail/spec/types/formatv0"
)

// The notation of the numbers that number, fraction and predicate take (format/README.md, section
// 15): an integer, a decimal with a point and a simple fraction, of at most 16 characters.
var (
	integerNotation  = regexp.MustCompile(`^-?(0|[1-9][0-9]*)$`)
	decimalNotation  = regexp.MustCompile(`^-?(0|[1-9][0-9]*)\.[0-9]+$`)
	fractionNotation = regexp.MustCompile(`^-?(0|[1-9][0-9]*) ?/ ?[1-9][0-9]*$`)
)

// maxNotation is the length of the longest text in notation. A text in notation is ASCII, so its
// length in bytes is its length in characters.
const maxNotation = 16

// written is a number in notation: its value, exactly, and whether it is written in lowest terms, as
// a simple fraction whose numerator and denominator have no common divisor but 1 is, and an integer
// or a decimal always is.
type written struct {
	value  *big.Rat
	lowest bool
}

var one = big.NewInt(1)

// notation reads a text in the notation that a kind takes — number all three, fraction a simple
// fraction, predicate an integer — and says whether it is in it. big.Rat reduces a fraction, so
// whether it is in lowest terms is read off the numerator and the denominator as they are written.
func notation(text string, kind formatv0.CheckKind) (written, bool) {
	if len(text) > maxNotation {
		return written{}, false
	}
	switch {
	case integerNotation.MatchString(text):
		if kind == formatv0.CheckKindFraction {
			return written{}, false
		}
		n, _ := new(big.Int).SetString(text, 10)
		return written{new(big.Rat).SetInt(n), true}, true
	case decimalNotation.MatchString(text):
		if kind != formatv0.CheckKindNumber {
			return written{}, false
		}
		r, _ := new(big.Rat).SetString(text)
		return written{r, true}, true
	case fractionNotation.MatchString(text):
		if kind == formatv0.CheckKindPredicate {
			return written{}, false
		}
		numerator, denominator, _ := strings.Cut(text, "/")
		n, _ := new(big.Int).SetString(strings.Trim(numerator, " "), 10)
		d, _ := new(big.Int).SetString(strings.Trim(denominator, " "), 10)
		divisor := new(big.Int).GCD(nil, nil, new(big.Int).Abs(n), d)
		return written{new(big.Rat).SetFrac(n, d), divisor.Cmp(one) == 0}, true
	}
	return written{}, false
}
