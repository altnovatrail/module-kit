package module

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/altnovatrail/module-kit/internal/yaml"
)

// A manifest that breaks no rule, of the version the library reads.
const manifest = `module_format_version: "0.3"
id: example/test
version: "0.1.0"
title: A test
subject: mathematics
audience:
  age_categories: [adult]
languages:
  explanation: en
authors:
  - name: The NovaTrail authors
license: Apache-2.0
`

// The first bytes of a file of each media format, and the header of a frame of MP3 without ID3.
var (
	png       = []byte("\x89PNG\r\n\x1a\n\x00\x00\x00\x0dIHDR")
	jpeg      = []byte{0xff, 0xd8, 0xff, 0xe0, 0x00, 0x10, 'J', 'F', 'I', 'F'}
	webp      = []byte("RIFF\x24\x00\x00\x00WEBPVP8 ")
	mp3       = []byte("ID3\x04\x00\x00\x00\x00\x00\x00")
	mp3Header = []byte{0xff, 0xfb, 0x90, 0x64}
)

func file(data string) *fstest.MapFile { return &fstest.MapFile{Data: []byte(data)} }

func bytesFile(data []byte) *fstest.MapFile { return &fstest.MapFile{Data: data} }

// module gives a module that breaks no rule, with the files of more, and without each file that
// more maps to nil.
func module(more fstest.MapFS) fstest.MapFS {
	fsys := fstest.MapFS{manifestPath: file(manifest), licensePath: file("The text of the license.\n")}
	for name, f := range more {
		if f == nil {
			delete(fsys, name)
		} else {
			fsys[name] = f
		}
	}
	return fsys
}

// short gives a breach as a case of TestCheck names it: its place, as Breach.String writes it, and
// its rule, or "not YAML".
func short(b Breach) string {
	s := b.String()
	return strings.Replace(s[:len(s)-len(": "+b.Message)], ": ", " ", 1)
}

func shortReport(r Report) []string {
	var breaches []string
	for _, b := range r.Breaches {
		breaches = append(breaches, short(b))
	}
	return breaches
}

