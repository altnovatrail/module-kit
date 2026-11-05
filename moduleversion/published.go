package moduleversion

import (
	"errors"
	"fmt"
	"io/fs"
	"regexp"

	"github.com/altnovatrail/spec/types/formatv0"

	"github.com/altnovatrail/module-kit/internal/canonical"
	"github.com/altnovatrail/module-kit/module"
	"github.com/altnovatrail/module-kit/modulepackage"
)

// A published version of a module as the comparison reads it: its digest, its places, and, when its
// format version is the one module-kit reads, its package, which Changes compares.
type publishedVersion struct {
	digest formatv0.Digest
	places places
	pkg    *modulepackage.Package
}

var (
	formatVersionForm = regexp.MustCompile(`^(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$`)
	digestForm        = regexp.MustCompile(`^sha256:[0-9a-f]{64}$`)
)

// readPublished reads a published package as files by the chain of digests from its digest
// (format/README.md, sections 2 and 25): inventory.json by the digest, and then, by the format
// version that the inventory names at the place every version keeps, a package of the version
// module-kit reads by modulepackage.ReadParts, and one of an earlier version by its places alone,
// server.json by the digest and the size that the inventory gives. A package of a newer version of
// the format is modulepackage.ErrFormatVersion.
func readPublished(p Published) (publishedVersion, error) {
	inventoryData, err := readNamed(p.Files, modulepackage.InventoryFile)
	if err != nil {
		return publishedVersion{}, err
	}
	if got := canonical.Digest(inventoryData); got != p.Digest {
		return publishedVersion{}, fmt.Errorf("%s has the digest %s, and the chain names %s: %w", modulepackage.InventoryFile, got, p.Digest, modulepackage.ErrChain)
	}
	inventory, err := canonical.Parse(inventoryData)
	if err != nil {
		return publishedVersion{}, fmt.Errorf("%s is no JSON text: %v: %w", modulepackage.InventoryFile, err, modulepackage.ErrNotCanonical)
	}
	object, _ := inventory.(map[string]any)
	v, _ := object["module_format_version"].(string)
	switch {
	case module.ReadsFormat(v):
		pkg, err := modulepackage.ReadParts(p.Files, p.Digest)
		if err != nil {
			return publishedVersion{}, err
		}
		pl, err := placesOfPackage(pkg)
		return publishedVersion{p.Digest, pl, pkg}, err
	case !formatVersionForm.MatchString(v) || module.NewerFormat(v):
		return publishedVersion{}, fmt.Errorf("%s names the format version %v, and module-kit reads %s and the places of the versions before it: %w",
			modulepackage.InventoryFile, object["module_format_version"], module.FormatVersion, modulepackage.ErrFormatVersion)
	}
	serverData, err := readNamed(p.Files, modulepackage.ServerFile)
	if err != nil {
		return publishedVersion{}, err
	}
	part, _ := object["server_part"].(map[string]any)
	digest, _ := part["digest"].(string)
	size, sized := part["size"].(int64)
	if got := canonical.Descriptor(serverData); !digestForm.MatchString(digest) || !sized || got.Digest != formatv0.Digest(digest) || got.Size != size {
		return publishedVersion{}, fmt.Errorf("%s has the digest %s and %d bytes, and %s names %v: %w", modulepackage.ServerFile,
			got.Digest, got.Size, modulepackage.InventoryFile, object["server_part"], modulepackage.ErrChain)
	}
	view, err := canonical.Parse(serverData)
	if err != nil {
		return publishedVersion{}, fmt.Errorf("%s is no JSON text: %v: %w", modulepackage.ServerFile, err, modulepackage.ErrNotCanonical)
	}
	pl, err := placesOf(inventory, view)
	return publishedVersion{p.Digest, pl, nil}, err
}

// readNamed reads a file of a package that the chain names: a missing one breaks the chain.
func readNamed(fsys fs.FS, name string) ([]byte, error) {
	data, err := fs.ReadFile(fsys, name)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, fmt.Errorf("%s is missing: %w", name, modulepackage.ErrChain)
	}
	return data, err
}

// placesOfPackage gives the places of a package that modulepackage built or read.
func placesOfPackage(pkg *modulepackage.Package) (places, error) {
	inventory, err := canonical.Parse(pkg.Files[modulepackage.InventoryFile])
	if err != nil {
		return places{}, fmt.Errorf("%s is no JSON text: %v: %w", modulepackage.InventoryFile, err, modulepackage.ErrNotCanonical)
	}
	view, err := canonical.Parse(pkg.Files[modulepackage.ServerFile])
	if err != nil {
		return places{}, fmt.Errorf("%s is no JSON text: %v: %w", modulepackage.ServerFile, err, modulepackage.ErrNotCanonical)
	}
	return placesOf(inventory, view)
}
