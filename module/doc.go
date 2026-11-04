// Package module reads an author's module and checks it as the module format says: its layout and
// the version of the format it names, each file against the schema of its part, and the format
// rules of each part, between parts and of the whole module — the tiers of the skill graph, the
// templates with every instance and solver, and the leak check of the texts a learner sees before
// solving among them (format/README.md, sections 2 and 4–24). A breach names the ID of its rule
// and its place. The CLI's check and build call it, and module-registry from stage 4
// (architecture 3.6).
//
// It holds no code yet: the epic of code brings it part by part (R16).
package module
