package page

import (
	"bytes"
	"errors"
	"io/fs"
	"slices"
	"strings"
	"testing"

	"github.com/altnovatrail/spec"
	"github.com/altnovatrail/spec/types/formatv0"
)

// The copy of page.yaml is format/page.yaml of spec.Format: the same file, byte for byte, of the
// version that spec.Format holds. A move to a new tag of spec that changes it fails it until the
// copy follows, as R20 has the move take the copy along.
func TestPageIsSpecs(t *testing.T) {
	want, err := fs.ReadFile(spec.Format, "page.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(pageFile, want) {
		t.Error("page.yaml of the copy is not that of spec.Format: copy it again from the module cache")
	}
}

// The measures of the numbers of page.yaml are those that section 14 of format/README.md names: the
// column 440 units, a line across the page 864, and the lowest unit 1360; the column of the numbers
// is as wide as 00., a line starts the gap after it, and the second column a gutter after the first.
func TestMeasures(t *testing.T) {
	got := format().measures
	want := measures{space: 8, numbers: 40, indent: 56, column: 440, second: 520, across: 864, lowest: 1360}
	if got != want {
		t.Errorf("the measures are %+v, and section 14 gives %+v", got, want)
	}
	if n := Numbers(); n.BoxWidth != 100 || n.Size != 30 {
		t.Errorf("the numbers are %+v", n)
	}
	if MostItems() != 99 {
		t.Errorf("most_items is %d", MostItems())
	}
}

// A code point's width comes from its class, in the order of section 14: of no width, a digit, a
// sign, a character or a space of its own, a wide one, and any other; a cluster is as wide as its
// first code point, and a flag is wide, as the emoji are.
func TestWidths(t *testing.T) {
	tab := format()
	for _, c := range []struct {
		r     rune
		width int64
	}{
		{'7', 53}, {'×', 56}, {'a', 50}, {'ш', 79}, {'\u00a0', 26}, {'\u2007', 53}, {'\u0301', 0},
		{'三', 100}, {'🍎', 100}, {'π', 60}, {'ש', 60}, {'\u200d', 0},
	} {
		if got := tab.width(c.r); got != c.width {
			t.Errorf("%U is %d wide, want %d", c.r, got, c.width)
		}
	}
	for _, c := range []struct {
		word     string
		clusters []string
		width    int64
	}{
		{"ё", []string{"ё"}, 50},
		{"е\u0308", []string{"е\u0308"}, 50},
		{"\u0301a", []string{"\u0301", "a"}, 50},
		{"👨\u200d👩\u200d👧", []string{"👨\u200d👩\u200d👧"}, 100},
		{"🇺🇦", []string{"🇺🇦"}, 100},
		{"🇺🇦🇺", []string{"🇺🇦", "🇺"}, 200},
		{"🇺\ufe0f🇦", []string{"🇺\ufe0f🇦"}, 100},
		{"00.", []string{"0", "0", "."}, 132},
	} {
		if got := tab.clusters(c.word); !slices.Equal(got, c.clusters) {
			t.Errorf("%q has the clusters %q, want %q", c.word, got, c.clusters)
		}
		if got := tab.wordWidth(c.word); got != c.width {
			t.Errorf("%q is %d wide, want %d", c.word, got, c.width)
		}
	}
}

// A text is as wide as size × (the sum of its words' widths + the space × the spaces between them)
// ÷ 100, rounded up, its words cut at the runs of the white space that breaks; and a value fits its
// box when it is no wider than box_width.
func TestTextWidth(t *testing.T) {
	for _, c := range []struct {
		text  string
		width int64
		fits  bool
	}{
		{"", 0, true},
		{" \t\n", 0, true},
		{"7 ×", 41, true},
		{"\t7  ×\n", 41, true},
		{"Tom has 3 bags of 4 apples. He has", 453, false},
		{"123456", 96, true},
		{"1234567", 112, false},
		{"2\u00a0км", 58, true},
	} {
		if got := TextWidth(c.text); got != c.width {
			t.Errorf("%q is %d wide, want %d", c.text, got, c.width)
		}
		if got := Fits(c.text); got != c.fits {
			t.Errorf("%q fits its box: %v, want %v", c.text, got, c.fits)
		}
	}
}

