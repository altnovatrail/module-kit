package module

import (
	"fmt"
	"maps"
	"slices"
	"strconv"
	"strings"

	"github.com/altnovatrail/spec/types/formatv0"
)

// The rules of a task and of a lesson that the check holds (format/README.md, sections 14, 15 and
// 21): those that need no check kind, no leak check and no template.
const (
	RuleTaskKnownSkills     = "task.known-skills"
	RuleTaskSkillRoles      = "task.skill-roles"
	RuleTaskOwnAnchors      = "task.own-anchors"
	RuleTaskKnownBand       = "task.known-band"
	RuleTaskKnownPictures   = "task.known-pictures"
	RuleTaskAgeCategories   = "task.age-categories"
	RuleTaskZones           = "task.zones"
	RuleTaskAssessability   = "task.assessability"
	RuleLessonKnownSkills   = "lesson.known-skills"
	RuleLessonKnownTasks    = "lesson.known-tasks"
	RuleLessonTaskSkills    = "lesson.task-skills"
	RuleLessonKnownPictures = "lesson.known-pictures"
	RuleLessonAreas         = "lesson.areas"
	RuleLessonAgeCategories = "lesson.age-categories"
	RuleLessonOrder         = "lesson.order"
)

// The size of a handwriting sheet, in its units (format/README.md, section 14).
const sheetWidth, sheetHeight = 1000, 1400

// The skills that a task names at one place, and the tokens of the place: the task's own, ofTask,
// and an item's.
type skillsAt struct {
	skills formatv0.Skills
	tokens []string
	ofTask bool
}

// The skills of a task and of its item at i, at their places.
func taskSkillsOf(task formatv0.Task) skillsAt {
	return skillsAt{task.Skills, []string{"skills"}, true}
}

func ownSkillsOf(item formatv0.Item, i int) skillsAt {
	return skillsAt{*item.Skills, []string{"items", strconv.Itoa(i), "skills"}, false}
}

// namedSkills gives every place where a task names skills: the task, then each item that names
// skills of its own.
func namedSkills(task formatv0.Task) []skillsAt {
	named := []skillsAt{taskSkillsOf(task)}
	for i, item := range task.Items {
		if item.Skills != nil {
			named = append(named, ownSkillsOf(item, i))
		}
	}
	return named
}

// itemSkills gives the skills of an item, with their place: its own, or else its task's
// (format/README.md, section 14).
func itemSkills(task formatv0.Task, i int) skillsAt {
	if item := task.Items[i]; item.Skills != nil {
		return ownSkillsOf(item, i)
	}
	return taskSkillsOf(task)
}

// place writes the place of an element of a list of skills at a place.
func (s skillsAt) place(list string, k int) string {
	return at(slices.Concat(s.tokens, []string{list, strconv.Itoa(k)})...)
}

// skillRoles checks task.skill-roles: no skill is both a primary and a supporting skill of an item —
// of the task, for an item that names no skills of its own, or of the item itself.
func (c *checker) skillRoles(path string, task formatv0.Task) {
	usesTasks := slices.ContainsFunc(task.Items, func(item formatv0.Item) bool { return item.Skills == nil })
	for _, named := range namedSkills(task) {
		if named.ofTask && !usesTasks {
			continue
		}
		for k, skill := range named.skills.Supporting {
			if slices.Contains(named.skills.Primary, formatv0.SkillID(skill)) {
				c.add(RuleTaskSkillRoles, path, 0, 0, named.place("supporting", k),
					fmt.Sprintf("the skill %s is a primary skill here too: a skill is one or the other", skill))
			}
		}
	}
}

// A zone of a task, and its place.
type zoneAt struct {
	zone  formatv0.SheetRectangle
	place string
}

// offSheet says whether a rectangle of a handwriting sheet goes past its right or bottom edge.
func offSheet(r formatv0.SheetRectangle) bool {
	return r.Left+r.Width > sheetWidth || r.Top+r.Height > sheetHeight
}

