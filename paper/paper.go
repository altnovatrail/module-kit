package paper

import (
	"math/big"
	"regexp"
	"slices"
	"unicode/utf8"

	observationv1 "github.com/altnovatrail/spec/contracts/novatrail/observation/v1"
	paperv1 "github.com/altnovatrail/spec/contracts/novatrail/paper/v1"
	"github.com/altnovatrail/spec/types/formatv0"
	"google.golang.org/protobuf/proto"

	"github.com/altnovatrail/module-kit/internal/normal"
)

// Reason is why the solver does not take a formula, as the contracts text names it.
type Reason string

// The reasons, in the order the solver names the first that applies: syntax, then the letters,
// then those of the operations as they are done, then the root, and last too_large of the answer.
const (
	Syntax              Reason = "syntax"                // not of the grammar, too long, a number of more than 15 digits or an exponent above 10
	TwoUnknowns         Reason = "two_unknowns"          // two letters or more
	UnknownInExpression Reason = "unknown_in_expression" // an expression with an unknown
	NoUnknown           Reason = "no_unknown"            // an equation with no unknown
	NotLinear           Reason = "not_linear"            // the unknown times itself, in a divisor, or to 2 or more
	DivisionByZero      Reason = "division_by_zero"      // a division by zero
	TooLarge            Reason = "too_large"             // a part of more than 100 digits on the way, or an answer longer than 16 characters
	NoRoot              Reason = "no_root"               // an equation no number solves
	EveryRoot           Reason = "every_root"            // an equation every number solves
)

// Rule is a rule of an answer of AI that the answer breaks, as the vectors name it.
type Rule string

// The rules of an answer, in the order Breach checks them.
const (
	NoKind       Rule = "no_kind"       // neither a number nor a choice
	Value        Rule = "value"         // a number not in the notation of number, or longer than 16 characters
	UnitsEmpty   Rule = "units_empty"   // a unit empty once it went through the steps of units
	UnitsSame    Rule = "units_same"    // two units the same then
	Unknown      Rule = "unknown"       // an unknown that is not one letter from a to z
	OptionsCount Rule = "options_count" // fewer than 2 options, or more than 10
	OptionLength Rule = "option_length" // a label of fewer than 1 character or more than 16
	OptionsSame  Rule = "options_same"  // a label empty, or two the same, once normalized as a choice
	ChoiceKey    Rule = "choice_key"    // a key that is none of the options as written
)

const (
	maxFormula  = 200
	maxDigits   = 15
	maxExponent = 10
	maxAnswer   = 16
	minOptions  = 2
	maxOptions  = 10
)

// limit bounds a numerator and a denominator on the way: a part of 10^100 or more has more than
// 100 digits.
var limit = new(big.Int).Exp(big.NewInt(10), big.NewInt(100), nil)

// linear is a part a·x + b of a formula, its coefficients exact.
type linear struct{ a, b *big.Rat }

func constant(b *big.Rat) linear { return linear{new(big.Rat), b} }

// bounded refuses a part whose numerator or denominator has more than 100 digits.
func bounded(p linear) (linear, Reason) {
	for _, r := range []*big.Rat{p.a, p.b} {
		if r.Num().CmpAbs(limit) >= 0 || r.Denom().Cmp(limit) >= 0 {
			return linear{}, TooLarge
		}
	}
	return p, ""
}

// Solve gives the answer of the solver on a formula, a Number with its value and the letter of an
// equation's unknown, or the reason it does not take it.
func Solve(formula string) (*paperv1.Number, Reason) {
	t, ok := parse(formula)
	if !ok {
		return nil, Syntax
	}
	if len(t.letters) > 1 {
		return nil, TwoUnknowns
	}
	letter := ""
	for l := range t.letters {
		letter = l
	}
	if t.right == nil && letter != "" {
		return nil, UnknownInExpression
	}
	if t.right != nil && letter == "" {
		return nil, NoUnknown
	}
	value, reason := root(t)
	if reason != "" {
		return nil, reason
	}
	text := written(value)
	if utf8.RuneCountInString(text) > maxAnswer {
		return nil, TooLarge
	}
	return &paperv1.Number{Value: text, Unknown: letter}, ""
}

