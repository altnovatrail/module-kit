package canonical

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"unicode/utf8"
)

// Parse reads a JSON text into the tree of JSON values that Text takes, as the reader of YAML gives
// one: a number written with neither a fraction nor an exponent as an int64, when it fits one, and
// any other number as a float64, which Text then refuses. A text with anything after its value is
// no JSON text, and neither is one that is not UTF-8, which encoding/json would read with U+FFFD in
// the place of each bad byte (RFC 8259, section 8.1), nor one with a number past the range of a
// double. Package modulepackage reads the texts of a package by it, and package moduleindex the
// module index (R34, R35).
func Parse(data []byte) (any, error) {
	if !utf8.Valid(data) {
		return nil, errors.New("not UTF-8")
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	var value any
	if err := decoder.Decode(&value); err != nil {
		return nil, err
	}
	if _, err := decoder.Token(); !errors.Is(err, io.EOF) {
		return nil, errors.New("more after its value")
	}
	return numbers(value)
}

// numbers puts an int64, or else a float64, in the place of each json.Number in a value. A number
// past the range of a float64, such as 1e400, is no number a reader can hold, and an error: as the
// float64 ±Inf it would stop the validator of the schemas.
func numbers(value any) (any, error) {
	var err error
	switch v := value.(type) {
	case map[string]any:
		for key, member := range v {
			if v[key], err = numbers(member); err != nil {
				return nil, err
			}
		}
	case []any:
		for i, element := range v {
			if v[i], err = numbers(element); err != nil {
				return nil, err
			}
		}
	case json.Number:
		if i, err := v.Int64(); err == nil && !strings.ContainsAny(string(v), ".eE") {
			return i, nil
		}
		f, err := v.Float64()
		if err != nil {
			return nil, fmt.Errorf("the number %s is past the range of a double", v)
		}
		return f, nil
	}
	return value, nil
}
