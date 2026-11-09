package moduleindex

import (
	"context"
	"encoding/json"
	"errors"
	"io/fs"
	"reflect"
	"slices"
	"strings"
	"sync"
	"testing"

	"github.com/altnovatrail/spec"
	"github.com/altnovatrail/spec/types/formatv0"

	"github.com/altnovatrail/module-kit/internal/canonical"
	"github.com/altnovatrail/module-kit/internal/schema"
	"github.com/altnovatrail/module-kit/modulepackage"
)

// The packages of the modules of the vectors of the index, built once, by the vector's case and the
// module's folder in it.
var vectorPackages sync.Map

// pkg gives the package of the module of a vector of the index: modules/<name>/ of the case.
func pkg(t testing.TB, vectorCase, name string) *modulepackage.Package {
	t.Helper()
	at := "vectors/indexes/" + vectorCase + "/modules/" + name
	if p, ok := vectorPackages.Load(at); ok {
		return p.(*modulepackage.Package)
	}
	folder, err := fs.Sub(spec.Format, at)
	if err != nil {
		t.Fatal(err)
	}
	p, err := modulepackage.Build(context.Background(), folder)
	if err != nil {
		t.Fatal(err)
	}
	vectorPackages.Store(at, p)
	return p
}

// The packages of the vectors that the tests build indexes of.
func smallest(t testing.TB) *modulepackage.Package    { return pkg(t, "first-release", "smallest") }
func timesTables(t testing.TB) *modulepackage.Package { return pkg(t, "first-release", "times-tables") }
func timesTables020(t testing.TB) *modulepackage.Package {
	return pkg(t, "next-release", "times-tables")
}

// textOf gives the canonical text of an index, as a release of modules holds it.
func textOf(t testing.TB, index formatv0.ModuleIndex) []byte {
	t.Helper()
	text, err := canonical.Text(schema.ValueOf(index))
	if err != nil {
		t.Fatal(err)
	}
	return text
}

// mustRead reads the text of an index.
func mustRead(t testing.TB, text []byte) formatv0.ModuleIndex {
	t.Helper()
	index, err := Read(text)
	if err != nil {
		t.Fatal(err)
	}
	return index
}

func reason(r formatv0.WithdrawalReason) *formatv0.WithdrawalReason { return &r }

// The rules of the index are those whose refusals the vectors of the index name, in the order of
// the table of section 26.
func TestRulesAreTheFormats(t *testing.T) {
	data, err := fs.ReadFile(spec.Format, "schemas/index-vector.schema.json")
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
		t.Errorf("Rules = %v; the refusals of the vectors of the index are %v", Rules, vectorSchema.Properties.Refusal.Enum)
	}
}

// The first release names each of its packages by its digest, the release's tag and the anchor
// skills it teaches, each once and in order; the order of the packages and a package given twice
// change nothing, and the index is the canonical text of what it names.
func TestBuildFirst(t *testing.T) {
	want := formatv0.ModuleIndex{ModuleFormatVersion: "0.4", Modules: map[formatv0.ModuleID]map[formatv0.ModuleVersion]formatv0.Entry{
		"example/smallest": {"0.1.0": {ModuleFormatVersion: "0.4", PackageDigest: smallest(t).Digest, Release: "v0.1.0"}},
		"example/times-tables": {"0.1.0": {ModuleFormatVersion: "0.4", PackageDigest: timesTables(t).Digest, Release: "v0.1.0",
			Teaches: []formatv0.AnchorSkillID{"core.division-with-remainder", "core.multiplication-table"}}},
	}}
	for _, packages := range [][]*modulepackage.Package{
		{smallest(t), timesTables(t)},
		{timesTables(t), smallest(t)},
		{timesTables(t), smallest(t), timesTables(t)},
	} {
		text, err := Build(Release{First: true, Tag: "v0.1.0", Packages: packages})
		if err != nil {
			t.Fatal(err)
		}
		if got := string(text); got != string(textOf(t, want)) {
			t.Errorf("Build = %s; want %s", got, textOf(t, want))
		}
	}
}

