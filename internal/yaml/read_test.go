package yaml

import (
	"errors"
	"fmt"
	"reflect"
	"slices"
	"strings"
	"testing"
)

// A case of TestRead: a file, and either the value it reads to or each breach that refuses it, as
// "<rule> <line>:<column>", with "not YAML" for no rule.
type readCase struct {
	name    string
	in      string
	value   any
	refused []string
}

// The cases hold every rule, and each place where go-yaml reads otherwise than YAML 1.2.2 and a
// pass of this package reads as YAML 1.2.2 does (R21). A case named after the YAML test suite is
// that case of it, data-2022-01-17.
var readCases = []readCase{
	// What is read.
	{name: "mapping, sequence and scalars",
		in: "title: A title\ncount: 3\nweight: 2.5\nlarge: 1e300\ndone: true\nnothing: null\nlist: [a, \"b\", 'c']\nempty: \"\"\n",
		value: map[string]any{"title": "A title", "count": int64(3), "weight": 2.5, "large": 1e300, "done": true,
			"nothing": nil, "list": []any{"a", "b", "c"}, "empty": ""}},
	{name: "integers at the bounds", in: "max: 9007199254740991\nmin: -9007199254740991\nzero: 0\n",
		value: map[string]any{"max": int64(9007199254740991), "min": int64(-9007199254740991), "zero": int64(0)}},
	{name: "a double that holds an integer", in: "a: 50.0\nb: 5e1\nc: 1e-400\nd: 9007199254740993.0\n",
		value: map[string]any{"a": 50.0, "b": 50.0, "c": 0.0, "d": 9007199254740992.0}},
	{name: "JSON syntax", in: `{"a": [1, -2.5e-3, "x"], "b": {"c": null, "d": false}}`,
		value: map[string]any{"a": []any{int64(1), -2.5e-3, "x"}, "b": map[string]any{"c": nil, "d": false}}},
	{name: "plain strings", in: "a: x1.5\nb: A title, and more\nc: 'quoted: yes'\nd: a:b # not a comment: a word\n",
		value: map[string]any{"a": "x1.5", "b": "A title, and more", "c": "quoted: yes", "d": "a:b"}},
	{name: "block scalars", in: "a: |\n  x\n  y\nb: >\n  x\n  y\nc: |2\n   x\nd: |\n\n  x\n",
		value: map[string]any{"a": "x\ny\n", "b": "x y\n", "c": " x\n", "d": "\nx\n"}},
	{name: "escapes", in: `a: "\u00e9\t\\/\x41\N"` + "\n",
		value: map[string]any{"a": "é\t\\/A\u0085"}},
	{name: "an explicit key and a single pair", in: "? a\n: [b: c]\n",
		value: map[string]any{"a": []any{map[string]any{"b": "c"}}}},
	{name: "comments and document markers", in: "# a comment\n--- # the document\na: b # a word\n...\n# after\n",
		value: map[string]any{"a": "b"}},
	{name: "line breaks CRLF", in: "a: b\r\nc: \"d\"\r\n", value: map[string]any{"a": "b", "c": "d"}},
	{name: "line breaks CR", in: "a: b\rc: d\r", value: map[string]any{"a": "b", "c": "d"}},
	{name: "a root block scalar with a line that begins with %", in: "--- |\n%!PS-Adobe-2.0\n",
		value: "%!PS-Adobe-2.0\n"},
	{name: "a quoted merge key", in: "\"<<\": base\n", value: map[string]any{"<<": "base"}},

	// U+0085, U+2028 and U+2029, which go-yaml takes for line breaks (text.go, standIns).
	{name: "line separator in a plain scalar", in: "a: x\u2028  y\n", value: map[string]any{"a": "x\u2028  y"}},
	{name: "line separator in a block scalar", in: "a: |\n  x\u2028  y\n", value: map[string]any{"a": "x\u2028  y\n"}},
	{name: "next line in a key and in quotes", in: "\u0085: \"x\u0085y\"\n", value: map[string]any{"\u0085": "x\u0085y"}},
	{name: "paragraph separator in a comment", in: "# c\u2029x: y\na: b\n", value: map[string]any{"a": "b"}},
	{name: "a stand-in written as an escape", in: "a: \"\\uE000\"\nb: x\u2028\n", value: map[string]any{"a": "\ue000", "b": "x\u2028"}},
	{name: "an escaped line separator", in: `t: "\L"`, value: map[string]any{"t": "\u2028"}},

	// What YAML 1.2.2 allows that is close to what it refuses (layout.go).
	{name: "a byte order mark in quotes", in: "a: \"x\ufeffy\"\nb: 'z\ufeff'\n", value: map[string]any{"a": "x\ufeffy", "b": "z\ufeff"}},
	{name: "lines of a flow collection and of quotes indented enough", in: "a:\n  b: [x,\n   z]\n  c: \"p\n   q\"\n",
		value: map[string]any{"a": map[string]any{"b": []any{"x", "z"}, "c": "p q"}}},
	{name: "a comment line anywhere in a flow collection", in: "a:\n  b: [x,\n# c\n   z]\n",
		value: map[string]any{"a": map[string]any{"b": []any{"x", "z"}}}},
	{name: "Y79Y/002, an empty line with a tab in a flow collection", in: "- [\n\t\n foo\n ]\n", value: []any{[]any{"foo"}}},
	{name: "DK95/02, a tab after the indentation", in: "foo: \"bar\n  \tbaz\"\n", value: map[string]any{"foo": "bar baz"}},
	{name: "a block scalar whose leading empty lines are shorter", in: "a: |\n \n  x\n", value: map[string]any{"a": "\nx\n"}},
	{name: "a block scalar ended by a document marker", in: "--- |\n  \n...\n", value: ""},
	{name: "an implicit key at the limit", in: "\"" + strings.Repeat("k", 1022) + "\": v\n",
		value: map[string]any{strings.Repeat("k", 1022): "v"}},
	{name: "an explicit key beyond the limit", in: "? \"" + strings.Repeat("k", 1100) + "\"\n: v\n",
		value: map[string]any{strings.Repeat("k", 1100): "v"}},

	// yaml.encoding.
	{name: "byte order mark", in: "\ufeffa: b\n", refused: []string{"yaml.encoding 1:1"}},
	{name: "byte order mark before a directive", in: "\ufeff%YAML 1.2\n---\na: b\n", refused: []string{"yaml.encoding 1:1", "yaml.no-directives 1:2"}},
	{name: "byte order mark after ...", in: "a: b\n...\n\ufeff# c\n", refused: []string{"yaml.encoding 3:1"}},
	{name: "byte order mark before a document", in: "a: b\n\ufeff# c\n---\nc: d\n", refused: []string{"yaml.encoding 2:1", "yaml.single-document 3:1"}},
	{name: "byte order mark on the line of a document", in: "a: b\n...\n\ufeff--- c\n", refused: []string{"yaml.encoding 3:1", "yaml.single-document 3:2"}},
	{name: "not UTF-8", in: "a: \"\xff\"\nb: \xfe\n", refused: []string{"yaml.encoding 1:5"}},
	{name: "not UTF-8 in two keys", in: "caf\xe9: 1\ncaf\xe8: 2\n", refused: []string{"yaml.encoding 1:4"}},

	// What is not YAML in its characters.
	{name: "control character", in: "a: \x01\n", refused: []string{"yaml.printable 1:4"}},
	{name: "control character in a comment", in: "# a\x7f\nb: c\n", refused: []string{"yaml.printable 1:4"}},
	{name: "C1 control character in quotes", in: "a: \"\u0080\"\n", refused: []string{"yaml.printable 1:5"}},
	{name: "U+FFFE written", in: "a: \"\ufffe\"\n", refused: []string{"yaml.printable 1:5"}},

	// yaml.single-document.
	{name: "comment only", in: "# A comment, and no document.\n", refused: []string{"yaml.single-document 1:1"}},
	{name: "empty file", in: "", refused: []string{"yaml.single-document 1:1"}},
	{name: "two documents", in: "a: 1\n---\nb: 2\n", refused: []string{"yaml.single-document 2:1"}},

	// yaml.no-directives.
	{name: "directive %YAML 1.2", in: "%YAML 1.2\n---\na: b\n", refused: []string{"yaml.no-directives 1:1"}},
	{name: "directive %YAML 1.1", in: "%YAML 1.1\n---\na: b\n", refused: []string{"yaml.no-directives 1:1"}},
	{name: "directive %TAG and its handle", in: "%TAG !n! tag:x,2026:\n---\na: !n!x b\n",
		refused: []string{"yaml.no-directives 1:1", "yaml.no-tags 3:4"}},
	{name: "directive after ...", in: "a: b\n...\n# c\n%YAML 1.2\n---\nc: d\n",
		refused: []string{"yaml.no-directives 4:1", "yaml.single-document 5:1"}},
	{name: "directive inside a document", in: "a: b\n%YAML 1.2\n---\nc: d\n", refused: []string{"not YAML 2:1"}},
	{name: "a directive that no document follows", in: "a: b\n...\n%YAML 1.2\nc: d\n",
		refused: []string{"not YAML 3:1", "yaml.no-directives 3:1", "yaml.single-document 4:1"}},
	{name: "a directive alone", in: "%YAML 1.2\n", refused: []string{"not YAML 1:1", "yaml.no-directives 1:1", "yaml.single-document 1:1"}},

	// yaml.no-anchors.
	{name: "anchor", in: "a: &x b\n", refused: []string{"yaml.no-anchors 1:4"}},
	{name: "anchor and alias", in: "a: &x [b]\nc: *x\n", refused: []string{"yaml.no-anchors 1:4", "yaml.no-anchors 2:4"}},
	{name: "alias with no anchor", in: "a: *x\n", refused: []string{"yaml.no-anchors 1:4"}},
	{name: "anchor with a dot", in: "a: &x.y b\n", refused: []string{"yaml.no-anchors 1:4"}},
	{name: "8XYN, anchor go-yaml does not scan", in: "- &\U0001F600 b\n", refused: []string{"yaml.no-anchors 1:3"}},
	{name: "alias go-yaml does not scan", in: "- *\U0001F600\n", refused: []string{"yaml.no-anchors 1:3"}},
	{name: "an anchor with no name", in: "a: &\n", refused: []string{"not YAML 1:5"}},
	{name: "an alias with no name", in: "a: *\n", refused: []string{"not YAML 1:5"}},
	{name: "an anchor with no name in a flow sequence", in: "a: [b, &]\n", refused: []string{"not YAML 1:9"}},
	{name: "2SXE, an anchor whose name ends in a colon", in: "&a: key: &a value\nfoo:\n  *a:\n", refused: []string{"yaml.no-anchors 1:1"}},
	{name: "W5VH, such an anchor among other characters", in: "a: &:@*!$\"<foo>: scalar a\nb: *:@*!$\"<foo>:\n", refused: []string{"yaml.no-anchors 1:4"}},
	{name: "an alias whose name ends in colons in a flow mapping, which go-yaml refuses as an alias with no anchor", in: "a: {b: c, *d:: e}\n", refused: []string{"yaml.no-anchors 1:11"}},
	{name: "& and a colon inside a plain scalar", in: "a: x &b: y\n", refused: []string{"not YAML 1:8"}},
	{name: "& and a colon in a comment", in: "a: b # &c:\nd: [\n", refused: []string{"not YAML 3:1"}},

	// yaml.no-tags.
	{name: "standard tag", in: "a: !!str 1\n", refused: []string{"yaml.no-tags 1:4"}},
	{name: "local tag", in: "a: !thing b\n", refused: []string{"yaml.no-tags 1:4"}},
	{name: "non-specific tag", in: "a: ! 12\nb: ! \"x\"\n", refused: []string{"yaml.no-tags 1:4", "yaml.no-tags 2:4"}},
	{name: "verbatim tag", in: "a: !<tag:yaml.org,2002:str> b\n", refused: []string{"yaml.no-tags 1:4"}},
	{name: "tag of a collection", in: "--- !!map\na: b\n", refused: []string{"yaml.no-tags 1:5"}},
	{name: "tag go-yaml does not scan", in: "a: !a#b c\n", refused: []string{"yaml.no-tags 1:4"}},
	{name: "a verbatim tag with a comma before quotes", in: "a: !<tag:x,y> \"q\\/\"\nb: c\n", refused: []string{"yaml.no-tags 1:4", "yaml.characters 1:17"}},
	{name: "an unclosed verbatim tag", in: "a: !<tag:x b\n", refused: []string{"not YAML 1:11"}},
	{name: "a tag a flow indicator follows in block context", in: "e: !x,y\n", refused: []string{"not YAML 1:6"}},
	{name: "a tag with a broken escape", in: "a: !!str%zz b\n", refused: []string{"not YAML 1:9"}},
	{name: "a tag whose handle no %TAG gave", in: "a: !n!x b\n", refused: []string{"not YAML 1:4"}},

	// yaml.string-keys.
	{name: "boolean key", in: "true: x\n", refused: []string{"yaml.string-keys 1:1"}},
	{name: "null key", in: "null: x\n", refused: []string{"yaml.string-keys 1:1"}},
	{name: "number key", in: "1: x\n", refused: []string{"yaml.string-keys 1:1"}},
	{name: "sequence key", in: "? [a, b]\n: x\n", refused: []string{"yaml.string-keys 1:3"}},
	{name: "mapping key", in: "{a: b}: x\n", refused: []string{"yaml.string-keys 1:1"}},

	// yaml.unique-keys.
	{name: "repeated key", in: "a: 1\na: 2\n", refused: []string{"yaml.unique-keys 2:1"}},
	{name: "escaped key", in: "a: 1\n\"\\u0061\": 2\n", refused: []string{"yaml.unique-keys 2:1"}},
	{name: "repeated key in a flow mapping", in: "{a: 1, 'a': 2}\n", refused: []string{"yaml.unique-keys 1:8"}},
	{name: "empty keys", in: "? \n: a\n? \n: b\n",
		refused: []string{"yaml.plain-strings 1:2", "yaml.plain-strings 3:2", "yaml.unique-keys 3:2"}},

	// yaml.json-scalars.
	{name: "integer above the range", in: "a: 9007199254740992\n", refused: []string{"yaml.json-scalars 1:4"}},
	{name: "integer below the range", in: "a: -9007199254740992\n", refused: []string{"yaml.json-scalars 1:4"}},
	{name: "integer beyond 64 bits", in: "a: 99999999999999999999\n", refused: []string{"yaml.json-scalars 1:4"}},
	{name: "negative zero", in: "a: -0\nb: -0.0\nc: -0e5\nd: -1e-400\n",
		refused: []string{"yaml.json-scalars 1:4", "yaml.json-scalars 2:4", "yaml.json-scalars 3:4", "yaml.json-scalars 4:4"}},
	{name: "overflow", in: "a: 1e400\n", refused: []string{"yaml.json-scalars 1:4"}},

	// yaml.plain-strings.
	{name: "empty plain scalar", in: "a:\n", refused: []string{"yaml.plain-strings 1:3"}},
	{name: "empty document", in: "---\n", refused: []string{"yaml.plain-strings 2:1"}},
	{name: "leading characters", in: "a: 0777\nb: +1\nc: .inf\nd: ~\ne: =\nf: <x\ng: -x\n", refused: []string{
		"yaml.plain-strings 1:4", "yaml.plain-strings 2:4", "yaml.plain-strings 3:4", "yaml.plain-strings 4:4",
		"yaml.plain-strings 5:4", "yaml.plain-strings 6:4", "yaml.plain-strings 7:4"}},
	{name: "other words", in: "a: Yes\nb: NULL\nc: off\nd: N\n", refused: []string{
		"yaml.plain-strings 1:4", "yaml.plain-strings 2:4", "yaml.plain-strings 3:4", "yaml.plain-strings 4:4"}},
	{name: "merge key", in: "<<: base\n", refused: []string{"yaml.plain-strings 1:1"}},
	{name: "a number over two lines", in: "a: 1\n  2\n", refused: []string{"yaml.plain-strings 1:4"}},

	// yaml.characters.
	{name: "escaped noncharacter", in: "a: \"\\uFFFE\"\nb: \"\\U0010FFFF\"\n", refused: []string{"yaml.characters 1:4", "yaml.characters 2:4"}},
	{name: "noncharacter in a key", in: "\"\\uFDD0\": x\n", refused: []string{"yaml.characters 1:1"}},
	{name: "written noncharacter", in: "a: x\ufdd0y\n", refused: []string{"yaml.characters 1:4"}},
	{name: "escaped slash", in: "a: \"a\\/b\"\nb: \"\\\\/\"\n", refused: []string{"yaml.characters 1:6"}},
	{name: "escaped slash after a tag and an anchor", in: "a: !!str \"x\\/\"\nb: &q \"y\\/\"\n",
		refused: []string{"yaml.no-tags 1:4", "yaml.characters 1:12", "yaml.no-anchors 2:4", "yaml.characters 2:9"}},
	{name: "escaped surrogate", in: "a: b\nc: \"x\\ud800\"\n", refused: []string{"yaml.characters 2:6"}},
	{name: "escaped surrogate pair", in: "a: \"\\ud83d\\ude00\"\n", refused: []string{"yaml.characters 1:5", "yaml.characters 1:11"}},
	{name: "long escaped surrogate", in: "a: \"\\U0000D800\"\n", refused: []string{"yaml.characters 1:5"}},
	{name: "escape of no character", in: "a: \"\\U00110000\"\n", refused: []string{"not YAML 1:5"}},

	// What YAML 1.2.2 refuses and go-yaml reads (layout.go).
	{name: "9C9N, a flow sequence indented too little", in: "---\nflow: [a,\nb,\nc]\n", refused: []string{"not YAML 3:1"}},
	{name: "a flow sequence indented too little in a nested mapping", in: "tasks:\n  rung: [a,\n  b]\n", refused: []string{"not YAML 3:3"}},
	{name: "a closing bracket indented too little", in: "a: [b,\n]\n", refused: []string{"not YAML 2:1"}},
	{name: "QB6E, a string in quotes indented too little", in: "---\nquoted: \"a\nb\nc\"\n", refused: []string{"not YAML 3:1"}},
	{name: "DK95/01, a tab for indentation in quotes", in: "foo: \"bar\n\tbaz\"\n", refused: []string{"not YAML 2:1"}},
	{name: "Y79Y/003, a tab for indentation in a flow collection", in: "- [\n\tfoo,\n foo\n ]\n", refused: []string{"not YAML 2:1"}},
	{name: "a line of quotes that begins with # in a flow collection", in: "a:\n  b: [\"x\n  #y\"]\n", refused: []string{"not YAML 3:3"}},
	{name: "S98Z, leading empty lines of a block scalar", in: "empty block scalar: >\n \n  \n   \n # comment\n", refused: []string{"not YAML 4:1"}},
	{name: "an implicit key beyond the limit", in: "\"" + strings.Repeat("k", 1023) + "\": v\n", refused: []string{"not YAML 1:1"}},
	{name: "a byte order mark in a plain scalar", in: "a: x\ufeffy\n", refused: []string{"not YAML 1:5"}},
	{name: "a byte order mark in a block scalar", in: "a: |\n  x\ufeffy\n", refused: []string{"not YAML 2:4"}},
	{name: "a byte order mark in a comment", in: "a: b # x\ufeff\n", refused: []string{"not YAML 1:9"}},

	// What go-yaml refuses as not YAML.
	{name: "an unclosed flow sequence", in: "a: [b\n", refused: []string{"not YAML 2:1"}},
	{name: "an error after two documents", in: "a: x\ufeffy\n---\nb: 1\n---\n[\n", refused: []string{"yaml.single-document 2:1", "not YAML 6:1"}},
}

