package template

import (
	"errors"
	"fmt"
	"slices"
	"strings"
)

// The type of an expression of a template: an int or a bool (format/README.md, section 23).
type exprType string

const (
	intType  exprType = "an int"
	boolType exprType = "a bool"
)

// The tokens of the expressions of templates that are symbols: those of two characters before those
// of one, so that the longer is read first, and the words and, or and not.
var (
	symbols     = []string{"==", "!=", "<=", ">=", "//", "<", ">", "+", "-", "*", "%", "(", ")"}
	comparisons = []string{"==", "!=", "<", "<=", ">", ">="}
	words       = []string{"and", "or", "not"}
)

// A token of an expression, at the place of its first character, from 1. A name or an integer is the
// longest run of the characters of a word.
type token struct {
	text    string
	at      int
	integer bool
	name    bool
}

func wordCharacter(c byte) bool { return c >= 'a' && c <= 'z' || c >= '0' && c <= '9' || c == '_' }
func digit(c byte) bool         { return c >= '0' && c <= '9' }

// tokensOf reads an expression into its tokens, which spaces, U+0020, or nothing separate
// (format/README.md, section 23). A character outside the tokens, a run that is neither an integer
// nor a name, and an integer with a leading zero are errors. Every character of a token is ASCII, so
// the place of a byte is the place of its character up to the first that is not, which is an error.
func tokensOf(src string) ([]token, error) {
	var found []token
	for at := 0; at < len(src); {
		c := src[at]
		switch {
		case c == ' ':
			at++
		case wordCharacter(c):
			end := at
			for end < len(src) && wordCharacter(src[end]) {
				end++
			}
			run := src[at:end]
			t := token{text: run, at: at + 1}
			switch {
			case !digit(run[0]):
				t.name = !slices.Contains(words, run)
			case strings.TrimLeft(run, "0123456789") != "":
				return nil, fmt.Errorf("%s at %d is neither an integer nor a name", run, at+1)
			case len(run) > 1 && run[0] == '0':
				return nil, fmt.Errorf("%s at %d: an integer has no leading zero", run, at+1)
			default:
				t.integer = true
			}
			found = append(found, t)
			at = end
		default:
			i := slices.IndexFunc(symbols, func(s string) bool { return strings.HasPrefix(src[at:], s) })
			if i < 0 {
				r := []rune(src[at:])[0]
				return nil, fmt.Errorf("%q at %d is no part of the expressions of templates", r, at+1)
			}
			found = append(found, token{text: symbols[i], at: at + 1})
			at += len(symbols[i])
		}
	}
	return found, nil
}

// typeOf reads src as an expression of the narrow subset of Starlark of templates and gives its
// type, when every name in it is one of names (format/README.md, section 23):
//
//	expression  = disjunction
//	disjunction = conjunction { "or" conjunction }
//	conjunction = negation { "and" negation }
//	negation    = "not" negation | comparison
//	comparison  = sum [ ( "==" | "!=" | "<" | "<=" | ">" | ">=" ) sum ]
//	sum         = product { ( "+" | "-" ) product }
//	product     = unary { ( "*" | "//" | "%" ) unary }
//	unary       = "-" unary | atom
//	atom        = integer | name | "(" expression ")"
//
// An integer and a name are ints; +, -, *, // and % take ints and give an int; a comparison takes
// two ints and gives a bool; not, and and or take bools and give a bool. The error says why src is
// no such expression, at the place of the token where that shows.
//
// An expression begins with a token, not a space, and nests at most maxDepth levels: each
// parenthesis, unary - and not is a level within the one around it, so -(a - (b)) nests three. The
// dialect reads a space before an expression as an indent, and go.starlark.net at the version of the
// format parses no deeper than about 75 parentheses, or 987 unary minus signs in a row, so every
// expression of the subset is one of the dialect (format/README.md, section 23; R102 spec; R32).
func typeOf(src string, names []string) (t exprType, err error) {
	if strings.HasPrefix(src, " ") {
		return "", errors.New("it begins with a space, where an expression begins with a token")
	}
	found, err := tokensOf(src)
	if err != nil {
		return "", err
	}
	p := &parser{tokens: found, names: names}
	defer func() {
		if r := recover(); r != nil {
			stop, ok := r.(problem)
			if !ok {
				panic(r)
			}
			t, err = "", stop
		}
	}()
	t = p.disjunction()
	if p.at < len(p.tokens) {
		p.fail("%s at %d stands where the expression has ended", p.tokens[p.at].text, p.tokens[p.at].at)
	}
	return t, nil
}

