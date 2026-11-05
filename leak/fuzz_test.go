package leak

import (
	"math/big"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"testing"
)

// runForm is the longest run of the form of a number that begins at a digit, written another way than
// longestRun: a simple fraction whose denominator holds a digit other than 0, a decimal, or an
// integer, the longest match of the three (format/README.md, section 20).
var runForm = func() *regexp.Regexp {
	r := regexp.MustCompile(`^(?:[0-9]+ ?/ ?[0-9]*[1-9][0-9]*|[0-9]+\.[0-9]+|[0-9]+)`)
	r.Longest()
	return r
}()

// modelNumbers reads the numbers of a text in its normal form by runForm, for FuzzNumbers to hold
// numbersIn to: each as the rational it writes.
func modelNumbers(form string) []string {
	var found []string
	runes := []rune(form)
	for at := 0; at < len(runes); {
		if !isDigit(runes[at]) {
			at++
			continue
		}
		run := runForm.FindString(string(runes[at:]))
		end := at + len(run)
		if !wordAt(runes, at-1) && !wordAt(runes, end) {
			found = append(found, modelValue(run))
		}
		at = end
	}
	return found
}

// modelValue gives the rational that a run of the form of a number writes, leading zeros and all.
func modelValue(run string) string {
	integer := func(digits string) *big.Int {
		n, _ := new(big.Int).SetString(strings.Trim(digits, " "), 10)
		return n
	}
	if numerator, denominator, ok := strings.Cut(run, "/"); ok {
		return new(big.Rat).SetFrac(integer(numerator), integer(denominator)).RatString()
	}
	if whole, part, ok := strings.Cut(run, "."); ok {
		places := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(len(part))), nil)
		return new(big.Rat).SetFrac(integer(whole+part), places).RatString()
	}
	return integer(run).String()
}

// FuzzNumbers writes a number in each form that the search reads — an integer, a decimal and a
// simple fraction with a space on either side of its slash or none, with leading zeros — and finds
// it between characters of no word, exactly and alone; with a character of a word right before or
// right after it — a letter, a mark, a digit of another script — the search finds nothing, not even
// a tail of it. In any text, the numbers are those that a regular expression of
// the longest run reads, modelNumbers, and the search goes without a panic (format/README.md,
// section 20).
func FuzzNumbers(f *testing.F) {
	f.Add(uint64(7), uint64(5), uint8(0), uint8(0), "x7.5")
	f.Add(uint64(9007199254740993), uint64(0), uint8(2), uint8(1), "07:30")
	f.Add(uint64(3), uint64(3), uint8(1), uint8(2), "6/8 and 1/0")
	f.Add(uint64(1), uint64(1), uint8(3), uint8(11), "1 1/2, 1\u00BD, 7\u00B2")
	search, err := For(taskOf(f, sheet))
	if err != nil {
		f.Fatal(err)
	}
	f.Fuzz(func(t *testing.T, n, d uint64, zeros, shape uint8, text string) {
		lead := strings.Repeat("0", int(zeros%4))
		whole := lead + strconv.FormatUint(n, 10)
		numerator := new(big.Int).SetUint64(n)
		var run string
		var value *big.Rat
		switch shape % 3 {
		case 0:
			run, value = whole, new(big.Rat).SetInt(numerator)
		case 1:
			digits := lead + strconv.FormatUint(d, 10)
			places := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(len(digits))), nil)
			run = whole + "." + digits
			value = new(big.Rat).Add(new(big.Rat).SetInt(numerator), new(big.Rat).SetFrac(new(big.Int).SetUint64(d), places))
		case 2:
			denominator := d%(1<<62) + 1
			before, after := "", ""
			if shape/3%2 == 1 {
				before = " "
			}
			if shape/6%2 == 1 {
				after = " "
			}
			run = whole + before + "/" + after + lead + strconv.FormatUint(denominator, 10)
			value = new(big.Rat).SetFrac(numerator, new(big.Int).SetUint64(denominator))
		}
		for _, around := range [][2]string{{"", ""}, {"(", ")"}, {"= ", "!"}, {"-", ", x"}, {"a ", " b"}, {"\uFFFD", ". "}} {
			text := around[0] + run + around[1]
			if got := numbersIn([]rune(text)); len(got) != 1 || got[0].Cmp(value) != 0 {
				t.Fatalf("the numbers of %q are %v, want %s alone", text, got, value.RatString())
			}
		}
		for _, around := range [][2]string{{"x", ""}, {"", "x"}, {"\u00E9", ""}, {"", "\u0301"}, {"\u0661", ""}, {"", "\u0661"}} {
			text := around[0] + run + around[1]
			if got := numbersIn([]rune(text)); len(got) != 0 {
				t.Fatalf("the numbers of %q are %v, want none: a character of a word touches the run", text, got)
			}
		}
		normal := form(text)
		var got []string
		for _, v := range numbersIn([]rune(normal)) {
			got = append(got, v.RatString())
		}
		if want := modelNumbers(normal); !slices.Equal(got, want) {
			t.Fatalf("the numbers of %q, %q in its normal form, are %q, and the model reads %q", text, normal, got, want)
		}
		_ = search.In(text)
	})
}
