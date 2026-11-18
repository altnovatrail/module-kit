package normal

import (
	"slices"
	"strings"
	"unicode"
	"unicode/utf8"

	"golang.org/x/text/unicode/norm"
)

// NFKC is golang.org/x/text/unicode/norm, which follows the Stream-Safe Text Process of UAX #15 too:
// after 30 non-starters in a row it puts U+034F in, and spec's second implementation does not (R27).

// UnicodeVersion is the version of Unicode of the tables of NFKC that the package is built with.
const UnicodeVersion = norm.Version

// Scripts takes step 1: every superscript or subscript digit becomes U+FFFD, which no notation
// takes, and so does every vulgar fraction that a number, a character of the general category N,
// stands right before or right after in the text as it is, so that NFKC does not make 7² into 72 or
// 1½ into 11⁄2.
func Scripts(text string) string {
	runes := []rune(text)
	out := slices.Clone(runes)
	isNumber := func(i int) bool { return i >= 0 && i < len(runes) && unicode.Is(unicode.N, runes[i]) }
	for i, r := range runes {
		if scriptDigit(r) || vulgarFraction(r) && (isNumber(i-1) || isNumber(i+1)) {
			out[i] = utf8.RuneError
		}
	}
	return string(out)
}

func scriptDigit(r rune) bool {
	return r == '\u00B2' || r == '\u00B3' || r == '\u00B9' || '\u2070' <= r && r <= '\u209F'
}

func vulgarFraction(r rune) bool {
	return '\u00BC' <= r && r <= '\u00BE' || '\u2150' <= r && r <= '\u215F' || r == '\u2189'
}

// Text takes steps 2 to 4, in their order, so that steps 3 and 4 take what NFKC gives too: NFKC;
// the hyphens and dashes and the minus sign become -, the fraction and division slashes /; and every
// run of White_Space becomes one space, with none at either end.
func Text(text string) string {
	var b strings.Builder
	space := false
	for _, r := range norm.NFKC.String(text) {
		switch {
		case unicode.Is(unicode.White_Space, r):
			space = true
			continue
		case r == '\u2010' || '\u2012' <= r && r <= '\u2015' || r == '\u2212':
			r = '-'
		case r == '\u2044' || r == '\u2215':
			r = '/'
		}
		if space && b.Len() > 0 {
			b.WriteByte(' ')
		}
		space = false
		b.WriteRune(r)
	}
	return b.String()
}

// Lower maps each character to its simple lowercase mapping (UnicodeData.txt, field 13), as
// unicode.ToLower does, a character at a time: no context counts, as a final sigma's would, and no
// character becomes two.
func Lower(text string) string {
	return strings.Map(unicode.ToLower, text)
}

// upToStep5 takes steps 1 to 5, which a number's value and its units both go through. The spaces
// it leaves at either end stay until step 7.
func upToStep5(text string) string {
	return strings.TrimRight(strings.TrimPrefix(Text(Scripts(text)), "="), ".!?")
}

// trimmed takes a text that steps 1 to 4 normalized, where white space is single spaces:
// strings.TrimSpace would drop other characters.
func trimmed(text string) string {
	return strings.Trim(text, " ")
}

// Number gives a value as number, fraction and predicate read it, steps 1 to 7, with units, the
// forms of a number's units that Unit gives: "" when it is blank.
func Number(value string, units []string) string {
	return trimmed(withoutUnit(upToStep5(value), units))
}

// Unit gives what a unit compares by: steps 1 to 5, as the value goes through them, the spaces at
// either end dropped, and each character in lowercase (step 6).
func Unit(unit string) string {
	return Lower(trimmed(upToStep5(unit)))
}

// withoutUnit takes step 6: of the units that end the text, the longest is dropped, with a space
// before it, unless nothing but spaces stands before it. The lengths are counted in characters: a
// character and its lowercase can differ in bytes, as İ and i do.
func withoutUnit(text string, units []string) string {
	runes := []rune(text)
	lowered := []rune(Lower(text))
	longest := 0
	for _, unit := range units {
		u := []rune(unit)
		if len(u) > longest && len(u) <= len(lowered) && slices.Equal(lowered[len(lowered)-len(u):], u) {
			longest = len(u)
		}
	}
	if longest == 0 {
		return text
	}
	before := string(runes[:len(runes)-longest])
	if trimmed(before) == "" {
		return text
	}
	return strings.TrimSuffix(before, " ")
}

// Choice gives what a choice compares: the text through steps 2 to 4, then in lowercase. Nothing
// more is dropped: a = or a . can be what the learner chose.
func Choice(text string) string {
	return Lower(Text(text))
}
