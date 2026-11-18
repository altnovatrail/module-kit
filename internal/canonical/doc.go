// Package canonical writes the canonical text of a JSON value — of a package, a module index and an
// open view — and stops with an error on a value the format cannot hold (R09). Only it writes those
// texts: not encoding/json, not jsontext and not the types of formatv0.
//
// Digest and Descriptor name bytes as a package does, by sha256: and their size (R34). Parse reads
// a JSON text into the tree Text takes, an integer as an int64 (R35). The fuzz test holds Text to
// jsontext's Canonicalize, which only that test imports, in a file built only under the experiment
// jsonv2.
package canonical
