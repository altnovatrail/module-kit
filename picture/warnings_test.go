package picture

import (
	"context"
	"errors"
	"io/fs"
	"path"
	"slices"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/altnovatrail/spec"
)

// Warnings reads the pictures under media/ alone, in the order of their paths, and skips what is not
// the module's and what has no transparency to check.
func TestWarnings(t *testing.T) {
	fsys := fstest.MapFS{
		"manifest.yaml":                {Data: []byte("id: x\n")},
		"media/b.png":                  {Data: pngOf(t, filled(2, 2, white, 1, 1, clear))},
		"media/a/opaque.png":           {Data: pngOf(t, filled(2, 2, white, -1, 0, white))},
		"media/a/z.webp":               {Data: losslessWebP(2, 2, half)},
		"media/c.jpg":                  {Data: []byte{0xff, 0xd8, 0xff}},
		"media/.notes/hidden.png":      {Data: []byte("not a picture")},
		"media/.hidden.png":            {Data: []byte("not a picture")},
		"tasks/not-media.png":          {Data: []byte("not a picture")},
		"media/d.png":                  {Data: []byte("\x89PNG\r\n\x1a\nnot a picture")},
		"media/e.png":                  {Data: []byte("not a picture, which module.media refuses")},
		"media/sound/cheer-up-now.mp3": {Data: []byte("ID3")},
	}
	warnings, err := Warnings(context.Background(), fsys)
	if err != nil {
		t.Fatal(err)
	}
	var paths []string
	for _, w := range warnings {
		paths = append(paths, w.Path)
		if !strings.HasPrefix(w.String(), w.Path+": warning: ") {
			t.Errorf("the warning reads %q", w)
		}
	}
	if want := []string{"media/a/z.webp", "media/b.png", "media/d.png"}; !slices.Equal(paths, want) {
		t.Errorf("warned of %q, want %q", paths, want)
	}
}

// A module with no media/ has no warnings; a folder that cannot be read, and a context that is done,
// give an error and none.
func TestWarningsWithout(t *testing.T) {
	if w, err := Warnings(context.Background(), fstest.MapFS{"manifest.yaml": {}}); err != nil || w != nil {
		t.Errorf("Warnings of a module with no media/ = %v, %v", w, err)
	}
	if _, err := Warnings(context.Background(), broken{}); !errors.Is(err, fs.ErrPermission) {
		t.Errorf("Warnings of a folder that cannot be read gives %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	fsys := fstest.MapFS{"media/a.png": {Data: pngOf(t, filled(1, 1, clear, -1, 0, clear))}}
	if w, err := Warnings(ctx, fsys); !errors.Is(err, context.Canceled) || w != nil {
		t.Errorf("Warnings with a context that is done = %v, %v", w, err)
	}
}

type broken struct{}

func (broken) Open(string) (fs.File, error) { return nil, fs.ErrPermission }

// No picture of the format's examples and test vectors is transparent or unreadable, but for the
// vectors of module.media, whose files are other formats under the extension.
func TestFormatPictures(t *testing.T) {
	checked := 0
	err := fs.WalkDir(spec.Format, ".", func(name string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() || strings.HasPrefix(name, "vectors/rules/module.media/") {
			return err
		}
		if _, ok := decoders[path.Ext(name)]; !ok {
			return nil
		}
		data, err := fs.ReadFile(spec.Format, name)
		if err != nil {
			return err
		}
		checked++
		if message, warn := Check(name, data); warn {
			t.Errorf("%s: %s", name, message)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if checked == 0 {
		t.Error("no picture of the format was checked")
	}
}
