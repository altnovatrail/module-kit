package main

import (
	"bytes"
	"testing"
)

func TestRun(t *testing.T) {
	const usage = "usage: module-kit --version\n"
	tests := []struct {
		name       string
		args       []string
		wantStatus int
		wantStdout string
		wantStderr string
	}{
		{"version", []string{"--version"}, 0, "module-kit dev\n", ""},
		{"no arguments", nil, 2, "", usage},
		{"another flag", []string{"--help"}, 2, "", usage},
		{"version and more", []string{"--version", "module.yaml"}, 2, "", usage},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			if status := run(tt.args, &stdout, &stderr); status != tt.wantStatus {
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
