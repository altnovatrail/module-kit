// Package moduleindex builds the module index of a release from the previous index, the release's
// packages, its tag and its withdrawals, refusing by the rules module-index.*; reads an index of
// any version of the format by the places every version keeps; and picks the version of a module
// for a new attempt: the newest by semver that is not withdrawn and lies in the format's window
// (format/README.md, sections 2 and 26; R63 spec; R16). The CLI's index and open call it, and api.
//
// It holds no code yet: the epic of code brings it (R16).
package moduleindex
