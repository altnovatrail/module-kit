package leak

import (
	"errors"
	"math/big"
	"slices"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/altnovatrail/spec/types/formatv0"

	"github.com/altnovatrail/module-kit/internal/normal"
	"github.com/altnovatrail/module-kit/verdict"
)

// Kind is what a leak is: a key of an answer slot, or a leak phrase of the task.
type Kind string

// The kinds of a leak.
const (
	Key    Kind = "key"    // the key of a check of a number, a fraction or a choice
	Phrase Kind = "phrase" // a phrase of the task's leak_phrases
)

// Leak is a key or a leak phrase of a task that the leak check finds in a text, Written as the task
// writes it. A key is of the answer slot Slot of the item Item, both from 0, the answer slots in the
// order of verdict.AnswerSlots; a leak phrase is the one at Phrase in the task's leak_phrases, from
// 0. The fields of the other kind are 0.
type Leak struct {
	Kind    Kind
	Written string
	Item    int
	Slot    int
	Phrase  int
}

// ErrTemplate is what a template gives: the leak check searches it through its instances, each a
// task with the numbers of the instance in place of its placeholders (format/README.md, sections 20
// and 23).
var ErrTemplate = errors.New("leak: a template is searched through its instances, and an instance is a task with its numbers in place")

// Search is what the leak check searches a text for, for one task (format/README.md, section 20):
// the key of every check of a number, a fraction and a choice, in every item, but a key that the
// same search finds in a text that the learner sees with its item; and every leak phrase. It holds
// nothing that changes, so one Search searches texts from many goroutines at once.
type Search struct {
	needles []needle
}

// A needle is a key or a leak phrase that a search looks for: a key of a number or a fraction by its
// magnitude, and a key of a choice and a leak phrase by its normal form.
type needle struct {
	leak      Leak
	magnitude *big.Rat
	form      string
}

// For gives the search of a task. A template gives ErrTemplate: api searches the words of the tutor
// on an attempt for the keys and the leak phrases of its instance, and the check of a module
// searches a template through each instance. A key of a number or a fraction that is not in the
// notation of its kind is not searched for, as spec's second implementation has it: a key that holds
// a placeholder outside a template breaks task.placeholders instead (section 23).
func For(task formatv0.Task) (*Search, error) {
	if task.Template != nil {
		return nil, ErrTemplate
	}
	s := &Search{}
	statement := statementTexts(task.Statement)
	for i, item := range task.Items {
		slots := verdict.AnswerSlots(item)
		own := append(slices.Clone(statement), itemTexts(item, slots)...)
		for j, slot := range slots {
			n, ok := keyOf(slot.Check)
			if !ok || slices.ContainsFunc(own, n.in) {
				continue
			}
			n.leak = Leak{Kind: Key, Written: *slot.Check.Key, Item: i, Slot: j}
			s.needles = append(s.needles, n)
		}
	}
	for k, phrase := range task.LeakPhrases {
		s.needles = append(s.needles, needle{leak: Leak{Kind: Phrase, Written: string(phrase), Phrase: k}, form: form(string(phrase))})
	}
	return s, nil
}

// In gives the keys and the leak phrases of the task that the leak check finds in a text, as the
// learner would see it: the keys by their items and answer slots, and then the leak phrases, in the
// order of the task. A key that two answer slots share is found for each. A byte of the text that
// is not UTF-8 counts as U+FFFD, which is no character of a word.
func (s *Search) In(text string) []Leak {
	t := searchedOf(text)
	var found []Leak
	for _, n := range s.needles {
		if n.in(t) {
			found = append(found, n.leak)
		}
	}
	return found
}

// Numbers gives the numbers that the leak check reads in a text, each as it stands in the text's
// normal form, in the order of the text: a key of a number or a fraction is found where one of them
// has its value (format/README.md, section 20). A text of the whole module SHOULD hold none, and the
// check of a module warns of each (R50). A byte of the text that is not UTF-8 counts as U+FFFD, as
// in Search.In.
func Numbers(text string) []string {
	runes := []rune(form(text))
	var written []string
	for _, r := range runsIn(runes) {
		written = append(written, string(runes[r.start:r.end]))
	}
	return written
}

