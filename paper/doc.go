// Package paper is the key of a task from the learner's paper: the solver, which computes the
// answer of arithmetic and of an equation in one unknown from the formula AI read of the
// statement, exactly and without Starlark or the network; the rules an answer of AI holds before
// it is a key; and the item and the values by which the package observation judges the reading of
// the work against the key, as on a task of a module.
//
// Solve gives the solver's answer or why it does not take the formula, Breach the rule an answer
// of AI breaks, Key the key of an attempt with its provenance, Items the task's one item by the
// answer of its key, and Brought and BroughtValue the values read brought to it. The formula, the
// solver, the answer, the key and the verdict are those of spec's contracts, whose
// test vectors the package passes.
package paper
