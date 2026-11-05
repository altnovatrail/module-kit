package moduleversion

import (
	"cmp"
	"fmt"
	"maps"
	"reflect"
	"slices"
	"strconv"
	"strings"

	"github.com/altnovatrail/spec/types/formatv0"

	"github.com/altnovatrail/module-kit/internal/normal"
	"github.com/altnovatrail/module-kit/internal/schema"
	"github.com/altnovatrail/module-kit/modulepackage"
	"github.com/altnovatrail/module-kit/template"
	"github.com/altnovatrail/module-kit/verdict"
)

// A Change is one change from a published version of a module to a new one, and the part of the
// version that it raises (format/README.md, section 27): where it is — the file of the package,
// modulepackage.ServerFile for the canonical view and modulepackage.DeviceFile for the listing of the
// device part, and the place in it as a JSON Pointer in URI fragment form, in the new version, or in
// the published one for what is gone — the numbers of the parameters of the instance of a template
// that it is of, and what changes, in words.
type Change struct {
	Part       Part             `json:"part"`
	File       string           `json:"file"`
	Pointer    string           `json:"pointer"`
	Parameters formatv0.Numbers `json:"parameters,omitzero"`
	What       string           `json:"what"`
}

// String writes the change as a line of the report, as a breach of the check of a module is written:
// the place, then the part, then what changes, as in
// "server.json #/tasks/apples: minor: the task apples is added".
func (c Change) String() string {
	what := c.What
	if len(c.Parameters) > 0 {
		what = "on the instance " + shownNumbers(c.Parameters) + ": " + what
	}
	return c.File + " " + c.Pointer + ": " + c.Part.String() + ": " + what
}

// shownNumbers writes numbers by their names, in the order of the names, as "a = 2, b = 3".
func shownNumbers(numbers formatv0.Numbers) string {
	var shown []string
	for _, name := range slices.Sorted(maps.Keys(numbers)) {
		shown = append(shown, fmt.Sprintf("%s = %d", name, numbers[name]))
	}
	return strings.Join(shown, ", ")
}

// A change is a Change with its place as tokens, by which the changes are ordered and a difference
// of the patch row finds the change of the table that covers it.
type change struct {
	Change
	tokens []string
}

// comparer gathers the changes from the canonical view of one version to that of another.
type comparer struct {
	changes []change
}

func (c *comparer) add(part Part, tokens []string, parameters formatv0.Numbers, what string) {
	c.changes = append(c.changes, change{
		Change: Change{Part: part, File: modulepackage.ServerFile, Pointer: schema.Pointer(tokens), Parameters: parameters, What: what},
		tokens: slices.Clone(tokens),
	})
}

// Changes compares two packages of one module and one version of the format — before, the published
// version, and after, the new one — and gives the part of the version that the comparison finds, the
// highest that a change raises, and every change, by the table of format/README.md, section 27: each
// row of the table at its place, and anything else that differs in the canonical view or in the
// listing of the device part, but the version, as a patch at each place where the two differ. A
// difference at or within the place of a change that a row gives is not given again. The changes come
// in the order of their parts, from the highest, and then of their places.
//
// The error is for two packages of other modules or of other versions of the format, and for a
// package that is not one modulepackage builds or reads.
func Changes(before, after *modulepackage.Package) (Part, []Change, error) {
	b, a := before.Inventory, after.Inventory
	if b.ID != a.ID {
		return None, nil, fmt.Errorf("moduleversion: %s and %s are two modules: no version of one compares with the other", b.ID, a.ID)
	}
	if b.ModuleFormatVersion != a.ModuleFormatVersion {
		return None, nil, fmt.Errorf("moduleversion: %s %s is of the format %s, and %s of %s: the table compares versions of one format", b.ID, b.Version, b.ModuleFormatVersion, a.Version, a.ModuleFormatVersion)
	}
	c := &comparer{}
	bv, av := before.View, after.View
	c.manifests(bv.Manifest, av.Manifest)
	c.skillGraphs(bv.SkillGraph, av.SkillGraph)
	c.catalogs(bv, av)
	c.solvingSteps(bv.SolvingSteps, av.SolvingSteps)
	if err := c.tasks(bv, av); err != nil {
		return None, nil, err
	}
	c.lessons(bv, av)
	if err := c.differences(before, after); err != nil {
		return None, nil, err
	}
	part := None
	changes := make([]Change, len(c.changes))
	slices.SortFunc(c.changes, compareChanges)
	for i, ch := range c.changes {
		part = max(part, ch.Part)
		changes[i] = ch.Change
	}
	return part, changes, nil
}

