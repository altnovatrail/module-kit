package moduleindex

import (
	"bytes"
	"cmp"
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/altnovatrail/spec/types/formatv0"

	"github.com/altnovatrail/module-kit/internal/canonical"
	"github.com/altnovatrail/module-kit/internal/schema"
	"github.com/altnovatrail/module-kit/internal/semver"
	"github.com/altnovatrail/module-kit/module"
	"github.com/altnovatrail/module-kit/modulepackage"
)

// File is the name of the module index among the files of a release of modules
// (format/README.md, section 26).
const File = "module-index.json"

// The rules of the module index, which refuse to build an index rather than a module
// (format/README.md, sections 7 and 26).
const (
	RuleEntriesKept      = "module-index.entries-kept"
	RuleOnePackage       = "module-index.one-package"
	RuleKnownWithdrawals = "module-index.known-withdrawals"
	RuleVersionAbove     = "module-index.version-above"
)

// Rules lists the rules of the module index, in the order of the table of section 26 of
// format/README.md.
var Rules = []string{RuleEntriesKept, RuleOnePackage, RuleKnownWithdrawals, RuleVersionAbove}

// The reasons Build builds no index other than the rules of the index, which its error wraps.
var (
	// ErrPrevious is a previous index that cannot be read: one built by a later format version than
	// module-kit reads, or one that is not what a build writes — the canonical text of a value valid
	// against the schema of the module index of the version module-kit reads, with teaches in order
	// —, or none where the release is not the first, or one where it is. Such a release fails: it does
	// not start a new index (format/README.md, section 26).
	ErrPrevious = errors.New("the previous index cannot be read")
	// ErrRelease is a release of which no index can be built: a tag that is not one of a release of
	// modules, withdrawals not in the form of $defs/withdrawals, a package that package
	// modulepackage did not build or read, or no version at all to name.
	ErrRelease = errors.New("not a release whose index can be built")
)

// A Release is what a release of modules gives the build of its index (format/README.md, section
// 26): the index of the release before it, its own packages, its tag and its withdrawals.
type Release struct {
	// Previous is the index of the release before, module-index.json as that release holds it,
	// byte for byte. The first release of modules has none, and says so by First: an empty
	// Previous is no index, so a file that could not be read does not start a new one.
	Previous []byte
	First    bool
	// Tag is the tag of the release, vX.Y.Z, which each new entry names.
	Tag formatv0.Release
	// Packages are the packages of the release, as package modulepackage built or read them. A
	// release may build every module or only those that changed: a package of a version that the
	// previous index names keeps its entry.
	Packages []*modulepackage.Package
	// Withdrawals are the versions that the release withdraws from publication, by module ID and
	// version, each with its reason, as $defs/withdrawals of module-index.schema.json gives them.
	// A release that withdraws none has nil.
	Withdrawals formatv0.Withdrawals
}

// A Breach is a rule of the module index that the build of a release breaks, at the version of a
// module that breaks it, with what is wrong.
type Breach struct {
	Rule    string
	Module  formatv0.ModuleID
	Version formatv0.ModuleVersion
	Message string
}

func (b Breach) String() string {
	return fmt.Sprintf("%s: %s %s: %s", b.Rule, b.Module, b.Version, b.Message)
}

// A Refusal is why Build builds no index: the release breaks rules of the module index, and it
// names every breach (format/README.md, section 7).
type Refusal struct {
	Breaches []Breach
}

func (r *Refusal) Error() string {
	lines := make([]string, len(r.Breaches))
	for i, b := range r.Breaches {
		lines[i] = b.String()
	}
	return "moduleindex: the release breaks the rules of the module index, so it has no index:\n" + strings.Join(lines, "\n")
}

