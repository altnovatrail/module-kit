package module

import (
	"cmp"
	"fmt"
	"maps"
	"slices"
	"strconv"
	"strings"

	"github.com/altnovatrail/module-kit/internal/schema"
	"github.com/altnovatrail/module-kit/leak"
)

// A Warning is what a text of the module asks of its author, though the format refuses nothing for
// it: a text of the whole module that holds a number — a rung template, a trap's advice, a
// technique, a solving step (format/README.md, section 20; the 2026-11-08 amendment to R57 in spec).
// A warning is no rule: it has no ID and no test vectors, and changes no report's verdict (R50).
// Path is the file from the module's folder, with / between names, and Pointer the place of the
// text in it.
type Warning struct {
	Path    string
	Pointer string
	Message string
}

func (w Warning) String() string {
	return strings.TrimSpace(w.Path+" "+w.Pointer) + ": warning: " + w.Message
}

// warn gives the warnings of the texts of the whole module: a rung template of the manifest, the
// advice of a trap, the title and description of a technique, and the title and description of a
// solving step, each that holds a number as the leak check reads it. task.no-leaks searches such a
// text for the keys of every task that can show it, so it SHOULD hold no number (format/README.md,
// section 20): the rule finds one only once it is the key of such a task, and the warning before.
// It reads each part whose file passed its schema, whatever else the module breaks, in the order of
// their places.
func (p parts) warn() []Warning {
	var warnings []Warning
	add := func(text, path string, tokens ...string) {
		if numbers := leak.Numbers(text); len(numbers) > 0 {
			warnings = append(warnings, Warning{Path: path, Pointer: at(tokens...), Message: numbersSaid(numbers)})
		}
	}
	if p.manifest != nil {
		for k, template := range p.manifest.RungTemplates {
			add(string(template), manifestPath, "rung_templates", strconv.Itoa(k))
		}
	}
	if p.traps != nil {
		for _, id := range slices.Sorted(maps.Keys(p.traps.Traps)) {
			add(string(p.traps.Traps[id].TrapAdvice), trapsPath, "traps", string(id), "trap_advice")
		}
	}
	if p.techniques != nil {
		for _, id := range slices.Sorted(maps.Keys(p.techniques.Techniques)) {
			technique := p.techniques.Techniques[id]
			add(string(technique.Title), techniquesPath, "techniques", string(id), "title")
			if technique.Description != nil {
				add(string(*technique.Description), techniquesPath, "techniques", string(id), "description")
			}
		}
	}
	if p.steps != nil {
		for k, step := range p.steps.Steps {
			add(string(step.Title), stepsPath, "steps", strconv.Itoa(k), "title")
			if step.Description != nil {
				add(string(*step.Description), stepsPath, "steps", strconv.Itoa(k), "description")
			}
		}
	}
	slices.SortStableFunc(warnings, func(a, b Warning) int {
		return cmp.Or(strings.Compare(a.Path, b.Path), schema.ComparePointers(a.Pointer, b.Pointer))
	})
	return warnings
}

// numbersSaid writes the message of a warning of a text that holds numbers, each as it stands in
// the text's normal form, once, in the order of the text.
func numbersSaid(numbers []string) string {
	var once []string
	for _, n := range numbers {
		if quoted := strconv.Quote(n); !slices.Contains(once, quoted) {
			once = append(once, quoted)
		}
	}
	held := "the number " + once[0]
	if len(once) > 1 {
		held = "the numbers " + strings.Join(once[:len(once)-1], ", ") + " and " + once[len(once)-1]
	}
	return fmt.Sprintf("the text holds %s: a text of the whole module is searched for the keys of every task that can show it, "+
		"so it SHOULD hold no number (format/README.md, section 20)", held)
}
