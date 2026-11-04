package module

import (
	"io/fs"
	"maps"
	"slices"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/altnovatrail/spec"
	"github.com/altnovatrail/spec/types/formatv0"

	"github.com/altnovatrail/module-kit/internal/yaml"
)

// examplePart gives an example of a part from spec.Format, examples/<part>/valid/<name>.yaml.
func examplePart(part, name string) *fstest.MapFile {
	data, err := fs.ReadFile(spec.Format, "examples/"+part+"/valid/"+name+".yaml")
	if err != nil {
		panic(err)
	}
	return bytesFile(data)
}

// A skill graph that breaks every rule within a skill graph: a skill that needs itself, two in a
// cycle beside a skill that needs one of them and is in none, a prerequisite and a replacement that
// name no skill, an anchor skill that the module teaches, and a retired ID given again.
const brokenGraph = `skills:
  a:
    title: A
    prerequisites: [a, gone, core.division-with-remainder]
    assessability: verdict
  b:
    title: B
    prerequisites: [c]
    teaches: [core.division-with-remainder]
    assessability: verdict
  c:
    title: C
    prerequisites: [b]
    assessability: verdict
  d:
    title: D
    prerequisites: [b, core.multiplication-table]
    assessability: verdict
retired_skills:
  d:
    replaced_by: []
  e:
    replaced_by: [a, lost]
`

// The bands of the base manifest, which a case puts others in the place of.
const baseBands = "bands:\n  - id: grades-3-4\n    title: Grades 3–4\n    grades: [3, 4]\n"

func bands(text string) *fstest.MapFile {
	if !strings.Contains(manifest, baseBands) {
		panic("the manifest of base has other bands")
	}
	return file(strings.Replace(manifest, baseBands, text, 1))
}

