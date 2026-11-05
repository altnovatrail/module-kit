package module

import (
	"cmp"
	"context"
	"io/fs"
	"maps"
	"slices"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/altnovatrail/spec"

	"github.com/altnovatrail/module-kit/internal/schema"
	"github.com/altnovatrail/module-kit/internal/yaml"
)

// The rules that a breach of the check may name, and "" for a file that is not YAML.
var knownRules = slices.Concat([]string{"", RuleManifest, RuleLayout, RuleFileIDs, RuleMedia, RuleLicense, RulePlainText, RuleFormatVersion},
	partRules, yaml.Rules)

// wellFormed fails on a report whose breach names a rule the check does not hold, or a place out of
// form, or whose breaches are out of their order.
func wellFormed(t *testing.T, report Report) {
	t.Helper()
	for _, b := range report.Breaches {
		if !slices.Contains(knownRules, b.Rule) || b.Line < 0 || b.Column < 0 || (b.Line == 0) != (b.Column == 0) ||
			b.Path != "" && !fs.ValidPath(b.Path) || b.Pointer != "" && !strings.HasPrefix(b.Pointer, "#") || b.Message == "" {
			t.Errorf("a breach out of form: %#v", b)
		}
		// A breach of a schema names a keyword of section 5 at a value, with no rule and no line.
		if b.Keyword != "" && (b.Rule != "" || b.Line != 0 || !strings.HasPrefix(b.Pointer, "#") ||
			!slices.Contains(schema.Keywords, b.Keyword)) {
			t.Errorf("a breach of a schema out of form: %#v", b)
		}
	}
	if !slices.IsSortedFunc(report.Breaches, func(a, b Breach) int {
		return cmp.Or(strings.Compare(a.Path, b.Path), cmp.Compare(a.Line, b.Line), cmp.Compare(a.Column, b.Column),
			schema.ComparePointers(a.Pointer, b.Pointer))
	}) {
		t.Errorf("the breaches are out of order:\n%s", report)
	}
}

// On any file at any path of a module, and on any file of a part on its own, the check gives a
// report without an error or a panic, and the report is in form. The corpus is the files of the
// vectors of the rules that the check holds but the YAML rules, whose vectors FuzzRead takes, and
// task.template-solver, each at its path in its case or, for a case in a file, at a path of its part,
// and the examples of the parts of the layout, each at a path of its part. The vectors of
// task.template-solver are left out because four of them run a solver to its budget of 100 000 000
// steps, a second under the race detector, and each run of the fuzz test runs it twice; the example
// number-thought-of-template brings a solver into the corpus.
func FuzzCheck(f *testing.F) {
	err := fs.WalkDir(spec.Format, "vectors/rules", func(name string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() {
			return err
		}
		parts := strings.SplitN(name, "/", 5)
		if !slices.Contains(knownRules, parts[2]) || slices.Contains(yaml.Rules, parts[2]) || parts[2] == RuleTaskTemplateSolver {
			return nil
		}
		data, err := fs.ReadFile(spec.Format, name)
		if err != nil {
			return err
		}
		at := ""
		if len(parts) == 5 {
			at = parts[4]
		}
		for _, r := range layout {
			if part, _, _ := strings.Cut(parts[2], "."); at == "" && r.part == part {
				at = strings.Replace(r.path, anyID, "a", 1)
			}
		}
		f.Add(at, data)
		return nil
	})
	if err != nil {
		f.Fatal(err)
	}
	for name, part := range partFiles(f) {
		if !strings.HasPrefix(name, "examples/") {
			continue
		}
		data, err := fs.ReadFile(spec.Format, name)
		if err != nil {
			f.Fatal(err)
		}
		for _, r := range layout {
			if r.part == part[0] {
				f.Add(strings.Replace(r.path, anyID, "a", 1), data)
			}
		}
	}
	f.Add("media/sounds/bell.mp3", []byte{0xff, 0xfb, 0x90})
	f.Add("tasks/a.yaml", []byte("id: b\n"))
	f.Fuzz(func(t *testing.T, name string, data []byte) {
		fsys := smallest(fstest.MapFS{noticePath: bytesFile(data)})
		// A path that a file of the module would also make a folder is left out: the file system
		// would hold both.
		folderOfFile := func(k string) bool { return strings.HasPrefix(name, k+"/") || strings.HasPrefix(k, name+"/") }
		if fs.ValidPath(name) && name != "." && !slices.ContainsFunc(slices.Collect(maps.Keys(fsys)), folderOfFile) {
			fsys[name] = bytesFile(data)
		}
		report, err := Check(context.Background(), fsys)
		if err != nil {
			t.Fatal(err)
		}
		wellFormed(t, report)
		for _, r := range layout {
			if r.part == "" {
				continue
			}
			report, err := CheckFile(r.part, data)
			if err != nil {
				t.Fatal(err)
			}
			wellFormed(t, report)
		}
	})
}
