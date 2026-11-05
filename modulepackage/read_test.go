package modulepackage

import (
	"context"
	"errors"
	"io/fs"
	"maps"
	"reflect"
	"testing"
	"testing/fstest"
)

// asFolder gives the files of a package as a folder that Read reads.
func asFolder(files map[string][]byte) fstest.MapFS {
	folder := fstest.MapFS{}
	for name, data := range files {
		folder[name] = &fstest.MapFile{Data: data}
	}
	return folder
}

// built gives the package of the module of a vector of packages.
func built(t testing.TB, name string) *Package {
	t.Helper()
	pkg, err := Build(context.Background(), vectorModule(t, name))
	if err != nil {
		t.Fatal(err)
	}
	return pkg
}

// A package as files reads back as the package Build built, and a file of the folder that the chain
// does not name changes nothing.
func TestRead(t *testing.T) {
	for _, name := range []string{"smallest", "with-media-and-notice", "with-solver"} {
		pkg := built(t, name)
		folder := asFolder(pkg.Files)
		folder["device/media/unlisted.png"] = &fstest.MapFile{Data: []byte("not named")}
		folder["README"] = &fstest.MapFile{Data: []byte("not named")}
		got, err := Read(folder, pkg.Digest)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if !reflect.DeepEqual(got, pkg) {
			t.Errorf("%s: the package reads back as another package", name)
		}
	}
}

// An error of the folder other than a missing file comes back as it is.
func TestReadUnreadable(t *testing.T) {
	_, err := Read(brokenFS{}, built(t, "smallest").Digest)
	if !errors.Is(err, fs.ErrPermission) {
		t.Errorf("Read = %v; want fs.ErrPermission", err)
	}
}

// parse reads an integer as an int64, and any other number as a float64, as the reader of YAML
// does: so the canonical text refuses 1.0 and 1e2 in a package, which it never writes.
func TestParse(t *testing.T) {
	got, err := parse([]byte(`{"a":1,"b":1.0,"c":1e2,"d":-9007199254740991,"e":[true,null,"x"]}`))
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]any{"a": int64(1), "b": 1.0, "c": 100.0, "d": int64(-9007199254740991), "e": []any{true, nil, "x"}}
	if !maps.EqualFunc(got.(map[string]any), want, reflect.DeepEqual) {
		t.Errorf("parse = %#v; want %#v", got, want)
	}
}
