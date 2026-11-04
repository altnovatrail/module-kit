package vectors

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/fs"
	"maps"
	"slices"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/altnovatrail/spec"
	"github.com/altnovatrail/spec/types/formatv0"
	"github.com/altnovatrail/spec/types/formatv0/open"

	"github.com/altnovatrail/module-kit/internal/yaml"
	"github.com/altnovatrail/module-kit/openview"
)

// readVector reads a file of a vector as the format has a reader do it: by the YAML rules, then
// against its schema, and only then into its type of formatv0, strictly, refusing a field that the
// type does not know (format/README.md, section 9; R72 spec). It gives the JSON value too. The
// runner of each kind reads its files so, as the tasks of the epic of code bring them.
func readVector(fsys fs.FS, name, schemaFile string, into any) (any, error) {
	data, err := fs.ReadFile(fsys, name)
	if err != nil {
		return nil, err
	}
	value, found, err := readBySchema(data, schemaFile)
	if err != nil {
		return nil, err
	}
	if len(found) > 0 {
		return nil, fmt.Errorf("refused by the YAML rules or by %s:\n%s", schemaFile, texts(found))
	}
	return value, decodeStrictly(value, into)
}

// readOpenView reads open-view.yaml of a vector of the open view, which has no schema of its own:
// by the YAML rules, then strictly into its type of formatv0/open (format/README.md, section 9;
// R72 spec). It gives the JSON value too.
func readOpenView(fsys fs.FS, name string, into any) (any, error) {
	data, err := fs.ReadFile(fsys, name)
	if err != nil {
		return nil, err
	}
	value, err := yaml.Read(data)
	if err != nil {
		return nil, fmt.Errorf("refused by the YAML rules:\n%v", err)
	}
	return value, decodeStrictly(value, into)
}

// decodeStrictly reads a JSON value into a type, refusing a field that the type does not know.
func decodeStrictly(value, into any) error {
	text, err := json.Marshal(value)
	if err != nil {
		return err
	}
	decoder := json.NewDecoder(bytes.NewReader(text))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(into); err != nil {
		return fmt.Errorf("not of the type %T: %v", into, err)
	}
	return nil
}

// A kind of file of the vectors: the schema it is valid against and the type that holds it.
type vectorFile struct {
	schema string
	value  func() any
}

func typeOf[T any]() func() any { return func() any { return new(T) } }

// The files of the vectors that the format gives a schema of their own: the file of each case of a
// kind of vectors, by the paths of spec.Format they stand at, with * for a name (format/README.md,
// sections 3 and 9). A published version of an earlier format in versions/ is the files of its
// package, read by their places alone (section 2), and open-view.yaml is the open view, which the
// type open reads.
var vectorFiles = map[string]vectorFile{
	"vectors/verdicts/*/*.yaml":       {"verdict-vector.schema.json", typeOf[formatv0.VerdictVector]()},
	"vectors/leaks/*.yaml":            {"leak-vector.schema.json", typeOf[formatv0.LeakVector]()},
	"vectors/random/*.yaml":           {"random-vector.schema.json", typeOf[formatv0.RandomVector]()},
	"vectors/templates/*.yaml":        {"template-vector.schema.json", typeOf[formatv0.TemplateVector]()},
	"vectors/tiers/*.yaml":            {"tier-vector.schema.json", typeOf[formatv0.TierVector]()},
	"vectors/packages/*/package.yaml": {"package-vector.schema.json", typeOf[formatv0.PackageVector]()},
	"vectors/indexes/*/index.yaml":    {"index-vector.schema.json", typeOf[formatv0.IndexVector]()},
	"vectors/versions/*/version.yaml": {"version-vector.schema.json", typeOf[formatv0.VersionVector]()},
}

// A part of a module's layout as the types hold it: its type of formatv0, its open view's of
// formatv0/open, and what gives the one from the other (R72 spec).
type partType struct {
	value func() any                  // a new value of its type of formatv0
	view  func() any                  // a new value of its type of formatv0/open
	open  func(part any) (any, error) // the open view of a value of the first, as a value of the second
}

