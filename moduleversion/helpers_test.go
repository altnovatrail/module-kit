package moduleversion

import (
	"context"
	"io/fs"
	"path"
	"testing"
	"testing/fstest"

	spec "github.com/altnovatrail/spec"
	"github.com/altnovatrail/spec/types/formatv0"

	"github.com/altnovatrail/module-kit/internal/canonical"

	"github.com/altnovatrail/module-kit/modulepackage"
)

// built builds the package of a module laid out at a path of spec.Format.
func built(t testing.TB, at string) *modulepackage.Package {
	t.Helper()
	folder, err := fs.Sub(spec.Format, at)
	if err != nil {
		t.Fatal(err)
	}
	pkg, err := modulepackage.Build(context.Background(), folder)
	if err != nil {
		t.Fatalf("%s: %v", at, err)
	}
	return pkg
}

// published gives a package as the CI of modules gives a published one: as files, without device/.
func published(pkg *modulepackage.Package) Published {
	files := fstest.MapFS{}
	for _, name := range []string{modulepackage.InventoryFile, modulepackage.ServerFile, modulepackage.DeviceFile} {
		files[name] = &fstest.MapFile{Data: pkg.Files[name]}
	}
	return Published{Digest: pkg.Digest, Files: files}
}

// versionCase gives the new version and the published one of a case of vectors/versions/ with one
// published version laid out as a module.
func versionCase(t testing.TB, name, version string) (next, before *modulepackage.Package) {
	t.Helper()
	at := path.Join("vectors/versions", name)
	return built(t, path.Join(at, "module")), built(t, path.Join(at, "published", version))
}

// digestOf gives the digest of the bytes of a file, as the digest of a package is that of its
// inventory.
func digestOf(data []byte) formatv0.Digest {
	return canonical.Digest(data)
}
