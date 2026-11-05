package verdict

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/altnovatrail/spec/types/formatv0"
	"go.starlark.net/starlark"
	"golang.org/x/text/unicode/norm"

	"github.com/altnovatrail/module-kit/internal/sandbox"
)

// Verdict is the verdict of an answer slot or of an item, as the format writes it
// (format/README.md, section 15).
type Verdict string

// The verdicts.
const (
	Correct    Verdict = "correct"    // the value is right
	NotYet     Verdict = "not_yet"    // the value is in the kind's notation and is not right
	Unreadable Verdict = "unreadable" // the learner is asked to write the answer again, which is no wrong answer
	Unchecked  Verdict = "unchecked"  // the check could not run
)

// Reason is why an answer slot is unreadable, as the format and the contracts write it: blank and
// notation, which a check kind gives itself (format/README.md, section 15), and the reasons that a
// reading gives before the value is judged, each named after a reason of the observation
// (contracts/README.md, section 10).
type Reason string

// The reasons that a check kind gives.
const (
	Blank    Reason = "blank"    // nothing was read, or nothing is left once the value is normalized
	Notation Reason = "notation" // the value is not in the kind's notation
)

// The reasons that a reading gives, by the table of contracts/README.md, section 10, which the
// package observation reads them by.
const (
	Illegible           Reason = "illegible"             // something is written, and it cannot be read
	Ambiguous           Reason = "ambiguous"             // it can be read more than one way, or two readers read it differently
	NoFinalAnswer       Reason = "no_final_answer"       // in a slot bound anywhere: the work holds no final answer
	SeveralFinalAnswers Reason = "several_final_answers" // in a slot bound anywhere: the work holds more than one
)

// UnicodeVersion is the version of Unicode that the normalization follows: that of the tables of
// NFKC the library is built with, which its tests hold equal to those of the standard library's
// package unicode. The format leaves the version to the reader's release, and this is where a
// release of module-kit names it (format/README.md, section 15; R27).
const UnicodeVersion = norm.Version

// ErrNoVerdict is what the check of a rubric gives: a rubric gives feedback by its criteria, and the
// item it checks has no verdict.
var ErrNoVerdict = errors.New("a rubric gives no verdict")

// ErrPlaceholder is what the check of a number or a fraction gives when its key or a trap value
// holds the placeholder of a template: the check is a template's, and an instance of it replaces
// the placeholder with a number (format/README.md, section 23).
var ErrPlaceholder = errors.New("the check holds the placeholder of a template, which an instance replaces with a number")

// valueName is the one name that a predicate is written on, the whole number read
// (format/README.md, section 15).
const valueName = "value"

// AnswerSlot is an answer slot of an item as its check reads it: its binding as the format writes
// it — anywhere, sheet, zone, choice or box (format/README.md, section 14) — its check, and, when it
// is bound choice, its options, of which the learner taps one.
type AnswerSlot struct {
	Binding string
	Check   formatv0.Check
	Options []formatv0.Text
}

// AnswerSlots gives the answer slots of an item, in their order: on a handwriting sheet its
// answer_slots, on a workbook the boxes of its line (format/README.md, section 14).
func AnswerSlots(item formatv0.Item) []AnswerSlot {
	var slots []AnswerSlot
	for _, slot := range item.AnswerSlots {
		s := AnswerSlot{Binding: string(slot.Binding), Check: slot.Check}
		if slot.Binding == formatv0.AnswerSlotBindingChoice {
			s.Options = slot.Options
		}
		slots = append(slots, s)
	}
	for _, part := range item.Line {
		if part.Binding != nil {
			s := AnswerSlot{Binding: string(*part.Binding)}
			if part.Check != nil {
				s.Check = *part.Check
			}
			slots = append(slots, s)
		}
	}
	return slots
}

// SlotVerdict is the verdict of an answer slot: with its reason when it is Unreadable, and with the
// trap value that the value matches, a copy of it, when it is NotYet and matches one — the trap it
// shows, and the text that the learner reads after it, if the author wrote one (format/README.md,
// sections 15 and 19).
type SlotVerdict struct {
	Verdict Verdict
	Reason  Reason
	Trap    *formatv0.TrapValue
}

// ItemVerdict is the verdict of an item, made of the verdicts of its answer slots, which it holds in
// their order.
type ItemVerdict struct {
	Verdict Verdict
	Slots   []SlotVerdict
}

