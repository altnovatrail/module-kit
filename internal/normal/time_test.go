package normal

import (
	"testing"

	"github.com/altnovatrail/spec/types/formatv0"
)

// A time is read in steps 1 to 5 of a number, and is in the notation of a time read when it is H:MM
// or HH:MM, its minutes two digits or three that begin with no zero; it compares by its hours and its
// minutes, each as a number, and never by their sum (format/README.md, section 15).
func TestTimeNotation(t *testing.T) {
	for _, c := range []struct {
		value, want string
		ok          bool
	}{
		{"7:30", "7:30", true},
		{" 07:30. ", "7:30", true},
		{"=19:05!", "19:05", true},
		{"5:60", "5:60", true},
		{"10:100", "10:100", true},
		{"７:３０", "7:30", true},
		{"7:3", "", false},
		{"7:030", "", false},
		{"007:30", "", false},
		{"7.30", "", false},
		{"7 : 30", "", false},
		{"7:30 pm", "", false},
		{"", "", false},
	} {
		clock, ok := TimeNotation(Time(c.value))
		if ok != c.ok || ok && clock.String() != c.want {
			t.Errorf("the time %q reads %v, %v; want %q, %v", c.value, clock, ok, c.want, c.ok)
		}
	}
}

// With notation: word a choice drops the marks and spaces at its end too, and a word of letters is a
// letter, then letters and marks; a choice of options, and one by tapping, drops nothing more.
func TestChoiceOfAWord(t *testing.T) {
	word := formatv0.Check{Kind: formatv0.CheckKindChoice, Notation: new(formatv0.CheckNotationWord)}
	options := formatv0.Check{Kind: formatv0.CheckKindChoice, Notation: new(formatv0.CheckNotationOptions)}
	for _, c := range []struct {
		check formatv0.Check
		text  string
		want  string
	}{
		{word, "Kim.", "kim"},
		{word, "Kim ?! ", "kim"},
		{word, "?", ""},
		{options, "Kim.", "kim."},
		{formatv0.Check{Kind: formatv0.CheckKindChoice}, "Kim.", "kim."},
		{formatv0.Check{Kind: formatv0.CheckKindNumber, Notation: new(formatv0.CheckNotationWord)}, "Kim.", "kim."},
	} {
		if got := ChoiceOf(c.check, c.text); got != c.want {
			t.Errorf("ChoiceOf(%v, %q) = %q, want %q", c.check.Notation, c.text, got, c.want)
		}
	}
	for text, want := range map[string]bool{"maya": true, "हिंदी": true, "́a": false, "a1": false, "two words": false, "": false, "o'neil": false} {
		if got := Word(text); got != want {
			t.Errorf("Word(%q) = %v, want %v", text, got, want)
		}
	}
}
