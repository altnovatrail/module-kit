// Package moduleversion compares a new version of a module with every published one: it finds the
// part of the version that the change asks for and holds the rules module-version.*, and it reads
// a package of an earlier version of the format by the places every version keeps
// (format/README.md, sections 2 and 27; R89 spec, R101 spec; R13). The CLI's diff calls it, and
// module-registry from stage 4.
//
// It holds no code yet: the epic of code brings it (R16).
package moduleversion
