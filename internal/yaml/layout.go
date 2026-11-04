package yaml

import (
	"fmt"
	"strings"
	"unicode/utf8"

	goyaml "go.yaml.in/yaml/v4"
)

// The pass over go-yaml's tree for what YAML 1.2.2 refuses and go-yaml reads all the same. Each is
// a case of the YAML test suite that go-yaml lets through (TestYAMLTestSuite), or a limit of YAML
// 1.2.2 that it does not keep:
//   - a line of a string in quotes or of a flow collection, in block context, that is indented no
//     more than the block collection around it (9C9N, QB6E), tabs among it (DK95/01, Y79Y/003);
//   - leading empty lines of a block scalar longer than its first line that is not empty (S98Z);
//   - an implicit key whose ":" stands more than 1024 characters after its start;
//   - a byte order mark outside a string in quotes and outside the places of the stream pass.
//
// The pass finds where a string in quotes or a flow collection ends in the text. go-yaml has read
// the text already, so the pass need not read its syntax again, only find the ends; the test suite
// holds it to YAML 1.2.2.

// A span is a string in quotes in src, from its opening quote to its closing one, at places from 0.
type span struct {
	line, column, endLine, endColumn int
}

func (s span) holds(line, column int) bool {
	return (line > s.line || line == s.line && column >= s.column) &&
		(line < s.endLine || line == s.endLine && column <= s.endColumn)
}

// start gives the place of a node's content: where go-yaml places the node, past its tag and its
// anchor and the space after them, comments and line breaks among it. A verbatim tag, !<…>, ends at
// its >, whatever it holds.
func (s *source) start(n *goyaml.Node) (line, column int) {
	line, column = n.Line-1, n.Column-1
	for {
		c := s.at(line, column)
		switch {
		case c == '!' && s.at(line, column+1) == '<':
			for c != '>' && c != '\n' && c != 0 {
				column++
				c = s.at(line, column)
			}
			column++
		case c == '!' || c == '&':
			for c != ' ' && c != '\t' && c != '\n' && c != 0 && !strings.ContainsRune(",[]{}", c) {
				column++
				c = s.at(line, column)
			}
		default:
			return line, column
		}
	separation:
		for {
			switch s.at(line, column) {
			case ' ', '\t':
				column++
			case '#', '\n':
				line, column = line+1, 0
			default:
				break separation
			}
		}
	}
}

// leadingSpaces counts the spaces a line of src begins with.
func (s *source) leadingSpaces(line int) int {
	n := 0
	for n < len(s.lines[line]) && s.lines[line][n] == ' ' {
		n++
	}
	return n
}

// quotedEnd gives the place of the closing quote of a string whose opening quote is at a place: two
// single quotes stand for one in single quotes, and a backslash escapes the next character in
// double ones.
func (s *source) quotedEnd(line, column int) (int, int) {
	quote := s.at(line, column)
	for column++; ; column++ {
		switch c := s.at(line, column); {
		case c == 0:
			return line, column
		case c == '\n':
			line, column = line+1, -1
		case quote == '"' && c == '\\':
			if s.at(line, column+1) == '\n' {
				line, column = line+1, -1
			} else {
				column++
			}
		case c == quote && quote == '\'' && s.at(line, column+1) == '\'':
			column++
		case c == quote:
			return line, column
		}
	}
}

// flowEnd gives the place of the bracket that closes a flow collection whose opening bracket is at
// a place, past the strings in quotes, whose ends ends gives by their starts, and past comments. A
// plain scalar in a flow collection holds no bracket (YAML 1.2.2, ns-plain-safe).
func (s *source) flowEnd(line, column int, ends map[[2]int][2]int) (int, int) {
	depth := 0
	afterSpace := false
	for {
		if end, ok := ends[[2]int{line, column}]; ok {
			line, column, afterSpace = end[0], end[1]+1, false
			continue
		}
		c := s.at(line, column)
		switch {
		case c == 0:
			return line, column
		case c == '\n':
			line, column, afterSpace = line+1, 0, true
			continue
		case c == '#' && (afterSpace || column == 0):
			line, column, afterSpace = line+1, 0, true
			continue
		case c == '[' || c == '{':
			depth++
		case c == ']' || c == '}':
			depth--
			if depth == 0 {
				return line, column
			}
		}
		afterSpace = c == ' ' || c == '\t'
		column++
	}
}

