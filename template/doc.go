// Package template fills the templates of tasks: it counts the instances of a template, builds
// each one with its numbers and the results of its solver, and picks the instance of an attempt
// from the attempt's seed (format/README.md, sections 22–24; R60 spec, R61 spec). api calls it for
// the instance of an attempt, and the check of a module for every instance.
//
// AttemptInstance picks the instance of an attempt by its number, from the place of randomness of
// the attempt's seed that the format gives a template, through internal/random (R10). The
// instances themselves and the results of a solver come with the tasks of templates and of
// solvers (R16).
package template
