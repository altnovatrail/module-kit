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
