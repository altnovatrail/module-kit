package moduleversion

import (
	"fmt"
	"maps"
	"reflect"
	"slices"
	"strconv"

	"github.com/altnovatrail/module-kit/internal/canonical"
	"github.com/altnovatrail/module-kit/internal/schema"
	"github.com/altnovatrail/module-kit/modulepackage"
)

// differences adds the row of the table that no other row names: anything else that differs in the
// canonical view or in the listing of the device part, but the version in the manifest, is a patch
// (format/README.md, section 27). The JSON values of the two texts of each are compared member by
// member and element by element, and each place where they differ is a change: a member or an
// element added, removed, or another value. The instances of a template compare as a whole, since
// their places carry no meaning. A difference is not given at or within the place of a change that
// another row gives, nor in skill_tiers when the skill graph differs, nor in the instances of a
// template when its task differs: they follow those, and give no part of their own.
func (c *comparer) differences(before, after *modulepackage.Package) error {
	var found []change
	for _, file := range []string{modulepackage.ServerFile, modulepackage.DeviceFile} {
		b, err := canonical.Parse(before.Files[file])
		if err != nil {
			return fmt.Errorf("moduleversion: %s of %s %s: %w", file, before.Inventory.ID, before.Inventory.Version, err)
		}
		a, err := canonical.Parse(after.Files[file])
		if err != nil {
			return fmt.Errorf("moduleversion: %s of %s %s: %w", file, after.Inventory.ID, after.Inventory.Version, err)
		}
		found = append(found, diff(file, nil, b, a)...)
	}
	rows := c.changes
	for _, d := range found {
		covered := slices.ContainsFunc(rows, func(r change) bool { return r.File == d.File && within(d.tokens, r.tokens) })
		follows := d.File == modulepackage.ServerFile && len(d.tokens) > 0 && (d.tokens[0] == "skill_tiers" && differsWithin(found, rows, "skill_graph") ||
			d.tokens[0] == "instances" && len(d.tokens) > 1 && differsWithin(found, rows, "tasks", d.tokens[1]))
		if !covered && !follows {
			c.changes = append(c.changes, d)
		}
	}
	return nil
}

// within says whether a place is at or within another, token by token.
func within(place, other []string) bool {
	return len(place) >= len(other) && slices.Equal(place[:len(other)], other)
}

// differsWithin says whether a difference of the canonical view or a change of a row is at or
// within the place of these tokens.
func differsWithin(found, rows []change, tokens ...string) bool {
	in := func(d change) bool { return d.File == modulepackage.ServerFile && within(d.tokens, tokens) }
	return slices.ContainsFunc(found, in) || slices.ContainsFunc(rows, in)
}

// diff gives the places where two JSON values differ, within the place at, in file.
func diff(file string, at []string, before, after any) []change {
	here := func(what string) []change {
		return []change{{Change: Change{Part: Patch, File: file, Pointer: schema.Pointer(at), What: what}, tokens: slices.Clone(at)}}
	}
	if file == modulepackage.ServerFile && (slices.Equal(at, []string{"manifest", "version"}) || len(at) == 2 && at[0] == "instances") {
		if at[0] == "instances" && !reflect.DeepEqual(before, after) {
			return here("the numbers of the instances change")
		}
		return nil
	}
	switch b := before.(type) {
	case map[string]any:
		a, ok := after.(map[string]any)
		if !ok {
			break
		}
		var found []change
		for _, name := range slices.Sorted(maps.Keys(b)) {
			place := append(slices.Clone(at), name)
			if v, ok := a[name]; ok {
				found = append(found, diff(file, place, b[name], v)...)
			} else if !(file == modulepackage.ServerFile && slices.Equal(place, []string{"manifest", "version"})) {
				found = append(found, removed(file, place)...)
			}
		}
		for _, name := range slices.Sorted(maps.Keys(a)) {
			if _, ok := b[name]; !ok {
				found = append(found, added(file, append(slices.Clone(at), name))...)
			}
		}
		return found
	case []any:
		a, ok := after.([]any)
		if !ok {
			break
		}
		var found []change
		for i := range min(len(b), len(a)) {
			found = append(found, diff(file, append(slices.Clone(at), strconv.Itoa(i)), b[i], a[i])...)
		}
		for i := len(a); i < len(b); i++ {
			found = append(found, removed(file, append(slices.Clone(at), strconv.Itoa(i)))...)
		}
		for i := len(b); i < len(a); i++ {
			found = append(found, added(file, append(slices.Clone(at), strconv.Itoa(i)))...)
		}
		return found
	default:
		if reflect.DeepEqual(before, after) {
			return nil
		}
	}
	return here("changes")
}

func added(file string, at []string) []change {
	return []change{{Change: Change{Part: Patch, File: file, Pointer: schema.Pointer(at), What: "is added"}, tokens: at}}
}

func removed(file string, at []string) []change {
	return []change{{Change: Change{Part: Patch, File: file, Pointer: schema.Pointer(at), What: "is removed"}, tokens: at}}
}
