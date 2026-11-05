package observation

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	observationv1 "github.com/altnovatrail/spec/contracts/novatrail/observation/v1"
	"github.com/altnovatrail/spec/types/formatv0"

	"github.com/altnovatrail/module-kit/verdict"
)

// ErrNotValid is what an observation that is not valid gives: it is no observation, as a failure of
// AI is none, and every answer slot that AI reads gets unchecked (contracts/README.md, section 10).
var ErrNotValid = errors.New("not valid")

// reasons is the table of contracts/README.md, section 10: each reason of the observation, with the
// reason of the verdict it gives. A reason that the package adds after this release is not in it,
// and not valid to this release; the task that takes the tag of spec that adds it names it here.
var reasons = map[observationv1.Unreadable]verdict.Reason{
	observationv1.Unreadable_UNREADABLE_BLANK:                 verdict.Blank,
	observationv1.Unreadable_UNREADABLE_ILLEGIBLE:             verdict.Illegible,
	observationv1.Unreadable_UNREADABLE_AMBIGUOUS:             verdict.Ambiguous,
	observationv1.Unreadable_UNREADABLE_NO_FINAL_ANSWER:       verdict.NoFinalAnswer,
	observationv1.Unreadable_UNREADABLE_SEVERAL_FINAL_ANSWERS: verdict.SeveralFinalAnswers,
}

// Submission is how the submission of an essay ends once it is read, as the test vectors of the
// observation write it (contracts/README.md, section 10): an essay gives no verdict.
type Submission string

// The ends of the submission of an essay.
const (
	SubmissionAssessment Submission = "assessment" // the essay read goes on to the assessment by its rubric
	SubmissionUnreadable Submission = "unreadable" // blank or illegible: the learner writes the sheet again
	SubmissionUnchecked  Submission = "unchecked"  // no observation: AI failed, or the observation is not valid
)

// Tapped is the option that the learner tapped in an answer slot bound choice: the answer slot by
// its place, and the text of the option as the open view of the task writes it, as api gets it in
// novatrail.api.v1.TappedChoice.
type Tapped struct {
	Place *observationv1.AnswerSlotPlace
	Value string
}

// place is an answer slot of a task by its place, its item's and its own, each from 1, with the
// answer slot as its check reads it.
type place struct {
	item, slot int32
	answerSlot verdict.AnswerSlot
}

func (p place) String() string { return fmt.Sprintf("%d.%d", p.item, p.slot) }

// task is what the rules of the observation read of a task: the answer slots that AI reads, in the
// task's order, those bound choice, which the learner taps, and whether it is an essay, a task whose
// one answer slot is bound sheet (format/README.md, section 14).
type task struct {
	read, tapped []place
	essay        bool
}

func taskOf(items []formatv0.Item) task {
	var t task
	count := 0
	for i, item := range items {
		for j, s := range verdict.AnswerSlots(item) {
			p := place{int32(i + 1), int32(j + 1), s}
			if s.Binding == string(formatv0.AnswerSlotBindingChoice) {
				t.tapped = append(t.tapped, p)
			} else {
				t.read = append(t.read, p)
			}
			count++
		}
	}
	t.essay = count == 1 && len(t.read) == 1 && t.read[0].answerSlot.Binding == string(formatv0.AnswerSlotBindingSheet)
	return t
}

// placeOf writes the place of an answer slot of an observation, ?.? where it has none.
func placeOf(p *observationv1.AnswerSlotPlace) string {
	if p == nil {
		return "?.?"
	}
	return fmt.Sprintf("%d.%d", p.GetItem(), p.GetSlot())
}

// Validate says whether an observation is valid for a task, given its items: nil when it is, and
// ErrNotValid, with every condition it breaks, when it is not (contracts/README.md, section 10). It
// holds every answer slot that AI reads, each once and in the task's order, and no other — neither a
// slot bound choice nor a place the task does not have; each with a value that is valid UTF-8 and
// not empty nor only white space, or with a reason of the table other than UNREADABLE_UNSPECIFIED;
// only a slot bound anywhere gives UNREADABLE_NO_FINAL_ANSWER, UNREADABLE_SEVERAL_FINAL_ANSWERS or
// a crop_box, the box only beside a value, whole thousandths from 0 to 1000 with ymin < ymax and
// xmin < xmax; and a slot bound sheet gives a value, UNREADABLE_BLANK or UNREADABLE_ILLEGIBLE. A
// nil observation is none, and not valid.
func Validate(items []formatv0.Item, o *observationv1.Observation) error {
	return valid(taskOf(items), o, "the observation")
}