// A problem is what stops the parser: why the expression is none of the subset.
type problem string

func (p problem) Error() string { return string(p) }

// maxDepth is the most levels an expression of a template nests (format/README.md, section 23).
const maxDepth = 64

// parser reads the tokens of an expression by the grammar of the subset, and types each part as it
// reads it.
type parser struct {
	tokens []token
	at     int
	names  []string // the names that the expression may name
	depth  int      // the levels of parentheses, unary - and not around the token read
}

// enter goes a level deeper at the token t, a (, a unary - or a not, and fails past maxDepth; the
// caller leaves the level when it has read what the token holds.
func (p *parser) enter(t token) {
	p.depth++
	if p.depth > maxDepth {
		p.fail("%s at %d nests the expression %d levels deep, and it nests at most %d", t.text, t.at, p.depth, maxDepth)
	}
}

func (p *parser) fail(format string, args ...any) {
	panic(problem(fmt.Sprintf(format, args...)))
}

// peek gives the next token, or one of no text at the end.
func (p *parser) peek() token {
	if p.at < len(p.tokens) {
		return p.tokens[p.at]
	}
	return token{}
}

func (p *parser) next() token {
	t := p.peek()
	p.at++
	return t
}

// symbolAhead says whether the next token is one of the symbols ops. A name or an integer is never
// one: the words and, or and not are symbols, and no symbol is a run of the characters of a word.
func (p *parser) symbolAhead(ops []string) bool {
	return slices.Contains(ops, p.peek().text)
}

// want fails unless an operand of the operator op, at its place, has the type that op takes.
func (p *parser) want(op token, got, takes exprType) {
	if got != takes {
		p.fail("%s at %d takes %s, and its operand is %s", op.text, op.at, takes, got)
	}
}

// binary reads operands joined by any of ops, left to right, each of the type takes, and gives the
// type gives.
func (p *parser) binary(operand func() exprType, ops []string, takes, gives exprType) exprType {
	t := operand()
	for p.symbolAhead(ops) {
		op := p.next()
		p.want(op, t, takes)
		p.want(op, operand(), takes)
		t = gives
	}
	return t
}

func (p *parser) disjunction() exprType {
	return p.binary(p.conjunction, []string{"or"}, boolType, boolType)
}

func (p *parser) conjunction() exprType {
	return p.binary(p.negation, []string{"and"}, boolType, boolType)
}

func (p *parser) negation() exprType {
	if !p.symbolAhead([]string{"not"}) {
		return p.comparison()
	}
	op := p.next()
	p.enter(op)
	p.want(op, p.negation(), boolType)
	p.depth--
	return boolType
}

// comparison reads a sum, compared with one more at most: comparisons do not chain.
func (p *parser) comparison() exprType {
	t := p.sum()
	if !p.symbolAhead(comparisons) {
		return t
	}
	op := p.next()
	p.want(op, t, intType)
	p.want(op, p.sum(), intType)
	if p.symbolAhead(comparisons) {
		p.fail("%s at %d and %s at %d: comparisons do not chain", op.text, op.at, p.peek().text, p.peek().at)
	}
	return boolType
}

func (p *parser) sum() exprType {
	return p.binary(p.product, []string{"+", "-"}, intType, intType)
}

func (p *parser) product() exprType {
	return p.binary(p.unary, []string{"*", "//", "%"}, intType, intType)
}

func (p *parser) unary() exprType {
	if !p.symbolAhead([]string{"-"}) {
		return p.atom()
	}
	op := p.next()
	p.enter(op)
	p.want(op, p.unary(), intType)
	p.depth--
	return intType
}

func (p *parser) atom() exprType {
	t := p.next()
	switch {
	case t.text == "":
		p.fail("it ends where an operand is due")
	case t.integer:
		return intType
	case t.name:
		if !slices.Contains(p.names, t.text) {
			p.fail("%s at %d is none of the names that it may name: %s", t.text, t.at, strings.Join(p.names, ", "))
		}
		return intType
	case t.text != "(":
		p.fail("%s at %d stands where an operand is due", t.text, t.at)
	}
	p.enter(t)
	inner := p.disjunction()
	p.depth--
	if !p.symbolAhead([]string{")"}) {
		p.fail("the ( at %d is not closed", t.at)
	}
	p.next()
	return inner
}
