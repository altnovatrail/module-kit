package moduleversion

import (
	"fmt"
	"maps"
	"slices"

	"github.com/altnovatrail/spec/types/formatv0"

	"github.com/altnovatrail/module-kit/internal/semver"
	"github.com/altnovatrail/module-kit/modulepackage"
)

// places is what the comparison reads of a package by the places that the format keeps in every
// version (format/README.md, section 2): from its inventory, the format version, the module's ID and
// version; from its canonical view, the manifest's languages, the IDs of its bands in order, the keys
// of the skills of its skill graph and its retired skills with what replaced each, and the keys of its
// tasks, its lessons, its traps and its techniques. The rules module-version.* but above and step read
// these alone, so they hold through a move to another version of the format (section 27).
type places struct {
	formatVersion string
	id            formatv0.ModuleID
	version       semver.Version
	languages     map[string]any
	bands         []string
	retired       map[string][]string
	// The keys of the skills, the tasks, the lessons, the traps and the techniques, by idKinds.
	ids map[string][]string
}

// The kinds of entries whose IDs module-version.ids-kept keeps, in the order of the rule, each with
// the member of the canonical view that holds them and, for a catalog, the member within it.
var idKinds = []string{"skills", "tasks", "lessons", "traps", "techniques"}

var idPlaces = map[string][]string{
	"skills":     {"skill_graph", "skills"},
	"tasks":      {"tasks"},
	"lessons":    {"lessons"},
	"traps":      {"traps", "traps"},
	"techniques": {"techniques", "techniques"},
}

// placesOf reads the places of an inventory and a canonical view, as JSON values, by the places alone
// (format/README.md, section 2). A package that does not hold one where the format keeps it is
// modulepackage.ErrNotCanonical: no version of the format builds it.
func placesOf(inventory, view any) (places, error) {
	missing := func(what string) (places, error) {
		return places{}, fmt.Errorf("the package holds no %s where the format keeps it in every version: %w", what, modulepackage.ErrNotCanonical)
	}
	inv, _ := inventory.(map[string]any)
	var p places
	var ok bool
	if p.formatVersion, ok = inv["module_format_version"].(string); !ok || !formatVersionForm.MatchString(p.formatVersion) {
		return missing("module_format_version in its inventory")
	}
	id, ok := inv["id"].(string)
	if !ok {
		return missing("id in its inventory")
	}
	p.id = formatv0.ModuleID(id)
	version, _ := inv["version"].(string)
	var err error
	if p.version, err = semver.Parse(version); err != nil {
		return missing("version in its inventory")
	}
	v, _ := view.(map[string]any)
	manifest, ok := v["manifest"].(map[string]any)
	if !ok {
		return missing("manifest in its canonical view")
	}
	if p.languages, ok = manifest["languages"].(map[string]any); !ok {
		return missing("manifest.languages")
	}
	bands, ok := member[[]any](manifest, "bands")
	for _, band := range bands {
		b, _ := band.(map[string]any)
		id, isString := b["id"].(string)
		ok = ok && isString
		p.bands = append(p.bands, id)
	}
	if !ok {
		return missing("manifest.bands, each with its id")
	}
	graph, _ := member[map[string]any](v, "skill_graph")
	retired, ok := member[map[string]any](graph, "retired_skills")
	if !ok {
		return missing("retired_skills of the skill graph")
	}
	p.retired = map[string][]string{}
	for skill, entry := range retired {
		replaced, ok := stringsAt(entry)
		if !ok {
			return missing("replaced_by of the retired skill " + skill)
		}
		p.retired[skill] = replaced
	}
	p.ids = map[string][]string{}
	for _, kind := range idKinds {
		at := idPlaces[kind]
		holder := v
		if len(at) == 2 {
			var ok bool
			if holder, ok = member[map[string]any](v, at[0]); !ok {
				return missing(at[0])
			}
		}
		entries, ok := member[map[string]any](holder, at[len(at)-1])
		if !ok {
			return missing(kind)
		}
		p.ids[kind] = slices.Sorted(maps.Keys(entries))
	}
	return p, nil
}

// member gives the member of an object as T, or the zero T when the object has no such member; ok is
// false when the member is there and of another type.
func member[T any](object map[string]any, name string) (T, bool) {
	var zero T
	m, there := object[name]
	if !there {
		return zero, true
	}
	t, ok := m.(T)
	return t, ok
}

// stringsAt reads the replaced_by of an entry of retired_skills: a list of strings.
func stringsAt(entry any) ([]string, bool) {
	e, _ := entry.(map[string]any)
	list, ok := e["replaced_by"].([]any)
	if !ok {
		return nil, false
	}
	s := []string{}
	for _, item := range list {
		text, ok := item.(string)
		if !ok {
			return nil, false
		}
		s = append(s, text)
	}
	return s, true
}
