// Package template fills the templates of tasks: it counts the instances of a template, runs its
// solver on them, builds each one with its numbers, and picks the instance of an attempt from the
// attempt's seed (format/README.md, sections 22–24; R60 spec, R61 spec). api calls it for the instance
// of an attempt, and the check of a module and the build of a package for every instance.
//
// Instances checks a template by the rules of templates that it breaks by itself — task.template-names,
// task.template-expressions, task.placeholders, task.template-instances and task.template-solver —
// and counts its instances in the order of the odometer, each with its numbers: the expressions of the
// narrow subset of Starlark, read and typed here and evaluated in the sandbox of internal/sandbox
// (R32); and the results of its solver, a program of Starlark run once on a thread of its own, whose
// solve is called on every instance in turn with the numbers that Inputs names, within the budget of
// the template (R33). Instance puts the numbers of an instance in the place of the placeholders, in a
// single pass, with no expression and no Starlark, as api does with the numbers that the package
// carries; the open view of an instance is openview.Task of it. AttemptInstance picks the instance of
// an attempt by its number, from the place of randomness of the attempt's seed that the format gives a
// template, through internal/random (R31).
package template
