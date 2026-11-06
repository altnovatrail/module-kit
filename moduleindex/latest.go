package moduleindex

import (
	"errors"
	"fmt"
	"maps"
	"reflect"
	"slices"

	"github.com/altnovatrail/spec/types/formatv0"
)

// ErrLatest is an index given as the newest that is not an index of a release at or after the one
// it is given with: it lacks an entry of that release's index, holds one otherwise than it was, or
// does not keep its withdrawal.
var ErrLatest = errors.New("not an index of a release at or after the release's")

// WithLatest gives the index of a release with the withdrawals that the index of the newest release
// marks: each entry of release as it is, but withdrawn whenever latest withdraws it, by latest's
// reason. Only the newest index knows of every withdrawal so far, and a version that it withdraws
// is no version to show, whatever the release that published it says (format/README.md, section 26;
// R25 spec, R63 spec). Newest of the index it gives is then the version that the page of samples of
// canvas takes of a module, as the CLI's open does (R15).
//
// Every later index keeps every entry as it was, and a withdrawal with its reason
// (module-index.entries-kept, module-index.known-withdrawals). So an entry of release that latest
// lacks, holds otherwise or does not keep withdrawn gives ErrLatest, and so latest from a release
// before release's does too; an entry of latest that release lacks, of a later release, is left
// out. latest may be release itself. Neither index is changed.
func WithLatest(release, latest formatv0.ModuleIndex) (formatv0.ModuleIndex, error) {
	merged := formatv0.ModuleIndex{
		ModuleFormatVersion: release.ModuleFormatVersion,
		Modules:             make(map[formatv0.ModuleID]map[formatv0.ModuleVersion]formatv0.Entry, len(release.Modules)),
	}
	for _, id := range slices.Sorted(maps.Keys(release.Modules)) {
		versions := release.Modules[id]
		merged.Modules[id] = make(map[formatv0.ModuleVersion]formatv0.Entry, len(versions))
		for _, version := range slices.Sorted(maps.Keys(versions)) {
			entry := versions[version]
			later, ok := latest.Modules[id][version]
			if !ok {
				return formatv0.ModuleIndex{}, fmt.Errorf("moduleindex: the newest index has no %s %s: %w", id, version, ErrLatest)
			}
			if !kept(entry, later) {
				return formatv0.ModuleIndex{}, fmt.Errorf("moduleindex: the newest index holds %s %s otherwise than the release's: %w", id, version, ErrLatest)
			}
			entry.Withdrawn = later.Withdrawn
			merged.Modules[id][version] = entry
		}
	}
	return merged, nil
}

// kept says whether a later index holds an entry as it was: every member the same, but that a
// version not yet withdrawn may since have been.
func kept(entry, later formatv0.Entry) bool {
	if entry.Withdrawn != nil && (later.Withdrawn == nil || *later.Withdrawn != *entry.Withdrawn) {
		return false
	}
	entry.Withdrawn, later.Withdrawn = nil, nil
	return reflect.DeepEqual(entry, later)
}
