package semver

import (
	"cmp"
	"fmt"
	"strings"
)

// A Version is a version of a module, X.Y.Z: its major, minor and patch numbers, each a decimal
// numeral with no leading zero. Its zero value is no version.
type Version struct {
	numbers [3]string
}

// Parse reads a version of a module, X.Y.Z, in the form of common.schema.json#/$defs/module_version:
// three decimal numerals with no leading zero, joined by ".", and nothing else.
func Parse(s string) (Version, error) {
	var v Version
	numbers := strings.Split(s, ".")
	if len(numbers) != len(v.numbers) {
		return Version{}, fmt.Errorf("semver: %q is no version X.Y.Z", s)
	}
	for i, n := range numbers {
		if !numeral(n) {
			return Version{}, fmt.Errorf("semver: %q is no version X.Y.Z: %q is no decimal numeral with no leading zero", s, n)
		}
		v.numbers[i] = n
	}
	return v, nil
}

// numeral reports whether s is a decimal numeral with no leading zero: 0, or a digit 1 to 9 and
// more digits.
func numeral(s string) bool {
	if s == "" || s[0] == '0' && len(s) > 1 {
		return false
	}
	for _, c := range []byte(s) {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

// Compare gives -1, 0 or +1 as v stands below, at or above w by Semantic Versioning, section 11: by
// the major number, then the minor, then the patch, each by its value and not its text, so that
// 1.10.0 stands above 1.9.0.
func (v Version) Compare(w Version) int {
	for i := range v.numbers {
		// A numeral has no leading zero, so the longer one is the larger, and of two as long the
		// order of their digits is the order of their values.
		if c := cmp.Or(cmp.Compare(len(v.numbers[i]), len(w.numbers[i])), strings.Compare(v.numbers[i], w.numbers[i])); c != 0 {
			return c
		}
	}
	return 0
}

// String writes the version as X.Y.Z.
func (v Version) String() string {
	return strings.Join(v.numbers[:], ".")
}

// The places of the numbers of a version, from the highest, as Raise and Zero take them.
const (
	Major = iota
	Minor
	Patch
)

// Raise gives the version with the number at place raised by one and every lower number 0, as a
// step to the next version by that part does (format/README.md, section 27): 1.2.3 raised at Minor
// is 1.3.0. The number has no bound: 9 is raised to 10 in its text, digit by digit.
func (v Version) Raise(place int) Version {
	w := v
	w.numbers[place] = increment(v.numbers[place])
	for i := place + 1; i < len(w.numbers); i++ {
		w.numbers[i] = "0"
	}
	return w
}

// Zero says whether the number at place is 0.
func (v Version) Zero(place int) bool {
	return v.numbers[place] == "0"
}

// increment adds one to a decimal numeral with no leading zero: the last digit that is not 9 goes
// up, the 9s after it become 0s, and a numeral of 9s alone gets a 1 before it.
func increment(n string) string {
	digits := []byte(n)
	for i := len(digits) - 1; i >= 0; i-- {
		if digits[i] != '9' {
			digits[i]++
			return string(digits)
		}
		digits[i] = '0'
	}
	return "1" + string(digits)
}
