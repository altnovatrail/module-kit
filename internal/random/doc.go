// Package random is the generator of the format and its places of randomness, written from the
// constants of the format's text rather than taken from math/rand/v2, with an error on an argument
// outside a function's domain (format/README.md, section 22; R59 spec; R10). Nothing outside
// module-kit imports it: what goes out is the instance of an attempt, from package template.
//
// A Stream is the generator, started from two words or from a place: a seed and its names. It
// gives numbers through five functions, those the vectors of the generator check, and through
// nothing else: no shuffle of a list, which module-kit does not do. The generator keeps no state of
// its own: a stream lives in the one call that builds it. math/rand/v2 and hash/fnv are the
// reference in its tests alone, and no randomness comes from anything but a seed.
package random