// valid is Validate on a task, its error naming the observation as which.
func valid(t task, o *observationv1.Observation, which string) error {
	if o == nil {
		return fmt.Errorf("observation: %s is %w: there is none", which, ErrNotValid)
	}
	if found := breaches(t, o); len(found) > 0 {
		return fmt.Errorf("observation: %s is %w: %s", which, ErrNotValid, strings.Join(found, "; "))
	}
	return nil
}

// breaches gives the conditions of validity that an observation breaks, none when it is valid.
func breaches(t task, o *observationv1.Observation) []string {
	var found []string
	given := o.GetAnswerSlots()
	if !samePlaces(given, t.read) {
		var at, read []string
		for _, s := range given {
			at = append(at, placeOf(s.GetAnswerSlot()))
		}
		for _, p := range t.read {
			read = append(read, p.String())
		}
		found = append(found, fmt.Sprintf("it holds the answer slots %s, and AI reads %s, each once and in order, and no other", listOf(at), listOf(read)))
	}
	byPlace := map[string]string{}
	for _, p := range t.read {
		byPlace[p.String()] = p.answerSlot.Binding
	}
	for _, s := range given {
		at := placeOf(s.GetAnswerSlot())
		binding, known := byPlace[at]
		box := s.GetCropBox()
		switch read := s.GetRead().(type) {
		case *observationv1.AnswerSlotObservation_Value:
			switch {
			case !utf8.ValidString(read.Value):
				found = append(found, fmt.Sprintf("%s: its value %+q is not UTF-8", at, read.Value))
			case strings.TrimFunc(read.Value, isWhiteSpace) == "":
				found = append(found, fmt.Sprintf("%s: its value %q is empty or only white space", at, read.Value))
			}
			switch {
			case box == nil, !known:
			case binding != string(formatv0.AnswerSlotBindingAnywhere):
				found = append(found, fmt.Sprintf("%s: a crop_box only in a slot bound anywhere, and this one is bound %s", at, binding))
			case !validBox(box):
				found = append(found, fmt.Sprintf("%s: the crop_box [%d, %d, %d, %d] is not of thousandths from 0 to 1000 with ymin < ymax and xmin < xmax", at, box.GetYmin(), box.GetXmin(), box.GetYmax(), box.GetXmax()))
			}
		case *observationv1.AnswerSlotObservation_Unreadable:
			reason := read.Unreadable
			_, inTable := reasons[reason]
			switch {
			case reason == observationv1.Unreadable_UNREADABLE_UNSPECIFIED:
				found = append(found, fmt.Sprintf("%s: its reason is UNREADABLE_UNSPECIFIED", at))
			case !inTable:
				found = append(found, fmt.Sprintf("%s: its reason %d is none that this release knows", at, reason))
			case !known:
			case finalAnswer(reason) && binding != string(formatv0.AnswerSlotBindingAnywhere):
				found = append(found, fmt.Sprintf("%s: %s only in a slot bound anywhere, and this one is bound %s", at, reason, binding))
			case binding == string(formatv0.AnswerSlotBindingSheet) && !essayReason(reason):
				found = append(found, fmt.Sprintf("%s: a slot bound sheet gives a value, UNREADABLE_BLANK or UNREADABLE_ILLEGIBLE, not %s", at, reason))
			}
			if box != nil {
				found = append(found, fmt.Sprintf("%s: a crop_box stands only beside a value", at))
			}
		default:
			found = append(found, fmt.Sprintf("%s: it gives neither a value nor a reason", at))
		}
	}
	return found
}

// samePlaces says whether the answer slots of an observation are those AI reads, in order.
func samePlaces(given []*observationv1.AnswerSlotObservation, read []place) bool {
	if len(given) != len(read) {
		return false
	}
	for i, s := range given {
		p := s.GetAnswerSlot()
		if p == nil || p.GetItem() != read[i].item || p.GetSlot() != read[i].slot {
			return false
		}
	}
	return true
}

