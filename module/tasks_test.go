package module

import (
	"context"
	"maps"
	"slices"
	"strings"
	"testing"
	"testing/fstest"
)

// replaced gives text with each old string of pairs, which it must hold, replaced by the new one
// after it.
func replaced(text string, pairs ...string) string {
	for i := 0; i < len(pairs); i += 2 {
		if !strings.Contains(text, pairs[i]) {
			panic("the text holds no " + pairs[i])
		}
		text = strings.Replace(text, pairs[i], pairs[i+1], 1)
	}
	return text
}

// edit gives a copy of a file of base with the replacements of pairs.
func edit(at string, pairs ...string) *fstest.MapFile {
	return file(replaced(string(base[at].Data), pairs...))
}

// Paths of base and the lines of its files that the cases replace.
const (
	campDays      = "tasks/camp-days-1.yaml"
	gapsLadder    = "tasks/gaps-ladder-1.yaml"
	daysLesson    = "lessons/days-and-dates.yaml"
	gapsLesson    = "lessons/gaps-and-posts.yaml"
	campSkills    = "skills:\n  primary: [time-calendar]\n"
	campItem      = "items:\n  - answer_slots:\n"
	campSolution  = "leak_phrases:"
	gapsGraph     = "  counting-gaps:\n    title: Gaps and boundaries\n    assessability: verdict\n"
	timeGraph     = "  time-calendar:\n    title: Calendar and age\n    prerequisites: [counting-gaps]\n    assessability: verdict\n"
	fence         = "picture: pictures/fence-posts-and-gaps.png"
	pathInGaps    = "picture: pictures/path-in-gaps.png"
	lampArea      = "area: {left: 100, top: 150, width: 800, height: 200}"
	lampStatement = "How many lamps are there?\n"
)

// A picture of a statement whose file the module does not have.
const missingPicture = "  - picture: pictures/missing.png\n    alt: Nothing.\n    width: 10\n    height: 10\n"

// The essay of the examples of a task, on a skill of base and in its band.
var essay = file(replaced(string(examplePart("task", "essay").Data),
	"primary: [paragraphs, endings]\n  supporting: [writing-a-first-sentence]", "primary: [counting-gaps]", "band: whole-essays", "band: grades-3-4"))

