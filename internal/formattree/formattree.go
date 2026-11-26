package formattree

import (
	"io/fs"
	"path"
	"regexp"
	"slices"
	"strings"

	"github.com/altnovatrail/module-kit/internal/yaml"
)

// The kinds of the cases, as Classify gives them: of an example by its folder, and of a vector by
// the folder of vectors/ that holds it.
const (
	KindValidExample   = "examples/valid"
	KindInvalidExample = "examples/invalid"
	KindModuleExample  = "examples/module"
	KindRules          = "rules"
	KindOpenView       = "open-view"
	KindVerdicts       = "verdicts"
	KindLeaks          = "leaks"
	KindPages          = "pages"
	KindRandom         = "random"
	KindTemplates      = "templates"
	KindTiers          = "tiers"
	KindPackages       = "packages"
	KindIndexes        = "indexes"
	KindVersions       = "versions"
)

// The things a case folder holds, by their names.
const (
	CaseModule    = "module"    // of packages/ and versions/: the module, laid out as a module
	CaseModules   = "modules"   // of indexes/: a folder laid out as a module for each module
	CasePublished = "published" // of versions/: a folder for each published version
	CasePackage   = "package.yaml"
	CaseIndex     = "index.yaml"
	CaseVersion   = "version.yaml"
	CasePart      = "part.yaml" // of open-view/: the part whose open view the case gives
	CaseOpenView  = "open-view.yaml"
	CaseDigest    = "digest.yaml"
)

// The manifest of a folder laid out as a module, which a published version of an earlier format
// lacks: it is the files of its package.
const manifest = "manifest.yaml"

// LayoutParts are the parts of a module's layout, whose open views have vectors.
var LayoutParts = []string{"lesson", "manifest", "skill-graph", "solving-steps", "task", "techniques", "traps"}

// ExampleParts are the parts that have examples of their own, each with its schema: the parts of a
// module's layout, the inventory of a package and the module index.
var ExampleParts = slices.Concat(LayoutParts, []string{"inventory", "module-index"})

// VerdictKinds are the check kinds that give verdicts, by the names of their folders; rubric gives
// none.
var VerdictKinds = []string{"choice", "fraction", "number", "predicate", "time"}

// FileVectors are the vectors whose cases are files.
var FileVectors = []string{KindLeaks, KindPages, KindRandom, KindTemplates, KindTiers}

// FolderVectors are the vectors whose cases are folders of their own kind, by the things each
// folder holds and nothing else, in order.
var FolderVectors = map[string][]string{
	KindPackages: {CaseModule, CasePackage},
	KindIndexes:  {CaseIndex, CaseModules},
	KindVersions: {CaseModule, CasePublished, CaseVersion},
}

// OpenViewFiles are the files a case of the open view of part holds, and nothing else, in order.
func OpenViewFiles(part string) []string {
	if part == "task" {
		return []string{CaseDigest, CaseOpenView, CasePart}
	}
	return []string{CaseOpenView, CasePart}
}

