package main

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The cases of diff that architecture 3.5 names besides the vectors: a first version by --first and
// by an index that does not name the module, a package for the new version, a module that its check
// refuses, and --json.
func TestDiffCases(t *testing.T) {
	root := "vectors/versions/media-changed"
	folders, indexFile := publishedFolders(t, root)
	module := fromSpec(t, root+"/module")
	otherIndex := writeTemp(t, "module-index.json", []byte(`{"module_format_version":"0.4","modules":{"example/smallest":{"0.1.0":`+
		`{"module_format_version":"0.4","package_digest":"sha256:fcbe6e68430776b0506c93ab51b155f8e7f92e2b3ecbe03e34e2afb91b4d4d08","release":"v0.1.0"}}}}`))
	built := filepath.Join(t.TempDir(), "package")
	var out, errs bytes.Buffer
	if status := run(context.Background(), []string{"build", module, built}, &out, &errs); status != statusAccepted {
		t.Fatalf("build: %d %s", status, errs.String())
	}
	refused := fromSpec(t, root+"/module")
	if err := os.Remove(filepath.Join(refused, "LICENSE")); err != nil {
		t.Fatal(err)
	}
	for _, c := range []struct {
		name   string
		args   []string
		status int
		want   string
	}{
		{"--first", []string{"diff", "--first", module}, statusAccepted, ": the first version, with nothing to compare it with\n"},
		{"an index that does not name the module", []string{"diff", "--index", otherIndex, module}, statusAccepted, ": the first version, with nothing to compare it with\n"},
		{"the new version as a package", diffArgs(indexFile, folders, built), statusAccepted, "\nfound: patch\n"},
		{"a module that its check refuses", diffArgs(indexFile, folders, refused), statusRefused, ": module.license: "},
		{"a folder of another package among --published", diffArgs(indexFile, append(folders, built), module), statusAccepted, "\nfound: patch\n"},
	} {
		t.Run(c.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			status := run(context.Background(), c.args, &stdout, &stderr)
			if status != c.status || stderr.Len() > 0 || !strings.Contains(stdout.String(), c.want) {
				t.Errorf("status %d, stdout %q, stderr %q; want %d and %q", status, stdout.String(), stderr.String(), c.status, c.want)
			}
		})
	}
}

// --json writes the comparison as data, and a module that its check refuses as its breaches.
func TestDiffJSON(t *testing.T) {
	root := "vectors/versions/media-changed"
	folders, indexFile := publishedFolders(t, root)
	var stdout, stderr bytes.Buffer
	if status := run(context.Background(), diffArgs(indexFile, folders, fromSpec(t, root+"/module"), "--json"), &stdout, &stderr); status != statusAccepted {
		t.Fatalf("status %d, %s", status, stderr.String())
	}
	var c map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &c); err != nil || c["version_part"] != "patch" || c["newest"] != "1.0.0" {
		t.Errorf("--json gives %s (%v)", stdout.String(), err)
	}
	refused := fromSpec(t, root+"/module")
	if err := os.Remove(filepath.Join(refused, "LICENSE")); err != nil {
		t.Fatal(err)
	}
	stdout.Reset()
	if status := run(context.Background(), []string{"diff", "--json", "--first", refused}, &stdout, &stderr); status != statusRefused {
		t.Fatalf("status %d, %s", status, stderr.String())
	}
	var r struct {
		Check []checkBreach `json:"check"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &r); err != nil || len(r.Check) != 1 || r.Check[0].Rule != "module.license" ||
		r.Check[0].Path != filepath.ToSlash(refused) {
		t.Errorf("--json of a refused module gives %s (%v)", stdout.String(), err)
	}
}

// What diff cannot compare is status 2, said on stderr: a command line it does not take, an index it
// cannot read, a version of the index whose package no --published folder holds, a broken chain of
// digests, and a folder that is neither a module nor a package.
func TestDiffFails(t *testing.T) {
	root := "vectors/versions/media-changed"
	folders, indexFile := publishedFolders(t, root)
	module := fromSpec(t, root+"/module")
	broken := filepath.Join(t.TempDir(), "broken")
	if err := os.CopyFS(broken, os.DirFS(folders[0])); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(broken, "server.json"), []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}
	notIndex := writeTemp(t, "module-index.json", []byte("[]"))
	for _, c := range []struct {
		name string
		args []string
		want string
	}{
		{"neither --index nor --first", []string{"diff", module}, "usage:"},
		{"both --index and --first", []string{"diff", "--first", "--index", indexFile, module}, "usage:"},
		{"an empty --index", []string{"diff", "--index", "", module}, "an empty path names no file"},
		{"no module", []string{"diff", "--first"}, "usage:"},
		{"two modules", []string{"diff", "--first", module, module}, "usage:"},
		{"an unknown flag", []string{"diff", "--colour", "--first", module}, "usage:"},
		{"an index that does not exist", []string{"diff", "--index", filepath.Join(t.TempDir(), "none.json"), module}, "the index"},
		{"an index that is none", []string{"diff", "--index", notIndex, module}, "no module index"},
		{"no package of a published version", diffArgs(indexFile, nil, module), "no --published folder holds it"},
		{"a --published folder that is no package", diffArgs(indexFile, append(folders, t.TempDir()), module), "holds no package"},
		{"a broken chain of digests", diffArgs(indexFile, []string{broken}, module), "not the file the chain of digests names"},
		{"a folder that is neither", []string{"diff", "--first", t.TempDir()}, "neither a module nor a package"},
		{"a module that does not exist", []string{"diff", "--first", filepath.Join(t.TempDir(), "none")}, "no such file"},
	} {
		t.Run(c.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			status := run(context.Background(), c.args, &stdout, &stderr)
			if status != statusFailed || stdout.Len() > 0 || !strings.Contains(stderr.String(), c.want) {
				t.Errorf("status %d, stdout %q, stderr %q; want 2 and %q", status, stdout.String(), stderr.String(), c.want)
			}
		})
	}
}
