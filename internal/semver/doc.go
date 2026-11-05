// Package semver reads a version of a module, X.Y.Z, and orders two of them by Semantic Versioning
// 2.0.0, as the format writes a version: three decimal numerals with no leading zero, and no
// pre-release and no build metadata (format/README.md, sections 26 and 27; R46 spec). Package
// moduleindex orders the versions of a module by it, for the rule module-index.version-above and
// the version of a new attempt, and package module the format versions, MAJOR.MINOR, as the
// versions MAJOR.MINOR.0; package moduleversion
// orders the published versions of a module by it and raises a version to the next by one of its
// parts, Raise (R35, R36).
//
// A numeral has no bound, as a version has none, so a Version keeps its numerals as text and orders
// them by their length first: no version is too large to read or to order.
package semver
