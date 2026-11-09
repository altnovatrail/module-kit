package normal

import (
	"fmt"
	"regexp"
	"strconv"
)

// Time gives a value as a time reads it: steps 1 to 5 of a number, with the spaces at either end
// dropped, so that "07:30." is "07:30"; "" when it is blank (format/README.md, section 15).
func Time(value string) string {
	return trimmed(upToStep5(value))
}

// timeForm is the notation of a time read: one or two digits of hours, :, and the minutes, two
// digits, or three that begin with no zero, so that 10:100 is in it and 7:030 is not.
var timeForm = regexp.MustCompile(`^([0-9]{1,2}):([0-9]{2}|[1-9][0-9]{2})$`)

// Clock is a time in the notation of a time read, by its hours and its minutes, each as a number:
// 07:30 and 7:30 are one clock, and 5:60 is not 6:00, since the sum of minutes is never taken.
type Clock struct {
	Hours, Minutes int
}

func (c Clock) String() string {
	return fmt.Sprintf("%d:%02d", c.Hours, c.Minutes)
}

// TimeNotation reads a text that Time normalized as a time, and says whether it is in the notation
// of a time read (format/README.md, section 15): 7:3, 007:30, 7.30, 7 : 30 and 7:30 pm are not.
func TimeNotation(text string) (Clock, bool) {
	m := timeForm.FindStringSubmatch(text)
	if m == nil {
		return Clock{}, false
	}
	hours, _ := strconv.Atoi(m[1])
	minutes, _ := strconv.Atoi(m[2])
	return Clock{hours, minutes}, true
}
