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
// (R22); and each file of a part against the schema of its part, by internal/schema, with a breach
// of a schema named by its keyword and place as section 8 names it (R23).
package module
