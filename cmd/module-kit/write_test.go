package main

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// A module that the format refuses has no package: status 1, its report, and nothing on the disk.
func TestBuildRefuses(t *testing.T) {
	dir := fromSpec(t, "vectors/rules/task.known-band/unknown-band")
	out := filepath.Join(t.TempDir(), "package")
	var stdout, stderr bytes.Buffer
	status := run(context.Background(), []string{"build", dir, out}, &stdout, &stderr)
	if status != statusRefused || !strings.Contains(stdout.String(), ": task.known-band: ") || stderr.Len() > 0 {
		t.Errorf("status %d, stdout %q, stderr %q; want 1 and task.known-band", status, stdout.String(), stderr.String())
	}
	if _, err := os.Stat(out); !os.IsNotExist(err) {
		t.Errorf("a refused module left %s: %v", out, err)
	}
}

// What build cannot do is status 2, said on stderr, and leaves the folder as it was: a folder that
// is not empty, a module that is no folder, a folder whose parent does not exist, and a context that
// is done.
func TestBuildFails(t *testing.T) {
	dir := fromSpec(t, "examples/module/valid/with-templates")
	full := t.TempDir()
	notes := filepath.Join(full, "notes.txt")
	if err := os.WriteFile(notes, []byte("mine"), 0o644); err != nil {
		t.Fatal(err)
	}
	done, cancel := context.WithCancel(context.Background())
	cancel()
	for _, c := range []struct {
		name     string
		ctx      context.Context
		dir, out string
		want     string
	}{
		{"a folder that is not empty", context.Background(), dir, full, "is not empty"},
		{"a module that is no folder", context.Background(), notes, filepath.Join(t.TempDir(), "p"), "is no folder"},
		{"a module that does not exist", context.Background(), filepath.Join(full, "absent"), filepath.Join(t.TempDir(), "p"), "absent"},
		{"no parent", context.Background(), dir, filepath.Join(t.TempDir(), "absent", "p"), "absent"},
		{"an output that is a file", context.Background(), dir, notes, "notes.txt"},
		{"a context that is done", done, dir, filepath.Join(t.TempDir(), "p"), "context canceled"},
	} {
		t.Run(c.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			status := run(c.ctx, []string{"build", c.dir, c.out}, &stdout, &stderr)
			if status != statusFailed || stdout.Len() > 0 || !strings.Contains(stderr.String(), c.want) {
				t.Errorf("status %d, stdout %q, stderr %q; want 2 and %q on stderr", status, stdout.String(), stderr.String(), c.want)
			}
		})
	}
	if data, err := os.ReadFile(notes); err != nil || string(data) != "mine" {
		t.Errorf("the folder that was not empty changed: %q, %v", data, err)
	}
}

// writePackage takes no path that leaves the folder of the package, and leaves nothing behind when
// it refuses one.
func TestWritePackageRefusesPaths(t *testing.T) {
	for _, name := range []string{"../escape", "/root", `device\LICENSE`, ""} {
		parent := t.TempDir()
		out := filepath.Join(parent, "package")
		if err := writePackage(out, map[string][]byte{name: nil}); err == nil {
			t.Errorf("writePackage took the path %q", name)
		}
		if entries, _ := os.ReadDir(parent); len(entries) > 0 {
			t.Errorf("writePackage left %d entries for the path %q", len(entries), name)
		}
	}
}
