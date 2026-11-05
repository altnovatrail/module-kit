package modulepackage

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"strings"
)

// parse reads a JSON text into a JSON value as the reader of YAML gives one: a number written with
// neither a fraction nor an exponent as an int64, when it fits one, and any other number as a
// float64, which the canonical text then refuses. A text with anything after its value is no JSON
// text.
func parse(data []byte) (any, error) {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	var value any
	if err := decoder.Decode(&value); err != nil {
		return nil, err
	}
	if _, err := decoder.Token(); !errors.Is(err, io.EOF) {
		return nil, errors.New("more after its value")
	}
	return numbers(value), nil
}

// numbers puts an int64, or else a float64, in the place of each json.Number in a value.
func numbers(value any) any {
	switch v := value.(type) {
	case map[string]any:
		for key, member := range v {
			v[key] = numbers(member)
		}
	case []any:
		for i, element := range v {
			v[i] = numbers(element)
		}
	case json.Number:
		if i, err := v.Int64(); err == nil && !strings.ContainsAny(string(v), ".eE") {
			return i
		}
		f, _ := v.Float64()
		return f
	}
	return value
}
