package module

import (
	"encoding/json"
	"io/fs"
	"maps"
	"path"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/altnovatrail/spec"
)

// What the check takes from the format as code is the format's: a move to a new tag of spec that
// changes it fails these tests until the check follows.

// The version of the format the library reads is the version of spec.Format.
func TestFormatVersionIsSpecs(t *testing.T) {
	data, err := fs.ReadFile(spec.Format, "VERSION")
	if err != nil {
		t.Fatal(err)
	}
	if v := strings.TrimSpace(string(data)); v != FormatVersion {
		t.Errorf("spec.Format is of version %s, and FormatVersion is %s", v, FormatVersion)
	}
}

// formatText reads the rows of a table of format/README.md whose header begins with the cells of
// header, each row by its cells with the backquotes of code taken out.
func formatText(t *testing.T, header string) [][]string {
	t.Helper()
	data, err := fs.ReadFile(spec.Format, "README.md")
	if err != nil {
		t.Fatal(err)
	}
	var rows [][]string
	in := false
	for line := range strings.Lines(string(data)) {
		line = strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(line, header):
			in = true
		case in && strings.HasPrefix(line, "|---"):
		case in && strings.HasPrefix(line, "|"):
			var cells []string
			for _, cell := range strings.Split(strings.Trim(line, "|"), "|") {
				cells = append(cells, strings.ReplaceAll(strings.TrimSpace(cell), "`", ""))
			}
			rows = append(rows, cells)
		default:
			in = false
		}
	}
	return rows
}

// The layout is the table of section 10: each path, with its part, in order.
func TestLayoutIsTheFormats(t *testing.T) {
	var text []row
	for _, cells := range formatText(t, "| Path | Part |") {
		part := cells[1]
		if part == "—" {
			part = ""
		}
		text = append(text, row{cells[0], part})
	}
	if !slices.Equal(text, layout) {
		t.Errorf("the layout of format/README.md is\n%v\nand the check's\n%v", text, layout)
	}
}

// The rules of a whole module are those of section 10, in its order, and the format version is a
// rule of the manifest.
func TestRulesAreTheFormats(t *testing.T) {
	var module, manifest []string
	for _, cells := range formatText(t, "| ID | Rule |") {
		switch {
		case strings.HasPrefix(cells[0], "module."):
			module = append(module, cells[0])
		case strings.HasPrefix(cells[0], "manifest."):
			manifest = append(manifest, cells[0])
		}
	}
	if want := []string{RuleManifest, RuleLayout, RuleFileIDs, RuleMedia, RuleLicense, RulePlainText}; !slices.Equal(module, want) {
		t.Errorf("format/README.md lists the rules %v, and the check %v", module, want)
	}
	if !slices.Contains(manifest, RuleFormatVersion) {
		t.Errorf("format/README.md lists no %s among %v", RuleFormatVersion, manifest)
	}
}

// The forms by which the check reads names are those of common.schema.json, with nothing more.
func TestFormsAreTheSchemas(t *testing.T) {
	data, err := fs.ReadFile(spec.Format, "schemas/common.schema.json")
	if err != nil {
		t.Fatal(err)
	}
	var common struct {
		Defs map[string]map[string]json.RawMessage `json:"$defs"`
	}
	if err := json.Unmarshal(data, &common); err != nil {
		t.Fatal(err)
	}
	for name, f := range map[string]form{"local_id": localID, "media_path": mediaPath, "module_format_version": {pattern: versionForm}} {
		def := common.Defs[name]
		var schema struct {
			Type      string
			Pattern   string
			MaxLength *int
			Not       *struct{ Pattern string }
		}
		raw, _ := json.Marshal(def)
		if err := json.Unmarshal(raw, &schema); err != nil {
			t.Fatal(err)
		}
		keys := slices.Sorted(maps.Keys(def))
		allowed := []string{"description", "maxLength", "not", "pattern", "type"}
		if schema.Type != "string" || slices.ContainsFunc(keys, func(k string) bool { return !slices.Contains(allowed, k) }) {
			t.Errorf("$defs/%s holds %v, and the check reads a string by pattern, maxLength and not alone", name, keys)
		}
		if schema.Pattern != f.pattern.String() {
			t.Errorf("$defs/%s: the pattern is %s, and the check's %s", name, schema.Pattern, f.pattern)
		}
		if schema.MaxLength == nil && f.maxLength != 0 || schema.MaxLength != nil && *schema.MaxLength != f.maxLength {
			t.Errorf("$defs/%s: maxLength is %v, and the check's %d", name, schema.MaxLength, f.maxLength)
		}
		if (schema.Not == nil) != (f.not == nil) || schema.Not != nil && schema.Not.Pattern != f.not.String() {
			t.Errorf("$defs/%s: not is %v, and the check's %v", name, schema.Not, f.not)
		}
	}
	// The check knows the signature of each extension that media_path takes, and of no other.
	match := regexp.MustCompile(`\\\.\(([a-z0-9|]+)\)\$$`).FindStringSubmatch(mediaPath.pattern.String())
	if match == nil {
		t.Fatalf("the pattern of media_path ends in no group of extensions: %s", mediaPath.pattern)
	}
	extensions := strings.Split(match[1], "|")
	slices.Sort(extensions)
	if known := slices.Sorted(maps.Keys(signatures)); !slices.Equal(extensions, known) {
		t.Errorf("media_path takes the extensions %v, and the check knows the signatures of %v", extensions, known)
	}
}

