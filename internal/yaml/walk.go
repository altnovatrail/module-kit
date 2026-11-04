package yaml

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	goyaml "go.yaml.in/yaml/v4"
)

// The pass over go-yaml's tree for the rules about nodes, which builds the value as it goes, as
// the Walker of spec's tools/format-check/yaml.mjs does. go-yaml resolves a plain scalar by its own
// schema, so the pass reads its text by section 4 instead. It never follows an alias.

// The largest integer a double holds exactly, and every one below it: 2^53 − 1.
const maxSafeInteger = 1<<53 - 1

// The styles of a scalar that is not plain: in quotes, or a block scalar.
const nonPlainStyles = goyaml.DoubleQuotedStyle | goyaml.SingleQuotedStyle | goyaml.LiteralStyle | goyaml.FoldedStyle

// value checks a node and gives its value. key says the node is a key of a mapping.
func (r *reader) value(n *goyaml.Node, key bool) any {
	line, column := n.Line-1, n.Column-1
	if n.Kind == goyaml.AliasNode {
		r.breach(RuleNoAnchors, line, column, "the alias *"+n.Value)
		return nil
	}
	if n.Anchor != "" {
		r.breach(RuleNoAnchors, line, column, "the anchor &"+n.Anchor)
	}
	// go-yaml marks every explicit tag with TaggedStyle but the non-specific tag !, which it keeps
	// as the node's tag.
	if n.Style&goyaml.TaggedStyle != 0 || n.Tag == "!" {
		r.breach(RuleNoTags, line, column, "the tag "+n.Tag)
	}
	switch n.Kind {
	case goyaml.DocumentNode:
		if len(n.Content) == 0 {
			return nil
		}
		return r.value(n.Content[0], false)
	case goyaml.ScalarNode:
		return r.scalar(n, key)
	case goyaml.MappingNode:
		if key {
			r.breach(RuleStringKeys, line, column, "a key that is a mapping, not a string")
		}
		return r.mapping(n)
	case goyaml.SequenceNode:
		if key {
			r.breach(RuleStringKeys, line, column, "a key that is a sequence, not a string")
		}
		items := make([]any, len(n.Content))
		for i, item := range n.Content {
			items[i] = r.value(item, false)
		}
		return items
	}
	return nil
}

// mapping checks the keys and the values of a mapping. A key that is no string has no member: the
// file breaks yaml.string-keys and has no value anyway.
func (r *reader) mapping(n *goyaml.Node) map[string]any {
	members := make(map[string]any, len(n.Content)/2)
	first := map[string]int{}
	for i := 0; i+1 < len(n.Content); i += 2 {
		k, v := n.Content[i], n.Content[i+1]
		name, isString := r.value(k, true).(string)
		value := r.value(v, false)
		if !isString {
			continue
		}
		if line, seen := first[name]; seen {
			r.breach(RuleUniqueKeys, k.Line-1, k.Column-1, fmt.Sprintf("the key %s again, after line %d", strconv.Quote(name), line))
			continue
		}
		first[name] = k.Line
		members[name] = value
	}
	return members
}

// scalar checks a scalar and gives its value: a plain one as section 4 reads it, any other the
// string it holds.
func (r *reader) scalar(n *goyaml.Node, key bool) any {
	text := r.restored(n.Value)
	line, column := n.Line-1, n.Column-1
	if n.Style&nonPlainStyles == 0 {
		return r.plain(text, key, line, column)
	}
	r.characters(text, line, column)
	if n.Style&goyaml.DoubleQuotedStyle != 0 {
		l, c := r.src.start(n)
		r.escapes(l, c)
	}
	return text
}

// The plain strings that some parser reads as something else: YAML 1.1's booleans, and the words
// of the Core schema in another letter case (yaml.plain-strings).
var otherWords = map[string]bool{"y": true, "n": true, "yes": true, "no": true, "on": true, "off": true, "true": true, "false": true, "null": true}

// plain reads a plain scalar: null, true, false or a JSON number, as written, or else a string
// (yaml.json-scalars, yaml.plain-strings).
func (r *reader) plain(text string, key bool, line, column int) any {
	switch {
	case text == "null", text == "true", text == "false":
		if key {
			r.breach(RuleStringKeys, line, column, "a key that is "+text+", not a string")
		}
		if text == "null" {
			return nil
		}
		return text == "true"
	case jsonNumber(text):
		if key {
			r.breach(RuleStringKeys, line, column, "a key that is the number "+text+", not a string")
		}
		return r.number(text, line, column)
	case text == "":
		r.breach(RulePlainStrings, line, column, `an empty plain scalar: write null, or "" for an empty string`)
	case strings.ContainsAny(text[:1], "0123456789+-.~<="):
		r.breach(RulePlainStrings, line, column, fmt.Sprintf("%s begins with %s: write it in quotes", strconv.Quote(text), text[:1]))
	case otherWords[strings.ToLower(text)]:
		r.breach(RulePlainStrings, line, column, strconv.Quote(text)+" reads as another value in some parsers: write it in quotes")
	}
	r.characters(text, line, column)
	return text
}

