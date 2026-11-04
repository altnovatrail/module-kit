package schema

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/fs"
	"regexp"
	"slices"
	"strings"
	"sync"
	"testing"
	"testing/fstest"

	"github.com/altnovatrail/spec"
)

// The copy of the schemas is schemas/ of spec.Format: the same files, byte for byte, of the version
// that spec.Format holds. A move to a new tag of spec that changes a schema fails it until the copy
// follows, as R20 has the move take the copy along.
func TestSchemasAreSpecs(t *testing.T) {
	version, err := fs.ReadFile(spec.Format, "VERSION")
	if err != nil {
		t.Fatal(err)
	}
	prefix := "https://altnovatrail.github.io/spec/format/" + strings.TrimSpace(string(version)) + "/"
	theirs := names(spec.Format)
	if !slices.Equal(Files(), theirs) {
		t.Errorf("the copy holds %v, and spec.Format %v", Files(), theirs)
	}
	for _, name := range theirs {
		want, err := fs.ReadFile(spec.Format, "schemas/"+name)
		if err != nil {
			t.Fatal(err)
		}
		got, err := files.ReadFile("schemas/" + name)
		if err != nil {
			t.Errorf("the copy holds no %s", name)
			continue
		}
		if !bytes.Equal(got, want) {
			t.Errorf("%s of the copy is not that of spec.Format: copy it again from the module cache", name)
		}
		var root struct {
			ID string `json:"$id"`
		}
		if err := json.Unmarshal(got, &root); err != nil || root.ID != prefix+name {
			t.Errorf("%s has the $id %q, and the version of spec.Format gives %q", name, root.ID, prefix+name)
		}
	}
}

// Keywords is the list of section 5 of format/README.md, in its order.
func TestKeywordsAreTheFormats(t *testing.T) {
	data, err := fs.ReadFile(spec.Format, "README.md")
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	start := strings.Index(text, "- A schema uses only these keywords:")
	end := strings.Index(text, "The other keywords of 2020-12 stay out of the format")
	if start < 0 || end < start {
		t.Fatal("format/README.md, section 5, lists no keywords where the test reads them")
	}
	var listed []string
	for _, m := range regexp.MustCompile("`([^`]+)`").FindAllStringSubmatch(text[start:end], -1) {
		listed = append(listed, m[1])
	}
	if !slices.Equal(listed, Keywords) {
		t.Errorf("section 5 lists\n%v\nand Keywords is\n%v", listed, Keywords)
	}
}

// Every schema of the format compiles, and so does a subschema of one by its pointer.
func TestEverySchemaCompiles(t *testing.T) {
	for _, name := range Files() {
		Validate(name, nil)
	}
	if errs := Validate("common.schema.json#/$defs/local_id", "counting-gaps"); errs != nil {
		t.Errorf("a local ID fails $defs/local_id: %v", errs)
	}
	if errs := Validate("task.schema.json#/$defs/check", map[string]any{"kind": "number", "key": "7"}); errs != nil {
		t.Errorf("a check fails $defs/check: %v", errs)
	}
}

// A location that names no schema of the format, or nothing in one, is a fault of module-kit.
func TestFaults(t *testing.T) {
	for _, location := range []string{"no.schema.json", "", "task.schema.json#/$defs/nothing"} {
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("Validate(%q) does not panic", location)
				}
			}()
			Validate(location, nil)
		}()
	}
	func() {
		defer func() {
			if recover() == nil {
				t.Error("a value that is no JSON value does not panic")
			}
		}()
		Validate("task.schema.json", map[string]any{"id": struct{}{}})
	}()
	func() {
		defer func() {
			if recover() == nil {
				t.Error("false beside no unevaluatedProperties does not panic")
			}
		}()
		testSet.validate("t.schema.json", map[string]any{"id": "a", "nothing": 1})
	}()
	for name, schema := range map[string]string{
		"no JSON":                `{`,
		"another $id":            `{"$id": "https://example.test/other.schema.json"}`,
		"a reference to nothing": `{"$id": "https://example.test/a.schema.json", "$ref": "none.schema.json"}`,
	} {
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("%s: no panic", name)
				}
			}()
			newSet(fstest.MapFS{"schemas/a.schema.json": {Data: []byte(schema)}}).validate("a.schema.json", nil)
		}()
	}
}

// testSet is a set of small schemas, one place for each way the validator nests what fails.
var testSet = newSet(fstest.MapFS{
	"schemas/t.schema.json": {Data: []byte(`{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "$id": "https://example.test/t.schema.json",
  "type": "object",
  "properties": {
    "id": {"$ref": "common.schema.json#/$defs/id"},
    "list": {"type": "array", "items": {"type": "integer", "maximum": 9}, "uniqueItems": true},
    "both": {"allOf": [{"minLength": 2}, {"pattern": "^a"}]},
    "either": {"anyOf": [{"type": "integer"}, {"type": "string", "minLength": 3}]},
    "one": {"oneOf": [{"const": "a"}, {"const": "b"}]},
    "never": {"not": {"type": "string"}},
    "has": {"type": "array", "contains": {"const": 1}},
    "named": {"type": "array", "items": {"type": "object", "propertyNames": {"enum": ["ok"]}}},
    "map": {"type": "object", "additionalProperties": {"type": "string"}, "propertyNames": {"pattern": "^[a-z]+$"}},
    "loose": {"propertyNames": {"maxLength": 1}},
    "kind": {"enum": ["x", "y"]},
    "deep": {"$ref": "#/$defs/deep", "unevaluatedProperties": false},
    "nothing": false,
    "pair": {"type": "object", "properties": {"a": {}, "b": {}}, "dependentRequired": {"a": ["b"]}, "unevaluatedProperties": false}
  },
  "required": ["id"],
  "if": {"properties": {"kind": {"const": "x"}}, "required": ["kind"]},
  "then": {"required": ["list"]},
  "else": {"properties": {"list": {"maxItems": 1}}},
  "unevaluatedProperties": false,
  "$defs": {
    "deep": {"type": "object", "properties": {"a/b~c": {"type": "string"}, "é": {"type": "string"}}}
  }
}`)},
	"schemas/common.schema.json": {Data: []byte(`{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "$id": "https://example.test/common.schema.json",
  "$defs": {"id": {"type": "string", "pattern": "^[a-z]+$", "maxLength": 5}}
}`)},
})