// The cases hold each rule of the parts at each place it is broken, on a whole module.
func TestPartRules(t *testing.T) {
	for _, c := range []struct {
		name    string
		fsys    fstest.MapFS
		refused []string
	}{
		{"bands with one ID, without grades and with a grade that the audience does not name", module(fstest.MapFS{
			manifestPath: bands("bands:\n  - id: grade-3\n    title: Grade 3\n    grades: [3, 5]\n  - id: grade-3\n    title: Again\n"),
		}), []string{
			"manifest.yaml #/audience/grades/1 manifest.bands", "manifest.yaml #/bands/0/grades/1 manifest.bands",
			"manifest.yaml #/bands/1 manifest.bands", "manifest.yaml #/bands/1/id manifest.bands",
		}},
		{"a grade in two bands", module(fstest.MapFS{
			manifestPath: bands(baseBands + "  - id: again\n    title: Again\n    grades: [4]\n"),
		}), []string{"manifest.yaml #/audience/grades/1 manifest.bands"}},
		{"grades in a module that names none", module(fstest.MapFS{
			manifestPath: file(strings.Replace(manifest, "  grades: [3, 4]\nbands", "bands", 1)),
		}), []string{"manifest.yaml #/bands/0/grades manifest.bands"}},
		{"a module that names grades and has no bands", module(fstest.MapFS{
			manifestPath:                  bands(""),
			"tasks/camp-days-1.yaml":      nil,
			"tasks/gaps-ladder-1.yaml":    nil,
			"lessons/days-and-dates.yaml": nil,
		}), nil},
		{"the rules within a skill graph, and those between parts that it breaks", module(fstest.MapFS{"skill-graph.yaml": file(brokenGraph)}), []string{
			"manifest.yaml # manifest.dependencies", "manifest.yaml # manifest.dependencies",
			"skill-graph.yaml #/retired_skills/d skill-graph.retired-ids",
			"skill-graph.yaml #/retired_skills/e/replaced_by/1 skill-graph.known-skills",
			"skill-graph.yaml #/skills/a/prerequisites skill-graph.acyclic",
			"skill-graph.yaml #/skills/a/prerequisites/1 skill-graph.known-skills",
			"skill-graph.yaml #/skills/a/prerequisites/2 skill-graph.own-anchors",
			"skill-graph.yaml #/skills/b/prerequisites skill-graph.acyclic",
			"skill-graph.yaml #/skills/c/prerequisites skill-graph.acyclic",
			"traps.yaml #/traps/off-by-one/skills/0 traps.known-skills",
		}},
		{"dependencies that leave one out and name one that no skill needs", module(fstest.MapFS{
			"skill-graph.yaml": file(strings.Replace(string(base["skill-graph.yaml"].Data), "prerequisites: [counting-gaps]",
				"prerequisites: [counting-gaps, core.multiplication-table, core.division-with-remainder]", 1)),
			manifestPath: file(manifest + "dependencies: [core.multiplication-table, core.linear-equations]\n"),
		}), []string{"manifest.yaml #/dependencies manifest.dependencies", "manifest.yaml #/dependencies/1 manifest.dependencies"}},
		{"no dependencies where a skill needs one", module(fstest.MapFS{
			"skill-graph.yaml": file(strings.Replace(string(base["skill-graph.yaml"].Data), "prerequisites: [counting-gaps]",
				"prerequisites: [counting-gaps, core.multiplication-table]", 1)),
		}), []string{"manifest.yaml # manifest.dependencies"}},
		{"dependencies with no skill graph", smallest(fstest.MapFS{
			manifestPath: file(string(example("smallest")[manifestPath].Data) + "dependencies: [core.linear-equations]\n"),
		}), []string{"manifest.yaml #/dependencies/0 manifest.dependencies"}},
		{"no rung templates, and a task with a verdict", module(fstest.MapFS{
			manifestPath: file(manifest[:strings.Index(manifest, "rung_templates:")]),
		}), []string{"manifest.yaml # manifest.rung-templates"}},
		{"no rung templates, and a box with a verdict", module(fstest.MapFS{
			manifestPath:                  file(manifest[:strings.Index(manifest, "rung_templates:")]),
			"tasks/camp-days-1.yaml":      nil,
			"tasks/gaps-ladder-1.yaml":    nil,
			"lessons/days-and-dates.yaml": nil,
			"tasks/tables-and-division-1.yaml": file(strings.NewReplacer("skills:\n  primary: [times-tables]", "skills:\n  primary: [counting-gaps]",
				"band: tables", "band: grades-3-4").Replace(string(examplePart("task", "workbook-page").Data))),
		}), []string{
			"manifest.yaml # manifest.rung-templates",
			"tasks/tables-and-division-1.yaml #/items/0/line/1/check/trap_values/0/trap task.known-traps",
			"tasks/tables-and-division-1.yaml #/items/1/line/0/check/trap_values/0/trap task.known-traps",
			"tasks/tables-and-division-1.yaml #/items/2/line/1/check/trap_values/0/trap task.known-traps",
			"tasks/tables-and-division-1.yaml #/items/3/line/1/check/trap_values/0/trap task.known-traps",
			"tasks/tables-and-division-1.yaml #/items/4/line/1/check/trap_values/0/trap task.known-traps",
			"tasks/tables-and-division-1.yaml #/items/4/line/3/check/trap_values/0/trap task.known-traps",
		}},
		{"traps of skills that the skill graph does not have", module(fstest.MapFS{
			"traps.yaml": file(strings.Replace(string(base["traps.yaml"].Data), "skills: [counting-gaps]", "skills: [time-calendar, gone]", 1)),
		}), []string{"traps.yaml #/traps/off-by-one/skills/1 traps.known-skills"}},
		{"traps of skills with no skill graph", smallest(fstest.MapFS{"traps.yaml": base["traps.yaml"]}),
			[]string{"traps.yaml #/traps/off-by-one/skills/0 traps.known-skills"}},
		{"traps and techniques that the catalogs do not have", module(fstest.MapFS{
			"tasks/gaps-ladder-1.yaml": file(strings.NewReplacer("trap: off-by-one", "trap: gone", "traps: [ignored-condition]", "traps: [lost]",
				"techniques: [draw, small-numbers]", "techniques: [draw, guess]").Replace(string(base["tasks/gaps-ladder-1.yaml"].Data))),
		}), []string{
			"tasks/gaps-ladder-1.yaml #/items/0/answer_slots/0/check/trap_values/0/trap task.known-traps",
			"tasks/gaps-ladder-1.yaml #/items/0/answer_slots/0/check/trap_values/1/trap task.known-traps",
			"tasks/gaps-ladder-1.yaml #/techniques/1 task.known-techniques",
			"tasks/gaps-ladder-1.yaml #/traps/0 task.known-traps",
		}},
		{"traps and techniques with no catalogs", module(fstest.MapFS{"traps.yaml": nil, "techniques.yaml": nil}), []string{
			"tasks/gaps-ladder-1.yaml #/items/0/answer_slots/0/check/trap_values/0/trap task.known-traps",
			"tasks/gaps-ladder-1.yaml #/items/0/answer_slots/0/check/trap_values/1/trap task.known-traps",
			"tasks/gaps-ladder-1.yaml #/techniques/0 task.known-techniques",
			"tasks/gaps-ladder-1.yaml #/techniques/1 task.known-techniques",
			"tasks/gaps-ladder-1.yaml #/traps/0 task.known-traps",
		}},
		{"the rules between parts wait for every file of a part to pass", module(fstest.MapFS{
			"traps.yaml":   nil,
			"tasks/a.yaml": file("id: a\n"),
		}), []string{"tasks/a.yaml # schema required"}},
		{"the rules within a part do not wait", module(fstest.MapFS{
			"skill-graph.yaml": file("skills:\n  a:\n    title: A\n    prerequisites: [a]\n    assessability: verdict\n"),
			"tasks/a.yaml":     file("id: a\n"),
		}), []string{"skill-graph.yaml #/skills/a/prerequisites skill-graph.acyclic", "tasks/a.yaml # schema required"}},
	} {
		t.Run(c.name, func(t *testing.T) {
			report, err := Check(c.fsys)
			if err != nil {
				t.Fatal(err)
			}
			if got := shortReport(report); !slices.Equal(got, c.refused) {
				t.Errorf("refused by\n%s\nwant\n%s\nreport:\n%s", strings.Join(got, "\n"), strings.Join(c.refused, "\n"), report)
			}
		})
	}
}

