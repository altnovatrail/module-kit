package main

import (
	"bytes"
	"context"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"

	"github.com/altnovatrail/spec"
	"github.com/altnovatrail/spec/types/formatv0"

	"github.com/altnovatrail/module-kit/internal/canonical"
	"github.com/altnovatrail/module-kit/internal/schema"
	"github.com/altnovatrail/module-kit/internal/yaml"
)

// The package of each test vector of packages, built as files, holds inventory.json, server.json and
// device.json byte for byte as the canonical texts of the vector, and under device/ LICENSE, NOTICE
// and media/ byte for byte as the module has them, and nothing else (format/README.md, section 25).
func TestBuildVectors(t *testing.T) {
	entries, err := fs.ReadDir(spec.Format, "vectors/packages")
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		t.Run(e.Name(), func(t *testing.T) {
			root := path.Join("vectors/packages", e.Name())
			dir := fromSpec(t, path.Join(root, "module"))
			out := filepath.Join(t.TempDir(), "package")
			var stdout, stderr bytes.Buffer
			status := run(context.Background(), []string{"build", dir, out}, &stdout, &stderr)
			want, digest := wantPackage(t, root, dir)
			if status != statusAccepted || stderr.Len() > 0 || !strings.Contains(stdout.String(), digest) {
				t.Fatalf("status %d, stdout %q, stderr %q; want 0 and the digest %s", status, stdout.String(), stderr.String(), digest)
			}
			got := filesUnder(t, out)
			if len(got) != len(want) {
				t.Errorf("the package holds %d files, and the vector's %d", len(got), len(want))
			}
			for name, data := range want {
				if !bytes.Equal(got[name], data) {
					t.Errorf("%s is\n%s\nand the vector's is\n%s", name, got[name], data)
				}
			}
		})
	}
}

// wantPackage reads package.yaml of the vector root and gives the files of its package: the
// canonical texts of its three members, and the files of the device part that its listing names, as
// the module at dir has them; and its digest.
func wantPackage(t *testing.T, root, dir string) (map[string][]byte, string) {
	t.Helper()
	data, err := fs.ReadFile(spec.Format, path.Join(root, "package.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	value, err := yaml.Read(data)
	if err != nil {
		t.Fatal(err)
	}
	vector, _ := value.(map[string]any)
	files := map[string][]byte{}
	for member, file := range map[string]string{"inventory": "inventory.json", "server_part": "server.json", "device_part": "device.json"} {
		if files[file], err = canonical.Text(vector[member]); err != nil {
			t.Fatal(err)
		}
	}
	listing, err := schema.Decode[formatv0.DeviceListing](vector["device_part"])
	if err != nil {
		t.Fatal(err)
	}
	names := []string{"LICENSE"}
	if listing.Notice != nil {
		names = append(names, "NOTICE")
	}
	for name := range listing.Media {
		names = append(names, "media/"+string(name))
	}
	for _, name := range names {
		if files["device/"+name], err = os.ReadFile(filepath.Join(dir, filepath.FromSlash(name))); err != nil {
			t.Fatal(err)
		}
	}
	digest, _ := vector["digest"].(string)
	return files, digest
}

// filesUnder gives every file under root by its path with / between names.
func filesUnder(t *testing.T, root string) map[string][]byte {
	t.Helper()
	files := map[string][]byte{}
	err := filepath.WalkDir(root, func(file string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		rel, err := filepath.Rel(root, file)
		if err != nil {
			return err
		}
		files[filepath.ToSlash(rel)], err = os.ReadFile(file)
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	return files
}

// build writes the package into an empty folder that exists, in its place, and leaves nothing beside
// it.
func TestBuildIntoAnEmptyFolder(t *testing.T) {
	parent := t.TempDir()
	out := filepath.Join(parent, "package")
	if err := os.Mkdir(out, 0o755); err != nil {
		t.Fatal(err)
	}
	dir := fromSpec(t, "examples/module/valid/smallest")
	var stdout, stderr bytes.Buffer
	if status := run(context.Background(), []string{"build", dir, out}, &stdout, &stderr); status != statusAccepted {
		t.Fatalf("status %d, stderr %q", status, stderr.String())
	}
	got := filesUnder(t, out)
	for _, name := range []string{"inventory.json", "server.json", "device.json", "device/LICENSE"} {
		if _, ok := got[name]; !ok {
			t.Errorf("the package holds no %s", name)
		}
	}
	if entries, _ := os.ReadDir(parent); len(entries) != 1 {
		t.Errorf("build left %d entries beside the package, want none", len(entries)-1)
	}
	// The folder of the package is made as its other folders are, not as the temporary folder that
	// only its owner may read.
	inner, err := os.Stat(filepath.Join(out, "device"))
	if err != nil {
		t.Fatal(err)
	}
	if info, err := os.Stat(out); err != nil || info.Mode().Perm() != inner.Mode().Perm() {
		t.Errorf("the folder of the package is %v, %v; want %v, as device/", info.Mode(), err, inner.Mode())
	}
}