func TestCollect(t *testing.T) {
	for _, c := range []struct {
		name  string
		value string
		want  []string
	}{
		{"valid", `{"id": "a"}`, nil},
		{"the root of no type", `[]`, []string{"type at #"}},
		{"required at the object", `{}`, []string{"required at #"}},
		{"through $ref", `{"id": "ABCDEF"}`, []string{"maxLength at #/id", "pattern at #/id"}},
		{"an element by its index", `{"id": "a", "kind": "x", "list": [1, 10, 1]}`, []string{"maximum at #/list/1", "uniqueItems at #/list"}},
		{"through allOf", `{"id": "a", "both": "b"}`, []string{"minLength at #/both", "pattern at #/both"}},
		{"anyOf without its branches", `{"id": "a", "either": "ab"}`, []string{"anyOf at #/either"}},
		{"oneOf without its branches", `{"id": "a", "one": "c"}`, []string{"oneOf at #/one"}},
		{"not", `{"id": "a", "never": "s"}`, []string{"not at #/never"}},
		{"contains without its elements", `{"id": "a", "has": [2, 3]}`, []string{"contains at #/has"}},
		{"then", `{"id": "a", "kind": "x"}`, []string{"required at #"}},
		{"else", `{"id": "a", "kind": "y", "list": [1, 2]}`, []string{"maxItems at #/list"}},
		{"an unknown field at its object", `{"id": "a", "colour": "red"}`, []string{"unevaluatedProperties at #"}},
		{"dependentRequired at the object", `{"id": "a", "pair": {"a": 1}}`, []string{"dependentRequired at #/pair"}},
		// A subschema that fails gives no annotations (JSON Schema 2020-12 core, section 7.7.1.2), so the
		// fields that $ref declares are unevaluated beside it, as ajv has them too.
		{"escaped tokens", `{"id": "a", "deep": {"a/b~c": 1, "é": 2, "x y": 3}}`, []string{
			"type at #/deep/%C3%A9", "type at #/deep/a~1b~0c",
			"unevaluatedProperties at #/deep", "unevaluatedProperties at #/deep", "unevaluatedProperties at #/deep",
		}},
		{"an unknown field beside a $ref that passes", `{"id": "a", "deep": {"é": "e", "x y": 3}}`, []string{"unevaluatedProperties at #/deep"}},
		// santhosh-tekuri/jsonschema#276: the error of propertyNames at the first element names the
		// second, which comes after it, unless the vocabulary propertyNames gives it.
		{"propertyNames at its object, before a sibling", `{"id": "a", "named": [{"bad": 1}, {"ok": 1}, {"ok": 2}]}`,
			[]string{"propertyNames at #/named/0"}},
		{"propertyNames on no object, with no type", `{"id": "a", "loose": 5}`, nil},
		{"propertyNames on no object", `{"id": "a", "map": "x"}`, []string{"type at #/map"}},
		{"propertyNames of a map", `{"id": "a", "map": {"B": "x", "c": 1, "d": "y"}}`,
			[]string{"propertyNames at #/map", "type at #/map/c"}},
	} {
		var value any
		if err := json.Unmarshal([]byte(c.value), &value); err != nil {
			t.Fatal(err)
		}
		var got []string
		for _, e := range testSet.validate("t.schema.json", value) {
			if e.Message == "" {
				t.Errorf("%s: %s has no message", c.name, e)
			}
			got = append(got, e.Keyword+" at "+e.Pointer())
		}
		slices.Sort(got)
		if !slices.Equal(got, c.want) {
			t.Errorf("%s: %q, want %q", c.name, got, c.want)
		}
	}
}

func TestPointer(t *testing.T) {
	for _, c := range []struct {
		tokens []string
		want   string
	}{
		{nil, "#"},
		{[]string{"answer_slots", "0"}, "#/answer_slots/0"},
		{[]string{"modules", "example/smallest", "0.1.0", "withdrawn"}, "#/modules/example~1smallest/0.1.0/withdrawn"},
		{[]string{"a~b", "~1"}, "#/a~0b/~01"},
		{[]string{"x y", "é", "a#b", "%"}, "#/x%20y/%C3%A9/a%23b/%25"},
		{[]string{""}, "#/"},
	} {
		if got := Pointer(c.tokens); got != c.want {
			t.Errorf("Pointer(%q) is %s, want %s", c.tokens, got, c.want)
		}
	}
}

// Validate is safe from many goroutines at once, compiling as it goes.
func TestConcurrentValidate(t *testing.T) {
	var wg sync.WaitGroup
	for i := range 16 {
		wg.Go(func() {
			name := Files()[i%len(Files())]
			for range 20 {
				Validate(name, map[string]any{"id": fmt.Sprint(i)})
			}
		})
	}
	wg.Wait()
}

func TestErrorString(t *testing.T) {
	e := Error{Keyword: "required", Location: []string{"items", "0"}, Message: "missing property 'line'"}
	if got, want := e.String(), "required at #/items/0: missing property 'line'"; got != want {
		t.Errorf("%q, want %q", got, want)
	}
}
