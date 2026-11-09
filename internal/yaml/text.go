package yaml

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

// The passes before go-yaml parses the text: they read what it reads otherwise than YAML 1.2.2,
// and give it a text it reads as YAML 1.2.2 does.

// decode is the first pass. It refuses, with yaml.encoding, a file that is not UTF-8, at its first
// byte that is no part of a UTF-8 character, and, with yaml.printable, a character outside the
// printable set of YAML 1.2.2, in a comment and in quotes too, where YAML 1.2.2 takes more, as
// format 0.4 has it (format/README.md, section 4; R102 spec); go-yaml refuses one too, but without
// its place (TestRead, "control character"). Either file is read no further, and
// ok is false: in a file that is not UTF-8 the text is not known, and two keys that differ only in
// the bytes that are not UTF-8 would read as the same key.
func (r *reader) decode(data []byte) (text string, ok bool) {
	var b strings.Builder
	b.Grow(len(data))
	ok = true
	line, column := 1, 1
	for at := 0; at < len(data); {
		c, size := utf8.DecodeRune(data[at:])
		switch {
		case c == utf8.RuneError && size == 1:
			if ok {
				r.breaches = append(r.breaches, Breach{RuleEncoding, line, column, fmt.Sprintf("byte %d is not UTF-8", at+1)})
			}
			ok = false
		case !printable(c):
			ok = false
			r.breaches = append(r.breaches, Breach{RulePrintable, line, column, codePoint(c) + ": the format allows no such character, not even in a comment or in quotes"})
		}
		b.WriteRune(c)
		at += size
		switch {
		case c == '\n', c == '\r' && (at == len(data) || data[at] != '\n'):
			line, column = line+1, 1
		case c != '\r':
			column++
		}
	}
	return b.String(), ok
}

// printable says whether YAML 1.2.2 allows a character in its streams (5.1, c-printable). U+FFFD,
// which decode puts for a byte that is no part of a UTF-8 character, is one.
func printable(c rune) bool {
	return c == 0x09 || c == 0x0a || c == 0x0d || (c >= 0x20 && c <= 0x7e) || c == 0x85 ||
		(c >= 0xa0 && c <= 0xd7ff) || (c >= 0xe000 && c <= 0xfffd) || (c >= 0x10000 && c <= 0x10ffff)
}

func codePoint(c rune) string {
	return fmt.Sprintf("U+%04X", c)
}

const byteOrderMark = "\ufeff"

// streamLevel is the second pass. It reads the lines that stand between documents: from the start
// of the file and after each line "...", while they hold nothing but comments. There a byte order
// mark breaks yaml.encoding, and so does one at the start of a line that a document "---" follows
// with nothing but comments between, and a line that begins with % is a directive, which breaks
// yaml.no-directives (YAML 1.2.2, 9.2). go-yaml refuses %YAML 1.2 and a byte order mark after
// "...", so the pass takes each mark out and turns each % into #, which leaves a comment of the
// same length (TestRead, "directive" and "byte order mark"). A directive that no document "---"
// follows is no YAML, and the pass says so, since go-yaml then reads what follows as a document.
//
// A line "---" or "..." ends whatever stands before it, even a block scalar or a string in quotes
// (YAML 1.2.2, c-forbidden), so the pass reads these lines alike in any file that is YAML.
func (r *reader) streamLevel(text string) string {
	lines := splitLines(text)
	between := true
	directive := -1 // the line of a directive that awaits its document "---"
	unfollowed := func() {
		if directive >= 0 {
			r.breach("", directive, 0, `a directive that no document "---" follows`)
			directive = -1
		}
	}
	for i := range lines {
		content := lines[i].content
		if strings.HasPrefix(content, byteOrderMark) && (between || beforeDocument(lines[i:], content[len(byteOrderMark):])) {
			r.breaches = append(r.breaches, Breach{RuleEncoding, i + 1, 1, "a byte order mark"})
			content = content[len(byteOrderMark):]
			r.shifted[i] = true
		}
		switch {
		case marker(content, "..."):
			unfollowed()
			between = true
		case !between || blankOrComment(content):
		case strings.HasPrefix(content, "%"):
			r.breach(RuleNoDirectives, i, 0, "the directive "+strings.TrimRight(content, " \t"))
			if fields := strings.Fields(content); len(fields) > 1 && fields[0] == "%TAG" {
				r.tagHandles[fields[1]] = true
			}
			directive = i
			content = "#" + content[1:]
		default:
			if !marker(content, "---") {
				unfollowed()
			}
			directive = -1
			between = false
		}
		lines[i].content = content
	}
	unfollowed()
	var b strings.Builder
	b.Grow(len(text))
	for _, l := range lines {
		b.WriteString(l.content)
		b.WriteString(l.end)
	}
	return b.String()
}

// beforeDocument says whether a line whose rest is given, and the lines after it, hold nothing but
// comments up to a document "---", so that a byte order mark at its start is one (YAML 1.2.2, 9.2,
// l-yaml-stream).
func beforeDocument(lines []line, rest string) bool {
	for i, l := range lines {
		content := l.content
		if i == 0 {
			content = rest
		}
		switch {
		case marker(content, "---"):
			return true
		case !blankOrComment(content):
			return false
		}
	}
	return false
}

