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
	"os"
	"path"
	"reflect"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/altnovatrail/spec"
	"github.com/altnovatrail/spec/contracts"
	"github.com/altnovatrail/spec/types/formatv0"

	"github.com/altnovatrail/module-kit/internal/schema"
	"github.com/altnovatrail/module-kit/internal/yaml"
	"github.com/altnovatrail/module-kit/module"
	"github.com/altnovatrail/module-kit/verdict"
)

// A vector is one case of the examples or the test vectors: a file, or a folder, named by its path
// in spec's repository, such as format/vectors/rules/yaml.encoding/byte-order-mark.yaml. The word
// case is Go's own, so the type takes the other one.
type vector struct {
	name string // its path in spec: the name of its subtest and of its line in testdata/skipped.txt
	kind string // what passing it means: a key of runners
	fsys fs.FS  // spec.Format or contracts.Vectors
	path string // the file or the folder within fsys
}

// The roots of the two file systems, as their paths in spec's repository begin.
const (
	formatRoot    = "format/"
	contractsRoot = "contracts/vectors/"
)

// errNoCode is what a kind gives until the task of the epic of code that brings the code for it.
var errNoCode = errors.New("module-kit has no code for this kind yet")

func noCode(vector) error { return errNoCode }

// rule passes a case of a format rule: refused with the ID of the rule its folder names and with no
// other, since a case breaks its rule and nothing else (format/README.md, section 9). The reader
// of YAML refuses a case of a rule yaml.* (R21). The check of a module refuses a case laid out as a
// module, and a case in a file is a file of the part its rule is of, which it checks on its own
// (R22); a rule that the check does not hold yet refuses neither.
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
		report, err := module.Check(folder)
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

