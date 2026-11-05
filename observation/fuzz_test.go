package observation

import (
	"context"
	"encoding/json"
	"errors"
	"io/fs"
	"testing"

	"github.com/altnovatrail/spec/contracts"
	observationv1 "github.com/altnovatrail/spec/contracts/novatrail/observation/v1"
	"github.com/altnovatrail/spec/types/formatv0"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	"github.com/altnovatrail/module-kit/verdict"
)

// vectorTasks gives the tasks of the vectors of the observation in contracts.Vectors, by their
// items, and every observation of their cases, so that the fuzz tests start from what the vectors
// hold. The runner of the vectors reads them strictly; this reads only what the fuzz tests need.
func vectorTasks(tb testing.TB) ([][]formatv0.Item, []*observationv1.Observation) {
	names, err := fs.Glob(contracts.Vectors, "observation/v1/*.json")
	if err != nil || len(names) == 0 {
		tb.Fatalf("no vectors of the observation: %v", err)
	}
	var tasks [][]formatv0.Item
	var observations []*observationv1.Observation
	for _, name := range names {
		data, err := fs.ReadFile(contracts.Vectors, name)
		if err != nil {
			tb.Fatal(err)
		}
		var v struct {
			Items []formatv0.Item `json:"items"`
			Cases []struct {
				Observations []json.RawMessage `json:"observations"`
			} `json:"cases"`
		}
		if err := json.Unmarshal(data, &v); err != nil {
			tb.Fatalf("%s: %v", name, err)
		}
		tasks = append(tasks, v.Items)
		for _, c := range v.Cases {
			for _, raw := range c.Observations {
				o := &observationv1.Observation{}
				if err := protojson.Unmarshal(raw, o); err != nil {
					tb.Fatalf("%s: %v", name, err)
				}
				observations = append(observations, o)
			}
		}
	}
	return tasks, observations
}

// The tasks of FuzzJoin, each with one answer slot that AI reads, of each kind that gives a verdict
// and of each binding written by hand.
var oneSlotTasks = [][]formatv0.Item{
	anywhere,
	{{AnswerSlots: []formatv0.AnswerSlot{sheetSlot(formatv0.AnswerSlotBindingZone, numberCheck("15", "minutes"))}}},
	{{Line: []formatv0.LinePart{boxOf(fractionCheck("3/4"))}}},
	{{Line: []formatv0.LinePart{boxOf(choiceCheck(">", "<", ">", "="))}}},
	{{AnswerSlots: []formatv0.AnswerSlot{sheetSlot(formatv0.AnswerSlotBindingZone, choiceCheck("even", "even", "odd"))}}},
	{{Line: []formatv0.LinePart{boxOf(predicateCheck("value % 7 == 0"))}}},
}

// readingOf gives one reading of the answer slot 1.1: a value, or the reason r, with a crop_box
// when withBox.
func readingOf(isValue bool, v string, r int32, withBox bool, ymin, xmin, ymax, xmax int32) *observationv1.Observation {
	s := reason(1, 1, observationv1.Unreadable(r))
	if isValue {
		s = value(1, 1, v)
	}
	if withBox {
		boxed(s, ymin, xmin, ymax, xmax)
	}
	return observation(s)
}