// The cases hold each rule of a task and of a lesson at each place it is broken, on a whole module.
var taskCases = []struct {
	name    string
	fsys    fstest.MapFS
	refused []string
}{
	{"skills that the skill graph does not have, and a skill both primary and supporting", module(fstest.MapFS{
		campDays: edit(campDays, campSkills, "skills:\n  primary: [time-calendar, gone]\n  supporting: [time-calendar, lost, core.multiplication-table]\n"),
	}), []string{
		campDays + " #/skills/primary/1 task.known-skills", campDays + " #/skills/supporting/0 task.skill-roles",
		campDays + " #/skills/supporting/1 task.known-skills",
	}},
	{"the skills of an item take the place of the task's", module(fstest.MapFS{
		campDays: edit(campDays, campSkills, "skills:\n  primary: [time-calendar]\n  supporting: [time-calendar]\n",
			campItem, "items:\n  - skills:\n      primary: [time-calendar, gone]\n      supporting: [time-calendar]\n    answer_slots:\n"),
	}), []string{campDays + " #/items/0/skills/primary/1 task.known-skills", campDays + " #/items/0/skills/supporting/0 task.skill-roles"}},
	{"a supporting anchor skill that the module teaches", module(fstest.MapFS{
		"skill-graph.yaml": edit("skill-graph.yaml", gapsGraph, "  counting-gaps:\n    title: Gaps and boundaries\n    teaches: [core.multiplication-table]\n    assessability: verdict\n"),
		campDays:           edit(campDays, campSkills, campSkills+"  supporting: [core.multiplication-table, core.linear-equations]\n"),
	}), []string{campDays + " #/skills/supporting/0 task.own-anchors"}},
	{"a band that the manifest does not have", module(fstest.MapFS{campDays: edit(campDays, "band: grades-3-4", "band: grades-5-6")}),
		[]string{campDays + " #/band task.known-band"}},
	{"no bands", module(fstest.MapFS{manifestPath: bands("")}),
		[]string{campDays + " #/band task.known-band", gapsLadder + " #/band task.known-band"}},
	{"pictures that the module does not have", module(fstest.MapFS{
		campDays: edit(campDays, "statement:\n", "statement:\n"+missingPicture, campSolution, missingPicture+campSolution),
	}), []string{campDays + " #/solution/1/picture task.known-pictures", campDays + " #/statement/0/picture task.known-pictures"}},
	{"an age category outside the audience, and a lesson that no learner can get", module(fstest.MapFS{
		campDays: edit(campDays, "difficulty: 3\n", "difficulty: 3\nage_categories: [adult]\n"),
	}), []string{daysLesson + " # lesson.age-categories", campDays + " #/age_categories/0 task.age-categories"}},
	{"zones off the sheet and overlapping", module(fstest.MapFS{
		"tasks/clock-walk-1.yaml": file(replaced(string(examplePart("task", "zones-and-picture").Data),
			"{left: 500, top: 1000, width: 400, height: 200}", "{left: 450, top: 1100, width: 400, height: 400}",
			"skills:\n  primary: [time-clocks]", "skills:\n  primary: [time-calendar]", "band: grades-1-2", "band: grades-3-4",
			"pictures/clock-quarter-past-three.png", "pictures/fence-posts-and-gaps.png", "trap: answered-other-question", "trap: off-by-one")),
	}), []string{"tasks/clock-walk-1.yaml #/items/1/answer_slots/0/zone task.zones", "tasks/clock-walk-1.yaml #/items/1/answer_slots/0/zone task.zones"}},
	{"a verdict on a skill of rubrics", module(fstest.MapFS{"skill-graph.yaml": edit("skill-graph.yaml", timeGraph, strings.Replace(timeGraph, "verdict", "rubric", 1))}),
		[]string{campDays + " #/skills/primary/0 task.assessability"}},
	{"a skill not assessed as primary, named once for every item", module(fstest.MapFS{"skill-graph.yaml": edit("skill-graph.yaml", gapsGraph, strings.Replace(gapsGraph, "verdict", "none", 1))}),
		[]string{gapsLadder + " #/skills/primary/0 task.assessability"}},
	{"a rubric on a skill of verdicts", module(fstest.MapFS{"tasks/undiscovered-island-1.yaml": essay}),
		[]string{"tasks/undiscovered-island-1.yaml #/skills/primary/0 task.assessability"}},
	{"a lesson of skills that the skill graph does not have, of a task it does not train and of one the module does not have", module(fstest.MapFS{
		daysLesson: edit(daysLesson, "skills: [time-calendar]", "skills: [counting-gaps, gone]", "  - task: camp-days-1\n", "  - task: camp-days-1\n  - task: missing\n"),
	}), []string{daysLesson + " #/segments/1/task lesson.task-skills", daysLesson + " #/segments/2/task lesson.known-tasks", daysLesson + " #/skills/1 lesson.known-skills"}},
	{"pictures of a lesson that the module does not have, and an area off the sheet", module(fstest.MapFS{
		gapsLesson: edit(gapsLesson, fence, "picture: pictures/missing.png", pathInGaps, "picture: pictures/missing.png",
			lampStatement, lampStatement+"        - picture: pictures/missing.png\n          alt: Nothing.\n          width: 10\n          height: 10\n",
			lampArea, "area: {left: 100, top: 150, width: 950, height: 200}"),
	}), []string{
		gapsLesson + " #/segments/0/explanation/1/picture lesson.known-pictures", gapsLesson + " #/segments/1/worked_example/demo/1/area lesson.areas",
		gapsLesson + " #/segments/1/worked_example/demo/1/picture lesson.known-pictures",
		gapsLesson + " #/segments/1/worked_example/statement/1/picture lesson.known-pictures",
	}},
	{"a lesson for an age category outside the audience", module(fstest.MapFS{
		daysLesson: edit(daysLesson, "segments:\n", "age_categories: [adult]\nsegments:\n"),
	}), []string{daysLesson + " # lesson.age-categories", daysLesson + " #/age_categories/0 lesson.age-categories"}},
	{"two lessons in a cycle", module(fstest.MapFS{
		"skill-graph.yaml": file(string(base["skill-graph.yaml"].Data) + "  first:\n    title: First\n    assessability: verdict\n" +
			"  second:\n    title: Second\n    prerequisites: [first]\n    assessability: verdict\n"),
		daysLesson: edit(daysLesson, "skills: [time-calendar]", "skills: [time-calendar, first]"),
		gapsLesson: edit(gapsLesson, "skills: [counting-gaps]", "skills: [counting-gaps, second]"),
	}), []string{daysLesson + " #/skills lesson.order", gapsLesson + " #/skills lesson.order"}},
	{"lessons in order, a lesson needing one skill by two routes", module(fstest.MapFS{
		"skill-graph.yaml": file(string(base["skill-graph.yaml"].Data) + "  extra:\n    title: Extra\n    prerequisites: [counting-gaps]\n    assessability: verdict\n"),
		daysLesson:         edit(daysLesson, "skills: [time-calendar]", "skills: [time-calendar, extra]"),
	}), nil},
	{"a task and a lesson in a module with no skill graph", smallest(fstest.MapFS{
		manifestPath:     file(manifest),
		campDays:         base[campDays],
		"lessons/a.yaml": file("id: a\ntitle: A\nskills: [time-calendar]\nsegments:\n  - explanation:\n      - text: A.\n"),
	}), []string{"lessons/a.yaml #/skills/0 lesson.known-skills", campDays + " #/skills/primary/0 task.known-skills"}},
}

