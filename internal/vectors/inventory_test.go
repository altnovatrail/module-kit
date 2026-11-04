package vectors

import (
	"errors"
	"io/fs"
	"maps"
	"slices"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/altnovatrail/spec"
)

// files makes a file system of empty files at the names given.
func files(names ...string) fstest.MapFS {
	fsys := fstest.MapFS{}
	for _, name := range names {
		fsys[name] = &fstest.MapFile{}
	}
	return fsys
}

// A file of every kind, and of none, comes out as the cases it belongs to, each once and in order,
// and every kind the inventory gives has a runner.
func TestInventoryNamesEachCase(t *testing.T) {
	format := files(
		"README.md",
		"VERSION",
		"schemas/task.schema.json",
		"examples/task/valid/smallest.yaml",
		"examples/task/invalid/missing-title.yaml",
		"examples/module/valid/smallest/manifest.yaml",
		"examples/module/valid/smallest/LICENSE",
		"vectors/rules/yaml.encoding/byte-order-mark.yaml",
		"vectors/rules/module.layout/unknown-file/manifest.yaml",
		"vectors/rules/module.layout/unknown-file/notes.txt",
		"vectors/open-view/task/hints-closed/part.yaml",
		"vectors/open-view/task/hints-closed/open-view.yaml",
		"vectors/verdicts/number/negative.yaml",
		"vectors/leaks/long-number.yaml",
		"vectors/random/one-name.yaml",
		"vectors/templates/escapes.yaml",
		"vectors/tiers/chain.yaml",
		"vectors/packages/smallest/package.yaml",
		"vectors/packages/smallest/module/.editor-notes",
		"vectors/indexes/first-release/index.yaml",
		"vectors/indexes/first-release/modules/smallest/manifest.yaml",
		"vectors/versions/zone-moved/version.yaml",
		"vectors/versions/zone-moved/module/manifest.yaml",
		"vectors/versions/zone-moved/published/1.0.0/manifest.yaml",
	)
	contracts := files(
		"observation/v1/boxes.json",
		"eval/v1/groups.json",
	)
	cases, err := inventory(format, contracts)
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]string{
		"contracts/vectors/observation/v1/boxes.json":             "observation/v1",
		"format/examples/module/valid/smallest":                   "examples/module",
		"format/examples/task/invalid/missing-title.yaml":         "examples/invalid",
		"format/examples/task/valid/smallest.yaml":                "examples/valid",
		"format/vectors/indexes/first-release":                    "indexes",
		"format/vectors/leaks/long-number.yaml":                   "leaks",
		"format/vectors/open-view/task/hints-closed":              "open-view",
		"format/vectors/packages/smallest":                        "packages",
		"format/vectors/random/one-name.yaml":                     "random",
		"format/vectors/rules/module.layout/unknown-file":         "rules",
		"format/vectors/rules/yaml.encoding/byte-order-mark.yaml": "rules",
		"format/vectors/templates/escapes.yaml":                   "templates",
		"format/vectors/tiers/chain.yaml":                         "tiers",
		"format/vectors/verdicts/number/negative.yaml":            "verdicts",
		"format/vectors/versions/zone-moved":                      "versions",
	}
	var names []string
	for _, c := range cases {
		names = append(names, c.name)
		if want[c.name] != c.kind {
			t.Errorf("%s is of the kind %q, not %q", c.name, c.kind, want[c.name])
		}
		if runners[c.kind] == nil {
			t.Errorf("%s is of the kind %s, which has no runner", c.name, c.kind)
		}
		if !strings.HasSuffix(c.name, c.path) {
			t.Errorf("%s has the path %s in its file system", c.name, c.path)
		}
	}
	if !slices.Equal(names, slices.Sorted(maps.Keys(want))) {
		t.Errorf("the cases are\n%s", strings.Join(names, "\n"))
	}
	if missing := kindsWithoutCase(cases); len(missing) > 0 {
		t.Errorf("the kinds without a case are %v", missing)
	}
}

