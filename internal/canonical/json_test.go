package canonical

import (
	"fmt"
	"math"
	"reflect"
	"testing"
)

// pastInt64 is 2^63, the smallest integer that does not fit an int64.
const pastInt64 = uint64(math.MaxInt64) + 1

// Parse reads 1.0 and 1e2 as float64s, as the reader of YAML does, so that the canonical text,
// which never writes them, refuses them. An integer is an int64 down to -MaxInteger, the smallest
// of the format, and a float64 past the range of an int64.
func TestParse(t *testing.T) {
	text := fmt.Sprintf(`{"a":1,"b":1.0,"c":1e2,"d":%d,"e":[true,null,"x"],"f":{"g":%d}}`, -MaxInteger, pastInt64)
	got, err := Parse([]byte(text))
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]any{"a": int64(1), "b": 1.0, "c": 100.0, "d": int64(-MaxInteger), "e": []any{true, nil, "x"},
		"f": map[string]any{"g": float64(pastInt64)}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Parse = %#v; want %#v", got, want)
	}
	for _, text := range []string{`{} {}`, `{}x`, `{"a":}`, ``, `nul`, "{\"a\":\"wrong\xffkey\"}", `1e400`, `{"a":[-1e400]}`, `{"a":{"b":2e308}}`} {
		if _, err := Parse([]byte(text)); err == nil {
			t.Errorf("Parse(%q) reads a JSON text", text)
		}
	}
}
