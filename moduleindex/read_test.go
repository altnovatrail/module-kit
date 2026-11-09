package moduleindex

import (
	"errors"
	"io/fs"
	"reflect"
	"strings"
	"testing"

	"github.com/altnovatrail/spec"
	"github.com/altnovatrail/spec/types/formatv0"

	"github.com/altnovatrail/module-kit/internal/canonical"
	"github.com/altnovatrail/module-kit/internal/schema"
	"github.com/altnovatrail/module-kit/internal/yaml"
)

// examples gives the text of each example of the module index, valid or invalid, as JSON, by its
// path in spec.Format.
func examples(t testing.TB, kind string) map[string][]byte {
	t.Helper()
	found, err := fs.Glob(spec.Format, "examples/module-index/"+kind+"/*.yaml")
	if err != nil || len(found) == 0 {
		t.Fatalf("no examples of the module index %s: %v", kind, err)
	}
	texts := map[string][]byte{}
	for _, name := range found {
		data, err := fs.ReadFile(spec.Format, name)
		if err != nil {
			t.Fatal(err)
		}
		value, err := yaml.Read(data)
		if err != nil {
			t.Fatal(err)
		}
		if texts[name], err = canonical.Text(value); err != nil {
			t.Fatal(err)
		}
	}
	return texts
}

// A valid example reads as its schema and its type read it, and an invalid one is no index: none of
// them breaks the module index at a place it does not keep. The one exception is unknown-reason,
// which its schema refuses for a reason that it does not know: Read takes the version as withdrawn
// all the same, by that reason.
func TestReadExamples(t *testing.T) {
	unknown := mustRead(t, examples(t, "invalid")["examples/module-index/invalid/unknown-reason.yaml"])
	if w := unknown.Modules["example/smallest"]["0.1.0"].Withdrawn; w == nil || *w != "broken_widget" {
		t.Errorf("Read of unknown-reason withdraws example/smallest 0.1.0 by %v; want broken_widget", w)
	}
	if v, ok := Newest(unknown, "example/smallest"); ok {
		t.Errorf("Newest of unknown-reason picks the withdrawn %s", v)
	}
	for name, text := range examples(t, "valid") {
		value, _ := canonical.Parse(text)
		want, err := schema.Decode[formatv0.ModuleIndex](value)
		if err != nil {
			t.Fatal(err)
		}
		if got := mustRead(t, text); !reflect.DeepEqual(got, want) {
			t.Errorf("%s: Read = %+v; want %+v", name, got, want)
		}
	}
	for name, text := range examples(t, "invalid") {
		if strings.HasSuffix(name, "/unknown-reason.yaml") {
			continue
		}
		if _, err := Read(text); !errors.Is(err, ErrNotIndex) {
			t.Errorf("%s: Read = %v; want ErrNotIndex", name, err)
		}
	}
}

// Read reads an index of a later version by the places every version keeps: a member at another
// place, at the root or in an entry, is not read; a reason of withdrawal that a later version adds
// is a withdrawal, kept as the index writes it; an entry of a format version module-kit does not
// read stands among the others; and the text need not be the canonical text.
func TestReadLaterVersion(t *testing.T) {
	text := `{
	  "module_format_version": "1.4",
	  "signed_by": "registry",
	  "modules": {
	    "example/smallest": {
	      "0.1.0": {"module_format_version": "0.4", "package_digest": "sha256:fcbe6e68430776b0506c93ab51b155f8e7f92e2b3ecbe03e34e2afb91b4d4d08",
	                "release": "v0.1.0", "withdrawn": "unsafe_widget", "place": {"url": "https://example.org"}},
	      "2.0.0": {"module_format_version": "1.4", "package_digest": "sha256:860ba9ff55af668cb068de2cb8b6e99d86f8739ed40e00d3c40bba34d71f7555",
	                "release": "v3.0.0", "teaches": ["core.multiplication-table"]}
	    }
	  }
	}`
	want := formatv0.ModuleIndex{ModuleFormatVersion: "1.4", Modules: map[formatv0.ModuleID]map[formatv0.ModuleVersion]formatv0.Entry{
		"example/smallest": {
			"0.1.0": {ModuleFormatVersion: "0.4", PackageDigest: "sha256:fcbe6e68430776b0506c93ab51b155f8e7f92e2b3ecbe03e34e2afb91b4d4d08", Release: "v0.1.0",
				Withdrawn: reason("unsafe_widget")},
			"2.0.0": {ModuleFormatVersion: "1.4", PackageDigest: "sha256:860ba9ff55af668cb068de2cb8b6e99d86f8739ed40e00d3c40bba34d71f7555", Release: "v3.0.0",
				Teaches: []formatv0.AnchorSkillID{"core.multiplication-table"}},
		},
	}}
	if got := mustRead(t, []byte(text)); !reflect.DeepEqual(got, want) {
		t.Errorf("Read = %+v; want %+v", got, want)
	}
}

