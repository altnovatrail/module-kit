package yaml

import (
	"errors"
	"strings"

	goyaml "go.yaml.in/yaml/v4"
)

// parseError is the pass over an error of go-yaml. Where the error is what a rule refuses, the
// file breaks that rule: go-yaml refuses an escaped surrogate, which YAML 1.2.2 reads and
// yaml.characters refuses; it scans no anchor or alias named as YAML 1.2.2 allows, such as &😀,
// nor every tag it allows; it refuses an alias with no anchor; and a tag whose handle a %TAG gave,
// which the stream pass made a comment, has none left. Where YAML 1.2.2 refuses the text too — an
// anchor with no name, a tag outside its grammar, a handle no %TAG gave — the file is not YAML.
// An anchor or an alias whose name ends in a colon go-yaml refuses for a reason that names none,
// and the file breaks yaml.no-anchors.
// The messages matched are go-yaml's at the version go.mod requires: TestRead fails on a version
// that words one otherwise. Any other error is the file's not being YAML, at go-yaml's place.
func (r *reader) parseError(err error) {
	var loadError *goyaml.LoadError
	if !errors.As(err, &loadError) {
		r.breaches = append(r.breaches, Breach{"", 1, 1, err.Error()})
		return
	}
	mark, context := loadError.Mark, loadError.ContextMark
	if context.Line == 0 {
		context = mark
	}
	switch {
	case loadError.Message == "found invalid Unicode character escape code" &&
		loadError.ContextMsg == "while scanning a quoted scalar" && context.Line > 0:
		// The context is the opening quote: the escapes of the string say which it is.
		if r.escapes(context.Line-1, context.Column-1) {
			return
		}
	case (loadError.ContextMsg == "while scanning an anchor" || loadError.ContextMsg == "while scanning an alias") &&
		r.src.named(context.Line-1, context.Column-1),
		loadError.Stage == goyaml.ComposerStage && strings.HasPrefix(loadError.Message, "unknown anchor "):
		r.at(RuleNoAnchors, context, "an anchor or an alias (go-yaml: "+loadError.Message+")")
		return
	case (loadError.ContextMsg == "while scanning a tag" || loadError.ContextMsg == "while parsing a tag") &&
		r.src.tag(context.Line-1, context.Column-1),
		loadError.Message == "found undefined tag handle" && r.tagHandles[r.src.handle(mark.Line-1, mark.Column-1)]:
		r.at(RuleNoTags, context, "a tag (go-yaml: "+loadError.Message+")")
		return
	}
	message := loadError.Message
	if loadError.ContextMsg != "" {
		message = loadError.ContextMsg + ": " + message
	}
	if line, column, ok := r.src.colonAnchor(mark.Line-1, mark.Column-1); ok {
		r.breach(RuleNoAnchors, line, column, "an anchor or an alias whose name ends in a colon, which go-yaml reads as a separator (go-yaml: "+message+")")
		return
	}
	r.at("", mark, message)
}

// colonAnchor finds the first anchor or alias, at or before the place go-yaml refused, whose name
// as YAML 1.2.2 reads it ends in a colon, as &a: and *a: do: go-yaml leaves the last colons of a
// name out on purpose, to read *a: b as an alias that is a key (yaml/go-yaml#109, closed as not
// planned), and refuses the file further on. Only an & or a * that starts a node counts, not
// one in a comment. go-yaml has refused the file already, so a place taken wrongly changes only
// which reason the refusal gives. line -1 is no place: then every line is looked at.
func (s *source) colonAnchor(line, column int) (int, int, bool) {
	last := len(s.lines) - 1
	if line >= 0 && line < last {
		last = line
	}
	for l := 0; l <= last; l++ {
		text := s.lines[l]
		for c, r := range text {
			if l == line && c > column {
				break
			}
			if r == '#' && (c == 0 || text[c-1] == ' ' || text[c-1] == '\t') {
				break
			}
			if (r != '&' && r != '*') || !s.nodeStart(l, c) {
				continue
			}
			end := c + 1
			for anchorChar(s.at(l, end)) {
				end++
			}
			if end > c+1 && text[end-1] == ':' {
				return l, c, true
			}
		}
	}
	return 0, 0, false
}