// line reads a line as the tests write it: [] is a box, | ends a piece of text.
func line(text string) []Piece {
	var pieces []Piece
	for _, part := range strings.Split(text, "|") {
		for i, piece := range strings.Split(part, "[]") {
			if i > 0 {
				pieces = append(pieces, Piece{Box: true})
			}
			if piece != "" {
				pieces = append(pieces, Piece{Text: piece})
			}
		}
	}
	return pieces
}

func lines(texts ...string) [][]Piece {
	items := make([][]Piece, len(texts))
	for i, text := range texts {
		items[i] = line(text)
	}
	return items
}

// The examples of section 14 of format/README.md: "347 ÷ 6 = [ ] r [ ]" is still short, two to a
// row, and "47 : 6 = [ ] (ост. [ ])" is long, alone in its row.
func TestShortAndLong(t *testing.T) {
	short, err := Lay(lines("347 ÷ 6 =[]r[]", "347 ÷ 6 =[]r[]"))
	if err != nil {
		t.Fatal(err)
	}
	if first, second := short.Items[0].Number, short.Items[1].Number; first.Baseline != second.Baseline || second.Left != 536 {
		t.Errorf("two short items stand at %+v and %+v", first, second)
	}
	if short.Bottom != 140 {
		t.Errorf("a row of two short items ends at %d", short.Bottom)
	}
	long, err := Lay(lines("47 : 6 =[](ост.[])", "47 : 6 =[](ост.[])"))
	if err != nil {
		t.Fatal(err)
	}
	if first, second := long.Items[0].Number, long.Items[1].Number; first.Left != 56 || second.Left != 56 || second.Baseline != 232 {
		t.Errorf("two long items stand at %+v and %+v", first, second)
	}
	if long.Bottom != 272 {
		t.Errorf("two rows of long items end at %d", long.Bottom)
	}
	// A short item before a long one stands alone at the left.
	mixed, err := Lay(lines("7 ×[]= 56", "47 : 6 =[](ост.[])"))
	if err != nil {
		t.Fatal(err)
	}
	if mixed.Items[1].Number.Baseline != 232 {
		t.Errorf("a long item after a short one stands at %+v", mixed.Items[1].Number)
	}
}

// Between words of two pieces stands the space, or nothing before a closing mark and after an
// opening one; a line does not break there, nor between a box and the word after it.
func TestJoints(t *testing.T) {
	laid, err := Lay(lines("apples|.[]", "([]|)"))
	if err != nil {
		t.Fatal(err)
	}
	parts := laid.Items[0].Line
	if parts[1].Left != parts[0].Left+parts[0].Width || parts[1].Text != "." {
		t.Errorf("a closing mark of its own piece stands at %+v after %+v", parts[1], parts[0])
	}
	if parts[2].Left != parts[1].Left+parts[1].Width+16 {
		t.Errorf("a box stands at %+v after %+v", parts[2], parts[1])
	}
	parts = laid.Items[1].Line
	if parts[1].Left != parts[0].Left+parts[0].Width+16 || parts[2].Left != parts[1].Left+100+16 {
		t.Errorf("an opening mark, a box and a closing mark stand at %+v", parts)
	}
	two, err := Lay(lines("a|b"))
	if err != nil {
		t.Fatal(err)
	}
	if parts := two.Items[0].Line; parts[1].Left != parts[0].Left+parts[0].Width+8 {
		t.Errorf("words of two pieces stand at %+v", parts)
	}
}

