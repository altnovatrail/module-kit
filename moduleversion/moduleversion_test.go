package moduleversion

import (
	"encoding/json"
	"errors"
	"io/fs"
	"slices"
	"strings"
	"testing"
	"testing/fstest"

	spec "github.com/altnovatrail/spec"

	"github.com/altnovatrail/module-kit/modulepackage"
)

// The parts are the words of version_part in the schema of the test vectors of versions.
func TestPartsAreTheFormats(t *testing.T) {
	data, err := fs.ReadFile(spec.Format, "schemas/version-vector.schema.json")
	if err != nil {
		t.Fatal(err)
	}
	var vectorSchema struct {
		Properties struct {
			VersionPart struct {
				Enum []string `json:"enum"`
			} `json:"version_part"`
		} `json:"properties"`
	}
	if err := json.Unmarshal(data, &vectorSchema); err != nil {
		t.Fatal(err)
	}
	var words []string
	for p := None; p <= Major; p++ {
		text, err := p.MarshalText()
		if err != nil {
			t.Fatal(err)
		}
		words = append(words, string(text))
	}
	slices.Sort(words)
	if want := slices.Sorted(slices.Values(vectorSchema.Properties.VersionPart.Enum)); !slices.Equal(words, want) {
		t.Errorf("the parts are %q, and the schema gives %q", words, want)
	}
	var p Part
	if err := p.UnmarshalText([]byte("minor")); err != nil || p != Minor || p.UnmarshalText([]byte("Minor")) == nil {
		t.Errorf("UnmarshalText reads %s, %v", p, err)
	}
	if _, err := Part(4).MarshalText(); err == nil || Part(-1).String() != "Part(-1)" {
		t.Errorf("a part out of range is written")
	}
}

// The next versions and their parts are those of section 27, X.Y.Z, 0.y.z and 0.0.z each.
func TestSteps(t *testing.T) {
	for from, want := range map[string]string{
		"1.2.3":  "1.2.4 patch, 1.3.0 minor, 2.0.0 major",
		"0.2.0":  "0.2.1 minor, 0.3.0 major, 1.0.0 major",
		"0.0.1":  "0.0.2 major, 0.1.0 major, 1.0.0 major",
		"0.0.0":  "0.0.1 major, 0.1.0 major, 1.0.0 major",
		"9.9.99": "9.9.100 patch, 9.10.0 minor, 10.0.0 major",
	} {
		steps, err := Steps(from)
		if err != nil {
			t.Fatal(err)
		}
		var got []string
		for _, s := range steps {
			got = append(got, s.Version+" "+s.Part.String())
		}
		if strings.Join(got, ", ") != want {
			t.Errorf("Steps(%s) = %s, want %s", from, strings.Join(got, ", "), want)
		}
	}
	if _, err := Steps("1.2"); err == nil {
		t.Error("Steps(1.2) gives steps")
	}
}

// StepPart gives the part of a next version, and None to a version that is none.
func TestStepPart(t *testing.T) {
	for _, c := range []struct {
		from, to string
		want     Part
	}{
		{"1.0.0", "1.0.1", Patch}, {"1.0.0", "1.1.0", Minor}, {"1.0.0", "2.0.0", Major}, {"1.0.3", "1.1.3", None},
		{"1.0.0", "1.0.2", None}, {"1.0.0", "1.0.0", None}, {"1.1.0", "1.0.0", None}, {"0.2.0", "0.2.1", Minor},
		{"0.2.0", "1.0.0", Major}, {"0.0.1", "0.0.2", Major}, {"0.0.1", "0.0.3", None},
	} {
		if got, err := StepPart(c.from, c.to); err != nil || got != c.want {
			t.Errorf("StepPart(%s, %s) = %s, %v, want %s", c.from, c.to, got, err, c.want)
		}
	}
	for _, c := range [][2]string{{"1.0", "1.0.1"}, {"1.0.0", "v1.0.1"}} {
		if _, err := StepPart(c[0], c[1]); err == nil {
			t.Errorf("StepPart(%s, %s) gives no error", c[0], c[1])
		}
	}
}

