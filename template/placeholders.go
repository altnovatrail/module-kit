package template

import (
	"fmt"
	"maps"
	"slices"
	"strconv"
	"strings"

	"github.com/altnovatrail/spec/types/formatv0"

	"github.com/altnovatrail/module-kit/internal/schema"
)

// A piece of a string of a template: text, with each doubled brace as one, or the name of a
// placeholder.
type piece struct {
	text        string
	placeholder formatv0.TemplateName
}

// templateName says whether a text is a name as a placeholder names it: lowercase letters, digits
// and _, beginning with a letter (task.schema.json, $defs/template_name). How long it is, and whether
// it is a keyword of Starlark, the schema of the template's names says: a placeholder that names no
// number of the template names nothing.
func templateName(text string) bool {
	return text != "" && text[0] >= 'a' && text[0] <= 'z' && strings.Trim(text, "abcdefghijklmnopqrstuvwxyz0123456789_") == ""
}

// piecesOf reads a string of a template from left to right: {{ is {, }} is }, and {name} is a
// placeholder (format/README.md, section 23). Any other brace is an error, at the place of its
// character from 1, as the rule task.placeholders has it.
func piecesOf(s string) ([]piece, error) {
	var pieces []piece
	var text strings.Builder
	for at, place := 0, 1; at < len(s); place++ {
		c := s[at]
		if (c == '{' || c == '}') && at+1 < len(s) && s[at+1] == c {
			text.WriteByte(c)
			at += 2
			place++
			continue
		}
		switch c {
		case '}':
			return nil, fmt.Errorf("a } at %d closes no placeholder: a brace of the text is written twice", place)
		case '{':
			end := strings.IndexByte(s[at:], '}')
			if end < 0 || !templateName(s[at+1:at+end]) {
				return nil, fmt.Errorf("a { at %d opens no placeholder: a placeholder is {name}, and a brace of the text is written twice", place)
			}
			if text.Len() > 0 {
				pieces = append(pieces, piece{text: text.String()})
				text.Reset()
			}
			name := s[at+1 : at+end]
			pieces = append(pieces, piece{placeholder: formatv0.TemplateName(name)})
			at += end + 1
			place += len(name) + 1
			continue
		}
		// A character of the text, of one byte or more: the place counts characters.
		end := at + 1
		for end < len(s) && s[end]&0xc0 == 0x80 {
			end++
		}
		text.WriteString(s[at:end])
		at = end
	}
	if text.Len() > 0 {
		pieces = append(pieces, piece{text: text.String()})
	}
	return pieces, nil
}

// placeholdersOf gives the names that the placeholders of the pieces name, in order.
func placeholdersOf(pieces []piece) []formatv0.TemplateName {
	var names []formatv0.TemplateName
	for _, p := range pieces {
		if p.placeholder != "" {
			names = append(names, p.placeholder)
		}
	}
	return names
}

// A string of a task outside its template, with its place: the tokens of the JSON Pointer to it.
type stringAt struct {
	text   string
	tokens []string
}

// stringsOf gives every string of a task outside its template, as its JSON value holds it, each with
// its place, the members of an object in the order of their names (format/README.md, section 23).
func stringsOf(task formatv0.Task) []stringAt {
	value := schema.ValueOf(task).(map[string]any)
	delete(value, "template")
	var found []stringAt
	var walk func(v any, tokens []string)
	walk = func(v any, tokens []string) {
		switch v := v.(type) {
		case string:
			found = append(found, stringAt{v, tokens})
		case []any:
			for i, element := range v {
				walk(element, append(slices.Clip(tokens), strconv.Itoa(i)))
			}
		case map[string]any:
			for _, name := range slices.Sorted(maps.Keys(v)) {
				walk(v[name], append(slices.Clip(tokens), name))
			}
		}
	}
	walk(value, nil)
	return found
}

// fill gives the instance of a template, as its JSON value: the task's value without its template,
// with every placeholder of every string replaced by its number, in decimal digits with - before a
// negative one, and every doubled brace by one, in a single pass, so that what is put in is not read
// again (format/README.md, section 23). A string that breaks task.placeholders, and a placeholder
// whose number numbers does not give, are errors.
func fill(task formatv0.Task, numbers formatv0.Numbers) (any, error) {
	value := schema.ValueOf(task).(map[string]any)
	delete(value, "template")
	var walk func(v any, tokens []string) (any, error)
	walk = func(v any, tokens []string) (any, error) {
		switch v := v.(type) {
		case string:
			pieces, err := piecesOf(v)
			if err != nil {
				return nil, fmt.Errorf("%s: %w", schema.Pointer(tokens), err)
			}
			var b strings.Builder
			for _, p := range pieces {
				if p.placeholder == "" {
					b.WriteString(p.text)
					continue
				}
				number, ok := numbers[p.placeholder]
				if !ok {
					return nil, fmt.Errorf("%s: no number of the instance is {%s}", schema.Pointer(tokens), p.placeholder)
				}
				b.WriteString(strconv.FormatInt(number, 10))
			}
			return b.String(), nil
		case []any:
			for i, element := range v {
				filled, err := walk(element, append(slices.Clip(tokens), strconv.Itoa(i)))
				if err != nil {
					return nil, err
				}
				v[i] = filled
			}
		case map[string]any:
			for _, name := range slices.Sorted(maps.Keys(v)) {
				filled, err := walk(v[name], append(slices.Clip(tokens), name))
				if err != nil {
					return nil, err
				}
				v[name] = filled
			}
		}
		return v, nil
	}
	return walk(value, nil)
}
