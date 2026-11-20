package vectors

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"maps"
	"net/url"
	"path"
	"reflect"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/altnovatrail/spec"
	"github.com/altnovatrail/spec/contracts"
	"github.com/altnovatrail/spec/types/formatv0"

	"github.com/altnovatrail/module-kit/internal/canonical"
	"github.com/altnovatrail/module-kit/internal/random"
	"github.com/altnovatrail/module-kit/internal/schema"
	"github.com/altnovatrail/module-kit/internal/yaml"
	"github.com/altnovatrail/module-kit/leak"
	"github.com/altnovatrail/module-kit/module"
	"github.com/altnovatrail/module-kit/moduleindex"
	"github.com/altnovatrail/module-kit/modulepackage"
	"github.com/altnovatrail/module-kit/moduleversion"
	"github.com/altnovatrail/module-kit/openview"
	"github.com/altnovatrail/module-kit/template"
	"github.com/altnovatrail/module-kit/verdict"
)

// A vector is one case of the examples or the test vectors: a file, or a folder, named by its path
// in spec's repository, such as format/vectors/rules/yaml.encoding/byte-order-mark.yaml. The word
// case is Go's own, so the type takes the other one.
type vector struct {
	name string // its path in spec: the name of its subtest
	kind string // what passing it means: a key of runners
	fsys fs.FS  // spec.Format or contracts.Vectors
	path string // the file or the folder within fsys
}

// The roots of the two file systems, as their paths in spec's repository begin.
const (
	formatRoot    = "format/"
	contractsRoot = "contracts/vectors/"
)

// rule passes a case of a format rule: refused with the ID of the rule its folder names and with no
// other, since a case breaks its rule and nothing else. The reader of YAML refuses a case of a rule
// yaml.* (R21). The check of a module refuses a case laid out as a module, and a case in a file is
// a file of the part its rule is of, which it checks on its own (R22).
func rule(v vector) error {
	id := strings.Split(v.path, "/")[2]
	rules, refusal, err := refuse(v, id)
	if err != nil {
		return err
	}
	if len(rules) == 0 {
		return fmt.Errorf("not refused by %s", id)
	}
	for _, r := range rules {
		if r != id {
			return fmt.Errorf("refused by more than %s:\n%s", id, refusal)
		}
	}
	return nil
}

// refuse reads or checks a case of a rule, and gives the rule of each breach that refuses it, with
// "" for a file that is not YAML, and the refusal as text.
func refuse(v vector, id string) (rules []string, refusal string, err error) {
	info, err := fs.Stat(v.fsys, v.path)
	if err != nil {
		return nil, "", err
	}
	if info.IsDir() {
		folder, err := fs.Sub(v.fsys, v.path)
		if err != nil {
			return nil, "", err
		}
		report, err := module.Check(context.Background(), folder)
		return reportRules(report), report.String(), err
	}
	data, err := fs.ReadFile(v.fsys, v.path)
	if err != nil {
		return nil, "", err
	}
	part, _, _ := strings.Cut(id, ".")
	if part != "yaml" {
		report, err := module.CheckFile(part, data)
		return reportRules(report), report.String(), err
	}
	_, err = yaml.Read(data)
	var read *yaml.Refusal
	if !errors.As(err, &read) {
		return nil, "", err
	}
	for _, b := range read.Breaches {
		rules = append(rules, b.Rule)
	}
	return rules, read.Error(), nil
}

func reportRules(report module.Report) []string {
	rules := make([]string, len(report.Breaches))
	for i, b := range report.Breaches {
		rules[i] = b.Rule
	}
	return rules
}

// A finding is one thing that the check of a file of a part finds: the rule it breaks, or the
// keyword of its schema that fails at a place in it, or, with neither, that it is not YAML.
type finding struct {
	rule, keyword, pointer, text string
}

// checkPart checks a file of a part on its own: a part of the layout of a module by the check of a
// module (R22), and the inventory of a package and the module index, which are no parts of a
// module, by the YAML rules and the roots of their schemas (R23). A module index that its schema
// accepts is held also to what the schema cannot say (indexForm).
func checkPart(part string, data []byte) ([]finding, error) {
	if !slices.Contains(layoutParts, part) {
		value, found, err := readBySchema(data, part+".schema.json")
		if err == nil && len(found) == 0 && part == "module-index" {
			found = indexForm(value)
		}
		return found, err
	}
	report, err := module.CheckFile(part, data)
	var found []finding
	for _, b := range report.Breaches {
		found = append(found, finding{b.Rule, b.Keyword, b.Pointer, b.String()})
	}
	return found, err
}