// number gives a JSON number's value. It is finite as a double, and how precisely the double holds
// it does not matter. One written without a fraction or an exponent is an integer, as go-yaml
// tells int from float, and lies within ±(2^53−1). Negative zero is written in no notation, nor
// reached by a negative number that the double rounds to zero, such as -1e-400 (yaml.json-scalars).
func (r *reader) number(text string, line, column int) any {
	if !strings.ContainsAny(text, ".eE") {
		n, err := strconv.ParseInt(text, 10, 64)
		switch {
		case text == "-0":
			r.breach(RuleJSONScalars, line, column, text+" reads as negative zero")
		case err != nil || n > maxSafeInteger || n < -maxSafeInteger:
			r.breach(RuleJSONScalars, line, column, text+": an integer outside ±(2^53−1)")
		}
		return n
	}
	// The grammar of a JSON number is Go's too, so the only error is a value out of range, which is
	// infinite here: one too small for a double is zero, with its sign.
	f, _ := strconv.ParseFloat(text, 64)
	switch {
	case math.IsInf(f, 0):
		r.breach(RuleJSONScalars, line, column, text+" is larger than any double")
	case f == 0 && math.Signbit(f):
		r.breach(RuleJSONScalars, line, column, text+" reads as negative zero")
	}
	return f
}

// jsonNumber says whether a text is a JSON number (RFC 8259, 6).
func jsonNumber(text string) bool {
	at := 0
	digits := func() int {
		start := at
		for at < len(text) && text[at] >= '0' && text[at] <= '9' {
			at++
		}
		return at - start
	}
	if at < len(text) && text[at] == '-' {
		at++
	}
	switch {
	case at < len(text) && text[at] == '0':
		at++
	case digits() == 0:
		return false
	}
	if at < len(text) && text[at] == '.' {
		at++
		if digits() == 0 {
			return false
		}
	}
	if at < len(text) && (text[at] == 'e' || text[at] == 'E') {
		at++
		if at < len(text) && (text[at] == '+' || text[at] == '-') {
			at++
		}
		if digits() == 0 {
			return false
		}
	}
	return at == len(text)
}

// characters refuses a noncharacter in a key or a string, written or escaped (yaml.characters). A
// surrogate is not there to see: UTF-8 holds none, and an escaped one go-yaml refuses, which
// parseError reads.
func (r *reader) characters(text string, line, column int) {
	for _, c := range text {
		if noncharacter(c) {
			r.breach(RuleCharacters, line, column, codePoint(c)+", a noncharacter")
		}
	}
}

// noncharacter says whether a character is one: U+FDD0 to U+FDEF, and the last two of every plane.
func noncharacter(c rune) bool {
	return (c >= 0xfdd0 && c <= 0xfdef) || c&0xfffe == 0xfffe
}

// escapes reads the escapes of a string in double quotes, from its opening quote at a place of
// src, from left to right, as YAML reads them, so that "\\/" is a backslash and a slash. \/ breaks
// yaml.characters, which go-yaml reads as a slash, and so does an escaped surrogate, which go-yaml
// refuses: alone, as half of a pair, and in the form \U0000D800. An escape of no character at all,
// above U+10FFFF, is no YAML. It says whether it found an escape go-yaml refuses.
func (r *reader) escapes(line, column int) (refused bool) {
	if r.src.at(line, column) != '"' {
		return false
	}
	for column++; ; {
		switch r.src.at(line, column) {
		case 0, '"':
			return refused
		case '\n':
			line, column = line+1, 0
			continue
		case '\\':
		default:
			column++
			continue
		}
		kind := r.src.at(line, column+1)
		width := escapeWidth(kind)
		escape := string(r.src.lines[line][column:min(column+2+width, len(r.src.lines[line]))])
		var code uint64
		var err error
		if width > 0 {
			code, err = strconv.ParseUint(escape[2:], 16, 32)
		}
		switch {
		case kind == '/':
			r.breach(RuleCharacters, line, column, `the escape \/: write /`)
		case width == 0 || err != nil:
		case code >= 0xd800 && code <= 0xdfff:
			refused = true
			r.breach(RuleCharacters, line, column, "the escape "+escape+", a surrogate")
		case code > 0x10ffff:
			refused = true
			r.breach("", line, column, "the escape "+escape+" names no character")
		}
		if kind == '\n' {
			// An escaped line break: the string goes on at the next line.
			line, column = line+1, 0
			continue
		}
		column += 2 + width
	}
}
