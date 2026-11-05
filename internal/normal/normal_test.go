package normal

import (
	"strings"
	"testing"
)

// The steps of the normalization, each on its own.
func TestSteps(t *testing.T) {
	for _, c := range []struct {
		step func(string) string
		in   string
		want string
	}{
		{Scripts, "7²", "7�"},
		{Scripts, "cm²", "cm�"},
		{Scripts, "1½", "1�"},
		{Scripts, "½1", "�1"},
		{Scripts, "½", "½"},
		{Scripts, "a½", "a½"},
		{Scripts, "½½", "��"},
		{Scripts, "²½", "��"},
		{Scripts, "１⅟", "１�"},
		{Text, "  a\u00A0\u2003 b  ", "a b"},
		{Text, "\u2212\u2010\u2012\u2013\u2014\u2015", "------"},
		{Text, "\u2011", "-"},
		{Text, "1\u20442 1\u22152", "1/2 1/2"},
		{Text, "½", "1/2"},
		{Text, "a\uFEFF", "a\uFEFF"},
		{Lower, "İΣA", "iσa"},
		{upToStep5, "= 7 rungs.", " 7 rungs"},
		{upToStep5, "7 .", "7 "},
		{Choice, " ＜ ", "<"},
		{Choice, "= ODD.", "= odd."},
		{Unit, " Min. ", "min"},
		{func(s string) string { return Number(s, []string{Unit("rungs")}) }, "= 7 Rungs.", "7"},
		{func(s string) string { return Number(s, []string{Unit("rungs")}) }, "rungs", "rungs"},
	} {
		if got := c.step(c.in); got != c.want {
			t.Errorf("a step of %q gives %q, want %q", c.in, got, c.want)
		}
	}
}

// NFKC here follows the Stream-Safe Text Process of UAX #15 too: after 30 non-starters in a row it
// puts U+034F in, which spec's second implementation does not (R27).
func TestNFKCIsStreamSafe(t *testing.T) {
	if got := Text("a" + strings.Repeat("\u0301", 31)); !strings.ContainsRune(got, '\u034F') {
		t.Errorf("31 combining marks normalize to %+q, with no U+034F", got)
	}
}
