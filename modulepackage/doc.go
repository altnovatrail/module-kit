// Package modulepackage builds the package of a checked module — its canonical view, the listing
// of its device part and its inventory, with their digests — and reads a package as files by the
// chain of digests from the digest an index gives: inventory, then parts, then files
// (format/README.md, section 25; R62 spec). It checks no signature: spec's release-verify does,
// where a file enters NovaTrail (R77 spec). The CLI's build, api, tutor's eval runner and
// module-registry call it.
//
// It holds no code yet: the epic of code brings it (R16).
package modulepackage