// indexForm gives what a module index that its schema accepts breaks, as indexFormProblems of
// spec's format-check finds it: its root names another format version than FormatVersion, the one
// in format/VERSION; or an entry lists teaches out of the order of the code units of their IDs,
// which are ASCII (R35).
func indexForm(value any) []finding {
	index, err := schema.Decode[formatv0.ModuleIndex](value)
	if err != nil {
		return []finding{{text: err.Error()}}
	}
	var found []finding
	if index.ModuleFormatVersion != module.FormatVersion {
		found = append(found, finding{pointer: "#/module_format_version",
			text: fmt.Sprintf("the index names the format version %s, and format/VERSION %s", index.ModuleFormatVersion, module.FormatVersion)})
	}
	for _, id := range slices.Sorted(maps.Keys(index.Modules)) {
		for _, version := range slices.Sorted(maps.Keys(index.Modules[id])) {
			if teaches := index.Modules[id][version].Teaches; !slices.IsSorted(teaches) {
				found = append(found, finding{pointer: schema.Pointer([]string{"modules", string(id), string(version), "teaches"}),
					text: fmt.Sprintf("the teaches of %s %s are not in the order of code units: %v", id, version, teaches)})
			}
		}
	}
	return found
}

// readBySchema reads a file that is no part of a module's layout as the format has it read: by the
// YAML rules, then against the schema at schemaFile. It gives the value, and what refuses it: each
// breach of a YAML rule, or else each keyword of the schema that fails.
func readBySchema(data []byte, schemaFile string) (any, []finding, error) {
	value, err := yaml.Read(data)
	var refusal *yaml.Refusal
	var found []finding
	switch {
	case errors.As(err, &refusal):
		for _, b := range refusal.Breaches {
			found = append(found, finding{rule: b.Rule, text: b.String()})
		}
		return nil, found, nil
	case err != nil:
		return nil, nil, err
	}
	for _, e := range schema.Validate(schemaFile, value) {
		found = append(found, finding{keyword: e.Keyword, pointer: e.Pointer(), text: e.String()})
	}
	return value, found, nil
}

func texts(found []finding) string {
	lines := make([]string, len(found))
	for i, f := range found {
		lines[i] = f.text
	}
	return strings.Join(lines, "\n")
}

// examplePart reads an example of a part, examples/<part>/<valid or invalid>/<case>.yaml, and what
// its check finds.
func examplePart(v vector) ([]byte, []finding, error) {
	data, err := fs.ReadFile(v.fsys, v.path)
	if err != nil {
		return nil, nil, err
	}
	found, err := checkPart(strings.Split(v.path, "/")[1], data)
	return data, found, err
}

// validExample passes a valid example of a part: valid against its schema, and breaking no rule
// within its part that module-kit holds (R23).
func validExample(v vector) error {
	_, found, err := examplePart(v)
	if err != nil {
		return err
	}
	if len(found) > 0 {
		return fmt.Errorf("a valid example is refused:\n%s", texts(found))
	}
	return nil
}

// invalidExample passes an invalid example of a part: it breaks its schema alone, and the
// validator, reporting all its errors, reports one with the keyword its first line names at the
// place it names (R23). Two places compare as the JSON Pointers they encode, since a fragment of a
// URI may encode one in more than one way.
func invalidExample(v vector) error {
	data, found, err := examplePart(v)
	if err != nil {
		return err
	}
	keyword, location, err := expectedError(data)
	if err != nil {
		return err
	}
	if len(found) == 0 {
		return fmt.Errorf("passes its schema, and names %s at %s", keyword, location)
	}
	named := false
	for _, f := range found {
		if f.keyword == "" {
			return fmt.Errorf("breaks more than its schema:\n%s", texts(found))
		}
		if pointer, err := decodePlace(f.pointer); err == nil && f.keyword == keyword && pointer == location {
			named = true
		}
	}
	if !named {
		return fmt.Errorf("gives no %s at %s; it gives\n%s", keyword, location, texts(found))
	}
	return nil
}

var errorLine = regexp.MustCompile(`^# error: ([^ ]+) at (#[^ ]*)$`)

// expectedError reads the error that an invalid example names on its first line, its place as the
// JSON Pointer that the fragment encodes.
func expectedError(data []byte) (keyword, location string, err error) {
	first, _, _ := strings.Cut(string(data), "\n")
	// A line break of YAML may be CR LF too.
	first = strings.TrimSuffix(first, "\r")
	m := errorLine.FindStringSubmatch(first)
	if m == nil {
		return "", "", fmt.Errorf("the first line, %q, names no error as # error: <keyword> at <location>", first)
	}
	location, err = decodePlace(m[2])
	if err != nil {
		return "", "", fmt.Errorf("the first line, %q: %v", first, err)
	}
	return m[1], location, nil
}

// decodePlace gives the JSON Pointer that a place in URI fragment form encodes: # and the pointer,
// percent-encoded, as RFC 6901 has it.
func decodePlace(place string) (string, error) {
	pointer, err := url.PathUnescape(strings.TrimPrefix(place, "#"))
	switch {
	case !strings.HasPrefix(place, "#"):
		return "", fmt.Errorf("%q is no fragment of a URI", place)
	case err != nil:
		return "", err
	case pointer != "" && !strings.HasPrefix(pointer, "/"):
		return "", fmt.Errorf("%q encodes no JSON Pointer", place)
	}
	return pointer, nil
}

// moduleExample passes an example of a whole module: it breaks no rule that module-kit holds, and
// each of its files of a part is valid against its schema.
func moduleExample(v vector) error {
	folder, err := fs.Sub(v.fsys, v.path)
	if err != nil {
		return err
	}
	report, err := module.Check(context.Background(), folder)
	if err != nil {
		return err
	}
	if len(report.Breaches) > 0 {
		return fmt.Errorf("an example of a whole module is refused:\n%s", report)
	}
	return nil
}