// checkPart checks a file of a part on its own, as its examples are checked (format/README.md,
// section 8): a part of the layout of a module by the check of a module (R22), and the inventory
// of a package and the module index, which are no parts of a module, by the YAML rules and their
// schemas, the roots of inventory.schema.json and module-index.schema.json.
func checkPart(part string, data []byte) ([]finding, error) {
	if !slices.Contains(layoutParts, part) {
		_, found, err := readBySchema(data, part+".schema.json")
		return found, err
	}
	report, err := module.CheckFile(part, data)
	var found []finding
	for _, b := range report.Breaches {
		found = append(found, finding{b.Rule, b.Keyword, b.Pointer, b.String()})
	}
	return found, err
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
// within its part that module-kit holds (format/README.md, section 8).
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
// place it names (format/README.md, section 8). Two places compare as the JSON Pointers they
// encode, since a fragment of a URI may encode one in more than one way.
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

// The first line of an invalid example: # error: <keyword> at <location>.
var errorLine = regexp.MustCompile(`^# error: ([^ ]+) at (#[^ ]*)$`)

// expectedError reads the error that an invalid example names on its first line, its place as the
// JSON Pointer that the fragment encodes.
func expectedError(data []byte) (keyword, location string, err error) {
	first, _, _ := strings.Cut(string(data), "\n")
	// A line break of YAML may be CR LF too (YAML 1.2.2, section 5.4).
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
// percent-encoded (RFC 6901, section 6).
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
// each of its files of a part is valid against its schema (format/README.md, section 8).
func moduleExample(v vector) error {
	folder, err := fs.Sub(v.fsys, v.path)
	if err != nil {
		return err
	}
	report, err := module.Check(folder)
	if err != nil {
		return err
	}
	if len(report.Breaches) > 0 {
		return fmt.Errorf("an example of a whole module is refused:\n%s", report)
	}
	return nil
}

// openView passes a case of the open view: the open view of part.yaml is what open-view.yaml holds,
// whatever the order of keys (format/README.md, section 9). It compares twice. The walk of the
// schemas gives a JSON value, which is compared with open-view.yaml as read, as spec's format-check
// compares them: so a field left out is told from one at its zero value. And the function of the
// part, which takes part.yaml read by its type of formatv0, gives the type of formatv0/open, which
// is compared with open-view.yaml read strictly by that type: so the open view the library gives is
// the one the vector holds, in the type that holds it.
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
	return nil
}

// differ says that an open view is not the one open-view.yaml holds, writing both out.
func differ(what string, got, want any) error {
	gotText, _ := json.MarshalIndent(got, "", "  ")
	wantText, _ := json.MarshalIndent(want, "", "  ")
	return fmt.Errorf("%s\n%s\nand open-view.yaml holds\n%s", what, gotText, wantText)
}

// tiers passes a case of tiers: the tiers of skill_graph are skill_tiers (format/README.md, section
// 9), the vector read by its schema and its type of formatv0.
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
// trap that the vector gives each answer slot, and the item the verdict it gives (format/README.md,
// section 15). The vector is read by its schema and its type of formatv0. A predicate runs in the
// sandbox of Starlark, which comes with its own task.
func verdicts(v vector) error {
	var vv formatv0.VerdictVector
	if _, err := readVector(v.fsys, v.path, "verdict-vector.schema.json", &vv); err != nil {
		return err
	}
	slots := verdict.AnswerSlots(vv.Item)
	if slices.ContainsFunc(slots, func(s verdict.AnswerSlot) bool { return s.Check.Kind == formatv0.CheckKindPredicate }) {
		return errNoCode
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

// verdictOf gives the verdict of an answer slot that a vector gives, as the library gives it.
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

// slotReading writes the verdict of an answer slot as a vector holds it: the verdict, the reason and
// the trap by its ID.
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

// quoted writes a value read: in quotes, or null when nothing was read.
func quoted(value *string) string {
	if value == nil {
		return "null"
	}
	return fmt.Sprintf("%q", *value)
}

// runners passes a case of each kind: nil when module-kit gives what the case says, and an error
// that says why not otherwise. A task of the epic of code puts the code it brings in place of
// noCode, and takes the cases it passes out of testdata/skipped.txt (architecture 3.6; R20).
var runners = map[string]func(vector) error{
	// A part's valid example is valid against its schema and meets every rule within its part; an
	// invalid one gives the error its first line names; a whole module breaks no rule
	// (format/README.md, section 8).
	"examples/valid":   validExample,
	"examples/invalid": invalidExample,
	"examples/module":  moduleExample,
	// The test vectors (format/README.md, section 9).
	"rules":     rule,     // refused with the ID of the rule the case is filed under
	"open-view": openView, // the open view of part.yaml is open-view.yaml
	"verdicts":  verdicts, // each reading gets its verdicts, reasons and traps
	"leaks":     noCode,   // each text gives the keys and leak phrases it names
	"random":    noCode,   // each draw gives its value
	"templates": noCode,   // the count, and each seed's numbers and instance
	"tiers":     tiers,    // the tier of each skill
	"packages":  noCode,   // the package of module/ is package.yaml
	"indexes":   noCode,   // the index the release builds, or the rule that refuses it
	"versions":  noCode,   // the part of the version found, or the rule that refuses it
	// The joining of two readings and the verdict on the observation (contracts/README.md,
	// section 10).
	"observation/v1": noCode,
}

// The parts of a module's layout, whose open views have vectors (format/README.md, section 10).
var layoutParts = []string{"lesson", "manifest", "skill-graph", "solving-steps", "task", "techniques", "traps"}

// The parts that have examples of their own, each with its schema: the parts of a module's layout,
// the inventory of a package and the module index (format/README.md, sections 3, 25 and 26).
var exampleParts = slices.Concat(layoutParts, []string{"inventory", "module-index"})

// The check kinds that give verdicts, by the names of their folders; rubric gives none
// (format/README.md, section 15).
var verdictKinds = []string{"choice", "fraction", "number", "predicate"}

// The vectors whose cases are files, each a YAML file of its own.
var fileVectors = []string{"leaks", "random", "templates", "tiers"}

// The vectors whose cases are folders of their own kind, by the things each folder holds and
// nothing else, in order (format/README.md, section 9).
var folderVectors = map[string][]string{
	"packages": {"module", "package.yaml"},
	"indexes":  {"index.yaml", "modules"},
	"versions": {"module", "published", "version.yaml"},
}

// The files a case of the open view holds, and nothing else, in order (format/README.md, section 9).
var openViewFiles = []string{"open-view.yaml", "part.yaml"}

// The ID of a format rule: <part>.<name>, in kebab-case (format/README.md, section 7). The rules of
// the module index and of versions of a module have their vectors among those of indexes and
// versions, not under rules/ (section 9).
var ruleID = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*\.[a-z0-9]+(-[a-z0-9]+)*$`)

// ruleFolder says whether a folder of rules/ is named by the ID of a rule that has its vectors there.
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
	case name == "README.md" || name == "VERSION":
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
			slices.Contains(openViewFiles, parts[4]):
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
	case len(parts) == 3 && parts[0] == "eval" && parts[1] == "v1":
		// The rules the device and the runner share: canvas's CI and tutor's pass them, not
		// module-kit's (contracts/README.md, section 12).
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
			want = openViewFiles
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

// parseSkipped reads the list of the cases module-kit does not pass yet: the name of a case on each
// line, in byte order, each once, every line ending in a newline.
func parseSkipped(data []byte) (map[string]bool, error) {
	if len(data) == 0 {
		return nil, nil
	}
	if !bytes.HasSuffix(data, []byte("\n")) {
		return nil, errors.New("the last line ends in no newline")
	}
	lines := strings.Split(strings.TrimSuffix(string(data), "\n"), "\n")
	for n, line := range lines {
		if line == "" || strings.TrimSpace(line) != line {
			return nil, fmt.Errorf("line %d is %q, not the name of a case", n+1, line)
		}
		if n > 0 && line <= lines[n-1] {
			return nil, fmt.Errorf("line %d, %s, does not come after %s: the cases go in byte order, each once", n+1, line, lines[n-1])
		}
	}
	skipped := make(map[string]bool, len(lines))
	for _, line := range lines {
		skipped[line] = true
	}
	return skipped, nil
}

// unknownSkips names each skipped case that spec does not have, in order.
func unknownSkips(cases []vector, skipped map[string]bool) []string {
	names := map[string]bool{}
	for _, c := range cases {
		names[c.name] = true
	}
	var unknown []string
	for _, name := range slices.Sorted(maps.Keys(skipped)) {
		if !names[name] {
			unknown = append(unknown, name)
		}
	}
	return unknown
}

// outcome says what the run of a case means: a failure to report, or that the case is skipped. A
// skipped case is run too, so that one which passes leaves the list rather than stay in it.
func outcome(skipped bool, err error) (failure string, skip bool) {
	switch {
	case skipped && err == nil:
		return "it passes: take it out of testdata/skipped.txt", false
	case skipped:
		return "", true
	case err != nil:
		return err.Error(), false
	}
	return "", false
}

// Every example and test vector of spec.Format, and every vector of the observation in
// contracts.Vectors, passes, but those testdata/skipped.txt names, which do not pass yet. A file of
// a kind the test does not know fails it, and so does a skipped case that spec does not have or
// that passes (R16; R20).
func TestVectors(t *testing.T) {
	cases, err := inventory(spec.Format, contracts.Vectors)
	if err != nil {
		t.Fatal(err)
	}
	for _, kind := range kindsWithoutCase(cases) {
		t.Errorf("no case is of the kind %s", kind)
	}
	data, err := os.ReadFile("testdata/skipped.txt")
	if err != nil {
		t.Fatal(err)
	}
	skipped, err := parseSkipped(data)
	if err != nil {
		t.Fatalf("testdata/skipped.txt: %v", err)
	}
	for _, name := range unknownSkips(cases, skipped) {
		t.Errorf("testdata/skipped.txt names %s, which spec does not have", name)
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			failure, skip := outcome(skipped[c.name], runners[c.kind](c))
			if failure != "" {
				t.Fatal(failure)
			}
			if skip {
				t.Skip("in testdata/skipped.txt")
			}
		})
	}
}
