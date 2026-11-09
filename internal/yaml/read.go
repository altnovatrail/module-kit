package yaml

import (
	"cmp"
	"errors"
	"fmt"
	"io"
	"slices"
	"strings"

	goyaml "go.yaml.in/yaml/v4"
)

// The rules of section 4 of format/README.md, in its order.
const (
	RuleEncoding       = "yaml.encoding"
	RulePrintable      = "yaml.printable"
	RuleSingleDocument = "yaml.single-document"
	RuleNoDirectives   = "yaml.no-directives"
	RuleNoAnchors      = "yaml.no-anchors"
	RuleNoTags         = "yaml.no-tags"
	RuleStringKeys     = "yaml.string-keys"
	RuleUniqueKeys     = "yaml.unique-keys"
	RuleJSONScalars    = "yaml.json-scalars"
	RulePlainStrings   = "yaml.plain-strings"
	RuleCharacters     = "yaml.characters"
)

// Rules are the IDs of the rules yaml.*, in the order of section 4.
var Rules = []string{
	RuleEncoding, RulePrintable, RuleSingleDocument, RuleNoDirectives, RuleNoAnchors, RuleNoTags,
	RuleStringKeys, RuleUniqueKeys, RuleJSONScalars, RulePlainStrings, RuleCharacters,
}

// A Breach is one reason a file is refused: a rule it breaks, or, when Rule is empty, that it is
// not YAML 1.2.2 at all. The format gives that refusal no ID: a file that is not YAML is no module
// file, whatever its rules say (format/README.md, section 4).
type Breach struct {
	Rule string
	// The place in the file as written: the line and the column, both from 1. The column counts
	// characters, a byte order mark among them, and a byte that is no part of a UTF-8 character
	// counts as one.
	Line, Column int
	Message      string
}

func (b Breach) String() string {
	rule := b.Rule
	if rule == "" {
		rule = "not YAML"
	}
	return fmt.Sprintf("%d:%d: %s: %s", b.Line, b.Column, rule, b.Message)
}

// A Refusal is why a file is not read: every breach found, in the order of their places, and at
// one place in the order of their rules.
type Refusal struct {
	Breaches []Breach
}

func (r *Refusal) Error() string {
	lines := make([]string, len(r.Breaches))
	for i, b := range r.Breaches {
		lines[i] = b.String()
	}
	return strings.Join(lines, "\n")
}

// panicMessage begins the message of a breach made of a panic of go-yaml, which recovering keeps
// from taking the process down. The fuzz test fails on it: such a panic is a fault to report.
const panicMessage = "go-yaml failed on this file: "

// Read reads a file of a module into its JSON value: a map[string]any, []any, string, int64 for a
// number written with neither a fraction nor an exponent, float64 for any other number, bool, or
// nil. A file that breaks a rule yaml.*, or is not YAML 1.2.2, gives a *Refusal and no value
// (format/README.md, section 4).
func Read(data []byte) (any, error) {
	r := reader{shifted: map[int]bool{}, tagHandles: map[string]bool{}}
	value := r.read(data)
	if len(r.breaches) > 0 {
		slices.SortStableFunc(r.breaches, func(a, b Breach) int {
			return cmp.Or(cmp.Compare(a.Line, b.Line), cmp.Compare(a.Column, b.Column), cmp.Compare(a.Rule, b.Rule))
		})
		return nil, &Refusal{Breaches: r.breaches}
	}
	return value, nil
}

// reader holds what the passes find and what they need of one another.
type reader struct {
	breaches []Breach
	// The text go-yaml reads, by lines, and the places in it that the passes look up.
	src *source
	// The lines of src, from 0, whose byte order mark the stream pass took out: their columns in
	// the file are one more.
	shifted map[int]bool
	// The handles the directives %TAG gave, which the stream pass made comments.
	tagHandles map[string]bool
	// What turns the stand-ins of U+0085, U+2028 and U+2029 back into them, or nil when the file
	// has none (text.go).
	restore *strings.Replacer
	// The strings in quotes, from their opening quotes to their closing ones (layout.go).
	quoted []span
}

// read runs the passes in order and gives the value of the file, which counts only when no pass
// found a breach.
func (r *reader) read(data []byte) any {
	text, ok := r.decode(data)
	if !ok {
		return nil
	}
	text = r.streamLevel(text)
	text = r.standIns(text)
	r.src = newSource(text)
	// On an error, the documents go-yaml read before it are YAML, and their breaches count too.
	docs, err := parse(text)
	if err != nil {
		r.parseError(err)
	} else if len(docs) == 0 {
		r.breaches = append(r.breaches, Breach{RuleSingleDocument, 1, 1, "no document: the file holds comments at most"})
		return nil
	}
	for _, doc := range docs[min(1, len(docs)):] {
		r.breach(RuleSingleDocument, doc.Line-1, doc.Column-1, "a second document")
	}
	r.quoted = r.quotedSpans(docs)
	var value any
	for i, doc := range docs {
		v := r.value(doc, false)
		if i == 0 {
			value = v
		}
	}
	// Where the strings in quotes of the text past the error stand is not known, so a byte order
	// mark outside them is looked for only in a text go-yaml read to its end.
	r.layout(docs, err == nil)
	return value
}

// parse reads the documents of the text with go-yaml, each as its tree of nodes, until the end or
// the first error. A panic of go-yaml comes back as an error.
func parse(text string) (docs []*goyaml.Node, err error) {
	defer func() {
		if p := recover(); p != nil {
			err = parserPanic{p}
		}
	}()
	loader, err := goyaml.NewLoader(strings.NewReader(text))
	if err != nil {
		return nil, err
	}
	for {
		var doc goyaml.Node
		err := loader.Load(&doc)
		if errors.Is(err, io.EOF) {
			return docs, nil
		}
		if err != nil {
			return docs, err
		}
		docs = append(docs, &doc)
	}
}

type parserPanic struct{ value any }

func (p parserPanic) Error() string { return panicMessage + fmt.Sprint(p.value) }

// breach records a breach at a place of src, from 0, which it turns into the place in the file.
func (r *reader) breach(rule string, line, column int, message string) {
	if r.shifted[line] {
		column++
	}
	r.breaches = append(r.breaches, Breach{rule, line + 1, column + 1, message})
}
