package template

import "testing"

// typeOf reads the subset of section 23 and types it, and says why an expression is none of it.
func TestTypeOf(t *testing.T) {
	names := []string{"a", "b", "ab_1"}
	for _, c := range []struct {
		src  string
		want exprType
		err  string
	}{
		{"0", intType, ""},
		{"a", intType, ""},
		{"2 * a + 4 * b", intType, ""},
		{"-a // 2", intType, ""},
		{"-(a // 2)", intType, ""},
		{"- - a", intType, ""},
		{"a%-4", intType, ""},
		{"ab_1 * (a - b) % 7", intType, ""},
		{"99999999999999999999999 * a", intType, ""},
		{"a != b", boolType, ""},
		{"a<=b", boolType, ""},
		{"not a == b", boolType, ""},
		{"not not a > b", boolType, ""},
		{"b != 0 and a % b == 0", boolType, ""},
		{"a < 1 or a > 8 and b == 2", boolType, ""},
		{"(a < b) or (b < a)", boolType, ""},
		{"", "", "it ends where an operand is due"},
		{"a + b   ", intType, ""},
		{"   ", "", "it ends where an operand is due"},
		{" a + b", intType, ""},
		{"a +", "", "it ends where an operand is due"},
		{"a / 2", "", "'/' at 3 is no part of the expressions of templates"},
		{"a = 2", "", "'=' at 3 is no part of the expressions of templates"},
		{"a\t+ 1", "", "'\\t' at 2 is no part of the expressions of templates"},
		{"a +\n1", "", "'\\n' at 4 is no part of the expressions of templates"},
		{"abs(a)", "", "abs at 1 is none of the names that it may name: a, b, ab_1"},
		{"a ** 2", "", "* at 4 stands where an operand is due"},
		{"a × 2", "", "'×' at 3 is no part of the expressions of templates"},
		{"True", "", "'T' at 1 is no part of the expressions of templates"},
		{"0x10", "", "0x10 at 1 is neither an integer nor a name"},
		{"1e5", "", "1e5 at 1 is neither an integer nor a name"},
		{"2a", "", "2a at 1 is neither an integer nor a name"},
		{"007", "", "007 at 1: an integer has no leading zero"},
		{"c + 1", "", "c at 1 is none of the names that it may name: a, b, ab_1"},
		{"if", "", "if at 1 is none of the names that it may name: a, b, ab_1"},
		{"a in b", "", "in at 3 stands where the expression has ended"},
		{"a b", "", "b at 3 stands where the expression has ended"},
		{"(a + 1", "", "the ( at 1 is not closed"},
		{"a + 1)", "", ") at 6 stands where the expression has ended"},
		{"()", "", ") at 2 stands where an operand is due"},
		{"0 < a < 9", "", "< at 3 and < at 7: comparisons do not chain"},
		{"a == b == c", "", "== at 3 and == at 8: comparisons do not chain"},
		{"a + (b < 2)", "", "+ at 3 takes an int, and its operand is a bool"},
		{"-(a < 2)", "", "- at 1 takes an int, and its operand is a bool"},
		{"(a < 2) < b", "", "< at 9 takes an int, and its operand is a bool"},
		{"a and b", "", "and at 3 takes a bool, and its operand is an int"},
		{"a < b or 1", "", "or at 7 takes a bool, and its operand is an int"},
		{"not a", "", "not at 1 takes a bool, and its operand is an int"},
	} {
		got, err := typeOf(c.src, names)
		if c.err == "" && (err != nil || got != c.want) || c.err != "" && (err == nil || err.Error() != c.err || got != "") {
			t.Errorf("%q: %q, %v; want %q, %q", c.src, got, err, c.want, c.err)
		}
	}
}
