package moduleindex

import (
	"errors"
	"io/fs"
	"maps"
	"reflect"
	"testing"

	"github.com/altnovatrail/spec"
	"github.com/altnovatrail/spec/types/formatv0"

	"github.com/altnovatrail/module-kit/internal/canonical"
	"github.com/altnovatrail/module-kit/internal/schema"
	"github.com/altnovatrail/module-kit/internal/semver"
	"github.com/altnovatrail/module-kit/internal/yaml"
	"github.com/altnovatrail/module-kit/module"
	"github.com/altnovatrail/module-kit/modulepackage"
)

// corpus adds the indexes of the vectors of the index, previous and built, and the examples of the
// module index, each as its canonical text, so that a change to any byte of them is tried.
func corpus(f *testing.F) {
	vectors, err := fs.Glob(spec.Format, "vectors/indexes/*/index.yaml")
	if err != nil || len(vectors) == 0 {
		f.Fatalf("no vectors of the index: %v", err)
	}
	for _, name := range vectors {
		data, err := fs.ReadFile(spec.Format, name)
		if err != nil {
			f.Fatal(err)
		}
		value, err := yaml.Read(data)
		if err != nil {
			f.Fatal(err)
		}
		for _, member := range []string{"previous", "index"} {
			if index, ok := value.(map[string]any)[member]; ok {
				text, err := canonical.Text(index)
				if err != nil {
					f.Fatal(err)
				}
				f.Add(text)
			}
		}
	}
	for _, kind := range []string{"valid", "invalid"} {
		for _, text := range examples(f, kind) {
			f.Add(text)
		}
	}
}

// On any text, Read gives an index or an error of ErrNotIndex, without a panic. An index that it
// gives has each member at a place it keeps in its form, so its canonical text reads back as the
// same index; and Newest picks of each module a version that the index names, not withdrawn, of
// the format version module-kit reads, above every other such version.
func FuzzRead(f *testing.F) {
	corpus(f)
	f.Fuzz(func(t *testing.T, data []byte) {
		index, err := Read(data)
		if err != nil {
			if !errors.Is(err, ErrNotIndex) || index.Modules != nil {
				t.Fatalf("Read gives %+v and an error of no reason of Read: %v", index, err)
			}
			return
		}
		text, err := canonical.Text(schema.ValueOf(index))
		if err != nil {
			t.Fatalf("the index that Read gives has no canonical text: %v", err)
		}
		if again := mustRead(t, text); !reflect.DeepEqual(again, index) {
			t.Fatalf("the canonical text of the index reads back as another index:\n%+v\n%+v", again, index)
		}
		for id, versions := range index.Modules {
			newest, ok := Newest(index, id)
			for v, entry := range versions {
				if entry.Withdrawn != nil || entry.ModuleFormatVersion != module.FormatVersion {
					if ok && v == newest {
						t.Fatalf("Newest(%s) picks %s, which is withdrawn or of the format %s", id, v, entry.ModuleFormatVersion)
					}
					continue
				}
				if !ok {
					t.Fatalf("Newest(%s) picks none, and %s is a version it may take", id, v)
				}
				parsed, _ := semver.Parse(string(v))
				picked, _ := semver.Parse(string(newest))
				if parsed.Compare(picked) > 0 {
					t.Fatalf("Newest(%s) picks %s, below %s", id, newest, v)
				}
			}
		}
	})
}

// On any previous index, Build of a release gives an index, a refusal or an error of ErrPrevious,
// without a panic; and an index that it builds holds every entry of the previous index as it was,
// and the entries of the packages of the release. The release is that of the vector next-release,
// and it withdraws the smallest module 0.1.0, so that a previous index that withdraws it for
// another reason is refused.
func FuzzBuildPrevious(f *testing.F) {
	corpus(f)
	packages := []*modulepackage.Package{smallest(f), timesTables020(f)}
	withdrawals := formatv0.Withdrawals{"example/smallest": {"0.1.0": "unsuitable_content"}}
	f.Fuzz(func(t *testing.T, previous []byte) {
		text, err := Build(Release{Previous: previous, Tag: "v0.2.0", Packages: packages, Withdrawals: withdrawals})
		var refusal *Refusal
		switch {
		case errors.As(err, &refusal):
			if len(refusal.Breaches) == 0 {
				t.Fatal("a refusal with no breach")
			}
			return
		case err != nil:
			if !errors.Is(err, ErrPrevious) {
				t.Fatalf("Build gives an error of no reason of a previous index: %v", err)
			}
			return
		}
		built, before := mustRead(t, text), mustRead(t, previous)
		if built.ModuleFormatVersion != module.FormatVersion {
			t.Fatalf("the index names the format version %s", built.ModuleFormatVersion)
		}
		for id, versions := range before.Modules {
			for v, entry := range versions {
				got := built.Modules[id][v]
				if id == "example/smallest" && v == "0.1.0" && entry.Withdrawn == nil {
					entry.Withdrawn = reason("unsuitable_content")
				}
				if !reflect.DeepEqual(got, entry) {
					t.Fatalf("the entry of %s %s is %+v, and the previous index has %+v", id, v, got, entry)
				}
			}
		}
		for _, p := range packages {
			if got := built.Modules[p.Inventory.ID][p.Inventory.Version]; got.PackageDigest != p.Digest {
				t.Fatalf("the entry of %s %s names %s, and the package is %s", p.Inventory.ID, p.Inventory.Version, got.PackageDigest, p.Digest)
			}
		}
		for id := range maps.Keys(built.Modules) {
			if _, ok := before.Modules[id]; !ok && id != "example/smallest" && id != "example/times-tables" {
				t.Fatalf("the index names %s, which neither the previous index nor the release does", id)
			}
		}
	})
}
