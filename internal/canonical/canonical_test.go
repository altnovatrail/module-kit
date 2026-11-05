package canonical

import (
	"errors"
	"math"
	"testing"
)

// Each character of the table of section 25 of format/README.md is written as the table says: ",
// \ and five control characters by their short escapes, every other one up to U+001F as \u00 and
// two lowercase hexadecimal digits, and the rest as it is, /, <, >, &, U+007F, U+2028, U+2029, a
// letter outside the Basic Multilingual Plane and a combining accent among them.
func TestEscapes(t *testing.T) {
	for _, c := range []struct {
		in, want string
	}{
		{`"`, `"\""`},
		{`\`, `"\\"`},
		{"\b", `"\b"`},
		{"\t", `"\t"`},
		{"\n", `"\n"`},
		{"\f", `"\f"`},
		{"\r", `"\r"`},
		{"\x00", `"\u0000"`},
		{"\x07", `"\u0007"`},
		{"\x0b", `"\u000b"`},
		{"\x1f", `"\u001f"`},
		{"/", `"/"`},
		{"<>&", `"<>&"`},
		{"\x7f", "\"\x7f\""},
		{"\u2028\u2029", "\"\u2028\u2029\""},
		{"\U0001F600", "\"\U0001F600\""},
		{"e\u0301", "\"e\u0301\""},
		{"", `""`},
	} {
		got, err := Text(c.in)
		if err != nil || string(got) != c.want {
			t.Errorf("Text(%q) = %s, %v; want %s", c.in, got, err, c.want)
		}
	}
}

// Every character from U+0000 to U+001F has an escape, and every one gives the escape the table
// gives it.
func TestEveryControlCharacter(t *testing.T) {
	shortOnes := map[rune]bool{'\b': true, '\t': true, '\n': true, '\f': true, '\r': true}
	for r := rune(0); r < 0x20; r++ {
		got, err := Text(string(r))
		if err != nil {
			t.Fatal(err)
		}
		if shortOnes[r] {
			if len(got) != 4 {
				t.Errorf("U+%04X is written %s, and it has a short escape", r, got)
			}
			continue
		}
		want := `"\u00` + string(hexDigits[r>>4]) + string(hexDigits[r&0xf]) + `"`
		if string(got) != want {
			t.Errorf("U+%04X is written %s; want %s", r, got, want)
		}
	}
}

// A value of each kind is written in the canonical form: the members of an object in the order of
// their names' UTF-16 code units, at every depth, with no whitespace; an integer in decimal digits,
// with - before a negative one; and null, true and false.
func TestText(t *testing.T) {
	for _, c := range []struct {
		name string
		in   any
		want string
	}{
		{"null", nil, `null`},
		{"true", true, `true`},
		{"false", false, `false`},
		{"zero", int64(0), `0`},
		{"negative", int64(-12), `-12`},
		{"the largest integer", int64(MaxInteger), `9007199254740991`},
		{"the smallest integer", int64(-MaxInteger), `-9007199254740991`},
		{"an empty object", map[string]any{}, `{}`},
		{"an empty array", []any{}, `[]`},
		{"an array in its order", []any{int64(3), int64(1), int64(2)}, `[3,1,2]`},
		{"members by their names", map[string]any{"b": int64(1), "a": int64(2), "aa": nil, "B": true, "1": false, "_": "x"},
			`{"1":false,"B":true,"_":"x","a":2,"aa":null,"b":1}`},
		{"members within members", map[string]any{"z": map[string]any{"y": []any{map[string]any{"b": int64(1), "a": int64(0)}}}, "a": []any{}},
			`{"a":[],"z":{"y":[{"a":0,"b":1}]}}`},
		// U+1F600 is the surrogate pair D83D DE00, which comes before U+FB01 by code units, though
		// its code point is above it; by code points, or as UTF-8 bytes, the order would be the
		// other way round.
		{"by code units, not by code points", map[string]any{"\uFB01": int64(1), "\U0001F600": int64(2)},
			"{\"\U0001F600\":2,\"\uFB01\":1}"},
		{"an escape in a name", map[string]any{"a\"b": int64(1)}, `{"a\"b":1}`},
	} {
		got, err := Text(c.in)
		if err != nil || string(got) != c.want {
			t.Errorf("%s: Text = %s, %v; want %s", c.name, got, err, c.want)
		}
	}
}

// Text stops with an *Error at the first value the canonical text cannot hold, naming its place:
// a number with a fraction, an integer given as a double, an integer past ±(2^53 − 1), a string or
// a name that is not UTF-8, and a value of a Go type that is no JSON value (R09).
func TestRefusals(t *testing.T) {
	for _, c := range []struct {
		name    string
		in      any
		pointer string
	}{
		{"a fraction", map[string]any{"a": []any{int64(1), 1.5}}, "#/a/1"},
		{"an integer as a double", map[string]any{"width": 50.0}, "#/width"},
		{"an infinity", math.Inf(1), "#"},
		{"not a number", math.NaN(), "#"},
		{"2^53", int64(MaxInteger + 1), "#"},
		{"−2^53", []any{int64(-MaxInteger - 1)}, "#/0"},
		{"2^63 − 1", int64(math.MaxInt64), "#"},
		{"a string not UTF-8", map[string]any{"text": "a\xffb"}, "#/text"},
		{"a name not UTF-8", map[string]any{"a\xff": int64(1)}, "#/a%FF"},
		{"an int", int(1), "#"},
		{"a struct", struct{}{}, "#"},
		{"a typed map", map[string]string{"a": "b"}, "#"},
		{"a member with a slash", map[string]any{"pictures/dot.png": int32(1)}, "#/pictures~1dot.png"},
	} {
		got, err := Text(c.in)
		var e *Error
		if !errors.As(err, &e) || got != nil {
			t.Errorf("%s: Text = %s, %v; want an *Error", c.name, got, err)
			continue
		}
		if e.Pointer != c.pointer || e.Message == "" {
			t.Errorf("%s: the error is at %s, %q; want it at %s", c.name, e.Pointer, e.Message, c.pointer)
		}
		if e.Error() != "canonical: "+e.Pointer+": "+e.Message {
			t.Errorf("%s: Error() = %q", c.name, e.Error())
		}
	}
}

// A digest is sha256: and 64 lowercase hexadecimal digits, and a descriptor its digest and size.
func TestDigest(t *testing.T) {
	const empty = "sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
	if got := Digest(nil); got != empty {
		t.Errorf("Digest(nil) = %s; want %s", got, empty)
	}
	d := Descriptor([]byte("{}"))
	if d.Size != 2 || d.Digest != "sha256:44136fa355b3678a1146ad16f7e8649e94fb4fc21fe77e8310c060f61caaff8a" {
		t.Errorf("Descriptor({}) = %+v", d)
	}
}