// compareChanges orders changes by their parts, from the highest, then by their places — the
// canonical view before the listing of the device part, and in a file member by member — then by the
// numbers of their instances, and then by their words.
func compareChanges(x, y change) int {
	return cmp.Or(cmp.Compare(y.Part, x.Part), boolOrder(x.File == modulepackage.DeviceFile, y.File == modulepackage.DeviceFile),
		schema.ComparePointers(x.Pointer, y.Pointer), strings.Compare(shownNumbers(x.Parameters), shownNumbers(y.Parameters)),
		strings.Compare(x.What, y.What))
}

// boolOrder orders false before true.
func boolOrder(a, b bool) int {
	switch {
	case a == b:
		return 0
	case a:
		return 1
	}
	return -1
}

// The sets that the table compares: what is in one and not in the other, in order.
func setOf[T cmp.Ordered](items []T) map[T]bool {
	s := make(map[T]bool, len(items))
	for _, item := range items {
		s[item] = true
	}
	return s
}

func without[T cmp.Ordered](a, b map[T]bool) []T {
	var gone []T
	for item := range a {
		if !b[item] {
			gone = append(gone, item)
		}
	}
	slices.Sort(gone)
	return gone
}

func sameSet[T cmp.Ordered](a, b []T) bool {
	x, y := setOf(a), setOf(b)
	return len(without(x, y)) == 0 && len(without(y, x)) == 0
}

func sorted[T cmp.Ordered](items []T) []T {
	s := slices.Clone(items)
	slices.Sort(s)
	return slices.Compact(s)
}

func words[T ~string](items []T) string {
	s := make([]string, len(items))
	for i, item := range items {
		s[i] = string(item)
	}
	return strings.Join(s, ", ")
}

// suited compares the age categories that a task or a lesson suits, or the module: a set that
// narrows is a major change, and one that widens a minor one.
func (c *comparer) suited(tokens []string, before, after []formatv0.AgeCategory) {
	x, y := setOf(before), setOf(after)
	if gone := without(x, y); len(gone) > 0 {
		c.add(Major, tokens, nil, "the age categories narrow: "+words(gone)+" no longer")
	} else if added := without(y, x); len(added) > 0 {
		c.add(Minor, tokens, nil, "the age categories widen: "+words(added)+" too")
	}
}

// own gives the age categories of a task or a lesson: its own, or else the module's of its version
// (format/README.md, sections 14 and 21).
func own(categories []formatv0.AgeCategory, m formatv0.Manifest) []formatv0.AgeCategory {
	if len(categories) > 0 {
		return categories
	}
	return m.Audience.AgeCategories
}

func (c *comparer) manifests(b, a formatv0.Manifest) {
	if b.Subject != a.Subject {
		c.add(Major, []string{"manifest", "subject"}, nil, "the subject changes")
	}
	if b.License != a.License {
		c.add(Major, []string{"manifest", "license"}, nil, "the license changes")
	}
	c.suited([]string{"manifest", "audience", "age_categories"}, b.Audience.AgeCategories, a.Audience.AgeCategories)
	if !sameSet(b.Audience.Grades, a.Audience.Grades) {
		c.add(Minor, []string{"manifest", "audience", "grades"}, nil, "the grades change")
	}
	// The bands kept, each at its place: the grades of one are a minor change, and a band after the
	// last a minor one too. A band removed, moved or put before another is module-version.bands-kept's.
	kept := len(a.Bands) >= len(b.Bands)
	for i, band := range b.Bands {
		kept = kept && a.Bands[i].ID == band.ID
	}
	if kept {
		for i, band := range b.Bands {
			if !sameSet(band.Grades, a.Bands[i].Grades) {
				c.add(Minor, []string{"manifest", "bands", strconv.Itoa(i), "grades"}, nil, fmt.Sprintf("the grades of the band %s change", band.ID))
			}
		}
		for i := len(b.Bands); i < len(a.Bands); i++ {
			c.add(Minor, []string{"manifest", "bands", strconv.Itoa(i)}, nil, fmt.Sprintf("the band %s comes after the last", a.Bands[i].ID))
		}
	}
	if x, y := digitHeight(b), digitHeight(a); x != y {
		c.add(Minor, []string{"manifest", "canvas", "digit_height"}, nil, fmt.Sprintf("the height of a digit on a crop changes from %d to %d", x, y))
	}
}

