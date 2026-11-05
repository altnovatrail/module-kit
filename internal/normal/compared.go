package normal

import "github.com/altnovatrail/spec/types/formatv0"

// Units gives the forms of the units of a number, Unit of each, those that are not empty: a check of
// another kind has none (format/README.md, section 15).
func Units(c formatv0.Check) []string {
	if c.Kind != formatv0.CheckKindNumber {
		return nil
	}
	var forms []string
	for _, unit := range c.Units {
		if form := Unit(string(unit)); form != "" {
			forms = append(forms, form)
		}
	}
	return forms
}

// Compared gives what a check compares a value by, and whether it compares it at all: a choice, its
// normal form, always; a number, a fraction and a predicate, its value in the kind's notation,
// exactly, written in lowest terms. units are the forms of a number's units, as Units gives them.
// The verdict compares a value with the key and the trap values by it, and the comparison of the
// versions of a module the keys and the trap values of two versions (format/README.md, sections 15
// and 27).
func Compared(c formatv0.Check, units []string, text string) (string, bool) {
	switch c.Kind {
	case formatv0.CheckKindChoice:
		return Choice(text), true
	case formatv0.CheckKindNumber, formatv0.CheckKindFraction, formatv0.CheckKindPredicate:
		if w, ok := Notation(Number(text, units), c.Kind); ok {
			return w.Value.RatString(), true
		}
	}
	return "", false
}
