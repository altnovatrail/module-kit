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

	"github.com/altnovatrail/module-kit/internal/canonical"
	"github.com/altnovatrail/module-kit/internal/yaml"
)

// vectorOf reads a YAML file of spec.Format into its JSON value.
func vectorOf(t testing.TB, name string) map[string]any {
	t.Helper()
	data, err := fs.ReadFile(spec.Format, name)
	if err != nil {
		t.Fatal(err)
	}
	value, err := yaml.Read(data)
	if err != nil {
		t.Fatal(err)
	}
	return value.(map[string]any)
}

// buildPackages builds the package of each module folder under root of spec.Format into a folder of
// its own, by the command build, and gives the folders.
func buildPackages(t *testing.T, root string) []string {
	t.Helper()
	entries, err := fs.ReadDir(spec.Format, root)
	if err != nil {
		t.Fatal(err)
	}
	var folders []string
	for _, e := range entries {
		module := fromSpec(t, path.Join(root, e.Name()))
		out := filepath.Join(t.TempDir(), e.Name())
		var stdout, stderr bytes.Buffer
		if status := run(context.Background(), []string{"build", module, out}, &stdout, &stderr); status != statusAccepted {
			t.Fatalf("build %s: status %d, %s%s", e.Name(), status, stdout.String(), stderr.String())
		}
		folders = append(folders, out)
	}
	return folders
}

// writeTemp writes data to a new file named name in a temporary folder and gives its path.
func writeTemp(t *testing.T, name string, data []byte) string {
	t.Helper()
	file := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(file, data, 0o644); err != nil {
		t.Fatal(err)
	}
	return file
}

// Each test vector of the module index, run through the command as the release of modules runs it:
// its modules built as packages, its previous index a file of its canonical text or --first, its
// withdrawals a YAML file or --no-withdrawals. The index written is the canonical text of the
// vector's, byte for byte, and a refusal exits 1 with the vector's rule on every line.
func TestIndexVectors(t *testing.T) {
	cases, err := fs.ReadDir(spec.Format, "vectors/indexes")
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range cases {
		t.Run(c.Name(), func(t *testing.T) {
			root := path.Join("vectors/indexes", c.Name())
			vector := vectorOf(t, path.Join(root, "index.yaml"))
			args := []string{"index", "--tag", vector["release"].(string)}
			if previous, ok := vector["previous"]; ok {
				text, err := canonical.Text(previous)
				if err != nil {
					t.Fatal(err)
				}
				args = append(args, "--previous", writeTemp(t, "module-index.json", text))
			} else {
				args = append(args, "--first")
			}
			if withdrawals, ok := vector["withdrawals"]; ok {
				// JSON is YAML by the rules of section 4, with every key in quotes, as they ask of a version.
				data, err := canonical.Text(withdrawals)
				if err != nil {
					t.Fatal(err)
				}
				args = append(args, "--withdrawals", writeTemp(t, "withdrawals.yaml", data))
			} else {
				args = append(args, "--no-withdrawals")
			}
			out := filepath.Join(t.TempDir(), "module-index.json")
			args = append(append(args, "--out", out), buildPackages(t, path.Join(root, "modules"))...)
			var stdout, stderr bytes.Buffer
			status := run(context.Background(), args, &stdout, &stderr)
			if rule, refused := vector["refusal"].(string); refused {
				lines := strings.Split(strings.TrimSuffix(stdout.String(), "\n"), "\n")
				if status != statusRefused || len(lines) < 2 || stderr.Len() > 0 {
					t.Fatalf("status %d, stdout %q, stderr %q; want 1 and %s", status, stdout.String(), stderr.String(), rule)
				}
				for _, line := range lines[:len(lines)-1] {
					if !strings.HasPrefix(line, rule+": ") {
						t.Errorf("the line %q is not of the rule %s", line, rule)
					}
				}
				if _, err := os.Stat(out); !os.IsNotExist(err) {
					t.Errorf("a refused release wrote %s", out)
				}
				return
			}
			if status != statusAccepted || stderr.Len() > 0 {
				t.Fatalf("status %d, stdout %q, stderr %q", status, stdout.String(), stderr.String())
			}
			want, err := canonical.Text(vector["index"])
			if err != nil {
				t.Fatal(err)
			}
			if got, err := os.ReadFile(out); err != nil || !bytes.Equal(got, want) {
				t.Errorf("the index is\n%s\nand the vector's is\n%s (%v)", got, want, err)
			}
		})
	}
}
