package modulepackage

import (
	"context"
	"errors"
	"io/fs"
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
