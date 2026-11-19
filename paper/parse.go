package paper

import (
	"math/big"
	"strings"
)

// kind is the kind of a node of a formula.
type kind int

const (
	number kind = iota
	unknown
	negation
	power
	plus
	minus
	times
	divide
)

// node is a node of a formula: a number with its value, the unknown, or an operation on left, and
// right too but for a negation and a power, whose exponent it holds.
type node struct {
	kind        kind
	value       *big.Rat
	left, right *node
	exponent    int
}

// tree is a formula parsed: its left side, its right side for an equation, and its letters.
type tree struct {
	left, right *node
	letters     map[string]bool
}

// token is a token of a formula: a number, with its digits before and after the point, a letter,
// or a sign.
type token struct {
	sign        byte
	whole, part string
	letter      string
}

// tokens gives the tokens of a formula, or false when a character is none of the grammar's or a
// number has more than 15 digits.
func tokens(formula string) ([]token, bool) {
	var out []token
	for at := 0; at < len(formula); {
		c := formula[at]
		switch {
		case c == ' ':
			at++
		case isDigit(c):
			start := at
			for at < len(formula) && isDigit(formula[at]) {
				at++
			}
			t := token{sign: 'n', whole: formula[start:at]}
			if at+1 < len(formula) && formula[at] == '.' && isDigit(formula[at+1]) {
				at++
				begin := at
				for at < len(formula) && isDigit(formula[at]) {
					at++
				}
				t.part = formula[begin:at]
			}
			if len(t.whole)+len(t.part) > maxDigits {
				return nil, false
			}
			out = append(out, t)
		case c >= 'a' && c <= 'z':
			out = append(out, token{sign: 'x', letter: string(c)})
			at++
		case strings.IndexByte("+-*/^()=", c) >= 0:
			out = append(out, token{sign: c})
			at++
		default:
			return nil, false
		}
	}
	return out, true
}

func isDigit(c byte) bool { return c >= '0' && c <= '9' }

// parser reads tokens by the grammar of a formula (R69).
type parser struct {
	tokens []token
	at     int
	failed bool
}

func (p *parser) peek() byte {
	if p.at < len(p.tokens) {
		return p.tokens[p.at].sign
	}
	return 0
}

func (p *parser) take(sign byte) bool {
	if p.peek() == sign {
		p.at++
		return true
	}
	return false
}

// parse gives the tree of a formula, or false when it is not of the grammar: longer than 200 bytes,
// a character outside it, a number of more than 15 digits or an exponent above 10.
func parse(formula string) (*tree, bool) {
	if len(formula) > maxFormula {
		return nil, false
	}
	ts, ok := tokens(formula)
	if !ok {
		return nil, false
	}
	p := &parser{tokens: ts}
	t := &tree{left: p.expression(), letters: map[string]bool{}}
	if p.take('=') {
		t.right = p.expression()
	}
	if p.failed || p.at != len(ts) {
		return nil, false
	}
	for _, one := range ts {
		if one.sign == 'x' {
			t.letters[one.letter] = true
		}
	}
	return t, true
}

func (p *parser) expression() *node {
	left := p.term()
	for {
		switch {
		case p.take('+'):
			left = &node{kind: plus, left: left, right: p.term()}
		case p.take('-'):
			left = &node{kind: minus, left: left, right: p.term()}
		default:
			return left
		}
	}
}

func (p *parser) term() *node {
	left := p.factor()
	for {
		switch {
		case p.take('*'):
			left = &node{kind: times, left: left, right: p.factor()}
		case p.take('/'):
			left = &node{kind: divide, left: left, right: p.factor()}
		default:
			return left
		}
	}
}

func (p *parser) factor() *node {
	if p.take('-') {
		return &node{kind: negation, left: p.factor()}
	}
	base := p.primary()
	if !p.take('^') {
		return base
	}
	if p.peek() != 'n' {
		p.failed = true
		return base
	}
	t := p.tokens[p.at]
	p.at++
	exponent, ok := new(big.Int).SetString(t.whole, 10)
	if t.part != "" || !ok || exponent.Cmp(big.NewInt(maxExponent)) > 0 {
		p.failed = true
		return base
	}
	return &node{kind: power, left: base, exponent: int(exponent.Int64())}
}

func (p *parser) primary() *node {
	if p.failed {
		return &node{kind: number, value: new(big.Rat)}
	}
	switch p.peek() {
	case 'n':
		t := p.tokens[p.at]
		p.at++
		digits, _ := new(big.Int).SetString(t.whole+t.part, 10)
		scale := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(len(t.part))), nil)
		return &node{kind: number, value: new(big.Rat).SetFrac(digits, scale)}
	case 'x':
		p.at++
		return &node{kind: unknown}
	case '(':
		p.at++
		inner := p.expression()
		if !p.take(')') {
			p.failed = true
		}
		return inner
	}
	p.failed = true
	return &node{kind: number, value: new(big.Rat)}
}
