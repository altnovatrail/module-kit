package vectors

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"maps"
	"os"
	"path"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/altnovatrail/spec"
	"github.com/altnovatrail/spec/contracts"

	"github.com/altnovatrail/module-kit/internal/yaml"
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
// of YAML refuses by the rules yaml.* (R21); the rules of the parts wait for the tasks that bring
// them.
func rule(v vector) error {
	id := strings.Split(v.path, "/")[2]
	if !strings.HasPrefix(id, "yaml.") {
		return errNoCode
	}
	data, err := fs.ReadFile(v.fsys, v.path)
	if err != nil {
		return err
	}
	_, err = yaml.Read(data)
	var refusal *yaml.Refusal
	if !errors.As(err, &refusal) {
		return fmt.Errorf("not refused by %s: %v", id, err)
	}
	for _, b := range refusal.Breaches {
		if b.Rule != id {
			return fmt.Errorf("refused by more than %s:\n%v", id, err)
		}
	}
	return nil
}

// runners passes a case of each kind: nil when module-kit gives what the case says, and an error
// that says why not otherwise. A task of the epic of code puts the code it brings in place of
// noCode, and takes the cases it passes out of testdata/skipped.txt (architecture 3.6; R20).
var runners = map[string]func(vector) error{
	// A part's valid example is valid against its schema and meets every rule within its part; an
	// invalid one gives the error its first line names; a whole module breaks no rule
	// (format/README.md, section 8).
	"examples/valid":   noCode,
	"examples/invalid": noCode,
	"examples/module":  noCode,
	// The test vectors (format/README.md, section 9).
	"rules":     rule,   // refused with the ID of the rule the case is filed under
	"open-view": noCode, // the open view of part.yaml is open-view.yaml
	"verdicts":  noCode, // each reading gets its verdicts, reasons and traps
	"leaks":     noCode, // each text gives the keys and leak phrases it names
	"random":    noCode, // each draw gives its value
	"templates": noCode, // the count, and each seed's numbers and instance
	"tiers":     noCode, // the tier of each skill
	"packages":  noCode, // the package of module/ is package.yaml
	"indexes":   noCode, // the index the release builds, or the rule that refuses it
	"versions":  noCode, // the part of the version found, or the rule that refuses it
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
