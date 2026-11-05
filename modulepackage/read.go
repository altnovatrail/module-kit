package modulepackage

import (
	"errors"
	"fmt"
	"io/fs"
	"maps"
	"slices"

	"github.com/altnovatrail/spec/types/formatv0"
)

// The reasons Read refuses a package, which its error wraps with the file it refuses.
var (
	// ErrChain is a file that is not the one the chain of digests names: its digest or its size is
	// not the one the digest of the package, the inventory or the listing gives it, or it is missing.
	ErrChain = errors.New("not the file the chain of digests names")
	// ErrFormatVersion is a package of a format version that module-kit does not read
	// (format/README.md, section 2).
	ErrFormatVersion = errors.New("a format version module-kit does not read")
	// ErrNotCanonical is a file of a package that is not the canonical text of a value valid against
	// its schema, or an inventory that names another module than its canonical view does.
	ErrNotCanonical = errors.New("not what the format builds")
)

// Read reads a package as files, the folder at the root of fsys, by the chain of digests from the
// digest of the package that a checked index gives (format/README.md, sections 25 and 26; R62 spec):
// inventory.json by that digest, then server.json and device.json by the digests and sizes the
// inventory gives, then each file of device/ by those the listing gives. A file is read only once
// the one before it in the chain has named it, and each is held to its schema and to its canonical
// text, so a package that Read accepts is the one Build built. Read checks no signature: the index
// was checked where it entered NovaTrail (R77 spec). A file of the folder that the chain does not
// name is no part of the package, and Read does not read it.
//
// It reads a package of the format version that module-kit reads, FormatVersion of package
// module; of another, it gives ErrFormatVersion as soon as the inventory names it. Its errors wrap
// ErrChain, ErrFormatVersion or ErrNotCanonical, with the file they are of, or are those of fsys.
func Read(fsys fs.FS, digest formatv0.Digest) (*Package, error) {
	return read(fsys, digest, true)
}

// ReadParts reads a package as files as Read does, by the chain of digests from the digest of the
// package, but without the files of device/: its inventory, its canonical view and the listing of its
// device part, which names each of those files by its digest and size. So the comparison of the
// versions of a module reads a published package, whose device part it compares by its listing
// (format/README.md, section 27). Files then holds InventoryFile, ServerFile and DeviceFile alone. Its
// errors are those of Read.
func ReadParts(fsys fs.FS, digest formatv0.Digest) (*Package, error) {
	return read(fsys, digest, false)
}

// read reads a package as files, and the files of device/ when withDevice is true.
func read(fsys fs.FS, digest formatv0.Digest, withDevice bool) (*Package, error) {
	files := map[string][]byte{}
	inventoryValue, err := readPart(fsys, InventoryFile, formatv0.Descriptor{Digest: digest, Size: anySize}, "inventory.schema.json", files)
	if err != nil {
		return nil, err
	}
	inventory := decode[formatv0.Inventory](inventoryValue)
	viewValue, err := readPart(fsys, ServerFile, inventory.ServerPart, "canonical-view.schema.json", files)
	if err != nil {
		return nil, err
	}
	listingValue, err := readPart(fsys, DeviceFile, inventory.DevicePart, "inventory.schema.json#/$defs/device_listing", files)
	if err != nil {
		return nil, err
	}
	view := decode[formatv0.CanonicalView](viewValue)
	if m := view.Manifest; m.ModuleFormatVersion != inventory.ModuleFormatVersion || m.ID != inventory.ID || m.Version != inventory.Version {
		return nil, fmt.Errorf("modulepackage: %s names %s %s of the format %s, and its canonical view %s %s of the format %s: %w", InventoryFile,
			inventory.ID, inventory.Version, inventory.ModuleFormatVersion, m.ID, m.Version, m.ModuleFormatVersion, ErrNotCanonical)
	}
	listing := decode[formatv0.DeviceListing](listingValue)
	if !withDevice {
		return &Package{Digest: digest, Inventory: inventory, View: view, Listing: listing, Files: files}, nil
	}
	named := map[string]formatv0.Descriptor{licenseFile: listing.License}
	if listing.Notice != nil {
		named[noticeFile] = *listing.Notice
	}
	for media, d := range listing.Media {
		named[mediaFolder+string(media)] = d
	}
	for _, file := range slices.Sorted(maps.Keys(named)) {
		if _, err := readFile(fsys, DeviceFolder+file, named[file], files); err != nil {
			return nil, err
		}
	}
	return &Package{Digest: digest, Inventory: inventory, View: view, Listing: listing, Files: files}, nil
}