// keyOf gives the needle of the key of a check, and whether the leak check searches for it: a key
// of a number or a fraction in the notation of its kind, by its magnitude, and a key of a choice, by
// its normal form. A predicate has no key, and a rubric neither.
func keyOf(c formatv0.Check) (needle, bool) {
	if c.Key == nil {
		return needle{}, false
	}
	switch c.Kind {
	case formatv0.CheckKindNumber, formatv0.CheckKindFraction:
		w, ok := normal.Notation(*c.Key, c.Kind)
		if !ok {
			return needle{}, false
		}
		return needle{magnitude: new(big.Rat).Abs(w.Value)}, true
	case formatv0.CheckKindChoice:
		return needle{form: form(*c.Key)}, true
	}
	return needle{}, false
}

// statementTexts gives, searched, the texts of a statement that the learner sees with every item:
// its paragraphs and the alt of its pictures.
func statementTexts(statement []formatv0.StatementPart) []*searched {
	var texts []*searched
	for _, part := range statement {
		if part.Text != nil {
			texts = append(texts, searchedOf(string(*part.Text)))
		}
		if part.Alt != nil {
			texts = append(texts, searchedOf(*part.Alt))
		}
	}
	return texts
}

// itemTexts gives, searched, the texts that the learner sees with an item beside its statement: the
// item's text, the text pieces of its line, and the options of its answer slots bound choice. The
// options of a choice written by hand are the check's, closed, and no text the learner sees.
func itemTexts(item formatv0.Item, slots []verdict.AnswerSlot) []*searched {
	var texts []*searched
	if item.Text != nil {
		texts = append(texts, searchedOf(string(*item.Text)))
	}
	for _, part := range item.Line {
		if part.Text != nil {
			texts = append(texts, searchedOf(string(*part.Text)))
		}
	}
	for _, slot := range slots {
		for _, option := range slot.Options {
			texts = append(texts, searchedOf(string(option)))
		}
	}
	return texts
}

// in says whether a needle is found in a text: a key of a number or a fraction where a number of the
// text has its magnitude, and a key of a choice or a leak phrase where its normal form stands.
func (n needle) in(t *searched) bool {
	if n.magnitude != nil {
		return slices.ContainsFunc(t.numbers(), func(v *big.Rat) bool { return v.Cmp(n.magnitude) == 0 })
	}
	return holds(t.form, n.form)
}

// form gives the normal form of a text, of a key of a choice and of a leak phrase: step 1 of section
// 15, then NFKC, the dashes and slashes of step 3 and the white space of step 4, and then each
// character to its simple lowercase mapping (format/README.md, section 20).
func form(text string) string {
	return normal.Choice(normal.Scripts(text))
}

// A searched text: its normal form, and the numbers in it, read once, when a key of a number or a
// fraction first asks for them: a search with no such key reads none.
type searched struct {
	form  string
	found []*big.Rat
	read  bool
}

func searchedOf(text string) *searched {
	return &searched{form: form(text)}
}

// numbers gives the numbers of the text, reading them the first time.
func (t *searched) numbers() []*big.Rat {
	if !t.read {
		t.found, t.read = numbersIn([]rune(t.form)), true
	}
	return t.found
}

// isWord says whether a character is a character of a word: one of the general categories L, M and
// N, a letter, a mark or a number (format/README.md, section 20). U+FFFD is none, and so is the
// utf8.RuneError that decoding gives at either end of a text.
func isWord(r rune) bool {
	return unicode.In(r, unicode.L, unicode.M, unicode.N)
}

