package schema

import (
	"encoding/json"
	"io/fs"
	"reflect"
	"sync"
	"testing"
	"testing/fstest"

	"github.com/altnovatrail/spec"

	"github.com/altnovatrail/module-kit/internal/yaml"
)

// openSet is a set of small schemas, one field for each way a mark counts or does not
// (format/README.md, section 6).
var openSet = newSet(fstest.MapFS{
	"schemas/o.schema.json": {Data: []byte(`{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "$id": "https://example.test/o.schema.json",
  "type": "object",
  "properties": {
    "id": {"type": "string", "x-open": true},
    "secret": {"type": "string"},
    "box": {"$ref": "common.schema.json#/$defs/box", "x-open": true},
    "list": {"type": "array", "items": {"$ref": "#/$defs/entry"}, "x-open": true},
    "closed_list": {"type": "array", "items": {"$ref": "#/$defs/entry"}},
    "pair": {"type": "array", "x-open": true,
      "prefixItems": [{"properties": {"a": {"x-open": true}, "b": {}}}],
      "items": {"properties": {"a": {}, "b": {"x-open": true}}}},
    "map": {"type": "object", "x-open": true,
      "patternProperties": {"^x-": {"x-open": true}, "^y": {}},
      "additionalProperties": {"type": "integer", "x-open": true}},
    "kind": {"enum": ["a", "b"], "x-open": true},
    "either": {"x-open": true, "anyOf": [
      {"properties": {"n": {"type": "integer", "x-open": true}}},
      {"properties": {"n": {"type": "string"}}}]},
    "one": {"x-open": true, "oneOf": [
      {"required": ["p"], "properties": {"p": {"x-open": true}}},
      {"required": ["q"], "properties": {"q": {}}}]},
    "deps": {"x-open": true, "properties": {"k": {"x-open": true}, "v": {}},
      "dependentSchemas": {"k": {"properties": {"v": {"x-open": true}}}}},
    "empty": {"type": "object", "properties": {"hidden": {}}, "x-open": true},
    "lone": {"if": {"required": ["a"]}, "x-open": true},
    "flags": {"type": "array", "items": true, "x-open": true},
    "bare": {"type": "array", "prefixItems": [{"properties": {"a": {"x-open": true}}}], "x-open": true},
    "loop": {"$ref": "#/$defs/round", "x-open": true},
    "twice": {"allOf": [{"$ref": "#/$defs/entry"}], "$ref": "#/$defs/entry", "x-open": true},
    "any": true
  },
  "allOf": [{"properties": {"extra": {"x-open": true}}}],
  "if": {"properties": {"kind": {"const": "a"}}, "required": ["kind"]},
  "then": {"properties": {"detail": {"x-open": true}}},
  "else": {"properties": {"detail": {}}},
  "$defs": {
    "entry": {"type": "object", "properties": {"name": {"x-open": true}, "key": {}}},
    "round": {"$ref": "#/$defs/about", "properties": {"x": {"x-open": true}}},
    "about": {"$ref": "#/$defs/round"}
  }
}`)},
	"schemas/common.schema.json": {Data: []byte(`{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "$id": "https://example.test/common.schema.json",
  "$defs": {"box": {"type": "object", "properties": {"w": {"x-open": true}, "h": {}}}}
}`)},
})

func jsonOf(t *testing.T, text string) any {
	t.Helper()
	var value any
	if err := json.Unmarshal([]byte(text), &value); err != nil {
		t.Fatal(err)
	}
	return value
}