func TestTaskRules(t *testing.T) {
	for _, c := range taskCases {
		t.Run(c.name, func(t *testing.T) {
			report, err := Check(context.Background(), c.fsys)
			if err != nil {
				t.Fatal(err)
			}
			if got := shortReport(report); !slices.Equal(got, c.refused) {
				t.Errorf("refused by\n%s\nwant\n%s\nreport:\n%s", strings.Join(got, "\n"), strings.Join(c.refused, "\n"), report)
			}
		})
	}
}

// CheckFile runs the rules within a task and a lesson on its file: task.skill-roles, task.zones and
// lesson.areas, and no rule between parts. Zones that touch at an edge do not overlap.
func TestTaskRulesOfAFile(t *testing.T) {
	zones := string(examplePart("task", "zones-and-picture").Data)
	for _, c := range []struct {
		part, data string
		refused    []string
	}{
		{"task", zones, nil},
		{"task", replaced(zones, "{left: 500, top: 1000, width: 400, height: 200}", "{left: 450, top: 900, width: 400, height: 200}"),
			[]string{"#/items/1/answer_slots/0/zone task.zones"}},
		{"task", replaced(zones, "{left: 100, top: 1000, width: 400, height: 200}", "{left: 700, top: 0, width: 301, height: 100}"),
			[]string{"#/items/0/answer_slots/0/zone task.zones"}},
		{"task", replaced(string(base[campDays].Data), campSkills, "skills:\n  primary: [time-calendar, gone]\n  supporting: [time-calendar]\n"),
			[]string{"#/skills/supporting/0 task.skill-roles"}},
		{"lesson", replaced(string(base[gapsLesson].Data), lampArea, "area: {left: 100, top: 1300, width: 800, height: 101}"),
			[]string{"#/segments/1/worked_example/demo/1/area lesson.areas"}},
		{"lesson", replaced(string(base[daysLesson].Data), "skills: [time-calendar]", "skills: [gone]"), nil},
	} {
		report, err := CheckFile(c.part, []byte(c.data))
		if err != nil {
			t.Fatal(err)
		}
		if got := shortReport(report); !slices.Equal(got, c.refused) {
			t.Errorf("%s: refused by %q, want %q\n%s", c.part, got, c.refused, report)
		}
	}
}