// The cases hold each rule of the check, each way a module's folder is listed, and the format
// version.
var checkCases = []struct {
	name    string
	fsys    fstest.MapFS
	refused []string
}{
	{name: "every path of the layout", fsys: module(fstest.MapFS{
		noticePath:                     file("Notices.\n"),
		"skill-graph.yaml":             file("skills: {}\n"),
		"traps.yaml":                   file("traps: {}\n"),
		"techniques.yaml":              file("techniques: {}\n"),
		"solving-steps.yaml":           file("steps: []\n"),
		"tasks/seven-times-8.yaml":     file("id: seven-times-8\n"),
		"lessons/a.yaml":               file("id: a\n"),
		"media/dot.png":                bytesFile(png),
		"media/pictures/a-b.jpg":       bytesFile(jpeg),
		"media/a/b/c/d.webp":           bytesFile(webp),
		"media/sounds/bell.mp3":        bytesFile(mp3),
		"media/sounds/frame.mp3":       bytesFile(mp3Header),
		"eval/samples/One Sample.json": file("{}"),
		"eval/README":                  file(""),
	})},
	{name: "what is not the module's", fsys: module(fstest.MapFS{
		".git/config":           file("x"),
		".DS_Store":             file("x"),
		".link":                 {Data: []byte("LICENSE"), Mode: fs.ModeSymlink},
		"tasks/.draft.yaml":     file("id: [\n"),
		"media/.thumbs/x.gif":   file("x"),
		"extras/.keep":          file(""),
		"empty":                 {Mode: fs.ModeDir},
		"media/empty/folder":    {Mode: fs.ModeDir},
		".hidden/Manifest.yaml": file("x"),
	})},
	{name: "an empty NOTICE", fsys: module(fstest.MapFS{noticePath: file("")})},
	{name: "an id that is no string, or none, is the schema's", fsys: module(fstest.MapFS{
		"tasks/a.yaml": file("id: 5\n"), "lessons/b.yaml": file("title: B\n"), "tasks/c.yaml": file("- c\n"),
	})},

	// module.manifest and module.layout.
	{name: "no manifest", fsys: module(fstest.MapFS{manifestPath: nil}), refused: []string{". module.manifest"}},
	{name: "Manifest.yaml in place of manifest.yaml", fsys: module(fstest.MapFS{manifestPath: nil, "Manifest.yaml": file(manifest)}),
		refused: []string{". module.manifest", "Manifest.yaml module.layout"}},
	{name: "a folder manifest.yaml", fsys: module(fstest.MapFS{manifestPath: nil, "manifest.yaml/manifest.yaml": file(manifest)}),
		refused: []string{". module.manifest", "manifest.yaml/manifest.yaml module.layout"}},
	{name: "files at no path of the layout", fsys: module(fstest.MapFS{
		"notes.txt":            file("x"),
		"extras/notes.txt":     file("x"),
		"notice":               file("x"),
		"NOTICE.txt":           file("x"),
		"License":              file("x"),
		"skill-graph.yml":      file("skills: {}\n"),
		"media":                file("x"),
		"eval":                 file("x"),
		"dot.png":              bytesFile(png),
		"tasks/Seven.yaml":     file("id: Seven\n"),
		"tasks/7-times-8.yaml": file("id: 7-times-8\n"),
		"tasks/con.yaml":       file("id: con\n"),
		"tasks/a.yml":          file("id: a\n"),
		"tasks/a.yaml.yaml":    file("id: a\n"),
		"tasks/sub/a.yaml":     file("id: a\n"),
		"tasks/" + strings.Repeat("a", 65) + ".yaml": file("x"),
		"lessons/a--b.yaml":                          file("id: a--b\n"),
		"Tasks/a.yaml":                               file("id: a\n"),
	}), refused: []string{
		"License module.layout", "NOTICE.txt module.layout", "Tasks/a.yaml module.layout", "dot.png module.layout",
		"eval module.layout", "extras/notes.txt module.layout", "lessons/a--b.yaml module.layout", "media module.layout",
		"notes.txt module.layout", "notice module.layout", "skill-graph.yml module.layout", "tasks/7-times-8.yaml module.layout",
		"tasks/Seven.yaml module.layout", "tasks/a.yaml.yaml module.layout", "tasks/a.yml module.layout",
		"tasks/" + strings.Repeat("a", 65) + ".yaml module.layout", "tasks/con.yaml module.layout", "tasks/sub/a.yaml module.layout",
	}},
	{name: "a local ID of 64 characters", fsys: module(fstest.MapFS{
		"tasks/" + strings.Repeat("a", 64) + ".yaml": file("id: " + strings.Repeat("a", 64) + "\n"),
	})},
	{name: "a symbolic link and a named pipe", fsys: module(fstest.MapFS{
		"link":       {Data: []byte("LICENSE"), Mode: fs.ModeSymlink},
		"media/link": {Data: []byte("../LICENSE"), Mode: fs.ModeSymlink},
		"eval/pipe":  {Mode: fs.ModeNamedPipe},
	}), refused: []string{"eval/pipe module.layout", "link module.layout", "media/link module.layout"}},

	// module.file-ids.
	{name: "an id unlike the name", fsys: module(fstest.MapFS{
		"tasks/seven-times-eight.yaml": file("id: eight-times-seven\n"),
		"lessons/first.yaml":           file("id: First\n"),
	}), refused: []string{"lessons/first.yaml #/id module.file-ids", "tasks/seven-times-eight.yaml #/id module.file-ids"}},

	// The YAML rules on each file of a part, and none on the files that no schema checks.
	{name: "files of parts that the YAML rules refuse", fsys: module(fstest.MapFS{
		"skill-graph.yaml": file("skills: {}\nskills: {}\n"),
		"traps.yaml":       file("traps: &a {}\n"),
		"tasks/a.yaml":     file("id: [a\n"),
		"lessons/b.yaml":   file("id: b\n---\nid: b\n"),
		"eval/x.yaml":      file("a: &a b\n"),
	}), refused: []string{
		"lessons/b.yaml:2:1 yaml.single-document", "skill-graph.yaml:2:1 yaml.unique-keys",
		"tasks/a.yaml:2:1 not YAML", "traps.yaml:1:8 yaml.no-anchors",
	}},
	{name: "a manifest that the YAML rules refuse, and the rest checked", fsys: module(fstest.MapFS{
		manifestPath: file(manifest + "title: Another\n"), licensePath: nil, "notes.txt": file("x"),
	}), refused: []string{". module.license", "manifest.yaml:13:1 yaml.unique-keys", "notes.txt module.layout"}},
	{name: "a manifest of 0.4 that the YAML rules refuse names no version the check reads", fsys: module(fstest.MapFS{
		manifestPath: file(strings.Replace(manifest, `"0.3"`, `"0.4"`, 1) + "subject: no\n"), "notes.txt": file("x"),
	}), refused: []string{"manifest.yaml:13:1 yaml.unique-keys", "manifest.yaml:13:10 yaml.plain-strings", "notes.txt module.layout"}},

	// The format version.
	{name: "a module of 0.2, refused for it alone", fsys: module(fstest.MapFS{
		manifestPath: file(strings.Replace(manifest, `"0.3"`, `"0.2"`, 1)), licensePath: nil,
		"notes.txt": file("x"), "link": {Data: []byte("x"), Mode: fs.ModeSymlink}, "tasks/a.yaml": file("id: b\n"),
	}), refused: []string{"manifest.yaml #/module_format_version manifest.format-version"}},
	{name: "a module of 0.4", fsys: module(fstest.MapFS{manifestPath: file(strings.Replace(manifest, `"0.3"`, `"0.4"`, 1))}),
		refused: []string{"manifest.yaml #/module_format_version manifest.format-version"}},
	{name: "a module of 1.0", fsys: module(fstest.MapFS{manifestPath: file(strings.Replace(manifest, `"0.3"`, `"1.0"`, 1))}),
		refused: []string{"manifest.yaml #/module_format_version manifest.format-version"}},
	{name: "a version that is no MAJOR.MINOR is the schema's", fsys: module(fstest.MapFS{
		manifestPath: file(strings.Replace(manifest, `"0.3"`, `"0.03"`, 1)),
	})},
	{name: "a version as a number is the schema's", fsys: module(fstest.MapFS{
		manifestPath: file(strings.Replace(manifest, `"0.3"`, `0.2`, 1)),
	})},
	{name: "a manifest with no version is the schema's", fsys: module(fstest.MapFS{manifestPath: file("id: example/test\n")})},
	{name: "a manifest that is no object is the schema's", fsys: module(fstest.MapFS{manifestPath: file("- \"0.2\"\n")})},

	// module.media.
	{name: "media at no media path", fsys: module(fstest.MapFS{
		"media/Pictures/dot.png": bytesFile(png),
		"media/dot.PNG":          bytesFile(png),
		"media/dot.gif":          file("GIF89a"),
		"media/dot.jpeg":         bytesFile(jpeg),
		"media/dot":              bytesFile(png),
		"media/con.png":          bytesFile(png),
		"media/com1/dot.png":     bytesFile(png),
		"media/a_b.png":          bytesFile(png),
		"media/" + strings.Repeat("a", 125) + ".png":         bytesFile(png),
		"media/" + strings.Repeat("a", 124) + ".png":         bytesFile(png),
		"media/console.png":                                  bytesFile(png),
		"media/pictures/" + strings.Repeat("a", 60) + ".mp3": bytesFile(mp3),
	}), refused: []string{
		"media/Pictures/dot.png module.media", "media/a_b.png module.media",
		"media/" + strings.Repeat("a", 125) + ".png module.media", "media/com1/dot.png module.media", "media/con.png module.media",
		"media/dot module.media", "media/dot.PNG module.media", "media/dot.gif module.media", "media/dot.jpeg module.media",
	}},
	{name: "media whose bytes are of another format", fsys: module(fstest.MapFS{
		"media/png-holding-jpeg.png": bytesFile(jpeg),
		"media/jpeg-holding-png.jpg": bytesFile(png),
		"media/wave.webp":            file("RIFF\x24\x00\x00\x00WAVEfmt "),
		"media/short.webp":           file("RIFF\x24\x00\x00\x00WEBPV"),
		"media/text.mp3":             file("A bell rings.\n"),
		"media/empty.png":            file(""),
		"media/short-frame.mp3":      bytesFile(mp3Header[:3]),
		"media/layer-two.mp3":        bytesFile([]byte{0xff, 0xfd, 0x90, 0x64}),
		"media/no-sync.mp3":          bytesFile([]byte{0xff, 0xdb, 0x90, 0x64}),
		"media/bad-bitrate.mp3":      bytesFile([]byte{0xff, 0xfb, 0xf0, 0x64}),
		"media/bad-rate.mp3":         bytesFile([]byte{0xff, 0xfb, 0x9c, 0x64}),
		"media/other-frame.mp3":      bytesFile([]byte{0xff, 0xe3, 0x00, 0x00}),
	}), refused: []string{
		"media/bad-bitrate.mp3 module.media", "media/bad-rate.mp3 module.media", "media/empty.png module.media",
		"media/jpeg-holding-png.jpg module.media", "media/layer-two.mp3 module.media", "media/no-sync.mp3 module.media",
		"media/png-holding-jpeg.png module.media", "media/short-frame.mp3 module.media", "media/short.webp module.media",
		"media/text.mp3 module.media", "media/wave.webp module.media",
	}},

	// module.license and module.plain-text.
	{name: "no LICENSE", fsys: module(fstest.MapFS{licensePath: nil}), refused: []string{". module.license"}},
	{name: "an empty LICENSE", fsys: module(fstest.MapFS{licensePath: file("")}), refused: []string{"LICENSE module.license"}},
	{name: "a folder LICENSE", fsys: module(fstest.MapFS{licensePath: nil, "LICENSE/text": file("x")}),
		refused: []string{". module.license", "LICENSE/text module.layout"}},
	{name: "LICENSE and NOTICE that are no plain text", fsys: module(fstest.MapFS{
		licensePath: file("Line one\r\nLine two\r\n"),
		noticePath:  file("\ufeffA notice\nAn ab\xffc\rline\n\xfe"),
	}), refused: []string{"LICENSE:1:9 module.plain-text", "NOTICE:1:1 module.plain-text", "NOTICE:2:6 module.plain-text", "NOTICE:2:8 module.plain-text"}},
}