// Build builds the module index of a release by the format version module-kit reads,
// FormatVersion of package module, and gives its file, module-index.json: the canonical text of the
// index (format/README.md, sections 25 and 26; R63 spec). Every entry of the previous index stands
// in the new one as it was; a package of a version that the previous index names keeps that entry,
// its release among it; every other package gets a new entry, which names the release's tag; and
// last, the release's withdrawals are marked.
//
// A release that breaks a rule of the index gives a *Refusal, with every breach in the order of the
// module, the version and the rule; the build is refused, rather than an index built without what
// breaks the rule. Any other error wraps ErrPrevious or ErrRelease, and comes before a refusal: the
// rules are read only on a release whose previous index, tag, withdrawals and packages are in their
// forms.
func Build(r Release) ([]byte, error) {
	previous, err := readPrevious(r)
	if err != nil {
		return nil, err
	}
	if errs := schema.Validate("module-index.schema.json#/$defs/release", string(r.Tag)); len(errs) > 0 {
		return nil, fmt.Errorf("moduleindex: the tag %q is not one of a release of modules, vX.Y.Z: %w", r.Tag, ErrRelease)
	}
	if r.Withdrawals != nil {
		if errs := schema.Validate("module-index.schema.json#/$defs/withdrawals", schema.ValueOf(r.Withdrawals)); len(errs) > 0 {
			return nil, fmt.Errorf("moduleindex: the withdrawals are not in the form of $defs/withdrawals: %s: %w", errs[0], ErrRelease)
		}
	}
	packages, err := released(r.Packages, r.Tag)
	if err != nil {
		return nil, err
	}
	modules := previous.Modules
	// The versions of each module that the previous index names, withdrawn ones too, which a new
	// version stands above. Its versions passed its schema, so each is in the form X.Y.Z.
	published := map[formatv0.ModuleID][]semver.Version{}
	for id, versions := range modules {
		for v := range versions {
			published[id] = append(published[id], mustParse(v))
		}
	}
	var breaches []Breach
	// The digest of each version that a package of the release gives.
	built := map[formatv0.ModuleID]map[formatv0.ModuleVersion]formatv0.Digest{}
	for _, p := range packages {
		id, version, digest := p.id, p.version, p.entry.PackageDigest
		at := Breach{Module: id, Version: version}
		if other, seen := built[id][version]; seen {
			if other != digest {
				breaches = append(breaches, at.breaks(RuleOnePackage, "the release builds two packages of it, %s and %s: a version has one package", other, digest))
			}
			continue
		}
		if built[id] == nil {
			built[id] = map[formatv0.ModuleVersion]formatv0.Digest{}
		}
		built[id][version] = digest
		if kept, ok := modules[id][version]; ok {
			if kept.PackageDigest != digest {
				breaches = append(breaches, at.breaks(RuleOnePackage, "the previous index names its package %s, and the release builds %s: a version is published once", kept.PackageDigest, digest))
			}
			continue
		}
		// The comparison of versions asks the same of a new version, and so the index catches a
		// release whose CI did not compare them (format/README.md, section 27). The entry stands
		// even so, so that a withdrawal of it breaks no other rule.
		var above []semver.Version
		for _, v := range published[id] {
			if v.Compare(p.parsed) > 0 {
				above = append(above, v)
			}
		}
		if len(above) > 0 {
			slices.SortFunc(above, semver.Version.Compare)
			breaches = append(breaches, at.breaks(RuleVersionAbove, "it is new in the release, and the previous index names %s above it: a new version stands above every published one, withdrawn ones too", joined(above)))
		}
		if modules[id] == nil {
			modules[id] = map[formatv0.ModuleVersion]formatv0.Entry{}
		}
		modules[id][version] = p.entry
	}
	breaches = append(breaches, withdraw(modules, r.Withdrawals)...)
	if len(breaches) > 0 {
		slices.SortFunc(breaches, func(a, b Breach) int {
			return cmp.Or(strings.Compare(string(a.Module), string(b.Module)), mustParse(a.Version).Compare(mustParse(b.Version)),
				strings.Compare(a.Rule, b.Rule), strings.Compare(a.Message, b.Message))
		})
		return nil, &Refusal{Breaches: breaches}
	}
	if len(modules) == 0 {
		return nil, fmt.Errorf("moduleindex: the release is the first and builds no package, and an index names at least one version: %w", ErrRelease)
	}
	text, err := canonical.Text(schema.ValueOf(formatv0.ModuleIndex{ModuleFormatVersion: module.FormatVersion, Modules: modules}))
	if err != nil {
		// The index holds strings alone.
		panic(fmt.Sprintf("moduleindex: the canonical text of an index: %v", err))
	}
	return text, nil
}

