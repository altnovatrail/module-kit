package yaml

import (
	"encoding/json"
	"errors"
	"maps"
	"os"
	"reflect"
	"slices"
	"testing"
)

// The YAML test suite, as testdata/yaml-test-suite/README.md says where it comes from: the cases
// YAML 1.2.2 reads in valid.json, and those it refuses in error.json.
const suiteRoot = "testdata/yaml-test-suite"

// A case of the suite.
type suiteCase struct {
	ID, Name, YAML string
	// The JSON value of what YAML reads, as text, when JSON can hold it.
	JSON *string
	// Whether YAML 1.2.2 refuses it: it is in error.json.
	Error bool
}

// suiteCases reads every case of the suite, the valid ones first.
func suiteCases() ([]suiteCase, error) {
	var cases []suiteCase
	for _, group := range []string{"valid", "error"} {
		data, err := os.ReadFile(suiteRoot + "/" + group + ".json")
		if err != nil {
			return nil, err
		}
		var list []suiteCase
		if err := json.Unmarshal(data, &list); err != nil {
			return nil, err
		}
		for i := range list {
			list[i].Error = group == "error"
		}
		cases = append(cases, list...)
	}
	return cases, nil
}

// The cases of the suite the reader does not pass, each with why. A case in the list that passes
// fails the test, so the list only shrinks (R21).
var suiteUnpassed = map[string]string{}

// Each case of the YAML test suite, data-2022-01-17, that YAML 1.2.2 refuses, the reader refuses
// too. Each case it reads, the reader reads to the value its JSON holds, or refuses by a rule
// yaml.* and not as not YAML (R21).
func TestYAMLTestSuite(t *testing.T) {
	cases, err := suiteCases()
	if err != nil {
		t.Fatal(err)
	}
	if len(cases) == 0 {
		t.Fatal("the suite holds no case")
	}
	ids := map[string]bool{}
	compared := 0
	for _, c := range cases {
		ids[c.ID] = true
		t.Run(c.ID, func(t *testing.T) {
			failure, wasCompared := c.check()
			if wasCompared {
				compared++
			}
			reason, unpassed := suiteUnpassed[c.ID]
			switch {
			case unpassed && failure == "":
				t.Fatal("it passes: take it out of suiteUnpassed")
			case unpassed:
				t.Skip(reason)
			case failure != "":
				t.Fatalf("%s:\n%s", c.Name, failure)
			}
		})
	}
	for _, id := range slices.Sorted(maps.Keys(suiteUnpassed)) {
		if !ids[id] {
			t.Errorf("suiteUnpassed names %s, which the suite does not have", id)
		}
	}
	t.Logf("%d cases, %d values compared with their JSON", len(cases), compared)
}

// check says why the reader does not pass a case, or "", and whether a value was compared.
func (c suiteCase) check() (failure string, compared bool) {
	value, err := Read([]byte(c.YAML))
	if c.Error {
		if err == nil {
			return "YAML 1.2.2 refuses it, and the reader reads it", false
		}
		return "", false
	}
	var refusal *Refusal
	if errors.As(err, &refusal) {
		for _, b := range refusal.Breaches {
			if b.Rule == "" {
				return "YAML 1.2.2 reads it, and the reader refuses it as not YAML:\n" + err.Error(), false
			}
		}
		return "", false
	}
	if err != nil {
		return err.Error(), false
	}
	if c.JSON == nil {
		// The case gives no JSON: its value is one JSON cannot hold.
		return "", false
	}
	var want any
	if err := json.Unmarshal([]byte(*c.JSON), &want); err != nil {
		return "its JSON: " + err.Error(), false
	}
	// Numbers compare by value: JSON reads every one as a double.
	written, err := json.Marshal(value)
	if err != nil {
		return err.Error(), false
	}
	var got any
	if err := json.Unmarshal(written, &got); err != nil {
		return err.Error(), false
	}
	if !reflect.DeepEqual(got, want) {
		return "read as " + string(written) + ", and its JSON is " + *c.JSON, true
	}
	return "", true
}
