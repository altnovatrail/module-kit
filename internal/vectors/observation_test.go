package vectors

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/altnovatrail/spec/contracts"
	observationv1 "github.com/altnovatrail/spec/contracts/novatrail/observation/v1"
	"github.com/altnovatrail/spec/types/formatv0"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	"github.com/altnovatrail/module-kit/internal/schema"
	"github.com/altnovatrail/module-kit/observation"
	"github.com/altnovatrail/module-kit/verdict"
)

// observationVector is a test vector of the observation, contracts/vectors/observation/v1/<case>.json
// (contracts/README.md, section 10): the task's canvas and items, as the module format writes them,
// and its cases.
type observationVector struct {
	Description string            `json:"description"`
	Canvas      formatv0.Canvas   `json:"canvas"`
	Items       []formatv0.Item   `json:"items"`
	Cases       []observationCase `json:"cases"`
}

// observationCase is a case of a vector of the observation: the readers' observations, whether
// each is valid, the options tapped, and either the joined observation, null when there is none,
// with each item's verdicts, or an essay's submission.
type observationCase struct {
	Description  string            `json:"description"`
	Observations []json.RawMessage `json:"observations"`
	Valid        []bool            `json:"valid"`
	Tapped       []struct {
		AnswerSlot struct {
			Item int32 `json:"item"`
			Slot int32 `json:"slot"`
		} `json:"answer_slot"`
		Value *string `json:"value"`
	} `json:"tapped"`
	Joined   json.RawMessage `json:"joined"`
	Verdicts []struct {
		AnswerSlots []struct {
			Verdict string `json:"verdict"`
			Reason  string `json:"reason"`
			Trap    string `json:"trap"`
		} `json:"answer_slots"`
		Verdict string `json:"verdict"`
	} `json:"verdicts"`
	Submission *string `json:"submission"`
}

