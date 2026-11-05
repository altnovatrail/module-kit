package schema

import (
	"cmp"
	"fmt"
	"net/url"
	"slices"
	"strings"

	"github.com/santhosh-tekuri/jsonschema/v6"
	"github.com/santhosh-tekuri/jsonschema/v6/kind"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

// The validator gives its errors as a tree, each node a keyword at a place in the value, with the
// errors inside it as its causes. collect turns the tree into the errors that the examples of the
// format name (format/README.md, section 8):
//
//   - $ref, allOf, properties, items and if, then and else are never named: the keyword that
//     failed inside them is. The validator gives the errors of properties, items, then and else
//     straight in their place, and those of $ref and allOf under a node of their own, which collect
//     goes through, as it goes through the group of several errors at one place and the root;
//   - every other keyword is named where it fails, and what failed inside it is not: the branches
//     of anyOf and oneOf, which are named when none of them passes, what not holds, the elements
//     that contains tried and the names that propertyNames read;
//   - unevaluatedProperties holds false alone, and false stands nowhere else (section 5): the
//     validator gives the false schema at the field it refuses, and the keyword is named at the
//     object, as an example names it. The keyword is told by the end of the place of the false
//     schema, which no field or definition of the format can share, since their names are
//     snake_case (section 5).

// english words the messages of the validator.
var english = message.NewPrinter(language.English)

func collect(e *jsonschema.ValidationError, errs []Error) []Error {
	location := slices.Clone(e.InstanceLocation)
	switch k := e.ErrorKind.(type) {
	case *kind.Schema, *kind.Group, *kind.Reference, *kind.AllOf:
		for _, cause := range e.Causes {
			errs = collect(cause, errs)
		}
		return errs
	case *kind.FalseSchema:
		if !strings.HasSuffix(e.SchemaURL, "/unevaluatedProperties") || len(location) == 0 {
			panic(fmt.Sprintf("schema: false stands at %s, beside unevaluatedProperties alone in the format", e.SchemaURL))
		}
		// A field that a subschema declares is unevaluated too when that subschema fails, since
		// it gives no annotations then (JSON Schema 2020-12 core, section 7.7.1.2).
		field := location[len(location)-1]
		return append(errs, Error{"unevaluatedProperties", location[:len(location)-1],
			fmt.Sprintf("no subschema that passes here evaluates the field %q", field)})
	case *kind.PropertyNames:
		// Its place is wrong: the vocabulary propertyNames gives the same error at the right one
		// (names.go).
		return errs
	case *kind.Not:
		return append(errs, Error{"not", location, "the value matches the schema that not holds"})
	default:
		path := k.KeywordPath()
		if len(path) == 0 {
			// No keyword fails: the value is no JSON value, which the reader of YAML never gives.
			panic(fmt.Sprintf("schema: %s at %s", k.LocalizedString(english), Pointer(location)))
		}
		return append(errs, Error{path[0], location, k.LocalizedString(english)})
	}
}

// escapeToken writes a token of a JSON Pointer: ~ as ~0 and / as ~1 (RFC 6901, section 3).
var escapeToken = strings.NewReplacer("~", "~0", "/", "~1")

// Pointer writes a place in a value, token by token, as a JSON Pointer in URI fragment form
// (RFC 6901, sections 3 and 6): # for the whole value, then / before each token, with ~ written ~0
// and / written ~1, and what a fragment of a URI cannot hold percent-encoded, as in
// #/modules/example~1smallest/0.1.0/withdrawn.
func Pointer(tokens []string) string {
	var b strings.Builder
	for _, token := range tokens {
		b.WriteString("/")
		b.WriteString(escapeToken.Replace(token))
	}
	return "#" + (&url.URL{Fragment: b.String()}).EscapedFragment()
}

// ComparePointers orders two places in a file, JSON Pointers in URI fragment form, member by member:
// an index of an array by its number, so that #/items/2 comes before #/items/10, before any other
// token, and other tokens as text. A place comes before the places within it.
func ComparePointers(a, b string) int {
	as, bs := strings.Split(a, "/"), strings.Split(b, "/")
	for i := range min(len(as), len(bs)) {
		if n := compareTokens(as[i], bs[i]); n != 0 {
			return n
		}
	}
	return cmp.Compare(len(as), len(bs))
}

func compareTokens(a, b string) int {
	switch ai, bi := isIndex(a), isIndex(b); {
	case ai && bi:
		return cmp.Or(cmp.Compare(len(a), len(b)), strings.Compare(a, b))
	case ai != bi:
		// Indices first, so that the order is one whatever tokens meet.
		if ai {
			return -1
		}
		return 1
	}
	return strings.Compare(a, b)
}

// isIndex says whether a token is the index of an element of an array, as a JSON Pointer writes it:
// decimal digits, with no leading zero (RFC 6901, section 4).
func isIndex(token string) bool {
	if token == "" || token[0] == '0' && len(token) > 1 {
		return false
	}
	for _, c := range []byte(token) {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}
