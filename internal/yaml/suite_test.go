package yaml

import (
	"encoding/json"
	"errors"
	"io/fs"
	"maps"
	"os"
	"path"
	"reflect"
	"slices"
	"testing"
)

// The YAML test suite, as testdata/yaml-test-suite/README.md says where it comes from.
const suiteRoot = "testdata/yaml-test-suite"

// The cases of the suite the reader does not pass, each with why. A case in the list that passes
// fails the test, so the list only shrinks (R21).
var suiteUnpassed = map[string]string{
	"2SXE": "go-yaml scans no anchor named a: and refuses the file as not YAML, though it breaks yaml.no-anchors",
	"W5VH": "go-yaml scans no anchor named :@*!$\"<foo>: and refuses the file as not YAML, though it breaks yaml.no-anchors",
}

// Each case of the YAML test suite, data-2022-01-17, that YAML 1.2.2 refuses, the reader refuses
// too. Each case it reads, the reader reads to the value its in.json holds, or refuses by a rule
// yaml.* and not as not YAML (R21).
func TestYAMLTestSuite(t *testing.T) {
	suite := os.DirFS(suiteRoot)
	cases, err := fs.Glob(suite, "*/in.yaml")
	if err != nil {
		t.Fatal(err)
	}
	nested, err := fs.Glob(suite, "*/*/in.yaml")
	if err != nil {
		t.Fatal(err)
	}
	cases = append(cases, nested...)
	slices.Sort(cases)
	if len(cases) == 0 {
		t.Fatal("the suite holds no case")
	}
	compared := 0
	for _, in := range cases {
		dir := path.Dir(in)
		t.Run(dir, func(t *testing.T) {
			failure, wasCompared := suiteCase(suite, dir)
			if wasCompared {
				compared++
			}
			reason, unpassed := suiteUnpassed[dir]
			switch {
			case unpassed && failure == "":
				t.Fatal("it passes: take it out of suiteUnpassed")
			case unpassed:
				t.Skip(reason)
			case failure != "":
				t.Fatal(failure)
			}
		})
	}
	for _, dir := range slices.Sorted(maps.Keys(suiteUnpassed)) {
		if _, err := fs.Stat(suite, dir+"/in.yaml"); err != nil {
			t.Errorf("suiteUnpassed names %s, which the suite does not have", dir)
		}
	}
	t.Logf("%d cases, %d values compared with in.json", len(cases), compared)
}

// suiteCase says why a case of the suite does not pass, or "", and whether a value was compared.
func suiteCase(suite fs.FS, dir string) (failure string, compared bool) {
	data, err := fs.ReadFile(suite, dir+"/in.yaml")
	if err != nil {
		return err.Error(), false
	}
	value, err := Read(data)
	if _, statErr := fs.Stat(suite, dir+"/error"); statErr == nil {
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
	text, err := fs.ReadFile(suite, dir+"/in.json")
	if err != nil {
		// The case gives no JSON: its value is one JSON cannot hold.
		return "", false
	}
	var want any
	if err := json.Unmarshal(text, &want); err != nil {
		return "in.json: " + err.Error(), false
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
		return "read as " + string(written) + ", and in.json holds " + string(text), true
	}
	return "", true
}
