package modulepackage

import (
	"errors"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/altnovatrail/spec/types/formatv0"

	"github.com/altnovatrail/module-kit/internal/canonical"
)

// A folder of a package that a case of TestReadRefuses reads, and the digest it reads it by.
type folderCase func(pkg *Package) (fstest.MapFS, formatv0.Digest)

// withFile gives the package's folder with one file in place of its own, or without it when data is
// nil, read by the package's digest.
func withFile(name string, data []byte) folderCase {
	return func(pkg *Package) (fstest.MapFS, formatv0.Digest) {
		folder := asFolder(pkg.Files)
		if data == nil {
			delete(folder, name)
		} else {
			folder[name] = &fstest.MapFile{Data: data}
		}
		return folder, pkg.Digest
	}
}

// withInventory gives the package's folder with another inventory, read by the digest of it, so that
// the chain holds up to the inventory.
func withInventory(text string) folderCase {
	return func(pkg *Package) (fstest.MapFS, formatv0.Digest) {
		folder := asFolder(pkg.Files)
		folder[InventoryFile] = &fstest.MapFile{Data: []byte(text)}
		return folder, canonical.Digest([]byte(text))
	}
}

// withParts gives the package's folder with a server part and a device part that edit makes of its
// own, and an inventory in canonical text that names them, read by its digest, so that the chain
// holds up to the parts.
func withParts(edit func(server, device string) (string, string)) folderCase {
	return func(pkg *Package) (fstest.MapFS, formatv0.Digest) {
		server, device := edit(string(pkg.Files[ServerFile]), string(pkg.Files[DeviceFile]))
		inventory, err := canonical.Text(map[string]any{
			versionField: string(pkg.Inventory.ModuleFormatVersion), "id": string(pkg.Inventory.ID),
			"version": string(pkg.Inventory.Version), "server_part": descriptorValue(canonical.Descriptor([]byte(server))),
			"device_part": descriptorValue(canonical.Descriptor([]byte(device))),
		})
		if err != nil {
			panic(err)
		}
		folder := asFolder(pkg.Files)
		folder[InventoryFile] = &fstest.MapFile{Data: inventory}
		folder[ServerFile] = &fstest.MapFile{Data: []byte(server)}
		folder[DeviceFile] = &fstest.MapFile{Data: []byte(device)}
		return folder, canonical.Digest(inventory)
	}
}

// Every way a package breaks the chain of digests, or is not what the format builds, is refused by
// its reason, with the file it is of.
func TestReadRefuses(t *testing.T) {
	pkg := built(t, "with-media-and-notice")
	other := func(pkg *Package) (fstest.MapFS, formatv0.Digest) {
		return asFolder(pkg.Files), canonical.Digest([]byte("another"))
	}
	for _, c := range []struct {
		name   string
		folder folderCase
		reason error
		file   string
	}{
		{"another digest", other, ErrChain, InventoryFile},
		{"no inventory", withFile(InventoryFile, nil), ErrChain, InventoryFile},
		{"a changed server part", withFile(ServerFile, []byte("{}")), ErrChain, ServerFile},
		{"no listing", withFile(DeviceFile, nil), ErrChain, DeviceFile},
		{"a changed picture", withFile("device/media/pictures/dot.png", []byte("\x89PNG")), ErrChain, "device/media/pictures/dot.png"},
		{"no notice", withFile("device/NOTICE", nil), ErrChain, "device/NOTICE"},
		{"another format version", withInventory(`{"module_format_version":"0.3"}`), ErrFormatVersion, InventoryFile},
		{"no format version", withInventory(`[]`), ErrFormatVersion, InventoryFile},
		{"no JSON", withInventory(`{`), ErrNotCanonical, InventoryFile},
		{"two values", withInventory(`{} {}`), ErrNotCanonical, InventoryFile},
		{"a number past a double", withInventory(`{"module_format_version":"0.4","id":1e700}`), ErrNotCanonical, InventoryFile},
		{"an inventory not of its schema", withInventory(`{"module_format_version":"0.4"}`), ErrNotCanonical, InventoryFile},
		{"an inventory not canonical", withInventory(strings.Replace(string(pkg.Files[InventoryFile]), `,"id":`, `, "id":`, 1)),
			ErrNotCanonical, InventoryFile},
		{"a server part not of its schema", withParts(func(_, device string) (string, string) { return `{}`, device }), ErrNotCanonical, ServerFile},
		{"a listing not canonical", withParts(func(server, device string) (string, string) {
			return server, strings.Replace(device, "{", "{ ", 1)
		}), ErrNotCanonical, DeviceFile},
		{"a size written 109.0", withParts(func(server, device string) (string, string) {
			return server, strings.Replace(device, `"size":109`, `"size":109.0`, 1)
		}), ErrNotCanonical, DeviceFile},
		{"an inventory of another module", withParts(func(server, device string) (string, string) {
			return strings.Replace(server, "example/with-media-and-notice", "example/another", 1), device
		}), ErrNotCanonical, InventoryFile},
	} {
		folder, digest := c.folder(pkg)
		got, err := Read(folder, digest)
		if !errors.Is(err, c.reason) || got != nil {
			t.Errorf("%s: Read gives a package: %t, and the error %v; want %v", c.name, got != nil, err, c.reason)
			continue
		}
		if !strings.Contains(err.Error(), c.file) {
			t.Errorf("%s: the error does not name %s: %v", c.name, c.file, err)
		}
	}
}
