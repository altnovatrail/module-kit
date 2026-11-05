// Package canonical writes the canonical text of a JSON value — of a package, a module index and an
// open view — and stops with an error on a value the format cannot hold (format/README.md,
// section 25; R09). Only it writes those texts: not encoding/json, not jsontext and not the types
// of formatv0.
//
// Text writes the canonical text of a tree of JSON values, and Digest and Descriptor name bytes as
// a package does, by sha256: and their size (R34). Parse reads a JSON text into such a tree, an
// integer as an int64, for the readers of a package and of the module index (R35). Its fuzz test
// holds Text to jsontext's Canonicalize, which lives in that test alone, in a file built only under
// the experiment jsonv2.
package canonical