// zones checks task.zones: every zone lies on the sheet, and no two zones overlap; zones that touch
// at an edge do not.
func (c *checker) zones(path string, task formatv0.Task) {
	var found []zoneAt
	for i, item := range task.Items {
		for j, slot := range item.AnswerSlots {
			if slot.Zone != nil {
				found = append(found, zoneAt{*slot.Zone, at("items", strconv.Itoa(i), "answer_slots", strconv.Itoa(j), "zone")})
			}
		}
	}
	for n, z := range found {
		if offSheet(z.zone) {
			c.add(RuleTaskZones, path, 0, 0, z.place,
				fmt.Sprintf("the zone goes past the sheet of %d × %d units", sheetWidth, sheetHeight))
		}
		for _, earlier := range found[:n] {
			a, b := earlier.zone, z.zone
			if a.Left < b.Left+b.Width && b.Left < a.Left+a.Width && a.Top < b.Top+b.Height && b.Top < a.Top+a.Height {
				c.add(RuleTaskZones, path, 0, 0, z.place, "the zone overlaps the zone at "+earlier.place)
			}
		}
	}
}

// areas checks lesson.areas: every area of a picture on the sheet of a worked example lies on the
// sheet.
func (c *checker) areas(path string, lesson formatv0.Lesson) {
	for i, segment := range lesson.Segments {
		if segment.WorkedExample == nil {
			continue
		}
		for k, part := range segment.WorkedExample.Demo {
			if part.Area != nil && offSheet(*part.Area) {
				c.add(RuleLessonAreas, path, 0, 0, at("segments", strconv.Itoa(i), "worked_example", "demo", strconv.Itoa(k), "area"),
					fmt.Sprintf("the area goes past the sheet of %d × %d units", sheetWidth, sheetHeight))
			}
		}
	}
}

// taskSkills checks task.known-skills and task.own-anchors: every local ID in the skills of a task
// and of its items is a skill of the module's skill graph, and no supporting skill is an anchor
// skill that a skill of the module teaches.
func (c *checker) taskSkills(path string, task formatv0.Task, p parts) {
	for _, named := range namedSkills(task) {
		for k, skill := range named.skills.Primary {
			if message, unknown := unknownSkill(p.skillGraph, "the task names", skill); unknown {
				c.add(RuleTaskKnownSkills, path, 0, 0, named.place("primary", k), message)
			}
		}
		for k, skill := range named.skills.Supporting {
			message, unknown := unknownSkill(p.skillGraph, "the task names", formatv0.SkillID(skill))
			switch taught := p.teachers[formatv0.SharedTaxonomy(skill)]; {
			case !isAnchor(skill) && unknown:
				c.add(RuleTaskKnownSkills, path, 0, 0, named.place("supporting", k), message)
			case isAnchor(skill) && len(taught) > 0:
				c.add(RuleTaskOwnAnchors, path, 0, 0, named.place("supporting", k),
					fmt.Sprintf("the supporting skill %s is an anchor skill that %s of the module teaches: name the skill", skill, strings.Join(taught, ", ")))
			}
		}
	}
}

// assessability checks task.assessability: each primary skill of an item that is a skill of the
// module's graph is of the assessability verdict when a kind with a verdict checks the item's answer
// slots, and rubric when a rubric does. A primary skill that the task's skills give many items is
// named once.
func (c *checker) assessability(path string, task formatv0.Task, g *formatv0.SkillGraph) {
	if g == nil {
		return
	}
	named := map[string]bool{}
	for i, item := range task.Items {
		found, _ := checks(item, i)
		want := formatv0.SkillAssessabilityVerdict
		if slices.ContainsFunc(found, func(check formatv0.Check) bool { return check.Kind == formatv0.CheckKindRubric }) {
			want = formatv0.SkillAssessabilityRubric
		}
		skills := itemSkills(task, i)
		for k, id := range skills.skills.Primary {
			skill, ok := g.Skills[id]
			if place := skills.place("primary", k); ok && skill.Assessability != want && !named[place] {
				named[place] = true
				c.add(RuleTaskAssessability, path, 0, 0, place,
					fmt.Sprintf("the primary skill %s is of the assessability %s, and a %s checks the item at #/items/%d", id, skill.Assessability, checkerOf(want), i))
			}
		}
	}
}

// checkerOf names what checks an item whose primary skills want an assessability.
func checkerOf(want formatv0.SkillAssessability) string {
	if want == formatv0.SkillAssessabilityRubric {
		return "rubric"
	}
	return "kind with a verdict"
}

