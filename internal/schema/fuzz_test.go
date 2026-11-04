package schema

import (
	"io/fs"
	"slices"
	"strconv"
	"testing"

	"github.com/altnovatrail/spec"

	"github.com/altnovatrail/module-kit/internal/yaml"
)

// holds says whether value has a place at location.
func holds(value any, location []string) bool {
	for _, token := range location {
		switch v := value.(type) {
		case map[string]any:
			member, ok := v[token]
			if !ok {
				return false
			}
			value = member
		case []any:
			i, err := strconv.Atoi(token)
			if err != nil || i < 0 || i >= len(v) || strconv.Itoa(i) != token {
				return false
			}
			value = v[i]
		default:
			return false
		}
	}
	return true
}

// within says whether view is what the open view of value may be: the same scalar, an array of as
// many elements, each within its own, or an object whose every field is one of value's and within
// it.
func within(view, value any) bool {
	switch v := view.(type) {
	case map[string]any:
		object, ok := value.(map[string]any)
		if !ok {
			return false
		}
		for key, field := range v {
			member, ok := object[key]
			if !ok || !within(field, member) {
				return false
			}
		}
		return true
	case []any:
		array, ok := value.([]any)
		if !ok || len(array) != len(v) {
			return false
		}
		for i, element := range v {
			if !within(element, array[i]) {
				return false
			}
		}
		return true
	}
	return view == value
}

// On any value that the reader of YAML gives, the open view at the root of every schema of the
// format is built without a panic, and is within the value. The corpus is the valid examples of
// the parts and the parts of the vectors of the open view.
func FuzzOpen(f *testing.F) {
	for _, pattern := range []string{"examples/*/valid/*.yaml", "vectors/open-view/*/*/part.yaml"} {
		names, err := fs.Glob(spec.Format, pattern)
		if err != nil {
			f.Fatal(err)
		}
		for _, name := range names {
			data, err := fs.ReadFile(spec.Format, name)
			if err != nil {
				f.Fatal(err)
			}
			f.Add(data)
		}
	}
	f.Fuzz(func(t *testing.T, data []byte) {
		value, err := yaml.Read(data)
		if err != nil {
			return
		}
		for _, name := range Files() {
			if view := Open(name, value); !within(view, value) {
				t.Errorf("%s: %v is not within %v", name, view, value)
			}
		}
	})
}

// On any value that the reader of YAML gives, every schema of the format gives its errors without a
// panic, each naming a keyword of section 5 at a place the value has, with a message. The corpus is
// the examples of the format.
func FuzzValidate(f *testing.F) {
	names, err := fs.Glob(spec.Format, "examples/*/*/*.yaml")
	if err != nil {
		f.Fatal(err)
	}
	for _, name := range names {
		data, err := fs.ReadFile(spec.Format, name)
		if err != nil {
			f.Fatal(err)
		}
		f.Add(data)
	}
	f.Fuzz(func(t *testing.T, data []byte) {
		value, err := yaml.Read(data)
		if err != nil {
			return
		}
		for _, name := range Files() {
			for _, e := range Validate(name, value) {
				if !slices.Contains(Keywords, e.Keyword) || e.Message == "" || !holds(value, e.Location) {
					t.Errorf("%s: an error out of form: %#v", name, e)
				}
			}
		}
	})
}