func TestCheck(t *testing.T) {
	for _, c := range checkCases {
		t.Run(c.name, func(t *testing.T) {
			report, err := Check(c.fsys)
			if err != nil {
				t.Fatal(err)
			}
			if got := shortReport(report); !slices.Equal(got, c.refused) {
				t.Errorf("refused by\n%s\nwant\n%s\nreport:\n%s", strings.Join(got, "\n"), strings.Join(c.refused, "\n"), report)
			}
		})
	}
}

// Every breach of a check of a module says what breaks the rule.
func TestMessages(t *testing.T) {
	report, err := Check(module(fstest.MapFS{
		manifestPath:     nil,
		licensePath:      file("a\rb\rc"),
		"tasks/a.yaml":   file("id: b\n"),
		"media/a.png":    bytesFile(jpeg),
		"media/A.png":    bytesFile(png),
		"notes.txt":      file("x"),
		"link":           {Mode: fs.ModeSymlink},
		"media/pipe.mp3": {Mode: fs.ModeNamedPipe},
	}))
	if err != nil {
		t.Fatal(err)
	}
	want := []string{
		".: module.manifest: the folder holds no file manifest.yaml",
		"LICENSE:1:2: module.plain-text: holds 2 carriage returns, the first here: every line ends with LF alone",
		"link: module.layout: a symbolic link: a module holds files and folders only",
		"media/A.png: module.media: A.png is no media_path of common.schema.json: names in kebab-case that begin with a letter, joined by /, and the extension .png, .jpg, .webp or .mp3",
		"media/a.png: module.media: does not begin with the signature of PNG, as its extension .png says it does",
		"media/pipe.mp3: module.layout: neither a file nor a folder: a module holds files and folders only",
		"notes.txt: module.layout: stands at no path of the layout of a module",
		`tasks/a.yaml #/id: module.file-ids: holds the id "b", and the file is named by the id a`,
	}
	if got := strings.Split(report.String(), "\n"); !slices.Equal(got, want) {
		t.Errorf("the report reads\n%s\nwant\n%s", report, strings.Join(want, "\n"))
	}
}

