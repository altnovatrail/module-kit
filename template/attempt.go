package template

import (
	"fmt"

	"github.com/altnovatrail/spec/types/formatv0"

	"github.com/altnovatrail/module-kit/internal/random"
)

// MaxInstances is the most instances a template has: the sizes of its domains multiply to at most
// 10 000 (format/README.md, section 23).
const MaxInstances = 10000

// ErrDomain is what AttemptInstance gives on an argument outside its domain: a count outside 1 to
// MaxInstances, or an ID that is no name of a place of randomness. It is the generator's own, so
// that errors.Is finds it whichever of the two gives it.
var ErrDomain = random.ErrDomain

// The names of the place of randomness of a template, around the ID of its task: its first name
// says what the place is of (format/README.md, section 22).
const (
	placeOfTemplate = "template"
	placeOfInstance = "instance"
)

// AttemptInstance gives the number of the instance that an attempt on a template gets, from 0:
// below(count) of the place of randomness of the attempt's seed whose names are template, the ID
// of the task and instance, where count is how many instances the template has, from 1 to
// MaxInstances (format/README.md, sections 22 and 23; R60 spec). Nothing else draws from that
// place, so the same package and the same seed give the same instance (R03 spec). The instances
// are numbered in the order of the odometer, as the package carries their numbers.
//
// api draws the seed from a cryptographic source when the attempt begins and passes it here as a
// number; neither the seed nor the words of the place reach the learner's device, since the words
// show the seed (format/README.md, section 22). A count outside its domain gives ErrDomain, and so
// does an ID that is no name of a place: empty, or holding | or what is not UTF-8. The ID of a
// valid task is neither (common.schema.json).
func AttemptInstance(seed uint64, task formatv0.TaskID, count int) (int, error) {
	if count < 1 || count > MaxInstances {
		return 0, fmt.Errorf("template: %d instances: a template has from 1 to %d: %w", count, MaxInstances, ErrDomain)
	}
	place, err := random.Place(seed, placeOfTemplate, string(task), placeOfInstance)
	if err != nil {
		return 0, fmt.Errorf("template: the task %q: %w", task, err)
	}
	n, _ := place.Below(uint64(count)) // the count is from 1, in the domain of below
	return int(n), nil
}