// digitHeight gives the height of a digit on a crop that a manifest gives: 24 when it gives none
// (format/README.md, section 11).
func digitHeight(m formatv0.Manifest) int64 {
	if m.Canvas == nil || m.Canvas.DigitHeight == nil {
		return 24
	}
	return *m.Canvas.DigitHeight
}

// keys compares the entries of a kind by their local IDs: one removed is a major change, and one
// added a minor one. It gives the IDs that both have, in order.
func keys[K ~string, V any](c *comparer, at []string, before, after map[K]V, what string) []K {
	var both []K
	for _, id := range slices.Sorted(maps.Keys(before)) {
		if _, ok := after[id]; ok {
			both = append(both, id)
		} else {
			c.add(Major, append(slices.Clone(at), string(id)), nil, fmt.Sprintf("the %s %s is removed", what, id))
		}
	}
	for _, id := range slices.Sorted(maps.Keys(after)) {
		if _, ok := before[id]; !ok {
			c.add(Minor, append(slices.Clone(at), string(id)), nil, fmt.Sprintf("the %s %s is added", what, id))
		}
	}
	return both
}

func (c *comparer) skillGraphs(b, a *formatv0.SkillGraph) {
	var bs, as map[formatv0.SkillID]formatv0.Skill
	var br, ar map[formatv0.SkillID]formatv0.RetiredSkill
	if b != nil {
		bs, br = b.Skills, b.RetiredSkills
	}
	if a != nil {
		as, ar = a.Skills, a.RetiredSkills
	}
	at := []string{"skill_graph", "skills"}
	for _, id := range keys(c, at, bs, as, "skill") {
		was, is := bs[id], as[id]
		here := append(slices.Clone(at), string(id))
		if !sameSet(was.Prerequisites, is.Prerequisites) {
			c.add(Minor, append(here, "prerequisites"), nil, fmt.Sprintf("the prerequisites of %s change", id))
		}
		taught, teaches := setOf(was.Teaches), setOf(is.Teaches)
		if gone := without(taught, teaches); len(gone) > 0 {
			c.add(Major, append(here, "teaches"), nil, fmt.Sprintf("%s no longer teaches %s", id, words(gone)))
		} else if added := without(teaches, taught); len(added) > 0 {
			c.add(Minor, append(here, "teaches"), nil, fmt.Sprintf("%s teaches %s too", id, words(added)))
		}
		if was.Assessability != is.Assessability {
			c.add(Major, append(here, "assessability"), nil, fmt.Sprintf("the assessability of %s changes", id))
		}
	}
	// A new entry of retired_skills comes with a retirement, and replaced_by changes only as a skill it
	// names is retired in turn: either goes with a major change. replaced_by is a set.
	for _, id := range sorted(slices.Concat(slices.Collect(maps.Keys(br)), slices.Collect(maps.Keys(ar)))) {
		was, inBefore := br[id]
		is, inAfter := ar[id]
		here := []string{"skill_graph", "retired_skills", string(id)}
		switch {
		case !inAfter:
			c.add(Major, here, nil, fmt.Sprintf("the entry %s of the retired skills is gone", id))
		case !inBefore:
			c.add(Major, here, nil, fmt.Sprintf("%s is retired", id))
		case !slices.Equal(sorted(was.ReplacedBy), sorted(is.ReplacedBy)):
			c.add(Major, here, nil, fmt.Sprintf("what replaces %s changes", id))
		}
	}
}

func (c *comparer) catalogs(b, a formatv0.CanonicalView) {
	var bt, at map[formatv0.TrapID]formatv0.Trap
	if b.Traps != nil {
		bt = b.Traps.Traps
	}
	if a.Traps != nil {
		at = a.Traps.Traps
	}
	for _, id := range keys(c, []string{"traps", "traps"}, bt, at, "trap") {
		if !sameSet(bt[id].Skills, at[id].Skills) {
			c.add(Minor, []string{"traps", "traps", string(id), "skills"}, nil, fmt.Sprintf("the skills of the trap %s change", id))
		}
	}
	var bq, aq map[formatv0.TechniqueID]formatv0.Technique
	if b.Techniques != nil {
		bq = b.Techniques.Techniques
	}
	if a.Techniques != nil {
		aq = a.Techniques.Techniques
	}
	keys(c, []string{"techniques", "techniques"}, bq, aq, "technique")
}