// CheckFile runs the rules within a part on its file, and none between parts.
func TestPartRulesOfAFile(t *testing.T) {
	for _, c := range []struct {
		part, data string
		refused    []string
	}{
		{"skill-graph", brokenGraph, []string{
			"#/retired_skills/d skill-graph.retired-ids", "#/retired_skills/e/replaced_by/1 skill-graph.known-skills",
			"#/skills/a/prerequisites skill-graph.acyclic", "#/skills/a/prerequisites/1 skill-graph.known-skills",
			"#/skills/a/prerequisites/2 skill-graph.own-anchors", "#/skills/b/prerequisites skill-graph.acyclic",
			"#/skills/c/prerequisites skill-graph.acyclic",
		}},
		{"manifest", string(bands("bands:\n  - id: a\n    title: A\n    grades: [3, 4]\n  - id: a\n    title: B\n    grades: [5]\n").Data),
			[]string{"#/bands/1/grades/0 manifest.bands", "#/bands/1/id manifest.bands"}},
		{"manifest", manifest[:strings.Index(manifest, "rung_templates:")] + "dependencies: [core.linear-equations]\n", nil},
		{"task", strings.Replace(string(base["tasks/gaps-ladder-1.yaml"].Data), "traps: [ignored-condition]", "traps: [lost]", 1), nil},
		{"traps", string(base["traps.yaml"].Data), nil},
	} {
		report, err := CheckFile(c.part, []byte(c.data))
		if err != nil {
			t.Fatal(err)
		}
		if got := shortReport(report); !slices.Equal(got, c.refused) {
			t.Errorf("%s: refused by %q, want %q", c.part, got, c.refused)
		}
	}
}

