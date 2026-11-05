package openview

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/fs"
	"path"
	"strings"
	"testing"

	"github.com/altnovatrail/spec"
	"github.com/altnovatrail/spec/types/formatv0"
	"github.com/altnovatrail/spec/types/formatv0/open"

	"github.com/altnovatrail/module-kit/internal/yaml"
)

// readPart reads a file of a part of spec.Format strictly into its type of formatv0.
func readPart(t *testing.T, name string, into any) {
	t.Helper()
	data, err := fs.ReadFile(spec.Format, name)
	if err != nil {
		t.Fatal(err)
	}
	value, err := yaml.Read(data)
	if err != nil {
		t.Fatal(err)
	}
	text, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	decoder := json.NewDecoder(bytes.NewReader(text))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(into); err != nil {
		t.Fatalf("%s: %v", name, err)
	}
}

// A function of a part: a new value of its type of formatv0, and its open view by the function.
type function struct {
	value func() any
	open  func(part any) error
}

func functionOf[P, O any](build func(P) (O, error)) function {
	return function{func() any { return new(P) }, func(part any) error {
		_, err := build(*part.(*P))
		return err
	}}
}

// Every valid example of a part of the layout has an open view, by the function of its part.
func TestParts(t *testing.T) {
	for part, f := range map[string]function{
		"lesson":        functionOf(Lesson),
		"manifest":      functionOf(Manifest),
		"skill-graph":   functionOf(SkillGraph),
		"solving-steps": functionOf(SolvingSteps),
		"task":          functionOf(Task),
		"techniques":    functionOf(Techniques),
		"traps":         functionOf(Traps),
	} {
		names, err := fs.Glob(spec.Format, path.Join("examples", part, "valid", "*.yaml"))
		if err != nil || len(names) == 0 {
			t.Fatalf("%s has no valid example: %v", part, err)
		}
		for _, name := range names {
			value := f.value()
			readPart(t, name, value)
			if err := f.open(value); err != nil {
				t.Errorf("%s: %v", name, err)
			}
		}
	}
}

// The checks, the solution and the other closed fields of a task are not in its open view, and its
// statement, canvas and items are.
func TestTask(t *testing.T) {
	var task formatv0.Task
	readPart(t, "vectors/open-view/task/zones-and-picture/part.yaml", &task)
	view, err := Task(task)
	if err != nil {
		t.Fatal(err)
	}
	if view.ID != task.ID || len(view.Statement) != len(task.Statement) || len(view.Items) != 2 ||
		view.Items[0].AnswerSlots[0].Zone == nil || view.Canvas.Kind != formatv0.CanvasKindHandwritingSheet {
		t.Errorf("the open view of the task is %+v", view)
	}
	text, err := json.Marshal(view)
	if err != nil {
		t.Fatal(err)
	}
	for _, closed := range []string{"check", "solution", "skills", "band", "difficulty"} {
		if strings.Contains(string(text), `"`+closed+`"`) {
			t.Errorf("the open view holds %s: %s", closed, text)
		}
	}
}

// A part that is not valid against its schema has no open view, and the error names what fails.
func TestInvalid(t *testing.T) {
	_, err := Task(formatv0.Task{ID: "a"})
	if err == nil || !strings.Contains(err.Error(), "not valid against task.schema.json") ||
		!strings.Contains(err.Error(), "required at #") {
		t.Errorf("an invalid task: %v", err)
	}
	if _, err := Traps(formatv0.Traps{}); err == nil {
		t.Error("a trap catalog without its traps has an open view")
	}
}

// A type that does not hold what the marks open is a fault of module-kit.
func TestTypeOfNoMarks(t *testing.T) {
	var task formatv0.Task
	readPart(t, "vectors/open-view/task/zones-and-picture/part.yaml", &task)
	defer func() {
		if recover() == nil {
			t.Error("an open view that its type does not hold does not panic")
		}
	}()
	_, _ = of[open.Traps]("task", task)
}

func ExampleTechniques() {
	view, err := Techniques(formatv0.Techniques{Techniques: map[formatv0.TechniqueID]formatv0.Technique{
		"counting-gaps": {Title: "Count the gaps"},
	}})
	fmt.Println(view.Techniques["counting-gaps"].Title, err)
	// Output: Count the gaps <nil>
}
