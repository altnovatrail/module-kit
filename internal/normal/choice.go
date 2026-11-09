package normal

import (
	"strings"
	"unicode"

	"github.com/altnovatrail/spec/types/formatv0"
)

// ChoiceOf gives what a choice of the check compares: Choice of the text, and, with notation: word,
// every ., !, ? and space at its end dropped as well, so that Kim. is kim and ? is blank
// (format/README.md, section 15). A choice by tapping has no notation, and is a choice of options.
func ChoiceOf(c formatv0.Check, text string) string {
	form := Choice(text)
	if WordNotation(c) {
		form = strings.TrimRight(form, ".!? ")
	}
	return form
}

// WordNotation says whether a choice is of the notation word, which takes any one word of letters
// besides its options: notation is options unless given (format/README.md, section 15).
func WordNotation(c formatv0.Check) bool {
	return c.Kind == formatv0.CheckKindChoice && c.Notation != nil && *c.Notation == formatv0.CheckNotationWord
}

// Word says whether a text is one word of letters: a letter, a character of the general category L,
// then letters and marks, M, as the vowel signs of a word in Devanagari, and nothing else
// (UnicodeData.txt, field 2). An empty text is none.
func Word(text string) bool {
	for i, r := range text {
		switch {
		case unicode.IsLetter(r):
		case i > 0 && unicode.IsMark(r):
		default:
			return false
		}
	}
	return text != ""
}

// Search gives the normal form in which the leak check searches a text: step 1 of section 15, then
// Choice — NFKC, the dashes and slashes of step 3, the white space of step 4, and each character in
// lowercase (format/README.md, section 20). Package leak searches by it, and the check of a module
// compares by it the text pieces of a line of a worked example with those of a task's line.
func Search(text string) string {
	return Choice(Scripts(text))
}
