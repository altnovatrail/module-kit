package moduleversion

import (
	"reflect"
	"slices"
	"testing"

	"github.com/altnovatrail/spec/types/formatv0"

	"github.com/altnovatrail/module-kit/verdict"
)

// What decides the verdict of a choice is its notation too, options when absent, and its options as
// a choice of that notation normalizes them; of a time, its key by its hours and its minutes, and its
// trap values so (format/README.md, section 27).
func TestMeaningOfNotationAndTime(t *testing.T) {
	choice := func(notation *formatv0.CheckNotation, options ...formatv0.Text) verdict.AnswerSlot {
		return verdict.AnswerSlot{Check: formatv0.Check{Kind: formatv0.CheckKindChoice, Key: new("Kim"), Options: options, Notation: notation}}
	}
	if !reflect.DeepEqual(meaningOf(choice(nil, "Kim", "Lea")), meaningOf(choice(new(formatv0.CheckNotationOptions), "Kim", "Lea"))) {
		t.Error("a choice of options means otherwise with its notation written")
	}
	word := meaningOf(choice(new(formatv0.CheckNotationWord), "Kim.", "Lea"))
	if word.notation != "word" || !slices.Equal(word.options, []string{"kim", "lea"}) ||
		reflect.DeepEqual(word, meaningOf(choice(nil, "Kim.", "Lea"))) {
		t.Errorf("a choice of notation word means %+v", word)
	}
	timeOf := func(key, trap string) meaning {
		return meaningOf(verdict.AnswerSlot{Check: formatv0.Check{Kind: formatv0.CheckKindTime, Key: new(key),
			TrapValues: []formatv0.TrapValue{{Value: trap, Trap: "hundred-minutes"}}}})
	}
	if a, b := timeOf("7:30", "06:70"), timeOf("7:30", "6:70"); !reflect.DeepEqual(a, b) || a.key != "7:30" || a.traps[0][0] != "6:70" {
		t.Errorf("a time means %+v and %+v", a, b)
	}
}
