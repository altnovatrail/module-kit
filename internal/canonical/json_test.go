package canonical

import (
	"reflect"
	"testing"
)

// Parse reads an integer as an int64, and any other number as a float64, as the reader of YAML
// does: so the canonical text refuses 1.0 and 1e2, which it never writes. A text with more after
// its value, a text not in UTF-8, one with a number past the range of a double and no JSON at all
// are no JSON text that Parse reads.
func TestParse(t *testing.T) {
	got, err := Parse([]byte(`{"a":1,"b":1.0,"c":1e2,"d":-9007199254740991,"e":[true,null,"x"],"f":{"g":9223372036854775808}}`))
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]any{"a": int64(1), "b": 1.0, "c": 100.0, "d": int64(-9007199254740991), "e": []any{true, nil, "x"},
		"f": map[string]any{"g": 9223372036854775808.0}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Parse = %#v; want %#v", got, want)
	}
	for _, text := range []string{`{} {}`, `{}x`, `{"a":}`, ``, `nul`, "{\"a\":\"wrong\xffkey\"}", `1e400`, `{"a":[-1e400]}`, `{"a":{"b":2e308}}`} {
		if _, err := Parse([]byte(text)); err == nil {
			t.Errorf("Parse(%q) reads a JSON text", text)
		}
	}
}