func listOf(places []string) string {
	if len(places) == 0 {
		return "none"
	}
	return strings.Join(places, ", ")
}

// isWhiteSpace says whether a character has the Unicode property White_Space, as step 4 of the
// module format's numbers has it (format/README.md, section 15).
func isWhiteSpace(r rune) bool { return unicode.Is(unicode.White_Space, r) }

// validBox says whether a crop_box is valid: from 0 to 1000, ymin < ymax and xmin < xmax (R53 spec).
func validBox(b *observationv1.CropBox) bool {
	for _, n := range []int32{b.GetYmin(), b.GetXmin(), b.GetYmax(), b.GetXmax()} {
		if n < 0 || n > 1000 {
			return false
		}
	}
	return b.GetYmin() < b.GetYmax() && b.GetXmin() < b.GetXmax()
}

// finalAnswer says whether a reason is one that only a slot bound anywhere gives.
func finalAnswer(r observationv1.Unreadable) bool {
	return r == observationv1.Unreadable_UNREADABLE_NO_FINAL_ANSWER || r == observationv1.Unreadable_UNREADABLE_SEVERAL_FINAL_ANSWERS
}

// essayReason says whether a reason is one that a slot bound sheet gives.
func essayReason(r observationv1.Unreadable) bool {
	return r == observationv1.Unreadable_UNREADABLE_BLANK || r == observationv1.Unreadable_UNREADABLE_ILLEGIBLE
}

// Join joins the observations of two readers of a task into one, answer slot by answer slot, first
// being the reader whom tutor names first in the pair (contracts/README.md, section 10). It runs no
// predicate and takes no context: the join compares values as the check normalizes them.
//   - A value that the kind normalizes to nothing, such as = under number, is UNREADABLE_BLANK
//     before any pairing.
//   - Two values join when both are in the kind's notation, the check compares them equal, as it
//     compares a value with a trap value, and gives them the same verdict, its reason and trap too
//     — a predicate the same integer, which gives one verdict —, or when both are out of the
//     notation and their normal forms are the same. A value in the notation and one out of it never
//     join. The joined value is the first, with the first's crop_box, or else the second's.
//   - Two reasons that are the same join into that reason, and every other pair into
//     UNREADABLE_AMBIGUOUS, with no crop_box.
//
// An error gives no joined observation: either observation is not valid, ErrNotValid, which says
// which; the task is an essay, which has one reader; or a check is one that the verdict refuses,
// such as one that holds the placeholder of a template, verdict.ErrPlaceholder, whatever was read.
// The joined observation is a new message, which shares nothing with the two.
func Join(items []formatv0.Item, first, second *observationv1.Observation) (*observationv1.Observation, error) {
	t := taskOf(items)
	if t.essay {
		return nil, errors.New("observation: an essay has one reader, and nothing to join")
	}
	if err := checked(t); err != nil {
		return nil, err
	}
	if err := valid(t, first, "the first observation"); err != nil {
		return nil, err
	}
	if err := valid(t, second, "the second observation"); err != nil {
		return nil, err
	}
	joined := &observationv1.Observation{}
	for i, p := range t.read {
		joined.AnswerSlots = append(joined.AnswerSlots, joinSlot(p, first.GetAnswerSlots()[i], second.GetAnswerSlots()[i]))
	}
	return joined, nil
}

// checked says whether the verdict takes the check of every answer slot that AI reads: OfSlot on no
// value reads the check alone — its kind, key, options, placeholder and predicate — and runs no
// Starlark. So a check that the verdict refuses, a rubric too, refuses the join and the verdict
// whatever was read, and whether there is an observation at all.
func checked(t task) error {
	for _, p := range t.read {
		if _, err := verdict.OfSlot(context.Background(), p.answerSlot, nil); err != nil {
			return fmt.Errorf("observation: answer slot %s: %w", p, err)
		}
	}
	return nil
}

// side is a reader's answer slot as the join takes it: a value with its crop_box, or a reason.
type side struct {
	value  *string
	box    *observationv1.CropBox
	reason observationv1.Unreadable
}