// OfItem gives the verdict of an item on the values read in its answer slots, one for each, in the
// order of AnswerSlots: a value as it is written, or nil when nothing was read (format/README.md,
// section 15). An error gives no verdict: ctx is done, the item has no answer slots, the values are
// not one for each answer slot, or OfSlot gives one for an answer slot — ErrNoVerdict for an item
// that a rubric checks, and the error of ctx when it stops a predicate as it runs.
func OfItem(ctx context.Context, item formatv0.Item, values []*string) (ItemVerdict, error) {
	if err := ctx.Err(); err != nil {
		return ItemVerdict{}, fmt.Errorf("verdict: %w", err)
	}
	slots := AnswerSlots(item)
	if len(slots) == 0 {
		return ItemVerdict{}, errors.New("verdict: the item has no answer slots")
	}
	if len(values) != len(slots) {
		return ItemVerdict{}, fmt.Errorf("verdict: %d values for the %d answer slots of the item", len(values), len(slots))
	}
	verdicts := make([]SlotVerdict, len(slots))
	for i, slot := range slots {
		v, err := slot.verdict(ctx, values[i])
		if err != nil {
			return ItemVerdict{}, fmt.Errorf("verdict: answer slot %d: %w", i+1, err)
		}
		verdicts[i] = v
	}
	return ItemVerdict{Verdict: ItemOf(verdicts), Slots: verdicts}, nil
}

// OfSlot gives the verdict of an answer slot on the value read in it: the value as it is written, or
// nil when nothing was read (format/README.md, section 15). The value and the trap values are
// normalized as the check kind normalizes them, and the key is read as it is written, in the kind's
// notation. A predicate on a value in its notation runs in the sandbox of Starlark, with the limits
// of a predicate and ctx (format/README.md, section 15; R11): True is correct and False not_yet,
// and an error, a result that is not a bool, the limit of steps, module-kit's clock and a panic of
// the interpreter are unchecked.
//
// An error gives no verdict: ctx is done before the verdict, or stops a predicate as it runs — a
// predicate that its limit of steps, the clock or a panic stopped first, or that ended, gives its
// verdict; the value is not UTF-8; the check is one that the format does not take — its kind is
// none of the five, a number, a fraction or a choice has no key, the key of a number or a fraction
// is not in its notation, it or a trap value holds the placeholder of a template, ErrPlaceholder, a
// choice has no options, or a predicate no expression; or the check is a rubric, ErrNoVerdict. A
// check that breaks a rule of the format that the check sees itself — a key that is none of the
// options, two options alike, a key of a fraction in lowest terms that can be reduced, a trap value
// that is not wrong, a predicate that is no expression of the dialect — still gives its verdict, so
// that the check of a module reads those rules by it.
func OfSlot(ctx context.Context, slot AnswerSlot, value *string) (SlotVerdict, error) {
	if err := ctx.Err(); err != nil {
		return SlotVerdict{}, fmt.Errorf("verdict: %w", err)
	}
	v, err := slot.verdict(ctx, value)
	if err != nil {
		return SlotVerdict{}, fmt.Errorf("verdict: %w", err)
	}
	return v, nil
}

// Same says whether the check of an answer slot compares two values equal, as it compares a value
// with a trap value, after the same normalization (format/README.md, section 15): for a choice, the
// same normal form, an option or not, empty too; for a number and a fraction, two values in the
// kind's notation, of one value; for a predicate, the same integer. A rubric compares nothing, and a
// value that is not UTF-8 is equal to none.
//
// A value out of a number's notation is equal to nothing here. The joining of two readings joins
// such values by their normal forms, NormalForm, and keeps apart a value out of the notation from
// one in it, which its normal form can spell, as 1/100000000000000 does (contracts/README.md,
// section 10).
func Same(slot AnswerSlot, a, b string) bool {
	if !utf8.ValidString(a) || !utf8.ValidString(b) {
		return false
	}
	units := unitsOf(slot.Check)
	x, xok := comparedBy(slot.Check, units, a)
	y, yok := comparedBy(slot.Check, units, b)
	return xok && yok && x == y
}

