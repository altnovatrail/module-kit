package semver

import (
	"math/big"
	"regexp"
	"strings"
	"testing"

	"github.com/altnovatrail/module-kit/internal/schema"
)

// Parse reads what the form of a version in the format accepts, and nothing else.
func TestParse(t *testing.T) {
	for _, s := range []string{"0.0.0", "1.2.3", "10.0.1", "0.1.0", "123456789012345678901234567890.0.7"} {
		v, err := Parse(s)
		if err != nil {
			t.Errorf("Parse(%q): %v", s, err)
		} else if v.String() != s {
			t.Errorf("Parse(%q) writes %q", s, v)
		}
	}
	for _, s := range []string{"", "1", "1.2", "1.2.3.4", "01.2.3", "1.02.3", "1.2.03", "1.2.3-rc.1", "1.2.3+build", "v1.2.3",
		"1..3", ".1.2", "1.2.", "1.2.-3", "1.2.+3", " 1.2.3", "1.2.3 ", "1.2.٣", "a.b.c"} {
		if _, err := Parse(s); err == nil {
			t.Errorf("Parse(%q) reads a version", s)
		}
	}
}

// Parse accepts exactly what common.schema.json#/$defs/module_version accepts.
func TestParseIsTheSchemas(t *testing.T) {
	for _, s := range []string{"0.0.0", "1.2.3", "01.2.3", "1.2", "1.2.3-rc.1", "9.9.99999999999999999999", "1.2.٣", "1.2.3\n"} {
		_, err := Parse(s)
		valid := len(schema.Validate("common.schema.json#/$defs/module_version", s)) == 0
		if (err == nil) != valid {
			t.Errorf("Parse(%q) gives %v, and the schema accepts it: %t", s, err, valid)
		}
	}
}

// Compare orders versions by their numbers, as Semantic Versioning does, and not by their text.
func TestCompare(t *testing.T) {
	ordered := []string{"0.0.0", "0.0.1", "0.0.9", "0.0.10", "0.1.0", "0.2.0", "0.10.0", "1.0.0", "1.0.1", "1.9.0", "1.10.0", "2.0.0",
		"10.0.0", "18446744073709551616.0.0"}
	for i, a := range ordered {
		for j, b := range ordered {
			want := 0
			switch {
			case i < j:
				want = -1
			case i > j:
				want = 1
			}
			if got := mustParse(t, a).Compare(mustParse(t, b)); got != want {
				t.Errorf("Compare(%s, %s) = %d, want %d", a, b, got, want)
			}
		}
	}
}

// Compare is the order of the values of the numbers, which math/big reads, on versions of every
// length.
func FuzzCompare(f *testing.F) {
	f.Add("1.10.0", "1.9.0")
	f.Add("0.0.0", "0.0.0")
	f.Add("99999999999999999999.1.2", "100000000000000000000.0.0")
	form := regexp.MustCompile(`^(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$`)
	f.Fuzz(func(t *testing.T, a, b string) {
		v, errV := Parse(a)
		w, errW := Parse(b)
		if (errV == nil) != form.MatchString(a) || (errW == nil) != form.MatchString(b) {
			t.Fatalf("Parse(%q) gives %v and Parse(%q) %v, against the form of a version", a, errV, b, errW)
		}
		if errV != nil || errW != nil {
			return
		}
		want := 0
		x, y := strings.Split(a, "."), strings.Split(b, ".")
		for i := range x {
			m, _ := new(big.Int).SetString(x[i], 10)
			n, _ := new(big.Int).SetString(y[i], 10)
			if want = m.Cmp(n); want != 0 {
				break
			}
		}
		if got := v.Compare(w); got != want {
			t.Fatalf("Compare(%s, %s) = %d, and the values of their numbers give %d", a, b, got, want)
		}
	})
}

func mustParse(t *testing.T, s string) Version {
	t.Helper()
	v, err := Parse(s)
	if err != nil {
		t.Fatal(err)
	}
	return v
}

// Raise raises one number by one, with a carry of any length, and sets every lower one to 0.
func TestRaise(t *testing.T) {
	for _, c := range []struct {
		from  string
		place int
		want  string
	}{
		{"1.2.3", Patch, "1.2.4"}, {"1.2.3", Minor, "1.3.0"}, {"1.2.3", Major, "2.0.0"},
		{"0.0.9", Patch, "0.0.10"}, {"0.9.9", Minor, "0.10.0"}, {"9.9.9", Major, "10.0.0"},
		{"1.0.99999999999999999999", Patch, "1.0.100000000000000000000"}, {"18446744073709551615.0.0", Major, "18446744073709551616.0.0"},
	} {
		if got := mustParse(t, c.from).Raise(c.place).String(); got != c.want {
			t.Errorf("Raise(%s, %d) = %s, want %s", c.from, c.place, got, c.want)
		}
	}
	if v := mustParse(t, "0.3.0"); !v.Zero(Major) || v.Zero(Minor) || !v.Zero(Patch) {
		t.Errorf("Zero of 0.3.0 is %t, %t, %t", v.Zero(Major), v.Zero(Minor), v.Zero(Patch))
	}
}