// openView passes a case of the open view: the open view of part.yaml is what open-view.yaml holds,
// whatever the order of keys. It compares twice: the JSON value of the walk of the schemas with
// open-view.yaml as read, as spec's format-check does, so that a field left out is told from one at
// its zero value; and the value of formatv0/open that the function of the part gives with
// open-view.yaml read strictly by that type, so that the library gives the open view in the type
// that holds it (R24). Of a task, the digest of its open view is what digest.yaml holds (R51).
func openView(v vector) error {
	name := strings.Split(v.path, "/")[2]
	kind := partTypes[name]
	part := kind.value()
	value, err := readVector(v.fsys, path.Join(v.path, "part.yaml"), name+".schema.json", part)
	if err != nil {
		return err
	}
	want := kind.view()
	expected, err := readOpenView(v.fsys, path.Join(v.path, "open-view.yaml"), want)
	if err != nil {
		return err
	}
	if walked := schema.Open(name+".schema.json", value); !reflect.DeepEqual(walked, expected) {
		return differ("the walk of the schemas gives", walked, expected)
	}
	got, err := kind.open(part)
	if err != nil {
		return err
	}
	if !reflect.DeepEqual(got, want) {
		return differ("the open view of part.yaml is", got, want)
	}
	if name != "task" {
		return nil
	}
	var dv digestVector
	if _, err := readOpenView(v.fsys, path.Join(v.path, "digest.yaml"), &dv); err != nil {
		return err
	}
	digest, err := openview.TaskDigest(*part.(*formatv0.Task))
	if err != nil {
		return err
	}
	if digest != dv.OpenViewDigest {
		return fmt.Errorf("the digest of the open view of part.yaml is %s, and digest.yaml holds %s", digest, dv.OpenViewDigest)
	}
	return nil
}

// digestVector is digest.yaml of a case of the open view of a task, which has no schema of its own:
// the digest of the open view alone, as a sample of the eval protocol names the task it answers.
type digestVector struct {
	OpenViewDigest formatv0.Digest `json:"open_view_digest"`
}

func differ(what string, got, want any) error {
	gotText, _ := json.MarshalIndent(got, "", "  ")
	wantText, _ := json.MarshalIndent(want, "", "  ")
	return fmt.Errorf("%s\n%s\nand open-view.yaml holds\n%s", what, gotText, wantText)
}

// leaks passes a case of the leak check: in each text, the leak check finds every key and leak
// phrase of the task that the vector names, as the task writes them, and no other, in any order
// (R30).
func leaks(v vector) error {
	var lv formatv0.LeakVector
	if _, err := readVector(v.fsys, v.path, "leak-vector.schema.json", &lv); err != nil {
		return err
	}
	search, err := leak.For(lv.Task)
	if err != nil {
		return err
	}
	for n, text := range lv.Texts {
		var found []string
		for _, l := range search.In(text.Text) {
			if !slices.Contains(found, l.Written) {
				found = append(found, l.Written)
			}
		}
		slices.Sort(found)
		if want := slices.Sorted(slices.Values(text.Leaks)); !slices.Equal(found, want) {
			return fmt.Errorf("text %d, %q: the leak check finds %q, and the vector gives %q", n+1, text.Text, found, want)
		}
	}
	return nil
}

// randomDraws passes a case of the generator: the seed and the names, when given, derive the words
// of state, and the stream that the words start gives each draw's value, in order (R10). Its
// integers, written as strings, are read by strconv; a fraction is compared as a double (R31).
func randomDraws(v vector) error {
	var rv formatv0.RandomVector
	if _, err := readVector(v.fsys, v.path, "random-vector.schema.json", &rv); err != nil {
		return err
	}
	high, err := unsigned(rv.State.High)
	if err != nil {
		return err
	}
	low, err := unsigned(rv.State.Low)
	if err != nil {
		return err
	}
	if rv.Seed != nil {
		seed, err := unsigned(*rv.Seed)
		if err != nil {
			return err
		}
		h, l, err := random.Words(seed, rv.Names)
		if err != nil {
			return err
		}
		if h != high || l != low {
			return fmt.Errorf("the seed %d and the names %q derive the words %d and %d, and the vector gives %d and %d", seed, rv.Names, h, l, high, low)
		}
	}
	stream := random.New(high, low)
	for n, d := range rv.Draws {
		got, want, err := draw(stream, d)
		if err != nil {
			return fmt.Errorf("draw %d: %w", n+1, err)
		}
		if got != want {
			return fmt.Errorf("draw %d gives %s, and the vector gives %s", n+1, got, want)
		}
	}
	return nil
}

