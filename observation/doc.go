// Package observation holds what module-kit does with an observation, novatrail.observation.v1:
// which one is valid, how the observations of two readers join into one, answer slot by answer
// slot, which verdict the observation gives with the options tapped beside it, and how the
// submission of an essay ends (contracts/README.md, section 10; R86 spec; R29). The verdict of each
// value is the package verdict's, by the check kinds of the module format (format/README.md,
// section 15). The observation's type is spec's, of the module github.com/altnovatrail/spec/contracts,
// so this package brings in the protobuf runtime, and verdict and the check of a module do not.
//
// api and tutor's eval runner call it:
//   - api joins the two readings, Join, and gives the verdict on the joined observation, Verdicts,
//     or on none, nil, when either is not valid or the call of AI gave none. It keeps the joined
//     observation and the options tapped, and the same release gives the same verdict on them
//     again, with no AI (R03 spec);
//   - the eval runner gives the verdict on the truth of a sample's markup, and on a reading of one
//     reader, by Validate and Verdicts, and on the readings of a pair by Join and Verdicts (R95 spec);
//   - an essay has one reader and no verdict: Essay says whether its submission goes on to the
//     assessment by the rubric, ends unreadable or ends unchecked.
//
// An observation that is not valid is ErrNotValid, from Validate, Join, Verdicts and Essay alike,
// and it is no observation: every answer slot that AI reads gets unchecked by it. So unchecked in
// the verdict on a valid observation is the check's own, a predicate that could not run (R28).
package observation