func quotedStyle(n *goyaml.Node) bool {
	return n.Kind == goyaml.ScalarNode && n.Style&(goyaml.DoubleQuotedStyle|goyaml.SingleQuotedStyle) != 0
}

func flowCollection(n *goyaml.Node) bool {
	return (n.Kind == goyaml.MappingNode || n.Kind == goyaml.SequenceNode) && n.Style&goyaml.FlowStyle != 0
}

func blockCollection(n *goyaml.Node) bool {
	return (n.Kind == goyaml.MappingNode || n.Kind == goyaml.SequenceNode) && n.Style&goyaml.FlowStyle == 0
}

// quotedSpans finds every string in quotes of the documents.
func (r *reader) quotedSpans(docs []*goyaml.Node) []span {
	var spans []span
	var walk func(*goyaml.Node)
	walk = func(n *goyaml.Node) {
		if quotedStyle(n) {
			line, column := r.src.start(n)
			endLine, endColumn := r.src.quotedEnd(line, column)
			spans = append(spans, span{line, column, endLine, endColumn})
		}
		for _, child := range n.Content {
			walk(child)
		}
	}
	for _, doc := range docs {
		walk(doc)
	}
	return spans
}

// layout runs the checks of this pass over every document, and over the whole text for byte order
// marks when go-yaml read it to its end.
func (r *reader) layout(docs []*goyaml.Node, whole bool) {
	ends := map[[2]int][2]int{}
	// The lines that start inside a string in quotes: there a line that begins with # is no comment.
	inQuotes := map[int]bool{}
	for _, s := range r.quoted {
		ends[[2]int{s.line, s.column}] = [2]int{s.endLine, s.endColumn}
		for l := s.line + 1; l <= s.endLine; l++ {
			inQuotes[l] = true
		}
	}
	l := layout{r: r, ends: ends, inQuotes: inQuotes}
	for _, doc := range docs {
		l.node(doc, 0, false)
	}
	if whole {
		r.byteOrderMarks()
	}
}

type layout struct {
	r        *reader
	ends     map[[2]int][2]int // the end of each string in quotes, by its start
	inQuotes map[int]bool
}

// node checks a node and what it holds. indent is the indentation that the lines of a string in
// quotes or a flow collection need here: the column, from 1, of the block collection around the
// node, or 0 at the root, as s-flow-line-prefix(n) asks n spaces where the block collection stands
// at n − 1 (YAML 1.2.2, 6.3 and 8.2). inFlow says the node is in a flow collection, whose own check
// covers the node's lines.
func (l *layout) node(n *goyaml.Node, indent int, inFlow bool) {
	if n.Kind == goyaml.AliasNode {
		return
	}
	flow := flowCollection(n)
	if !inFlow && (flow || quotedStyle(n)) {
		l.lines(n, indent, flow)
	}
	if n.Kind == goyaml.ScalarNode && n.Style&(goyaml.LiteralStyle|goyaml.FoldedStyle) != 0 {
		l.blockScalar(n, indent)
	}
	if n.Kind == goyaml.MappingNode {
		l.implicitKeys(n)
	}
	inner := indent
	if !inFlow && blockCollection(n) {
		inner = l.indentation(n)
	}
	for _, child := range n.Content {
		l.node(child, inner, inFlow || flow)
	}
}

// indentation gives the column, from 1, at which a block collection's entries stand: where go-yaml
// places it, unless it has a tag or an anchor, where go-yaml places those, and the entries stand
// past them.
func (l *layout) indentation(n *goyaml.Node) int {
	if n.Anchor == "" && n.Style&goyaml.TaggedStyle == 0 && n.Tag != "!" {
		return n.Column
	}
	_, column := l.r.src.start(n)
	return column + 1
}

