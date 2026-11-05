package moduleversion

import (
	"errors"
	"fmt"
	"io/fs"
	"slices"
	"strings"

	"github.com/altnovatrail/spec/types/formatv0"

	"github.com/altnovatrail/module-kit/modulepackage"
)

// A Published is a published version of a module: its package as files, without device/, and the
// digest of the package by which the module index of the newest release names it
// (format/README.md, sections 25–27). A withdrawn version is a published one too.
type Published struct {
	Digest formatv0.Digest
	Files  fs.FS
}

// A Comparison is what the comparison of a new version of a module with its published versions
// gives (format/README.md, section 27): the module and the new version, the published versions in
// the order of Semantic Versioning, and the newest of them, with which the new version is compared.
// Unchanged says that the new version is the newest published one as it was, and then nothing is
// compared. Otherwise Part is the part of the version that the comparison finds, Changes every change
// that it finds, Next the next versions after the newest published one whose step raises a part no
// lower than Part, and Raised the part that the step to the new version raises, None when it is not
// the next version after the newest one by one of its parts. Breaches are the rules module-version.*
// that the new version breaks, in the order the rules are checked: above, step, and then those that
// read the places every version keeps; a comparison with a breach refuses the new version, and diff
// exits 1. A first version, with no published one, may be any, and nothing is compared.
type Comparison struct {
	Module    formatv0.ModuleID `json:"module"`
	Version   string            `json:"version"`
	Published []string          `json:"published"`
	Newest    string            `json:"newest,omitzero"`
	Unchanged bool              `json:"unchanged,omitzero"`
	Part      Part              `json:"version_part"`
	Next      []Step            `json:"next_versions,omitzero"`
	Raised    Part              `json:"raised"`
	Changes   []Change          `json:"changes,omitzero"`
	Breaches  []Breach          `json:"breaches,omitzero"`
}

// Refused says whether a rule of versions refuses the new version.
func (c *Comparison) Refused() bool {
	return len(c.Breaches) > 0
}

// String writes the comparison as diff reports it: a line of the module and the versions, then the
// part found, the next versions and the part raised, and then each change on a line of its own, in
// the order of Changes.
func (c *Comparison) String() string {
	head := fmt.Sprintf("%s %s", c.Module, c.Version)
	switch {
	case c.Newest == "":
		return head + ": the first version, with nothing to compare it with"
	case c.Unchanged:
		return head + ": the newest published version as it was, so nothing is compared"
	}
	lines := []string{
		head + ", compared with " + c.Newest,
		"published: " + strings.Join(c.Published, ", "),
		"found: " + c.Part.String(),
	}
	var next []string
	for _, s := range c.Next {
		next = append(next, fmt.Sprintf("%s (%s)", s.Version, s.Part))
	}
	lines = append(lines, "next versions after "+c.Newest+": "+strings.Join(next, ", "))
	if c.Raised == None {
		lines = append(lines, c.Version+" is not the next version after "+c.Newest+" by one of its parts")
	} else {
		lines = append(lines, c.Version+" raises: "+c.Raised.String())
	}
	for _, breach := range c.Breaches {
		lines = append(lines, "refused: "+breach.String())
	}
	for _, change := range c.Changes {
		lines = append(lines, change.String())
	}
	return strings.Join(lines, "\n")
}