// solvingSteps compares the solving steps, which have no IDs: by their number and, at each place, by
// what the step gives, as a set, and its exit condition. Their titles are words.
func (c *comparer) solvingSteps(b, a *formatv0.SolvingSteps) {
	at := []string{"solving_steps"}
	switch {
	case b != nil && a == nil:
		c.add(Major, at, nil, "the solving steps are removed")
		return
	case b == nil && a != nil:
		c.add(Minor, at, nil, "the solving steps are added")
		return
	case b == nil:
		return
	}
	at = append(at, "steps")
	if len(b.Steps) != len(a.Steps) {
		c.add(Minor, at, nil, fmt.Sprintf("the number of the solving steps changes from %d to %d", len(b.Steps), len(a.Steps)))
		return
	}
	for i, was := range b.Steps {
		is := a.Steps[i]
		if !slices.Equal(sorted(was.Gives), sorted(is.Gives)) || was.ExitCondition != is.ExitCondition {
			c.add(Minor, append(slices.Clone(at), strconv.Itoa(i)), nil, fmt.Sprintf("what step %d gives or its exit condition changes", i+1))
		}
	}
}

func (c *comparer) tasks(b, a formatv0.CanonicalView) error {
	for _, id := range keys(c, []string{"tasks"}, b.Tasks, a.Tasks, "task") {
		was, is := b.Tasks[id], a.Tasks[id]
		here := []string{"tasks", string(id)}
		c.suited(append(slices.Clone(here), "age_categories"), own(was.AgeCategories, b.Manifest), own(is.AgeCategories, a.Manifest))
		c.guidance(here, was, is)
		if (was.Template == nil) != (is.Template == nil) {
			becomes := "becomes"
			if is.Template == nil {
				becomes = "is no longer"
			}
			c.add(Major, append(slices.Clone(here), "template"), nil, fmt.Sprintf("the task %s %s a template", id, becomes))
			continue
		}
		if was.Canvas.Kind != is.Canvas.Kind {
			c.add(Major, append(slices.Clone(here), "canvas", "kind"), nil, fmt.Sprintf("the canvas kind changes from %s to %s", was.Canvas.Kind, is.Canvas.Kind))
			continue
		}
		if was.Template == nil {
			c.items(here, was, is, true, nil)
			continue
		}
		c.items(here, was, is, false, nil)
		if err := c.instances(id, b, a); err != nil {
			return err
		}
	}
	return nil
}

// guidance compares what a task gives the learner model and the engine beside its checks: its
// skills and those of its items at the places of both, its band, its difficulty, its techniques and
// its traps, each a minor change.
func (c *comparer) guidance(here []string, was, is formatv0.Task) {
	if !sameSkills(&was.Skills, &is.Skills) {
		c.add(Minor, append(slices.Clone(here), "skills"), nil, "the skills change")
	}
	for i := range min(len(was.Items), len(is.Items)) {
		if !sameSkills(was.Items[i].Skills, is.Items[i].Skills) {
			c.add(Minor, append(slices.Clone(here), "items", strconv.Itoa(i), "skills"), nil, "the skills change")
		}
	}
	if was.Band != is.Band {
		c.add(Minor, append(slices.Clone(here), "band"), nil, fmt.Sprintf("the band changes from %s to %s", was.Band, is.Band))
	}
	if was.Difficulty != is.Difficulty {
		c.add(Minor, append(slices.Clone(here), "difficulty"), nil, fmt.Sprintf("the difficulty changes from %d to %d", was.Difficulty, is.Difficulty))
	}
	if !sameSet(was.Techniques, is.Techniques) {
		c.add(Minor, append(slices.Clone(here), "techniques"), nil, "the techniques change")
	}
	if !sameSet(was.Traps, is.Traps) {
		c.add(Minor, append(slices.Clone(here), "traps"), nil, "the traps change")
	}
}

// sameSkills says whether two sets of skills by roles are the same, each role as a set; skills that
// an item does not give are the same as none other.
func sameSkills(a, b *formatv0.Skills) bool {
	if a == nil || b == nil {
		return a == b
	}
	return sameSet(a.Primary, b.Primary) && sameSet(a.Supporting, b.Supporting)
}