func TestOpen(t *testing.T) {
	for _, c := range []struct {
		name, value, want string
	}{
		{"a field without the mark is closed", `{"id": "a", "secret": "s"}`, `{"id": "a"}`},
		{"the root is open, and stays when all its fields are closed", `{"secret": "s"}`, `{}`},
		{"a scalar at the root is itself", `"a"`, `"a"`},
		{"the mark beside $ref, and the fields of what it names", `{"box": {"w": 1, "h": 2}}`, `{"box": {"w": 1}}`},
		{"an element of an open array is open", `{"list": [{"name": "n", "key": "k"}, {"key": "k"}]}`,
			`{"list": [{"name": "n"}, {}]}`},
		{"a closed array is closed, whatever its elements", `{"closed_list": [{"name": "n"}]}`, `{}`},
		{"prefixItems at its index, and items after it", `{"pair": [{"a": 1, "b": 2}, {"a": 3, "b": 4}]}`,
			`{"pair": [{"a": 1}, {"b": 4}]}`},
		{"patternProperties, and additionalProperties where none matches", `{"map": {"x-a": 1, "yes": 2, "z": 3}}`,
			`{"map": {"x-a": 1, "z": 3}}`},
		{"the branch of anyOf that passes", `{"either": {"n": 1}}`, `{"either": {"n": 1}}`},
		{"not the branch of anyOf that fails", `{"either": {"n": "one"}}`, `{"either": {}}`},
		{"the branch of oneOf that passes", `{"one": {"p": 1}}`, `{"one": {"p": 1}}`},
		{"not the branch of oneOf that fails", `{"one": {"q": 1}}`, `{"one": {}}`},
		{"dependentSchemas whose key the object has", `{"deps": {"k": 1, "v": 2}}`, `{"deps": {"k": 1, "v": 2}}`},
		{"not dependentSchemas whose key the object lacks", `{"deps": {"v": 2}}`, `{"deps": {}}`},
		{"an open object with no open field", `{"empty": {"hidden": 1}}`, `{"empty": {}}`},
		{"if with neither then nor else", `{"lone": {"a": 1}}`, `{"lone": {}}`},
		{"true carries no mark", `{"any": 1}`, `{}`},
		{"nor does it under items", `{"flags": [1, {"a": 1}]}`, `{"flags": [1, {}]}`},
		{"an element past prefixItems, with no items", `{"bare": [{"a": 1}, {"a": 2}]}`, `{"bare": [{"a": 1}, {}]}`},
		{"references in a circle, each taken once", `{"loop": {"x": 1, "y": 2}}`, `{"loop": {"x": 1}}`},
		{"one definition by two routes", `{"twice": {"name": "n", "key": "k"}}`, `{"twice": {"name": "n"}}`},
		{"every branch of allOf", `{"extra": 1}`, `{"extra": 1}`},
		{"then when if passes", `{"kind": "a", "detail": 1}`, `{"kind": "a", "detail": 1}`},
		{"else when if fails", `{"kind": "b", "detail": 1}`, `{"kind": "b"}`},
	} {
		value := jsonOf(t, c.value)
		got := openSet.open("o.schema.json", value)
		if want := jsonOf(t, c.want); !reflect.DeepEqual(got, want) {
			t.Errorf("%s: %v, want %v", c.name, got, want)
		}
		if !reflect.DeepEqual(value, jsonOf(t, c.value)) {
			t.Errorf("%s: the value changed to %v", c.name, value)
		}
	}
}

// Open starts at a subschema that its location names, as Validate does.
func TestOpenAtSubschema(t *testing.T) {
	got := openSet.open("o.schema.json#/$defs/entry", jsonOf(t, `{"name": "n", "key": "k"}`))
	if want := jsonOf(t, `{"name": "n"}`); !reflect.DeepEqual(got, want) {
		t.Errorf("%v, want %v", got, want)
	}
	got = openSet.open("common.schema.json#/%24defs/box", jsonOf(t, `{"w": 1, "h": 2}`))
	if want := jsonOf(t, `{"w": 1}`); !reflect.DeepEqual(got, want) {
		t.Errorf("a percent-encoded fragment: %v, want %v", got, want)
	}
}

// A location or a reference that names nothing in the set is a fault of module-kit.
func TestOpenFaults(t *testing.T) {
	for _, location := range []string{
		"none.schema.json", "o.schema.json#/$defs/nothing", "o.schema.json#nothing", "o.schema.json#%zz",
		"o.schema.json#/allOf/1", "o.schema.json#/allOf/00", "o.schema.json#/allOf/x", "o.schema.json#/properties/id/type/x",
	} {
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("Open(%q) does not panic", location)
				}
			}()
			openSet.open(location, nil)
		}()
	}
	func() {
		defer func() {
			if recover() == nil {
				t.Error("a reference to nothing does not panic")
			}
		}()
		newSet(fstest.MapFS{"schemas/a.schema.json": {Data: []byte(`{"$id": "https://example.test/a.schema.json", "$ref": "none.schema.json"}`)}}).
			open("a.schema.json", nil)
	}()
}

// The tokens of a pointer are read as RFC 6901 has them, ~01 as ~1.
func TestTokensOf(t *testing.T) {
	for fragment, want := range map[string][]string{
		"":              nil,
		"/":             {""},
		"/a~1b/~01/%25": {"a/b", "~1", "%"},
	} {
		if got := tokensOf("a.schema.json", fragment); !reflect.DeepEqual(got, want) {
			t.Errorf("tokensOf(%q) is %q, want %q", fragment, got, want)
		}
	}
}

// Open is safe from many goroutines at once, compiling as it goes.
func TestConcurrentOpen(t *testing.T) {
	names, err := fs.Glob(spec.Format, "vectors/open-view/task/*/part.yaml")
	if err != nil || len(names) == 0 {
		t.Fatal("no vector of the open view of a task", err)
	}
	values := make([]any, len(names))
	for i, name := range names {
		data, err := fs.ReadFile(spec.Format, name)
		if err != nil {
			t.Fatal(err)
		}
		if values[i], err = yaml.Read(data); err != nil {
			t.Fatal(name, err)
		}
	}
	set := newSet(files)
	var wg sync.WaitGroup
	for i := range 16 {
		wg.Go(func() {
			for range 5 {
				set.open("task.schema.json", values[i%len(values)])
			}
		})
	}
	wg.Wait()
}
