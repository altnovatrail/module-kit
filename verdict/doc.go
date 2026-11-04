// Package verdict gives the verdict of an answer slot and of an item on the values read in them,
// by the check kinds of the module format and their normalization, with the reason of a value
// that cannot be read and the trap of a wrong one (format/README.md, section 15). It also joins the
// observations of two readers into one, answer slot by answer slot, and gives the verdict on the
// joined observation (contracts/README.md, section 10; R86 spec). api calls it on a learner's
// work, and tutor's eval runner on the truth of a sample's markup (R95 spec); the check of a
// module reads by it the rules of a task that the check of an answer slot sees itself.
//
// It holds so far the check kinds: OfItem and OfSlot give the verdict on values as they are
// written, Same compares two values as a check does, and UnicodeVersion names the version of
// Unicode that the normalization follows (R27). A predicate runs in the sandbox of Starlark, and
// CheckPredicate holds the rule of its expression (R28). A rubric gives no verdict. The joining of
// two readings comes with its own task (R16).
package verdict