// nodeStart says whether the character at a place stands where a node may start: first on its line
// but for spaces and tabs, after a flow indicator [, { or , or after an indicator - ? or : and the
// space or tab that follows it.
func (s *source) nodeStart(line, column int) bool {
	p := column - 1
	for p >= 0 && (s.lines[line][p] == ' ' || s.lines[line][p] == '\t') {
		p--
	}
	switch {
	case p < 0:
		return true
	case strings.ContainsRune("[{,", s.lines[line][p]):
		return true
	}
	return p < column-1 && strings.ContainsRune("-?:", s.lines[line][p])
}

// anchorChar says whether a character is an ns-anchor-char of YAML 1.2.2.
func anchorChar(c rune) bool {
	return c != ' ' && c != '\t' && c != '\n' && c != 0 && c != byteOrderMark && !strings.ContainsRune(",[]{}", c)
}

// at records a breach at a place of go-yaml, whose line 0 is no place: the start of the file.
func (r *reader) at(rule string, mark goyaml.Mark, message string) {
	if mark.Line == 0 {
		r.breaches = append(r.breaches, Breach{rule, 1, 1, message})
		return
	}
	r.breach(rule, mark.Line-1, mark.Column-1, message)
}

// named says whether the anchor or the alias whose & or * stands at a place has a name, at least
// one ns-anchor-char.
func (s *source) named(line, column int) bool {
	c := s.at(line, column+1)
	return c != ' ' && c != '\t' && c != '\n' && c != 0 && !strings.ContainsRune(",[]{}", c)
}

// tag says whether the text at a place is a tag of YAML 1.2.2 with a blank or the end of a line
// after it. go-yaml refuses a tag in block context that a flow indicator follows, as YAML 1.2.2
// does, and reads one in a flow collection; so a flow indicator after a tag that go-yaml refused
// makes it no tag.
func (s *source) tag(line, column int) bool {
	at := column + 1
	// uri passes the ns-uri-char from at, or the ns-tag-char of a suffix, and says how many.
	uri := func(suffix bool) int {
		n := 0
		for ; ; n++ {
			c := s.at(line, at)
			switch {
			case c == '%':
				if !hexDigit(s.at(line, at+1)) || !hexDigit(s.at(line, at+2)) {
					return n
				}
				at += 3
			case wordCharacter(c) || strings.ContainsRune("#;/?:@&=+$_.~*'()", c) || !suffix && strings.ContainsRune(",[]!", c):
				at++
			default:
				return n
			}
		}
	}
	if s.at(line, at) == '<' {
		at++
		if uri(false) == 0 || s.at(line, at) != '>' {
			return false
		}
		at++
	} else {
		handle := at
		for wordCharacter(s.at(line, handle)) {
			handle++
		}
		if s.at(line, handle) == '!' {
			// A handle, !! or !name!, needs a suffix.
			at = handle + 1
		}
		if uri(true) == 0 && at != column+1 {
			return false
		}
	}
	c := s.at(line, at)
	return c == ' ' || c == '\t' || c == '\n' || c == 0
}

// handle gives the handle of the tag at a place: !, !! or !name!.
func (s *source) handle(line, column int) string {
	end := column + 1
	for wordCharacter(s.at(line, end)) {
		end++
	}
	if s.at(line, end) != '!' {
		return "!"
	}
	return string(s.lines[line][column : end+1])
}

// wordCharacter is ns-word-char of YAML 1.2.2.
func wordCharacter(c rune) bool {
	return c >= '0' && c <= '9' || c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c == '-'
}

func hexDigit(c rune) bool {
	return c >= '0' && c <= '9' || c >= 'a' && c <= 'f' || c >= 'A' && c <= 'F'
}