// A long item's lines are as high as a box or a line, and between two lines stands apart when both
// have a box, lead when one has, and nothing when neither has. A box holds on to the word after it,
// and a word wider than a line breaks between its clusters, 57 letters a to a line.
func TestHeights(t *testing.T) {
	for _, c := range []struct {
		name, line      string
		baselines, tops []int64
		bottom          int64
	}{
		{"a box and a word of 120 letters", "[]" + strings.Repeat("a", 120), []int64{185, 231, 277}, []int64{40}, 290},
		{"fourteen boxes", "Boxes:" + strings.Repeat("[]", 14), []int64{100},
			[]int64{40, 40, 40, 40, 40, 40, 172, 172, 172, 172, 172, 172, 304, 304}, 404},
	} {
		laid, err := Lay(lines(c.line))
		if err != nil {
			t.Fatal(err)
		}
		var baselines, tops []int64
		for _, part := range laid.Items[0].Line {
			if part.Box {
				tops = append(tops, part.Top)
			} else {
				baselines = append(baselines, part.Baseline)
			}
		}
		if !slices.Equal(baselines, c.baselines) || !slices.Equal(tops, c.tops) || laid.Bottom != c.bottom {
			t.Errorf("%s: the baselines %v, the tops %v and the bottom %d, want %v, %v and %d", c.name,
				baselines, tops, laid.Bottom, c.baselines, c.tops, c.bottom)
		}
	}
}

// A task of no items lays out no item, and its bottom is the margin.
func TestNoItems(t *testing.T) {
	laid, err := Lay(nil)
	if err != nil || len(laid.Items) != 0 || laid.Bottom != 40 {
		t.Errorf("no items lay out %+v, %v", laid, err)
	}
}

// The page has no room for a row that would pass the lowest unit, and for an item after the first
// most_items; it is refused at the first item of that row, a *NoRoom.
func TestNoRoom(t *testing.T) {
	long := "Tom has 3 bags of 4 apples. He has[]apples."
	for _, c := range []struct {
		name  string
		items [][]Piece
		at    int
	}{
		{"eleven rows of a long item", lines(strings.Split(strings.Repeat(long+"\n", 11), "\n")[:11]...), 10},
		{"twenty-one short items", lines(strings.Split(strings.Repeat("7 =[]\n", 21), "\n")[:21]...), 20},
		{"a piece of many words", lines(strings.Repeat("many words ", 400) + "[]"), 0},
		{"a run that no line may break", lines("Marks:" + strings.Repeat(" .x", 2000) + "[]"), 0},
		{"a word of many letters", lines(strings.Repeat("a", 5000) + "[]"), 0},
	} {
		_, err := Lay(c.items)
		var noRoom *NoRoom
		if !errors.As(err, &noRoom) || noRoom.Item != c.at {
			t.Errorf("%s: %v, want no room at %d", c.name, err, c.at)
		}
	}
	err := error(&NoRoom{3})
	if want := "page: the page has no room at the row of item 3, counted from 0"; err.Error() != want {
		t.Errorf("the error says %q", err)
	}
}

// tableWith gives the table of page.yaml with numbers of its own.
func tableWith(t *testing.T, change func(*formatv0.Page)) *table {
	t.Helper()
	p := format().page
	change(&p)
	tab, err := newTable(p)
	if err != nil {
		t.Fatal(err)
	}
	return tab
}

// With numbers of their own, the layout reaches what those of page.yaml do not: an item after the
// first most_items, a line that would go past the room left on its last line, and a box wider than
// a line.
func TestOtherNumbers(t *testing.T) {
	few := tableWith(t, func(p *formatv0.Page) { p.MostItems = 3 })
	_, err := few.lay(lines("7 =[]", "7 =[]", "7 =[]", "7 =[]"))
	var noRoom *NoRoom
	if !errors.As(err, &noRoom) || noRoom.Item != 3 {
		t.Errorf("a fourth item of three at most: %v", err)
	}
	// A page of 100 units has 20 below the first row's top: one line at most, and an item of two lines
	// of text is stopped once its last line ends.
	low := tableWith(t, func(p *formatv0.Page) { p.Sheet.Height = 100 })
	_, err = low.lay(lines(strings.Repeat("a ", 50) + "[]"))
	if !errors.As(err, &noRoom) || noRoom.Item != 0 {
		t.Errorf("two lines in the room of one: %v", err)
	}
	wide := tableWith(t, func(p *formatv0.Page) { p.Numbers.BoxWidth = 900 })
	laid, err := wide.lay(lines("a[]"))
	if err != nil {
		t.Fatal(err)
	}
	if parts := laid.Items[0].Line; len(parts) != 2 || parts[1].Left != 96 || parts[1].Top != 40+46+12 {
		t.Errorf("a box wider than a line stands at %+v", parts)
	}
}

