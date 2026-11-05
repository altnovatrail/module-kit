package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"

	"github.com/altnovatrail/spec"

	"github.com/altnovatrail/module-kit/internal/canonical"
	"github.com/altnovatrail/module-kit/modulepackage"
)

// publishedFolders gives a package as files for each published version of a test vector of versions
// — built by the command build from a module's folder, or the vector's own package of an earlier
// version of the format — and the text of a module index that names each by its digest, as the CI of
// modules hands diff the index of the newest release and the packages it names.
func publishedFolders(t *testing.T, root string) (folders []string, indexFile string) {
	t.Helper()
	entries, err := fs.ReadDir(spec.Format, path.Join(root, "published"))
	if err != nil {
		t.Fatal(err)
	}
	versions := map[string]any{}
	id := ""
	for _, e := range entries {
		src := path.Join(root, "published", e.Name())
		folder := filepath.Join(t.TempDir(), e.Name())
		if _, err := fs.Stat(spec.Format, path.Join(src, "manifest.yaml")); err == nil {
			var stdout, stderr bytes.Buffer
			if status := run(context.Background(), []string{"build", fromSpec(t, src), folder}, &stdout, &stderr); status != statusAccepted {
				t.Fatalf("build %s: %d %s%s", src, status, stdout.String(), stderr.String())
			}
		} else {
			folder = fromSpec(t, src)
		}
		inventory, err := os.ReadFile(filepath.Join(folder, modulepackage.InventoryFile))
		if err != nil {
			t.Fatal(err)
		}
		var inv map[string]any
		if err := json.Unmarshal(inventory, &inv); err != nil {
			t.Fatal(err)
		}
		id = inv["id"].(string)
		versions[inv["version"].(string)] = map[string]any{
			"module_format_version": inv["module_format_version"],
			"package_digest":        string(canonical.Digest(inventory)),
			"release":               "v0.1.0",
		}
		folders = append(folders, folder)
	}
	index := map[string]any{"module_format_version": "0.3", "modules": map[string]any{id: versions}}
	text, err := canonical.Text(index)
	if err != nil {
		t.Fatal(err)
	}
	return folders, writeTemp(t, "module-index.json", text)
}

// diffArgs gives the command line of diff on a module against an index and package folders.
func diffArgs(indexFile string, folders []string, module string, more ...string) []string {
	args := append([]string{"diff"}, more...)
	args = append(args, "--index", indexFile)
	for _, f := range folders {
		args = append(args, "--published", f)
	}
	return append(args, module)
}

// Each test vector of versions, run through the command: the part of the version it gives, status 0,
// on the line "found:"; or the rule of its refusal, status 1, on a line "refused:" and on no other.
func TestDiffVectors(t *testing.T) {
	cases, err := fs.ReadDir(spec.Format, "vectors/versions")
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range cases {
		t.Run(c.Name(), func(t *testing.T) {
			root := path.Join("vectors/versions", c.Name())
			vector := vectorOf(t, path.Join(root, "version.yaml"))
			folders, indexFile := publishedFolders(t, root)
			var stdout, stderr bytes.Buffer
			status := run(context.Background(), diffArgs(indexFile, folders, fromSpec(t, path.Join(root, "module"))), &stdout, &stderr)
			report := stdout.String()
			if rule, refused := vector["refusal"].(string); refused {
				// A refusal breaks its rule and no other, perhaps at more than one published version.
				if status != statusRefused || strings.Count(report, "\nrefused: "+rule+": ") == 0 ||
					strings.Count(report, "\nrefused: ") != strings.Count(report, "\nrefused: "+rule+": ") {
					t.Errorf("status %d, stdout %q, stderr %q; want 1 and refused by %s alone", status, report, stderr.String(), rule)
				}
				return
			}
			part := vector["version_part"].(string)
			found := strings.Contains(report, "\nfound: "+part+"\n") ||
				part == "none" && strings.HasSuffix(report, ": the newest published version as it was, so nothing is compared\n")
			if status != statusAccepted || stderr.Len() > 0 || !found || strings.Contains(report, "refused:") {
				t.Errorf("status %d, stdout %q, stderr %q; want 0 and found: %s", status, report, stderr.String(), part)
			}
		})
	}
}
