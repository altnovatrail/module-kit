package moduleversion

import (
	"fmt"

	"github.com/altnovatrail/module-kit/internal/semver"
)

// A Part is a part of a version of a module, as the comparison finds it and a step to the next
// version raises it, from the lowest (format/README.md, section 27): None, when nothing but the
// version changes; Patch, the same entries, checks and path in other words or pictures; Minor,
// everything there stays and means the same, and the module gives more or leads otherwise; and
// Major, something of what a learner had is gone or means something else. It is written by its word,
// as the test vectors of versions write it.
type Part int

// The parts of a version, in their order.
const (
	None Part = iota
	Patch
	Minor
	Major
)

var partWords = [...]string{"none", "patch", "minor", "major"}

func (p Part) String() string {
	if p < None || p > Major {
		return fmt.Sprintf("Part(%d)", int(p))
	}
	return partWords[p]
}

// MarshalText writes the part by its word, as --json of diff gives it.
func (p Part) MarshalText() ([]byte, error) {
	if p < None || p > Major {
		return nil, fmt.Errorf("moduleversion: %v is no part of a version", p)
	}
	return []byte(p.String()), nil
}

// A Step is a next version after a version of a module, and the part that the step to it raises.
type Step struct {
	Version string `json:"version"`
	Part    Part   `json:"part"`
}

// Steps gives the next versions after a version of a module, each by one of its parts, and the part
// that each step raises, in the order patch, minor and major of the number it raises
// (format/README.md, section 27). From X.Y.Z with X from 1 on, X.Y.(Z+1) is a patch, X.(Y+1).0 a
// minor and (X+1).0.0 a major. While X is 0, y plays the major, as Cargo reads it: from 0.y.z with y
// from 1 on, 0.y.(z+1) is a minor, which serves a patch too, and 0.(y+1).0 and 1.0.0 are majors; and
// from 0.0.z, 0.0.(z+1), 0.1.0 and 1.0.0 are each a major.
func Steps(version string) ([]Step, error) {
	v, err := semver.Parse(version)
	if err != nil {
		return nil, fmt.Errorf("moduleversion: %w", err)
	}
	parts := [3]Part{Patch, Minor, Major}
	switch {
	case !v.Zero(semver.Major):
	case !v.Zero(semver.Minor):
		parts = [3]Part{Minor, Major, Major}
	default:
		parts = [3]Part{Major, Major, Major}
	}
	return []Step{
		{v.Raise(semver.Patch).String(), parts[0]},
		{v.Raise(semver.Minor).String(), parts[1]},
		{v.Raise(semver.Major).String(), parts[2]},
	}, nil
}

// StepPart gives the part that the step from one version of a module to another raises, or None
// when to is not the next version after from by one of its parts (format/README.md, section 27).
func StepPart(from, to string) (Part, error) {
	steps, err := Steps(from)
	if err != nil {
		return None, err
	}
	next, err := semver.Parse(to)
	if err != nil {
		return None, fmt.Errorf("moduleversion: %w", err)
	}
	for _, s := range steps {
		if s.Version == next.String() {
			return s.Part, nil
		}
	}
	return None, nil
}

// SameMajor says whether two versions of a module are of the same major, as diff reads it and api
// does when it moves a learner to a newer version of a module (format/README.md, section 27): one X
// from 1 on, one 0.y for y from 1 on, and each 0.0.z alone.
func SameMajor(a, b string) (bool, error) {
	v, err := semver.Parse(a)
	if err != nil {
		return false, fmt.Errorf("moduleversion: %w", err)
	}
	w, err := semver.Parse(b)
	if err != nil {
		return false, fmt.Errorf("moduleversion: %w", err)
	}
	// The next major after a version depends on X alone, and the next minor on X and Y alone.
	switch {
	case !v.Zero(semver.Major):
		return v.Raise(semver.Major) == w.Raise(semver.Major), nil
	case !v.Zero(semver.Minor):
		return v.Raise(semver.Minor) == w.Raise(semver.Minor), nil
	}
	return v == w, nil
}

// UnmarshalText reads a part by its word, as MarshalText writes it.
func (p *Part) UnmarshalText(text []byte) error {
	for i, word := range partWords {
		if string(text) == word {
			*p = Part(i)
			return nil
		}
	}
	return fmt.Errorf("moduleversion: %q is no part of a version", text)
}
