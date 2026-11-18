package module

import (
	"fmt"
	"maps"
	"slices"
	"strconv"
	"strings"

	"github.com/altnovatrail/spec/types/formatv0"

	"github.com/altnovatrail/module-kit/internal/schema"
)

// The IDs of the rules of the manifest, the skill graph and the catalogs, named by the part a breach of
// each is reported in (R25).
const (
	RuleManifestDependencies  = "manifest.dependencies"
	RuleManifestBands         = "manifest.bands"
	RuleManifestRungTemplates = "manifest.rung-templates"
	RuleSkillGraphAcyclic     = "skill-graph.acyclic"
	RuleSkillGraphKnownSkills = "skill-graph.known-skills"
	RuleSkillGraphRetiredIDs  = "skill-graph.retired-ids"
	RuleSkillGraphOwnAnchors  = "skill-graph.own-anchors"
	RuleTrapsKnownSkills      = "traps.known-skills"
	RuleTaskKnownTraps        = "task.known-traps"
	RuleTaskKnownTechniques   = "task.known-techniques"
)

// partRules lists the rules of the parts that the check holds, for the tests that hold each to the
// text of the format and each breach to a rule the check holds.
var partRules = []string{
	RuleManifestDependencies, RuleManifestBands, RuleManifestRungTemplates, RuleSkillGraphAcyclic, RuleSkillGraphKnownSkills,
	RuleSkillGraphRetiredIDs, RuleSkillGraphOwnAnchors, RuleTrapsKnownSkills, RuleTaskKnownTraps, RuleTaskKnownTechniques,
	RuleTaskKnownSkills, RuleTaskSkillRoles, RuleTaskOwnAnchors, RuleTaskKnownBand, RuleTaskKnownPictures, RuleTaskAgeCategories,
	RuleTaskZones, RuleTaskAssessability, RuleLessonKnownSkills, RuleLessonKnownTasks, RuleLessonTaskSkills,
	RuleLessonKnownPictures, RuleLessonAreas, RuleLessonAgeCategories, RuleLessonOrder, RuleLessonExampleFills, RuleLessonWorkbookPage, RuleTaskWorkbookPage, RuleTaskChoiceKey, RuleTaskChoiceOptions,
	RuleTaskFractionKey, RuleTaskTrapValues, RuleTaskPredicateExpression, RuleTaskNoLeaks, RuleLessonNoLeaks,
	RuleTaskTemplateNames, RuleTaskTemplateExpressions, RuleTaskPlaceholders, RuleTaskTemplateInstances, RuleTaskTemplateSolver,
	RuleTaskInstances,
}

// typed reads the value of a file of a part that passed its schema into the type of formatv0 that
// holds the part, strictly (R72 spec). The schema accepted the value, so the type reads it, and a
// value that it does not is a fault of module-kit.
func typed[T any](value any) T {
	t, err := schema.Decode[T](value)
	if err != nil {
		panic(fmt.Sprintf("module: a value that its schema accepts: %v", err))
	}
	return t
}

func at(tokens ...string) string {
	return schema.Pointer(tokens)
}

// A module's parts that passed their schemas, read into their types once for every rule (R25): a part
// the module does not have, or whose file did not pass, is nil; the tasks, the lessons and the
// expansions of templates stand by their paths, the expansions set by expandAll once every part is
// added; media holds the paths of the files in media/. file is true for the file of CheckFile, which
// has no path.
type parts struct {
	file       bool
	manifest   *formatv0.Manifest
	skillGraph *formatv0.SkillGraph
	traps      *formatv0.Traps
	techniques *formatv0.Techniques
	steps      *formatv0.SolvingSteps
	tasks      map[string]formatv0.Task
	templates  map[string]*expansion
	lessons    map[string]formatv0.Lesson
	media      map[string]bool
	// The skills of the skill graph that teach each anchor skill, none when there is no graph.
	teachers map[formatv0.SharedTaxonomy][]string
}

func newParts(file bool) parts {
	return parts{file: file, tasks: map[string]formatv0.Task{}, templates: map[string]*expansion{}, lessons: map[string]formatv0.Lesson{},
		media: map[string]bool{}}
}

// add puts the value of a file that passed its schema among the parts; a part that no rule reads is
// dropped.
func (p *parts) add(path, part string, value any) {
	switch part {
	case partManifest:
		p.manifest = new(typed[formatv0.Manifest](value))
	case partSkillGraph:
		p.skillGraph = new(typed[formatv0.SkillGraph](value))
		p.teachers = teachersOf(*p.skillGraph)
	case partTraps:
		p.traps = new(typed[formatv0.Traps](value))
	case partTechniques:
		p.techniques = new(typed[formatv0.Techniques](value))
	case partSteps:
		p.steps = new(typed[formatv0.SolvingSteps](value))
	case partTask:
		p.tasks[path] = typed[formatv0.Task](value)
	case partLesson:
		p.lessons[path] = typed[formatv0.Lesson](value)
	}
}

// pathOf gives the path of a breach in the file at path, or none for the file of CheckFile (R22).
func (p parts) pathOf(path string) string {
	if p.file {
		return ""
	}
	return path
}

