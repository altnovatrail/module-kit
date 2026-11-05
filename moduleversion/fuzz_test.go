package moduleversion

import (
	"reflect"
	"strconv"
	"testing"

	"github.com/altnovatrail/module-kit/internal/canonical"
	"github.com/altnovatrail/module-kit/modulepackage"
)

// The differences of two JSON values are none exactly when the values are equal, and each is at a
// place that one of the two values has.
func FuzzDiff(f *testing.F) {
	for _, seed := range [][2]string{
		{`{}`, `{}`}, {`{"a":1}`, `{"a":2}`}, {`{"a":[1,2]}`, `{"a":[1]}`}, {`[1,{"b":"x"}]`, `[1,{"b":"y","c":null}]`},
		{`{"license":{"digest":"sha256:0","size":1}}`, `{"license":{"digest":"sha256:0","size":1},"media":{"a/b.png":{"size":2}}}`},
		{`{"a~b":true}`, `{"a~b":false}`}, {`"x"`, `{"x":1}`},
	} {
		f.Add(seed[0], seed[1])
	}
	f.Fuzz(func(t *testing.T, x, y string) {
		a, err := canonical.Parse([]byte(x))
		if err != nil {
			return
		}
		b, err := canonical.Parse([]byte(y))
		if err != nil {
			return
		}
		found := diff(modulepackage.DeviceFile, nil, a, b)
		if (len(found) == 0) != reflect.DeepEqual(a, b) {
			t.Fatalf("%s and %s give %d differences", x, y, len(found))
		}
		for _, d := range found {
			if !has(a, d.tokens) && !has(b, d.tokens) {
				t.Fatalf("%s and %s differ at %s, which neither has", x, y, d.Pointer)
			}
		}
	})
}

// has says whether a JSON value has a place, token by token.
func has(value any, tokens []string) bool {
	for _, token := range tokens {
		switch v := value.(type) {
		case map[string]any:
			next, ok := v[token]
			if !ok {
				return false
			}
			value = next
		case []any:
			i, err := strconv.Atoi(token)
			if err != nil || i < 0 || i >= len(v) {
				return false
			}
			value = v[i]
		default:
			return false
		}
	}
	return true
}