// knownBand checks task.known-band: the band of a task is a band of the manifest.
func (c *checker) knownBand(path string, task formatv0.Task, m formatv0.Manifest) {
	if slices.ContainsFunc(m.Bands, func(b formatv0.Band) bool { return b.ID == task.Band }) {
		return
	}
	message := fmt.Sprintf("the task is in the band %s, which the manifest does not have", task.Band)
	if len(m.Bands) == 0 {
		message = fmt.Sprintf("the task is in the band %s, and the manifest has no bands", task.Band)
	}
	c.add(RuleTaskKnownBand, path, 0, 0, at("band"), message)
}

// ageCategories checks the age categories of a task or a lesson against the manifest's audience
// (task.age-categories, lesson.age-categories).
func (c *checker) ageCategories(rule, path string, categories []formatv0.AgeCategory, m formatv0.Manifest) {
	for k, category := range categories {
		if !slices.Contains(m.Audience.AgeCategories, category) {
			c.add(rule, path, 0, 0, at("age_categories", strconv.Itoa(k)),
				fmt.Sprintf("the age category %s is none of the manifest's audience.age_categories", category))
		}
	}
}

// pictures checks that the picture of every part of a list is a file of the module, in media/
// (task.known-pictures, lesson.known-pictures). P is a part of a statement or of a demo.
func pictures[P formatv0.StatementPart | formatv0.DemoPart](c *checker, rule, path string, media map[string]bool, parts []P, place ...string) {
	for k, part := range parts {
		var picture *formatv0.PicturePath
		switch part := any(part).(type) {
		case formatv0.StatementPart:
			picture = part.Picture
		case formatv0.DemoPart:
			picture = part.Picture
		}
		if picture != nil && !media[mediaFolder+string(*picture)] {
			c.add(rule, path, 0, 0, at(slices.Concat(place, []string{strconv.Itoa(k), "picture"})...),
				fmt.Sprintf("the picture %s is no file of the module in %s", *picture, mediaFolder))
		}
	}
}

// taskRules checks the rules of a task between parts.
func (c *checker) taskRules(path string, task formatv0.Task, p parts) {
	c.taskSkills(path, task, p)
	c.assessability(path, task, p.skillGraph)
	c.knownBand(path, task, *p.manifest)
	c.ageCategories(RuleTaskAgeCategories, path, task.AgeCategories, *p.manifest)
	pictures(c, RuleTaskKnownPictures, path, p.media, task.Statement, "statement")
	pictures(c, RuleTaskKnownPictures, path, p.media, task.Solution, "solution")
}

// lessonTasks gives each task that a lesson names and the module has, by the segment that names it.
func lessonTasks(lesson formatv0.Lesson, tasks map[string]formatv0.Task) map[int]formatv0.Task {
	found := map[int]formatv0.Task{}
	for i, segment := range lesson.Segments {
		if segment.Task == nil {
			continue
		}
		if task, ok := tasks[taskPath(string(*segment.Task))]; ok {
			found[i] = task
		}
	}
	return found
}

// lessonRules checks the rules of a lesson between parts: lesson.known-skills, lesson.known-tasks,
// lesson.task-skills, lesson.known-pictures and lesson.age-categories.
func (c *checker) lessonRules(path string, lesson formatv0.Lesson, p parts) {
	for k, skill := range lesson.Skills {
		if message, unknown := unknownSkill(p.skillGraph, "the lesson teaches", skill); unknown {
			c.add(RuleLessonKnownSkills, path, 0, 0, at("skills", strconv.Itoa(k)), message)
		}
	}
	tasks := lessonTasks(lesson, p.tasks)
	for i, segment := range lesson.Segments {
		task, has := tasks[i]
		switch {
		case segment.Task != nil && !has:
			c.add(RuleLessonKnownTasks, path, 0, 0, at("segments", strconv.Itoa(i), "task"),
				fmt.Sprintf("the task %s is no task of the module: it has no file %s", *segment.Task, taskPath(string(*segment.Task))))
		case has && !trains(task, lesson.Skills):
			c.add(RuleLessonTaskSkills, path, 0, 0, at("segments", strconv.Itoa(i), "task"),
				fmt.Sprintf("the task %s trains no skill of the lesson: no item of it has one among its primary skills", *segment.Task))
		}
		pictures(c, RuleLessonKnownPictures, path, p.media, segment.Explanation, "segments", strconv.Itoa(i), "explanation")
		if example := segment.WorkedExample; example != nil {
			pictures(c, RuleLessonKnownPictures, path, p.media, example.Statement, "segments", strconv.Itoa(i), "worked_example", "statement")
			pictures(c, RuleLessonKnownPictures, path, p.media, example.Demo, "segments", strconv.Itoa(i), "worked_example", "demo")
		}
	}
	c.ageCategories(RuleLessonAgeCategories, path, lesson.AgeCategories, *p.manifest)
	c.suits(path, lesson, tasks, p.manifest.Audience.AgeCategories)
}