func sideOf(p place, s *observationv1.AnswerSlotObservation) side {
	read, ok := s.GetRead().(*observationv1.AnswerSlotObservation_Value)
	if !ok {
		return side{reason: s.GetUnreadable()}
	}
	if form, _ := verdict.NormalForm(p.answerSlot, read.Value); form == "" {
		return side{reason: observationv1.Unreadable_UNREADABLE_BLANK}
	}
	return side{value: &read.Value, box: s.GetCropBox()}
}

// joinSlot joins the two readers' readings of one answer slot.
func joinSlot(p place, a, b *observationv1.AnswerSlotObservation) *observationv1.AnswerSlotObservation {
	joined := &observationv1.AnswerSlotObservation{AnswerSlot: &observationv1.AnswerSlotPlace{Item: p.item, Slot: p.slot}}
	x, y := sideOf(p, a), sideOf(p, b)
	if x.value != nil && y.value != nil {
		if alike(p.answerSlot, *x.value, *y.value) {
			joined.Read = &observationv1.AnswerSlotObservation_Value{Value: *x.value}
			box := x.box
			if box == nil {
				box = y.box
			}
			if box != nil {
				joined.CropBox = &observationv1.CropBox{Ymin: box.GetYmin(), Xmin: box.GetXmin(), Ymax: box.GetYmax(), Xmax: box.GetXmax()}
			}
			return joined
		}
	}
	reason := observationv1.Unreadable_UNREADABLE_AMBIGUOUS
	if x.value == nil && y.value == nil && x.reason == y.reason {
		reason = x.reason
	}
	joined.Read = &observationv1.AnswerSlotObservation_Unreadable{Unreadable: reason}
	return joined
}

// alike says whether two values that are not blank join under the check of an answer slot. The
// verdicts compared are those of number, fraction and choice, which run no predicate, on a check
// that checked took and values that Validate took: OfSlot gives no error on them, and one is a
// defect of module-kit.
func alike(s verdict.AnswerSlot, a, b string) bool {
	formA, inA := verdict.NormalForm(s, a)
	formB, inB := verdict.NormalForm(s, b)
	switch {
	case inA != inB:
		return false
	case !inA:
		return formA == formB
	case !verdict.Same(s, a, b):
		return false
	case s.Check.Kind == formatv0.CheckKindPredicate:
		return true
	}
	var got [2]verdict.SlotVerdict
	for i, v := range []string{a, b} {
		read, err := verdict.OfSlot(context.Background(), s, &v)
		if err != nil {
			panic(fmt.Sprintf("observation: the verdict refuses %q, which the check and the observation took: %v", v, err))
		}
		got[i] = read
	}
	return got[0].Verdict == got[1].Verdict && got[0].Reason == got[1].Reason && trapOf(got[0]) == trapOf(got[1])
}

// trapOf gives the ID of the trap of a verdict, or "" when it has none.
func trapOf(v verdict.SlotVerdict) formatv0.TrapID {
	if v.Trap == nil {
		return ""
	}
	return v.Trap.Trap
}

