package moduleversion

import (
	"errors"
	"fmt"
	"io/fs"
	"slices"
	"strings"

	"github.com/altnovatrail/spec/types/formatv0"

	"github.com/altnovatrail/module-kit/internal/semver"
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
// the next version after the newest one by one of its parts. A first version, with no published
// one, may be any, and nothing is compared.
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
	for _, change := range c.Changes {
		lines = append(lines, change.String())
	}
	return strings.Join(lines, "\n")
}

// Compare compares a new version of a module, its package next, with its published versions
// (format/README.md, section 27). It reads each published package by the chain of digests from its
// digest, modulepackage.ReadParts, and takes the newest by Semantic Versioning, withdrawn or not:
// when the new version is that version and its package has that digest, the module has not changed;
// otherwise the new version is compared with it by Changes. With no published version, the new one
// is the first.
//
// The error is for a published package that ReadParts refuses — of an earlier version of the format
// too, modulepackage.ErrFormatVersion —, one of another module than next, and two of one version.
func Compare(next *modulepackage.Package, published []Published) (*Comparison, error) {
	if next == nil {
		return nil, errors.New("moduleversion: no new version to compare")
	}
	type version struct {
		v   semver.Version
		pkg *modulepackage.Package
	}
	var versions []version
	for _, p := range published {
		pkg, err := modulepackage.ReadParts(p.Files, p.Digest)
		if err != nil {
			return nil, fmt.Errorf("moduleversion: the published package %s: %w", p.Digest, err)
		}
		if pkg.Inventory.ID != next.Inventory.ID {
			return nil, fmt.Errorf("moduleversion: the published package %s is of the module %s, and the new version of %s", p.Digest, pkg.Inventory.ID, next.Inventory.ID)
		}
		v, err := semver.Parse(string(pkg.Inventory.Version))
		if err != nil {
			return nil, fmt.Errorf("moduleversion: the published package %s: %w", p.Digest, err)
		}
		versions = append(versions, version{v, pkg})
	}
	slices.SortFunc(versions, func(a, b version) int { return a.v.Compare(b.v) })
	c := &Comparison{Module: next.Inventory.ID, Version: string(next.Inventory.Version), Published: []string{}}
	for i, v := range versions {
		if i > 0 && v.v.Compare(versions[i-1].v) == 0 {
			return nil, fmt.Errorf("moduleversion: two published packages of %s, %s and %s", v.v, versions[i-1].pkg.Digest, v.pkg.Digest)
		}
		c.Published = append(c.Published, v.v.String())
	}
	if len(versions) == 0 {
		return c, nil
	}
	newest := versions[len(versions)-1].pkg
	c.Newest = string(newest.Inventory.Version)
	if newest.Inventory.Version == next.Inventory.Version && newest.Digest == next.Digest {
		c.Unchanged = true
		return c, nil
	}
	part, changes, err := Changes(newest, next)
	if err != nil {
		return nil, err
	}
	c.Part, c.Changes = part, changes
	steps, err := Steps(c.Newest)
	if err != nil {
		return nil, err
	}
	for _, s := range steps {
		if s.Part >= part {
			c.Next = append(c.Next, s)
		}
	}
	if c.Raised, err = StepPart(c.Newest, c.Version); err != nil {
		return nil, err
	}
	return c, nil
}