// trains says whether a task trains one of the skills: for at least one item of it, the item's
// primary skills include one.
func trains(task formatv0.Task, skills []formatv0.SkillID) bool {
	for i := range task.Items {
		if slices.ContainsFunc(itemSkills(task, i).skills.Primary, func(s formatv0.SkillID) bool { return slices.Contains(skills, s) }) {
			return true
		}
	}
	return false
}

// suits checks the second half of lesson.age-categories: at least one age category suits the lesson
// and each task of the module that it names — the lesson's own, or else the module's, and each
// task's own, or else the module's.
func (c *checker) suits(path string, lesson formatv0.Lesson, tasks map[int]formatv0.Task, module []formatv0.AgeCategory) {
	own := func(categories []formatv0.AgeCategory) []formatv0.AgeCategory {
		if len(categories) == 0 {
			return module
		}
		return categories
	}
	common := slices.Clone(own(lesson.AgeCategories))
	for _, i := range slices.Sorted(maps.Keys(tasks)) {
		theirs := own(tasks[i].AgeCategories)
		common = slices.DeleteFunc(common, func(a formatv0.AgeCategory) bool { return !slices.Contains(theirs, a) })
	}
	if len(common) == 0 {
		c.add(RuleLessonAgeCategories, path, 0, 0, at(),
			"no age category suits both the lesson and each task of the module that it names: no learner can get it")
	}
}

// order checks lesson.order: the order of the module's lessons has no cycle. Lesson A comes after
// lesson B when a skill of A needs, through its prerequisites among the module's skills and theirs,
// a skill of B that is not one of A's own skills (format/README.md, section 21). The order comes
// from the skill graph, so a graph whose prerequisites run in a cycle, which breaks
// skill-graph.acyclic, gives no order, and the rule waits for it, as a rule between parts waits for
// a file that its schema refuses (R23).
func (c *checker) order(lessons map[string]formatv0.Lesson, g *formatv0.SkillGraph) {
	if g == nil {
		return
	}
	edges := edgesOf(*g)
	if len(cycles(edges)) > 0 {
		return
	}
	teaching := map[formatv0.SkillID][]string{} // the lessons that teach each skill, by path
	for path, lesson := range lessons {
		for _, skill := range lesson.Skills {
			teaching[skill] = append(teaching[skill], path)
		}
	}
	after := map[string][]string{} // the lessons that each lesson comes after, by path
	for path, a := range lessons {
		own := map[formatv0.SkillID]bool{}
		for _, skill := range a.Skills {
			own[skill] = true
		}
		needs := map[formatv0.SkillID]bool{}
		next := slices.Clone(a.Skills)
		for len(next) > 0 {
			skill := next[len(next)-1]
			next = next[:len(next)-1]
			for _, prerequisite := range edges[skill] {
				if needs[prerequisite] {
					continue
				}
				needs[prerequisite] = true
				next = append(next, prerequisite)
				if own[prerequisite] {
					continue
				}
				for _, other := range teaching[prerequisite] {
					if other != path && !slices.Contains(after[path], other) {
						after[path] = append(after[path], other)
					}
				}
			}
		}
	}
	for path, others := range cycles(after) {
		c.add(RuleLessonOrder, path, 0, 0, at("skills"),
			"the lesson comes, through the lessons that it comes after, after itself, in a cycle with "+strings.Join(others, ", "))
	}
}
