package module

import (
	"fmt"
	"maps"
	"slices"
	"strconv"
	"strings"

	"github.com/altnovatrail/spec/types/formatv0"

	"github.com/altnovatrail/module-kit/internal/normal"
	"github.com/altnovatrail/module-kit/leak"
	"github.com/altnovatrail/module-kit/verdict"
)

// The rules of the leak check on the texts of a module (format/README.md, sections 20 and 21).
const (
	RuleTaskNoLeaks   = "task.no-leaks"
	RuleLessonNoLeaks = "lesson.no-leaks"
)

// A text of a module that a rule of the leak check searches: the text, the place of the breach that
// it gives in the file of its task or its lesson, and, for a text of another part, where the text
// stands, its file and its place in it, such as techniques.yaml#/techniques/count-gaps/title.
type searchedText struct {
	text, place, at string
	// phrasesOnly is set for a text piece of a line of a worked example on a workbook, which is
	// searched for leak phrases but for no key: the keys of its lesson's pages are numbers that
	// the example's lines write by their nature (format/README.md, section 20).
	phrasesOnly bool
}

// said writes what holds a leak in a message: the text, or the text at its file and place.
func (t searchedText) said() string {
	if t.at == "" {
		return "the text"
	}
	return "the text at " + t.at
}

// searchOf gives the leak check of the task at path, a task that is no template, built once for
// every rule that searches for its keys and leak phrases. The check of a module reads a template
// through its instances (format/README.md, section 20), each on a checker of its own, so a template
// here is a fault of module-kit.
func (c *checker) searchOf(path string, task formatv0.Task) *leak.Search {
	if s, ok := c.searches[path]; ok {
		return s
	}
	s, err := leak.For(task)
	if err != nil {
		panic(fmt.Sprintf("module: the leak check of a task that is no template: %v", err))
	}
	if c.searches == nil {
		c.searches = map[string]*leak.Search{}
	}
	c.searches[path] = s
	return s
}

// ownShown gives the texts of a task itself that task.no-leaks searches, which the learner can see
// before the task is solved: its givens, unknown and leading question, the text of each of its trap
// values and the simpler case of each of its items; and, of an instance of a template, the picture
// path of each part of its statement and its solution, whose names the numbers of the instance fill.
// The picture path of a task that is no template is its author's name, and not searched
// (format/README.md, section 20). They are in its file, so the rule searches them within the part.
func ownShown(task formatv0.Task, instance bool) []searchedText {
	var texts []searchedText
	if task.LeadingQuestion != nil {
		texts = append(texts, searchedText{text: string(*task.LeadingQuestion), place: at("leading_question")})
	}
	if instance {
		for _, list := range []struct {
			parts []formatv0.StatementPart
			name  string
		}{{task.Statement, "statement"}, {task.Solution, "solution"}} {
			for k, part := range list.parts {
				if part.Picture != nil {
					texts = append(texts, searchedText{text: string(*part.Picture), place: at(list.name, strconv.Itoa(k), "picture")})
				}
			}
		}
	}
	for k, given := range task.Givens {
		texts = append(texts, searchedText{text: string(given), place: at("givens", strconv.Itoa(k))})
	}
	if task.Unknown != nil {
		texts = append(texts, searchedText{text: string(*task.Unknown), place: at("unknown")})
	}
	for i, item := range task.Items {
		slots, places := slotsOf(item, i)
		for n, slot := range slots {
			for k, trapValue := range slot.Check.TrapValues {
				if trapValue.Text != nil {
					tokens := slices.Concat(places[n], []string{"check", "trap_values", strconv.Itoa(k), "text"})
					texts = append(texts, searchedText{text: string(*trapValue.Text), place: at(tokens...)})
				}
			}
		}
		if item.SimplerCase != nil {
			texts = append(texts, searchedText{text: string(*item.SimplerCase), place: at("items", strconv.Itoa(i), "simpler_case")})
		}
	}
	return texts
}