// The report of a module of another version names it, and a file that is not YAML says so.
func TestBreachString(t *testing.T) {
	for _, c := range []struct {
		breach Breach
		want   string
	}{
		{Breach{Rule: RuleFormatVersion, Path: manifestPath, Pointer: "#/module_format_version", Message: "m"},
			"manifest.yaml #/module_format_version: manifest.format-version: m"},
		{Breach{Path: "tasks/a.yaml", Line: 2, Column: 1, Message: "m"}, "tasks/a.yaml:2:1: not YAML: m"},
		{Breach{Rule: RuleManifest, Path: ".", Message: "m"}, ".: module.manifest: m"},
		{Breach{Rule: RuleFormatVersion, Pointer: "#/module_format_version", Message: "m"}, "#/module_format_version: manifest.format-version: m"},
		{Breach{Rule: yaml.RuleNoTags, Line: 1, Column: 4, Message: "m"}, "1:4: yaml.no-tags: m"},
		{Breach{Rule: yaml.RuleNoTags, Message: "m"}, "yaml.no-tags: m"},
	} {
		if got := c.breach.String(); got != c.want {
			t.Errorf("%q, want %q", got, c.want)
		}
	}
	report, err := Check(module(fstest.MapFS{manifestPath: file(strings.Replace(manifest, `"0.3"`, `"0.2"`, 1))}))
	if err != nil {
		t.Fatal(err)
	}
	if want := "manifest.yaml #/module_format_version: manifest.format-version: names the format version 0.2, and module-kit reads 0.3 alone"; report.String() != want {
		t.Errorf("%q, want %q", report, want)
	}
}