// withinPart checks the rules within one part, in Check and in CheckFile alike, since an example of
// a part is checked so (R25). A template is checked through its instances by the rules that read the
// strings of a task, where its placeholders stand, and as written by those that read none of them,
// which every instance breaks alike (R32).
func (c *checker) withinPart(p parts) {
	if p.manifest != nil {
		c.bands(p.pathOf(manifestPath), *p.manifest)
	}
	if p.skillGraph != nil {
		c.skillGraph(p.pathOf(skillGraphPath), *p.skillGraph, p.teachers)
	}
	for path, task := range p.tasks {
		c.skillRoles(path, task)
		c.zones(path, task)
		if e, ok := p.templates[path]; ok {
			c.templateRules(path, e)
			continue
		}
		c.checkRules(path, task)
		c.workbookPage(path, task)
		c.taskLeaks(path, task, ownShown(task, false))
	}
	for path, lesson := range p.lessons {
		c.areas(path, lesson)
		c.exampleFills(path, lesson)
		c.exampleOnPage(path, lesson)
	}
}

// betweenParts checks the rules between parts. They run only when every file of a part passed, as
// format-check of spec has it, so that a file that its schema refuses breaks that and nothing more
// (R23); the manifest is then among them.
func (c *checker) betweenParts(p parts) {
	c.dependencies(p)
	c.rungTemplates(p)
	c.trapSkills(p)
	for path, task := range p.tasks {
		c.knownTraps(path, task, p.traps)
		c.knownTechniques(path, task, p.techniques)
		c.taskRules(path, task, p)
		if e, ok := p.templates[path]; ok {
			c.onInstances(e, func(sub *checker, i instance) {
				if len(i.invalid) == 0 {
					sub.taskPictures(path, i.task, p.media)
					sub.taskLeaks(path, i.task, partsShown(i.task, p))
				}
			})
			continue
		}
		c.taskPictures(path, task, p.media)
		c.taskLeaks(path, task, partsShown(task, p))
	}
	for path, lesson := range p.lessons {
		c.lessonRules(path, lesson, p)
		c.lessonLeaks(path, lesson, p)
	}
	c.order(p.lessons, p.skillGraph)
}

// bands checks manifest.bands: when the manifest has bands, their IDs differ; in a module that names
// grades, every band names grades, every grade of the audience is in exactly one band, and no band
// names a grade the audience does not; in a module that names none, no band names grades.
func (c *checker) bands(path string, m formatv0.Manifest) {
	first := map[formatv0.BandID]int{}
	for i, band := range m.Bands {
		if j, ok := first[band.ID]; ok {
			c.add(RuleManifestBands, path, 0, 0, at("bands", strconv.Itoa(i), "id"),
				fmt.Sprintf("the band ID %s is that of the band at #/bands/%d too: the IDs of bands differ", band.ID, j))
			continue
		}
		first[band.ID] = i
	}
	grades := m.Audience.Grades
	if len(m.Bands) == 0 {
		return
	}
	for i, band := range m.Bands {
		switch {
		case len(grades) == 0 && len(band.Grades) > 0:
			c.add(RuleManifestBands, path, 0, 0, at("bands", strconv.Itoa(i), "grades"),
				"the band names grades, and the module names none in audience.grades")
		case len(grades) > 0 && len(band.Grades) == 0:
			c.add(RuleManifestBands, path, 0, 0, at("bands", strconv.Itoa(i)),
				"the band names no grades, and the module names grades in audience.grades: every band names them")
		}
		for k, grade := range band.Grades {
			if len(grades) > 0 && !slices.Contains(grades, grade) {
				c.add(RuleManifestBands, path, 0, 0, at("bands", strconv.Itoa(i), "grades", strconv.Itoa(k)),
					fmt.Sprintf("grade %d is no grade of audience.grades", grade))
			}
		}
	}
	for j, grade := range grades {
		var holding []string
		for _, band := range m.Bands {
			if slices.Contains(band.Grades, grade) {
				holding = append(holding, string(band.ID))
			}
		}
		switch len(holding) {
		case 1:
		case 0:
			c.add(RuleManifestBands, path, 0, 0, at("audience", "grades", strconv.Itoa(j)),
				fmt.Sprintf("grade %d is in no band: every grade of the audience is in exactly one", grade))
		default:
			c.add(RuleManifestBands, path, 0, 0, at("audience", "grades", strconv.Itoa(j)),
				fmt.Sprintf("grade %d is in the bands %s: every grade of the audience is in exactly one", grade, strings.Join(holding, ", ")))
		}
	}
}