// A file of no kind the test knows is refused by its path in spec, whatever else is there.
func TestInventoryRefusesWhatItDoesNotKnow(t *testing.T) {
	format := []string{
		"CHANGELOG.md",
		"schemas/task.json",
		"schemas/old/task.schema.json",
		"examples/widget/valid/smallest.yaml",
		"examples/task/draft/smallest.yaml",
		"examples/task/valid/smallest.json",
		"examples/module/invalid/no-manifest/LICENSE",
		"vectors/widgets/smallest.yaml",
		"vectors/rules/Yaml_Encoding/byte-order-mark.yaml",
		"vectors/rules/module-index.one-package/twice.yaml",
		"vectors/rules/module-version.above/same-version/manifest.yaml",
		"vectors/rules/yaml.encoding/byte-order-mark.txt",
		"vectors/open-view/canvas/smallest/part.yaml",
		"vectors/open-view/task/smallest/notes.yaml",
		"vectors/verdicts/rubric/essay.yaml",
		"vectors/leaks/nested/long-number.yaml",
		"vectors/packages/smallest.yaml",
		"vectors/packages/smallest/notes.yaml",
		"vectors/packages/smallest/package.yaml/inside",
		"vectors/indexes/first-release/withdrawals.yaml",
	}
	contracts := []string{
		"api/v1/progress.json",
		"observation/v2/boxes.json",
		"observation/v1/boxes.yaml",
	}
	for _, name := range format {
		t.Run(formatRoot+name, func(t *testing.T) {
			_, err := inventory(files("README.md", name), files())
			if err == nil || !strings.Contains(err.Error(), formatRoot+name) {
				t.Errorf("the inventory gives %v", err)
			}
		})
	}
	for _, name := range contracts {
		t.Run(contractsRoot+name, func(t *testing.T) {
			_, err := inventory(files(), files(name))
			if err == nil || !strings.Contains(err.Error(), contractsRoot+name) {
				t.Errorf("the inventory gives %v", err)
			}
		})
	}
}

// A folder of its own kind that leaves out a thing its kind names is refused by its path, with what
// it holds.
func TestInventoryRefusesAnIncompleteFolder(t *testing.T) {
	for _, test := range []struct {
		name  string
		files []string
	}{
		{"format/vectors/open-view/task/hints-closed", []string{"vectors/open-view/task/hints-closed/part.yaml"}},
		{"format/vectors/packages/smallest", []string{"vectors/packages/smallest/module/manifest.yaml"}},
		{"format/vectors/indexes/first-release", []string{"vectors/indexes/first-release/index.yaml"}},
		{"format/vectors/versions/zone-moved", []string{
			"vectors/versions/zone-moved/module/manifest.yaml",
			"vectors/versions/zone-moved/version.yaml",
		}},
	} {
		t.Run(test.name, func(t *testing.T) {
			_, err := inventory(files(test.files...), files())
			if err == nil || !strings.Contains(err.Error(), test.name+" holds") {
				t.Errorf("the inventory gives %v", err)
			}
		})
	}
}

// An empty file system has a case of no kind, and each kind is named.
func TestKindsWithoutCase(t *testing.T) {
	missing := kindsWithoutCase(nil)
	if !slices.Equal(missing, slices.Sorted(maps.Keys(runners))) {
		t.Errorf("with no case, the kinds without one are %v", missing)
	}
}

func TestParseSkipped(t *testing.T) {
	for _, test := range []struct {
		data  string
		names []string // nil for a list refused
	}{
		{"", []string{}},
		{"format/a\n", []string{"format/a"}},
		{"format/a\nformat/b\n", []string{"format/a", "format/b"}},
		{"format/a", nil},
		{"format/b\nformat/a\n", nil},
		{"format/a\nformat/a\n", nil},
		{"format/a\n\nformat/b\n", nil},
		{"format/a \n", nil},
		{"\n", nil},
	} {
		skipped, err := parseSkipped([]byte(test.data))
		switch {
		case test.names == nil && err == nil:
			t.Errorf("%q is taken as %v", test.data, skipped)
		case test.names != nil && err != nil:
			t.Errorf("%q is refused: %v", test.data, err)
		case test.names != nil && !slices.Equal(slices.Sorted(maps.Keys(skipped)), test.names):
			t.Errorf("%q gives %v", test.data, skipped)
		}
	}
}

// A skipped case that spec does not have is named, and one it has is not.
func TestUnknownSkips(t *testing.T) {
	cases := []vector{{name: "format/a"}, {name: "format/c"}}
	unknown := unknownSkips(cases, map[string]bool{"format/a": true, "format/b": true, "format/d": true})
	if !slices.Equal(unknown, []string{"format/b", "format/d"}) {
		t.Errorf("the unknown skips are %v", unknown)
	}
}

