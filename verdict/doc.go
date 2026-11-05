// Package verdict gives the verdict of an answer slot and of an item on the values read in them,
// by the check kinds of the module format and their normalization, with the reason of a value
// that cannot be read and the trap of a wrong one (format/README.md, section 15). api and tutor's
// eval runner call it through the package observation, which joins the observations of two
// readers and gives the verdict on the joined one (contracts/README.md, section 10; R29); the
// check of a module reads by it the rules of a task that the check of an answer slot sees itself.
//
// OfItem and OfSlot give the verdict on values as they are written, Same compares two values as a
// check does, NormalForm gives a value as the kind normalizes it, ItemOf makes an item's verdict of
// its answer slots', and UnicodeVersion names the version of Unicode that the normalization follows
// (R27). A predicate runs in the sandbox of Starlark, and CheckPredicate holds the rule of its
// expression (R28). A rubric gives no verdict. The package imports no protobuf: the observation is
// the package observation's.
package verdict
