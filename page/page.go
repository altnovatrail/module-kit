package page

import (
	_ "embed"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"unicode/utf8"

	"github.com/altnovatrail/spec/types/formatv0"

	"github.com/altnovatrail/module-kit/internal/schema"
	"github.com/altnovatrail/module-kit/internal/yaml"
)

// pageFile is format/page.yaml of spec.Format at the tag go.mod requires, byte for byte
// (TestPageIsSpecs).
//
//go:embed page.yaml
var pageFile []byte

var format = sync.OnceValue(func() *table { return mustLoad(pageFile) })

// Numbers gives the numbers of the layout, numbers of page.yaml, in units of the page: the type
// size, the width and the height of a box, and the rest.
func Numbers() formatv0.PageNumbers {
	return format().page.Numbers
}

// MostItems gives the most items that a page holds, most_items of page.yaml: the item after them has
// no room.
func MostItems() int {
	return int(format().page.MostItems)
}

// TextWidth gives the width, in units of the page, of a text as the layout sets it on one line, as
// one piece of a line: its words, cut at the runs of the white space that breaks, joined by single
// spaces, size × (the sum of their widths + the space × the spaces between them) ÷ 100, rounded up.
// A text of white space alone, or none, is 0 wide.
func TextWidth(text string) int64 {
	return format().textWidth(text)
}

// Fits says whether a value written in a box fits the box: as a text of the line, it is no wider
// than the numbers' BoxWidth, as lesson.workbook-page holds the value of each box of a worked
// example.
func Fits(value string) bool {
	return TextWidth(value) <= Numbers().BoxWidth
}

// mustLoad reads the table of a file of page.yaml. A file that does not read is a fault in
// module-kit, since the copy is spec's, and mustLoad panics on it, as regexp.MustCompile does.
func mustLoad(data []byte) *table {
	t, err := load(data)
	if err != nil {
		panic(err)
	}
	return t
}

// load reads a file of page.yaml as the format has a reader do it: by the YAML rules, against its
// schema, page.schema.json, and only then into its type of formatv0, strictly; and gives its table.
func load(data []byte) (*table, error) {
	value, err := yaml.Read(data)
	if err != nil {
		return nil, fmt.Errorf("page: page.yaml: %w", err)
	}
	if errs := schema.Validate("page.schema.json", value); len(errs) > 0 {
		texts := make([]string, len(errs))
		for i, e := range errs {
			texts[i] = e.String()
		}
		return nil, fmt.Errorf("page: page.yaml is refused by page.schema.json: %s", strings.Join(texts, "; "))
	}
	p, err := schema.Decode[formatv0.Page](value)
	if err != nil {
		return nil, fmt.Errorf("page: page.yaml: %w", err)
	}
	return newTable(p)
}

// A table is a value of page.yaml as the layout reads it: which code point is of which width, from
// the table of widths alone, never from a Unicode database, and the measures that the numbers give.
// It holds nothing that changes, so one table lays out pages from many goroutines at once.
type table struct {
	page     formatv0.Page
	digits   map[rune]bool
	signs    map[rune]bool
	own      map[rune]int64 // a character or a space of a width of its own
	zero     []span
	wide     []span
	breaks   map[rune]bool
	joiner   rune
	regional span
	measures
}

type span struct{ first, last rune }

// The measures are in units of the page.
type measures struct {
	space   int64 // between words of two pieces: the space's own width, rounded up
	numbers int64 // the column of the numbers, as wide as 00.
	indent  int64 // where a line starts after the left of its column: the numbers and the gap
	column  int64 // a column of short items
	second  int64 // where the second column starts
	across  int64 // a line across the page
	lowest  int64 // the lowest unit that a row may reach
}

// errCodePoint is what a code point that page.yaml does not write as U+ and hexadecimal digits
// gives; its schema lets none through.
var errCodePoint = errors.New("page: page.yaml writes a code point not as U+ and hexadecimal digits")

func newTable(p formatv0.Page) (*table, error) {
	t := &table{page: p, digits: set(p.Widths.Digits.Characters), signs: set(p.Widths.Signs.Characters),
		own: map[rune]int64{}, breaks: map[rune]bool{}}
	for _, class := range p.Widths.Characters {
		for _, r := range class.Characters {
			t.own[r] = class.Width
		}
	}
	for _, space := range p.Widths.Spaces {
		r, err := codePoint(space.CodePoint)
		if err != nil {
			return nil, err
		}
		t.own[r] = space.Width
	}
	var err error
	if t.zero, err = spans(p.Widths.Zero); err != nil {
		return nil, err
	}
	if t.wide, err = spans(p.Widths.Wide.Ranges); err != nil {
		return nil, err
	}
	for _, text := range p.Widths.Breaks {
		r, err := codePoint(text)
		if err != nil {
			return nil, err
		}
		t.breaks[r] = true
	}
	if t.joiner, err = codePoint(p.Widths.Joiner); err != nil {
		return nil, err
	}
	if t.regional, err = spanOf(p.Widths.RegionalIndicators); err != nil {
		return nil, err
	}
	n, sheet := p.Numbers, p.Sheet
	t.space = t.units(p.Widths.Space)
	t.numbers = t.units(t.wordWidth("00."))
	t.indent = t.numbers + n.Gap
	content := sheet.Width - 2*n.Margin
	t.column = floorDiv(content-n.Gutter, 2)
	t.second = n.Margin + t.column + n.Gutter
	t.across = content - t.indent
	t.lowest = sheet.Height - n.Margin
	return t, nil
}

