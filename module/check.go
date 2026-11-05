package module

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"io/fs"
	"regexp"
	"slices"
	"strings"

	"github.com/altnovatrail/spec/types/formatv0"

	"github.com/altnovatrail/module-kit/internal/schema"
	"github.com/altnovatrail/module-kit/internal/semver"
	"github.com/altnovatrail/module-kit/internal/yaml"
	"github.com/altnovatrail/module-kit/leak"
)

// FormatVersion is the version of the module format that this library reads: the version of the
// schemas that internal/schema holds, a copy of those of spec.Format, and of the test vectors
// (format/README.md, section 2; R23). While the format is v0, it is the one version the library
// reads.
const FormatVersion = "0.3"

// ReadsFormat says whether the library reads a module written in the format version v, as a
// manifest names it in module_format_version: while the format is v0, only FormatVersion
// (format/README.md, section 2).
func ReadsFormat(v string) bool {
	return v == FormatVersion
}

// NewerFormat says whether v, a format version MAJOR.MINOR, is newer than FormatVersion: by its
// major number, then its minor, each by its value, so that 0.10 is newer than 0.9. A v that is no
// version MAJOR.MINOR is newer than none. The build of the module index reads a previous index by
// it (R35).
func NewerFormat(v string) bool {
	// MAJOR.MINOR are two numerals with no leading zero, and so is each number of X.Y.Z: the version
	// MAJOR.MINOR orders as the version of a module MAJOR.MINOR.0.
	newer, err := semver.Parse(v + ".0")
	if err != nil {
		return false
	}
	reader, _ := semver.Parse(FormatVersion + ".0")
	return newer.Compare(reader) > 0
}

// The rules the check holds, by the part a breach of each is reported in (format/README.md,
// sections 10 and 11).
const (
	RuleManifest      = "module.manifest"
	RuleLayout        = "module.layout"
	RuleFileIDs       = "module.file-ids"
	RuleMedia         = "module.media"
	RuleLicense       = "module.license"
	RulePlainText     = "module.plain-text"
	RuleFormatVersion = "manifest.format-version"
)

// A Breach is one reason a module is refused: a rule it breaks; or, when Rule is empty, a keyword of
// the schema of its part that a file fails, which Keyword names; or, when both are empty, a file of a
// part that is not YAML 1.2.2, which the format refuses without an ID (format/README.md, sections 4
// and 8).
type Breach struct {
	Rule string
	// The keyword of JSON Schema 2020-12 that fails, such as required or pattern, when the breach is
	// of the schema; its place is then Pointer, where the schema holding the keyword applies, as an
	// example of the format names it (format/README.md, section 8).
	Keyword string
	// The file or folder from the module's folder, with / between names, as the folder lists them;
	// "." for the module's folder itself, as fs.FS names it, and empty for the file that CheckFile
	// checks, which has no path.
	Path string
	// The place in the file, line and column from 1: in a file of a part as yaml.Breach counts them,
	// and in LICENSE and NOTICE with every line ended by LF alone. 0 when the breach is of the file
	// as a whole or of a value in it.
	Line, Column int
	// The value in the file that breaks the rule, as a JSON Pointer in URI fragment form, such as
	// #/module_format_version (format/README.md, section 8); empty when the breach is of no value.
	Pointer string
	Message string
}

func (b Breach) String() string {
	place := b.Path
	if b.Line > 0 {
		place = strings.TrimPrefix(fmt.Sprintf("%s:%d:%d", b.Path, b.Line, b.Column), ":")
	}
	if b.Pointer != "" {
		place = strings.TrimSpace(place + " " + b.Pointer)
	}
	rule := b.Rule
	switch {
	case rule != "":
	case b.Keyword != "":
		rule = "schema " + b.Keyword
	default:
		rule = "not YAML"
	}
	if place == "" {
		return rule + ": " + b.Message
	}
	return place + ": " + rule + ": " + b.Message
}

// A Report is what the check of a module finds: every breach, in the order of their places — the
// path, then the line and column, then the value, member by member, with an element of an array by
// its index —, and at one place in the order of their rules and keywords, so that it does not
// depend on the order in which the folder lists its files. A report with no breach accepts the
// module.
type Report struct {
	Breaches []Breach
}

