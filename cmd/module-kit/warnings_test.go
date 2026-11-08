package main

import (
	"bytes"
	"context"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// transparentPicture writes over the picture of a module a PNG of the same name whose every pixel is
// fully transparent, and gives the path that check names it by.
func transparentPicture(t *testing.T, dir string) string {
	t.Helper()
	var b bytes.Buffer
	if err := png.Encode(&b, image.NewNRGBA(image.Rect(0, 0, 2, 2))); err != nil {
		t.Fatal(err)
	}
	file := filepath.Join(dir, "media", "pictures", "dot.png")
	if err := os.WriteFile(file, b.Bytes(), 0o644); err != nil {
		t.Fatal(err)
	}
	return filepath.ToSlash(file)
}

// A picture with a pixel that is not fully opaque gives a warning, which changes no exit status: a
// module that breaks no rule is accepted still, with the warning before its last line, and a module
// that breaks a rule is refused, with the warning after its breaches.
func TestCheckWarns(t *testing.T) {
	accepted := fromSpec(t, "examples/module/valid/with-tasks")
	file := transparentPicture(t, accepted)
	var stdout, stderr bytes.Buffer
	status := run(context.Background(), []string{"check", accepted}, &stdout, &stderr)
	lines := strings.Split(strings.TrimSuffix(stdout.String(), "\n"), "\n")
	want := file + ": warning: the pixel at x = 0, y = 0 has alpha 0 of 65535"
	if status != statusAccepted || stderr.Len() > 0 || len(lines) != 2 || !strings.HasPrefix(lines[0], want) ||
		lines[1] != accepted+": the module breaks no rule of the format 0.3" {
		t.Errorf("status %d, stdout %q, stderr %q; want 0, a warning that begins %q and the last line", status, stdout.String(), stderr.String(), want)
	}

	refused := fromSpec(t, "examples/module/valid/with-tasks")
	file = transparentPicture(t, refused)
	if err := os.Remove(filepath.Join(refused, "LICENSE")); err != nil {
		t.Fatal(err)
	}
	stdout.Reset()
	status = run(context.Background(), []string{"check", refused}, &stdout, &stderr)
	lines = strings.Split(strings.TrimSuffix(stdout.String(), "\n"), "\n")
	if status != statusRefused || len(lines) != 3 || !strings.Contains(lines[0], ": module.license: ") ||
		!strings.HasPrefix(lines[1], file+": warning: ") || !strings.HasSuffix(lines[2], ": 1 breach") {
		t.Errorf("status %d, stdout %q; want 1, the breach, the warning and the last line", status, stdout.String())
	}
}

// A number in a text of the whole module gives a warning too, which changes no exit status: the
// warnings of the texts come first, then those of the pictures, and the last line after them.
func TestCheckWarnsOfNumbers(t *testing.T) {
	dir := fromSpec(t, "examples/module/valid/with-tasks")
	picture := transparentPicture(t, dir)
	manifest := filepath.Join(dir, "manifest.yaml")
	data, err := os.ReadFile(manifest)
	if err != nil {
		t.Fatal(err)
	}
	numbered := strings.Replace(string(data), "Which step of your work", "Which of your 1000 steps", 1)
	if err := os.WriteFile(manifest, []byte(numbered), 0o644); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	status := run(context.Background(), []string{"check", dir}, &stdout, &stderr)
	lines := strings.Split(strings.TrimSuffix(stdout.String(), "\n"), "\n")
	want := filepath.ToSlash(manifest) + ` #/rung_templates/0: warning: the text holds the number "1000": `
	if status != statusAccepted || stderr.Len() > 0 || len(lines) != 3 || !strings.HasPrefix(lines[0], want) ||
		!strings.HasPrefix(lines[1], picture+": warning: ") || lines[2] != dir+": the module breaks no rule of the format 0.3" {
		t.Errorf("status %d, stdout %q, stderr %q; want 0, a warning that begins %q, that of the picture and the last line", status, stdout.String(), stderr.String(), want)
	}
}

// A folder of media that cannot be read stops check with status 2, as any other file would.
func TestCheckWarningsFail(t *testing.T) {
	dir := fromSpec(t, "examples/module/valid/with-tasks")
	media := filepath.Join(dir, "media", "pictures")
	if err := os.Chmod(media, 0o000); err != nil {
		t.Fatal(err)
	}
	defer os.Chmod(media, 0o755)
	if _, err := os.ReadDir(media); err == nil {
		t.Skip("the folder can be read with no permission, as by root")
	}
	var stdout, stderr bytes.Buffer
	if status := run(context.Background(), []string{"check", dir}, &stdout, &stderr); status != statusFailed || !strings.Contains(stderr.String(), "permission denied") {
		t.Errorf("status %d, stderr %q; want 2 and permission denied", status, stderr.String())
	}
}
