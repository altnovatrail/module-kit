package module

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strconv"

	"github.com/altnovatrail/spec/types/formatv0"

	"github.com/altnovatrail/module-kit/verdict"
)

// The rules of a task that the check of an answer slot sees itself (format/README.md, section 15).
const (
	RuleTaskChoiceKey           = "task.choice-key"
	RuleTaskChoiceOptions       = "task.choice-options"
	RuleTaskFractionKey         = "task.fraction-key"
	RuleTaskTrapValues          = "task.trap-values"
	RuleTaskPredicateExpression = "task.predicate-expression"
)

// slotsOf gives the answer slots of an item as their checks read them, in the order of
// verdict.AnswerSlots, with the place of each in the task: an answer slot of a handwriting sheet,
// and a box of a workbook's line (format/README.md, section 14).
func slotsOf(item formatv0.Item, i int) (slots []verdict.AnswerSlot, places [][]string) {
	slots = verdict.AnswerSlots(item)
	for j := range item.AnswerSlots {
		places = append(places, []string{"items", strconv.Itoa(i), "answer_slots", strconv.Itoa(j)})
	}
	for j, part := range item.Line {
		if part.Binding != nil {
			places = append(places, []string{"items", strconv.Itoa(i), "line", strconv.Itoa(j)})
		}
	}
	if len(places) != len(slots) {
		panic(fmt.Sprintf("module: %d places for the %d answer slots of item %d", len(places), len(slots), i))
	}
	return slots, places
}

// A breach of a rule that the check of an answer slot sees itself: its rule, its place within the
// answer slot, and what breaks it.
type slotBreach struct {
	rule    string
	tokens  []string
	message string
}

// checkRules checks the rules of a task that the check of each answer slot sees itself:
// task.choice-key, task.choice-options, task.fraction-key, task.trap-values and
// task.predicate-expression. It reads a task that is no template: a template is checked through its
// instances, each a task of its own (templates.go). The key or a trap value of a number or a fraction
// that holds a placeholder outside a template breaks task.placeholders, since no instance puts a
// number in its place, and nothing more of its check is read (format/README.md, section 23).
func (c *checker) checkRules(path string, task formatv0.Task) {
	for i, item := range task.Items {
		slots, places := slotsOf(item, i)
		for n, slot := range slots {
			if holdsPlaceholder(slot) {
				c.add(RuleTaskPlaceholders, path, 0, 0, at(slices.Concat(places[n], []string{"check"})...),
					"the key or a trap value holds the placeholder of a template, and the task is no template: no instance puts a number in its place")
				continue
			}
			for _, b := range slotRules(slot) {
				c.add(b.rule, path, 0, 0, at(slices.Concat(places[n], b.tokens)...), b.message)
			}
		}
	}
}

// holdsPlaceholder says whether the key or a trap value of a number or a fraction holds the
// placeholder of a template, as the verdict reads it: the check gives verdict.ErrPlaceholder.
func holdsPlaceholder(slot verdict.AnswerSlot) bool {
	_, err := verdict.OfSlot(context.Background(), slot, nil)
	return errors.Is(err, verdict.ErrPlaceholder)
}

// slotRules gives the breaches of the rules that the check of an answer slot sees itself, by the
// verdicts that the check gives, and a predicate by verdict.CheckPredicate (format/README.md,
// section 15). It reads one answer slot of a task, or of an instance of a template, which is checked
// as a task.
func slotRules(slot verdict.AnswerSlot) []slotBreach {
	var found []slotBreach
	check := slot.Check
	if check.Kind == formatv0.CheckKindChoice {
		found = append(found, choiceRules(slot)...)
	}
	if check.Kind == formatv0.CheckKindPredicate {
		if err := verdict.CheckPredicate(*check.Predicate); err != nil {
			found = append(found, slotBreach{RuleTaskPredicateExpression, []string{"check", "predicate"},
				fmt.Sprintf("the predicate is no expression of the dialect on value and the built-in names alone: %v", err)})
		}
	}
	if check.Kind == formatv0.CheckKindFraction && check.Form != nil && *check.Form == formatv0.CheckFormReduced &&
		verdictOf(slot, *check.Key).Verdict != verdict.Correct {
		found = append(found, slotBreach{RuleTaskFractionKey, []string{"check", "key"},
			fmt.Sprintf("the key %s can be reduced, and the check takes only a fraction in lowest terms", *check.Key)})
	}
	for k, trapValue := range check.TrapValues {
		tokens := []string{"check", "trap_values", strconv.Itoa(k), "value"}
		if v := verdictOf(slot, trapValue.Value); v.Verdict != verdict.NotYet {
			found = append(found, slotBreach{RuleTaskTrapValues, tokens,
				fmt.Sprintf("the trap value %s gets %s from the check, not not_yet: a trap value is a wrong value", trapValue.Value, said(v))})
		} else if check.Kind == formatv0.CheckKindChoice && !slices.Contains(slot.ChoiceOptions(), formatv0.Text(trapValue.Value)) {
			found = append(found, slotBreach{RuleTaskTrapValues, tokens,
				fmt.Sprintf("the trap value %s is none of the options as written", trapValue.Value)})
		}
		for _, earlier := range check.TrapValues[:k] {
			if verdict.Same(slot, earlier.Value, trapValue.Value) {
				found = append(found, slotBreach{RuleTaskTrapValues, tokens,
					fmt.Sprintf("the trap value %s is equal to the trap value %s before it, as the check compares them", trapValue.Value, earlier.Value)})
				break
			}
		}
	}
	return found
}

// choiceRules gives the breaches of task.choice-key and task.choice-options: the key is one of the
// options as written, and no two options are alike once normalized.
func choiceRules(slot verdict.AnswerSlot) []slotBreach {
	var found []slotBreach
	options := slot.ChoiceOptions()
	if key := *slot.Check.Key; !slices.Contains(options, formatv0.Text(key)) {
		found = append(found, slotBreach{RuleTaskChoiceKey, []string{"check", "key"},
			fmt.Sprintf("the key %s is none of the options as written", key)})
	}
	list := []string{"check", "options"}
	if slot.Binding == string(formatv0.AnswerSlotBindingChoice) {
		list = []string{"options"}
	}
	for k, option := range options {
		for _, earlier := range options[:k] {
			if verdict.Same(slot, string(earlier), string(option)) {
				found = append(found, slotBreach{RuleTaskChoiceOptions, slices.Concat(list, []string{strconv.Itoa(k)}),
					fmt.Sprintf("the option %s is the option %s before it, once normalized as a choice normalizes them", option, earlier)})
				break
			}
		}
	}
	return found
}

// verdictOf gives the verdict of the check of an answer slot on a value written in a module. The
// schema and the YAML rules took the check and the value, and a placeholder breaks another rule, so
// a check that gives no verdict is a fault of module-kit.
func verdictOf(slot verdict.AnswerSlot, value string) verdict.SlotVerdict {
	v, err := verdict.OfSlot(context.Background(), slot, &value)
	if err != nil {
		panic(fmt.Sprintf("module: a check that its schema accepts: %v", err))
	}
	return v
}

// said writes a verdict as a breach names it: with its reason when it is unreadable.
func said(v verdict.SlotVerdict) string {
	if v.Reason != "" {
		return string(v.Verdict) + " (" + string(v.Reason) + ")"
	}
	return string(v.Verdict)
}