// A case passes, fails, is skipped, or fails for passing while skipped.
func TestOutcome(t *testing.T) {
	failed := errors.New("the verdict is not_yet, not correct")
	for _, test := range []struct {
		skipped bool
		err     error
		failure string
		skip    bool
	}{
		{false, nil, "", false},
		{false, failed, failed.Error(), false},
		{true, failed, "", true},
		{true, nil, "it passes: take it out of testdata/skipped.txt", false},
	} {
		failure, skip := outcome(test.skipped, test.err)
		if failure != test.failure || skip != test.skip {
			t.Errorf("skipped %v, %v: %q, %v", test.skipped, test.err, failure, skip)
		}
	}
}

// The first line of an invalid example names a keyword at a JSON Pointer in URI fragment form, and
// the place it names is the pointer the fragment encodes.
func TestExpectedError(t *testing.T) {
	for _, c := range []struct {
		line, keyword, location, refused string
	}{
		{"# error: required at #", "required", "", ""},
		{"# error: type at #/difficulty", "type", "/difficulty", ""},
		{"# error: type at #/difficulty\r", "type", "/difficulty", ""},
		{"# error: enum at #/modules/example~1smallest/0.1.0/withdrawn", "enum", "/modules/example~1smallest/0.1.0/withdrawn", ""},
		{"# error: pattern at #/x%20y/%C3%A9", "pattern", "/x y/é", ""},
		{"# error: required at", "", "", "names no error"},
		{"# error: required at difficulty", "", "", "names no error"},
		{"# error: required at # and more", "", "", "names no error"},
		{"# error: required at #difficulty", "", "", "encodes no JSON Pointer"},
		{"# error: required at #/%zz", "", "", "invalid URL escape"},
		{"id: a", "", "", "names no error"},
	} {
		keyword, location, err := expectedError([]byte(c.line + "\nid: a\n"))
		switch {
		case c.refused == "" && (err != nil || keyword != c.keyword || location != c.location):
			t.Errorf("%q: %q at %q, %v; want %q at %q", c.line, keyword, location, err, c.keyword, c.location)
		case c.refused != "" && (err == nil || !strings.Contains(err.Error(), c.refused)):
			t.Errorf("%q: %v, want %q", c.line, err, c.refused)
		}
	}
	if _, err := decodePlace("/difficulty"); err == nil {
		t.Error("a place with no # is read")
	}
}

// An example fails its runner when it is refused though valid, when it passes though invalid, when
// it breaks more than its schema, and when it gives another error than its first line names.
func TestExampleRunners(t *testing.T) {
	manifest, err := fs.ReadFile(spec.Format, "examples/manifest/valid/olympiad-math.yaml")
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range []struct {
		path, data string
		run        func(vector) error
		refused    string
	}{
		{"examples/manifest/valid/a.yaml", "id: x\n", validExample, "a valid example is refused"},
		{"examples/inventory/valid/a.yaml", "id: x\n", validExample, "a valid example is refused"},
		{"examples/manifest/invalid/a.yaml", "# error: required at #\n" + string(manifest), invalidExample, "passes its schema"},
		{"examples/manifest/invalid/a.yaml", "# error: required at #\nid: &a x\n", invalidExample, "breaks more than its schema"},
		{"examples/inventory/invalid/a.yaml", "# error: required at #\nid: &a x\n", invalidExample, "breaks more than its schema"},
		{"examples/manifest/invalid/a.yaml", "# error: type at #/id\nid: x\n", invalidExample, "gives no type at /id"},
		{"examples/manifest/invalid/a.yaml", "id: x\n", invalidExample, "names no error"},
		{"examples/manifest/invalid/a.yaml", "# error: required at #\n" + "id: x\n", invalidExample, ""},
		{"examples/module/valid/a", "", moduleExample, "an example of a whole module is refused"},
	} {
		fsys := fstest.MapFS{c.path: {Data: []byte(c.data)}}
		if strings.HasPrefix(c.path, "examples/module/") {
			fsys = fstest.MapFS{c.path + "/manifest.yaml": {Data: []byte("id: x\n")}}
		}
		err := c.run(vector{name: c.path, fsys: fsys, path: c.path})
		if c.refused == "" && err != nil || c.refused != "" && (err == nil || !strings.Contains(err.Error(), c.refused)) {
			t.Errorf("%s %q: %v, want %q", c.path, c.data, err, c.refused)
		}
	}
	for _, run := range []func(vector) error{validExample, invalidExample, moduleExample} {
		if run(vector{fsys: fstest.MapFS{}, path: "examples/task/valid/none.yaml"}) == nil {
			t.Error("a case that is not there passes")
		}
	}
}