// root gives the value of an expression, or the one root of an equation, or why there is none.
func root(t *tree) (*big.Rat, Reason) {
	left, reason := evaluate(t.left)
	if reason != "" {
		return nil, reason
	}
	if t.right == nil {
		return left.b, ""
	}
	right, reason := evaluate(t.right)
	if reason != "" {
		return nil, reason
	}
	// The sides brought together are a step on the way too, bounded as any.
	side, reason := bounded(linear{new(big.Rat).Sub(left.a, right.a), new(big.Rat).Sub(right.b, left.b)})
	if reason != "" {
		return nil, reason
	}
	a, b := side.a, side.b
	if a.Sign() == 0 {
		if b.Sign() == 0 {
			return nil, EveryRoot
		}
		return nil, NoRoot
	}
	return b.Quo(b, a), ""
}

// written is a rational as the solver writes its answer: the integer, or p/q in lowest terms with
// its sign before p.
func written(r *big.Rat) string {
	if r.IsInt() {
		return r.Num().String()
	}
	return r.Num().String() + "/" + r.Denom().String()
}

// evaluate gives the value of a node as a linear part, each part before its operation and the left
// before the right, so that the first refusal of the operations is the one named.
func evaluate(n *node) (linear, Reason) {
	switch n.kind {
	case number:
		return bounded(constant(n.value))
	case unknown:
		return linear{big.NewRat(1, 1), new(big.Rat)}, ""
	case negation:
		of, reason := evaluate(n.left)
		if reason != "" {
			return linear{}, reason
		}
		return linear{new(big.Rat).Neg(of.a), new(big.Rat).Neg(of.b)}, ""
	case power:
		return raised(n)
	}
	left, reason := evaluate(n.left)
	if reason != "" {
		return linear{}, reason
	}
	right, reason := evaluate(n.right)
	if reason != "" {
		return linear{}, reason
	}
	switch n.kind {
	case plus:
		return bounded(linear{new(big.Rat).Add(left.a, right.a), new(big.Rat).Add(left.b, right.b)})
	case minus:
		return bounded(linear{new(big.Rat).Sub(left.a, right.a), new(big.Rat).Sub(left.b, right.b)})
	case times:
		if left.a.Sign() != 0 && right.a.Sign() != 0 {
			return linear{}, NotLinear
		}
		part, factor := left, right.b
		if left.a.Sign() == 0 {
			part, factor = right, left.b
		}
		return bounded(linear{new(big.Rat).Mul(part.a, factor), new(big.Rat).Mul(part.b, factor)})
	}
	if right.a.Sign() != 0 {
		return linear{}, NotLinear
	}
	if right.b.Sign() == 0 {
		return linear{}, DivisionByZero
	}
	return bounded(linear{new(big.Rat).Quo(left.a, right.b), new(big.Rat).Quo(left.b, right.b)})
}

// raised gives a power: of what holds the unknown, itself for 1, 1 for 0 and not linear from 2;
// of a number, its product step by step, each step bounded.
func raised(n *node) (linear, Reason) {
	base, reason := evaluate(n.left)
	if reason != "" {
		return linear{}, reason
	}
	if base.a.Sign() != 0 {
		switch n.exponent {
		case 0:
			return constant(big.NewRat(1, 1)), ""
		case 1:
			return base, ""
		}
		return linear{}, NotLinear
	}
	value := constant(big.NewRat(1, 1))
	for range n.exponent {
		if value, reason = bounded(constant(new(big.Rat).Mul(value.b, base.b))); reason != "" {
			return linear{}, reason
		}
	}
	return value, ""
}

// Breach gives the first rule of an answer of AI that it breaks, or "" when it is an answer, which
// a key may hold.
func Breach(answer *paperv1.Answer) Rule {
	if n := answer.GetNumber(); n != nil {
		// The notation of the check kind number, which judges the reading by the key.
		if _, ok := normal.Notation(n.GetValue(), formatv0.CheckKindNumber); !ok {
			return Value
		}
		// Every unit is read for an empty one before any two are compared: the rules' order.
		forms := make([]string, 0, len(n.GetUnits()))
		for _, unit := range n.GetUnits() {
			form := normal.Unit(unit)
			if form == "" {
				return UnitsEmpty
			}
			forms = append(forms, form)
		}
		for i, form := range forms {
			if slices.Contains(forms[:i], form) {
				return UnitsSame
			}
		}
		if u := n.GetUnknown(); u != "" && (len(u) != 1 || u[0] < 'a' || u[0] > 'z') {
			return Unknown
		}
		return ""
	}
	if c := answer.GetChoice(); c != nil {
		options := c.GetOptions()
		if len(options) < minOptions || len(options) > maxOptions {
			return OptionsCount
		}
		for _, option := range options {
			if n := utf8.RuneCountInString(option); n < 1 || n > maxAnswer {
				return OptionLength
			}
		}
		forms := make([]string, 0, len(options))
		for _, option := range options {
			form := normal.Choice(option)
			if form == "" || slices.Contains(forms, form) {
				return OptionsSame
			}
			forms = append(forms, form)
		}
		if !slices.Contains(options, c.GetKey()) {
			return ChoiceKey
		}
		return ""
	}
	return NoKind
}