// dependencies checks manifest.dependencies: the manifest's dependencies are exactly the anchor
// skills that the prerequisites of the module's skills name.
func (c *checker) dependencies(p parts) {
	needed := map[string][]string{} // the skills that need each anchor skill
	if p.skillGraph != nil {
		for _, id := range slices.Sorted(maps.Keys(p.skillGraph.Skills)) {
			for _, prerequisite := range p.skillGraph.Skills[id].Prerequisites {
				if isAnchor(prerequisite) {
					needed[prerequisite] = append(needed[prerequisite], string(id))
				}
			}
		}
	}
	listed := map[string]bool{}
	for i, anchor := range p.manifest.Dependencies {
		listed[string(anchor)] = true
		if _, ok := needed[string(anchor)]; !ok {
			c.add(RuleManifestDependencies, manifestPath, 0, 0, at("dependencies", strconv.Itoa(i)),
				fmt.Sprintf("no prerequisite of a skill of the module names %s", anchor))
		}
	}
	place := at("dependencies")
	if len(p.manifest.Dependencies) == 0 {
		place = at()
	}
	for _, anchor := range slices.Sorted(maps.Keys(needed)) {
		if !listed[anchor] {
			c.add(RuleManifestDependencies, manifestPath, 0, 0, place,
				fmt.Sprintf("the dependencies leave out %s, which %s needs", anchor, strings.Join(needed[anchor], ", ")))
		}
	}
}

// The check kinds that give a verdict: an item whose answer slots they check is an item with a
// verdict (R25).
var verdictKinds = []formatv0.CheckKind{
	formatv0.CheckKindNumber, formatv0.CheckKindFraction, formatv0.CheckKindChoice, formatv0.CheckKindPredicate,
}

// checks gives the check of each answer slot of an item, with its place in the task: an answer slot
// of a handwriting sheet, and a box of a workbook's line.
func checks(item formatv0.Item, i int) (found []formatv0.Check, places [][]string) {
	slots, slotPlaces := slotsOf(item, i)
	for n, slot := range slots {
		found = append(found, slot.Check)
		places = append(places, slices.Concat(slotPlaces[n], []string{"check"}))
	}
	return found, places
}

// hasVerdict says whether a task has an item with a verdict: one with an answer slot that a kind
// with a verdict checks, as format-check of spec reads it (R25).
func hasVerdict(task formatv0.Task) bool {
	for i, item := range task.Items {
		found, _ := checks(item, i)
		for _, check := range found {
			if slices.Contains(verdictKinds, check.Kind) {
				return true
			}
		}
	}
	return false
}

// rungTemplates checks manifest.rung-templates: a module with an item with a verdict has rung
// templates.
func (c *checker) rungTemplates(p parts) {
	if len(p.manifest.RungTemplates) > 0 {
		return
	}
	for _, path := range slices.Sorted(maps.Keys(p.tasks)) {
		if hasVerdict(p.tasks[path]) {
			c.add(RuleManifestRungTemplates, manifestPath, 0, 0, at(),
				fmt.Sprintf("the manifest has no rung_templates, and %s has an item with a verdict", path))
			return
		}
	}
}

// trapSkills checks traps.known-skills: every skill of a trap is a skill of the module's skill graph.
func (c *checker) trapSkills(p parts) {
	if p.traps == nil {
		return
	}
	for id, trap := range p.traps.Traps {
		for i, skill := range trap.Skills {
			if message, unknown := unknownSkill(p.skillGraph, "the trap names", skill); unknown {
				c.add(RuleTrapsKnownSkills, trapsPath, 0, 0, at("traps", string(id), "skills", strconv.Itoa(i)), message)
			}
		}
	}
}

// knownTraps checks task.known-traps: every trap that a task names, in its traps and in the trap
// values of its checks, is a trap of the module's trap catalog.
func (c *checker) knownTraps(path string, task formatv0.Task, catalog *formatv0.Traps) {
	check := func(trap formatv0.TrapID, place ...string) {
		switch {
		case catalog == nil:
			c.add(RuleTaskKnownTraps, path, 0, 0, at(place...),
				fmt.Sprintf("the task names the trap %s, and the module has no trap catalog", trap))
		case !hasKey(catalog.Traps, trap):
			c.add(RuleTaskKnownTraps, path, 0, 0, at(place...),
				fmt.Sprintf("the task names the trap %s, which the trap catalog does not have", trap))
		}
	}
	for i, trap := range task.Traps {
		check(trap, "traps", strconv.Itoa(i))
	}
	for i, item := range task.Items {
		found, places := checks(item, i)
		for n, slot := range found {
			for k, value := range slot.TrapValues {
				check(value.Trap, slices.Concat(places[n], []string{"trap_values", strconv.Itoa(k), "trap"})...)
			}
		}
	}
}

// knownTechniques checks task.known-techniques: every technique that a task names is a technique of
// the module's technique catalog.
func (c *checker) knownTechniques(path string, task formatv0.Task, catalog *formatv0.Techniques) {
	for i, technique := range task.Techniques {
		switch {
		case catalog == nil:
			c.add(RuleTaskKnownTechniques, path, 0, 0, at("techniques", strconv.Itoa(i)),
				fmt.Sprintf("the task names the technique %s, and the module has no technique catalog", technique))
		case !hasKey(catalog.Techniques, technique):
			c.add(RuleTaskKnownTechniques, path, 0, 0, at("techniques", strconv.Itoa(i)),
				fmt.Sprintf("the task names the technique %s, which the technique catalog does not have", technique))
		}
	}
}

func hasKey[K comparable, V any](m map[K]V, key K) bool {
	_, ok := m[key]
	return ok
}
