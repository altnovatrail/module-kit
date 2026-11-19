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

// paperCase is a case of a vector of the key of a task from paper, contracts/vectors/paper/v1/:
// a formula with the solver's answer or the reason it refuses it, an answer of AI with whether it
// is one, or a key with a value read and the verdict on it.
type paperCase struct {
	Description string          `json:"description"`
	Formula     *string         `json:"formula"`
	Answer      json.RawMessage `json:"answer"`
	Refused     *string         `json:"refused"`
	Key         json.RawMessage `json:"key"`
	Value       *string         `json:"value"`
	Verdict     *struct {
		Verdict string `json:"verdict"`
		Reason  string `json:"reason"`
	} `json:"verdict"`
}

// paperCases reads a vector of the key strictly: JSON with no field that it does not name, and its
// messages by protojson, which refuses a field that the message does not have.
func paperCases(v vector) ([]paperCase, error) {
	data, err := fs.ReadFile(v.fsys, v.path)
	if err != nil {
		return nil, err
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var pv struct {
		Description string      `json:"description"`
		Cases       []paperCase `json:"cases"`
	}
	if err := decoder.Decode(&pv); err != nil {
		return nil, fmt.Errorf("not a vector of the key: %v", err)
	}
	if len(pv.Cases) == 0 {
		return nil, errors.New("a vector of the key with no case")
	}
	return pv.Cases, nil
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
	cases, err := paperCases(v)
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
	cases, err := paperCases(v)
	if err != nil {
		return err
	}
	var failed []error
	for _, c := range cases {
		var key bool
		if err := json.Unmarshal(c.Key, &key); err != nil {
			return fmt.Errorf("%s: key is not true or false: %v", c.Description, err)
		}
		answer := &paperv1.Answer{}
		if err := message(c.Answer, answer); err != nil {
			return fmt.Errorf("%s: %v", c.Description, err)
		}
		if rule := paper.Breach(answer); (rule == "") != key {
			failed = append(failed, fmt.Errorf("%s: the rule broken is %q, and the vector says key %v", c.Description, rule, key))
		}
	}
	return errors.Join(failed...)
}

// verdictVectors passes verdicts.json as api gives the verdict: the value read, or a blank slot
// for null, in the one answer slot, brought to the key, then validated and judged by the package
// observation against the task's one item.
func verdictVectors(v vector) error {
	cases, err := paperCases(v)
	if err != nil {
		return err
	}
	var failed []error
	for _, c := range cases {
		if c.Verdict == nil {
			return fmt.Errorf("%s: a case of the verdict gives no verdict", c.Description)
		}
		answer := &paperv1.Answer{}
		if err := message(c.Key, answer); err != nil {
			return fmt.Errorf("%s: %v", c.Description, err)
		}
		slot := &observationv1.AnswerSlotObservation{AnswerSlot: &observationv1.AnswerSlotPlace{Item: 1, Slot: 1}}
		read := "null"
		if c.Value != nil {
			slot.Read = &observationv1.AnswerSlotObservation_Value{Value: *c.Value}
			read = fmt.Sprintf("%q", *c.Value)
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
			failed = append(failed, fmt.Errorf("%s: %s gives %s %s, and the vector %s %s", c.Description, read, s.Verdict, s.Reason, c.Verdict.Verdict, c.Verdict.Reason))
		}
	}
	return errors.Join(failed...)
}
