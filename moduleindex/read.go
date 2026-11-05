package moduleindex

import (
	"errors"
	"fmt"

	"github.com/altnovatrail/spec/types/formatv0"

	"github.com/altnovatrail/module-kit/internal/canonical"
	"github.com/altnovatrail/module-kit/internal/schema"
	"github.com/altnovatrail/module-kit/internal/semver"
	"github.com/altnovatrail/module-kit/module"
)

// ErrNotIndex is a text that Read does not read as a module index: no JSON object, or one whose
// members at the places every version of the format keeps are not in their forms.
var ErrNotIndex = errors.New("no module index by the places every version of the format keeps")

// The members that keep their places and meanings in every version of the format: at the root of
// the index, and in an entry (format/README.md, sections 2 and 26).
var (
	rootMembers  = []string{versionField, "modules"}
	entryMembers = []string{versionField, "package_digest", "release", "teaches", withdrawnField}
)

const (
	versionField   = "module_format_version"
	withdrawnField = "withdrawn"
)

// Read reads a module index of any version of the format by the places that every version keeps,
// and not by the schema of its own version, which refuses what a later version adds
// (format/README.md, sections 2 and 26): module_format_version and modules at its root, and
// module_format_version, package_digest, release, teaches and withdrawn in an entry. A member at
// another place, which a later version may add, is not read. Those members keep their forms in
// every version too, so one that is not in its form is no index of any version, and Read gives
// ErrNotIndex. A version is withdrawn whatever its reason: Withdrawn holds the reason as the index
// writes it, one that a later version adds among them.
//
// Read gives every entry, of whatever format version, so that the comparison of versions sees
// every published version of a module (format/README.md, section 27). A pick of a version, Newest,
// skips an entry whose format version module-kit does not read, rather than refuse the index. Read
// checks no signature: the index was checked where it entered NovaTrail (R77 spec). Nor does it
// hold the text to the canonical text, which a later version may change; a name that stands twice
// in one object, which no build writes, counts by its last member.
func Read(data []byte) (formatv0.ModuleIndex, error) {
	value, err := canonical.Parse(data)
	if err != nil {
		return formatv0.ModuleIndex{}, fmt.Errorf("moduleindex: no JSON text: %v: %w", err, ErrNotIndex)
	}
	root, ok := value.(map[string]any)
	if !ok {
		return formatv0.ModuleIndex{}, fmt.Errorf("moduleindex: the text is no JSON object: %w", ErrNotIndex)
	}
	kept := members(root, rootMembers)
	// The reason of each withdrawn version, which may be one that no schema module-kit has names:
	// the schema checks the kept members without it, and it comes back after.
	reasons := map[formatv0.ModuleID]map[formatv0.ModuleVersion]string{}
	if modules, ok := kept["modules"].(map[string]any); ok {
		for id, versions := range modules {
			versions, ok := versions.(map[string]any)
			if !ok {
				continue
			}
			for version, entry := range versions {
				entry, ok := entry.(map[string]any)
				if !ok {
					continue
				}
				e := members(entry, entryMembers)
				if w, ok := e[withdrawnField]; ok {
					reason, ok := w.(string)
					if !ok {
						return formatv0.ModuleIndex{}, fmt.Errorf("moduleindex: %s %s is withdrawn by %v, which is no reason: %w", id, version, w, ErrNotIndex)
					}
					delete(e, withdrawnField)
					if reasons[formatv0.ModuleID(id)] == nil {
						reasons[formatv0.ModuleID(id)] = map[formatv0.ModuleVersion]string{}
					}
					reasons[formatv0.ModuleID(id)][formatv0.ModuleVersion(version)] = reason
				}
				versions[version] = e
			}
		}
	}
	if errs := schema.Validate("module-index.schema.json", kept); len(errs) > 0 {
		return formatv0.ModuleIndex{}, fmt.Errorf("moduleindex: %s: %w", errs[0], ErrNotIndex)
	}
	index, err := schema.Decode[formatv0.ModuleIndex](kept)
	if err != nil {
		panic(fmt.Sprintf("moduleindex: an index its schema accepts: %v", err))
	}
	for id, versions := range reasons {
		for version, reason := range versions {
			entry := index.Modules[id][version]
			r := formatv0.WithdrawalReason(reason)
			entry.Withdrawn = &r
			index.Modules[id][version] = entry
		}
	}
	return index, nil
}

// members gives the members of an object at the places it names, which it has.
func members(object map[string]any, places []string) map[string]any {
	kept := map[string]any{}
	for _, name := range places {
		if member, ok := object[name]; ok {
			kept[name] = member
		}
	}
	return kept
}

// Newest gives the version of the module id that a new attempt takes: the newest by Semantic
// Versioning of the versions that the index names, of those not withdrawn, whatever the reason, and
// whose format version module-kit reads, ReadsFormat of package module (format/README.md, section
// 26; R63 spec; R16). ok is false when the index names no such version. The same version is the
// one by which a module counts as teaching the anchor skills of its entry (concepts 4 spec). A
// version not in the form X.Y.Z, which no index that Read or Build gives names, is skipped too.
func Newest(index formatv0.ModuleIndex, id formatv0.ModuleID) (version formatv0.ModuleVersion, ok bool) {
	var newest semver.Version
	for v, entry := range index.Modules[id] {
		parsed, err := semver.Parse(string(v))
		if err != nil || entry.Withdrawn != nil || !module.ReadsFormat(string(entry.ModuleFormatVersion)) {
			continue
		}
		if !ok || parsed.Compare(newest) > 0 {
			version, newest, ok = v, parsed, true
		}
	}
	return version, ok
}