// NormalForm gives a value as the check of an answer slot normalizes it, and whether it is in the
// kind's notation, with no Starlark (format/README.md, section 15): for a choice, its normal form,
// in the notation when it is the key's or an option's; for a number, a fraction and a predicate,
// steps 1 to 7, with a number's units dropped, in the notation when the kind takes what is left. An
// empty form is blank, and in no notation. A rubric normalizes nothing: the form is the value as it
// is, in no notation, and so is a value that is not UTF-8. The joining of two readings compares the
// values out of the notation by it (contracts/README.md, section 10).
func NormalForm(slot AnswerSlot, value string) (form string, inNotation bool) {
	if !utf8.ValidString(value) {
		return value, false
	}
	c := slot.Check
	switch c.Kind {
	case formatv0.CheckKindChoice:
		form = choiceForm(value)
		if form == "" {
			return form, false
		}
		if c.Key != nil && choiceForm(*c.Key) == form {
			return form, true
		}
		for _, option := range slot.ChoiceOptions() {
			if choiceForm(string(option)) == form {
				return form, true
			}
		}
		return form, false
	case formatv0.CheckKindNumber, formatv0.CheckKindFraction, formatv0.CheckKindPredicate:
		form = numberText(value, unitsOf(c))
		_, ok := notation(form, c.Kind)
		return form, ok
	}
	return value, false
}

// comparedBy gives what a check compares a value by, and whether it compares it at all: a choice,
// its normal form, always; a number, a fraction and a predicate, its value in the kind's notation,
// exactly, written in lowest terms. units are the forms of a number's units.
func comparedBy(c formatv0.Check, units []string, text string) (string, bool) {
	switch c.Kind {
	case formatv0.CheckKindChoice:
		return choiceForm(text), true
	case formatv0.CheckKindNumber, formatv0.CheckKindFraction, formatv0.CheckKindPredicate:
		if w, ok := notation(numberText(text, units), c.Kind); ok {
			return w.value.RatString(), true
		}
	}
	return "", false
}

// verdict gives the verdict of the answer slot on a value, as OfSlot does, with an error that names
// no package.
func (s AnswerSlot) verdict(ctx context.Context, value *string) (SlotVerdict, error) {
	c := s.Check
	if value != nil && !utf8.ValidString(*value) {
		return SlotVerdict{}, errors.New("the value read is not UTF-8")
	}
	switch c.Kind {
	case formatv0.CheckKindNumber, formatv0.CheckKindFraction:
		return s.numberVerdict(value)
	case formatv0.CheckKindChoice:
		return s.choiceVerdict(value)
	case formatv0.CheckKindPredicate:
		return s.predicateVerdict(ctx, sandbox.Predicate, value)
	case formatv0.CheckKindRubric:
		return SlotVerdict{}, ErrNoVerdict
	}
	return SlotVerdict{}, fmt.Errorf("the check kind %q is none of the format's", c.Kind)
}

// predicateVerdict gives the verdict of a predicate: blank, out of the notation — an integer alone
// is in it —, or what the expression gives on the integer read, a Starlark int bound to value, in
// the sandbox with the limits and ctx. True is correct and False not_yet; an error, a result that is
// not a bool, the limit of steps, the clock and a panic are unchecked, and ctx done is an error
// (format/README.md, section 15; R11).
func (s AnswerSlot) predicateVerdict(ctx context.Context, limits sandbox.Limits, value *string) (SlotVerdict, error) {
	c := s.Check
	if c.Predicate == nil {
		return SlotVerdict{}, errors.New("the predicate check has no predicate")
	}
	if value == nil {
		return SlotVerdict{Verdict: Unreadable, Reason: Blank}, nil
	}
	text := numberText(*value, nil)
	if text == "" {
		return SlotVerdict{Verdict: Unreadable, Reason: Blank}, nil
	}
	read, ok := notation(text, c.Kind)
	if !ok {
		return SlotVerdict{Verdict: Unreadable, Reason: Notation}, nil
	}
	result, err := sandbox.Eval(ctx, limits, *c.Predicate, starlark.StringDict{valueName: starlark.MakeBigInt(read.value.Num())})
	switch {
	case errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded):
		return SlotVerdict{}, err
	case err != nil:
		return SlotVerdict{Verdict: Unchecked}, nil
	case result == starlark.True:
		return SlotVerdict{Verdict: Correct}, nil
	case result == starlark.False:
		return SlotVerdict{Verdict: NotYet}, nil
	}
	return SlotVerdict{Verdict: Unchecked}, nil
}

// CheckPredicate says whether the expression of a predicate is one of the dialect of all the
// Starlark of the format, which names nothing but value, the names that it binds itself and the
// built-in names of the dialect: the rule task.predicate-expression (format/README.md, sections 15
// and 24). The error says why not, at the line and the column where go.starlark.net finds it. A
// predicate that is not gives unchecked on every value in its notation.
func CheckPredicate(expression string) error {
	return sandbox.CheckExpr(expression, valueName)
}