// marker says whether a line is the marker "---" or "...": the three characters in its first
// column, and after them a space, a tab or the end of the line.
func marker(content, m string) bool {
	return strings.HasPrefix(content, m) && (len(content) == 3 || content[3] == ' ' || content[3] == '\t')
}

func blankOrComment(content string) bool {
	rest := strings.TrimLeft(content, " \t")
	return rest == "" || rest[0] == '#'
}

// The characters YAML 1.1 takes for line breaks and YAML 1.2.2 for characters like any other.
var lineSeparators = []rune{0x85, 0x2028, 0x2029}

// standIns is the third pass. go-yaml takes U+0085, U+2028 and U+2029 for line breaks, as YAML
// 1.1 does, and reads a text with them otherwise than YAML 1.2.2, which takes them for characters
// like any other: `a: x<U+2028>  y` is "x<U+2028>y" to it, and where YAML 1.2.2 reads one line it
// may refuse the file or count lines that are not there (TestRead, "line separator"). The pass puts
// for each a character of the private use area that the text neither holds nor writes with an
// escape. Such a character is ns-char to YAML 1.2.2, as the three are, and go-yaml reads it as one,
// so go-yaml reads the text as YAML 1.2.2 reads the file. The values get the three back through
// r.restore, and since one character stands for one, no place moves.
func (r *reader) standIns(text string) string {
	var present []rune
	for _, c := range lineSeparators {
		if strings.ContainsRune(text, c) {
			present = append(present, c)
		}
	}
	if len(present) == 0 {
		return text
	}
	used := map[rune]bool{}
	for _, c := range text {
		used[c] = true
	}
	for _, c := range escapedCharacters(text) {
		used[c] = true
	}
	var forth, back []string
	next := rune(0xe000)
	for _, c := range present {
		for used[next] {
			next++
			if next == 0xf900 {
				// The basic plane's area is full: on to plane 15's.
				next = 0xf0000
			}
		}
		used[next] = true
		forth = append(forth, string(c), string(next))
		back = append(back, string(next), string(c))
	}
	r.restore = strings.NewReplacer(back...)
	return strings.NewReplacer(forth...).Replace(text)
}

// escapedCharacters gives every character the text writes with an escape \x, \u or \U, wherever
// it stands: more than the strings in double quotes hold is no harm here. An escape above
// U+10FFFF names no character, so it is no stand-in either.
func escapedCharacters(text string) []rune {
	var found []rune
	for at := strings.IndexByte(text, '\\'); at >= 0 && at+1 < len(text); {
		width := escapeWidth(rune(text[at+1]))
		if width > 0 && at+2+width <= len(text) {
			if c, err := strconv.ParseUint(text[at+2:at+2+width], 16, 32); err == nil && c <= unicode.MaxRune {
				found = append(found, rune(c))
			}
		}
		next := strings.IndexByte(text[at+1:], '\\')
		if next < 0 {
			break
		}
		at += 1 + next
	}
	return found
}

// restored gives a value of go-yaml with U+0085, U+2028 and U+2029 back in place of their
// stand-ins.
func (r *reader) restored(s string) string {
	if r.restore == nil {
		return s
	}
	return r.restore.Replace(s)
}

// A line of a text: what it holds, and the line break that ends it — CRLF, CR, LF, or nothing at
// the end of the text. These are the line breaks of YAML 1.2.2 (5.4), and once U+0085, U+2028 and
// U+2029 have stand-ins, go-yaml's too.
type line struct {
	content, end string
}

func splitLines(text string) []line {
	var lines []line
	for {
		at := strings.IndexAny(text, "\r\n")
		if at < 0 {
			return append(lines, line{text, ""})
		}
		end := text[at : at+1]
		if strings.HasPrefix(text[at:], "\r\n") {
			end = "\r\n"
		}
		lines = append(lines, line{text[:at], end})
		text = text[at+len(end):]
	}
}

// source is the text go-yaml reads, by lines of characters, so that a place of go-yaml — a line
// and a column of characters — finds its character.
type source struct {
	lines [][]rune
}

func newSource(text string) *source {
	split := splitLines(text)
	s := &source{lines: make([][]rune, len(split))}
	for i, l := range split {
		s.lines[i] = []rune(l.content)
	}
	return s
}

// at gives the character at a place, from 0: '\n' just past the end of a line, and 0 past the end
// of the text.
func (s *source) at(line, column int) rune {
	switch {
	case line < 0 || line >= len(s.lines):
		return 0
	case column >= len(s.lines[line]):
		return '\n'
	}
	return s.lines[line][column]
}

// escapeWidth gives how many hexadecimal digits follow the escape \x, \u or \U, and 0 after any
// other.
func escapeWidth(kind rune) int {
	switch kind {
	case 'x':
		return 2
	case 'u':
		return 4
	case 'U':
		return 8
	}
	return 0
}
