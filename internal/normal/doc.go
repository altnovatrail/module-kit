// Package normal normalizes a text as the check kinds of the module format and its leak check do,
// and reads a number in the notation that a kind takes: the steps 1 to 7 of a number and the units
// of step 6, the normal form of a choice of either notation, a time by its hours and minutes, and a
// number in notation as an exact rational. The verdict and the leak check share it, so that both
// read a text alike.
//
// Its NFKC is golang.org/x/text/unicode/norm, and the version of Unicode that it follows is
// UnicodeVersion, which package verdict names for a release.
package normal