// The same major is one X from 1 on, one 0.y for y from 1 on, and each 0.0.z alone.
func TestSameMajor(t *testing.T) {
	for _, c := range []struct {
		a, b string
		want bool
	}{
		{"1.0.0", "1.9.3", true}, {"1.0.0", "2.0.0", false}, {"0.2.0", "0.2.7", true}, {"0.2.0", "0.3.0", false},
		{"0.2.0", "1.2.0", false}, {"0.0.1", "0.0.1", true}, {"0.0.1", "0.0.2", false}, {"0.0.1", "0.1.1", false},
	} {
		for _, pair := range [][2]string{{c.a, c.b}, {c.b, c.a}} {
			if got, err := SameMajor(pair[0], pair[1]); err != nil || got != c.want {
				t.Errorf("SameMajor(%s, %s) = %t, %v, want %t", pair[0], pair[1], got, err, c.want)
			}
		}
	}
	for _, c := range [][2]string{{"1.0", "1.0.1"}, {"1.0.0", "1"}} {
		if _, err := SameMajor(c[0], c[1]); err == nil {
			t.Errorf("SameMajor(%s, %s) gives no error", c[0], c[1])
		}
	}
}

// Each change stands at its place, in the new version or in the published one for what is gone, an
// instance's with its parameters, and a difference that a row of the table names is not given again.
func TestChanges(t *testing.T) {
	for name, want := range map[string][]string{
		"instance-key-changed": {
			"server.json #/tasks/times/items/0/line/1: major: on the instance a = 2, b = 2: what decides the verdict or the trap of answer slot 1 changes",
			"server.json #/tasks/times/items/0/line/1: major: on the instance a = 2, b = 3: what decides the verdict or the trap of answer slot 1 changes",
			"server.json #/tasks/times/items/0/line/1: major: on the instance a = 3, b = 2: what decides the verdict or the trap of answer slot 1 changes",
			"server.json #/tasks/times/items/0/line/1: major: on the instance a = 3, b = 3: what decides the verdict or the trap of answer slot 1 changes",
			"server.json #/tasks/times/items/0/line/0/text: patch: changes",
			"server.json #/tasks/times/items/0/line/2: patch: is removed",
		},
		"units-reordered-and-recased": {
			"server.json #/tasks/apples/items/0/answer_slots/0/check/units/0: patch: changes",
			"server.json #/tasks/apples/items/0/answer_slots/0/check/units/1: patch: changes",
		},
		"band-after-the-last": {
			"server.json #/manifest/audience/grades: minor: the grades change",
			"server.json #/manifest/bands/2: minor: the band grades-5-6 comes after the last",
		},
		"task-removed":       {"server.json #/tasks/pick: major: the task pick is removed"},
		"media-added":        {"device.json #/media: patch: is added"},
		"zone-moved":         {"server.json #/tasks/two-sums/items/1/answer_slots/0/zone/top: patch: changes"},
		"prerequisite-added": {"server.json #/skill_graph/skills/subtracting/prerequisites: minor: the prerequisites of subtracting change"},
		"constraint-added": {
			"server.json #/instances/times: minor: on the instance a = 2, b = 2: the instance is no longer one",
			"server.json #/instances/times: minor: on the instance a = 3, b = 3: the instance is no longer one",
			"server.json #/tasks/times/template/constraint: patch: is added",
		},
	} {
		next, before := versionCase(t, name, "1.0.0")
		_, changes, err := Changes(before, next)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		var got []string
		for _, c := range changes {
			got = append(got, c.String())
		}
		if !slices.Equal(got, want) {
			t.Errorf("%s: the changes are\n%s\nwant\n%s", name, strings.Join(got, "\n"), strings.Join(want, "\n"))
		}
	}
	next, before := versionCase(t, "only-the-version", "1.0.0")
	if part, changes, err := Changes(before, next); err != nil || part != None || len(changes) != 0 {
		t.Errorf("only the version: %s, %v, %v", part, changes, err)
	}
	other := built(t, "examples/module/valid/smallest")
	if _, _, err := Changes(before, other); err == nil {
		t.Error("two modules compare")
	}
	moved := *next
	moved.Inventory.ModuleFormatVersion = "0.2"
	if _, _, err := Changes(before, &moved); err == nil {
		t.Error("two versions of the format compare")
	}
	broken := *next
	broken.Files = map[string][]byte{modulepackage.ServerFile: []byte("{")}
	if _, _, err := Changes(before, &broken); err == nil {
		t.Error("a package with no canonical view compares")
	}
}

