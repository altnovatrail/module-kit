package module

import (
	"context"
	"fmt"
	"runtime"
	"slices"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/altnovatrail/spec/types/formatv0"

	"github.com/altnovatrail/module-kit/internal/schema"
	"github.com/altnovatrail/module-kit/template"
)

// The rules of templates (format/README.md, sections 23 and 24): those that a template breaks by
// itself, which package template reads, and task.instances, which holds each of its instances to the
// schema of a task.
const (
	RuleTaskTemplateNames       = template.RuleNames
	RuleTaskTemplateExpressions = template.RuleExpressions
	RuleTaskPlaceholders        = template.RulePlaceholders
	RuleTaskTemplateInstances   = template.RuleInstances
	RuleTaskTemplateSolver      = template.RuleSolver
	RuleTaskInstances           = "task.instances"
)

// An instance of a template, as the check of a module reads it: its numbers as a breach names them,
// the task that it is, and the keywords of the schema of a task that it fails, which break
// task.instances.
type instance struct {
	shown   string
	task    formatv0.Task
	invalid []schema.Error
}

// What the check of a module reads of a template: the breaches of its own rules, and, when it breaks
// none, its instances, in their order, with the results of its solver among their numbers
// (format/README.md, sections 23 and 24): numbers as template.Instances gives them, and each
// instance as the rules read it.
type expansion struct {
	breaches  []template.Breach
	numbers   []formatv0.Numbers
	instances []instance
}

// expand counts the instances of a template, running its solver with ctx, and builds each one, the
// task with its numbers in place, against the schema of a task. The error is that of ctx, once it is
// done. A template whose placeholders break no rule has an instance for its numbers, so an error of
// template.Instance is a fault of module-kit.
func expand(ctx context.Context, task formatv0.Task) (*expansion, error) {
	numbers, breaches, err := template.Instances(ctx, task)
	if err != nil {
		return nil, err
	}
	e := &expansion{breaches: breaches, numbers: numbers}
	for _, n := range numbers {
		t, err := template.Instance(task, n)
		if err != nil {
			panic(fmt.Sprintf("module: an instance of a template that breaks no rule of templates: %v", err))
		}
		e.instances = append(e.instances, instance{
			shown:   template.Shown(*task.Template, n),
			task:    t,
			invalid: causes(schema.Validate("task.schema.json", schema.ValueOf(t))),
		})
	}
	return e, nil
}

// expandAll expands every template among the tasks of parts, one on each CPU that the process may use
// at once, runtime.GOMAXPROCS(0), each solver on a Starlark thread of its own, and puts each
// expansion under the path of its task, whatever order they end in (architecture 3.4). The solvers
// share the wall time of their clocks and the memory of the process, so no more run at once than the
// CPUs that serve them. The error is that of ctx, once it is done; the templates that have not begun
// then do not run.
func (p *parts) expandAll(ctx context.Context) error {
	var paths []string
	for path, task := range p.tasks {
		if task.Template != nil {
			paths = append(paths, path)
		}
	}
	slices.Sort(paths)
	expansions := make([]*expansion, len(paths))
	errs := make([]error, len(paths))
	var next atomic.Int64
	var wg sync.WaitGroup
	for range min(runtime.GOMAXPROCS(0), len(paths)) {
		wg.Go(func() {
			for {
				i := int(next.Add(1)) - 1
				if i >= len(paths) {
					return
				}
				if err := ctx.Err(); err != nil {
					errs[i] = err
					continue
				}
				expansions[i], errs[i] = expand(ctx, p.tasks[paths[i]])
			}
		})
	}
	wg.Wait()
	for i, path := range paths {
		if errs[i] != nil {
			return errs[i]
		}
		p.templates[path] = expansions[i]
	}
	return nil
}

// instancesOf gives the numbers of the instances of each template of a module by the local ID of its
// task, as the canonical view holds them (format/README.md, section 25).
func (p *parts) instancesOf() map[string][]formatv0.Numbers {
	if len(p.templates) == 0 {
		return nil
	}
	before, after, _ := strings.Cut(taskRow.path, anyID)
	instances := make(map[string][]formatv0.Numbers, len(p.templates))
	for path, e := range p.templates {
		instances[strings.TrimSuffix(strings.TrimPrefix(path, before), after)] = e.numbers
	}
	return instances
}

// causes gives the keywords that an instance of a template fails but those of unevaluatedProperties,
// unless they are all. An instance holds the fields of its template, which the schema of a task took,
// so a field is left unevaluated only where a subschema that evaluates it fails, which is a keyword
// of its own: the author mends that one (R23).
func causes(errs []schema.Error) []schema.Error {
	if kept := slices.DeleteFunc(slices.Clone(errs), func(e schema.Error) bool { return e.Keyword == "unevaluatedProperties" }); len(kept) > 0 {
		return kept
	}
	return errs
}

// templateRules checks the rules of templates on a template, and the rules of a task within its part
// on each of its instances (format/README.md, section 23): its own rules; task.instances, of an
// instance that the schema of a task does not take, each keyword at each place by the first instance
// that fails it there, since each keyword is a mistake of its own; and on every other instance the
// rules that the check of an answer slot sees itself and task.no-leaks on the texts of the task.
func (c *checker) templateRules(path string, e *expansion) {
	for _, b := range e.breaches {
		c.add(b.Rule, path, 0, 0, b.Pointer, b.Message)
	}
	failed := map[string]bool{} // each keyword at each place that an instance has failed
	for _, i := range e.instances {
		for _, err := range i.invalid {
			if place := err.Pointer() + " " + err.Keyword; !failed[place] {
				failed[place] = true
				c.add(RuleTaskInstances, path, 0, 0, err.Pointer(),
					fmt.Sprintf("on the instance %s: the instance is not valid against task.schema.json: %s: %s", i.shown, err.Keyword, err.Message))
			}
		}
	}
	c.onInstances(e, func(sub *checker, i instance) {
		if len(i.invalid) == 0 {
			sub.checkRules(path, i.task)
			sub.taskLeaks(path, i.task, ownShown(i.task))
		}
	})
}

// onInstances checks rules on each instance of a template, by check on a checker of its own. The
// author mends a rule broken at a place there, so the report holds the breaches at a place of the
// first instance that breaks the rule there, each naming the instance's numbers (format/README.md,
// section 23): 10 000 instances whose trap value is their key give one breach, not 10 000. Another
// instance that breaks the rule at the same place otherwise — a text that holds the key of one
// instance and the key of another — shows once the place is mended. A rule of a task reads an
// instance only when it is valid against the schema of a task, as it reads a file that passed its
// schema.
func (c *checker) onInstances(e *expansion, check func(sub *checker, i instance)) {
	first := map[string]int{} // the instance that first breaks each rule at each place
	for n, i := range e.instances {
		sub := &checker{}
		check(sub, i)
		for _, b := range sub.breaches {
			place := b.Rule + " " + b.Path + " " + b.Pointer
			if k, ok := first[place]; ok && k != n {
				continue
			}
			first[place] = n
			b.Message = "on the instance " + i.shown + ": " + b.Message
			c.breaches = append(c.breaches, b)
		}
	}
}
