package module

import (
	"context"
	"encoding/json"
	"io/fs"
	"maps"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/altnovatrail/spec"

	"github.com/altnovatrail/module-kit/internal/formattree"
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

// formatText reads the rows of a table of the format's text, README.md of spec.Format, whose header
// begins with the cells of header, each row by its cells with the backquotes of code taken out.
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

// The layout is the format's table of the layout: each path, with its part, in order.
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

// The rules of a whole module are those of the format's text, in its order, the format version is a
// rule of the manifest, and the text lists every rule of the parts that the check holds.
func TestRulesAreTheFormats(t *testing.T) {
	var module, manifest, all []string
	for _, cells := range formatText(t, "| ID | Rule |") {
		all = append(all, cells[0])
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
	for _, rule := range partRules {
		if !slices.Contains(all, rule) {
			t.Errorf("format/README.md lists no rule %s", rule)
		}
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

// Every module laid out in spec.Format and every file of a part breaks nothing that the check holds
// but the rule it is a case of: each of them breaks that rule and nothing else, its files are valid
// against their schemas, and a valid example and a module that is no case break nothing at all. An
// invalid example breaks its schema alone. That a case breaks its rule, and an invalid example the
// keyword it names, the test of vectors checks. Each module and each file is a subtest of its own,
// named by its path in spec.Format, and they run in parallel: they share nothing but spec.Format,
// which is read-only.
func TestChecksTheFormat(t *testing.T) {
	// beside says whether a breach is other than the one its case may have: of its rule, or, for an
	// invalid example, of its schema.
	beside := func(b Breach, rule string, invalid bool) bool {
		if invalid {
			return b.Keyword == ""
		}
		return rule == "" || b.Rule != rule || b.Keyword != ""
	}
	folders, err := formattree.Modules(spec.Format)
	if err != nil {
		t.Fatal(err)
	}
	files, err := formattree.PartFiles(spec.Format)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%d modules and %d files of a part", len(folders), len(files))
	for _, name := range slices.Sorted(maps.Keys(folders)) {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			folder, err := fs.Sub(spec.Format, name)
			if err != nil {
				t.Fatal(err)
			}
			report, err := Check(context.Background(), folder)
			if err != nil {
				t.Fatal(err)
			}
			for _, b := range report.Breaches {
				if beside(b, folders[name], false) {
					t.Error(b)
				}
			}
		})
	}
	for _, name := range slices.Sorted(maps.Keys(files)) {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			data, err := fs.ReadFile(spec.Format, name)
			if err != nil {
				t.Fatal(err)
			}
			report, err := CheckFile(files[name].Part, data)
			if err != nil {
				t.Fatal(err)
			}
			invalid := files[name].Kind == formattree.KindInvalidExample
			for _, b := range report.Breaches {
				if beside(b, files[name].Rule, invalid) {
					t.Error(b)
				}
			}
			if invalid && len(report.Breaches) == 0 {
				t.Error("an invalid example breaks nothing")
			}
		})
	}
}