// holds says whether a phrase in its normal form — a key of a choice or a leak phrase — is found in
// a text's normal form: where it stands, with no character of a word right outside an end of it
// that is a character of a word itself. A phrase whose normal form is empty is found nowhere. Every
// place where it stands is tried, so that "odd" is found in "oddly odd". Both forms are UTF-8, so a
// match of their bytes begins and ends at characters.
func holds(text, phrase string) bool {
	if phrase == "" {
		return false
	}
	first, _ := utf8.DecodeRuneInString(phrase)
	last, _ := utf8.DecodeLastRuneInString(phrase)
	opens, closes := isWord(first), isWord(last)
	for from := 0; from < len(text); {
		i := strings.Index(text[from:], phrase)
		if i < 0 {
			return false
		}
		at, end := from+i, from+i+len(phrase)
		before, _ := utf8.DecodeLastRuneInString(text[:at])
		after, _ := utf8.DecodeRuneInString(text[end:])
		if !(opens && isWord(before)) && !(closes && isWord(after)) {
			return true
		}
		_, size := utf8.DecodeRuneInString(text[at:])
		from = at + size
	}
	return false
}

// isDigit says whether a character is a digit, 0 to 9: a number of the text begins at one, and the
// digits of other scripts begin none (format/README.md, section 20).
func isDigit(r rune) bool {
	return '0' <= r && r <= '9'
}

// wordAt says whether a character of a word stands at i, inside the text.
func wordAt(runes []rune, i int) bool {
	return i >= 0 && i < len(runes) && isWord(runes[i])
}

// numbersIn gives the numbers of a text in its normal form, each by its value, from its start
// (format/README.md, section 20): at a digit, the longest run of the form of a number, which is a
// number of the text when no character of a word stands right before it and none right after it;
// either way the search goes on after the whole run, so that no tail of a longer token, as the 5 of
// x7.5, is read as a number. A number is never negative: no sign is read, and a key is compared by
// its magnitude.
func numbersIn(runes []rune) []*big.Rat {
	var found []*big.Rat
	for _, r := range runsIn(runes) {
		found = append(found, r.value)
	}
	return found
}

// A number of a text in its normal form: where its run begins and ends, and its value.
type run struct {
	start, end int
	value      *big.Rat
}

// runsIn gives the numbers of a text in its normal form with their places, as numbersIn reads them.
func runsIn(runes []rune) []run {
	var found []run
	for at := 0; at < len(runes); {
		if !isDigit(runes[at]) {
			at++
			continue
		}
		end, value := longestRun(runes, at)
		if !wordAt(runes, at-1) && !wordAt(runes, end) {
			found = append(found, run{at, end, value})
		}
		at = end
	}
	return found
}

// digitsEnd gives where the digits from at end.
func digitsEnd(runes []rune, at int) int {
	for at < len(runes) && isDigit(runes[at]) {
		at++
	}
	return at
}

// integerOf gives the integer that digits write, leading zeros and all.
func integerOf(digits []rune) *big.Int {
	n, _ := new(big.Int).SetString(string(digits), 10)
	return n
}

// longestRun gives where the longest run of the form of a number that begins at a digit ends, and
// its value, the exact rational it writes: a simple fraction, digits, an optional space, /, an
// optional space and digits that are not all zero; a decimal, digits, . and digits; or an integer,
// digits. A fraction and a decimal never both begin at one digit, since the digits of the whole are
// followed by . or by what a fraction takes.
func longestRun(runes []rune, at int) (int, *big.Rat) {
	whole := digitsEnd(runes, at)
	end, value := whole, new(big.Rat).SetInt(integerOf(runes[at:whole]))
	if whole+1 < len(runes) && runes[whole] == '.' && isDigit(runes[whole+1]) {
		end = digitsEnd(runes, whole+1)
		places := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(end-whole-1)), nil)
		value = new(big.Rat).SetFrac(integerOf(slices.Concat(runes[at:whole], runes[whole+1:end])), places)
	}
	slash := whole
	if slash < len(runes) && runes[slash] == ' ' {
		slash++
	}
	if slash < len(runes) && runes[slash] == '/' {
		start := slash + 1
		if start < len(runes) && runes[start] == ' ' {
			start++
		}
		if stop := digitsEnd(runes, start); stop > start {
			if denominator := integerOf(runes[start:stop]); denominator.Sign() != 0 {
				end, value = stop, new(big.Rat).SetFrac(integerOf(runes[at:whole]), denominator)
			}
		}
	}
	return end, value
}