// Compare takes the newest published version by Semantic Versioning, gives a module that has not
// changed and a first version nothing to compare, and refuses what it cannot compare.
func TestCompare(t *testing.T) {
	next := built(t, "vectors/versions/after-a-withdrawn-version/module")
	older := built(t, "vectors/versions/after-a-withdrawn-version/published/1.0.0")
	newest := built(t, "vectors/versions/after-a-withdrawn-version/published/1.1.0")
	c, err := Compare(next, []Published{published(newest), published(older)})
	if err != nil {
		t.Fatal(err)
	}
	if c.Newest != "1.1.0" || !slices.Equal(c.Published, []string{"1.0.0", "1.1.0"}) || c.Part != Major || c.Raised != Major {
		t.Errorf("the comparison is %+v", c)
	}
	if c, err := Compare(newest, []Published{published(older), published(newest)}); err != nil || !c.Unchanged || c.Part != None || len(c.Changes) > 0 {
		t.Errorf("the newest version as it was: %+v, %v", c, err)
	}
	if c, err := Compare(next, nil); err != nil || c.Newest != "" || len(c.Published) != 0 {
		t.Errorf("a first version: %+v, %v", c, err)
	}
	if _, err := Compare(nil, nil); err == nil {
		t.Error("no new version compares")
	}
	twice := published(older)
	if _, err := Compare(next, []Published{published(older), twice}); err == nil {
		t.Error("two packages of one version compare")
	}
	if _, err := Compare(next, []Published{published(built(t, "examples/module/valid/smallest"))}); err == nil {
		t.Error("a package of another module compares")
	}
	if _, err := Compare(next, []Published{{Digest: newest.Digest, Files: fstest.MapFS{}}}); !errors.Is(err, modulepackage.ErrChain) {
		t.Errorf("a missing package: %v; want ErrChain", err)
	}
	earlier, err := fs.Sub(spec.Format, "vectors/versions/format-transition/published/1.0.0")
	if err != nil {
		t.Fatal(err)
	}
	inventory, err := fs.ReadFile(earlier, modulepackage.InventoryFile)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Compare(next, []Published{{Digest: digestOf(inventory), Files: earlier}}); !errors.Is(err, modulepackage.ErrFormatVersion) {
		t.Errorf("a package of an earlier version of the format: %v; want ErrFormatVersion", err)
	}
}

// The report writes the versions, the part found, the next versions and the part raised, and a line
// for each change; --json gives the same data.
func TestComparisonString(t *testing.T) {
	next, before := versionCase(t, "skips-a-patch", "1.0.0")
	c, err := Compare(next, []Published{published(before)})
	if err != nil {
		t.Fatal(err)
	}
	head := strings.Join([]string{
		"example/versions 1.0.2, compared with 1.0.0",
		"published: 1.0.0",
		"found: patch",
		"next versions after 1.0.0: 1.0.1 (patch), 1.1.0 (minor), 2.0.0 (major)",
		"1.0.2 is not the next version after 1.0.0 by one of its parts",
	}, "\n")
	if got := c.String(); !strings.HasPrefix(got, head+"\n") {
		t.Errorf("the report is\n%s\nwant it to begin with\n%s", got, head)
	}
	data, err := json.Marshal(c)
	if err != nil {
		t.Fatal(err)
	}
	var back struct {
		Module, Version, Newest, Raised string
		Part                            string `json:"version_part"`
		Next                            []Step `json:"next_versions"`
		Changes                         []map[string]any
	}
	if err := json.Unmarshal(data, &back); err != nil {
		t.Fatal(err)
	}
	if back.Module != "example/versions" || back.Version != "1.0.2" || back.Newest != "1.0.0" || back.Part != "patch" || back.Raised != "none" ||
		len(back.Next) != 3 || len(back.Changes) != len(c.Changes) || back.Changes[0]["file"] != modulepackage.ServerFile {
		t.Errorf("--json gives %s", data)
	}
	for _, c := range []*Comparison{{Module: "m", Version: "0.1.0", Published: []string{}}, {Module: "m", Version: "1.0.0", Newest: "1.0.0", Unchanged: true}} {
		if s := c.String(); strings.Contains(s, "\n") || !strings.HasPrefix(s, "m ") {
			t.Errorf("the report of nothing compared is %q", s)
		}
	}
}
