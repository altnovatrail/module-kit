package page

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/altnovatrail/spec/types/formatv0"
)

// Piece is a piece of an item's line, as the open view holds it: a piece of text, Text, or, when Box
// is set, a box, an answer slot, whose Text is nothing (format/README.md, section 14).
type Piece struct {
	Text string
	Box  bool
}

// Text is a text set on the page: Text, its words joined by single spaces, from Left, on its
// alphabetic baseline, Baseline, fitted to Width. Every number is in units of the page.
type Text struct {
	Text     string
	Left     int64
	Baseline int64
	Width    int64
}

// Part is a part of an item's line on the page: a text, as a Text, or, when Box is set, a box, from
// Left and Top, Width wide and Height high. The fields of the other kind are 0, and a box has no
// Text.
type Part struct {
	Box      bool
	Text     string
	Left     int64
	Baseline int64
	Width    int64
	Top      int64
	Height   int64
}

// PagePart gives the part as a vector of the page holds it: a text with its left, its baseline and
// its width, and a box with its left, top, width and height.
func (p Part) PagePart() formatv0.PagePart {
	if p.Box {
		return formatv0.PagePart{Left: &p.Left, Top: &p.Top, Width: &p.Width, Height: &p.Height}
	}
	return formatv0.PagePart{Text: &p.Text, Left: &p.Left, Baseline: &p.Baseline, Width: &p.Width}
}

// Item is an item on the page: its number, n., right in the column of the numbers on the baseline
// of its first line, and the texts and boxes of its line, in the order they are read.
type Item struct {
	Number Text
	Line   []Part
}

// PageItem gives the item as a vector of the page holds it.
func (i Item) PageItem() formatv0.PageItem {
	line := make([]formatv0.PagePart, len(i.Line))
	for n, part := range i.Line {
		line[n] = part.PagePart()
	}
	return formatv0.PageItem{
		Number: formatv0.PageText{Text: i.Number.Text, Left: i.Number.Left, Baseline: i.Number.Baseline, Width: i.Number.Width},
		Line:   line,
	}
}

// Layout is the page of a task's items: each item on it, in the task's order, and Bottom, the bottom
// of the last row. A task of no items has none, and its bottom is the margin, where the first row
// would start.
type Layout struct {
	Items  []Item
	Bottom int64
}

// NoRoom is what Lay gives when the page has no room: Item, counted from 0, is the first item of the
// row that has none, at which the page is refused (task.workbook-page) — or the first item after the
// most_items of page.yaml, which has none either (format/README.md, section 14).
type NoRoom struct {
	Item int
}

func (e *NoRoom) Error() string {
	return fmt.Sprintf("page: the page has no room at the row of item %d, counted from 0", e.Item)
}

// Lay lays out the items of a task on a workbook, each its line of pieces, on the page, as section 14
// of format/README.md does by page.yaml. Rows start at the margin, each the gap of items below the
// one before: two short items side by side, in the two columns, and a short item before a long one,
// or before none, alone at the left; a long item across the page, its lines filled greedily, run by
// run. When a row would pass the lowest unit, or an item comes after the most_items of page.yaml,
// it gives a *NoRoom with the first item of that row, and nothing ever shrinks.
func Lay(items [][]Piece) (Layout, error) {
	return format().lay(items)
}

// A token is a word of a piece of text, with its width in hundredths of the type size, or a box;
// piece is the place of its piece in the line.
type token struct {
	piece int
	box   bool
	word  string
	width int64
}

// A joint is what lies between a token and the one before it in a line: inner, the space within the
// text of one piece, which the text's width holds; room, what the layout sets there otherwise;
// and whether a line may break there.
type joint struct {
	inner  bool
	room   int64
	breaks bool
}

// An entry is a token of a line with the joint before it; start says that it starts its line, with
// nothing before it.
type entry struct {
	token
	joint
	start bool
}

// A part laid out on a line, from the line's start: a text or a box.
type laid struct {
	box   bool
	text  string
	left  int64
	width int64
}

