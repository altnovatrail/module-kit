package moduleversion

import (
	"encoding/json"
	"io/fs"
	"slices"
	"strings"
	"testing"

	spec "github.com/altnovatrail/spec"

	"github.com/altnovatrail/module-kit/modulepackage"
)

// The rules are those of the schema of the test vectors of versions, in its order.
func TestRulesAreTheFormats(t *testing.T) {
	data, err := fs.ReadFile(spec.Format, "schemas/version-vector.schema.json")
	if err != nil {
		t.Fatal(err)
	}
	var vectorSchema struct {
		Properties struct {
			Refusal struct {
				Enum []string `json:"enum"`
			} `json:"refusal"`
		} `json:"properties"`
	}
	if err := json.Unmarshal(data, &vectorSchema); err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(Rules, vectorSchema.Properties.Refusal.Enum) {
		t.Errorf("the rules are %q, and the schema gives %q", Rules, vectorSchema.Properties.Refusal.Enum)
	}
}

// asPublished gives a folder of spec.Format as a published package: a module built into its package,
// or a package of an earlier version of the format as its files are, by the digest of inventory.json.
func asPublished(t *testing.T, at string) Published {
	t.Helper()
	files, err := fs.Sub(spec.Format, at)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := fs.Stat(files, "manifest.yaml"); err == nil {
		return published(built(t, at))
	}
	inventory, err := fs.ReadFile(files, modulepackage.InventoryFile)
	if err != nil {
		t.Fatal(err)
	}
	return Published{Digest: digestOf(inventory), Files: files}
}

// compareCase compares the new version of a case of vectors/versions/ with its published versions.
func compareCase(t *testing.T, name string, versions ...string) *Comparison {
	t.Helper()
	at := "vectors/versions/" + name
	var published []Published
	for _, v := range versions {
		published = append(published, asPublished(t, at+"/published/"+v))
	}
	c, err := Compare(built(t, at+"/module"), published)
	if err != nil {
		t.Fatalf("%s: %v", name, err)
	}
	return c
}

// Each rule refuses with its words, and the report names it on a line of its own; through a move of
// the format, the part is major, and the rules hold by the places of section 2.
func TestBreaches(t *testing.T) {
	for _, c := range []struct {
		name     string
		versions []string
		want     string
	}{
		{"below-the-newest", []string{"1.0.0", "1.1.0"}, "module-version.above: 1.0.1 is below 1.1.0, which is published: a new version stands above every published one"},
		{"same-version-another-package", []string{"1.0.0"}, "module-version.above: 1.0.0 is published already, with the package sha256:"},
		{"skips-a-patch", []string{"1.0.0"}, "module-version.step: 1.0.2 is not the next version after 1.0.0 by one of its parts"},
		{"patch-step-for-a-minor-change", []string{"1.0.0"}, "module-version.step: 1.0.1 raises the patch part after 1.0.0, and the comparison finds a minor change: "},
		{"skill-dropped-unretired", []string{"1.0.0"}, "module-version.retired-complete: the skill "},
		{"retired-entry-dropped", []string{"2.0.0"}, "module-version.retired-kept: the entry "},
		{"replaced-by-corrected", []string{"2.0.0"}, "module-version.retired-kept: "},
		{"retired-entry-of-no-skill", []string{"1.0.0"}, "module-version.retired-known: the new entry "},
		{"task-id-returns", []string{"1.0.0", "2.0.0"}, "module-version.ids-kept: "},
		{"target-language-added", []string{"1.0.0"}, "module-version.languages-kept: the languages differ from those of 1.0.0"},
		{"band-removed", []string{"1.0.0"}, "module-version.bands-kept: the bands of 1.0.0, "},
		{"format-transition-skill-unretired", []string{"1.0.0"}, "module-version.retired-complete: the skill adding of a published version"},
		{"format-transition-task-id-returns", []string{"0.9.0", "1.0.0"}, "module-version.ids-kept: pick names one of the tasks again"},
	} {
		got := compareCase(t, c.name, c.versions...)
		if len(got.Breaches) != 1 || !strings.HasPrefix(got.Breaches[0].String(), c.want) || !got.Refused() {
			t.Errorf("%s: the breaches are %v, want one that begins %q", c.name, got.Breaches, c.want)
			continue
		}
		if !strings.Contains(got.String(), "\nrefused: "+got.Breaches[0].String()) {
			t.Errorf("%s: the report does not name the breach:\n%s", c.name, got)
		}
	}
	moved := compareCase(t, "format-transition", "1.0.0")
	if moved.Part != Major || moved.Refused() || len(moved.Changes) != 1 || moved.Changes[0].String() !=
		"inventory.json #/module_format_version: major: the module moves from version 0.2 of the format to 0.4" {
		t.Errorf("a move of the format: %s", moved)
	}
	if c := compareCase(t, "earlier-format-older-version", "0.9.0", "1.0.0"); c.Part != Minor || c.Refused() || c.Newest != "1.0.0" {
		t.Errorf("an older version of an earlier format: %s", c)
	}
	if c := compareCase(t, "newest-unchanged", "1.0.0"); !c.Unchanged || c.Refused() {
		t.Errorf("the newest version as it was is refused: %s", c)
	}
}