// The ID of a format rule: <part>.<name>, in kebab-case.
var ruleID = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*\.[a-z0-9]+(-[a-z0-9]+)*$`)

// RuleFolder says whether name is a folder of rules/: the ID of a format rule. The rules of the
// module index and of versions of a module have their vectors among those of indexes and versions.
func RuleFolder(name string) bool {
	return ruleID.MatchString(name) && !strings.HasPrefix(name, "module-index.") &&
		!strings.HasPrefix(name, "module-version.")
}

// Classify says where a file of spec.Format belongs: the path of its case within the folder and
// the case's kind; no kind for a file of no case — the text, VERSION and the schemas —; and false
// for a file of no kind it knows.
func Classify(name string) (casePath, kind string, known bool) {
	parts := strings.Split(name, "/")
	isYAML := strings.HasSuffix(name, ".yaml")
	switch {
	case name == "README.md" || name == "VERSION" || name == "page.yaml":
		return "", "", true
	case len(parts) == 2 && parts[0] == "schemas":
		return "", "", strings.HasSuffix(parts[1], ".schema.json")
	case parts[0] == "examples":
		switch {
		case len(parts) >= 5 && parts[1] == "module" && parts[2] == "valid":
			return path.Join(parts[:4]...), KindModuleExample, true
		case len(parts) == 4 && slices.Contains(ExampleParts, parts[1]) && (parts[2] == "valid" || parts[2] == "invalid") && isYAML:
			return name, "examples/" + parts[2], true
		}
	case parts[0] == "vectors" && len(parts) >= 3:
		vectors := parts[1]
		switch {
		case vectors == KindRules && len(parts) >= 4 && RuleFolder(parts[2]):
			// A file of its own, or a folder laid out as a module.
			if len(parts) == 4 && !isYAML {
				return "", "", false
			}
			return path.Join(parts[:4]...), vectors, true
		case vectors == KindOpenView && len(parts) == 5 && slices.Contains(LayoutParts, parts[2]) &&
			slices.Contains(OpenViewFiles(parts[2]), parts[4]):
			return path.Join(parts[:4]...), vectors, true
		case vectors == KindVerdicts && len(parts) == 4 && slices.Contains(VerdictKinds, parts[2]) && isYAML:
			return name, vectors, true
		case slices.Contains(FileVectors, vectors) && len(parts) == 3 && isYAML:
			return name, vectors, true
		case len(parts) >= 4 && slices.Contains(FolderVectors[vectors], parts[3]) &&
			(len(parts) == 4) == strings.HasSuffix(parts[3], ".yaml"):
			return path.Join(parts[:3]...), vectors, true
		}
	}
	return "", "", false
}

// cases gives the kind of each case of fsys, by its path. A file of no kind is no case here: the
// vector test refuses it.
func cases(fsys fs.FS) (map[string]string, error) {
	found := map[string]string{}
	err := fs.WalkDir(fsys, ".", func(name string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() {
			return err
		}
		if casePath, kind, _ := Classify(name); kind != "" {
			found[casePath] = kind
		}
		return nil
	})
	return found, err
}

// Modules gives the folders of fsys, a tree laid out as spec.Format is, that are laid out as a
// module, each with the rule it is a case of, or "" for a module that breaks no rule: an example of
// a whole module, a case of a rule that is a folder, the module of a case of packages/ and of
// versions/, each module of a case of indexes/, and each published version of a case of versions/
// that is laid out as a module.
func Modules(fsys fs.FS) (map[string]string, error) {
	found, err := cases(fsys)
	if err != nil {
		return nil, err
	}
	modules := map[string]string{}
	inside := func(folder string, laidOut bool) error {
		entries, err := fs.ReadDir(fsys, folder)
		if err != nil {
			return err
		}
		for _, e := range entries {
			name := path.Join(folder, e.Name())
			if _, err := fs.Stat(fsys, path.Join(name, manifest)); e.IsDir() && (!laidOut || err == nil) {
				modules[name] = ""
			}
		}
		return nil
	}
	for casePath, kind := range found {
		switch kind {
		case KindModuleExample:
			modules[casePath] = ""
		case KindRules:
			if !strings.HasSuffix(casePath, ".yaml") {
				modules[casePath] = path.Base(path.Dir(casePath))
			}
		case KindPackages:
			modules[path.Join(casePath, CaseModule)] = ""
		case KindVersions:
			modules[path.Join(casePath, CaseModule)] = ""
			err = inside(path.Join(casePath, CasePublished), true)
		case KindIndexes:
			err = inside(path.Join(casePath, CaseModules), false)
		}
		if err != nil {
			return nil, err
		}
	}
	return modules, nil
}

// A PartFile is a file of a part on its own: the term of its part, the rule it is a case of, or "",
// and the kind of its case.
type PartFile struct {
	Part, Rule, Kind string
}

// PartFiles gives the files of fsys, a tree laid out as spec.Format is, that are files of a part on
// their own, by their paths: the examples of the parts of the layout, the cases in a file of a
// part's rule, and the parts of the vectors of the open view. A case of a rule yaml.* is a file that
// the YAML reader alone reads, of no part.
func PartFiles(fsys fs.FS) (map[string]PartFile, error) {
	found, err := cases(fsys)
	if err != nil {
		return nil, err
	}
	files := map[string]PartFile{}
	for casePath, kind := range found {
		parts := strings.Split(casePath, "/")
		switch kind {
		case KindValidExample, KindInvalidExample:
			if slices.Contains(LayoutParts, parts[1]) {
				files[casePath] = PartFile{parts[1], "", kind}
			}
		case KindRules:
			if rule := parts[2]; strings.HasSuffix(casePath, ".yaml") && !slices.Contains(yaml.Rules, rule) {
				part, _, _ := strings.Cut(rule, ".")
				files[casePath] = PartFile{part, rule, kind}
			}
		case KindOpenView:
			files[path.Join(casePath, CasePart)] = PartFile{parts[2], "", kind}
		}
	}
	return files, nil
}
