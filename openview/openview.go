package openview

import (
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
	v, err := viewOf(part, value)
	if err != nil {
		var none O
		return none, err
	}
	view, err := schema.Decode[O](v)
	if err != nil {
		panic(fmt.Sprintf("openview: the open view of the %s: %v", part, err))
	}
	return view, nil
}

// viewOf gives the open view of a part as the JSON value that the walk of its schema gives, before
// any type holds it, so that a field is in it exactly when the open view has it. It fails when the
// part is not valid against its schema.
func viewOf(part string, value any) (any, error) {
	file := part + ".schema.json"
	v := schema.ValueOf(value)
	if errs := schema.Validate(file, v); len(errs) > 0 {
		lines := make([]string, len(errs))
		for i, e := range errs {
			lines[i] = e.String()
		}
		return nil, fmt.Errorf("openview: the %s is not valid against %s:\n%s", part, file, strings.Join(lines, "\n"))
	}
	return schema.Open(file, v), nil
}