// Every breach of a rule of a task and of a lesson says what breaks the rule.
func TestTaskRuleMessages(t *testing.T) {
	messages := map[string]bool{}
	for _, c := range taskCases {
		report, err := Check(context.Background(), c.fsys)
		if err != nil {
			t.Fatal(err)
		}
		for _, b := range report.Breaches {
			messages[b.Rule+": "+b.Message] = true
		}
	}
	for _, want := range []string{
		"task.known-skills: the task names the skill gone, which the skill graph does not have",
		"task.known-skills: the task names the skill time-calendar, and the module has no skill graph",
		"task.skill-roles: the skill time-calendar is a primary skill here too: a skill is one or the other",
		"task.own-anchors: the supporting skill core.multiplication-table is an anchor skill that counting-gaps of the module teaches: name the skill",
		"task.known-band: the task is in the band grades-5-6, which the manifest does not have",
		"task.known-band: the task is in the band grades-3-4, and the manifest has no bands",
		"task.known-pictures: the picture pictures/missing.png is no file of the module in media/",
		"task.age-categories: the age category adult is none of the manifest's audience.age_categories",
		"task.zones: the zone goes past the sheet of 1000 × 1400 units",
		"task.zones: the zone overlaps the zone at #/items/0/answer_slots/0/zone",
		"task.assessability: the primary skill time-calendar is of the assessability rubric, and a kind with a verdict checks the item at #/items/0",
		"task.assessability: the primary skill counting-gaps is of the assessability none, and a kind with a verdict checks the item at #/items/0",
		"task.assessability: the primary skill counting-gaps is of the assessability verdict, and a rubric checks the item at #/items/0",
		"lesson.known-skills: the lesson teaches the skill gone, which the skill graph does not have",
		"lesson.known-skills: the lesson teaches the skill time-calendar, and the module has no skill graph",
		"lesson.known-tasks: the task missing is no task of the module: it has no file tasks/missing.yaml",
		"lesson.task-skills: the task camp-days-1 trains no skill of the lesson: no item of it has one among its primary skills",
		"lesson.known-pictures: the picture pictures/missing.png is no file of the module in media/",
		"lesson.areas: the area goes past the sheet of 1000 × 1400 units",
		"lesson.age-categories: the age category adult is none of the manifest's audience.age_categories",
		"lesson.age-categories: no age category suits both the lesson and each task of the module that it names: no learner can get it",
		"lesson.order: the lesson comes, through the lessons that it comes after, after itself, in a cycle with " + gapsLesson,
	} {
		if !messages[want] {
			t.Errorf("no breach says %q; they say\n%s", want, strings.Join(slices.Sorted(maps.Keys(messages)), "\n"))
		}
	}
}

// A module whose only item with a verdict is a box of a workbook has rung templates too, and a cycle
// of the skill graph gives no order of lessons to break: lesson.order waits for the graph.
func TestReviewedCases(t *testing.T) {
	workbook := replaced(string(examplePart("task", "workbook-page").Data), "primary: [times-tables]", "primary: [counting-gaps]",
		"primary: [division-with-remainder]", "primary: [time-calendar]", "band: grade-3", "band: grades-3-4")
	for _, trap := range []string{"times-table-slip", "wrong-operation", "reversed-comparison", "remainder-too-big"} {
		workbook = strings.ReplaceAll(workbook, "trap: "+trap, "trap: off-by-one")
	}
	for _, c := range []struct {
		name    string
		fsys    fstest.MapFS
		refused []string
	}{
		{"boxes with a verdict and no rung templates", module(fstest.MapFS{
			manifestPath: file(manifest[:strings.Index(manifest, "rung_templates:")]),
			campDays:     nil, gapsLadder: nil, daysLesson: nil, gapsLesson: nil,
			"tasks/tables-and-division-1.yaml": file(workbook),
		}), []string{"manifest.yaml # manifest.rung-templates"}},
		{"a cycle of skills under two lessons", module(fstest.MapFS{
			"skill-graph.yaml": edit("skill-graph.yaml", gapsGraph, strings.Replace(gapsGraph, "    assessability", "    prerequisites: [time-calendar]\n    assessability", 1)),
		}), []string{
			"skill-graph.yaml #/skills/counting-gaps/prerequisites skill-graph.acyclic",
			"skill-graph.yaml #/skills/time-calendar/prerequisites skill-graph.acyclic",
		}},
	} {
		report, err := Check(context.Background(), c.fsys)
		if err != nil {
			t.Fatal(err)
		}
		if got := shortReport(report); !slices.Equal(got, c.refused) {
			t.Errorf("%s: refused by\n%s\nwant\n%s\nreport:\n%s", c.name, strings.Join(got, "\n"), strings.Join(c.refused, "\n"), report)
		}
	}
}
