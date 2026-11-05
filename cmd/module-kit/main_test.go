package main

import (
	"bytes"
	"context"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"testing"

	"github.com/altnovatrail/spec"

	"github.com/altnovatrail/module-kit/verdict"
)

// fromSpec copies the folder root of spec.Format into a new temporary folder and gives its path: the
// CLI reads a module from the disk, and CI checks out no submodule.
func fromSpec(t testing.TB, root string) string {
	t.Helper()
	dir := filepath.Join(t.TempDir(), path.Base(root))
	sub, err := fs.Sub(spec.Format, root)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.CopyFS(dir, sub); err != nil {
		t.Fatal(err)
	}
	return dir
}

func TestRun(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		wantStatus int
		wantStdout string
		wantStderr string
	}{
		{"version", []string{"--version"}, 0, "module-kit dev\nUnicode " + verdict.UnicodeVersion + "\n", ""},
		{"no arguments", nil, 2, "", usage},
		{"another flag", []string{"--help"}, 2, "", usage},
		{"version and more", []string{"--version", "module.yaml"}, 2, "", usage},
		{"check with no module", []string{"check"}, 2, "", usage},
		{"check with two modules", []string{"check", "a", "b"}, 2, "", usage},
		{"build with no folder", []string{"build", "a"}, 2, "", usage},
		{"another command", []string{"index"}, 2, "", usage},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			if status := run(context.Background(), tt.args, &stdout, &stderr); status != tt.wantStatus {
				t.Errorf("run(%q) = %d, want %d", tt.args, status, tt.wantStatus)
			}
			if got := stdout.String(); got != tt.wantStdout {
				t.Errorf("run(%q) wrote %q to stdout, want %q", tt.args, got, tt.wantStdout)
			}
			if got := stderr.String(); got != tt.wantStderr {
				t.Errorf("run(%q) wrote %q to stderr, want %q", tt.args, got, tt.wantStderr)
			}
		})
	}
}

func TestCurrentVersionTakesTheReleaseVersion(t *testing.T) {
	defer func(saved string) { version = saved }(version)
	version = "v0.1.0"
	if got := currentVersion(); got != "v0.1.0" {
		t.Errorf("currentVersion() = %q with version set to v0.1.0, want v0.1.0", got)
	}
}