// items compares the items and the answer slots of two tasks by their places: their number, the
// binding of each answer slot and, withChecks, what decides its verdict and trap and the criteria of
// a rubric (format/README.md, section 27). A template as it is written is compared without its
// checks, which its instances compare, each with its numbers, parameters.
func (c *comparer) items(here []string, was, is formatv0.Task, withChecks bool, parameters formatv0.Numbers) {
	if len(was.Items) != len(is.Items) {
		c.add(Major, append(slices.Clone(here), "items"), parameters, fmt.Sprintf("the number of items changes from %d to %d", len(was.Items), len(is.Items)))
		return
	}
	for i := range was.Items {
		ws, ss := slotsOf(was.Items[i]), slotsOf(is.Items[i])
		item := append(slices.Clone(here), "items", strconv.Itoa(i))
		if len(ws) != len(ss) {
			c.add(Major, item, parameters, fmt.Sprintf("the number of answer slots of item %d changes from %d to %d", i+1, len(ws), len(ss)))
			continue
		}
		for n, w := range ws {
			s := ss[n]
			at := append(slices.Clone(item), s.tokens...)
			switch {
			case w.Binding != s.Binding:
				c.add(Major, at, parameters, fmt.Sprintf("the binding of answer slot %d changes from %s to %s", n+1, w.Binding, s.Binding))
			case !withChecks:
			case !reflect.DeepEqual(meaningOf(w.AnswerSlot), meaningOf(s.AnswerSlot)):
				c.add(Major, at, parameters, fmt.Sprintf("what decides the verdict or the trap of answer slot %d changes", n+1))
			case w.Check.Kind == formatv0.CheckKindRubric && !slices.Equal(w.Check.Criteria, s.Check.Criteria):
				c.add(Minor, at, parameters, fmt.Sprintf("the criteria of the rubric of answer slot %d change", n+1))
			}
		}
	}
}

// A slot is an answer slot of an item and its place in the item: answer_slots and its index on a
// handwriting sheet, line and the index of its box on a workbook.
type slot struct {
	verdict.AnswerSlot
	tokens []string
}

// slotsOf gives the answer slots of an item in the order of verdict.AnswerSlots, with their places.
func slotsOf(item formatv0.Item) []slot {
	var places [][]string
	for j := range item.AnswerSlots {
		places = append(places, []string{"answer_slots", strconv.Itoa(j)})
	}
	for k, part := range item.Line {
		if part.Binding != nil {
			places = append(places, []string{"line", strconv.Itoa(k)})
		}
	}
	slots := verdict.AnswerSlots(item)
	s := make([]slot, len(slots))
	for i, a := range slots {
		s[i] = slot{a, places[i]}
	}
	return s
}

// A meaning is what decides the verdict and the trap of an answer slot, as its check kind reads it
// (format/README.md, sections 15 and 27): the kind; for number, the key as a rational number and the
// set of the forms of its units; for fraction, the key as a rational number and the form, any when
// absent; for choice, the key and the set of options — the check's own, or the answer slot's when it
// is bound choice — each as a choice normalizes it; for predicate, the predicate as it is written;
// and the set of trap values, each the value as the check compares it, with its trap. A value out of
// a number's notation compares as "", as none of its values does.
type meaning struct {
	kind                 formatv0.CheckKind
	key, form, predicate string
	units, options       []string
	traps                [][2]string
}

func meaningOf(s verdict.AnswerSlot) meaning {
	c := s.Check
	m := meaning{kind: c.Kind}
	units := normal.Units(c)
	compared := func(text string) string {
		form, _ := normal.Compared(c, units, text)
		return form
	}
	switch c.Kind {
	case formatv0.CheckKindNumber, formatv0.CheckKindFraction, formatv0.CheckKindChoice:
		if c.Key != nil {
			m.key = compared(*c.Key)
		}
		for _, t := range c.TrapValues {
			m.traps = append(m.traps, [2]string{compared(t.Value), string(t.Trap)})
		}
		slices.SortFunc(m.traps, func(x, y [2]string) int { return cmp.Or(strings.Compare(x[0], y[0]), strings.Compare(x[1], y[1])) })
		m.traps = slices.Compact(m.traps)
	case formatv0.CheckKindPredicate:
		if c.Predicate != nil {
			m.predicate = *c.Predicate
		}
	}
	switch c.Kind {
	case formatv0.CheckKindNumber:
		m.units = sorted(units)
	case formatv0.CheckKindFraction:
		m.form = string(formatv0.CheckFormAny)
		if c.Form != nil {
			m.form = string(*c.Form)
		}
	case formatv0.CheckKindChoice:
		for _, option := range s.ChoiceOptions() {
			m.options = append(m.options, normal.Choice(string(option)))
		}
		m.options = sorted(m.options)
	}
	return m
}