func TestRead(t *testing.T) {
	for _, test := range readCases {
		t.Run(test.name, func(t *testing.T) {
			value, err := Read([]byte(test.in))
			var refusal *Refusal
			switch {
			case test.refused == nil && err != nil:
				t.Fatalf("refused:\n%v", err)
			case test.refused == nil && !reflect.DeepEqual(value, test.value):
				t.Fatalf("read as %#v, not %#v", value, test.value)
			case test.refused == nil:
			case !errors.As(err, &refusal):
				t.Fatalf("read as %#v, with %v", value, err)
			case value != nil:
				t.Fatalf("refused with the value %#v", value)
			case !slices.Equal(places(refusal), test.refused):
				t.Fatalf("refused with\n%v\nnot %q", err, test.refused)
			}
		})
	}
}

// places gives each breach of a refusal as "<rule> <line>:<column>".
func places(refusal *Refusal) []string {
	var got []string
	for _, b := range refusal.Breaches {
		rule := b.Rule
		if rule == "" {
			rule = "not YAML"
		}
		got = append(got, fmt.Sprintf("%s %d:%d", rule, b.Line, b.Column))
	}
	return got
}

// A refusal names each breach on a line of its own.
func TestRefusalError(t *testing.T) {
	refusal := &Refusal{Breaches: []Breach{
		{RuleEncoding, 1, 1, "a byte order mark"},
		{"", 2, 3, "something"},
	}}
	want := "1:1: yaml.encoding: a byte order mark\n2:3: not YAML: something"
	if got := refusal.Error(); got != want {
		t.Errorf("%q, not %q", got, want)
	}
}

// An error of go-yaml that is no LoadError, such as a panic it recovered from, refuses the file as
// not YAML at its start.
func TestParseErrorOfAnotherKind(t *testing.T) {
	var r reader
	r.parseError(parserPanic{"index out of range"})
	want := []Breach{{"", 1, 1, panicMessage + "index out of range"}}
	if !slices.Equal(r.breaches, want) {
		t.Errorf("%v, not %v", r.breaches, want)
	}
}

// Every character the text writes with an escape is found, an escape above U+10FFFF is none, and a
// backslash at the end is no escape.
func TestEscapedCharacters(t *testing.T) {
	got := escapedCharacters(`a \x41 \u00e9 \U0001F600 \uZZZZ \\u0042 \UFFFFFFFF \U00110000 \u12`)
	want := []rune{'A', 'é', '😀', 'B'}
	if !slices.Equal(got, want) {
		t.Errorf("%q, not %q", got, want)
	}
	if got := escapedCharacters(`a\`); got != nil {
		t.Errorf("%q from a backslash at the end", got)
	}
}