func (r Report) String() string {
	lines := make([]string, len(r.Breaches))
	for i, b := range r.Breaches {
		lines[i] = b.String()
	}
	return strings.Join(lines, "\n")
}

// Check reads the module whose folder is the root of fsys and checks it by the format version
// FormatVersion: its layout, each file of a part by the YAML rules and against the schema of its
// part, and the files that both accept by the rules of the parts (format/README.md, sections 4, 5
// and 10–17). A manifest that names a version the library does not read gives a report of that
// breach alone, since the rest of the module is written by rules the library does not know. The
// error is for a folder or file that could not be read, and then there is no report.
//
// The solvers of the module's templates run with ctx, one on each CPU that the process may use at
// once, runtime.GOMAXPROCS(0), and the report holds their breaches in the order of their places, as
// it holds every other (architecture 3.4). Once ctx is done the check stops, and its error is that
// of ctx.
func Check(ctx context.Context, fsys fs.FS) (Report, error) {
	report, _, err := CheckModule(ctx, fsys)
	return report, err
}

// CheckModule is Check, and gives besides the numbers of the instances of each template of a module
// that breaks no rule, by the local ID of its task, in the order of section 23 with the results of
// its solver, as the canonical view holds them (format/README.md, section 25); none when the module
// has no template or is refused. The build of a package takes them, so that a solver runs once for
// the check and the package together.
func CheckModule(ctx context.Context, fsys fs.FS) (Report, map[string][]formatv0.Numbers, error) {
	c := checker{fsys: fsys, parts: map[string]any{}}
	p, err := c.check(ctx)
	if err != nil {
		return Report{}, nil, err
	}
	report := c.report()
	if len(report.Breaches) > 0 || p == nil {
		return report, nil, nil
	}
	return report, p.instancesOf(), nil
}

// CheckFile checks a file of a part on its own, as an example of the part and a test vector of a
// part's rule in a file are checked (format/README.md, sections 8 and 9): by the YAML rules and
// against the schema of the part, and, when both accept it, by the rules within its part; a manifest
// first by the format version it names, which, when the library does not read it, is the report's
// only breach, as in Check. Its breaches have no path.
// part is the term of a part of the layout, such as manifest or task; any other is an error.
func CheckFile(part string, data []byte) (Report, error) {
	if part == "" || !slices.ContainsFunc(layout, func(r row) bool { return r.part == part }) {
		return Report{}, fmt.Errorf("%q is no part of the layout of a module", part)
	}
	c := checker{}
	value, read := c.read("", data)
	if b, other := otherVersion("", part, value); other {
		return Report{Breaches: []Breach{b}}, nil
	}
	if read && c.conform("", part, value) {
		p := newParts(true)
		p.add("", part, value)
		// A file on its own is an example or a test vector, whose solver runs with a context that is
		// never done: it stops by its budget or its clock.
		if err := p.expandAll(context.Background()); err != nil {
			panic(fmt.Sprintf("module: the instances of a template, with a context that is never done: %v", err))
		}
		c.withinPart(p)
	}
	return c.report(), nil
}

// checker holds what the check of one module finds.
type checker struct {
	fsys     fs.FS
	breaches []Breach
	// The files of the module, by their paths from its folder, as the folder lists them (layout.go).
	files []string
	// The value of each file of a part that the YAML rules and the schema of its part accept, by
	// its path. A rule reads only a file that passed, and a rule between parts runs only when every
	// file of a part passed (allParts), as format-check of spec has it: a file that its schema
	// refuses breaks the schema and nothing more (R23).
	parts map[string]any
	// The leak check of each task that a rule of the leak check searches for, by its path (leaks.go).
	searches map[string]*leak.Search
}

// allParts says whether every file of a part passed, the manifest among them: a rule between parts
// runs only then (R23).
func (c *checker) allParts() bool {
	if !slices.Contains(c.files, manifestPath) {
		return false
	}
	for _, file := range c.files {
		if r, _, ok := rowOf(file); ok && r.part != "" {
			if _, passed := c.parts[file]; !passed {
				return false
			}
		}
	}
	return true
}

func (c *checker) add(rule, path string, line, column int, pointer, message string) {
	c.breaches = append(c.breaches, Breach{Rule: rule, Path: path, Line: line, Column: column, Pointer: pointer, Message: message})
}