// lines checks the lines after the first of a string in quotes or of a flow collection in block
// context: each that holds more than spaces and tabs begins with at least indent spaces, since a
// tab is no indentation (YAML 1.2.2, 6.1). In a flow collection a comment line may stand anywhere,
// but not a line of a string in quotes that begins with #.
func (l *layout) lines(n *goyaml.Node, indent int, flow bool) {
	src := l.r.src
	line, column := src.start(n)
	var end int
	if flow {
		if c := src.at(line, column); c != '[' && c != '{' {
			// A single pair in a flow sequence: its sequence checks its lines.
			return
		}
		end, _ = src.flowEnd(line, column, l.ends)
	} else {
		end = l.ends[[2]int{line, column}][0]
	}
	what := "string in quotes"
	if flow {
		what = "flow collection"
	}
	for k := line + 1; k <= end && k < len(src.lines); k++ {
		text := src.lines[k]
		spaces := src.leadingSpaces(k)
		rest := strings.TrimLeft(string(text[spaces:]), " \t")
		if rest == "" || flow && rest[0] == '#' && !l.inQuotes[k] {
			continue
		}
		if spaces < indent {
			l.r.breach("", k, spaces, fmt.Sprintf("this line of a %s begins with %d spaces, and here it needs %d; a tab is no indentation", what, spaces, indent))
			return
		}
	}
}

// blockScalar checks a block scalar with no indentation indicator: none of its leading empty
// lines, which hold nothing but spaces, holds more spaces than its first line that is not empty,
// when that line is the scalar's, with at least indent spaces (YAML 1.2.2, 8.1.1.1). go-yaml takes
// such a first line for a comment when it begins with # (S98Z).
func (l *layout) blockScalar(n *goyaml.Node, indent int) {
	src := l.r.src
	line, column := src.start(n)
	for c := column + 1; ; c++ {
		h := src.at(line, c)
		if h >= '1' && h <= '9' {
			return
		}
		if h != '+' && h != '-' {
			break
		}
	}
	longest, longestLine := -1, -1
	for k := line + 1; k < len(src.lines); k++ {
		text := src.lines[k]
		spaces := src.leadingSpaces(k)
		if spaces == len(text) {
			if spaces > longest {
				longest, longestLine = spaces, k
			}
			continue
		}
		content := string(text)
		if spaces >= indent && longest > spaces && !marker(content, "---") && !marker(content, "...") {
			l.r.breach("", longestLine, 0, fmt.Sprintf("this leading empty line of a block scalar holds %d spaces, more than the %d of its first line that is not empty", longest, spaces))
		}
		return
	}
}

// The most characters between the start of an implicit key and its ":" (YAML 1.2.2, 7.4.2, 8.2.2).
const implicitKeyLimit = 1024

// implicitKeys checks that the ":" of each implicit key of a mapping stands at most 1024 characters
// after the key's start. A key with "?" before it is explicit.
func (l *layout) implicitKeys(n *goyaml.Node) {
	src := l.r.src
	for i := 0; i < len(n.Content); i += 2 {
		k := n.Content[i]
		keyLine, keyColumn := k.Line-1, k.Column-1
		before := keyColumn - 1
		for before >= 0 && (src.at(keyLine, before) == ' ' || src.at(keyLine, before) == '\t') {
			before--
		}
		if before >= 0 && src.at(keyLine, before) == '?' {
			continue
		}
		line, column := src.start(k)
		switch {
		case quotedStyle(k):
			end := l.ends[[2]int{line, column}]
			line, column = end[0], end[1]
		case flowCollection(k) && strings.ContainsRune("[{", src.at(line, column)):
			line, column = src.flowEnd(line, column, l.ends)
		case k.Kind == goyaml.ScalarNode && k.Value != "":
			column += utf8.RuneCountInString(k.Value) - 1
		default:
			continue
		}
		if line != keyLine {
			continue
		}
		colon := column + 1
		for src.at(line, colon) == ' ' || src.at(line, colon) == '\t' {
			colon++
		}
		if src.at(line, colon) == ':' && colon-keyColumn > implicitKeyLimit {
			l.r.breach("", keyLine, keyColumn, fmt.Sprintf("the \":\" of this implicit key stands %d characters after its start, and YAML allows %d", colon-keyColumn, implicitKeyLimit))
		}
	}
}

// byteOrderMarks refuses a byte order mark that the stream pass did not take out and that stands
// outside a string in quotes: YAML 1.2.2 allows one only at the start of a document and in quotes
// (5.2), and nb-char excludes it elsewhere.
func (r *reader) byteOrderMarks() {
	for line, text := range r.src.lines {
		for column, c := range text {
			if c != 0xfeff {
				continue
			}
			inside := false
			for _, s := range r.quoted {
				if s.holds(line, column) {
					inside = true
					break
				}
			}
			if !inside {
				r.breach("", line, column, "U+FEFF, a byte order mark, which YAML allows only at the start of a document or in quotes")
			}
		}
	}
}
