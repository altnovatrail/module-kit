package main

import (
	"bytes"
	"context"
	"errors"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
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
		{"build with two folders", []string{"build", "a", "b", "c"}, 2, "", usage},
		{"another command", []string{"index"}, 2, "", usage},
		{"open with no flags", []string{"open"}, 2, "", usage},
		{"a flag diff does not take", []string{"diff", "--colour"}, 2, "",
			"module-kit diff: flag provided but not defined: -colour\n" + usage},
		{"an empty --previous", []string{"index", "--previous", "", "--tag", "v0.1.0", "--no-withdrawals", "--out", "o", "p"}, 2, "",
			"module-kit index: an empty path names no file: say that there is none by its own flag\n" + usage},
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

// Each command that cannot do its work says so on stderr after its own name, with no usage.
func TestRunNamesTheCommandThatFailed(t *testing.T) {
	file := filepath.Join(t.TempDir(), "module.yaml")
	if err := os.WriteFile(file, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(t.TempDir(), "out")
	notFolder := file + " is no folder: a module is the folder of its files\n"
	tests := []struct {
		args       []string
		wantStderr string
	}{
		{[]string{"check", file}, "module-kit check: " + notFolder},
		{[]string{"build", file, out}, "module-kit build: " + notFolder},
		{[]string{"index", "--first", "--tag", "v0.1.0", "--no-withdrawals", "--out", out, file}, "module-kit index: " + notFolder},
		{[]string{"diff", "--first", file}, "module-kit diff: " + notFolder},
		{[]string{"open", "--index", file, "--latest", file, "--seed", "1", "--out", out}, "module-kit open: the index, " + file + ": "},
	}
	for _, tt := range tests {
		t.Run(tt.args[0], func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			if status := run(context.Background(), tt.args, &stdout, &stderr); status != statusFailed {
				t.Errorf("run(%q) = %d, want %d", tt.args, status, statusFailed)
			}
			if stdout.Len() > 0 {
				t.Errorf("run(%q) wrote %q to stdout, want nothing", tt.args, stdout.String())
			}
			if got := stderr.String(); !strings.HasPrefix(got, tt.wantStderr) || strings.Contains(got, usage) {
				t.Errorf("run(%q) wrote %q to stderr, want %q at its start and no usage", tt.args, got, tt.wantStderr)
			}
		})
	}
}

func TestUsageOf(t *testing.T) {
	wrong := errors.New("given twice")
	tests := []struct {
		err  error
		want string
	}{
		{nil, "usage"},
		{errUsage, "usage"},
		{errors.Join(errUsage, wrong), "usage"},
		{wrong, "usage\ngiven twice"},
	}
	for _, tt := range tests {
		if got := usageOf(tt.err).Error(); got != tt.want {
			t.Errorf("usageOf(%v).Error() = %q, want %q", tt.err, got, tt.want)
		}
	}
}

func TestCurrentVersionTakesTheReleaseVersion(t *testing.T) {
	defer func(saved string) { version = saved }(version)
	version = "v0.1.0"
	if got := currentVersion(); got != "v0.1.0" {
		t.Errorf("currentVersion() = %q with version set to v0.1.0, want v0.1.0", got)
	}
}
