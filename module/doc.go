// Package module reads an author's module and checks it as the module format says: its layout and
// the format version it names, each file by the YAML rules and against the schema of its part, and
// the format rules within each part, between parts and of the whole module — the tiers of the skill
// graph, the templates with every instance and solver, the leak check of the texts a learner sees
// before solving, and the page of a workbook. A breach names the ID of its rule and its place, and a
// warning of a report refuses nothing (R50). The CLI's check and build call it, and module-registry
// from stage 4.
//
// Check runs the solvers of a module's templates one on each CPU, with the caller's context, and
// CheckModule gives the numbers of their instances to the build of a package, so that each solver
// runs once (R38).
package module
