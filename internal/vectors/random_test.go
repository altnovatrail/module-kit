package vectors

import (
	"strings"
	"testing"
	"testing/fstest"

	"github.com/altnovatrail/spec/types/formatv0"

	"github.com/altnovatrail/module-kit/internal/random"
)

// The runner of the generator refuses a vector that any of its fields would make wrong: the words
// that the seed and the names derive, an integer that a word of 64 bits does not hold, an argument
// outside a function's domain, and the value of each function.
func TestRandomDrawsRefuse(t *testing.T) {
	const place = "seed: \"7\"\nnames: [a]\nstate: {high: \"3879510333719213901\", low: \"7\"}\ndraws:\n"
	const words = "state: {high: \"1\", low: \"2\"}\ndraws:\n"
	for _, c := range []struct {
		name, data, says string
	}{
		{"a vector", place + "  - integer: {value: \"152559422570776952\"}\n", ""},
		{"not of its schema", place, "random-vector.schema.json"},
		{"the words of another place", strings.Replace(place, "[a]", "[b]", 1) + "  - integer: {value: \"1\"}\n", "derive the words"},
		{"a seed past 2^64 − 1", strings.Replace(place, `"7"`, `"18446744073709551616"`, 1) + "  - integer: {value: \"1\"}\n", "out of range"},
		{"a high word past 2^64 − 1", strings.Replace(words, `"1"`, `"18446744073709551616"`, 1) + "  - integer: {value: \"1\"}\n", "out of range"},
		{"a low word past 2^64 − 1", strings.Replace(words, `"2"`, `"18446744073709551616"`, 1) + "  - integer: {value: \"1\"}\n", "out of range"},
		{"an integer", words + "  - integer: {value: \"1\"}\n", "draw 1 gives 14192431797130687760"},
		{"below", words + "  - below: {bound: \"3\", value: \"3\"}\n", "draw 1 gives"},
		{"a bound of 0", words + "  - below: {bound: \"0\", value: \"0\"}\n", "draw 1: random: below(0)"},
		{"a bound past 2^64 − 1", words + "  - below: {bound: \"18446744073709551616\", value: \"0\"}\n", "draw 1: strconv"},
		{"between", words + "  - between: {from: \"-1\", to: \"1\", value: \"2\"}\n", "draw 1 gives"},
		{"between backwards", words + "  - between: {from: \"1\", to: \"-1\", value: \"0\"}\n", "draw 1: random: between(1, -1)"},
		{"a least integer past −2^63", words + "  - between: {from: \"-9223372036854775809\", to: \"0\", value: \"0\"}\n", "draw 1: strconv"},
		{"a greatest integer past 2^63 − 1", words + "  - between: {from: \"0\", to: \"9223372036854775808\", value: \"0\"}\n", "draw 1: strconv"},
		{"a fraction", words + "  - fraction: {value: 0.5}\n", "draw 1 gives"},
		{"a permutation", words + "  - permutation: {count: 2, value: [2, 0]}\n", "draw 1 gives"},
		{"the second draw", words + "  - integer: {value: \"14192431797130687760\"}\n  - integer: {value: \"1\"}\n", "draw 2 gives"},
	} {
		err := randomDraws(vector{fsys: fstest.MapFS{"v.yaml": {Data: []byte(c.data)}}, path: "v.yaml"})
		if c.says == "" && err != nil || c.says != "" && (err == nil || !strings.Contains(err.Error(), c.says)) {
			t.Errorf("%s: %v, want %q", c.name, err, c.says)
		}
	}
	if err := randomDraws(vector{fsys: fstest.MapFS{}, path: "none.yaml"}); err == nil {
		t.Error("a file that is not there passes")
	}
	// The schema lets no draw through with no field, but the runner does not lean on it.
	if _, _, err := draw(random.New(1, 2), formatv0.Draw{}); err == nil || !strings.Contains(err.Error(), "names no function") {
		t.Errorf("a draw with no field: %v", err)
	}
}