// FuzzJoin joins any two readings of one answer slot: Join gives a joined observation when both are
// valid, and only then; the joined observation is valid and gets a verdict; the two in the other
// order join into a value, or into the same reason, where these do; and a reading joined with
// itself is itself, but for a value that the kind normalizes to nothing, which is blank.
func FuzzJoin(f *testing.F) {
	for i, v := range []string{"7", "7 rungs", "7²", "0.5", "1/2", "6/8", "=", "\uFEFF", "\u0085", "-0", "012", "＞", "EVEN", "odd.", "15 minutes"} {
		f.Add(uint8(i), true, v, int32(0), true, v+".", int32(0), i%2 == 0, int32(1), int32(2), int32(3), int32(4))
	}
	f.Add(uint8(0), false, "", int32(4), false, "", int32(5), false, int32(0), int32(0), int32(0), int32(0))
	f.Add(uint8(1), false, "", int32(1000), true, "x", int32(0), true, int32(0), int32(0), int32(1000), int32(1000))
	f.Fuzz(func(t *testing.T, which uint8, aIsValue bool, a string, ra int32, bIsValue bool, b string, rb int32, withBox bool, ymin, xmin, ymax, xmax int32) {
		items := oneSlotTasks[int(which)%len(oneSlotTasks)]
		first := readingOf(aIsValue, a, ra, withBox, ymin, xmin, ymax, xmax)
		second := readingOf(bIsValue, b, rb, false, 0, 0, 0, 0)
		bothValid := Validate(items, first) == nil && Validate(items, second) == nil
		joined, err := Join(items, first, second)
		if !bothValid {
			if joined != nil || !errors.Is(err, ErrNotValid) {
				t.Fatalf("an observation not valid joins into %s, %v", written(joined), err)
			}
			return
		}
		if err != nil {
			t.Fatal(err)
		}
		if err := Validate(items, joined); err != nil {
			t.Fatalf("the joined %s: %v", written(joined), err)
		}
		if _, err := Verdicts(context.Background(), items, joined, nil); err != nil {
			t.Fatalf("the verdict on %s: %v", written(joined), err)
		}
		swapped, err := Join(items, second, first)
		if err != nil {
			t.Fatal(err)
		}
		x, y := joined.AnswerSlots[0], swapped.AnswerSlots[0]
		_, xValue := x.Read.(*observationv1.AnswerSlotObservation_Value)
		_, yValue := y.Read.(*observationv1.AnswerSlotObservation_Value)
		if xValue != yValue || x.GetUnreadable() != y.GetUnreadable() {
			t.Fatalf("%s and %s join into %s, and the other way round into %s", written(first), written(second), written(joined), written(swapped))
		}
		itself, err := Join(items, first, first)
		if err != nil {
			t.Fatal(err)
		}
		want := proto.Clone(first).(*observationv1.Observation)
		if form, _ := verdict.NormalForm(verdict.AnswerSlots(items[0])[0], a); aIsValue && form == "" {
			want = observation(reason(1, 1, observationv1.Unreadable_UNREADABLE_BLANK))
		}
		if !proto.Equal(itself, want) {
			t.Fatalf("%s joined with itself is %s", written(first), written(itself))
		}
	})
}

// FuzzObservation reads any bytes as an observation of the tasks of the vectors: Validate, Join,
// Verdicts and Essay never panic, and a valid observation joins with itself and gets a verdict,
// or an end of its essay.
func FuzzObservation(f *testing.F) {
	tasks, observations := vectorTasks(f)
	for _, o := range observations {
		data, err := proto.Marshal(o)
		if err != nil {
			f.Fatal(err)
		}
		f.Add(data)
	}
	f.Fuzz(func(t *testing.T, data []byte) {
		o := &observationv1.Observation{}
		if proto.Unmarshal(data, o) != nil {
			return
		}
		for _, items := range tasks {
			err := Validate(items, o)
			if err != nil && !errors.Is(err, ErrNotValid) {
				t.Fatalf("Validate gives %v", err)
			}
			if taskOf(items).essay {
				end, essayErr := Essay(items, o)
				if (err == nil) != (essayErr == nil) || err == nil && end == "" {
					t.Fatalf("an essay valid by %v ends %q, %v", err, end, essayErr)
				}
				continue
			}
			joined, joinErr := Join(items, o, o)
			if (err == nil) != (joinErr == nil) {
				t.Fatalf("an observation valid by %v joins with itself: %v", err, joinErr)
			}
			if _, verdictErr := Verdicts(context.Background(), items, o, nil); (err == nil) != (verdictErr == nil) {
				t.Fatalf("an observation valid by %v gets its verdict: %v", err, verdictErr)
			}
			if joined != nil {
				if _, verdictErr := Verdicts(context.Background(), items, joined, nil); verdictErr != nil {
					t.Fatalf("the joined observation gets no verdict: %v", verdictErr)
				}
			}
		}
	})
}