func (t *table) lay(items [][]Piece) (Layout, error) {
	n := t.page.Numbers
	lines := make([][]entry, len(items))
	lineOf := func(i int) []entry {
		if lines[i] == nil {
			lines[i] = t.entries(items[i])
		}
		return lines[i]
	}
	short := func(i int) bool {
		_, width := t.parts(lineOf(i))
		return width <= t.column-t.indent
	}
	most := int(t.page.MostItems)
	var placed []Item
	top, bottom := n.Margin, n.Margin
	for i := 0; i < len(items); {
		room := t.lowest - top
		if i >= most {
			return Layout{}, &NoRoom{i}
		}
		var row []Item
		var height int64
		if short(i) {
			cells := []int{i}
			if i+1 < len(items) && i+1 < most && short(i+1) {
				cells = append(cells, i+1)
			}
			for c, cell := range cells {
				left := n.Margin
				if c == 1 {
					left = t.second
				}
				item, h := t.item([][]entry{lineOf(cell)}, cell, left, top)
				row = append(row, item)
				height = max(height, h)
			}
		} else {
			wrapped, ok := t.wrap(lineOf(i), room)
			if !ok {
				return Layout{}, &NoRoom{i}
			}
			item, h := t.item(wrapped, i, n.Margin, top)
			row, height = []Item{item}, h
		}
		if height > room {
			return Layout{}, &NoRoom{i}
		}
		placed = append(placed, row...)
		bottom = top + height
		i += len(row)
		top = bottom + n.Items
	}
	return Layout{Items: placed, Bottom: bottom}, nil
}

// entries gives the tokens of an item's line, the words of its pieces of text and its boxes, in
// order, each with the joint before it.
func (t *table) entries(pieces []Piece) []entry {
	var tokens []token
	for p, piece := range pieces {
		if piece.Box {
			tokens = append(tokens, token{piece: p, box: true})
			continue
		}
		for _, word := range t.words(piece.Text) {
			tokens = append(tokens, token{piece: p, word: word, width: t.wordWidth(word)})
		}
	}
	entries := make([]entry, len(tokens))
	for i, tok := range tokens {
		entries[i] = entry{token: tok, start: i == 0}
		if i > 0 {
			entries[i].joint = t.joint(tokens[i-1], tok)
		}
	}
	return entries
}

// joint gives what lies between two tokens next to each other (format/README.md, section 14).
// Between two words of one piece stands the space, within its text; between words of two pieces,
// the space's own width, or nothing when the second begins with a closing mark or the first ends
// with an opening one. Between a word and a box, either way, stands the gap, and between two boxes
// apart. A line breaks between two tokens but after an opening mark, before a closing one, and
// between a box and the word after it, which the box holds on to as a number holds on to its unit.
func (t *table) joint(a, b token) joint {
	n := t.page.Numbers
	switch {
	case !a.box && !b.box:
		glued := t.closes(b.word) || t.opens(a.word)
		if a.piece == b.piece {
			return joint{inner: true, breaks: !glued}
		}
		width := t.space
		if glued {
			width = 0
		}
		return joint{room: width, breaks: !glued}
	case !a.box:
		return joint{room: n.Gap, breaks: !t.opens(a.word)}
	case !b.box:
		return joint{room: n.Gap}
	}
	return joint{room: n.Apart, breaks: true}
}

// closes says whether a word begins with a closing mark, and opens whether it ends with an opening
// one, of joints of page.yaml.
func (t *table) closes(word string) bool {
	r, _ := utf8.DecodeRuneInString(word)
	return strings.ContainsRune(t.page.Joints.Closing, r)
}

func (t *table) opens(word string) bool {
	r, _ := utf8.DecodeLastRuneInString(word)
	return strings.ContainsRune(t.page.Joints.Opening, r)
}

// parts lays out a line of entries from its start: the words of one piece that follow each other
// on it make one text, as wide as size × (the sum of their widths + the space × the spaces between
// them) ÷ 100, rounded up, and a box is the box's width. It gives each text and box at its place,
// and the width of the whole line.
func (t *table) parts(entries []entry) ([]laid, int64) {
	var parts []laid
	var x, start, hundredths int64
	var words []string
	end := func() {
		if len(words) == 0 {
			return
		}
		width := t.units(hundredths + t.page.Widths.Space*int64(len(words)-1))
		parts = append(parts, laid{text: strings.Join(words, " "), left: start, width: width})
		x = start + width
		words, hundredths = nil, 0
	}
	for _, e := range entries {
		if !e.box && !e.start && e.inner && len(words) > 0 {
			words = append(words, e.word)
			hundredths += e.width
			continue
		}
		end()
		if !e.start {
			x += e.room
		}
		if e.box {
			parts = append(parts, laid{box: true, left: x, width: t.page.Numbers.BoxWidth})
			x += t.page.Numbers.BoxWidth
			continue
		}
		start, words, hundredths = x, []string{e.word}, e.width
	}
	end()
	return parts, x
}

// fits says whether a line of entries is no wider than a line across the page.
func (t *table) fits(entries ...[]entry) bool {
	_, width := t.parts(slices.Concat(entries...))
	return width <= t.across
}

// fresh gives an entry as the start of a line.
func fresh(e entry) entry {
	return entry{token: e.token, start: true}
}

