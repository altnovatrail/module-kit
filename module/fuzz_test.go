package module

import (
	"cmp"
	"io/fs"
	"maps"
	"slices"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/altnovatrail/spec"

	"github.com/altnovatrail/module-kit/internal/yaml"
)

// The rules that a breach of the check may name, and "" for a file that is not YAML.
var knownRules = slices.Concat([]string{"", RuleManifest, RuleLayout, RuleFileIDs, RuleMedia, RuleLicense, RulePlainText, RuleFormatVersion}, yaml.Rules)

// wellFormed fails on a report whose breach names a rule the check does not hold, or a place out of
// form, or whose breaches are out of their order.
func wellFormed(t *testing.T, report Report) {
	t.Helper()
	for _, b := range report.Breaches {
		if !slices.Contains(knownRules, b.Rule) || b.Line < 0 || b.Column < 0 || (b.Line == 0) != (b.Column == 0) ||
			b.Path != "" && !fs.ValidPath(b.Path) || b.Pointer != "" && !strings.HasPrefix(b.Pointer, "#") || b.Message == "" {
			t.Errorf("a breach out of form: %#v", b)
		}
	}
	if !slices.IsSortedFunc(report.Breaches, func(a, b Breach) int {
		return cmp.Or(strings.Compare(a.Path, b.Path), cmp.Compare(a.Line, b.Line), cmp.Compare(a.Column, b.Column))
	}) {
		t.Errorf("the breaches are out of order:\n%s", report)
	}
}

// On any file at any path of a module, and on any file of a part on its own, the check gives a
// report without an error or a panic, and the report is in form. The corpus is the files of the
// vectors of the rules module.* and manifest.format-version, each at its path in its case.
func FuzzCheck(f *testing.F) {
	err := fs.WalkDir(spec.Format, "vectors/rules", func(name string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() {
			return err
		}
		parts := strings.SplitN(name, "/", 5)
		if !strings.HasPrefix(parts[2], "module.") && parts[2] != RuleFormatVersion {
			return nil
		}
		data, err := fs.ReadFile(spec.Format, name)
		if err != nil {
			return err
		}
		at := manifestPath
		if len(parts) == 5 {
			at = parts[4]
		}
		f.Add(at, data)
		return nil
	})
	if err != nil {
		f.Fatal(err)
	}
	f.Add("media/sounds/bell.mp3", []byte{0xff, 0xfb, 0x90})
	f.Add("tasks/a.yaml", []byte("id: b\n"))
	f.Fuzz(func(t *testing.T, name string, data []byte) {
		fsys := module(fstest.MapFS{noticePath: bytesFile(data)})
		// A path that a file of the module would also make a folder is left out: the file system
		// would hold both.
		folderOfFile := func(k string) bool { return strings.HasPrefix(name, k+"/") || strings.HasPrefix(k, name+"/") }
		if fs.ValidPath(name) && name != "." && !slices.ContainsFunc(slices.Collect(maps.Keys(fsys)), folderOfFile) {
			fsys[name] = bytesFile(data)
		}
		report, err := Check(fsys)
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