// partsShown gives the texts of the module's other parts that task.no-leaks searches for a task: the
// title and description of each technique it names, the title and description of each solving step,
// which every task goes through, the trap_advice of each trap it names — in its traps and in the trap
// values of its checks — and, when it has an item with a verdict, the manifest's rung templates. A
// breach stands where the task names the technique or the trap, the first place for a trap named
// twice, and at the task as a whole for a solving step and a rung template. A technique or
// a trap that the module's catalog does not have breaks task.known-techniques or task.known-traps
// instead.
func partsShown(task formatv0.Task, p parts) []searchedText {
	var texts []searchedText
	if p.techniques != nil {
		for k, id := range task.Techniques {
			technique, ok := p.techniques.Techniques[id]
			if !ok {
				continue
			}
			place := at("techniques", strconv.Itoa(k))
			texts = append(texts, searchedText{text: string(technique.Title), place: place, at: techniquesPath + at("techniques", string(id), "title")})
			if technique.Description != nil {
				texts = append(texts, searchedText{text: string(*technique.Description), place: place, at: techniquesPath + at("techniques", string(id), "description")})
			}
		}
	}
	if p.traps != nil {
		named := map[formatv0.TrapID]bool{}
		for _, t := range namedTraps(task) {
			trap, ok := p.traps.Traps[t.id]
			if !ok || named[t.id] {
				continue
			}
			named[t.id] = true
			texts = append(texts, searchedText{text: string(trap.TrapAdvice), place: t.place, at: trapsPath + at("traps", string(t.id), "trap_advice")})
		}
	}
	if p.steps != nil {
		for k, step := range p.steps.Steps {
			texts = append(texts, searchedText{text: string(step.Title), place: at(), at: stepsPath + at("steps", strconv.Itoa(k), "title")})
			if step.Description != nil {
				texts = append(texts, searchedText{text: string(*step.Description), place: at(), at: stepsPath + at("steps", strconv.Itoa(k), "description")})
			}
		}
	}
	if hasVerdict(task) {
		for k, template := range p.manifest.RungTemplates {
			texts = append(texts, searchedText{text: string(template), place: at(), at: manifestPath + at("rung_templates", strconv.Itoa(k))})
		}
	}
	return texts
}

// A trap that a task names, and the place where it names it.
type trapAt struct {
	id    formatv0.TrapID
	place string
}

// namedTraps gives the traps that a task names, in its traps and then in the trap values of its
// checks, each with its place (format/README.md, section 16).
func namedTraps(task formatv0.Task) []trapAt {
	var found []trapAt
	for k, id := range task.Traps {
		found = append(found, trapAt{id, at("traps", strconv.Itoa(k))})
	}
	for i, item := range task.Items {
		slots, places := slotsOf(item, i)
		for n, slot := range slots {
			for k, trapValue := range slot.Check.TrapValues {
				found = append(found, trapAt{trapValue.Trap, at(slices.Concat(places[n], []string{"check", "trap_values", strconv.Itoa(k), "trap"})...)})
			}
		}
	}
	return found
}

// taskLeaks checks task.no-leaks on texts of a module: none of them holds a key or a leak phrase of
// the task, as the leak check finds them (format/README.md, section 20). A breach for each text that
// holds one, naming each it holds. It reads a task that is no template: a template is searched
// through its instances, each a task of its own (section 23; templates.go).
func (c *checker) taskLeaks(path string, task formatv0.Task, texts []searchedText) {
	if len(texts) == 0 {
		return
	}
	search := c.searchOf(path, task)
	for _, t := range texts {
		if found := search.In(t.text); len(found) > 0 {
			c.add(RuleTaskNoLeaks, path, 0, 0, t.place,
				fmt.Sprintf("%s holds %s, and the learner can see it before the task is solved", t.said(), leaksSaid(task, "", found)))
		}
	}
}

// lessonShown gives the texts of a lesson that lesson.no-leaks searches, each at its place: its
// title, the paragraphs, the alt of the pictures and the texts of the numbered lists of each
// explanation, the paragraphs and the alt of the pictures of each worked example's statement and
// demo, and, for leak phrases alone, the text pieces of the lines of a worked example on a workbook
// (format/README.md, sections 20 and 21).
func lessonShown(lesson formatv0.Lesson) []searchedText {
	texts := []searchedText{{text: string(lesson.Title), place: at("title")}}
	add := func(text *formatv0.Text, alt *string, tokens ...string) {
		if text != nil {
			texts = append(texts, searchedText{text: string(*text), place: at(slices.Concat(tokens, []string{"text"})...)})
		}
		if alt != nil {
			texts = append(texts, searchedText{text: *alt, place: at(slices.Concat(tokens, []string{"alt"})...)})
		}
	}
	for i, segment := range lesson.Segments {
		segmentAt := []string{"segments", strconv.Itoa(i)}
		for k, part := range segment.Explanation {
			partAt := slices.Concat(segmentAt, []string{"explanation", strconv.Itoa(k)})
			add(part.Text, part.Alt, partAt...)
			for j, text := range part.NumberedList {
				texts = append(texts, searchedText{text: string(text), place: at(slices.Concat(partAt, []string{"numbered_list", strconv.Itoa(j)})...)})
			}
		}
		if example := segment.WorkedExample; example != nil {
			for k, part := range example.Statement {
				add(part.Text, part.Alt, slices.Concat(segmentAt, []string{"worked_example", "statement", strconv.Itoa(k)})...)
			}
			for k, part := range example.Demo {
				add(part.Text, part.Alt, slices.Concat(segmentAt, []string{"worked_example", "demo", strconv.Itoa(k)})...)
			}
			for l, line := range example.Lines {
				for k, part := range line.Line {
					if part.Text != nil {
						texts = append(texts, searchedText{text: string(*part.Text), phrasesOnly: true,
							place: at(slices.Concat(segmentAt, []string{"worked_example", "lines", strconv.Itoa(l), "line", strconv.Itoa(k), "text"})...)})
					}
				}
			}
		}
	}
	return texts
}

