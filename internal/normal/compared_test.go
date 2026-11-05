package normal

import (
	"testing"

	"github.com/altnovatrail/spec/types/formatv0"
)

// Compared gives one form to the values that a check compares equal, and none to a value out of a
// number's notation; Units drops a unit with no form and gives none to a check of another kind.
func TestCompared(t *testing.T) {
	number := formatv0.Check{Kind: formatv0.CheckKindNumber, Units: []formatv0.Text{" Rungs ", "=", "m"}}
	if got := Units(number); len(got) != 2 || got[0] != "rungs" || got[1] != "m" {
		t.Errorf("Units(%v) = %q", number.Units, got)
	}
	if got := Units(formatv0.Check{Kind: formatv0.CheckKindChoice, Units: []formatv0.Text{"m"}}); got != nil {
		t.Errorf("Units of a choice = %q", got)
	}
	fraction := formatv0.Check{Kind: formatv0.CheckKindFraction}
	choice := formatv0.Check{Kind: formatv0.CheckKindChoice}
	rubric := formatv0.Check{Kind: formatv0.CheckKindRubric}
	for _, c := range []struct {
		check      formatv0.Check
		text, want string
		ok         bool
	}{
		{number, "2.50", "5/2", true}, {number, "5/2 rungs", "5/2", true}, {number, "7", "7", true}, {number, "seven", "", false},
		{fraction, "6/8", "3/4", true}, {fraction, "0.75", "", false},
		{choice, " Yes ", "yes", true}, {choice, "", "", true}, {rubric, "7", "", false},
	} {
		got, ok := Compared(c.check, Units(c.check), c.text)
		if got != c.want || ok != c.ok {
			t.Errorf("Compared(%s, %q) = %q, %t, want %q, %t", c.check.Kind, c.text, got, ok, c.want, c.ok)
		}
	}
}
