package moduleversion

import (
	"fmt"
	"slices"
	"strings"

	"github.com/altnovatrail/module-kit/internal/canonical"
)

// The rules of versions of a module, which refuse a new version rather than a module
// (format/README.md, sections 7 and 27).
const (
	RuleAbove           = "module-version.above"
	RuleStep            = "module-version.step"
	RuleRetiredComplete = "module-version.retired-complete"
	RuleRetiredKept     = "module-version.retired-kept"
	RuleRetiredKnown    = "module-version.retired-known"
	RuleIDsKept         = "module-version.ids-kept"
	RuleLanguagesKept   = "module-version.languages-kept"
	RuleBandsKept       = "module-version.bands-kept"
)

// Rules are the IDs of the rules module-version.*, in the order of section 27.
var Rules = []string{RuleAbove, RuleStep, RuleRetiredComplete, RuleRetiredKept, RuleRetiredKnown, RuleIDsKept, RuleLanguagesKept, RuleBandsKept}

// A Breach is a rule of versions that the new version breaks, and why.
type Breach struct {
	Rule    string `json:"rule"`
	Message string `json:"message"`
}

func (b Breach) String() string {
	return b.Rule + ": " + b.Message
}

// placeBreaches gives the breaches of the rules that read only the places of section 2, so that they
// hold through a move to another version of the format too (format/README.md, section 27): every rule
// but above and step. versions are the published ones in the order of Semantic Versioning, the newest
// last; next is the new version.
func placeBreaches(versions []publishedVersion, next places) []Breach {
	var breaches []Breach
	breach := func(rule, format string, args ...any) {
		breaches = append(breaches, Breach{rule, fmt.Sprintf(format, args...)})
	}
	newest := versions[len(versions)-1].places
	skills := setOf(next.ids["skills"])
	ever := map[string]bool{}
	for _, v := range versions {
		for _, id := range v.places.ids["skills"] {
			ever[id] = true
		}
	}
	for _, id := range sortedKeys(ever) {
		if _, retired := next.retired[id]; !skills[id] && !retired {
			breach(RuleRetiredComplete, "the skill %s of a published version is neither in skills nor in retired_skills", id)
		}
	}
	for _, id := range sortedKeys(newest.retired) {
		replaced := newest.retired[id]
		now, kept := next.retired[id]
		if !kept {
			breach(RuleRetiredKept, "the entry %s of retired_skills of %s is gone", id, newest.version)
			continue
		}
		// Each skill that replaced id stays, or, retired in turn, gives way to its own replacements.
		var expanded []string
		for _, skill := range replaced {
			if own, retired := next.retired[skill]; retired {
				expanded = append(expanded, own...)
			} else {
				expanded = append(expanded, skill)
			}
		}
		if !slices.Equal(sorted(expanded), sorted(now)) {
			breach(RuleRetiredKept, "%s is replaced by %s, and %s has it replaced by %s: an entry only expands",
				id, namesOrNone(now), newest.version, namesOrNone(replaced))
		}
	}
	for _, id := range sortedKeys(next.retired) {
		if _, before := newest.retired[id]; !before && !ever[id] {
			breach(RuleRetiredKnown, "the new entry %s of retired_skills names no skill of a published version", id)
		}
	}
	for _, kind := range idKinds {
		kept := setOf(newest.ids[kind])
		gone := map[string]bool{}
		for _, v := range versions {
			for _, id := range v.places.ids[kind] {
				if !kept[id] {
					gone[id] = true
				}
			}
		}
		for _, id := range next.ids[kind] {
			if gone[id] {
				breach(RuleIDsKept, "%s names one of the %s again, which the newest published version, %s, does not have: a published ID never names another entry",
					id, kind, newest.version)
			}
		}
	}
	nextLanguages, _ := canonical.Text(next.languages)
	for _, v := range versions {
		if languages, _ := canonical.Text(v.places.languages); string(languages) != string(nextLanguages) {
			breach(RuleLanguagesKept, "the languages differ from those of %s: a module in other languages is another module", v.places.version)
		}
		for i, band := range v.places.bands {
			if i >= len(next.bands) || next.bands[i] != band {
				breach(RuleBandsKept, "the bands of %s, %s, are not the first bands of the new version, %s",
					v.places.version, strings.Join(v.places.bands, ", "), namesOrNone(next.bands))
				break
			}
		}
	}
	return breaches
}

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	return keys
}

func namesOrNone(names []string) string {
	if len(names) == 0 {
		return "none"
	}
	return strings.Join(names, ", ")
}
