package modulepackage

import (
	"context"
	"fmt"
	"io/fs"
	"strings"

	"github.com/altnovatrail/spec/types/formatv0"

	"github.com/altnovatrail/module-kit/internal/canonical"
	"github.com/altnovatrail/module-kit/internal/schema"
	"github.com/altnovatrail/module-kit/module"
)

// The files of a package as files, by their paths in its folder (format/README.md, section 25):
// the canonical texts of the inventory, the canonical view and the listing of the device part, and
// the folder of the device part's files, byte for byte.
const (
	InventoryFile = "inventory.json"
	ServerFile    = "server.json"
	DeviceFile    = "device.json"
	DeviceFolder  = "device/"
)

// A Package is the package of a module: its digest, its inventory, its canonical view, the listing
// of its device part, and its files, by their paths in the folder of a package as files, each byte
// for byte as the folder holds it: InventoryFile, ServerFile, DeviceFile, and under DeviceFolder
// LICENSE, NOTICE when the module has one, and media/<path> for each media file.
type Package struct {
	Digest    formatv0.Digest
	Inventory formatv0.Inventory
	View      formatv0.CanonicalView
	Listing   formatv0.DeviceListing
	Files     map[string][]byte
}

// A Refusal is why Build builds no package: the module breaks the format, and the report of its
// check names every breach (R45 spec).
type Refusal struct {
	Report module.Report
}

func (r *Refusal) Error() string {
	return "modulepackage: the module breaks the format, so it has no package:\n" + r.Report.String()
}

// Build builds the package of the module whose folder is the root of fsys, by the format version
// its manifest names (format/README.md, section 25; R62 spec). It checks the module first, and a
// module that breaks the format gives a *Refusal with the report of its check (R45 spec). The
// solver of each template runs with ctx, and an error of ctx comes back as it is. Any other error is
// a file that could not be read, or a fault of module-kit: of a module that passed its check, the
// canonical text refuses nothing (R09).
//
// The package depends on the files of the module and on the format version alone: its files and
// folders whose names begin with ".", the eval set, eval/, and the comments of its files of parts
// stay out, and nothing is added — no version of module-kit, no time, no path.
func Build(ctx context.Context, fsys fs.FS) (*Package, error) {
	report, err := module.Check(fsys)
	if err != nil {
		return nil, err
	}
	if len(report.Breaches) > 0 {
		return nil, &Refusal{Report: report}
	}
	files, err := moduleFiles(fsys)
	if err != nil {
		return nil, err
	}
	view, err := canonicalView(ctx, fsys, files)
	if err != nil {
		return nil, err
	}
	listing, files0, err := deviceListing(fsys, files)
	if err != nil {
		return nil, err
	}
	return assemble(view, listing, files0)
}

// assemble writes the canonical texts of a canonical view and a listing, and of the inventory that
// names them, and gives the package with its files: files0 holds those of the device part already.
func assemble(view, listing map[string]any, files0 map[string][]byte) (*Package, error) {
	server, err := canonical.Text(view)
	if err != nil {
		return nil, fmt.Errorf("modulepackage: the canonical view: %w", err)
	}
	device, err := canonical.Text(listing)
	if err != nil {
		return nil, fmt.Errorf("modulepackage: the listing of the device part: %w", err)
	}
	manifest, _ := view[manifestMember].(map[string]any)
	inventory := map[string]any{
		versionField:  manifest[versionField],
		"id":          manifest["id"],
		"version":     manifest["version"],
		"server_part": descriptorValue(canonical.Descriptor(server)),
		"device_part": descriptorValue(canonical.Descriptor(device)),
	}
	text, err := canonical.Text(inventory)
	if err != nil {
		return nil, fmt.Errorf("modulepackage: the inventory: %w", err)
	}
	files0[InventoryFile] = text
	files0[ServerFile] = server
	files0[DeviceFile] = device
	return &Package{
		Digest:    canonical.Digest(text),
		Inventory: decode[formatv0.Inventory](inventory),
		View:      decode[formatv0.CanonicalView](view),
		Listing:   decode[formatv0.DeviceListing](listing),
		Files:     files0,
	}, nil
}

// moduleFiles lists the files of a module that passed its check, by their paths from its folder, in
// the order of the folder: every file but those whose name or folder begins with "." and those of
// the eval set, which stay out of the package (format/README.md, section 25).
func moduleFiles(fsys fs.FS) ([]string, error) {
	var files []string
	err := fs.WalkDir(fsys, ".", func(name string, entry fs.DirEntry, err error) error {
		switch {
		case err != nil:
			return err
		case name == ".":
		case strings.HasPrefix(entry.Name(), ".") || name == evalFolder:
			if entry.IsDir() {
				return fs.SkipDir
			}
		case !entry.IsDir():
			files = append(files, name)
		}
		return nil
	})
	return files, err
}

// descriptorValue gives a descriptor as a JSON value.
func descriptorValue(d formatv0.Descriptor) map[string]any {
	return map[string]any{"digest": string(d.Digest), "size": d.Size}
}

// decode reads a JSON value that its schema accepts into its type of formatv0, strictly (R72 spec).
// The schema accepted it, so the type reads it, and a value that it does not is a fault of
// module-kit.
func decode[T any](value any) T {
	t, err := schema.Decode[T](value)
	if err != nil {
		panic(fmt.Sprintf("modulepackage: a value that its schema accepts: %v", err))
	}
	return t
}