// instances compares a template of both versions through its instances, matched by the numbers of
// their parameters, not by their places: parameters of other names match none, a major change; an
// instance in one version alone is a minor change; and the checks of two matched ones compare, each
// the task with its numbers in place (format/README.md, section 27).
func (c *comparer) instances(id formatv0.TaskID, b, a formatv0.CanonicalView) error {
	was, is := b.Tasks[id], a.Tasks[id]
	names := slices.Sorted(maps.Keys(was.Template.Parameters))
	if !slices.Equal(names, slices.Sorted(maps.Keys(is.Template.Parameters))) {
		c.add(Major, []string{"tasks", string(id), "template", "parameters"}, nil, "the parameters have other names: no instance matches")
		return nil
	}
	parametersOf := func(numbers formatv0.Numbers) formatv0.Numbers {
		p := formatv0.Numbers{}
		for _, name := range names {
			p[name] = numbers[name]
		}
		return p
	}
	byParameters := func(all []formatv0.Numbers) map[string]formatv0.Numbers {
		m := map[string]formatv0.Numbers{}
		for _, numbers := range all {
			m[shownNumbers(parametersOf(numbers))] = numbers
		}
		return m
	}
	old, now := byParameters(b.Instances[id]), byParameters(a.Instances[id])
	at := []string{"instances", string(id)}
	for _, key := range slices.Sorted(maps.Keys(old)) {
		numbers := old[key]
		match, ok := now[key]
		if !ok {
			c.add(Minor, at, parametersOf(numbers), "the instance is no longer one")
			continue
		}
		before, err := template.Instance(was, numbers)
		if err != nil {
			return fmt.Errorf("moduleversion: %w", err)
		}
		after, err := template.Instance(is, match)
		if err != nil {
			return fmt.Errorf("moduleversion: %w", err)
		}
		c.items([]string{"tasks", string(id)}, before, after, true, parametersOf(numbers))
	}
	for _, key := range slices.Sorted(maps.Keys(now)) {
		if _, ok := old[key]; !ok {
			c.add(Minor, at, parametersOf(now[key]), "the instance is new")
		}
	}
	return nil
}

// lessons compares the lessons, their segments by their places — what each is, and for a task the
// task —, their skills and their age categories. A segment after the last is a minor change, and any
// other change of the segments a major one.
func (c *comparer) lessons(b, a formatv0.CanonicalView) {
	for _, id := range keys(c, []string{"lessons"}, b.Lessons, a.Lessons, "lesson") {
		was, is := b.Lessons[id], a.Lessons[id]
		here := []string{"lessons", string(id)}
		old, now := segmentsOf(was), segmentsOf(is)
		changed := -1
		for i, segment := range old {
			if i >= len(now) || now[i] != segment {
				changed = i
				break
			}
		}
		switch {
		case changed >= 0:
			c.add(Major, append(slices.Clone(here), "segments", strconv.Itoa(changed)), nil, fmt.Sprintf("segment %d is removed, put before another, moved or replaced", changed+1))
		case len(now) > len(old):
			c.add(Minor, append(slices.Clone(here), "segments", strconv.Itoa(len(old))), nil, "a segment comes after the last")
		}
		if !sameSet(was.Skills, is.Skills) {
			c.add(Minor, append(slices.Clone(here), "skills"), nil, "the skills change")
		}
		c.suited(append(slices.Clone(here), "age_categories"), own(was.AgeCategories, b.Manifest), own(is.AgeCategories, a.Manifest))
	}
}

// segmentsOf gives what each segment of a lesson is: an explanation, a worked example, or the task
// that it names.
func segmentsOf(l formatv0.Lesson) []string {
	s := make([]string, len(l.Segments))
	for i, segment := range l.Segments {
		switch {
		case segment.Task != nil:
			s[i] = "task " + string(*segment.Task)
		case segment.WorkedExample != nil:
			s[i] = "worked_example"
		default:
			s[i] = "explanation"
		}
	}
	return s
}
