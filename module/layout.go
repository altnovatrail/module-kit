package module

import (
	"fmt"
	"io/fs"
	"regexp"
	"slices"
	"strings"
	"unicode/utf8"
)

// The layout of a module and the rules of its folder (format/README.md, section 10): which files a
// module holds, where each stands and which part's schema checks it.

// A row of the layout: a path as section 10 writes it, from the module's folder, and the term of the
// part whose schema checks a file at it, or "" for a file that no schema checks. <path> at the end
// of a path stands for one name or more, joined by /; <id> before the extension of the last name
// stands for one name of the form of a local ID.
type row struct {
	path string
	part string
}

// The parts of the layout and the files of the module's folder that the rules read by name.
const (
	partManifest = "manifest"
	manifestPath = "manifest.yaml"
	licensePath  = "LICENSE"
	noticePath   = "NOTICE"
	mediaFolder  = "media/"
	// The path of the module's folder itself, as fs.FS names its root.
	moduleFolder = "."
)

// layout is the table of section 10, row by row; a test holds it to the table. The eval set, eval/,
// stands in the layout, though the package leaves it out (section 25).
var layout = []row{
	{manifestPath, partManifest},
	{"skill-graph.yaml", "skill-graph"},
	{"traps.yaml", "traps"},
	{"techniques.yaml", "techniques"},
	{"solving-steps.yaml", "solving-steps"},
	{"tasks/<id>.yaml", "task"},
	{"lessons/<id>.yaml", "lesson"},
	{mediaFolder + "<path>", ""},
	{"eval/<path>", ""},
	{licensePath, ""},
	{noticePath, ""},
}

const (
	anyPath = "<path>"
	anyID   = "<id>"
)

// holds says whether a file stands at the row, and, when the row's path has <id>, gives the name the
// file has there before its extension. Names compare exactly, letter case and all.
func (r row) holds(file string) (id string, ok bool) {
	if folder, below := strings.CutSuffix(r.path, anyPath); below {
		return "", len(file) > len(folder) && strings.HasPrefix(file, folder)
	}
	folder, extension, hasID := strings.Cut(r.path, anyID)
	if !hasID {
		return "", file == r.path
	}
	name, ok := strings.CutPrefix(file, folder)
	if !ok {
		return "", false
	}
	id, ok = strings.CutSuffix(name, extension)
	return id, ok && localID.matches(id)
}

// rowOf gives the row of the layout that a file stands at, with the name at <id> when its path has
// one, and false when the file stands at none.
func rowOf(file string) (row, string, bool) {
	for _, r := range layout {
		if id, ok := r.holds(file); ok {
			return r, id, true
		}
	}
	return row{}, "", false
}

// A form of a name from common.schema.json, as the layout reads names by it: a pattern, a length in
// characters at most, and a pattern that the name MUST NOT match.
type form struct {
	pattern   *regexp.Regexp
	maxLength int
	not       *regexp.Regexp
}

func (f form) matches(name string) bool {
	return utf8.RuneCountInString(name) <= f.maxLength && f.pattern.MatchString(name) && !f.not.MatchString(name)
}

// The forms $defs/local_id and $defs/media_path of common.schema.json, by which <id> and module.media
// read names (format/README.md, section 10); a test holds them to the schema.
var (
	localID = form{
		pattern:   regexp.MustCompile(`^[a-z][a-z0-9]*(-[a-z0-9]+)*$`),
		maxLength: 64,
		not:       regexp.MustCompile(`^(con|prn|aux|nul|com[1-9]|lpt[1-9])$`),
	}
	mediaPath = form{
		pattern:   regexp.MustCompile(`^([a-z][a-z0-9]*(-[a-z0-9]+)*/)*[a-z][a-z0-9]*(-[a-z0-9]+)*\.(png|jpg|webp|mp3)$`),
		maxLength: 128,
		not:       regexp.MustCompile(`(^|/)(con|prn|aux|nul|com[1-9]|lpt[1-9])(/|\.)`),
	}
)

// list lists the files of the module, by their paths from its folder, as the folder lists their
// names. A file or folder whose name begins with "." is not the module's, nor anything in such a
// folder, and a folder with no file in it holds nothing. A symbolic link, which list does not
// follow, and anything else that is neither a file nor a folder break module.layout.
func (c *checker) list() error {
	return c.listFolder(".")
}

func (c *checker) listFolder(folder string) error {
	entries, err := fs.ReadDir(c.fsys, folder)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".") {
			continue
		}
		name := entry.Name()
		if folder != "." {
			name = folder + "/" + name
		}
		switch kind := entry.Type(); {
		case kind&fs.ModeSymlink != 0:
			c.add(RuleLayout, name, 0, 0, "", "a symbolic link: a module holds files and folders only")
		case kind.IsDir():
			if err := c.listFolder(name); err != nil {
				return err
			}
		case kind.IsRegular():
			c.files = append(c.files, name)
		default:
			c.add(RuleLayout, name, 0, 0, "", "neither a file nor a folder: a module holds files and folders only")
		}
	}
	return nil
}

// readManifest checks module.manifest, and reads the manifest as the first file of a part: the
// format version it names, then the schema of a manifest. stop is true when the manifest names a
// format version that the library does not read: its breach is then the report's only one.
func (c *checker) readManifest() (stop bool, err error) {
	if !slices.Contains(c.files, manifestPath) {
		c.add(RuleManifest, moduleFolder, 0, 0, "", "the folder holds no file "+manifestPath)
		return false, nil
	}
	data, err := fs.ReadFile(c.fsys, manifestPath)
	if err != nil {
		return false, err
	}
	value, read := c.read(manifestPath, data)
	if b, other := otherVersion(manifestPath, partManifest, value); other {
		c.breaches = []Breach{b}
		return true, nil
	}
	if read && c.conform(manifestPath, partManifest, value) {
		c.parts[manifestPath] = value
	}
	return false, nil
}

// place checks that each file stands at a path of the layout (module.layout), reads each file of a
// part but the manifest by the YAML rules and against the schema of its part, and checks the id
// that a file at <id> holds (module.file-ids) when both accept the file.
func (c *checker) place() error {
	for _, file := range c.files {
		r, id, ok := rowOf(file)
		if !ok {
			c.add(RuleLayout, file, 0, 0, "", "stands at no path of the layout of a module")
			continue
		}
		if r.part == "" || file == manifestPath {
			continue
		}
		data, err := fs.ReadFile(c.fsys, file)
		if err != nil {
			return err
		}
		c.part(file, r.part, data)
		if value, passed := c.parts[file]; passed && id != "" {
			c.fileID(file, id, value)
		}
	}
	return nil
}

// fileID checks module.file-ids on the value of a file at <id> that its schema accepts, so that the
// id is a string of the form of a local ID: the rule says only what the schema cannot, that the id
// is the file's name (format/README.md, section 7). A file that the schema refuses breaks it and
// nothing more (R23).
func (c *checker) fileID(file, name string, value any) {
	entry, _ := value.(map[string]any)
	if id, ok := entry["id"].(string); ok && id != name {
		c.add(RuleFileIDs, file, 0, 0, "#/id", fmt.Sprintf("holds the id %q, and the file is named by the id %s", id, name))
	}
}