// numberVerdict gives the verdict of a number or a fraction: blank, out of the notation, a value
// other than the key's, a fraction of the key's value that can be reduced under form: reduced, or
// right.
func (s AnswerSlot) numberVerdict(value *string) (SlotVerdict, error) {
	c := s.Check
	if c.Key == nil {
		return SlotVerdict{}, fmt.Errorf("the %s check has no key", c.Kind)
	}
	written := []string{*c.Key}
	for _, trapValue := range c.TrapValues {
		written = append(written, trapValue.Value)
	}
	for _, w := range written {
		if strings.Contains(w, "{") {
			return SlotVerdict{}, fmt.Errorf("%w: %q", ErrPlaceholder, w)
		}
	}
	key, ok := notation(*c.Key, c.Kind)
	if !ok {
		return SlotVerdict{}, fmt.Errorf("the key %q is not in the notation of %s", *c.Key, c.Kind)
	}
	if value == nil {
		return SlotVerdict{Verdict: Unreadable, Reason: Blank}, nil
	}
	text := numberText(*value, unitsOf(c))
	if text == "" {
		return SlotVerdict{Verdict: Unreadable, Reason: Blank}, nil
	}
	read, ok := notation(text, c.Kind)
	switch {
	case !ok:
		return SlotVerdict{Verdict: Unreadable, Reason: Notation}, nil
	case read.value.Cmp(key.value) != 0, c.Kind == formatv0.CheckKindFraction && reducedForm(c) && !read.lowest:
		return s.notYet(*value), nil
	}
	return SlotVerdict{Verdict: Correct}, nil
}

// choiceVerdict gives the verdict of a choice: blank, the key, another option, or none of the
// options, which is out of the notation (format/README.md, section 15).
func (s AnswerSlot) choiceVerdict(value *string) (SlotVerdict, error) {
	c := s.Check
	if c.Key == nil {
		return SlotVerdict{}, errors.New("the choice check has no key")
	}
	options := s.ChoiceOptions()
	if len(options) == 0 {
		return SlotVerdict{}, errors.New("the choice check has no options")
	}
	if value == nil {
		return SlotVerdict{Verdict: Unreadable, Reason: Blank}, nil
	}
	read := choiceForm(*value)
	switch {
	case read == "":
		return SlotVerdict{Verdict: Unreadable, Reason: Blank}, nil
	case read == choiceForm(*c.Key):
		return SlotVerdict{Verdict: Correct}, nil
	}
	for _, option := range options {
		if choiceForm(string(option)) == read {
			return s.notYet(*value), nil
		}
	}
	return SlotVerdict{Verdict: Unreadable, Reason: Notation}, nil
}

// ChoiceOptions gives the options of a choice as written: the answer slot's when it is bound
// choice, and the check's own otherwise (format/README.md, section 15).
func (s AnswerSlot) ChoiceOptions() []formatv0.Text {
	if s.Binding == string(formatv0.AnswerSlotBindingChoice) {
		return s.Options
	}
	return s.Check.Options
}

// notYet gives NotYet with the first trap value that the check compares equal to the value: a value
// matches one at most, since no two trap values of a check are equal (task.trap-values).
func (s AnswerSlot) notYet(value string) SlotVerdict {
	units := unitsOf(s.Check)
	read, _ := comparedBy(s.Check, units, value)
	for _, trapValue := range s.Check.TrapValues {
		if compared, ok := comparedBy(s.Check, units, trapValue.Value); ok && compared == read {
			return SlotVerdict{Verdict: NotYet, Trap: &trapValue}
		}
	}
	return SlotVerdict{Verdict: NotYet}
}

// reducedForm says whether a fraction check takes only a fraction in lowest terms.
func reducedForm(c formatv0.Check) bool {
	return c.Form != nil && *c.Form == formatv0.CheckFormReduced
}

// ItemOf gives the verdict of an item, made of its answer slots' verdicts: unchecked when there are
// none, since nothing was checked; correct when every one is; otherwise not_yet when one is, since a
// value read clearly and wrong is no doubt; otherwise unreadable when one is; otherwise unchecked
// (format/README.md, section 15). The verdict on an observation makes an item's of its answer
// slots' by it.
func ItemOf(slots []SlotVerdict) Verdict {
	count := map[Verdict]int{}
	for _, s := range slots {
		count[s.Verdict]++
	}
	switch {
	case len(slots) == 0:
		return Unchecked
	case count[Correct] == len(slots):
		return Correct
	case count[NotYet] > 0:
		return NotYet
	case count[Unreadable] > 0:
		return Unreadable
	}
	return Unchecked
}