// A member at a place that every version keeps, not in its form, is no index of any version.
func TestReadRefuses(t *testing.T) {
	entry := `"module_format_version":"0.4","package_digest":"sha256:fcbe6e68430776b0506c93ab51b155f8e7f92e2b3ecbe03e34e2afb91b4d4d08","release":"v0.1.0"`
	for _, text := range []string{
		``,
		`[]`,
		`"module_format_version"`,
		`{"module_format_version":"0.4"}`,
		`{"modules":{"example/smallest":{"0.1.0":{` + entry + `}}}}`,
		`{"module_format_version":"0.4","modules":{}}`,
		`{"module_format_version":"0.4","modules":{"example/smallest":{}}}`,
		`{"module_format_version":"0.4","modules":{"example/smallest":[]}}`,
		`{"module_format_version":"0.4","modules":{"example/smallest":{"0.1.0":"v0.1.0"}}}`,
		`{"module_format_version":"0.4","modules":{"Example/Smallest":{"0.1.0":{` + entry + `}}}}`,
		`{"module_format_version":"0.4","modules":{"example/smallest":{"0.1":{` + entry + `}}}}`,
		`{"module_format_version":"0.3.0","modules":{"example/smallest":{"0.1.0":{` + entry + `}}}}`,
		`{"module_format_version":"0.4","modules":{"example/smallest":{"0.1.0":{` + entry + `,"withdrawn":true}}}}`,
		`{"module_format_version":"0.4","modules":{"example/smallest":{"0.1.0":{` + entry + `,"withdrawn":"wrong` + "\xff" + `key"}}}}`,
		`{"module_format_version":"0.4","modules":{"example/smallest":{"0.1.0":{` + entry + `,"teaches":[]}}}}`,
		`{"module_format_version":"0.4","modules":{"example/smallest":{"0.1.0":{` + entry + `,"teaches":["multiplication"]}}}}`,
		`{"module_format_version":"0.4","modules":{"example/smallest":{"0.1.0":{` + strings.Replace(entry, `"v0.1.0"`, `"0.1.0"`, 1) + `}}}}`,
		`{"module_format_version":"0.4","modules":{"example/smallest":{"0.1.0":{` + strings.Replace(entry, `sha256:fc`, `sha256:FC`, 1) + `}}}}`,
		`{"module_format_version":"0.4","modules":{"example/smallest":{"0.1.0":{` + strings.Replace(entry, `"module_format_version":"0.4",`, ``, 1) + `}}}}`,
	} {
		if index, err := Read([]byte(text)); !errors.Is(err, ErrNotIndex) || index.Modules != nil {
			t.Errorf("Read(%s) = %+v, %v; want ErrNotIndex", text, index, err)
		}
	}
}

// The version of a new attempt is the newest by Semantic Versioning, not withdrawn for any reason,
// whose format version module-kit reads.
func TestNewest(t *testing.T) {
	entry := func(format formatv0.ModuleFormatVersion, withdrawn *formatv0.WithdrawalReason) formatv0.Entry {
		return formatv0.Entry{ModuleFormatVersion: format, PackageDigest: smallest(t).Digest, Release: "v0.1.0", Withdrawn: withdrawn}
	}
	index := formatv0.ModuleIndex{ModuleFormatVersion: "0.4", Modules: map[formatv0.ModuleID]map[formatv0.ModuleVersion]formatv0.Entry{
		"example/a": {"1.2.0": entry("0.4", nil), "1.9.0": entry("0.4", nil), "1.10.0": entry("0.4", nil), "1.11.0": entry("0.4", reason("wrong_key")),
			"1.12.0": entry("0.4", reason("unsafe_widget")), "2.0.0": entry("0.5", nil), "1.13": entry("0.4", nil)},
		"example/b": {"0.1.0": entry("0.3", nil), "0.2.0": entry("0.4", reason("answer_leak"))},
		"example/c": {"0.1.0": entry("0.4", nil)},
	}}
	for _, c := range []struct {
		id   formatv0.ModuleID
		want formatv0.ModuleVersion
		ok   bool
	}{
		{"example/a", "1.10.0", true},
		{"example/b", "", false},
		{"example/c", "0.1.0", true},
		{"example/none", "", false},
	} {
		if got, ok := Newest(index, c.id); got != c.want || ok != c.ok {
			t.Errorf("Newest(%s) = %q, %t; want %q, %t", c.id, got, ok, c.want, c.ok)
		}
	}
}
