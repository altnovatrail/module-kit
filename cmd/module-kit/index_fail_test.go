package main

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// What index cannot build is status 2, said on stderr, and writes nothing: a command line it does
// not take, a previous index or withdrawals it cannot read, a package folder that is none, an out
// that exists already.
func TestIndexFails(t *testing.T) {
	packages := buildPackages(t, "vectors/indexes/first-release/modules")
	exists := writeTemp(t, "module-index.json", []byte("{}"))
	notIndex := writeTemp(t, "previous.json", []byte("not JSON"))
	badWithdrawals := writeTemp(t, "withdrawals.yaml", []byte("\"example/smallest\": {\"0.1.0\": \"lost\"}\n"))
	notYAML := writeTemp(t, "withdrawals.yaml", []byte("a: [\n"))
	out := func() string { return filepath.Join(t.TempDir(), "module-index.json") }
	base := []string{"--tag", "v0.1.0", "--no-withdrawals"}
	for _, c := range []struct {
		name string
		args []string
		want string
	}{
		{"neither --previous nor --first", append([]string{"--out", out()}, append(base, packages...)...), "usage:"},
		{"both --previous and --first", append([]string{"--first", "--previous", exists, "--out", out()}, append(base, packages...)...), "usage:"},
		{"an empty --previous", append([]string{"--previous", "", "--out", out()}, append(base, packages...)...), "an empty path names no file"},
		{"--previous twice", append([]string{"--previous", exists, "--previous", exists, "--out", out()}, append(base, packages...)...), "given twice"},
		{"neither --withdrawals nor --no-withdrawals", append([]string{"--first", "--tag", "v0.1.0", "--out", out()}, packages...), "usage:"},
		{"no package", append([]string{"--first", "--out", out()}, base...), "usage:"},
		{"no --out", append([]string{"--first"}, append(base, packages...)...), "usage:"},
		{"no --tag", append([]string{"--first", "--no-withdrawals", "--out", out()}, packages...), "usage:"},
		{"an unknown flag", append([]string{"--first", "--colour", "--out", out()}, append(base, packages...)...), "usage:"},
		{"a previous index that does not exist", append([]string{"--previous", filepath.Join(t.TempDir(), "none.json"), "--out", out()}, append(base, packages...)...), "the previous index"},
		{"a previous index that is no index", append([]string{"--previous", notIndex, "--out", out()}, append(base, packages...)...), "the previous index cannot be read"},
		{"withdrawals that do not exist", append([]string{"--first", "--tag", "v0.1.0", "--withdrawals", filepath.Join(t.TempDir(), "none.yaml"), "--out", out()}, packages...), "the withdrawals"},
		{"withdrawals that are not YAML", append([]string{"--first", "--tag", "v0.1.0", "--withdrawals", notYAML, "--out", out()}, packages...), "the withdrawals"},
		{"withdrawals with an unknown reason", append([]string{"--first", "--tag", "v0.1.0", "--withdrawals", badWithdrawals, "--out", out()}, packages...), "not in the form of $defs/withdrawals"},
		{"a tag that is none", append([]string{"--first", "--tag", "0.1", "--no-withdrawals", "--out", out()}, packages...), "not one of a release"},
		{"a folder that is no package", append([]string{"--first", "--out", out()}, append(base, t.TempDir())...), "holds no package"},
		{"an out that exists", append([]string{"--first", "--out", exists}, append(base, packages...)...), "exists already"},
		{"an out in no folder", append([]string{"--first", "--out", filepath.Join(t.TempDir(), "none", "module-index.json")}, append(base, packages...)...), "none"},
	} {
		t.Run(c.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			status := run(context.Background(), append([]string{"index"}, c.args...), &stdout, &stderr)
			if status != statusFailed || stdout.Len() > 0 || !strings.Contains(stderr.String(), c.want) {
				t.Errorf("status %d, stdout %q, stderr %q; want 2 and %q", status, stdout.String(), stderr.String(), c.want)
			}
		})
	}
	if data, err := os.ReadFile(exists); err != nil || string(data) != "{}" {
		t.Errorf("an out that existed was changed: %q, %v", data, err)
	}
}

// index writes the index where nothing was, and leaves nothing beside it.
func TestIndexWrites(t *testing.T) {
	packages := buildPackages(t, "vectors/indexes/first-release/modules")
	parent := t.TempDir()
	out := filepath.Join(parent, "module-index.json")
	var stdout, stderr bytes.Buffer
	args := append([]string{"index", "--first", "--tag", "v0.1.0", "--no-withdrawals", "--out", out}, packages...)
	if status := run(context.Background(), args, &stdout, &stderr); status != statusAccepted {
		t.Fatalf("status %d, %s", status, stderr.String())
	}
	if want := "the module index of v0.1.0, of "; !strings.HasPrefix(stdout.String(), want) {
		t.Errorf("stdout %q, want it to begin %q", stdout.String(), want)
	}
	if entries, _ := os.ReadDir(parent); len(entries) != 1 {
		t.Errorf("index left %d entries beside the file, want none", len(entries)-1)
	}
}