// Every breach of a rule of the parts says what breaks the rule.
func TestPartRuleMessages(t *testing.T) {
	report, err := CheckFile("skill-graph", []byte(brokenGraph))
	if err != nil {
		t.Fatal(err)
	}
	want := []string{
		"#/retired_skills/d: skill-graph.retired-ids: d is retired and the ID of a skill too: a published ID never names another skill",
		"#/retired_skills/e/replaced_by/1: skill-graph.known-skills: the replacement lost is no skill of the module",
		"#/skills/a/prerequisites: skill-graph.acyclic: the skill names itself among its prerequisites",
		"#/skills/a/prerequisites/1: skill-graph.known-skills: the prerequisite gone is no skill of the module",
		"#/skills/a/prerequisites/2: skill-graph.own-anchors: the prerequisite core.division-with-remainder is an anchor skill that b of the module teaches: name the skill",
		"#/skills/b/prerequisites: skill-graph.acyclic: the skill is, through its prerequisites, a prerequisite of itself, in a cycle with c",
		"#/skills/c/prerequisites: skill-graph.acyclic: the skill is, through its prerequisites, a prerequisite of itself, in a cycle with b",
	}
	if got := strings.Split(report.String(), "\n"); !slices.Equal(got, want) {
		t.Errorf("the report reads\n%s\nwant\n%s", report, strings.Join(want, "\n"))
	}
	messages := map[string]bool{}
	for _, fsys := range []fstest.MapFS{
		module(fstest.MapFS{manifestPath: bands("bands:\n  - id: a\n    title: A\n    grades: [3, 4, 5]\n  - id: a\n    title: B\n  - id: c\n    title: C\n    grades: [4]\n")}),
		module(fstest.MapFS{manifestPath: file(strings.Replace(manifest, "  grades: [3, 4]\nbands", "bands", 1))}),
		module(fstest.MapFS{manifestPath: bands("bands:\n  - id: a\n    title: A\n    grades: [3]\n")}),
		module(fstest.MapFS{
			"skill-graph.yaml": file(strings.Replace(string(base["skill-graph.yaml"].Data), "prerequisites: [counting-gaps]",
				"prerequisites: [counting-gaps, core.multiplication-table]", 1)),
			manifestPath: file(manifest[:strings.Index(manifest, "rung_templates:")] + "dependencies: [core.linear-equations]\n"),
			"traps.yaml": file(strings.Replace(string(base["traps.yaml"].Data), "skills: [counting-gaps]", "skills: [gone]", 1)),
			"tasks/gaps-ladder-1.yaml": file(strings.NewReplacer("traps: [ignored-condition]", "traps: [lost]",
				"techniques: [draw, small-numbers]", "techniques: [guess]").Replace(string(base["tasks/gaps-ladder-1.yaml"].Data))),
		}),
		module(fstest.MapFS{"traps.yaml": nil, "techniques.yaml": nil}),
		smallest(fstest.MapFS{"traps.yaml": base["traps.yaml"]}),
	} {
		report, err := Check(fsys)
		if err != nil {
			t.Fatal(err)
		}
		for _, b := range report.Breaches {
			messages[b.Rule+": "+b.Message] = true
		}
	}
	for _, want := range []string{
		"manifest.bands: the band ID a is that of the band at #/bands/0 too: the IDs of bands differ",
		"manifest.bands: the band names no grades, and the module names grades in audience.grades: every band names them",
		"manifest.bands: grade 5 is no grade of audience.grades",
		"manifest.bands: grade 4 is in the bands a, c: every grade of the audience is in exactly one",
		"manifest.bands: the band names grades, and the module names none in audience.grades",
		"manifest.bands: grade 4 is in no band: every grade of the audience is in exactly one",
		"manifest.dependencies: no prerequisite of a skill of the module names core.linear-equations",
		"manifest.dependencies: the dependencies leave out core.multiplication-table, which time-calendar needs",
		"manifest.rung-templates: the manifest has no rung_templates, and tasks/camp-days-1.yaml has an item with a verdict",
		"traps.known-skills: the trap names the skill gone, which the skill graph does not have",
		"traps.known-skills: the trap names the skill counting-gaps, and the module has no skill graph",
		"task.known-traps: the task names the trap lost, which the trap catalog does not have",
		"task.known-traps: the task names the trap off-by-one, and the module has no trap catalog",
		"task.known-techniques: the task names the technique guess, which the technique catalog does not have",
		"task.known-techniques: the task names the technique draw, and the module has no technique catalog",
	} {
		if !messages[want] {
			t.Errorf("no breach says %q; they say\n%s", want, strings.Join(slices.Sorted(maps.Keys(messages)), "\n"))
		}
	}
}