// A file of page.yaml that is not YAML, that its schema refuses, or that writes a code point
// otherwise than U+ and hexadecimal digits gives no table, and mustLoad panics on it.
func TestLoad(t *testing.T) {
	if _, err := load([]byte("sheet: [\n")); err == nil {
		t.Error("a file that is not YAML loads")
	}
	if _, err := load([]byte("sheet: {width: 1, height: 1}\n")); err == nil || !strings.Contains(err.Error(), "page.schema.json") {
		t.Errorf("a file that its schema refuses: %v", err)
	}
	if tab, err := load(pageFile); err != nil || tab.page.Numbers != Numbers() {
		t.Errorf("the copy loads %v", err)
	}
	for _, change := range []func(*formatv0.Page){
		func(p *formatv0.Page) { p.Widths.Spaces = []formatv0.PageSpaceWidth{{CodePoint: "00A0"}} },
		func(p *formatv0.Page) { p.Widths.Zero = []string{"U+0300–U+03ZZ"} },
		func(p *formatv0.Page) { p.Widths.Wide.Ranges = []string{"U+"} },
		func(p *formatv0.Page) { p.Widths.Breaks = []string{"U+110000000"} },
		func(p *formatv0.Page) { p.Widths.Joiner = "200D" },
		func(p *formatv0.Page) { p.Widths.RegionalIndicators = "U+1F1E6-U+1F1FF" },
	} {
		p := format().page
		change(&p)
		if _, err := newTable(p); !errors.Is(err, errCodePoint) {
			t.Errorf("%+v gives %v", p.Widths, err)
		}
	}
	defer func() {
		if recover() == nil {
			t.Error("mustLoad does not panic on a file that is not YAML")
		}
	}()
	mustLoad([]byte("sheet: [\n"))
}

// floorDiv rounds down, toward minus infinity, whatever the signs.
func TestFloorDiv(t *testing.T) {
	for _, c := range [][3]int64{{7, 2, 3}, {-7, 2, -4}, {7, -2, -4}, {-7, -2, 3}, {-6, 2, -3}, {0, 5, 0}} {
		if got := floorDiv(c[0], c[1]); got != c[2] {
			t.Errorf("floorDiv(%d, %d) = %d, want %d", c[0], c[1], got, c[2])
		}
	}
}

// An item on the page converts to a vector's, a text with its baseline and a box with its top and
// height.
func TestPageItem(t *testing.T) {
	laid, err := Lay(lines("7 ×[]"))
	if err != nil {
		t.Fatal(err)
	}
	item := laid.Items[0].PageItem()
	if item.Number != (formatv0.PageText{Text: "1.", Left: 56, Baseline: 100, Width: 24}) {
		t.Errorf("the number is %+v", item.Number)
	}
	text, box := item.Line[0], item.Line[1]
	if *text.Text != "7 ×" || *text.Left != 96 || *text.Baseline != 100 || *text.Width != 41 || text.Top != nil || text.Height != nil {
		t.Errorf("the text is %+v", text)
	}
	if box.Text != nil || box.Baseline != nil || *box.Left != 153 || *box.Top != 40 || *box.Width != 100 || *box.Height != 100 {
		t.Errorf("the box is %+v", box)
	}
}
