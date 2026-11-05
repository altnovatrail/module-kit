package modulepackage

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"

	"github.com/altnovatrail/spec/types/formatv0"

	"github.com/altnovatrail/module-kit/internal/canonical"
	"github.com/altnovatrail/module-kit/internal/schema"
	"github.com/altnovatrail/module-kit/module"
)

// anySize stands for the size of the inventory, which the digest of the package names alone.
const anySize = -1

// readPart reads a file of a package that the chain names by want, and its JSON value, which is
// valid against the schema at location and of which the file is the canonical text. Of the
// inventory, the format version is read first, at the place every version keeps
// (format/README.md, section 2), so that a package of another version is refused by it.
func readPart(fsys fs.FS, file string, want formatv0.Descriptor, location string, files map[string][]byte) (any, error) {
	data, err := readFile(fsys, file, want, files)
	if err != nil {
		return nil, err
	}
	value, err := parse(data)
	if err != nil {
		return nil, fmt.Errorf("modulepackage: %s is no JSON text of the format: %v: %w", file, err, ErrNotCanonical)
	}
	if file == InventoryFile {
		object, _ := value.(map[string]any)
		if v, ok := object[versionField].(string); !ok || !module.ReadsFormat(v) {
			return nil, fmt.Errorf("modulepackage: %s names the format version %v, and module-kit reads %s alone: %w",
				file, object[versionField], module.FormatVersion, ErrFormatVersion)
		}
	}
	if errs := schema.Validate(location, value); len(errs) > 0 {
		return nil, fmt.Errorf("modulepackage: %s is not valid against %s: %s: %w", file, location, errs[0], ErrNotCanonical)
	}
	if text, err := canonical.Text(value); err != nil || !bytes.Equal(text, data) {
		return nil, fmt.Errorf("modulepackage: %s is not the canonical text of its value: %w", file, ErrNotCanonical)
	}
	return value, nil
}

// readFile reads a file of a package and holds it to the digest and size that name it, and keeps
// it in files.
func readFile(fsys fs.FS, file string, want formatv0.Descriptor, files map[string][]byte) ([]byte, error) {
	data, err := fs.ReadFile(fsys, file)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, fmt.Errorf("modulepackage: %s is missing: %w", file, ErrChain)
	}
	if err != nil {
		return nil, err
	}
	got := canonical.Descriptor(data)
	if got.Digest != want.Digest || want.Size != anySize && got.Size != want.Size {
		named := string(want.Digest)
		if want.Size != anySize {
			named = fmt.Sprintf("%s and %d bytes", want.Digest, want.Size)
		}
		return nil, fmt.Errorf("modulepackage: %s has the digest %s and %d bytes, and the chain names %s: %w",
			file, got.Digest, got.Size, named, ErrChain)
	}
	files[file] = data
	return data, nil
}
