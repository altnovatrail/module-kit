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
