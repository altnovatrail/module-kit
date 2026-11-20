package vectors

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"

	observationv1 "github.com/altnovatrail/spec/contracts/novatrail/observation/v1"
	paperv1 "github.com/altnovatrail/spec/contracts/novatrail/paper/v1"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	"github.com/altnovatrail/module-kit/observation"
	"github.com/altnovatrail/module-kit/paper"
)

// paperRunners passes the files of the vectors of the key of a task from paper,
// contracts/vectors/paper/v1/, each a kind of its own, by the name of the file without .json.
var paperRunners = map[string]func(vector) error{
	"answers":  answerVectors,
	"solver":   solverVectors,
	"verdicts": verdictVectors,
}

// solverCase is a case of solver.json: a formula with the solver's answer, or the reason it
// refuses it.
type solverCase struct {
	Description string          `json:"description"`
	Formula     *string         `json:"formula"`
	Answer      json.RawMessage `json:"answer"`
	Refused     *string         `json:"refused"`
}

// answerCase is a case of answers.json: an answer of AI and whether it is one.
type answerCase struct {
	Description string          `json:"description"`
	Answer      json.RawMessage `json:"answer"`
	Key         *bool           `json:"key"`
}

// verdictCase is a case of verdicts.json: a key, the value read, and the verdict on it. Value is
// raw, so that a value missing, nil, is told from null, which is nothing read.
type verdictCase struct {
	Description string          `json:"description"`
	Key         json.RawMessage `json:"key"`
	Value       json.RawMessage `json:"value"`
	Verdict     *struct {
		Verdict string `json:"verdict"`
		Reason  string `json:"reason"`
	} `json:"verdict"`
}

// paperCases reads a vector of the key strictly, each case into C: JSON with no field that the
// vector or its kind of case does not name, and a description to each case. Its messages are read
// by protojson, which refuses a field that the message does not have.
func paperCases[C any](v vector, description func(C) string) ([]C, error) {
	data, err := fs.ReadFile(v.fsys, v.path)
	if err != nil {
		return nil, err
	}
	var pv struct {
		Description string            `json:"description"`
		Cases       []json.RawMessage `json:"cases"`
	}
	if err := strict(data, &pv); err != nil {
		return nil, fmt.Errorf("not a vector of the key: %v", err)
	}
	if len(pv.Cases) == 0 {
		return nil, errors.New("a vector of the key with no case")
	}
	cases := make([]C, len(pv.Cases))
	for i, raw := range pv.Cases {
		if err := strict(raw, &cases[i]); err != nil {
			return nil, fmt.Errorf("case %d: %v", i+1, err)
		}
		if description(cases[i]) == "" {
			return nil, fmt.Errorf("case %d has no description", i+1)
		}
	}
	return cases, nil
}

func strict(data []byte, v any) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	return decoder.Decode(v)
}

func message(raw json.RawMessage, m proto.Message) error {
	if raw == nil {
		return errors.New("a message the case does not give")
	}
	return protojson.Unmarshal(raw, m)
}

// solverVectors passes solver.json: the solver gives each formula its answer, or refuses it by the
// reason the case names.
func solverVectors(v vector) error {
	cases, err := paperCases(v, func(c solverCase) string { return c.Description })
	if err != nil {
		return err
	}
	var failed []error
	for _, c := range cases {
		if c.Formula == nil || (c.Answer == nil) == (c.Refused == nil) {
			return fmt.Errorf("%s: a case of the solver gives a formula and either an answer or a reason", c.Description)
		}
		got, reason := paper.Solve(*c.Formula)
		if c.Refused != nil {
			if got != nil || reason != paper.Reason(*c.Refused) {
				failed = append(failed, fmt.Errorf("%s: %q gives %v %q, and the vector refuses it by %s", c.Description, *c.Formula, got, reason, *c.Refused))
			}
			continue
		}
		want := &paperv1.Number{}
		if err := message(c.Answer, want); err != nil {
			return fmt.Errorf("%s: %v", c.Description, err)
		}
		if reason != "" || !proto.Equal(got, want) {
			failed = append(failed, fmt.Errorf("%s: %q gives %v %q, and the vector %v", c.Description, *c.Formula, got, reason, want))
		}
	}
	return errors.Join(failed...)
}

// answerVectors passes answers.json: an answer of AI breaks a rule of one exactly when the case
// says it is no key.
func answerVectors(v vector) error {
	cases, err := paperCases(v, func(c answerCase) string { return c.Description })
	if err != nil {
		return err
	}
	var failed []error
	for _, c := range cases {
		if c.Key == nil {
			return fmt.Errorf("%s: a case of an answer gives key, true or false", c.Description)
		}
		answer := &paperv1.Answer{}
		if err := message(c.Answer, answer); err != nil {
			return fmt.Errorf("%s: %v", c.Description, err)
		}
		if rule := paper.Breach(answer); (rule == "") != *c.Key {
			failed = append(failed, fmt.Errorf("%s: the rule broken is %q, and the vector says key %v", c.Description, rule, *c.Key))
		}
	}
	return errors.Join(failed...)
}

// verdictVectors passes verdicts.json as api gives the verdict: the value read, or a blank slot
// for null, in the one answer slot, brought to the key, then validated and judged by the package
// observation against the task's one item. The key of a case is one an answer of AI may be: it
// breaks no rule.
func verdictVectors(v vector) error {
	cases, err := paperCases(v, func(c verdictCase) string { return c.Description })
	if err != nil {
		return err
	}
	var failed []error
	for _, c := range cases {
		if c.Verdict == nil || c.Value == nil {
			return fmt.Errorf("%s: a case of the verdict gives a value, null too, and a verdict", c.Description)
		}
		var value *string
		if err := json.Unmarshal(c.Value, &value); err != nil {
			return fmt.Errorf("%s: value is not a string or null: %v", c.Description, err)
		}
		answer := &paperv1.Answer{}
		if err := message(c.Key, answer); err != nil {
			return fmt.Errorf("%s: %v", c.Description, err)
		}
		if rule := paper.Breach(answer); rule != "" {
			return fmt.Errorf("%s: the key breaks the rule %s of an answer, and no key does", c.Description, rule)
		}
		slot := &observationv1.AnswerSlotObservation{AnswerSlot: &observationv1.AnswerSlotPlace{Item: 1, Slot: 1}}
		if value != nil {
			slot.Read = &observationv1.AnswerSlotObservation_Value{Value: *value}
		} else {
			slot.Read = &observationv1.AnswerSlotObservation_Unreadable{Unreadable: observationv1.Unreadable_UNREADABLE_BLANK}
		}
		o := paper.Brought(answer, &observationv1.Observation{AnswerSlots: []*observationv1.AnswerSlotObservation{slot}})
		items := paper.Items(answer)
		if err := observation.Validate(items, o); err != nil {
			return fmt.Errorf("%s: %v", c.Description, err)
		}
		got, err := observation.Verdicts(context.Background(), items, o, nil)
		if err != nil {
			return fmt.Errorf("%s: %v", c.Description, err)
		}
		s := got[0].Slots[0]
		if string(s.Verdict) != c.Verdict.Verdict || string(s.Reason) != c.Verdict.Reason {
			failed = append(failed, fmt.Errorf("%s: %s gives %s %s, and the vector %s %s", c.Description, quoted(value), s.Verdict, s.Reason, c.Verdict.Verdict, c.Verdict.Reason))
		}
	}
	return errors.Join(failed...)
}
