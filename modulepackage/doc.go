// Package modulepackage builds the package of a checked module — its canonical view, the listing
// of its device part and its inventory, with their digests — and reads a package as files by the
// chain of digests from the digest an index gives: inventory, then parts, then files
// (format/README.md, section 25; R62 spec). It checks no signature: spec's release-verify does,
// where a file enters NovaTrail (R77 spec). The CLI's build, api, tutor's eval runner and
// module-registry call it.
//
// Build checks a module by package module and builds its package — the canonical view, with the
// tiers of its skill graph and the numbers of every instance of its templates with the results of
// their solvers, the listing of the device part and the inventory, each by its canonical text from
// internal/canonical — as bytes of its files; the CLI writes them to a folder. Read takes the files
// of a package as an fs.FS and the digest of the package, and reads the inventory by that digest,
// then the two parts, then the files of device/, each by the digest and size that name it, held to
// its schema and its canonical text (R34).
package modulepackage
