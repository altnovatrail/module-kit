// Package openview builds the open view of a part of a module and of an instance of a template —
// what the learner's app receives: the part without its closed fields — from the marks x-open in
// the schemas of the format, and the digest of the open view of a task and of an instance
// (format/README.md, section 6; R05 spec; R93 spec). api, tutor's eval runner and the CLI's open
// call it.
//
// Manifest, SkillGraph, Traps, Techniques, SolvingSteps, Lesson and Task take a part as the type of
// formatv0 holds it and give its open view as the type of formatv0/open holds it (R72 spec). The
// walk of the schemas is internal/schema's: a mark counts only where its subschema applies and
// passes, and no list of fields is kept here (R45 spec). The open view of an instance of a template
// is the open view of a task: Task of what template.Instance gives (format/README.md, section 23;
// R32). The digest comes with the task of the package (R16).
package openview
