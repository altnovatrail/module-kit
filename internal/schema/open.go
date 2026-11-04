package schema

import (
	"fmt"
	"net/url"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

// Open gives the open view of value at location, as Validate takes them: value without its closed
// fields, which is what the learner's app receives of a part (format/README.md, section 6). A
// field is open when a subschema of its place that applies and passes carries "x-open": true, an
// element of an open array is open, and the root is open; an open object whose fields are all
// closed stays, as an empty object. The view is made of new maps and slices, and value is left as
// it is.
//
// A mark counts only where its subschema applies and passes, as JSON Schema 2020-12 collects an
// annotation (core, section 7.7.1.2), so Open asks the validator whether each branch of anyOf and
// oneOf, and each if, passes. The view is the format's of a value valid against the schema at
// location; of another value Open gives what the same walk gives. Like Validate, it panics on a
// location or a reference that names nothing in the set, and it may be called from many goroutines
// at once.
func Open(location string, value any) any {
	return format().open(location, value)
}

func (s *set) open(location string, value any) any {
	file, fragment, _ := strings.Cut(location, "#")
	return s.view(value, s.applicable(value, s.subschemaAt(file, fragment)))
}

// A subschema is what stands at a place in a schema of the set: the name of the schema's file, the
// tokens of the place, and the JSON there — an object, or true or false, which carry no marks.
type subschema struct {
	file string
	at   []string
	json any
}

// location names the subschema as Validate takes it, with the pointer after # even at the root, and
// so tells two subschemas apart.
func (n subschema) location() string {
	return n.file + Pointer(n.at)
}

func (n subschema) object() map[string]any {
	node, _ := n.json.(map[string]any)
	return node
}

// child is the subschema at tokens below n. A token that names nothing there — no member, or no
// element by its index in decimal digits — is a fault, as a location that names nothing is.
func (n subschema) child(tokens ...string) subschema {
	node := n.json
	for k, token := range tokens {
		found := false
		switch v := node.(type) {
		case map[string]any:
			node, found = v[token]
		case []any:
			i, err := strconv.Atoi(token)
			if found = err == nil && i >= 0 && i < len(v) && strconv.Itoa(i) == token; found {
				node = v[i]
			}
		}
		if !found {
			panic(fmt.Sprintf("schema: %s%s names nothing", n.file, Pointer(slices.Concat(n.at, tokens[:k+1]))))
		}
	}
	return subschema{n.file, slices.Concat(n.at, tokens), node}
}

// marked says whether a subschema carries the mark that opens its field.
func marked(n subschema) bool {
	return n.object()["x-open"] == true
}

// subschemaAt is the subschema at fragment, a JSON Pointer in URI fragment form, in the schema of
// file.
func (s *set) subschemaAt(file, fragment string) subschema {
	doc, ok := s.docs[file]
	if !ok {
		panic(fmt.Sprintf("schema: %s is no schema of the set", file))
	}
	return subschema{file: file, json: doc}.child(tokensOf(file, fragment)...)
}

// unescapeToken reads a token of a JSON Pointer: ~1 as / and ~0 as ~ (RFC 6901, section 4).
var unescapeToken = strings.NewReplacer("~1", "/", "~0", "~")

// tokensOf reads the tokens of a JSON Pointer in URI fragment form, without its #.
func tokensOf(file, fragment string) []string {
	pointer, err := url.PathUnescape(fragment)
	if err != nil || pointer != "" && !strings.HasPrefix(pointer, "/") {
		panic(fmt.Sprintf("schema: %s#%s encodes no JSON Pointer", file, fragment))
	}
	if pointer == "" {
		return nil
	}
	tokens := strings.Split(pointer[1:], "/")
	for i, token := range tokens {
		tokens[i] = unescapeToken.Replace(token)
	}
	return tokens
}

// resolve gives the subschema that a reference in the schema of from names. A reference is
// relative, to a definition of the same schema, another schema or a definition of another schema,
// and the $id of every schema of the set is the same but for the name of its file (format/README.md,
// section 5), so the part before # is that name.
func (s *set) resolve(from subschema, reference string) subschema {
	file, fragment, _ := strings.Cut(reference, "#")
	if file == "" {
		file = from.file
	}
	return s.subschemaAt(file, fragment)
}

// passes says whether value is valid against a subschema on its own.
func (s *set) passes(n subschema, value any) bool {
	return s.schema(n.location()).Validate(value) == nil
}

// applicable gives the subschemas that start at a place, and every subschema they bring to it, that
// apply to value and pass (format/README.md, section 6): what a $ref names, beside the other
// keywords there; every branch of allOf, each branch of anyOf that passes and the branch of oneOf
// that passes; then when if passes, and else when it fails; and each branch of dependentSchemas
// whose key value has. true and false carry no marks and bring nothing. A subschema reached twice —
// two references to one definition, or references that run in a circle — is taken once.
func (s *set) applicable(value any, start ...subschema) []subschema {
	var found []subschema
	seen := map[string]bool{}
	var visit func(n subschema)
	visit = func(n subschema) {
		node := n.object()
		location := n.location()
		if node == nil || seen[location] {
			return
		}
		seen[location] = true
		found = append(found, n)
		if reference, ok := node["$ref"].(string); ok {
			visit(s.resolve(n, reference))
		}
		branches, _ := node["allOf"].([]any)
		for i := range branches {
			visit(n.child("allOf", strconv.Itoa(i)))
		}
		for _, keyword := range []string{"anyOf", "oneOf"} {
			branches, _ := node[keyword].([]any)
			for i := range branches {
				if branch := n.child(keyword, strconv.Itoa(i)); s.passes(branch, value) {
					visit(branch)
				}
			}
		}
		_, then := node["then"]
		_, otherwise := node["else"]
		if _, ok := node["if"]; ok && (then || otherwise) {
			branch := "else"
			if s.passes(n.child("if"), value) {
				branch = "then"
			}
			if _, ok := node[branch]; ok {
				visit(n.child(branch))
			}
		}
		dependents, _ := node["dependentSchemas"].(map[string]any)
		object, _ := value.(map[string]any)
		for key := range dependents {
			if _, ok := object[key]; ok {
				visit(n.child("dependentSchemas", key))
			}
		}
	}
	for _, n := range start {
		visit(n)
	}
	return found
}

// view is the open view of value, an open place, given the subschemas that apply there and pass.
func (s *set) view(value any, applied []subschema) any {
	switch v := value.(type) {
	case []any:
		elements := make([]any, len(v))
		for i, element := range v {
			var subs []subschema
			for _, n := range applied {
				if sub, ok := elementSchema(n, i); ok {
					subs = append(subs, sub)
				}
			}
			elements[i] = s.view(element, s.applicable(element, subs...))
		}
		return elements
	case map[string]any:
		fields := map[string]any{}
		for key, field := range v {
			if subs := s.fieldSchemas(applied, key); slices.ContainsFunc(subs, marked) {
				fields[key] = s.view(field, s.applicable(field, subs...))
			}
		}
		return fields
	}
	return value
}

// fieldSchemas gives the subschemas of the field key that the subschemas of its object give: under
// properties, under each key of patternProperties that matches it, and under additionalProperties
// where neither matches in the same subschema.
func (s *set) fieldSchemas(applied []subschema, key string) []subschema {
	var subs []subschema
	for _, n := range applied {
		node := n.object()
		matched := false
		if properties, ok := node["properties"].(map[string]any); ok {
			if _, ok := properties[key]; ok {
				subs = append(subs, n.child("properties", key))
				matched = true
			}
		}
		patterns, _ := node["patternProperties"].(map[string]any)
		for pattern := range patterns {
			if s.pattern(pattern).MatchString(key) {
				subs = append(subs, n.child("patternProperties", pattern))
				matched = true
			}
		}
		if _, ok := node["additionalProperties"]; ok && !matched {
			subs = append(subs, n.child("additionalProperties"))
		}
	}
	return subs
}

// elementSchema gives the subschema of the element at index that the subschema of its array gives:
// under prefixItems at the index, or else under items.
func elementSchema(n subschema, index int) (subschema, bool) {
	node := n.object()
	if prefix, ok := node["prefixItems"].([]any); ok && index < len(prefix) {
		return n.child("prefixItems", strconv.Itoa(index)), true
	}
	if _, ok := node["items"]; ok {
		return n.child("items"), true
	}
	return subschema{}, false
}

// pattern compiles a key of patternProperties once. The format writes a pattern only with what Go
// and JavaScript read the same way (format/README.md, section 5), and a pattern that does not
// compile is a fault, as a schema that does not compile is.
func (s *set) pattern(p string) *regexp.Regexp {
	if re, ok := s.patterns.Load(p); ok {
		return re.(*regexp.Regexp)
	}
	re, _ := s.patterns.LoadOrStore(p, regexp.MustCompile(p))
	return re.(*regexp.Regexp)
}
