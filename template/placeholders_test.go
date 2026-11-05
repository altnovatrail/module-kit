package template

import (
	"slices"
	"testing"

	"github.com/altnovatrail/spec/types/formatv0"
)

// A string of a template is read from left to right: a doubled brace is one brace of the text, and
// {name} is a placeholder; any other brace is an error at the place of its character.
func TestPiecesOf(t *testing.T) {
	text := func(s string) piece { return piece{text: s} }
	name := func(n string) piece { return piece{placeholder: formatv0.TemplateName(n)} }
	for _, c := range []struct {
		s    string
		want []piece
		err  string
	}{
		{"", nil, ""},
		{"no braces", []piece{text("no braces")}, ""},
		{"{a}", []piece{name("a")}, ""},
		{"{a}{b}", []piece{name("a"), name("b")}, ""},
		{"= {product}", []piece{text("= "), name("product")}, ""},
		{"{first} ×", []piece{name("first"), text(" ×")}, ""},
		{"as {{{a}, {b}}}.", []piece{text("as {"), name("a"), text(", "), name("b"), text("}.")}, ""},
		{"{{a}}", []piece{text("{a}")}, ""},
		{"{{}}", []piece{text("{}")}, ""},
		{"{a_1}", []piece{name("a_1")}, ""},
		{"{x9_}", []piece{name("x9_")}, ""},
		{"{{{{", []piece{text("{{")}, ""},
		{"{", nil, "a { at 1 opens no placeholder: a placeholder is {name}, and a brace of the text is written twice"},
		{"× {", nil, "a { at 3 opens no placeholder: a placeholder is {name}, and a brace of the text is written twice"},
		{"{}", nil, "a { at 1 opens no placeholder: a placeholder is {name}, and a brace of the text is written twice"},
		{"{A}", nil, "a { at 1 opens no placeholder: a placeholder is {name}, and a brace of the text is written twice"},
		{"{1a}", nil, "a { at 1 opens no placeholder: a placeholder is {name}, and a brace of the text is written twice"},
		{"{_a}", nil, "a { at 1 opens no placeholder: a placeholder is {name}, and a brace of the text is written twice"},
		{"{a b}", nil, "a { at 1 opens no placeholder: a placeholder is {name}, and a brace of the text is written twice"},
		{"{a{b}", nil, "a { at 1 opens no placeholder: a placeholder is {name}, and a brace of the text is written twice"},
		{"{{{a}", []piece{text("{"), name("a")}, ""},
		{"{a}}", nil, "a } at 4 closes no placeholder: a brace of the text is written twice"},
		{"½}", nil, "a } at 2 closes no placeholder: a brace of the text is written twice"},
		{"{{é}{a}}", nil, "a } at 4 closes no placeholder: a brace of the text is written twice"},
	} {
		got, err := piecesOf(c.s)
		if c.err == "" && (err != nil || !slices.Equal(got, c.want)) || c.err != "" && (err == nil || err.Error() != c.err) {
			t.Errorf("%q: %v, %v; want %v, %q", c.s, got, err, c.want, c.err)
		}
	}
}

// The strings of a task outside its template are every string of its value, at its place, and none
// of its template.
func TestStringsOf(t *testing.T) {
	task := formatv0.Task{
		ID:        "t",
		Statement: []formatv0.StatementPart{{Text: new(formatv0.Text("{a} apples"))}},
		Items:     []formatv0.Item{{Text: new(formatv0.Text("a)"))}},
		Template:  &formatv0.Template{Parameters: map[formatv0.TemplateName]formatv0.Domain{"a": {From: 1, To: 2}}, Constraint: new(formatv0.TemplateExpression("a > 0"))},
	}
	var got []string
	for _, s := range stringsOf(task) {
		got = append(got, at(s.tokens...)+" "+s.text)
	}
	want := []string{"#/band ", "#/canvas/kind ", "#/id t", "#/items/0/text a)", "#/statement/0/text {a} apples"}
	if !slices.Equal(got, want) {
		t.Errorf("%q, want %q", got, want)
	}
}