// An item with a verdict is one whose answer slots or boxes a kind with a verdict checks.
func TestHasVerdict(t *testing.T) {
	for name, want := range map[string]bool{
		"essay": false, "essay-about-yourself": false, "workbook-page": true, "zones-and-picture": true, "best-first-sentence": true,
	} {
		value, err := yaml.Read(examplePart("task", name).Data)
		if err != nil {
			t.Fatal(err)
		}
		if got := hasVerdict(typed[formatv0.Task](value)); got != want {
			t.Errorf("%s: %t, want %t", name, got, want)
		}
	}
}

func graph(prerequisites map[string][]string) formatv0.SkillGraph {
	g := formatv0.SkillGraph{Skills: map[formatv0.SkillID]formatv0.Skill{}}
	for id, needs := range prerequisites {
		g.Skills[formatv0.SkillID(id)] = formatv0.Skill{Title: formatv0.Text(id), Prerequisites: needs}
	}
	return g
}

func TestTiers(t *testing.T) {
	for _, c := range []struct {
		name          string
		prerequisites map[string][]string
		tiers         formatv0.SkillTiers
		err           string
	}{
		{"no skills", map[string][]string{}, formatv0.SkillTiers{}, ""},
		{"an anchor skill counts for none", map[string][]string{"a": {"core.multiplication-table"}, "b": {"a", "core.linear-equations"}},
			formatv0.SkillTiers{"a": 1, "b": 2}, ""},
		{"the highest prerequisite", map[string][]string{"a": nil, "b": {"a"}, "c": {"b"}, "d": {"a", "c"}},
			formatv0.SkillTiers{"a": 1, "b": 2, "c": 3, "d": 4}, ""},
		{"a prerequisite that names no skill", map[string][]string{"a": {"gone"}}, nil, "the skill a needs gone, which is no skill of the module"},
		{"a cycle", map[string][]string{"a": nil, "b": {"c"}, "c": {"b"}}, nil, "the prerequisites of b run in a cycle: its tier is no number"},
		{"a skill that needs itself", map[string][]string{"a": {"a"}}, nil, "the prerequisites of a run in a cycle: its tier is no number"},
	} {
		got, err := Tiers(graph(c.prerequisites))
		if c.err == "" && (err != nil || !maps.Equal(got, c.tiers)) || c.err != "" && (err == nil || err.Error() != c.err || got != nil) {
			t.Errorf("%s: %v, %v; want %v, %q", c.name, got, err, c.tiers, c.err)
		}
	}
}

// The cycles of a graph are its strongly connected components of more than one skill, and the
// skills that need themselves; a skill that needs one on a cycle is on none.
func TestCycles(t *testing.T) {
	got := cycles(edgesOf(graph(map[string][]string{
		"a": {"a"}, "b": {"c"}, "c": {"d", "core.linear-equations"}, "d": {"b", "gone"}, "e": {"b"}, "f": {"g"}, "g": {"f", "e"},
	})))
	want := map[formatv0.SkillID][]string{"a": {}, "b": {"c", "d"}, "c": {"b", "d"}, "d": {"b", "c"}, "f": {"g"}, "g": {"f"}}
	if !maps.EqualFunc(got, want, slices.Equal) {
		t.Errorf("%v, want %v", got, want)
	}
}

// A value that its schema accepts and its type does not read is a fault of module-kit.
func TestTypedFault(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("a value that its type does not read does not panic")
		}
	}()
	typed[formatv0.Traps](map[string]any{"colour": "red"})
}

// A skill that names itself and is on a larger cycle too gets one breach that says both.
func TestSelfAndCycle(t *testing.T) {
	report, err := CheckFile("skill-graph", []byte("skills:\n  a:\n    title: A\n    prerequisites: [a, b]\n    assessability: verdict\n"+
		"  b:\n    title: B\n    prerequisites: [a]\n    assessability: verdict\n"))
	if err != nil {
		t.Fatal(err)
	}
	want := []string{
		"#/skills/a/prerequisites: skill-graph.acyclic: the skill names itself among its prerequisites; " +
			"the skill is, through its prerequisites, a prerequisite of itself, in a cycle with b",
		"#/skills/b/prerequisites: skill-graph.acyclic: the skill is, through its prerequisites, a prerequisite of itself, in a cycle with a",
	}
	if got := strings.Split(report.String(), "\n"); !slices.Equal(got, want) {
		t.Errorf("the report reads\n%s\nwant\n%s", report, strings.Join(want, "\n"))
	}
}