// readPrevious reads the index of the release before, or gives an empty one to the first release.
// The build keeps every entry as it was in an index of its own version, which its schema accepts,
// so a previous index that this schema refuses — such as one with a member or a reason that a later
// version adds — cannot be read, and nor can one built by a later version of the format, which a
// release of modules gives only when its CLI went back to an older module-kit. What can be read is
// what a build writes: the canonical text, with teaches in order.
func readPrevious(r Release) (formatv0.ModuleIndex, error) {
	switch {
	case r.First && len(r.Previous) > 0:
		return formatv0.ModuleIndex{}, fmt.Errorf("moduleindex: the release is the first, and it is given a previous index: %w", ErrPrevious)
	case r.First:
		return formatv0.ModuleIndex{Modules: map[formatv0.ModuleID]map[formatv0.ModuleVersion]formatv0.Entry{}}, nil
	case len(r.Previous) == 0:
		return formatv0.ModuleIndex{}, fmt.Errorf("moduleindex: the release is given no previous index, and it is not the first: %w", ErrPrevious)
	}
	value, err := canonical.Parse(r.Previous)
	if err != nil {
		return formatv0.ModuleIndex{}, fmt.Errorf("moduleindex: the previous index is no JSON text: %v: %w", err, ErrPrevious)
	}
	// The format version first, at the place every version keeps (format/README.md, section 2), so
	// that an index of a later version is refused by it rather than by what the version added.
	root, _ := value.(map[string]any)
	if v, ok := root[versionField].(string); ok && module.NewerFormat(v) {
		return formatv0.ModuleIndex{}, fmt.Errorf("moduleindex: the previous index was built by the format version %s, newer than %s, which module-kit reads: %w",
			v, module.FormatVersion, ErrPrevious)
	}
	if errs := schema.Validate("module-index.schema.json", value); len(errs) > 0 {
		return formatv0.ModuleIndex{}, fmt.Errorf("moduleindex: the previous index is not valid against module-index.schema.json of the format %s: %s: %w",
			module.FormatVersion, errs[0], ErrPrevious)
	}
	if text, err := canonical.Text(value); err != nil || !bytes.Equal(text, r.Previous) {
		return formatv0.ModuleIndex{}, fmt.Errorf("moduleindex: the previous index is not the canonical text of its value: %w", ErrPrevious)
	}
	index, err := schema.Decode[formatv0.ModuleIndex](value)
	if err != nil {
		panic(fmt.Sprintf("moduleindex: an index its schema accepts: %v", err))
	}
	for id, versions := range index.Modules {
		for version, entry := range versions {
			// The schema holds the items unique, so in order is strictly ascending.
			if !slices.IsSorted(entry.Teaches) {
				return formatv0.ModuleIndex{}, fmt.Errorf("moduleindex: in the previous index, the teaches of %s %s are not in the order of the code units of their IDs: %w",
					id, version, ErrPrevious)
			}
		}
	}
	return index, nil
}

// A releasedPackage is a package of the release, with its version read and the entry it gets when
// the release publishes it first.
type releasedPackage struct {
	id      formatv0.ModuleID
	version formatv0.ModuleVersion
	parsed  semver.Version
	entry   formatv0.Entry
}

// released reads the packages of a release, each as package modulepackage builds or reads one: a
// package of the format version module-kit reads, whose module ID, version and entry are in their
// forms. It gives them in the order of their module IDs, versions and digests, so that the build and
// its breaches do not depend on the order they come in.
func released(packages []*modulepackage.Package, tag formatv0.Release) ([]releasedPackage, error) {
	out := make([]releasedPackage, 0, len(packages))
	for _, pkg := range packages {
		if pkg == nil {
			return nil, fmt.Errorf("moduleindex: a package of the release is nil: %w", ErrRelease)
		}
		p := releasedPackage{id: pkg.Inventory.ID, version: pkg.Inventory.Version, entry: entryOf(pkg, tag)}
		parsed, err := semver.Parse(string(pkg.Inventory.Version))
		bad := append(schema.Validate("common.schema.json#/$defs/module_id", string(pkg.Inventory.ID)),
			schema.Validate("module-index.schema.json#/$defs/entry", schema.ValueOf(p.entry))...)
		if err != nil || len(bad) > 0 || !module.ReadsFormat(string(pkg.Inventory.ModuleFormatVersion)) {
			return nil, fmt.Errorf("moduleindex: %s %s of the format %s is no package that package modulepackage builds or reads: %w",
				pkg.Inventory.ID, pkg.Inventory.Version, pkg.Inventory.ModuleFormatVersion, ErrRelease)
		}
		p.parsed = parsed
		out = append(out, p)
	}
	slices.SortFunc(out, func(a, b releasedPackage) int {
		return cmp.Or(strings.Compare(string(a.id), string(b.id)), a.parsed.Compare(b.parsed),
			strings.Compare(string(a.entry.PackageDigest), string(b.entry.PackageDigest)))
	})
	return out, nil
}

