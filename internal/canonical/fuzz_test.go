//go:build goexperiment.jsonv2

package canonical

import (
	"encoding/json/jsontext"
	"testing"
	"unicode/utf8"
)

// On every value that the format admits, Text writes what jsontext's Canonicalize writes, RFC 8785
// in the standard library of Go (R09). The fuzz test builds a value from its input: a stream of
// bytes read as a small program, so that every input is a value of the format — objects, arrays,
// strings of valid UTF-8, integers within ±(2^53 − 1), true, false and null. Canonicalize is the
// reference here alone: under GOEXPERIMENT=nojsonv2 this file is not built, and the other tests run
// without it.
func FuzzText(f *testing.F) {
	for _, seed := range []string{
		"", "o", "a", "s\"\\\b\t\n\f\r\x07", "i\x01\x02\x03\x04\x05\x06\x07", "t", "f", "n",
		"oa\x00k\U0001F600\uFB01", "s<>&\u2028\u2029e\u0301/",
	} {
		f.Add([]byte(seed))
	}
	f.Fuzz(func(t *testing.T, data []byte) {
		value := build(&data, 0)
		got, err := Text(value)
		if err != nil {
			t.Fatalf("Text refuses %#v: %v", value, err)
		}
		want := jsontext.Value(got).Clone()
		if err := want.Canonicalize(); err != nil {
			t.Fatalf("Canonicalize refuses %s: %v", got, err)
		}
		if string(got) != string(want) {
			t.Fatalf("Text writes\n%s\nand Canonicalize\n%s", got, want)
		}
	})
}

// The depth below which build nests no object or array, so that a value stays small.
const maxDepth = 4

// build reads one value from the front of data, and takes what it read off data. Its first byte
// says what the value is; a value ends when data does.
func build(data *[]byte, depth int) any {
	if len(*data) == 0 {
		return nil
	}
	kind := (*data)[0]
	*data = (*data)[1:]
	switch kind % 8 {
	case 0:
		if depth > maxDepth {
			return nil
		}
		object := map[string]any{}
		for n := next(data) % 6; n > 0; n-- {
			object[text(data)] = build(data, depth+1)
		}
		return object
	case 1:
		if depth > maxDepth {
			return nil
		}
		var array []any
		for n := next(data) % 6; n > 0; n-- {
			array = append(array, build(data, depth+1))
		}
		return append([]any{}, array...)
	case 2:
		return text(data)
	case 3:
		var n int64
		for range 7 {
			n = n<<8 | int64(next(data))
		}
		n %= MaxInteger + 1
		if next(data)%2 == 1 {
			n = -n
		}
		return n
	case 4:
		return true
	case 5:
		return false
	case 6:
		return int64(next(data)) - 128
	}
	return nil
}

// next takes one byte off data, or 0 when it is empty.
func next(data *[]byte) byte {
	if len(*data) == 0 {
		return 0
	}
	b := (*data)[0]
	*data = (*data)[1:]
	return b
}

// text takes a string off data: as many bytes as its first byte says, at most 15, with each byte
// that is not part of a UTF-8 character replaced, so that the string is valid UTF-8.
func text(data *[]byte) string {
	n := min(int(next(data)%16), len(*data))
	s := string((*data)[:n])
	*data = (*data)[n:]
	if !utf8.ValidString(s) {
		s = string([]rune(s))
	}
	return s
}