func draw(s *random.Stream, d formatv0.Draw) (got, want string, err error) {
	switch {
	case d.Integer != nil:
		return strconv.FormatUint(s.Integer(), 10), string(d.Integer.Value), nil
	case d.Below != nil:
		bound, err := unsigned(d.Below.Bound)
		if err != nil {
			return "", "", err
		}
		value, err := s.Below(bound)
		return strconv.FormatUint(value, 10), string(d.Below.Value), err
	case d.Between != nil:
		from, err := strconv.ParseInt(string(d.Between.From), 10, 64)
		if err != nil {
			return "", "", err
		}
		to, err := strconv.ParseInt(string(d.Between.To), 10, 64)
		if err != nil {
			return "", "", err
		}
		value, err := s.Between(from, to)
		return strconv.FormatInt(value, 10), string(d.Between.Value), err
	case d.Fraction != nil:
		value := s.Fraction()
		// strconv writes each double the shortest text that reads back to it, so two doubles are one
		// when it writes them alike.
		return strconv.FormatFloat(value, 'g', -1, 64), strconv.FormatFloat(d.Fraction.Value, 'g', -1, 64), nil
	case d.Permutation != nil:
		order, err := s.Permutation(int(d.Permutation.Count))
		return fmt.Sprint(order), fmt.Sprint(d.Permutation.Value), err
	}
	return "", "", errors.New("it names no function the runner knows")
}

func unsigned(text formatv0.Unsigned) (uint64, error) {
	return strconv.ParseUint(string(text), 10, 64)
}

// templates passes a case of templates: its template breaks no rule of a task, as the check of a
// task on its own reads it, on any instance; it has count instances; its solver, when it has one,
// gets the numbers of inputs, in any order; and each attempt's seed picks the instance of the
// attempt's numbers, the results of the solver among them as module-kit's own run of the solver
// finds them, which is the attempt's instance, whatever the order of keys, and has its open view,
// whose digest is the attempt's open_view_digest (R32; R33; R51).
func templates(v vector) error {
	var tv formatv0.TemplateVector
	value, err := readVector(v.fsys, v.path, "template-vector.schema.json", &tv)
	if err != nil {
		return err
	}
	switch inputs := slices.Sorted(slices.Values(tv.Inputs)); {
	case tv.Task.Template.Solver == nil && len(inputs) > 0:
		return fmt.Errorf("the vector gives the inputs %v of a template without a solver", inputs)
	case tv.Task.Template.Solver != nil && !slices.Equal(template.Inputs(tv.Task), inputs):
		return fmt.Errorf("the solver gets %v, and the vector gives %v", template.Inputs(tv.Task), inputs)
	}
	vector := value.(map[string]any)
	// A file written in JSON syntax is YAML, and the same rules apply to it.
	task, err := json.Marshal(vector["task"])
	if err != nil {
		return err
	}
	report, err := module.CheckFile("task", task)
	if err != nil {
		return err
	}
	if len(report.Breaches) > 0 {
		return fmt.Errorf("the template is refused:\n%s", report)
	}
	instances, breaches, err := template.Instances(context.Background(), tv.Task)
	if err != nil {
		return err
	}
	if len(breaches) > 0 || int64(len(instances)) != tv.Count {
		return fmt.Errorf("the template has %d instances, %v, and the vector gives %d", len(instances), breaches, tv.Count)
	}
	for n, attempt := range tv.Attempts {
		seed, err := unsigned(attempt.Seed)
		if err != nil {
			return err
		}
		picked, err := template.AttemptInstance(seed, tv.Task.ID, len(instances))
		if err != nil {
			return err
		}
		if numbers := instances[picked]; !maps.Equal(numbers, attempt.Numbers) {
			return fmt.Errorf("attempt %d: the seed %d picks the instance %v, and the vector gives %v", n+1, seed, numbers, attempt.Numbers)
		}
		got, err := template.Instance(tv.Task, instances[picked])
		if err != nil {
			return fmt.Errorf("attempt %d: %w", n+1, err)
		}
		want := vector["attempts"].([]any)[n].(map[string]any)["instance"]
		if gotValue := schema.ValueOf(got); !reflect.DeepEqual(gotValue, want) {
			gotText, _ := json.MarshalIndent(gotValue, "", "  ")
			wantText, _ := json.MarshalIndent(want, "", "  ")
			return fmt.Errorf("attempt %d: the instance is\n%s\nand the vector gives\n%s", n+1, gotText, wantText)
		}
		if _, err := openview.Task(got); err != nil {
			return fmt.Errorf("attempt %d: %w", n+1, err)
		}
		digest, err := openview.TaskDigest(got)
		if err != nil {
			return fmt.Errorf("attempt %d: %w", n+1, err)
		}
		if digest != attempt.OpenViewDigest {
			return fmt.Errorf("attempt %d: the digest of the instance's open view is %s, and the vector gives %s", n+1, digest, attempt.OpenViewDigest)
		}
	}
	return nil
}