func (c *checker) report() Report {
	slices.SortStableFunc(c.breaches, func(a, b Breach) int {
		return cmp.Or(strings.Compare(a.Path, b.Path), cmp.Compare(a.Line, b.Line), cmp.Compare(a.Column, b.Column),
			schema.ComparePointers(a.Pointer, b.Pointer), strings.Compare(a.Rule, b.Rule), strings.Compare(a.Keyword, b.Keyword),
			strings.Compare(a.Message, b.Message))
	})
	return Report{Breaches: c.breaches}
}

// check runs the rules of a whole module in the order of section 10, with the format version that
// the manifest names before them, and gives the parts that the rules read: none when the manifest
// stops the check.
func (c *checker) check(ctx context.Context) (*parts, error) {
	if err := c.list(); err != nil {
		return nil, err
	}
	switch stop, err := c.readManifest(); {
	case err != nil:
		return nil, err
	case stop:
		return nil, nil
	}
	if err := c.place(); err != nil {
		return nil, err
	}
	if err := c.media(); err != nil {
		return nil, err
	}
	if err := c.texts(); err != nil {
		return nil, err
	}
	return c.rules(ctx)
}

// rules checks the rules of the parts on the files of a part that passed, read into their types
// once, with the templates among them expanded with ctx: those within a part on each, and then, when
// every file of a part passed, those between parts (rules.go). It gives the parts it read.
func (c *checker) rules(ctx context.Context) (*parts, error) {
	p := newParts(false)
	for path, value := range c.parts {
		r, _, _ := rowOf(path)
		p.add(path, r.part, value)
	}
	for _, file := range c.files {
		if strings.HasPrefix(file, mediaFolder) {
			p.media[file] = true
		}
	}
	if err := p.expandAll(ctx); err != nil {
		return nil, err
	}
	c.withinPart(p)
	if c.allParts() {
		c.betweenParts(p)
	}
	return &p, nil
}

// read reads a file of a part by the YAML rules, and gives its value, or nil and false when they
// refuse it.
func (c *checker) read(path string, data []byte) (any, bool) {
	value, err := yaml.Read(data)
	var refusal *yaml.Refusal
	if errors.As(err, &refusal) {
		for _, b := range refusal.Breaches {
			c.add(b.Rule, path, b.Line, b.Column, "", b.Message)
		}
		return nil, false
	}
	return value, true
}

// conform checks the value of a file of a part against the schema of its part, schemas/<part>.schema.json
// of the version the library reads, and says whether the schema accepts it. Each keyword that fails
// is a breach at its place in the value (format/README.md, section 8).
func (c *checker) conform(path, part string, value any) bool {
	errs := schema.Validate(part+".schema.json", value)
	for _, e := range errs {
		c.breaches = append(c.breaches, Breach{Keyword: e.Keyword, Path: path, Pointer: e.Pointer(), Message: e.Message})
	}
	return len(errs) == 0
}

// part reads a file of a part at path by the YAML rules and checks it against the schema of its
// part. A file that both accept is one that the rules may read: its value stands in parts.
func (c *checker) part(path, part string, data []byte) {
	value, read := c.read(path, data)
	if read && c.conform(path, part, value) {
		c.parts[path] = value
	}
}

// The field at the root of a manifest that names the format version, at the same place in every
// version of the format, and the form of a version, MAJOR.MINOR, as module_format_version in
// common.schema.json has it (format/README.md, section 2).
const versionField = "module_format_version"

var versionForm = regexp.MustCompile(`^(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$`)

// otherVersion gives the breach of manifest.format-version of a file of a part, and true, when the
// file is a manifest that names a version the library does not read. A manifest that names no
// version — that is no object, has no such field, or has one that is no string of the form
// MAJOR.MINOR — breaks the schema of a manifest, which the library reads by its own version, rather
// than this rule; so the example format-version-as-number.yaml has it.
func otherVersion(path, part string, value any) (Breach, bool) {
	manifest, _ := value.(map[string]any)
	v, ok := manifest[versionField].(string)
	if part != partManifest || !ok || !versionForm.MatchString(v) || ReadsFormat(v) {
		return Breach{}, false
	}
	return Breach{Rule: RuleFormatVersion, Path: path, Pointer: "#/" + versionField,
		Message: fmt.Sprintf("names the format version %s, and module-kit reads %s alone", v, FormatVersion)}, true
}