func set(text string) map[rune]bool {
	s := map[rune]bool{}
	for _, r := range text {
		s[r] = true
	}
	return s
}

// codePoint reads a code point as page.yaml writes it: U+00A0.
func codePoint(text string) (rune, error) {
	digits, ok := strings.CutPrefix(text, "U+")
	if !ok {
		return 0, fmt.Errorf("%w: %q", errCodePoint, text)
	}
	n, err := strconv.ParseUint(digits, 16, 21)
	if err != nil {
		return 0, fmt.Errorf("%w: %q", errCodePoint, text)
	}
	return rune(n), nil
}

// spanOf reads a code point, or a range of them, as page.yaml writes it: U+0300–U+036F, with an en
// dash between the two.
func spanOf(text string) (span, error) {
	first, last, found := strings.Cut(text, "–")
	if !found {
		last = first
	}
	from, err := codePoint(first)
	if err != nil {
		return span{}, err
	}
	to, err := codePoint(last)
	if err != nil {
		return span{}, err
	}
	return span{from, to}, nil
}

func spans(texts []string) ([]span, error) {
	found := make([]span, len(texts))
	for i, text := range texts {
		s, err := spanOf(text)
		if err != nil {
			return nil, err
		}
		found[i] = s
	}
	return found, nil
}

func (s span) holds(r rune) bool { return r >= s.first && r <= s.last }

func within(r rune, spans []span) bool {
	for _, s := range spans {
		if s.holds(r) {
			return true
		}
	}
	return false
}

// floorDiv rounds toward minus infinity, as the layout rounds a half and a third: Go's / truncates.
func floorDiv(a, b int64) int64 {
	q := a / b
	if a%b != 0 && (a < 0) != (b < 0) {
		q--
	}
	return q
}

// width is in hundredths of the type size.
func (t *table) width(r rune) int64 {
	switch {
	case within(r, t.zero):
		return 0
	case t.digits[r]:
		return t.page.Widths.Digits.Width
	case t.signs[r]:
		return t.page.Widths.Signs.Width
	}
	if w, ok := t.own[r]; ok {
		return w
	}
	if within(r, t.wide) {
		return t.page.Widths.Wide.Width
	}
	return t.page.Widths.Unknown
}

func (t *table) words(text string) []string {
	return strings.FieldsFunc(text, func(r rune) bool { return t.breaks[r] })
}

// clusters cuts a word into its clusters of code points, in order: a code point of no width, one
// after the joiner, and the second of two regional indicators join the cluster before them.
func (t *table) clusters(word string) []string {
	var clusters []string
	joined := false
	regional := 0 // 1 after a regional indicator that begins a cluster, 2 once its flag is whole
	for _, r := range word {
		flag := t.regional.holds(r)
		if last := len(clusters) - 1; last >= 0 && (joined || within(r, t.zero) || flag && regional == 1) {
			clusters[last] += string(r)
			if flag && regional == 1 {
				regional = 2
			}
		} else {
			clusters = append(clusters, string(r))
			regional = 0
			if flag {
				regional = 1
			}
		}
		joined = r == t.joiner
	}
	return clusters
}

func (t *table) clusterWidth(cluster string) int64 {
	r, _ := utf8.DecodeRuneInString(cluster)
	return t.width(r)
}

// wordWidth is in hundredths of the type size.
func (t *table) wordWidth(word string) int64 {
	var sum int64
	for _, cluster := range t.clusters(word) {
		sum += t.clusterWidth(cluster)
	}
	return sum
}

func (t *table) units(hundredths int64) int64 {
	product := t.page.Numbers.Size * hundredths
	units := product / 100
	if product%100 > 0 {
		units++
	}
	return units
}

func (t *table) textWidth(text string) int64 {
	words := t.words(text)
	var sum int64
	for _, word := range words {
		sum += t.wordWidth(word)
	}
	return t.units(sum + t.page.Widths.Space*int64(max(len(words)-1, 0)))
}