// packages passes a case of packages: the package that module-kit builds of module/ is package.yaml
// (R34). The canonical view, the listing of the device part and the inventory compare as their
// canonical texts, so whatever the order of keys, and their digests and sizes, and the digest of
// the package, pin those texts byte for byte. The results of a solver are module-kit's own run of
// it, compared with the vector's in the canonical view. The package as files is read back by the
// chain of digests from its digest, and gives the same package.
func packages(v vector) error {
	var pv formatv0.PackageVector
	value, err := readVector(v.fsys, path.Join(v.path, "package.yaml"), "package-vector.schema.json", &pv)
	if err != nil {
		return err
	}
	folder, err := fs.Sub(v.fsys, path.Join(v.path, "module"))
	if err != nil {
		return err
	}
	pkg, err := modulepackage.Build(context.Background(), folder)
	if err != nil {
		return err
	}
	want := value.(map[string]any)
	for _, part := range []struct{ member, file string }{
		{"server_part", modulepackage.ServerFile},
		{"device_part", modulepackage.DeviceFile},
		{"inventory", modulepackage.InventoryFile},
	} {
		text, err := canonical.Text(want[part.member])
		if err != nil {
			return fmt.Errorf("%s of the vector: %w", part.member, err)
		}
		if got := pkg.Files[part.file]; !bytes.Equal(got, text) {
			return fmt.Errorf("%s is\n%s\nand the vector's %s is\n%s", part.file, got, part.member, text)
		}
	}
	switch {
	case pkg.Digest != pv.Digest:
		return fmt.Errorf("the digest of the package is %s, and the vector gives %s", pkg.Digest, pv.Digest)
	case !reflect.DeepEqual(pkg.Inventory, pv.Inventory):
		return fmt.Errorf("the inventory is %+v, and the vector gives %+v", pkg.Inventory, pv.Inventory)
	case canonical.Descriptor(pkg.Files[modulepackage.ServerFile]) != pv.Inventory.ServerPart:
		return fmt.Errorf("the server part is not the one the vector's inventory names")
	case canonical.Descriptor(pkg.Files[modulepackage.DeviceFile]) != pv.Inventory.DevicePart:
		return fmt.Errorf("the device part is not the one the vector's inventory names")
	}
	read, err := modulepackage.Read(fstest.MapFS(asFiles(pkg.Files)), pv.Digest)
	if err != nil {
		return fmt.Errorf("the package as files: %w", err)
	}
	if !reflect.DeepEqual(read, pkg) {
		return fmt.Errorf("the package as files reads back as another package")
	}
	return nil
}

// indexes passes a case of the module index: the release that index.yaml gives — the previous
// index, unless it is the first release, its tag and its withdrawals — with the package of each
// module of modules/ builds the index that index.yaml gives, byte for byte as its canonical text, or
// is refused by the rule it names and by no other (R35). The previous index goes to the build as
// its canonical text, as a release of modules holds it, and it reads back, as the index built does,
// as the vector gives it.
func indexes(v vector) error {
	var iv formatv0.IndexVector
	value, err := readVector(v.fsys, path.Join(v.path, "index.yaml"), "index-vector.schema.json", &iv)
	if err != nil {
		return err
	}
	release := moduleindex.Release{First: iv.Previous == nil, Tag: iv.Release, Withdrawals: iv.Withdrawals}
	modules, err := fs.ReadDir(v.fsys, path.Join(v.path, "modules"))
	if err != nil {
		return err
	}
	for _, m := range modules {
		folder, err := fs.Sub(v.fsys, path.Join(v.path, "modules", m.Name()))
		if err != nil {
			return err
		}
		pkg, err := modulepackage.Build(context.Background(), folder)
		if err != nil {
			return fmt.Errorf("modules/%s: %w", m.Name(), err)
		}
		release.Packages = append(release.Packages, pkg)
	}
	want := value.(map[string]any)
	if iv.Previous != nil {
		if release.Previous, err = canonical.Text(want["previous"]); err != nil {
			return fmt.Errorf("previous of the vector: %w", err)
		}
		if err := readsBack(release.Previous, *iv.Previous); err != nil {
			return fmt.Errorf("the previous index: %w", err)
		}
	}
	text, err := moduleindex.Build(release)
	if iv.Refusal != nil {
		var refusal *moduleindex.Refusal
		if !errors.As(err, &refusal) {
			return fmt.Errorf("not refused by %s: %v", *iv.Refusal, err)
		}
		for _, b := range refusal.Breaches {
			if b.Rule != string(*iv.Refusal) {
				return fmt.Errorf("refused by more than %s:\n%v", *iv.Refusal, refusal)
			}
		}
		return nil
	}
	if err != nil {
		return err
	}
	index, err := canonical.Text(want["index"])
	if err != nil {
		return fmt.Errorf("index of the vector: %w", err)
	}
	if !bytes.Equal(text, index) {
		return fmt.Errorf("the index is\n%s\nand the vector's is\n%s", text, index)
	}
	return readsBack(text, *iv.Index)
}

