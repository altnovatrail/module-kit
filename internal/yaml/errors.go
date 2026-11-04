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
// Each message is the one go-yaml gives at the version go.mod requires, and TestRead holds each: a
// version that words one otherwise fails it. Any other error is the file's not being YAML, at
// go-yaml's place.
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
	r.at("", mark, message)
}

// at records a breach at a place of go-yaml, or at the start of the file when go-yaml gives none.
func (r *reader) at(rule string, mark goyaml.Mark, message string) {
	if mark.Line == 0 {
		r.breaches = append(r.breaches, Breach{rule, 1, 1, message})
		return
	}
	r.breach(rule, mark.Line-1, mark.Column-1, message)
}

// named says whether the anchor or the alias whose & or * stands at a place has a name, at least
// one ns-anchor-char (YAML 1.2.2, 6.9.2).
func (s *source) named(line, column int) bool {
	c := s.at(line, column+1)
	return c != ' ' && c != '\t' && c != '\n' && c != 0 && !strings.ContainsRune(",[]{}", c)
}

// tag says whether the text at a place is a tag of YAML 1.2.2 with a space, a tab or the end of a
// line after it (6.8.2, 6.9.1): verbatim, !<…>; a shorthand, a handle and a suffix; or the
// non-specific !. go-yaml refuses a tag in block context that a flow indicator follows, which YAML
// 1.2.2 refuses there too, and reads one in a flow collection; so a flow indicator after a tag that
// go-yaml refused makes it no tag.
func (s *source) tag(line, column int) bool {
	at := column + 1
	// uri passes the characters of a URI from at, ns-uri-char, or of a tag's suffix, ns-tag-char,
	// which is no ! and no flow indicator, and says how many it passed.
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
			// !! or !name!: a handle, which needs a suffix.
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

// wordCharacter is ns-word-char: a digit, an ASCII letter or - (YAML 1.2.2, 5.6).
func wordCharacter(c rune) bool {
	return c >= '0' && c <= '9' || c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c == '-'
}

func hexDigit(c rune) bool {
	return c >= '0' && c <= '9' || c >= 'a' && c <= 'f' || c >= 'A' && c <= 'F'
}
