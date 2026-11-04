// Package random is the generator of the format and its places of randomness, written from the
// constants of the format's text rather than taken from math/rand/v2, with an error on an argument
// outside a function's domain (format/README.md, section 22; R59 spec; R10). Nothing outside
// module-kit imports it: what goes out is the instance of an attempt, from package template.
//
// It holds no code yet: the epic of code brings it (R16).
package random
