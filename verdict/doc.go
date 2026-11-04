// Package verdict gives the verdict of an answer slot and of an item on the values read in them,
// by the check kinds of the module format and their normalization, with the reason of a value
// that cannot be read and the trap of a wrong one (format/README.md, section 15). It also joins the
// observations of two readers into one, answer slot by answer slot, and gives the verdict on the
// joined observation (contracts/README.md, section 10; R86 spec). api calls it on a learner's
// work, and tutor's eval runner on the truth of a sample's markup (R95 spec).
//
// It holds no code yet: the epic of code brings it (R16).
package verdict