// observations passes a vector of the observation: each observation of a case is valid as valid
// says; an essay's submission ends as submission says; any other case's two observations join into
// joined, or into none when either is not valid, and the joined observation, or none, with the
// options tapped, gives each item the verdicts that verdicts gives it (contracts/README.md, section
// 10). The vector is read as that section has it: JSON with no field that it does not name, its
// canvas and items by the schema of a task and strictly by the types of formatv0, and its
// observations by protojson, which refuses a field that the message does not have.
func observations(v vector) error {
	data, err := fs.ReadFile(v.fsys, v.path)
	if err != nil {
		return err
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var ov observationVector
	if err := decoder.Decode(&ov); err != nil {
		return fmt.Errorf("not a vector of the observation: %v", err)
	}
	if err := observationTask(data); err != nil {
		return err
	}
	for n, c := range ov.Cases {
		if err := observationCaseOf(ov.Items, c); err != nil {
			return fmt.Errorf("case %d, %q: %w", n+1, c.Description, err)
		}
	}
	return nil
}

// observationTask holds the canvas and the items of a vector to the schema of a task, as the module
// format writes them.
func observationTask(data []byte) error {
	var task struct {
		Canvas any   `json:"canvas"`
		Items  []any `json:"items"`
	}
	if err := json.Unmarshal(data, &task); err != nil {
		return err
	}
	var found []string
	for _, e := range schema.Validate("task.schema.json#/properties/canvas", task.Canvas) {
		found = append(found, "canvas: "+e.String())
	}
	for i, item := range task.Items {
		for _, e := range schema.Validate("task.schema.json#/$defs/item", item) {
			found = append(found, fmt.Sprintf("item %d: %s", i+1, e))
		}
	}
	if len(found) > 0 {
		return fmt.Errorf("the task is not one of the module format:\n%s", strings.Join(found, "\n"))
	}
	return nil
}

// observationCaseOf passes one case of a vector of the observation.
func observationCaseOf(items []formatv0.Item, c observationCase) error {
	if len(c.Valid) != len(c.Observations) {
		return fmt.Errorf("%d observations, and valid says of %d", len(c.Observations), len(c.Valid))
	}
	read := make([]*observationv1.Observation, len(c.Observations))
	for n, raw := range c.Observations {
		read[n] = &observationv1.Observation{}
		if err := protojson.Unmarshal(raw, read[n]); err != nil {
			return fmt.Errorf("observation %d: %v", n+1, err)
		}
		err := observation.Validate(items, read[n])
		if err != nil && !errors.Is(err, observation.ErrNotValid) {
			return fmt.Errorf("observation %d: %w", n+1, err)
		}
		if (err == nil) != c.Valid[n] {
			return fmt.Errorf("observation %d: valid says %t, and Validate gives %v", n+1, c.Valid[n], err)
		}
	}
	if c.Submission != nil {
		if len(read) != 1 {
			return fmt.Errorf("%d observations, and an essay has one", len(read))
		}
		end, err := observation.Essay(items, read[0])
		if errors.Is(err, observation.ErrNotValid) {
			end, err = observation.SubmissionUnchecked, nil
		}
		if err != nil {
			return err
		}
		if string(end) != *c.Submission {
			return fmt.Errorf("the submission ends %s, and the vector gives %s", end, *c.Submission)
		}
		return nil
	}
	if len(read) != 2 {
		return fmt.Errorf("%d observations, and a case that is no essay has two", len(read))
	}
	joined, err := observation.Join(items, read[0], read[1])
	if err != nil && !errors.Is(err, observation.ErrNotValid) {
		return err
	}
	if string(c.Joined) == "null" {
		if joined != nil {
			return fmt.Errorf("the observations join into %v, and the vector gives null", joined)
		}
	} else {
		want := &observationv1.Observation{}
		if err := protojson.Unmarshal(c.Joined, want); err != nil {
			return fmt.Errorf("joined: %v", err)
		}
		if !proto.Equal(joined, want) {
			return fmt.Errorf("the observations join into %v (%v), and the vector gives %v", joined, err, want)
		}
	}
	var tapped []observation.Tapped
	for _, t := range c.Tapped {
		if t.Value != nil {
			tapped = append(tapped, observation.Tapped{Place: &observationv1.AnswerSlotPlace{Item: t.AnswerSlot.Item, Slot: t.AnswerSlot.Slot}, Value: *t.Value})
		}
	}
	got, err := observation.Verdicts(context.Background(), items, joined, tapped)
	if err != nil {
		return err
	}
	if len(got) != len(c.Verdicts) {
		return fmt.Errorf("%d items get verdicts, and the vector gives %d", len(got), len(c.Verdicts))
	}
	for i, want := range c.Verdicts {
		if len(got[i].Slots) != len(want.AnswerSlots) {
			return fmt.Errorf("item %d has %d answer slots, and the vector gives %d", i+1, len(got[i].Slots), len(want.AnswerSlots))
		}
		for j, w := range want.AnswerSlots {
			expected := verdict.SlotVerdict{Verdict: verdict.Verdict(w.Verdict), Reason: verdict.Reason(w.Reason)}
			if w.Trap != "" {
				expected.Trap = &formatv0.TrapValue{Trap: formatv0.TrapID(w.Trap)}
			}
			if g, e := slotReading(got[i].Slots[j]), slotReading(expected); g != e {
				return fmt.Errorf("answer slot %d.%d: %s, and the vector gives %s", i+1, j+1, g, e)
			}
		}
		if string(got[i].Verdict) != want.Verdict {
			return fmt.Errorf("item %d is %s, and the vector gives %s", i+1, got[i].Verdict, want.Verdict)
		}
	}
	return nil
}

// The runner of the observation refuses a vector that any of its fields would make wrong: it reads
// each of them, and reads the file strictly.
func TestObservationsRefuse(t *testing.T) {
	for _, c := range []struct {
		name, file, old, new, says string
	}{
		{"a field the vector does not name", "zones.json", `"canvas":`, `"extra": 1, "canvas":`, "unknown field"},
		{"not JSON", "zones.json", `"cases": [`, `"cases": [[`, "not a vector"},
		{"an item not of the format", "zones.json", `"binding": "zone"`, `"binding": "nowhere"`, "the task is not one of the module format"},
		{"a field the observation does not have", "zones.json", `{"answer_slots": [{"answer_slot": {"item": 1, "slot": 1}`, `{"x": 1, "answer_slots": [{"answer_slot": {"item": 1, "slot": 1}`, "observation 1"},
		{"valid", "zones.json", `"valid": [true, true],`, `"valid": [false, true],`, "valid says false"},
		{"joined", "zones.json", `"joined": {"answer_slots": [{"answer_slot": {"item": 1, "slot": 1}, "value": "15"}`, `"joined": {"answer_slots": [{"answer_slot": {"item": 1, "slot": 1}, "value": "16"}`, "the observations join into"},
		{"tapped", "zones.json", `"tapped": [{"answer_slot": {"item": 1, "slot": 2}, "value": "after"}`, `"tapped": [{"answer_slot": {"item": 1, "slot": 2}, "value": "before"}`, "answer slot 1.2"},
		{"an answer slot's verdict", "zones.json", `{"answer_slots": [{"verdict": "correct"}, {"verdict": "correct"}`, `{"answer_slots": [{"verdict": "not_yet"}, {"verdict": "correct"}`, "answer slot 1.1"},
		{"an item's verdict", "zones.json", `{"verdict": "correct"}], "verdict": "correct"}`, `{"verdict": "correct"}], "verdict": "not_yet"}`, "item 1 is correct"},
		{"submission", "essay.json", `"submission": "assessment"`, `"submission": "unreadable"`, "the submission ends assessment"},
	} {
		data, err := fs.ReadFile(contracts.Vectors, "observation/v1/"+c.file)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Contains(data, []byte(c.old)) {
			t.Fatalf("%s: %s holds no %s", c.name, c.file, c.old)
		}
		tampered := bytes.Replace(data, []byte(c.old), []byte(c.new), 1)
		err = observations(vector{fsys: fstest.MapFS{"v.json": {Data: tampered}}, path: "v.json"})
		if err == nil || !strings.Contains(err.Error(), c.says) {
			t.Errorf("%s: %v, want an error saying %q", c.name, err, c.says)
		}
	}
	if err := observations(vector{fsys: fstest.MapFS{}, path: "none.json"}); err == nil {
		t.Error("a file that is not there passes")
	}
}