// Verdicts gives the verdict of every item of a task on an observation, in the task's order, each
// with its answer slots' (contracts/README.md, section 10; format/README.md, section 15): a value
// gets the verdict that the check kind gives that value read, a predicate's in the sandbox of
// Starlark with ctx; a reason gets unreadable, with the reason of the verdict that the table names;
// and a slot bound choice gets the verdict of the option tapped, or blank when none was. The
// observation is one that Join gave, or one that Validate accepts — a reading of one reader, the
// truth of a sample's markup —, or nil when there is none: AI failed, a limit ran out, or an
// observation is not valid. Then every answer slot that AI reads is unchecked, and the slots bound
// choice still get their verdicts.
//
// tapped lists the options tapped, one for each answer slot bound choice that the learner tapped
// one of, in the task's order, as api gets them.
//
// An error gives no verdict: ctx is done before the verdict, or stops a predicate as it runs; the
// task has a rubric, verdict.ErrNoVerdict; an item has no answer slots; the observation is not
// valid, ErrNotValid; an option tapped is not of an answer slot bound choice of the task, or not in
// order, or twice; or a check is one that the verdict refuses (verdict.OfSlot), whatever was read.
func Verdicts(ctx context.Context, items []formatv0.Item, o *observationv1.Observation, tapped []Tapped) ([]verdict.ItemVerdict, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("observation: %w", err)
	}
	t := taskOf(items)
	if err := checked(t); err != nil {
		return nil, err
	}
	read := map[string]*observationv1.AnswerSlotObservation{}
	if o != nil {
		if err := valid(t, o, "the observation"); err != nil {
			return nil, err
		}
		for _, s := range o.GetAnswerSlots() {
			read[placeOf(s.GetAnswerSlot())] = s
		}
	}
	taps, err := tapsOf(t, tapped)
	if err != nil {
		return nil, err
	}
	verdicts := make([]verdict.ItemVerdict, len(items))
	for i, item := range items {
		slots := verdict.AnswerSlots(item)
		if len(slots) == 0 {
			return nil, fmt.Errorf("observation: item %d has no answer slots", i+1)
		}
		of := make([]verdict.SlotVerdict, len(slots))
		for j, s := range slots {
			at := place{int32(i + 1), int32(j + 1), s}
			v, err := slotVerdict(ctx, at, o != nil, read[at.String()], taps)
			if err != nil {
				return nil, fmt.Errorf("observation: answer slot %s: %w", at, err)
			}
			of[j] = v
		}
		verdicts[i] = verdict.ItemVerdict{Verdict: verdict.ItemOf(of), Slots: of}
	}
	return verdicts, nil
}

// slotVerdict gives the verdict of one answer slot: of the option tapped in a slot bound choice; of
// the slot's reading when there is an observation; and unchecked when there is none.
func slotVerdict(ctx context.Context, p place, observed bool, s *observationv1.AnswerSlotObservation, taps map[string]*string) (verdict.SlotVerdict, error) {
	if p.answerSlot.Binding == string(formatv0.AnswerSlotBindingChoice) {
		return verdict.OfSlot(ctx, p.answerSlot, taps[p.String()])
	}
	if !observed {
		return verdict.SlotVerdict{Verdict: verdict.Unchecked}, nil
	}
	if read, ok := s.GetRead().(*observationv1.AnswerSlotObservation_Value); ok {
		return verdict.OfSlot(ctx, p.answerSlot, &read.Value)
	}
	return verdict.SlotVerdict{Verdict: verdict.Unreadable, Reason: reasons[s.GetUnreadable()]}, nil
}

// tapsOf gives the options tapped by the places of their answer slots, which are bound choice, each
// once and in the task's order.
func tapsOf(t task, tapped []Tapped) (map[string]*string, error) {
	taps := map[string]*string{}
	next := 0
	for n, tap := range tapped {
		at := placeOf(tap.Place)
		found := -1
		for i := next; i < len(t.tapped); i++ {
			if t.tapped[i].String() == at {
				found = i
				break
			}
		}
		if found < 0 {
			return nil, fmt.Errorf("observation: option tapped %d: the answer slot %s is no answer slot bound choice of the task after those tapped before it", n+1, at)
		}
		value := tap.Value
		taps[at] = &value
		next = found + 1
	}
	return taps, nil
}

// Essay says how the submission of an essay ends on its observation: SubmissionAssessment when it
// gives the essay's text, which goes on to the assessment by the rubric, without the picture;
// SubmissionUnreadable when the sheet is blank or illegible, and the learner writes it again; and
// SubmissionUnchecked when there is no observation, nil — AI failed, or a limit ran out
// (contracts/README.md, section 10; R75 spec, R87 spec). A rubric normalizes nothing, so the text
// is the value as the reader gave it. An error gives no end: the task is no essay, whose one answer
// slot is bound sheet; or the observation is not valid, ErrNotValid, and the submission then ends
// unchecked.
func Essay(items []formatv0.Item, o *observationv1.Observation) (Submission, error) {
	t := taskOf(items)
	if !t.essay {
		return "", errors.New("observation: the task is no essay: an essay's one answer slot is bound sheet")
	}
	if o == nil {
		return SubmissionUnchecked, nil
	}
	if err := valid(t, o, "the observation"); err != nil {
		return "", err
	}
	if _, ok := o.GetAnswerSlots()[0].GetRead().(*observationv1.AnswerSlotObservation_Value); ok {
		return SubmissionAssessment, nil
	}
	return SubmissionUnreadable, nil
}
