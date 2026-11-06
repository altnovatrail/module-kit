package moduleindex

import (
	"errors"
	"io/fs"
	"reflect"
	"testing"

	"github.com/altnovatrail/spec"
	"github.com/altnovatrail/spec/types/formatv0"

	"github.com/altnovatrail/module-kit/internal/canonical"
	"github.com/altnovatrail/module-kit/internal/schema"
	"github.com/altnovatrail/module-kit/internal/yaml"
)

// pair gives the previous index and the index built of a vector of the index that has both.
func pair(t testing.TB, name string) (previous, built formatv0.ModuleIndex) {
	t.Helper()
	data, err := fs.ReadFile(spec.Format, "vectors/indexes/"+name+"/index.yaml")
	if err != nil {
		t.Fatal(err)
	}
	value, err := yaml.Read(data)
	if err != nil {
		t.Fatal(err)
	}
	var indexes [2]formatv0.ModuleIndex
	for i, member := range []string{"previous", "index"} {
		text, err := canonical.Text(value.(map[string]any)[member])
		if err != nil {
			t.Fatal(err)
		}
		indexes[i] = mustRead(t, text)
	}
	return indexes[0], indexes[1]
}

// The index of a release takes the withdrawals of a later one and keeps every other member of its
// own entries, and leaves out the entries that only the later one has; Newest then skips what the
// later one withdraws.
func TestWithLatest(t *testing.T) {
	for _, c := range []struct {
		vector string
		newest map[formatv0.ModuleID]formatv0.ModuleVersion // "" when no version is left
	}{
		{"withdrawal", map[formatv0.ModuleID]formatv0.ModuleVersion{"example/smallest": "", "example/times-tables": "0.2.0"}},
		{"next-release", map[formatv0.ModuleID]formatv0.ModuleVersion{"example/smallest": "0.1.0", "example/times-tables": "0.1.0"}},
	} {
		t.Run(c.vector, func(t *testing.T) {
			previous, built := pair(t, c.vector)
			merged, err := WithLatest(previous, built)
			if err != nil {
				t.Fatal(err)
			}
			if merged.ModuleFormatVersion != previous.ModuleFormatVersion {
				t.Errorf("the format version is %s, not the release's %s", merged.ModuleFormatVersion, previous.ModuleFormatVersion)
			}
			for id, versions := range previous.Modules {
				if len(merged.Modules[id]) != len(versions) {
					t.Errorf("%s has %d versions, not the release's %d", id, len(merged.Modules[id]), len(versions))
				}
				for version, entry := range versions {
					got := merged.Modules[id][version]
					if !reflect.DeepEqual(got.Withdrawn, built.Modules[id][version].Withdrawn) {
						t.Errorf("%s %s is withdrawn by %v, not by the newest index's %v", id, version, got.Withdrawn, built.Modules[id][version].Withdrawn)
					}
					got.Withdrawn, entry.Withdrawn = nil, nil
					if !reflect.DeepEqual(got, entry) {
						t.Errorf("%s %s is %+v, not the release's %+v", id, version, got, entry)
					}
				}
			}
			for id, want := range c.newest {
				if got, ok := Newest(merged, id); got != want || ok != (want != "") {
					t.Errorf("Newest of %s is %q, %v; want %q", id, got, ok, want)
				}
			}
			if _, err := WithLatest(previous, previous); err != nil {
				t.Errorf("the release's own index as the newest: %v", err)
			}
		})
	}
}

// An index from before the release's, or one that holds an entry otherwise or drops a withdrawal,
// is not the index of a newer release.
func TestWithLatestRefuses(t *testing.T) {
	withdrawalPrevious, withdrawalBuilt := pair(t, "withdrawal")
	nextPrevious, nextBuilt := pair(t, "next-release")
	otherDigest := clone(nextBuilt)
	entry := otherDigest.Modules["example/smallest"]["0.1.0"]
	entry.PackageDigest = "sha256:0000000000000000000000000000000000000000000000000000000000000000"
	otherDigest.Modules["example/smallest"]["0.1.0"] = entry
	otherReason := clone(withdrawalBuilt)
	entry = otherReason.Modules["example/times-tables"]["0.1.0"]
	reason := formatv0.WithdrawalReason("answer_leak")
	entry.Withdrawn = &reason
	otherReason.Modules["example/times-tables"]["0.1.0"] = entry
	for _, c := range []struct {
		name            string
		release, latest formatv0.ModuleIndex
	}{
		{"an earlier index, without a version", nextBuilt, nextPrevious},
		{"an earlier index, without a withdrawal", withdrawalBuilt, withdrawalPrevious},
		{"another digest", nextPrevious, otherDigest},
		{"another reason", withdrawalPrevious, otherReason},
		{"no index", nextPrevious, formatv0.ModuleIndex{}},
	} {
		if _, err := WithLatest(c.release, c.latest); !errors.Is(err, ErrLatest) {
			t.Errorf("%s: %v; want ErrLatest", c.name, err)
		}
	}
	if nextBuilt.Modules["example/smallest"]["0.1.0"].PackageDigest == otherDigest.Modules["example/smallest"]["0.1.0"].PackageDigest {
		t.Error("clone shares the entries of its index")
	}
}

// clone copies an index, so that a change to the copy leaves the index as it is.
func clone(index formatv0.ModuleIndex) formatv0.ModuleIndex {
	copied := formatv0.ModuleIndex{ModuleFormatVersion: index.ModuleFormatVersion, Modules: map[formatv0.ModuleID]map[formatv0.ModuleVersion]formatv0.Entry{}}
	for id, versions := range index.Modules {
		copied.Modules[id] = map[formatv0.ModuleVersion]formatv0.Entry{}
		for version, entry := range versions {
			copied.Modules[id][version] = entry
		}
	}
	return copied
}

// On any two indexes that Read gives, WithLatest gives an index with the release's entries, each
// withdrawn as latest withdraws it, or ErrLatest, without a panic.
func FuzzWithLatest(f *testing.F) {
	previous, built := pair(f, "withdrawal")
	var texts [][]byte
	for _, index := range []formatv0.ModuleIndex{previous, built} {
		text, err := canonical.Text(schema.ValueOf(index))
		if err != nil {
			f.Fatal(err)
		}
		texts = append(texts, text)
	}
	for _, a := range texts {
		for _, b := range texts {
			f.Add(a, b)
		}
	}
	f.Fuzz(func(t *testing.T, a, b []byte) {
		release, err := Read(a)
		if err != nil {
			return
		}
		latest, err := Read(b)
		if err != nil {
			return
		}
		merged, err := WithLatest(release, latest)
		if err != nil {
			if !errors.Is(err, ErrLatest) {
				t.Fatalf("WithLatest: %v", err)
			}
			return
		}
		for id, versions := range release.Modules {
			for version := range versions {
				got, ok := merged.Modules[id][version]
				if !ok || !reflect.DeepEqual(got.Withdrawn, latest.Modules[id][version].Withdrawn) {
					t.Fatalf("%s %s is %+v, %v in the merged index", id, version, got, ok)
				}
			}
		}
	})
}