// versions passes a case of versions of a module: the new version, module/, compared with every
// published version, published/X.Y.Z/, gives the part of the version that version.yaml gives, or is
// refused by the rule it names. A published version laid out as a module is built into its
// package, which goes to the comparison as files, without device/, as the CI of modules gives it
// (R36); one of an earlier version of the format is its files as they are, by the digest of
// inventory.json. Each folder is named by the version its package names.
func versions(v vector) error {
	var vv formatv0.VersionVector
	if _, err := readVector(v.fsys, path.Join(v.path, "version.yaml"), "version-vector.schema.json", &vv); err != nil {
		return err
	}
	folder, err := fs.Sub(v.fsys, path.Join(v.path, "module"))
	if err != nil {
		return err
	}
	next, err := modulepackage.Build(context.Background(), folder)
	if err != nil {
		return fmt.Errorf("module: %w", err)
	}
	entries, err := fs.ReadDir(v.fsys, path.Join(v.path, "published"))
	if err != nil {
		return err
	}
	var published []moduleversion.Published
	for _, e := range entries {
		dir := path.Join(v.path, "published", e.Name())
		files, err := fs.Sub(v.fsys, dir)
		if err != nil {
			return err
		}
		if _, err := fs.Stat(files, "manifest.yaml"); err != nil {
			inventory, err := fs.ReadFile(files, modulepackage.InventoryFile)
			if err != nil {
				return err
			}
			published = append(published, moduleversion.Published{Digest: canonical.Digest(inventory), Files: files})
			continue
		}
		pkg, err := modulepackage.Build(context.Background(), files)
		if err != nil {
			return fmt.Errorf("published/%s: %w", e.Name(), err)
		}
		if string(pkg.Inventory.Version) != e.Name() {
			return fmt.Errorf("published/%s holds the version %s", e.Name(), pkg.Inventory.Version)
		}
		parts := map[string][]byte{}
		for _, name := range []string{modulepackage.InventoryFile, modulepackage.ServerFile, modulepackage.DeviceFile} {
			parts[name] = pkg.Files[name]
		}
		published = append(published, moduleversion.Published{Digest: pkg.Digest, Files: fstest.MapFS(asFiles(parts))})
	}
	c, err := moduleversion.Compare(next, published)
	if err != nil {
		return err
	}
	var rules []string
	for _, b := range c.Breaches {
		rules = append(rules, b.Rule)
	}
	if vv.Refusal != nil {
		if len(rules) == 0 || slices.ContainsFunc(rules, func(r string) bool { return r != string(*vv.Refusal) }) {
			return fmt.Errorf("refused by %v, and the vector gives %s alone:\n%s", rules, *vv.Refusal, c)
		}
		return nil
	}
	if len(rules) > 0 {
		return fmt.Errorf("refused by %v, and the vector gives the part %s:\n%s", rules, *vv.VersionPart, c)
	}
	if c.Part.String() != string(*vv.VersionPart) {
		return fmt.Errorf("the comparison finds a %s change, and the vector gives %s:\n%s", c.Part, *vv.VersionPart, c)
	}
	return nil
}

func readsBack(text []byte, want formatv0.ModuleIndex) error {
	got, err := moduleindex.Read(text)
	if err != nil {
		return err
	}
	if !reflect.DeepEqual(got, want) {
		return fmt.Errorf("it reads back as %+v, and the vector gives %+v", got, want)
	}
	return nil
}

func asFiles(files map[string][]byte) map[string]*fstest.MapFile {
	m := make(map[string]*fstest.MapFile, len(files))
	for name, data := range files {
		m[name] = &fstest.MapFile{Data: data}
	}
	return m
}

// tiers passes a case of tiers: the tiers of skill_graph are skill_tiers (R25).
func tiers(v vector) error {
	var tv formatv0.TierVector
	if _, err := readVector(v.fsys, v.path, "tier-vector.schema.json", &tv); err != nil {
		return err
	}
	got, err := module.Tiers(tv.SkillGraph)
	if err != nil {
		return err
	}
	if !maps.Equal(got, tv.SkillTiers) {
		return fmt.Errorf("the tiers are %v, and the vector gives %v", got, tv.SkillTiers)
	}
	return nil
}

// verdicts passes a case of verdicts: each reading of the item gets the verdict, the reason and the
// trap that the vector gives each answer slot, and the item the verdict it gives (R27). A predicate
// runs in the sandbox of Starlark, with the limits of a predicate (R28).
func verdicts(v vector) error {
	var vv formatv0.VerdictVector
	if _, err := readVector(v.fsys, v.path, "verdict-vector.schema.json", &vv); err != nil {
		return err
	}
	for n, reading := range vv.Readings {
		values := make([]*string, len(reading.AnswerSlots))
		for i, slot := range reading.AnswerSlots {
			values[i] = slot.Value
		}
		got, err := verdict.OfItem(context.Background(), vv.Item, values)
		if err != nil {
			return fmt.Errorf("reading %d: %w", n+1, err)
		}
		for i, want := range reading.AnswerSlots {
			if g, w := slotReading(got.Slots[i]), slotReading(verdictOf(want)); g != w {
				return fmt.Errorf("reading %d, answer slot %d, the value %s: %s, and the vector gives %s", n+1, i+1, quoted(want.Value), g, w)
			}
		}
		if string(got.Verdict) != string(reading.Verdict) {
			return fmt.Errorf("reading %d: the item is %s, and the vector gives %s", n+1, got.Verdict, reading.Verdict)
		}
	}
	return nil
}

func verdictOf(s formatv0.SlotReading) verdict.SlotVerdict {
	v := verdict.SlotVerdict{Verdict: verdict.Verdict(s.Verdict)}
	if s.Reason != nil {
		v.Reason = verdict.Reason(*s.Reason)
	}
	if s.Trap != nil {
		v.Trap = &formatv0.TrapValue{Trap: *s.Trap}
	}
	return v
}

