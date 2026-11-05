// Package module reads an author's module and checks it as the module format says: its layout and
// the version of the format it names, each file against the schema of its part, and the format
// rules of each part, between parts and of the whole module — the tiers of the skill graph, the
// templates with every instance and solver, and the leak check of the texts a learner sees before
// solving among them (format/README.md, sections 2 and 4–24). A breach names the ID of its rule
// and its place. The CLI's check and build call it, and module-registry from stage 4
// (architecture 3.6).
//
// The epic of code brings it part by part (R16). It holds so far the folder of a module as an
// fs.FS, by the layout of section 10 and the rules module.*; the format version of section 2 and
// manifest.format-version; the YAML rules of section 4 on each file of a part, by internal/yaml
// (R22); each file of a part against the schema of its part, by internal/schema, with a breach of a
// schema named by its keyword and place as section 8 names it (R23); the rules of the manifest, the
// skill graph and the catalogs of sections 11, 12, 16 and 17, read on the types of formatv0, with
// the tiers of the skill graph, Tiers, which the package will hold (R25); and the rules of a task
// and of a lesson of sections 14, 15 and 21 that need no check kind, no leak check and no template
// (R26); the rules of a task that the check of an answer slot sees itself, by package verdict
// (R27, R28); task.no-leaks and lesson.no-leaks, by package leak (R30); and the rules of templates of
// section 23, by package template, with the rules of a task on every instance of a template that
// read its strings, and task.instances, which holds each instance to the schema of a task (R32).
package module
