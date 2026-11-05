package modulepackage

import (
	"errors"
	"io/fs"
	"reflect"
	"testing"
	"testing/fstest"

	"github.com/altnovatrail/module-kit/internal/canonical"
)

// On any inventory, server part and listing, Read gives a package or an error of its own, without a
// panic: a package only when the chain holds and each file is the canonical text of a value valid
// against its schema, and then the package is the one it read, its files byte for byte; and any
// other error wraps ErrChain, ErrFormatVersion or ErrNotCanonical. The fuzz test reads the three
// files by the digest of the inventory it is given, and the files of the device part are those of
// the package of with-media-and-notice, so that an input that keeps their listing reads to the end.
// The corpus is the three files of the package of every vector of packages, so that a change to any
// byte of them is tried.
func FuzzRead(f *testing.F) {
	device := built(f, "with-media-and-notice").Files
	for _, name := range []string{"smallest", "strings", "with-lessons", "with-media-and-notice", "with-solver", "with-templates"} {
		pkg := built(f, name)
		f.Add(pkg.Files[InventoryFile], pkg.Files[ServerFile], pkg.Files[DeviceFile])
	}
	f.Fuzz(func(t *testing.T, inventory, server, listing []byte) {
		folder := fstest.MapFS{
			InventoryFile: {Data: inventory},
			ServerFile:    {Data: server},
			DeviceFile:    {Data: listing},
		}
		for name, data := range device {
			if _, isPart := folder[name]; !isPart {
				folder[name] = &fstest.MapFile{Data: data}
			}
		}
		digest := canonical.Digest(inventory)
		pkg, err := Read(folder, digest)
		if err != nil {
			if pkg != nil {
				t.Fatalf("Read gives a package and the error %v", err)
			}
			if !errors.Is(err, ErrChain) && !errors.Is(err, ErrFormatVersion) && !errors.Is(err, ErrNotCanonical) {
				t.Fatalf("an error of no reason of Read: %v", err)
			}
			return
		}
		if pkg.Digest != digest || !bytesEqual(pkg.Files[InventoryFile], inventory) || !bytesEqual(pkg.Files[ServerFile], server) ||
			!bytesEqual(pkg.Files[DeviceFile], listing) {
			t.Fatal("Read gives a package of other files than it read")
		}
		for name, data := range pkg.Files {
			if !reflect.DeepEqual(data, readOrNil(folder, name)) {
				t.Fatalf("the package holds %s, which is not the file the folder holds", name)
			}
		}
	})
}

func bytesEqual(a, b []byte) bool { return string(a) == string(b) }

func readOrNil(fsys fs.FS, name string) []byte {
	data, _ := fs.ReadFile(fsys, name)
	return data
}
