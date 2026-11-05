package modulepackage

import (
	"context"
	"errors"
	"io/fs"
	"maps"
	"reflect"
	"slices"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/altnovatrail/spec"
)

// vectorModule gives the module of a vector of packages, as a file system of its own that a test
// may add to, and the folder it is read from.
func vectorModule(t testing.TB, name string) fstest.MapFS {
	t.Helper()
	root := "vectors/packages/" + name + "/module"
	module := fstest.MapFS{}
	err := fs.WalkDir(spec.Format, root, func(file string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() {
			return err
		}
		data, err := fs.ReadFile(spec.Format, file)
		module[strings.TrimPrefix(file, root+"/")] = &fstest.MapFile{Data: data}
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	return module
}

// The package holds the device part's files byte for byte, under device/, beside the three texts,
// and nothing else: of the module with media, a notice, an eval set and a file of editor's notes,
// the eval set and the notes stay out.
func TestBuildFiles(t *testing.T) {
	pkg, err := Build(context.Background(), vectorModule(t, "with-media-and-notice"))
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"device.json", "device/LICENSE", "device/NOTICE", "device/media/pictures/dot.png", "inventory.json", "server.json"}
	if got := slices.Sorted(maps.Keys(pkg.Files)); !slices.Equal(got, want) {
		t.Errorf("the package holds %v; want %v", got, want)
	}
	license, _ := fs.ReadFile(spec.Format, "vectors/packages/with-media-and-notice/module/LICENSE")
	if string(pkg.Files["device/LICENSE"]) != string(license) {
		t.Error("device/LICENSE is not LICENSE byte for byte")
	}
	if pkg.Listing.Notice == nil || len(pkg.Listing.Media) != 1 {
		t.Errorf("the listing is %+v", pkg.Listing)
	}
}

// The package is the same whatever the folder holds beside the module: a file or a folder whose name
// begins with ".", anything in such a folder, and the eval set (format/README.md, section 25).
func TestBuildLeavesOut(t *testing.T) {
	module := vectorModule(t, "smallest")
	before, err := Build(context.Background(), module)
	if err != nil {
		t.Fatal(err)
	}
	module[".notes"] = &fstest.MapFile{Data: []byte("notes")}
	module[".git/HEAD"] = &fstest.MapFile{Data: []byte("ref")}
	module["eval/sample.json"] = &fstest.MapFile{Data: []byte("{}")}
	after, err := Build(context.Background(), module)
	if err != nil {
		t.Fatal(err)
	}
	if after.Digest != before.Digest || len(after.Files) != len(before.Files) {
		t.Errorf("the digest is %s with files beside the module, and %s without", after.Digest, before.Digest)
	}
}

// A module that breaks the format has no package: Build gives a *Refusal with its report.
func TestBuildRefuses(t *testing.T) {
	module := vectorModule(t, "smallest")
	delete(module, "LICENSE")
	pkg, err := Build(context.Background(), module)
	var refusal *Refusal
	if !errors.As(err, &refusal) || pkg != nil {
		t.Fatalf("Build = %v, %v; want a *Refusal", pkg, err)
	}
	if len(refusal.Report.Breaches) != 1 || refusal.Report.Breaches[0].Rule != "module.license" {
		t.Errorf("the report is\n%s", refusal.Report)
	}
	if !strings.Contains(refusal.Error(), "module.license") {
		t.Errorf("Error() = %q", refusal.Error())
	}
}

// A folder that cannot be read gives its error, and no report.
func TestBuildUnreadable(t *testing.T) {
	if _, err := Build(context.Background(), brokenFS{}); err == nil || errors.As(err, new(*Refusal)) {
		t.Errorf("Build = %v; want the error of the folder", err)
	}
}

type brokenFS struct{}

func (brokenFS) Open(string) (fs.File, error) { return nil, fs.ErrPermission }

// A context that is done stops the solvers, and Build gives its error, not a refusal.
func TestBuildContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	pkg, err := Build(ctx, vectorModule(t, "with-solver"))
	if !errors.Is(err, context.Canceled) || pkg != nil {
		t.Errorf("Build = %v, %v; want context.Canceled", pkg, err)
	}
}

// An integer that a file writes as a double, 5e1 or 50.0, is an integer in the canonical view: of
// the module of strings, the picture's size is 50.
func TestIntegers(t *testing.T) {
	pkg, err := Build(context.Background(), vectorModule(t, "strings"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(pkg.Files[ServerFile]), `"height":50,`) {
		t.Errorf("server.json does not hold the height 50:\n%s", pkg.Files[ServerFile])
	}
	for _, c := range []struct {
		in, want any
	}{
		{50.0, int64(50)},
		{-3.0, int64(-3)},
		{1.5, 1.5},
		{9007199254740992.0, 9007199254740992.0},
		{map[string]any{"a": []any{2.0}}, map[string]any{"a": []any{int64(2)}}},
		{"5e1", "5e1"},
	} {
		// DeepEqual tells an int64 from a float64 of the same number.
		if got := integers(c.in); !reflect.DeepEqual(got, c.want) {
			t.Errorf("integers(%v) = %#v; want %#v", c.in, got, c.want)
		}
	}
}
