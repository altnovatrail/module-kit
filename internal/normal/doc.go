// Package normal normalizes a text as the check kinds of the module format and its leak check do,
// and reads a number in the notation that a kind takes (format/README.md, sections 15 and 20): the
// steps 1 to 7 of a number and the units of step 6, the normal form of a choice of either notation,
// a time by its hours and minutes, and a number in notation as an exact rational. The verdict reads
// the values read by it, and the leak check the texts it searches, the keys and the leak phrases; it
// came out of package verdict when the leak check came to need it too (R27, R30).
//
// Its NFKC is golang.org/x/text/unicode/norm, and the version of Unicode that it follows is
// UnicodeVersion, which package verdict names for a release.
package normal