func partOf[P, O any](build func(P) (O, error)) partType {
	return partType{typeOf[P](), typeOf[O](), func(part any) (any, error) {
		view, err := build(*part.(*P))
		return &view, err
	}}
}

// The types of the parts of a module's layout, whose open views have vectors.
var partTypes = map[string]partType{
	"lesson":        partOf(openview.Lesson),
	"manifest":      partOf(openview.Manifest),
	"skill-graph":   partOf(openview.SkillGraph),
	"solving-steps": partOf(openview.SolvingSteps),
	"task":          partOf(openview.Task),
	"techniques":    partOf(openview.Techniques),
	"traps":         partOf(openview.Traps),
}

// Every file of a vector that has a schema is read by its schema and its type of formatv0, so that
// the runners of those kinds can read them so. The runner of the open view reads both files of its
// vectors itself, each by its type.
func TestVectorFiles(t *testing.T) {
	if parts := slices.Sorted(maps.Keys(partTypes)); !slices.Equal(parts, layoutParts) {
		t.Errorf("the types are of the parts %v, and the layout has %v", parts, layoutParts)
	}
	for _, pattern := range slices.Sorted(maps.Keys(vectorFiles)) {
		names, err := fs.Glob(spec.Format, pattern)
		if err != nil {
			t.Fatal(err)
		}
		if len(names) == 0 {
			t.Errorf("no file stands at %s", pattern)
		}
		kind := vectorFiles[pattern]
		for _, name := range names {
			if _, err := readVector(spec.Format, name, kind.schema, kind.value()); err != nil {
				t.Errorf("%s: %v", name, err)
			}
		}
	}
}

// readOpenView refuses a file that is not YAML and one whose type does not know a field: the closed
// field of a part is no field of its open view.
func TestReadOpenView(t *testing.T) {
	for _, c := range []struct {
		name, data, refused string
	}{
		{"an open view", "techniques: {}\n", ""},
		{"not YAML", "a: [\n", "refused by the YAML rules"},
		{"a closed field", "techniques: {}\nnothing: 1\n", "not of the type"},
	} {
		_, err := readOpenView(fstest.MapFS{"v.yaml": {Data: []byte(c.data)}}, "v.yaml", new(open.Techniques))
		if c.refused == "" && err != nil || c.refused != "" && (err == nil || !strings.Contains(err.Error(), c.refused)) {
			t.Errorf("%s: %v, want %q", c.name, err, c.refused)
		}
	}
	if _, err := readOpenView(fstest.MapFS{}, "none.yaml", new(open.Techniques)); err == nil {
		t.Error("a file that is not there is read")
	}
}

// readVector refuses a file that is not YAML, one that its schema refuses and one whose type does
// not know a field.
func TestReadVector(t *testing.T) {
	const tier = "skill_graph:\n  skills:\n    a:\n      title: A\n      assessability: verdict\nskill_tiers:\n  a: 1\n"
	for _, c := range []struct {
		name, data string
		into       any
		refused    string
	}{
		{"a vector", tier, new(formatv0.TierVector), ""},
		{"not YAML", "a: [\n", new(formatv0.TierVector), "not YAML"},
		{"refused by its schema", "skill_tiers: {}\n", new(formatv0.TierVector), "refused by the YAML rules or by tier-vector.schema.json"},
		{"of another type", tier, new(struct {
			SkillGraph any `json:"skill_graph"`
		}), "not of the type"},
	} {
		_, err := readVector(fstest.MapFS{"v.yaml": {Data: []byte(c.data)}}, "v.yaml", "tier-vector.schema.json", c.into)
		if c.refused == "" && err != nil || c.refused != "" && (err == nil || !strings.Contains(err.Error(), c.refused)) {
			t.Errorf("%s: %v, want %q", c.name, err, c.refused)
		}
	}
	if _, err := readVector(fstest.MapFS{}, "none.yaml", "tier-vector.schema.json", new(formatv0.TierVector)); err == nil {
		t.Error("a file that is not there is read")
	}
}
