// Package moduleindex builds the module index of a release from the previous index, the release's
// packages, its tag and its withdrawals, refusing by the rules module-index.*; reads an index of
// any version of the format by the places every version keeps; and picks the version of a module
// for a new attempt: the newest by semver that is not withdrawn and lies in the format's window
// (format/README.md, sections 2 and 26; R63 spec; R16). The CLI's index and open call it, and api.
//
// Build takes a Release — the previous index as its file, or First, the packages that package
// modulepackage built or read, the tag and the withdrawals — and gives module-index.json, the
// canonical text of the index from internal/canonical, or a *Refusal that names every breach of the
// rules of the index. Read reads an index by its members at their places alone, gives every entry
// of whatever format version, and takes a version as withdrawn whatever its reason. Newest picks
// the version of a new attempt, skipping an entry whose format version module-kit does not read.
// WithLatest marks in the index of a release the withdrawals that the index of the newest release
// knows, for the CLI's open (R15; R43). Versions are ordered by internal/semver. It checks no
// signature: spec's release-verify does, where a file enters NovaTrail (R77 spec; R35).
package moduleindex
