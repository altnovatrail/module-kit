package module

import (
	"cmp"
	"errors"
	"fmt"
	"io/fs"
	"regexp"
	"slices"
	"strings"

	"github.com/altnovatrail/module-kit/internal/yaml"
)

// FormatVersion is the version of the module format that this library reads: the version of the
// schemas and test vectors it takes from spec.Format (format/README.md, section 2). While the
// format is v0, it is the one version the library reads.
const FormatVersion = "0.3"

// ReadsFormat says whether the library reads a module written in the format version v, as a
// manifest names it in module_format_version: while the format is v0, only FormatVersion
// (format/README.md, section 2).
func ReadsFormat(v string) bool {
	return v == FormatVersion
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

// A Breach is one reason a module is refused: a rule it breaks, or, when Rule is empty, a file of a
// part that is not YAML 1.2.2, which the format refuses without an ID (format/README.md, section 4).
type Breach struct {
	Rule string
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
	rule := cmp.Or(b.Rule, "not YAML")
	if place == "" {
		return rule + ": " + b.Message
	}
	return place + ": " + rule + ": " + b.Message
}

// A Report is what the check of a module finds: every breach, in the order of their places — the
// path, then the line and column, then the value —, and at one place in the order of their rules,
// so that it does not depend on the order in which the folder lists its files. A report with no
// breach accepts the module.
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
// FormatVersion: its layout, and each file of a part by the YAML rules (format/README.md, sections
// 4, 10 and 11). A manifest that names a version the library does not read gives a report of that
// breach alone, since the rest of the module is written by rules the library does not know. The
// error is for a folder or file that could not be read, and then there is no report.
func Check(fsys fs.FS) (Report, error) {
	c := checker{fsys: fsys}
	if err := c.check(); err != nil {
		return Report{}, err
	}
	return c.report(), nil
}

// CheckFile checks a file of a part on its own, as an example of the part and a test vector of a
// part's rule in a file are checked (format/README.md, sections 8 and 9): by the YAML rules, and a
// manifest by the format version it names. Its breaches have no path. part is the term of a part of
// the layout, such as manifest or task; any other is an error.
func CheckFile(part string, data []byte) (Report, error) {
	if part == "" || !slices.ContainsFunc(layout, func(r row) bool { return r.part == part }) {
		return Report{}, fmt.Errorf("%q is no part of the layout of a module", part)
	}
	c := checker{}
	value := c.part("", part, data)
	if b, other := otherVersion("", part, value); other {
		c.breaches = append(c.breaches, b)
	}
	return c.report(), nil
}

// checker holds what the check of one module finds.
type checker struct {
	fsys     fs.FS
	breaches []Breach
	// The files of the module, by their paths from its folder, as the folder lists them (layout.go).
	files []string
}

func (c *checker) add(rule, path string, line, column int, pointer, message string) {
	c.breaches = append(c.breaches, Breach{rule, path, line, column, pointer, message})
}

func (c *checker) report() Report {
	slices.SortStableFunc(c.breaches, func(a, b Breach) int {
		return cmp.Or(strings.Compare(a.Path, b.Path), cmp.Compare(a.Line, b.Line), cmp.Compare(a.Column, b.Column),
			strings.Compare(a.Pointer, b.Pointer), strings.Compare(a.Rule, b.Rule), strings.Compare(a.Message, b.Message))
	})
	return Report{Breaches: c.breaches}
}

// check runs the rules of a whole module in the order of section 10, with the format version that
// the manifest names before them.
func (c *checker) check() error {
	if err := c.list(); err != nil {
		return err
	}
	switch stop, err := c.readManifest(); {
	case err != nil:
		return err
	case stop:
		return nil
	}
	if err := c.place(); err != nil {
		return err
	}
	if err := c.media(); err != nil {
		return err
	}
	return c.texts()
}

// part reads a file of a part by the YAML rules, and gives its value, or nil when they refuse it.
func (c *checker) part(path, part string, data []byte) any {
	value, err := yaml.Read(data)
	var refusal *yaml.Refusal
	if errors.As(err, &refusal) {
		for _, b := range refusal.Breaches {
			c.add(b.Rule, path, b.Line, b.Column, "", b.Message)
		}
		return nil
	}
	return value
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
