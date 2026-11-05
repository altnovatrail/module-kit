package modulepackage

import (
	"context"
	"fmt"
	"io/fs"
	"maps"
	"math"
	"path"
	"slices"
	"strings"

	"github.com/altnovatrail/spec/types/formatv0"

	"github.com/altnovatrail/module-kit/internal/canonical"
	"github.com/altnovatrail/module-kit/internal/schema"
	"github.com/altnovatrail/module-kit/internal/yaml"
	"github.com/altnovatrail/module-kit/module"
	"github.com/altnovatrail/module-kit/template"
)

// The members of the canonical view that the code names, and the field of the format version
// (format/README.md, sections 2 and 25).
const (
	versionField     = "module_format_version"
	manifestMember   = "manifest"
	skillGraphMember = "skill_graph"
	skillTiersMember = "skill_tiers"
	tasksMember      = "tasks"
	instancesMember  = "instances"
)

// The member of the canonical view of each file of a part that a module has once, and of each
// folder of a part that it has many of (format/README.md, section 25).
var (
	onceParts = map[string]string{
		"manifest.yaml":      manifestMember,
		"skill-graph.yaml":   skillGraphMember,
		"traps.yaml":         "traps",
		"techniques.yaml":    "techniques",
		"solving-steps.yaml": "solving_steps",
	}
	manyParts = map[string]string{"tasks/": tasksMember, "lessons/": "lessons"}
)

// canonicalView builds the canonical view of a module that passed its check: each part under its
// member, as the JSON value of its file is, with an integer that the reader of YAML gave as a
// double as an integer; skill_tiers, when it has a skill graph; and instances, when a task is a
// template, with the results of its solver (format/README.md, sections 12, 23, 24 and 25).
func canonicalView(ctx context.Context, fsys fs.FS, files []string) (map[string]any, error) {
	view := map[string]any{}
	for _, file := range files {
		member, id, ok := memberOf(file)
		if !ok {
			continue
		}
		data, err := fs.ReadFile(fsys, file)
		if err != nil {
			return nil, err
		}
		value, err := yaml.Read(data)
		if err != nil {
			return nil, fmt.Errorf("modulepackage: %s, which passed the check: %w", file, err)
		}
		value = integers(value)
		if id == "" {
			view[member] = value
			continue
		}
		entries, _ := view[member].(map[string]any)
		if entries == nil {
			entries = map[string]any{}
			view[member] = entries
		}
		entries[id] = value
	}
	if graph, ok := view[skillGraphMember]; ok {
		tiers, err := module.Tiers(decode[formatv0.SkillGraph](graph))
		if err != nil {
			return nil, fmt.Errorf("modulepackage: the tiers of a skill graph that passed the check: %w", err)
		}
		view[skillTiersMember] = schema.ValueOf(tiers)
	}
	instances, err := instancesOf(ctx, view)
	if err != nil {
		return nil, err
	}
	if len(instances) > 0 {
		view[instancesMember] = instances
	}
	return view, nil
}

// memberOf gives the member of the canonical view that a file of a part stands under, with the local
// ID of its entry for a part that a module has many of, and false for a file of no part.
func memberOf(file string) (member, id string, ok bool) {
	if member, ok := onceParts[file]; ok {
		return member, "", true
	}
	folder, name := path.Split(file)
	if member, ok := manyParts[folder]; ok {
		return member, strings.TrimSuffix(name, ".yaml"), true
	}
	return "", "", false
}

// instancesOf counts the instances of each template among the tasks, in the order of their IDs,
// running its solver with ctx, and gives their numbers by the ID of the template
// (format/README.md, sections 23 and 24). A template of a module that passed its check breaks no
// rule of templates, so a breach is a fault of module-kit.
func instancesOf(ctx context.Context, view map[string]any) (map[string]any, error) {
	tasks, _ := view[tasksMember].(map[string]any)
	instances := map[string]any{}
	for _, id := range slices.Sorted(maps.Keys(tasks)) {
		task := decode[formatv0.Task](tasks[id])
		if task.Template == nil {
			continue
		}
		numbers, breaches, err := template.Instances(ctx, task)
		if err != nil {
			return nil, err
		}
		if len(breaches) > 0 {
			return nil, fmt.Errorf("modulepackage: the template %s, which passed the check, breaks %s: %s", id, breaches[0].Rule, breaches[0].Message)
		}
		instances[id] = schema.ValueOf(numbers)
	}
	return instances, nil
}

// integers gives a JSON value of the reader of YAML with every double that is an integer within
// ±(2^53 − 1) — 50.0 or 5e1 — as that integer, as the canonical view holds it (format/README.md,
// section 25). Any other double stays, and the canonical text refuses it.
func integers(value any) any {
	switch v := value.(type) {
	case map[string]any:
		for key, member := range v {
			v[key] = integers(member)
		}
	case []any:
		for i, element := range v {
			v[i] = integers(element)
		}
	case float64:
		if v == math.Trunc(v) && math.Abs(v) <= canonical.MaxInteger {
			return int64(v)
		}
	}
	return value
}