// moduleFolders finds the folders of spec.Format laid out as a module (format/README.md, sections
// 8 and 9), each with the rule it is a case of, or "" for a module that breaks no rule.
func moduleFolders(t *testing.T) map[string]string {
	t.Helper()
	folders := map[string]string{}
	err := fs.WalkDir(spec.Format, ".", func(name string, entry fs.DirEntry, err error) error {
		if err != nil || !entry.IsDir() {
			return err
		}
		parts := strings.Split(name, "/")
		laidOut := func() bool {
			_, err := fs.Stat(spec.Format, path.Join(name, manifestPath))
			return err == nil
		}
		switch {
		case len(parts) == 4 && parts[0] == "examples" && parts[1] == "module" && parts[2] == "valid":
			folders[name] = ""
		case len(parts) == 4 && parts[0] == "vectors" && parts[1] == "rules":
			folders[name] = parts[2]
		case len(parts) == 4 && parts[0] == "vectors" && (parts[1] == "packages" || parts[1] == "versions") && parts[3] == "module",
			len(parts) == 5 && parts[0] == "vectors" && parts[1] == "indexes" && parts[3] == "modules",
			// A published version of an earlier format is the files of its package, and no module.
			len(parts) == 5 && parts[0] == "vectors" && parts[1] == "versions" && parts[3] == "published" && laidOut():
			folders[name] = ""
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return folders
}

// partFiles finds the files of spec.Format that are files of a part on their own: the examples of
// the parts of the layout, the cases in a file of a part's rule and the parts of the vectors of the
// open view. Each with its part, and the rule it is a case of, or "".
func partFiles(t *testing.T) map[string][2]string {
	t.Helper()
	files := map[string][2]string{}
	err := fs.WalkDir(spec.Format, ".", func(name string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() {
			return err
		}
		parts := strings.Split(name, "/")
		inLayout := func(part string) bool {
			return part != "" && slices.ContainsFunc(layout, func(r row) bool { return r.part == part })
		}
		switch {
		case len(parts) == 4 && parts[0] == "examples" && inLayout(parts[1]):
			files[name] = [2]string{parts[1], ""}
		case len(parts) == 4 && parts[0] == "vectors" && parts[1] == "rules" && !strings.HasPrefix(parts[2], "yaml."):
			part, _, _ := strings.Cut(parts[2], ".")
			files[name] = [2]string{part, parts[2]}
		case len(parts) == 5 && parts[0] == "vectors" && parts[1] == "open-view" && parts[4] == "part.yaml":
			files[name] = [2]string{parts[2], ""}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return files
}

// Every module laid out in spec.Format and every file of a part breaks no rule that the check holds,
// but the rule it is a case of: each of them breaks that rule and nothing else (format/README.md,
// sections 8 and 9). That a case breaks its rule, the test of vectors checks.
func TestChecksTheFormat(t *testing.T) {
	folders := moduleFolders(t)
	for _, name := range slices.Sorted(maps.Keys(folders)) {
		folder, err := fs.Sub(spec.Format, name)
		if err != nil {
			t.Fatal(err)
		}
		report, err := Check(folder)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		for _, b := range report.Breaches {
			if b.Rule != folders[name] {
				t.Errorf("%s: %s", name, b)
			}
		}
	}
	files := partFiles(t)
	for _, name := range slices.Sorted(maps.Keys(files)) {
		data, err := fs.ReadFile(spec.Format, name)
		if err != nil {
			t.Fatal(err)
		}
		report, err := CheckFile(files[name][0], data)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		for _, b := range report.Breaches {
			if b.Rule != files[name][1] {
				t.Errorf("%s: %s", name, b)
			}
		}
	}
	t.Logf("%d modules and %d files of a part", len(folders), len(files))
}
