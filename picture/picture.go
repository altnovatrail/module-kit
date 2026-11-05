package picture

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"image"
	"image/png"
	"io/fs"
	"path"
	"strings"

	"golang.org/x/image/webp"
)

// MaxPixels is the most pixels a picture may have for Warnings to decode it: 1<<25, 33 554 432, such
// as 5792 × 5792. A picture decodes into at most 8 bytes a pixel, so one takes at most 256 MiB, and a
// file whose header names a larger size, as a picture packed to explode would, is not decoded but
// warned of (R39).
const MaxPixels = 1 << 25

// maxPixels is MaxPixels, which the fuzz test lowers.
var maxPixels = MaxPixels

// The folder of a module's media files and the extensions of the pictures that may hold
// transparency (format/README.md, section 10).
const mediaFolder = "media"

// decoders holds, by extension, the format's name, whether a file begins with its signature as
// module.media reads it (format/README.md, section 10), and how to read its size and decode it.
var decoders = map[string]struct {
	format    string
	signature func(data []byte) bool
	config    func(data []byte) (image.Config, error)
	decode    func(data []byte) (image.Image, error)
}{
	".png": {"PNG",
		func(data []byte) bool { return bytes.HasPrefix(data, []byte("\x89PNG\r\n\x1a\n")) },
		func(data []byte) (image.Config, error) { return png.DecodeConfig(bytes.NewReader(data)) },
		func(data []byte) (image.Image, error) { return png.Decode(bytes.NewReader(data)) }},
	".webp": {"WebP",
		func(data []byte) bool {
			return len(data) >= 14 && string(data[:4]) == "RIFF" && string(data[8:14]) == "WEBPVP"
		},
		func(data []byte) (image.Config, error) { return webp.DecodeConfig(bytes.NewReader(data)) },
		func(data []byte) (image.Image, error) { return webp.Decode(bytes.NewReader(data)) }},
}

// A Warning is what a picture of a module asks of its author, though the format refuses nothing for
// it: Path is the file from the module's folder, with / between names, and Message says what.
type Warning struct {
	Path    string
	Message string
}

func (w Warning) String() string {
	return w.Path + ": warning: " + w.Message
}

// Warnings reads each PNG and WebP file under media/ of the module whose folder is the root of fsys,
// in the order of their paths, and gives a warning for each picture with a pixel that is not fully
// opaque, and for each it could not check. A module with no media/ has none. Files and folders whose
// names begin with "." are not the module's (format/README.md, section 10) and are skipped. The
// error is for a folder or a file that could not be read, or ctx once it is done, and then there are
// no warnings.
func Warnings(ctx context.Context, fsys fs.FS) ([]Warning, error) {
	var warnings []Warning
	err := fs.WalkDir(fsys, mediaFolder, func(name string, entry fs.DirEntry, err error) error {
		if err != nil {
			if name == mediaFolder && errors.Is(err, fs.ErrNotExist) {
				return fs.SkipDir
			}
			return err
		}
		if strings.HasPrefix(entry.Name(), ".") && name != mediaFolder {
			if entry.IsDir() {
				return fs.SkipDir
			}
			return nil
		}
		if entry.IsDir() {
			return nil
		}
		if _, ok := decoders[path.Ext(name)]; !ok {
			return nil
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		data, err := fs.ReadFile(fsys, name)
		if err != nil {
			return err
		}
		if message, warn := Check(name, data); warn {
			warnings = append(warnings, Warning{Path: name, Message: message})
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return warnings, nil
}

// Check reads one picture, the bytes of a file named name, by the format its extension names, .png
// or .webp in lowercase, as media_path has them, and says what to warn of, if anything: a pixel that is not fully opaque, the first in
// the order of rows; a size above MaxPixels; or a file it cannot decode. A file of another extension
// gives no warning, and neither does one that does not begin with the signature of its format: the
// check of the module refuses it by module.media, and a warning would say it twice.
func Check(name string, data []byte) (message string, warn bool) {
	d, ok := decoders[path.Ext(name)]
	if !ok || !d.signature(data) {
		return "", false
	}
	config, err := d.config(data)
	if err != nil {
		return fmt.Sprintf("module-kit cannot read it as %s, so it could not check that every pixel is opaque: %v", d.format, err), true
	}
	if pixels := int64(config.Width) * int64(config.Height); pixels > int64(maxPixels) {
		return fmt.Sprintf("it is %d × %d pixels, more than the %d that module-kit decodes, so it could not check that every pixel is opaque",
			config.Width, config.Height, maxPixels), true
	}
	m, err := d.decode(data)
	if err != nil {
		return fmt.Sprintf("module-kit cannot decode it as %s, so it could not check that every pixel is opaque: %v", d.format, err), true
	}
	if x, y, a, ok := firstTransparent(m); ok {
		return fmt.Sprintf("the pixel at x = %d, y = %d has alpha %d of 65535, not fully opaque: a picture SHOULD have an opaque, light background, "+
			"or it may vanish on a dark sheet and in forced colours (format/README.md, section 14)", x, y, a), true
	}
	return "", false
}

// firstTransparent gives the first pixel of m, in the order of rows, whose alpha is below full, with
// its alpha of 0xffff, and false when every pixel is opaque. An image that knows it is opaque says so
// at once.
func firstTransparent(m image.Image) (x, y int, alpha uint32, found bool) {
	if o, ok := m.(interface{ Opaque() bool }); ok && o.Opaque() {
		return 0, 0, 0, false
	}
	b := m.Bounds()
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			if _, _, _, a := m.At(x, y).RGBA(); a != 0xffff {
				return x - b.Min.X, y - b.Min.Y, a, true
			}
		}
	}
	return 0, 0, 0, false
}