func slotReading(v verdict.SlotVerdict) string {
	text := string(v.Verdict)
	if v.Reason != "" {
		text += ", " + string(v.Reason)
	}
	if v.Trap != nil {
		text += ", trap " + string(v.Trap.Trap)
	}
	return text
}

func quoted(value *string) string {
	if value == nil {
		return "null"
	}
	return fmt.Sprintf("%q", *value)
}

// runners passes a case of each kind: nil when module-kit gives what the case says, and an error
// that says why not otherwise (R20).
var runners = withPaper(map[string]func(vector) error{
	"examples/valid":   validExample,
	"examples/invalid": invalidExample,
	"examples/module":  moduleExample,
	"rules":            rule,
	"open-view":        openView,
	"verdicts":         verdicts,
	"leaks":            leaks,
	"random":           randomDraws,
	"templates":        templates,
	"pages":            pages,
	"tiers":            tiers,
	"packages":         packages,
	"indexes":          indexes,
	"versions":         versions,
	"observation/v1":   observations,
})

// withPaper adds to runners each file of paper/v1/, a kind of its own.
func withPaper(runners map[string]func(vector) error) map[string]func(vector) error {
	for name, run := range paperRunners {
		runners[paperKind(name)] = run
	}
	return runners
}

func paperKind(name string) string { return "paper/v1/" + name }

// The parts of a module's layout, whose open views have vectors.
var layoutParts = []string{"lesson", "manifest", "skill-graph", "solving-steps", "task", "techniques", "traps"}

// The parts that have examples of their own, each with its schema: the parts of a module's layout,
// the inventory of a package and the module index.
var exampleParts = slices.Concat(layoutParts, []string{"inventory", "module-index"})

// The check kinds that give verdicts, by the names of their folders; rubric gives none.
var verdictKinds = []string{"choice", "fraction", "number", "predicate", "time"}

var fileVectors = []string{"leaks", "pages", "random", "templates", "tiers"}

// The vectors whose cases are folders of their own kind, by the things each folder holds and
// nothing else, in order.
var folderVectors = map[string][]string{
	"packages": {"module", "package.yaml"},
	"indexes":  {"index.yaml", "modules"},
	"versions": {"module", "published", "version.yaml"},
}

// The files a case of the open view holds, and nothing else, in order.
func openViewFiles(part string) []string {
	if part == "task" {
		return []string{"digest.yaml", "open-view.yaml", "part.yaml"}
	}
	return []string{"open-view.yaml", "part.yaml"}
}