// entryOf gives the entry of a package that a release publishes first: its format version, the
// digest of its package, the release's tag, and the anchor skills it teaches, when it teaches any
// (format/README.md, section 26).
func entryOf(pkg *modulepackage.Package, tag formatv0.Release) formatv0.Entry {
	return formatv0.Entry{
		ModuleFormatVersion: pkg.Inventory.ModuleFormatVersion,
		PackageDigest:       pkg.Digest,
		Release:             tag,
		Teaches:             teachesOf(pkg.View),
	}
}

// teachesOf gives the anchor skills that the teaches of a skill of a canonical view's skill graph
// names, each once, in the order of the UTF-16 code units of their IDs; nil when it names none.
func teachesOf(view formatv0.CanonicalView) []formatv0.AnchorSkillID {
	if view.SkillGraph == nil {
		return nil
	}
	var teaches []formatv0.AnchorSkillID
	for _, skill := range view.SkillGraph.Skills {
		for _, anchor := range skill.Teaches {
			teaches = append(teaches, formatv0.AnchorSkillID(anchor))
		}
	}
	// The ID of an anchor skill is ASCII, core. and a name in kebab-case, so the order of its bytes
	// is the order of its UTF-16 code units.
	slices.Sort(teaches)
	return slices.Compact(teaches)
}

// withdraw marks the withdrawals of a release in the entries of its index, and gives the breaches:
// a withdrawal of a version that the index does not name, and one for another reason than the
// version is withdrawn for already. A withdrawal for the same reason changes nothing.
func withdraw(modules map[formatv0.ModuleID]map[formatv0.ModuleVersion]formatv0.Entry, withdrawals formatv0.Withdrawals) []Breach {
	var breaches []Breach
	for id, versions := range withdrawals {
		for version, reason := range versions {
			at := Breach{Module: id, Version: version}
			entry, ok := modules[id][version]
			switch {
			case !ok:
				breaches = append(breaches, at.breaks(RuleKnownWithdrawals, "the release withdraws it for %s, and the index names no such version", reason))
			case entry.Withdrawn == nil:
				entry.Withdrawn = &reason
				modules[id][version] = entry
			case *entry.Withdrawn != reason:
				breaches = append(breaches, at.breaks(RuleEntriesKept, "it is withdrawn for %s, and the release withdraws it for %s: a withdrawal keeps its reason", *entry.Withdrawn, reason))
			}
		}
	}
	return breaches
}

// breaks gives the breach of a rule at the version of b.
func (b Breach) breaks(rule, format string, args ...any) Breach {
	b.Rule, b.Message = rule, fmt.Sprintf(format, args...)
	return b
}

// mustParse reads a version of a module that a schema or released has held to the form X.Y.Z; one
// not in it is a fault of module-kit.
func mustParse(v formatv0.ModuleVersion) semver.Version {
	parsed, err := semver.Parse(string(v))
	if err != nil {
		panic(fmt.Sprintf("moduleindex: a version the build holds to its form: %v", err))
	}
	return parsed
}

// joined writes versions as a list in a sentence: 0.1.0, 0.2.0 and 1.0.0.
func joined(versions []semver.Version) string {
	words := make([]string, len(versions))
	for i, v := range versions {
		words[i] = v.String()
	}
	if len(words) < 2 {
		return strings.Join(words, "")
	}
	return strings.Join(words[:len(words)-1], ", ") + " and " + words[len(words)-1]
}