func TestCheckFile(t *testing.T) {
	for _, c := range []struct {
		part, data string
		refused    []string
	}{
		{part: "manifest", data: manifest},
		{part: "manifest", data: strings.Replace(manifest, `"0.3"`, `"0.2"`, 1), refused: []string{"#/module_format_version manifest.format-version"}},
		{part: "manifest", data: "module_format_version: 0.2\n"},
		{part: "manifest", data: "module_format_version: \"0.2\"\nmodule_format_version: \"0.2\"\n", refused: []string{"2:1 yaml.unique-keys"}},
		{part: "task", data: "id: a\n"},
		{part: "task", data: "module_format_version: \"0.2\"\n"},
		{part: "lesson", data: "a: !x b\n", refused: []string{"1:4 yaml.no-tags"}},
	} {
		report, err := CheckFile(c.part, []byte(c.data))
		if err != nil {
			t.Fatal(err)
		}
		if got := shortReport(report); !slices.Equal(got, c.refused) {
			t.Errorf("%s %q: refused by %q, want %q", c.part, c.data, got, c.refused)
		}
	}
	for _, part := range []string{"module", "inventory", "media", ""} {
		if _, err := CheckFile(part, []byte(manifest)); err == nil {
			t.Errorf("CheckFile takes %q for a part", part)
		}
	}
}