// The ID of a format rule: <part>.<name>, in kebab-case. The rules of the module index and of
// versions of a module have their vectors among those of indexes and versions, not under rules/.
var ruleID = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*\.[a-z0-9]+(-[a-z0-9]+)*$`)

func ruleFolder(name string) bool {
	return ruleID.MatchString(name) && !strings.HasPrefix(name, "module-index.") &&
		!strings.HasPrefix(name, "module-version.")
}

// classifyFormat says where a file of spec.Format belongs: the path of its case within the folder
// and the case's kind; no kind for a file of no case — the text, VERSION and the schemas —; and
// false for a file of no kind the test knows.
func classifyFormat(name string) (casePath, kind string, known bool) {
	parts := strings.Split(name, "/")
	yaml := strings.HasSuffix(name, ".yaml")
	switch {
	case name == "README.md" || name == "VERSION" || name == "page.yaml":
		return "", "", true
	case len(parts) == 2 && parts[0] == "schemas":
		return "", "", strings.HasSuffix(parts[1], ".schema.json")
	case parts[0] == "examples":
		switch {
		case len(parts) >= 5 && parts[1] == "module" && parts[2] == "valid":
			return path.Join(parts[:4]...), "examples/module", true
		case len(parts) == 4 && slices.Contains(exampleParts, parts[1]) && (parts[2] == "valid" || parts[2] == "invalid") && yaml:
			return name, "examples/" + parts[2], true
		}
	case parts[0] == "vectors" && len(parts) >= 3:
		vectors := parts[1]
		switch {
		case vectors == "rules" && len(parts) >= 4 && ruleFolder(parts[2]):
			// A file of its own, or a folder laid out as a module.
			if len(parts) == 4 && !yaml {
				return "", "", false
			}
			return path.Join(parts[:4]...), vectors, true
		case vectors == "open-view" && len(parts) == 5 && slices.Contains(layoutParts, parts[2]) &&
			slices.Contains(openViewFiles(parts[2]), parts[4]):
			return path.Join(parts[:4]...), vectors, true
		case vectors == "verdicts" && len(parts) == 4 && slices.Contains(verdictKinds, parts[2]) && yaml:
			return name, vectors, true
		case slices.Contains(fileVectors, vectors) && len(parts) == 3 && yaml:
			return name, vectors, true
		case len(parts) >= 4 && slices.Contains(folderVectors[vectors], parts[3]) &&
			(len(parts) == 4) == strings.HasSuffix(parts[3], ".yaml"):
			return path.Join(parts[:3]...), vectors, true
		}
	}
	return "", "", false
}

// classifyContracts does the same for a file of contracts.Vectors.
func classifyContracts(name string) (casePath, kind string, known bool) {
	parts := strings.Split(name, "/")
	switch {
	case len(parts) == 3 && parts[0] == "observation" && parts[1] == "v1" && strings.HasSuffix(name, ".json"):
		return name, "observation/v1", true
	case len(parts) == 3 && parts[0] == "paper" && parts[1] == "v1" && strings.HasSuffix(parts[2], ".json") && paperRunners[strings.TrimSuffix(parts[2], ".json")] != nil:
		return name, paperKind(strings.TrimSuffix(parts[2], ".json")), true
	case len(parts) == 3 && parts[0] == "eval" && parts[1] == "v1":
		// The rules the device and the runner share: canvas's CI and tutor's pass them, not
		// module-kit's (R20).
		return "", "", true
	}
	return "", "", false
}

// inventory sorts every file of the format and of the contracts' vectors into the cases module-kit
// passes, in the order of their names. It refuses a file of no kind it knows, naming each one: a
// new kind comes with a move to a new tag of spec, and with the code that passes it.
func inventory(format, vectors fs.FS) ([]vector, error) {
	found := map[string]vector{}
	var unknown []string
	walk := func(root string, fsys fs.FS, classify func(string) (string, string, bool)) error {
		return fs.WalkDir(fsys, ".", func(name string, entry fs.DirEntry, err error) error {
			if err != nil || entry.IsDir() {
				return err
			}
			casePath, kind, known := classify(name)
			switch {
			case !known:
				unknown = append(unknown, root+name)
			case kind != "":
				found[root+casePath] = vector{name: root + casePath, kind: kind, fsys: fsys, path: casePath}
			}
			return nil
		})
	}
	if err := walk(formatRoot, format, classifyFormat); err != nil {
		return nil, err
	}
	if err := walk(contractsRoot, vectors, classifyContracts); err != nil {
		return nil, err
	}
	if len(unknown) > 0 {
		return nil, fmt.Errorf("files of no kind the test knows:\n%s", strings.Join(unknown, "\n"))
	}
	// A folder of its own kind holds every thing its kind names: the walk refused any other, and a
	// thing left out makes no case.
	var incomplete []string
	for _, c := range found {
		want := folderVectors[c.kind]
		if c.kind == "open-view" {
			want = openViewFiles(strings.Split(c.path, "/")[2])
		}
		if want == nil {
			continue
		}
		entries, err := fs.ReadDir(c.fsys, c.path)
		if err != nil {
			return nil, err
		}
		var held []string
		for _, entry := range entries {
			held = append(held, entry.Name())
		}
		if !slices.Equal(held, want) {
			incomplete = append(incomplete, fmt.Sprintf("%s holds %v, and a case of %s holds %v", c.name, held, c.kind, want))
		}
	}
	if len(incomplete) > 0 {
		slices.Sort(incomplete)
		return nil, fmt.Errorf("cases that do not hold what their kind does:\n%s", strings.Join(incomplete, "\n"))
	}
	cases := make([]vector, 0, len(found))
	for _, name := range slices.Sorted(maps.Keys(found)) {
		cases = append(cases, found[name])
	}
	return cases, nil
}

// kindsWithoutCase names each kind of runners that no case has, in order: a kind the test knows
// and spec no longer has, or a file system that came empty.
func kindsWithoutCase(cases []vector) []string {
	seen := map[string]bool{}
	for _, c := range cases {
		seen[c.kind] = true
	}
	var missing []string
	for _, kind := range slices.Sorted(maps.Keys(runners)) {
		if !seen[kind] {
			missing = append(missing, kind)
		}
	}
	return missing
}

// Every example and test vector of spec.Format, and every vector of the observation and of the key
// of a task from paper in contracts.Vectors, passes, and none is skipped, so a vector a new tag of spec brings comes with
// its code. A file of a kind the test does not know fails it, and so does a kind left with no case
// (R20; R42). The cases share nothing but the read-only file systems and tables of the test, so
// they run in parallel, but for the verdicts of a predicate, which run one at a time and before the
// rest, since a parallel subtest waits for its parent's function to return: two of them take the
// predicate to its million steps, which under -race takes a good part of the sandbox's clock of a
// second, and cases running beside them can push one past it (sandbox.Predicate; R59). ci-test
// runs them in a go test of their own, after the rest, so that no other package runs beside them
// (R60).
func TestVectors(t *testing.T) {
	cases, err := inventory(spec.Format, contracts.Vectors)
	if err != nil {
		t.Fatal(err)
	}
	for _, kind := range kindsWithoutCase(cases) {
		t.Errorf("no case is of the kind %s", kind)
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if !strings.HasPrefix(c.name, formatRoot+"vectors/verdicts/predicate/") {
				t.Parallel()
			}
			run, ok := runners[c.kind]
			if !ok {
				t.Fatalf("no runner passes the kind %s", c.kind)
			}
			if err := run(c); err != nil {
				t.Fatal(err)
			}
		})
	}
}