// Compare compares a new version of a module, its package next, with its published versions
// (format/README.md, section 27). It reads each published package by the chain of digests from its
// digest — a package of the version of the format that module-kit reads by
// modulepackage.ReadParts, one of an earlier version by the places that every version keeps — and
// takes the newest by Semantic Versioning, withdrawn or not. When the new version is that version and
// its package has that digest, the module has not changed, and nothing is compared. Otherwise the new
// version is compared with it by Changes, or, when the newest names an earlier version of the format,
// the part is major while the format is v0; and the rules module-version.* are held: above against
// every published version, step and retired-kept against the newest, and the others against every
// one by the places of section 2, so through a move of the format too. With no published version,
// the new one is the first.
//
// The error is for a published package that the chain of digests or the places of section 2 refuse —
// of a newer version of the format, modulepackage.ErrFormatVersion, among them —, one of another
// module than next, and two of one version.
func Compare(next *modulepackage.Package, published []Published) (*Comparison, error) {
	if next == nil {
		return nil, errors.New("moduleversion: no new version to compare")
	}
	nextPlaces, err := placesOfPackage(next)
	if err != nil {
		return nil, fmt.Errorf("moduleversion: the new version: %w", err)
	}
	var versions []publishedVersion
	for _, p := range published {
		v, err := readPublished(p)
		if err != nil {
			return nil, fmt.Errorf("moduleversion: the published package %s: %w", p.Digest, err)
		}
		if v.places.id != next.Inventory.ID {
			return nil, fmt.Errorf("moduleversion: the published package %s is of the module %s, and the new version of %s", p.Digest, v.places.id, next.Inventory.ID)
		}
		versions = append(versions, v)
	}
	slices.SortFunc(versions, func(a, b publishedVersion) int { return a.places.version.Compare(b.places.version) })
	c := &Comparison{Module: next.Inventory.ID, Version: string(next.Inventory.Version), Published: []string{}}
	for i, v := range versions {
		if i > 0 && v.places.version.Compare(versions[i-1].places.version) == 0 {
			return nil, fmt.Errorf("moduleversion: two published packages of %s, %s and %s", v.places.version, versions[i-1].digest, v.digest)
		}
		c.Published = append(c.Published, v.places.version.String())
	}
	if len(versions) == 0 {
		return c, nil
	}
	newest := versions[len(versions)-1]
	c.Newest = newest.places.version.String()
	if c.Newest == c.Version && newest.digest == next.Digest {
		c.Unchanged = true
		return c, nil
	}
	// module-version.above: above every published version, withdrawn ones too.
	for _, v := range versions {
		switch order := nextPlaces.version.Compare(v.places.version); {
		case order == 0:
			c.Breaches = append(c.Breaches, Breach{RuleAbove, fmt.Sprintf("%s is published already, with the package %s: a version is published once", c.Version, v.digest)})
		case order < 0:
			c.Breaches = append(c.Breaches, Breach{RuleAbove, fmt.Sprintf("%s is below %s, which is published: a new version stands above every published one", c.Version, v.places.version)})
		}
	}
	// The part: through a move to another version of the format, a major one while the format is v0;
	// otherwise what the comparison with the newest published version finds.
	if newest.pkg != nil {
		if c.Part, c.Changes, err = Changes(newest.pkg, next); err != nil {
			return nil, err
		}
	} else {
		c.Part = Major
		c.Changes = []Change{{Part: Major, File: modulepackage.InventoryFile, Pointer: "#/module_format_version",
			What: fmt.Sprintf("the module moves from version %s of the format to %s", newest.places.formatVersion, nextPlaces.formatVersion)}}
	}
	steps, err := Steps(c.Newest)
	if err != nil {
		return nil, err
	}
	for _, s := range steps {
		if s.Part >= c.Part {
			c.Next = append(c.Next, s)
		}
	}
	if c.Raised, err = StepPart(c.Newest, c.Version); err != nil {
		return nil, err
	}
	// module-version.step applies to a new version that module-version.above lets pass.
	if len(c.Breaches) == 0 {
		switch {
		case c.Raised == None:
			c.Breaches = append(c.Breaches, Breach{RuleStep, fmt.Sprintf("%s is not the next version after %s by one of its parts", c.Version, c.Newest)})
		case c.Raised < c.Part:
			c.Breaches = append(c.Breaches, Breach{RuleStep, fmt.Sprintf("%s raises the %s part after %s, and the comparison finds a %s change: %s",
				c.Version, c.Raised, c.Newest, c.Part, changesOf(c.Changes, c.Part))})
		}
	}
	c.Breaches = append(c.Breaches, placeBreaches(versions, nextPlaces)...)
	return c, nil
}

// changesOf writes the changes of one part, as a breach of module-version.step names them.
func changesOf(changes []Change, part Part) string {
	var said []string
	for _, change := range changes {
		if change.Part == part {
			said = append(said, change.Pointer+": "+change.What)
		}
	}
	return strings.Join(said, "; ")
}
