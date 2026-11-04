// Package canonical writes the canonical text of a JSON value — of a package, a module index and an
// open view — and stops with an error on a value the format cannot hold (format/README.md,
// section 25; R09). Only it writes those texts: not encoding/json, not jsontext and not the types
// of formatv0.
//
// It holds no code yet: the epic of code brings it (R16).
package canonical