// wrap fills the lines of a long item across the page greedily, run by run, where a run is the
// tokens between which no line may break: a run that does not fit after the line so far starts the
// next line; a run wider than a line goes token by token; and a word wider than a line is broken
// between its clusters, as many to a line as fit (format/README.md, section 14).
//
// It stops, and gives false, once the item has more lines than could fit in room, the room left
// below the row's top: every line is at least as high as the lower of a line of text and a box, and
// so many lines could not fit anyway: the row would have no room. That only stops early what
// could not fit, as spec's second implementation does, so that a piece of many words is not laid
// out to its end.
func (t *table) wrap(entries []entry, room int64) ([][]entry, bool) {
	n := t.page.Numbers
	most := floorDiv(room, min(n.Line, n.BoxHeight)) + 1
	var lines [][]entry
	var current []entry
	// end ends the current line, and says whether the lines are now too many.
	end := func() bool {
		if len(current) > 0 {
			lines = append(lines, current)
		}
		current = nil
		return int64(len(lines)) > most
	}
	for _, run := range runs(entries) {
		if len(current) > 0 && t.fits(current, run) {
			current = append(current, run...)
			continue
		}
		if end() {
			return nil, false
		}
		alone := slices.Concat([]entry{fresh(run[0])}, run[1:])
		if t.fits(alone) {
			current = alone
			continue
		}
		for _, e := range run {
			if len(current) > 0 && t.fits(current, []entry{e}) {
				current = append(current, e)
				continue
			}
			if end() {
				return nil, false
			}
			if t.fits([]entry{fresh(e)}) || e.box {
				// A box wider than a line, which the numbers of page.yaml never make, stands alone on
				// its line: nothing shrinks.
				current = []entry{fresh(e)}
				continue
			}
			// A word wider than a line: its clusters, as many to a line as fit.
			var chunk string
			var width int64
			for _, cluster := range t.clusters(e.word) {
				w := t.clusterWidth(cluster)
				if chunk != "" && t.units(width+w) > t.across {
					current = []entry{{token: token{piece: e.piece, word: chunk, width: width}, start: true}}
					if end() {
						return nil, false
					}
					chunk, width = "", 0
				}
				chunk += cluster
				width += w
			}
			current = []entry{{token: token{piece: e.piece, word: chunk, width: width}, start: true}}
		}
	}
	if end() {
		return nil, false
	}
	return lines, true
}

// runs cuts a line of entries where a line may break: each run starts with the entry after a break,
// with the joint before it.
func runs(entries []entry) [][]entry {
	var found [][]entry
	for _, e := range entries {
		if len(found) == 0 || e.start || e.breaks {
			found = append(found, []entry{e})
			continue
		}
		found[len(found)-1] = append(found[len(found)-1], e)
	}
	return found
}

// item places an item, n from 0, its lines laid out from the left of its column and from top: a
// line with a box is the box's height, and one without the height of a line; between two lines
// stands apart when both have a box, lead when one has, and nothing when neither has. A text stands
// on its line's baseline, its top plus half its height, rounded down, plus a third of the type size,
// rounded down, and a box's top is its line's top. The number stands right in the column of the
// numbers, on the baseline of the first line. It gives the item and its height. An item has one
// line at least: a short item is one, and a long one has a token.
func (t *table) item(lines [][]entry, n int, left, top int64) (Item, int64) {
	nums := t.page.Numbers
	var item Item
	var first int64
	var before bool // whether the line before has a box
	y := top
	for i, entries := range lines {
		parts, _ := t.parts(entries)
		boxed := slices.ContainsFunc(parts, func(p laid) bool { return p.box })
		if i > 0 {
			switch {
			case before && boxed:
				y += nums.Apart
			case before || boxed:
				y += nums.Lead
			}
		}
		height := nums.Line
		if boxed {
			height = nums.BoxHeight
		}
		baseline := y + floorDiv(height, 2) + floorDiv(nums.Size, 3)
		if i == 0 {
			first = baseline
		}
		x := left + t.indent
		for _, p := range parts {
			if p.box {
				item.Line = append(item.Line, Part{Box: true, Left: x + p.left, Top: y, Width: p.width, Height: nums.BoxHeight})
			} else {
				item.Line = append(item.Line, Part{Text: p.text, Left: x + p.left, Baseline: baseline, Width: p.width})
			}
		}
		y += height
		before = boxed
	}
	label := strconv.Itoa(n+1) + "."
	width := t.units(t.wordWidth(label))
	item.Number = Text{Text: label, Left: left + t.numbers - width, Baseline: first, Width: width}
	return item, y - top
}
