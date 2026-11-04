package yaml

import (
	"errors"
	"fmt"
	"io/fs"
	"maps"
	"math"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/altnovatrail/spec"
)

// On any input the reader gives a value or a refusal, and does not panic, nor lets go-yaml panic.
// A refusal names a rule yaml.* or that the file is not YAML, at a place in it. A value is one the
// format allows, and its JSON text reads back to it: a file in JSON syntax is YAML, under the same
// rules (format/README.md, section 4). The corpus is the vectors of the rules yaml.*, the YAML test
// suite and the cases of TestRead.
func FuzzRead(f *testing.F) {
	vectors, err := fs.Glob(spec.Format, "vectors/rules/yaml.*/*.yaml")
	if err != nil || len(vectors) == 0 {
		f.Fatalf("the vectors of the rules yaml.*: %v", err)
	}
	for _, name := range vectors {
		data, err := fs.ReadFile(spec.Format, name)
		if err != nil {
			f.Fatal(err)
		}
		f.Add(data)
	}
	cases, err := suiteCases()
	if err != nil {
		f.Fatal(err)
	}
	for _, c := range cases {
		f.Add([]byte(c.YAML))
	}
	for _, test := range readCases {
		f.Add([]byte(test.in))
	}
	f.Fuzz(func(t *testing.T, data []byte) {
		value, err := Read(data)
		if err != nil {
			var refusal *Refusal
			if !errors.As(err, &refusal) || len(refusal.Breaches) == 0 {
				t.Fatalf("an error that is no refusal: %v", err)
			}
			if value != nil {
				t.Fatalf("refused with the value %#v", value)
			}
			for _, b := range refusal.Breaches {
				if b.Rule != "" && !slices.Contains(Rules, b.Rule) || b.Line < 1 || b.Column < 1 ||
					strings.HasPrefix(b.Message, panicMessage) {
					t.Fatalf("the breach %v", b)
				}
			}
			return
		}
		if problem := outsideFormat(value); problem != "" {
			t.Fatalf("read as %#v, %s", value, problem)
		}
		var text strings.Builder
		if !writeJSON(&text, value) {
			return
		}
		again, err := Read([]byte(text.String()))
		if err != nil || !reflect.DeepEqual(again, value) {
			t.Fatalf("read as %#v, whose JSON text\n%s\nreads as %#v, %v", value, text.String(), again, err)
		}
	})
}

// outsideFormat says what of a value the format does not allow, or "".
func outsideFormat(value any) string {
	switch v := value.(type) {
	case nil, bool:
	case int64:
		if v > maxSafeInteger || v < -maxSafeInteger {
			return fmt.Sprintf("the integer %d outside ±(2^53−1)", v)
		}
	case float64:
		if math.IsInf(v, 0) || math.IsNaN(v) || v == 0 && math.Signbit(v) {
			return fmt.Sprintf("the double %v", v)
		}
	case string:
		return outsideStrings(v)
	case []any:
		for _, item := range v {
			if problem := outsideFormat(item); problem != "" {
				return problem
			}
		}
	case map[string]any:
		for key, member := range v {
			if problem := outsideStrings(key); problem != "" {
				return problem
			}
			if problem := outsideFormat(member); problem != "" {
				return problem
			}
		}
	default:
		return fmt.Sprintf("a value of the type %T", v)
	}
	return ""
}

func outsideStrings(s string) string {
	if !utf8.ValidString(s) {
		return fmt.Sprintf("the string %q, not UTF-8", s)
	}
	for _, c := range s {
		if noncharacter(c) {
			return fmt.Sprintf("the string %q, with a noncharacter", s)
		}
	}
	return ""
}

// The longest key writeJSON writes: an implicit key of a flow mapping ends within 1024 characters,
// and the reader reads a longer one only as an explicit key, which JSON has not.
const longestKey = implicitKeyLimit - 4

// writeJSON writes a value as a JSON text that the reader reads back to it, and says whether it
// could. A double keeps a point or an exponent, so that it is read as a double again, and every
// character that YAML does not print, or prints only in quotes, is escaped: json.Marshal does
// neither (R21).
func writeJSON(b *strings.Builder, value any) bool {
	switch v := value.(type) {
	case nil:
		b.WriteString("null")
	case bool:
		b.WriteString(strconv.FormatBool(v))
	case int64:
		b.WriteString(strconv.FormatInt(v, 10))
	case float64:
		s := strconv.FormatFloat(v, 'g', -1, 64)
		if !strings.ContainsAny(s, ".eE") {
			s += ".0"
		}
		b.WriteString(s)
	case string:
		writeString(b, v)
	case []any:
		b.WriteByte('[')
		for i, item := range v {
			if i > 0 {
				b.WriteString(", ")
			}
			if !writeJSON(b, item) {
				return false
			}
		}
		b.WriteByte(']')
	case map[string]any:
		b.WriteByte('{')
		for i, key := range slices.Sorted(maps.Keys(v)) {
			if i > 0 {
				b.WriteString(", ")
			}
			start := b.Len()
			writeString(b, key)
			if utf8.RuneCountInString(b.String()[start:]) > longestKey {
				return false
			}
			b.WriteString(": ")
			if !writeJSON(b, v[key]) {
				return false
			}
		}
		b.WriteByte('}')
	}
	return true
}

func writeString(b *strings.Builder, s string) {
	b.WriteByte('"')
	for _, c := range s {
		switch {
		case c == '"' || c == '\\':
			b.WriteByte('\\')
			b.WriteRune(c)
		case c < 0x20 || !printable(c) || c == 0xfeff || c == 0x85 || c == 0x2028 || c == 0x2029:
			fmt.Fprintf(b, `\u%04x`, c)
		default:
			b.WriteRune(c)
		}
	}
	b.WriteByte('"')
}

// The writer of JSON texts escapes what YAML prints only in quotes, and keeps a double a double.
func TestWriteJSON(t *testing.T) {
	var b strings.Builder
	writeJSON(&b, map[string]any{"b": []any{int64(1), 1e20, 50.0, 2.5}, "a": "x\"\\\u007f\ufeff\u2028\n", "c": nil, "d": true})
	want := `{"a": "x\"\\\u007f\ufeff\u2028\u000a", "b": [1, 1e+20, 50.0, 2.5], "c": null, "d": true}`
	if b.String() != want {
		t.Errorf("%s, not %s", b.String(), want)
	}
	if writeJSON(&b, map[string]any{strings.Repeat("k", longestKey): nil}) {
		t.Error("a key too long for an implicit key written")
	}
}