// Key gives the key of one task of an attempt, with no statement, which the caller sets: the
// solver's answer with the provenance of the solver, when there is one, or else the answer of AI
// with the provenance of AI when it breaks no rule. With neither the key has no answer and no
// provenance, which says the task has none.
func Key(solved *paperv1.Number, ofAI *paperv1.Answer) *paperv1.TaskKey {
	if solved != nil {
		return &paperv1.TaskKey{Provenance: paperv1.Provenance_PROVENANCE_SOLVER, Answer: &paperv1.Answer{Kind: &paperv1.Answer_Number{Number: solved}}}
	}
	if ofAI != nil && Breach(ofAI) == "" {
		return &paperv1.TaskKey{Provenance: paperv1.Provenance_PROVENANCE_AI, Answer: ofAI}
	}
	return &paperv1.TaskKey{}
}

// Items gives the task's one item by the answer of its key: one answer slot bound anywhere, with
// the check kind number or choice written by hand, and no trap values. With them, the package
// observation joins and judges the readings of the work as on a task of a module, once Brought has
// brought their values to the key.
//
// An answer of no kind, nil too, is a task with no key, and gives no item: no check can be written
// of it, and the package observation judges none. Its item is unchecked, which the caller gives it
// without the package observation.
func Items(answer *paperv1.Answer) []formatv0.Item {
	var check formatv0.Check
	if n := answer.GetNumber(); n != nil {
		value := n.GetValue()
		check = formatv0.Check{Kind: formatv0.CheckKindNumber, Key: &value}
		for _, unit := range n.GetUnits() {
			check.Units = append(check.Units, formatv0.Text(unit))
		}
	} else if c := answer.GetChoice(); c != nil {
		key := c.GetKey()
		options := formatv0.CheckNotationOptions
		check = formatv0.Check{Kind: formatv0.CheckKindChoice, Key: &key, Notation: &options}
		for _, option := range c.GetOptions() {
			check.Options = append(check.Options, formatv0.Text(option))
		}
	} else {
		return nil
	}
	return []formatv0.Item{{AnswerSlots: []formatv0.AnswerSlot{{Binding: formatv0.AnswerSlotBindingAnywhere, Check: check}}}}
}

var (
	// A value that begins with a letter, an optional space and =, after steps 1 to 4.
	ofUnknown = regexp.MustCompile(`^(\pL) ?=`)
	// A label in brackets: (C), [C], or C) with a bracket at its end alone.
	inBrackets = []*regexp.Regexp{regexp.MustCompile(`^\((.+)\)$`), regexp.MustCompile(`^\[(.+)\]$`), regexp.MustCompile(`^([^()[\]]+)\)$`)}
)

// BroughtValue gives a value read brought to the key: x = 4 as 4 when the key's unknown is x, and
// (C) as C of a choice; any other value as it is.
func BroughtValue(answer *paperv1.Answer, value string) string {
	if n := answer.GetNumber(); n != nil && n.GetUnknown() != "" {
		text := normal.Text(normal.Scripts(value))
		if m := ofUnknown.FindStringSubmatch(text); m != nil && normal.Lower(m[1]) == n.GetUnknown() {
			return text[len(m[0])-1:]
		}
		return value
	}
	if answer.GetChoice() != nil {
		text := normal.Text(value)
		for _, re := range inBrackets {
			if m := re.FindStringSubmatch(text); m != nil {
				return m[1]
			}
		}
	}
	return value
}

// Brought gives a copy of an observation with each value brought to the key, so that the join of
// two readings compares x = 4 and 4 as the same value.
func Brought(answer *paperv1.Answer, o *observationv1.Observation) *observationv1.Observation {
	brought := proto.Clone(o).(*observationv1.Observation)
	for _, slot := range brought.GetAnswerSlots() {
		if v, ok := slot.GetRead().(*observationv1.AnswerSlotObservation_Value); ok {
			v.Value = BroughtValue(answer, v.Value)
		}
	}
	return brought
}
