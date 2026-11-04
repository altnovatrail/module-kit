package openview

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/altnovatrail/spec/types/formatv0"
	"github.com/altnovatrail/spec/types/formatv0/open"

	"github.com/altnovatrail/module-kit/internal/schema"
)

// Manifest gives the open view of a manifest. It fails when the manifest is not valid against its
// schema, since the open view is the format's of a valid part alone.
func Manifest(part formatv0.Manifest) (open.Manifest, error) {
	return of[open.Manifest]("manifest", part)
}

// SkillGraph gives the open view of a skill graph, as Manifest does of a manifest.
func SkillGraph(part formatv0.SkillGraph) (open.SkillGraph, error) {
	return of[open.SkillGraph]("skill-graph", part)
}

// Traps gives the open view of a trap catalog, as Manifest does of a manifest: every field of the
// catalog is closed, so its open view is an empty object.
func Traps(part formatv0.Traps) (open.Traps, error) {
	return of[open.Traps]("traps", part)
}

// Techniques gives the open view of a technique catalog, as Manifest does of a manifest.
func Techniques(part formatv0.Techniques) (open.Techniques, error) {
	return of[open.Techniques]("techniques", part)
}

// SolvingSteps gives the open view of the solving steps, as Manifest does of a manifest.
func SolvingSteps(part formatv0.SolvingSteps) (open.SolvingSteps, error) {
	return of[open.SolvingSteps]("solving-steps", part)
}

// Lesson gives the open view of a lesson, as Manifest does of a manifest.
func Lesson(part formatv0.Lesson) (open.Lesson, error) {
	return of[open.Lesson]("lesson", part)
}

// Task gives the open view of a task, as Manifest does of a manifest: without its checks, its
// solution, its traps, its leak phrases and its template, among the other closed fields.
func Task(part formatv0.Task) (open.Task, error) {
	return of[open.Task]("task", part)
}

// of gives the open view of a part of a module's layout, named as its schema's file is, as the
// type open has it. The part goes as JSON to the schema of its version, and the open view comes
// back strictly into its type: a field that the type does not know is a fault, since the type is
// what the marks of the same schemas give (R72 spec).
func of[O any](part string, value any) (O, error) {
	file := part + ".schema.json"
	v := jsonValue(value)
	if errs := schema.Validate(file, v); len(errs) > 0 {
		lines := make([]string, len(errs))
		for i, e := range errs {
			lines[i] = e.String()
		}
		var none O
		return none, fmt.Errorf("openview: the %s is not valid against %s:\n%s", part, file, strings.Join(lines, "\n"))
	}
	view, err := schema.Decode[O](schema.Open(file, v))
	if err != nil {
		panic(fmt.Sprintf("openview: the open view of the %s: %v", part, err))
	}
	return view, nil
}

// jsonValue gives a value of a type of formatv0 as the reader of YAML gives a JSON value:
// map[string]any, []any, string, int64 for a number without a fraction or an exponent, float64 for
// another, bool or nil. A part of formatv0 holds only strings, integers, booleans and lists and
// maps of them, so it is always JSON.
func jsonValue(value any) any {
	data, err := json.Marshal(value)
	if err != nil {
		panic(fmt.Sprintf("openview: %T is no JSON: %v", value, err))
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	var v any
	// The text is what json.Marshal has just written.
	_ = decoder.Decode(&v)
	return numbers(v)
}

// numbers puts int64 or float64 in the place of each json.Number in v.
func numbers(v any) any {
	switch v := v.(type) {
	case map[string]any:
		for key, member := range v {
			v[key] = numbers(member)
		}
	case []any:
		for i, element := range v {
			v[i] = numbers(element)
		}
	case json.Number:
		if i, err := v.Int64(); err == nil {
			return i
		}
		// json.Marshal writes a float64 that it can read back.
		f, _ := v.Float64()
		return f
	}
	return v
}
