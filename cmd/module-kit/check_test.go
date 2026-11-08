package main

import (
	"bytes"
	"context"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"

	"github.com/altnovatrail/spec"
)

// Each module of the examples that the format accepts is accepted, with status 0 and one line.
func TestCheckAccepts(t *testing.T) {
	entries, err := fs.ReadDir(spec.Format, "examples/module/valid")
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		t.Run(e.Name(), func(t *testing.T) {
			dir := fromSpec(t, "examples/module/valid/"+e.Name())
			var stdout, stderr bytes.Buffer
			status := run(context.Background(), []string{"check", dir}, &stdout, &stderr)
			want := dir + ": the module breaks no rule of the format 0.3\n"
			if status != statusAccepted || stdout.String() != want || stderr.Len() > 0 {
				t.Errorf("status %d, stdout %q, stderr %q; want 0 and %q", status, stdout.String(), stderr.String(), want)
			}
		})
	}
}

// Each test vector of a rule that is a module's folder is refused with status 1: each line of the
// report names its place from the folder given and the ID of the rule, but a warning, which names
// its place and refuses nothing — a key of task.no-leaks in a text of the whole module is a number
// there too (R50) —, and the last counts them.
func TestCheckRefuses(t *testing.T) {
	rules, err := fs.ReadDir(spec.Format, "vectors/rules")
	if err != nil {
		t.Fatal(err)
	}
	cases := 0
	for _, rule := range rules {
		root := path.Join("vectors/rules", rule.Name())
		entries, err := fs.ReadDir(spec.Format, root)
		if err != nil {
			t.Fatal(err)
		}
		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			cases++
			t.Run(rule.Name()+"/"+e.Name(), func(t *testing.T) {
				dir := fromSpec(t, path.Join(root, e.Name()))
				var stdout, stderr bytes.Buffer
				status := run(context.Background(), []string{"check", dir}, &stdout, &stderr)
				lines := strings.Split(strings.TrimSuffix(stdout.String(), "\n"), "\n")
				if status != statusRefused || stderr.Len() > 0 || len(lines) < 2 {
					t.Fatalf("status %d, stdout %q, stderr %q; want 1 and a report", status, stdout.String(), stderr.String())
				}
				breaches := 0
				for _, line := range lines[:len(lines)-1] {
					warning := strings.Contains(line, ": warning: ")
					if !warning {
						breaches++
					}
					if !strings.HasPrefix(line, filepath.ToSlash(dir)) || !warning && !strings.Contains(line, ": "+rule.Name()+": ") {
						t.Errorf("the line %q names no place in %s, or not the rule %s", line, dir, rule.Name())
					}
				}
				noun := "breaches"
				if breaches == 1 {
					noun = "breach"
				}
				if count := fmt.Sprintf(": %d %s", breaches, noun); breaches == 0 || !strings.HasSuffix(lines[len(lines)-1], count) {
					t.Errorf("the report %q has %d breaches, and its last line does not end with %q", stdout.String(), breaches, count)
				}
				if last := lines[len(lines)-1]; !strings.HasPrefix(last, dir+": the format 0.3 refuses the module: ") {
					t.Errorf("the last line is %q", last)
				}
			})
		}
	}
	if cases == 0 {
		t.Fatal("no test vector of a rule is a folder")
	}
}

// The report counts one breach in the singular, and a breach of the module's folder itself has the
// folder's path.
func TestReport(t *testing.T) {
	dir := fromSpec(t, "examples/module/valid/smallest")
	if err := os.Remove(filepath.Join(dir, "LICENSE")); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	status := run(context.Background(), []string{"check", dir}, &stdout, &stderr)
	want := filepath.ToSlash(dir) + ": module.license: "
	if status != statusRefused || !strings.HasPrefix(stdout.String(), want) || !strings.HasSuffix(stdout.String(), ": 1 breach\n") {
		t.Errorf("status %d, stdout %q; want 1, a line that begins %q and 1 breach", status, stdout.String(), want)
	}
}

// What check cannot read is status 2, said on stderr: a path that does not exist, a file that is no
// folder, and a stop by a context that is done.
func TestCheckFails(t *testing.T) {
	file := filepath.Join(t.TempDir(), "manifest.yaml")
	if err := os.WriteFile(file, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	done, cancel := context.WithCancel(context.Background())
	cancel()
	for _, c := range []struct {
		name string
		ctx  context.Context
		dir  string
		want string
	}{
		{"no such folder", context.Background(), filepath.Join(t.TempDir(), "absent"), "no such file"},
		{"a file", context.Background(), file, "is no folder"},
		{"a context that is done", done, fromSpec(t, "examples/module/valid/with-templates"), "context canceled"},
	} {
		t.Run(c.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			status := run(c.ctx, []string{"check", c.dir}, &stdout, &stderr)
			if status != statusFailed || stdout.Len() > 0 || !strings.Contains(stderr.String(), c.want) {
				t.Errorf("status %d, stdout %q, stderr %q; want 2 and %q on stderr", status, stdout.String(), stderr.String(), c.want)
			}
		})
	}
}