// A later release keeps every entry of the previous index as it was, an entry of an earlier format
// version among them; a package of a version that it names keeps that entry, its release among
// it; a new version gets an entry that names the release, even one below another new version of
// the release; and a withdrawal marks the entry, one of a version new in the release too, while
// one for the reason it has already changes nothing.
func TestBuildNext(t *testing.T) {
	previous := formatv0.ModuleIndex{ModuleFormatVersion: "0.4", Modules: map[formatv0.ModuleID]map[formatv0.ModuleVersion]formatv0.Entry{
		"example/smallest": {"0.1.0": {ModuleFormatVersion: "0.4", PackageDigest: smallest(t).Digest, Release: "v0.1.0", Withdrawn: reason("answer_leak")}},
		"example/old":      {"1.0.0": {ModuleFormatVersion: "0.1", PackageDigest: timesTables(t).Digest, Release: "v0.1.0"}},
	}}
	got, err := Build(Release{
		Previous:    textOf(t, previous),
		Tag:         "v0.4.0",
		Packages:    []*modulepackage.Package{timesTables020(t), smallest(t), timesTables(t)},
		Withdrawals: formatv0.Withdrawals{"example/smallest": {"0.1.0": "answer_leak"}, "example/times-tables": {"0.1.0": "wrong_key"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	teaches := []formatv0.AnchorSkillID{"core.division-with-remainder", "core.multiplication-table"}
	want := formatv0.ModuleIndex{ModuleFormatVersion: "0.4", Modules: map[formatv0.ModuleID]map[formatv0.ModuleVersion]formatv0.Entry{
		"example/smallest": previous.Modules["example/smallest"],
		"example/old":      previous.Modules["example/old"],
		"example/times-tables": {
			"0.1.0": {ModuleFormatVersion: "0.4", PackageDigest: timesTables(t).Digest, Release: "v0.4.0", Teaches: teaches, Withdrawn: reason("wrong_key")},
			"0.2.0": {ModuleFormatVersion: "0.4", PackageDigest: timesTables020(t).Digest, Release: "v0.4.0", Teaches: teaches},
		},
	}}
	if string(got) != string(textOf(t, want)) {
		t.Errorf("Build = %s; want %s", got, textOf(t, want))
	}
}

// Each rule refuses the build with a breach at the version that breaks it, and a release that
// breaks several gets every breach, in the order of the module, the version and the rule.
func TestBuildRefuses(t *testing.T) {
	previous := textOf(t, formatv0.ModuleIndex{ModuleFormatVersion: "0.4", Modules: map[formatv0.ModuleID]map[formatv0.ModuleVersion]formatv0.Entry{
		"example/smallest": {"0.1.0": {ModuleFormatVersion: "0.4", PackageDigest: smallest(t).Digest, Release: "v0.1.0", Withdrawn: reason("wrong_key")}},
		"example/times-tables": {
			"0.1.0":  {ModuleFormatVersion: "0.4", PackageDigest: timesTables(t).Digest, Release: "v0.1.0"},
			"0.1.10": {ModuleFormatVersion: "0.4", PackageDigest: timesTables(t).Digest, Release: "v0.2.0", Withdrawn: reason("answer_leak")},
			"0.2.0":  {ModuleFormatVersion: "0.3", PackageDigest: timesTables020(t).Digest, Release: "v0.2.0"},
		},
	}})
	republished := pkg(t, "republished-version", "times-tables")
	below := pkg(t, "below-a-published-version", "times-tables")
	cases := []struct {
		name    string
		release Release
		want    []string
	}{
		{"another reason", Release{Previous: previous, Tag: "v0.3.0", Withdrawals: formatv0.Withdrawals{"example/smallest": {"0.1.0": "answer_leak"}}},
			[]string{"module-index.entries-kept: example/smallest 0.1.0: it is withdrawn for wrong_key, and the release withdraws it for answer_leak: a withdrawal keeps its reason"}},
		{"republished", Release{Previous: previous, Tag: "v0.3.0", Packages: []*modulepackage.Package{republished}},
			[]string{"module-index.one-package: example/times-tables 0.1.0: the previous index names its package " + string(timesTables(t).Digest) +
				", and the release builds " + string(republished.Digest) + ": a version is published once"}},
		{"two packages of a version", Release{First: true, Tag: "v0.1.0", Packages: []*modulepackage.Package{republished, timesTables(t)}},
			[]string{"module-index.one-package: example/times-tables 0.1.0: the release builds two packages of it, " + string(min(republished.Digest, timesTables(t).Digest)) +
				" and " + string(max(republished.Digest, timesTables(t).Digest)) + ": a version has one package"}},
		{"unknown withdrawal", Release{Previous: previous, Tag: "v0.3.0", Withdrawals: formatv0.Withdrawals{"example/other": {"1.0.0": "unsuitable_content"}}},
			[]string{"module-index.known-withdrawals: example/other 1.0.0: the release withdraws it for unsuitable_content, and the index names no such version"}},
		{"below", Release{Previous: previous, Tag: "v0.3.0", Packages: []*modulepackage.Package{below}},
			[]string{"module-index.version-above: example/times-tables 0.1.1: it is new in the release, and the previous index names 0.1.10 and 0.2.0 above it: a new version stands above every published one, withdrawn ones too"}},
		{"several", Release{Previous: previous, Tag: "v0.3.0", Packages: []*modulepackage.Package{below, republished},
			Withdrawals: formatv0.Withdrawals{"example/times-tables": {"0.1.1": "wrong_key", "0.9.0": "wrong_key"}, "example/smallest": {"0.1.0": "unsuitable_content"}}},
			[]string{
				"module-index.entries-kept: example/smallest 0.1.0: it is withdrawn for wrong_key, and the release withdraws it for unsuitable_content: a withdrawal keeps its reason",
				"module-index.one-package: example/times-tables 0.1.0: the previous index names its package " + string(timesTables(t).Digest) +
					", and the release builds " + string(republished.Digest) + ": a version is published once",
				"module-index.version-above: example/times-tables 0.1.1: it is new in the release, and the previous index names 0.1.10 and 0.2.0 above it: a new version stands above every published one, withdrawn ones too",
				"module-index.known-withdrawals: example/times-tables 0.9.0: the release withdraws it for wrong_key, and the index names no such version",
			}},
	}
	for _, c := range cases {
		text, err := Build(c.release)
		var refusal *Refusal
		if !errors.As(err, &refusal) || text != nil {
			t.Errorf("%s: Build = %s, %v; want a refusal", c.name, text, err)
			continue
		}
		got := make([]string, len(refusal.Breaches))
		for i, b := range refusal.Breaches {
			got[i] = b.String()
		}
		if !slices.Equal(got, c.want) {
			t.Errorf("%s: the breaches are\n%s\nwant\n%s", c.name, strings.Join(got, "\n"), strings.Join(c.want, "\n"))
		}
		if want := "moduleindex: the release breaks the rules of the module index, so it has no index:\n" + strings.Join(c.want, "\n"); refusal.Error() != want {
			t.Errorf("%s: Error() = %q", c.name, refusal.Error())
		}
	}
}

// A release whose previous index cannot be read fails rather than start a new index, and so does
// one that cannot be given an index: each with the error of its reason.
func TestBuildErrors(t *testing.T) {
	index := formatv0.ModuleIndex{ModuleFormatVersion: "0.4", Modules: map[formatv0.ModuleID]map[formatv0.ModuleVersion]formatv0.Entry{
		"example/smallest": {"0.1.0": {ModuleFormatVersion: "0.4", PackageDigest: smallest(t).Digest, Release: "v0.1.0"}},
	}}
	previous := textOf(t, index)
	withMember := strings.Replace(string(previous), `"release":"v0.1.0"`, `"release":"v0.1.0","place":"elsewhere"`, 1)
	withReason := strings.Replace(string(previous), `"release":"v0.1.0"`, `"release":"v0.1.0","withdrawn":"unsafe_widget"`, 1)
	badID := *smallest(t)
	badID.Inventory.ID = "Example/Smallest"
	badVersion := *smallest(t)
	badVersion.Inventory.Version = "0.1"
	otherFormat := *smallest(t)
	otherFormat.Inventory.ModuleFormatVersion = "0.3"
	badTeaches := *timesTables(t)
	badTeaches.View.SkillGraph = &formatv0.SkillGraph{Skills: map[formatv0.SkillID]formatv0.Skill{"a": {Teaches: []formatv0.SharedTaxonomy{"multiplication"}}}}
	later := index
	later.ModuleFormatVersion = "0.10"
	unordered := strings.Replace(string(textOf(t, formatv0.ModuleIndex{ModuleFormatVersion: "0.4", Modules: map[formatv0.ModuleID]map[formatv0.ModuleVersion]formatv0.Entry{
		"example/times-tables": {"0.1.0": {ModuleFormatVersion: "0.4", PackageDigest: timesTables(t).Digest, Release: "v0.1.0",
			Teaches: []formatv0.AnchorSkillID{"core.division-with-remainder", "core.multiplication-table"}}},
	}})), `"core.division-with-remainder","core.multiplication-table"`, `"core.multiplication-table","core.division-with-remainder"`, 1)
	packages := []*modulepackage.Package{smallest(t)}
	cases := []struct {
		name    string
		release Release
		want    error
	}{
		{"first with a previous index", Release{First: true, Previous: previous, Tag: "v0.2.0"}, ErrPrevious},
		{"no previous index", Release{Tag: "v0.2.0", Packages: packages}, ErrPrevious},
		{"an empty previous index", Release{Previous: []byte{}, Tag: "v0.2.0", Packages: packages}, ErrPrevious},
		{"not JSON", Release{Previous: []byte("module_format_version: 0.3"), Tag: "v0.2.0"}, ErrPrevious},
		{"a member of a later version", Release{Previous: []byte(withMember), Tag: "v0.2.0"}, ErrPrevious},
		{"a reason of a later version", Release{Previous: []byte(withReason), Tag: "v0.2.0"}, ErrPrevious},
		{"not the canonical text", Release{Previous: append(slices.Clone(previous), '\n'), Tag: "v0.2.0"}, ErrPrevious},
		{"not UTF-8", Release{Previous: append(slices.Clone(previous[:len(previous)-1]), 0xff, '}'), Tag: "v0.2.0"}, ErrPrevious},
		{"a later format version", Release{Previous: textOf(t, later), Tag: "v0.2.0"}, ErrPrevious},
		{"teaches out of order", Release{Previous: []byte(unordered), Tag: "v0.2.0"}, ErrPrevious},
		{"a tag that is no tag", Release{Previous: previous, Tag: "0.2.0"}, ErrRelease},
		{"empty withdrawals", Release{Previous: previous, Tag: "v0.2.0", Withdrawals: formatv0.Withdrawals{}}, ErrRelease},
		{"a module that withdraws none", Release{Previous: previous, Tag: "v0.2.0", Withdrawals: formatv0.Withdrawals{"example/smallest": {}}}, ErrRelease},
		{"an unknown reason", Release{Previous: previous, Tag: "v0.2.0", Withdrawals: formatv0.Withdrawals{"example/smallest": {"0.1.0": "unsafe_widget"}}}, ErrRelease},
		{"a nil package", Release{Previous: previous, Tag: "v0.2.0", Packages: []*modulepackage.Package{nil}}, ErrRelease},
		{"a package of no version", Release{Previous: previous, Tag: "v0.2.0", Packages: []*modulepackage.Package{&badVersion}}, ErrRelease},
		{"a package of no module ID", Release{Previous: previous, Tag: "v0.2.0", Packages: []*modulepackage.Package{&badID}}, ErrRelease},
		{"a package of another format version", Release{Previous: previous, Tag: "v0.2.0", Packages: []*modulepackage.Package{&otherFormat}}, ErrRelease},
		{"a package that teaches no anchor skill", Release{Previous: previous, Tag: "v0.2.0", Packages: []*modulepackage.Package{&badTeaches}}, ErrRelease},
		{"a bad package beside a breach", Release{Previous: previous, Tag: "v0.2.0", Packages: []*modulepackage.Package{&badID},
			Withdrawals: formatv0.Withdrawals{"example/other": {"1.0.0": "wrong_key"}}}, ErrRelease},
		{"nothing to name", Release{First: true, Tag: "v0.1.0"}, ErrRelease},
	}
	for _, c := range cases {
		text, err := Build(c.release)
		if !errors.Is(err, c.want) || text != nil {
			t.Errorf("%s: Build = %s, %v; want %v", c.name, text, err, c.want)
		}
	}
	// The previous index is read as it is: the build changes none of its entries.
	text, err := Build(Release{Previous: previous, Tag: "v0.2.0"})
	if err != nil || string(text) != string(previous) {
		t.Errorf("Build of a release with nothing new = %s, %v; want the previous index", text, err)
	}
}

// The anchor skills of an entry are those of the skills of the version's skill graph, each once,
// in the order of their IDs, and none for a module without a skill graph or whose skills teach none.
func TestTeachesOf(t *testing.T) {
	if got := teachesOf(formatv0.CanonicalView{}); got != nil {
		t.Errorf("teachesOf of a view without a skill graph = %v", got)
	}
	view := formatv0.CanonicalView{SkillGraph: &formatv0.SkillGraph{Skills: map[formatv0.SkillID]formatv0.Skill{
		"a": {Teaches: []formatv0.SharedTaxonomy{"core.multiplication-table", "core.addition-within-20"}},
		"b": {},
		"c": {Teaches: []formatv0.SharedTaxonomy{"core.multiplication-table"}},
	}}}
	want := []formatv0.AnchorSkillID{"core.addition-within-20", "core.multiplication-table"}
	if got := teachesOf(view); !reflect.DeepEqual(got, want) {
		t.Errorf("teachesOf = %v; want %v", got, want)
	}
	view.SkillGraph.Skills = map[formatv0.SkillID]formatv0.Skill{"b": {}}
	if got := teachesOf(view); got != nil {
		t.Errorf("teachesOf of skills that teach none = %v", got)
	}
}