// lessonLeaks checks lesson.no-leaks: no text of a lesson holds a key or a leak phrase of a task of
// the module that it names, as the leak check finds them, after the task's segment too, since the
// open view of a lesson reaches the learner's app whole (format/README.md, sections 20 and 21). A
// breach for each text and each task whose keys or leak phrases it holds. A task that the module does
// not have breaks lesson.known-tasks instead, and a template is searched through its instances, for
// the keys and leak phrases of each (section 23).
func (c *checker) lessonLeaks(path string, lesson formatv0.Lesson, p parts) {
	named := lessonTasks(lesson, p.tasks)
	texts := lessonShown(lesson)
	for _, i := range slices.Sorted(maps.Keys(named)) {
		file := taskPath(string(*lesson.Segments[i].Task))
		e, ok := p.templates[file]
		if !ok {
			c.lessonTaskLeaks(path, i, file, named[i], texts)
			c.exampleSolves(path, lesson, i, file, named[i])
			continue
		}
		c.onInstances(e, func(sub *checker, inst instance) {
			if len(inst.invalid) == 0 {
				sub.lessonTaskLeaks(path, i, file, inst.task, texts)
				sub.exampleSolves(path, lesson, i, file, inst.task)
			}
		})
	}
}

// lessonTaskLeaks checks lesson.no-leaks on the texts of a lesson for one task that it names, at the
// segment segment, a task of the file file that is no template: an instance, for a template.
func (c *checker) lessonTaskLeaks(path string, segment int, file string, task formatv0.Task, texts []searchedText) {
	search := c.searchOf(file, task)
	for _, t := range texts {
		found := search.In(t.text)
		if t.phrasesOnly {
			found = slices.DeleteFunc(found, func(l leak.Leak) bool { return l.Kind != leak.Phrase })
		}
		if len(found) > 0 {
			c.add(RuleLessonNoLeaks, path, 0, 0, t.place,
				fmt.Sprintf("the text holds %s, of the task that #/segments/%d/task names, and the learner sees the whole lesson before the task is solved", leaksSaid(task, file, found), segment))
		}
	}
}

// leaksSaid writes what the leak check found, each a key by the place of its answer slot or a leak
// phrase by its place, in the task's file, which file names when the breach is in another.
func leaksSaid(task formatv0.Task, file string, found []leak.Leak) string {
	said := make([]string, len(found))
	for n, l := range found {
		switch l.Kind {
		case leak.Key:
			_, places := slotsOf(task.Items[l.Item], l.Item)
			said[n] = fmt.Sprintf("the key %q of the answer slot at %s%s", l.Written, file, at(places[l.Slot]...))
		case leak.Phrase:
			said[n] = fmt.Sprintf("the leak phrase %q at %s%s", l.Written, file, at("leak_phrases", strconv.Itoa(l.Phrase)))
		}
	}
	if len(said) == 1 {
		return said[0]
	}
	return strings.Join(said[:len(said)-1], ", ") + " and " + said[len(said)-1]
}

// exampleSolves checks the part of lesson.no-leaks that reads the lines of a worked example on a
// workbook, for one task that the lesson names, at the segment segment, a task of the file file that
// is no template: an instance, for a template. A line of the example solves a line of the task when
// it has the same pieces in the same order, its texts are those of the task's line in their normal
// form, and the value of each box is equal, as its box's check compares it, to that box's key
// (format/README.md, section 20). A box whose check has no key, a predicate's, is solved by no value.
func (c *checker) exampleSolves(path string, lesson formatv0.Lesson, segment int, file string, task formatv0.Task) {
	for i, s := range lesson.Segments {
		example := s.WorkedExample
		if example == nil {
			continue
		}
		for l, line := range example.Lines {
			for n, item := range task.Items {
				if solves(line, item) {
					c.add(RuleLessonNoLeaks, path, 0, 0, at("segments", strconv.Itoa(i), "worked_example", "lines", strconv.Itoa(l)),
						fmt.Sprintf("the line solves item %d of the task that #/segments/%d/task names, %s: it has the item's pieces, and each box the item's key, and the learner sees the whole lesson before the task is solved", n+1, segment, file))
				}
			}
		}
	}
}

// solves says whether a line of a worked example solves the line of an item of a task, as
// exampleSolves reads it.
func solves(line formatv0.ExampleLine, item formatv0.Item) bool {
	if len(line.Line) == 0 || len(line.Line) != len(item.Line) {
		return false
	}
	slots := verdict.AnswerSlots(item)
	box := 0
	for k, piece := range line.Line {
		part := item.Line[k]
		switch {
		case piece.Text != nil && part.Text != nil:
			if normal.Search(string(*piece.Text)) != normal.Search(string(*part.Text)) {
				return false
			}
		case piece.Binding != nil && part.Binding != nil:
			slot := slots[box]
			box++
			if piece.Value == nil || slot.Check.Key == nil || !verdict.Same(slot, string(*piece.Value), *slot.Check.Key) {
				return false
			}
		default:
			return false
		}
	}
	return true
}