func TestReadsFormat(t *testing.T) {
	for v, want := range map[string]bool{"0.3": true, "0.2": false, "0.4": false, "1.0": false, "0.3.0": false, "": false} {
		if ReadsFormat(v) != want {
			t.Errorf("ReadsFormat(%q) is %t", v, !want)
		}
	}
}

// reversed lists each folder in the reverse order of its names.
type reversed struct{ fstest.MapFS }

func (r reversed) ReadDir(name string) ([]fs.DirEntry, error) {
	entries, err := r.MapFS.ReadDir(name)
	slices.Reverse(entries)
	return entries, err
}

// The report is the same however the folder lists its files.
func TestOrderOfListing(t *testing.T) {
	for _, c := range checkCases {
		inOrder, err := Check(c.fsys)
		if err != nil {
			t.Fatal(err)
		}
		reverse, err := Check(reversed{c.fsys})
		if err != nil {
			t.Fatal(err)
		}
		if !slices.Equal(inOrder.Breaches, reverse.Breaches) {
			t.Errorf("%s: listed in reverse, the report reads\n%s\nand in order\n%s", c.name, reverse, inOrder)
		}
	}
}

// failing fails to open one name. It has no other method, so that every read opens.
type failing struct {
	fsys fstest.MapFS
	name string
}

var errFailing = errors.New("the test fails this name")

func (f failing) Open(name string) (fs.File, error) {
	if name == f.name {
		return nil, errFailing
	}
	return f.fsys.Open(name)
}

// A folder or file that cannot be read is an error, and no report.
func TestReadErrors(t *testing.T) {
	fsys := module(fstest.MapFS{
		noticePath: file("x"), "tasks/a.yaml": file("id: a\n"), "media/a.png": bytesFile(png),
	})
	for _, name := range []string{".", "tasks", manifestPath, "tasks/a.yaml", "media/a.png", licensePath, noticePath} {
		if _, err := Check(failing{fsys, name}); !errors.Is(err, errFailing) {
			t.Errorf("%s fails, and Check gives %v", name, err)
		}
	}
}

// A symbolic link on a file system, which the check does not follow.
func TestSymbolicLinkOnDisk(t *testing.T) {
	dir := t.TempDir()
	for name, data := range map[string]string{manifestPath: manifest, licensePath: "x", "eval/a.json": "{}"} {
		if err := os.MkdirAll(filepath.Dir(filepath.Join(dir, name)), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, name), []byte(data), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Symlink(licensePath, filepath.Join(dir, noticePath)); err != nil {
		t.Skip("no symbolic link on this file system:", err)
	}
	if err := os.Symlink("eval", filepath.Join(dir, "media")); err != nil {
		t.Fatal(err)
	}
	report, err := Check(os.DirFS(dir))
	if err != nil {
		t.Fatal(err)
	}
	if got, want := shortReport(report), []string{"NOTICE module.layout", "media module.layout"}; !slices.Equal(got, want) {
		t.Errorf("refused by %q, want %q", got, want)
	}
}

func ExampleCheck() {
	report, err := Check(fstest.MapFS{
		"Manifest.yaml": {Data: []byte(manifest)},
		"LICENSE":       {Data: []byte("The text of the license.\r\n")},
	})
	if err != nil {
		panic(err)
	}
	fmt.Println(report)
	// Output:
	// .: module.manifest: the folder holds no file manifest.yaml
	// LICENSE:1:25: module.plain-text: holds a carriage return: every line ends with LF alone
	// Manifest.yaml: module.layout: stands at no path of the layout of a module
}
