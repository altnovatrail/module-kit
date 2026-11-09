package page

import (
	"errors"
	"strconv"
	"testing"
)

// FuzzLay lays out a task of count items, each the same line, written as the tests write lines:
// [] is a box, | ends a piece of text. Whatever the line, the layout gives a page or a *NoRoom with
// an item of the task; and on a page every text is as wide as TextWidth sets it, every part stands
// within its column or across the page, no row passes the lowest unit, and each number stands right
// in the column of the numbers.
func FuzzLay(f *testing.F) {
	for _, seed := range []string{
		"7 ×[]= 56", "[]÷ 8 = 4", "Tom has 3 bags of 4 apples. He has[]apples.", "47 ÷ 6 =[]r[]",
		"He has[]|.", "(|[]|)", "apples|.[]", "Fill in the boxes:[][][][][][][]", "\t7  ×\n[] = 56 ",
		"2\u00a0км =[]м", "👨\u200d👩\u200d👧 =[]", "🇺🇦[]", "е\u0308 = ё[]", "[]",
		"The longest word: supercalifragilisticsupercalifragilisticsupercalifragilisticsupercalifragilistic[]",
		"Marks: .x .x .x .x .x .x .x .x .x .x .x .x .x .x .x .x .x .x .x .x .x .x .x .x .x .x .x .x .x[]",
	} {
		f.Add(seed, uint8(3))
	}
	tab := format()
	f.Fuzz(func(t *testing.T, text string, count uint8) {
		n := int(count%24) + 1
		items := make([][]Piece, n)
		for i := range items {
			items[i] = line(text)
		}
		laid, err := Lay(items)
		var noRoom *NoRoom
		if errors.As(err, &noRoom) {
			if noRoom.Item < 0 || noRoom.Item >= n {
				t.Fatalf("no room at item %d of %d", noRoom.Item, n)
			}
			return
		}
		if err != nil || len(laid.Items) != n || laid.Bottom > tab.lowest {
			t.Fatalf("%d items of %q lay out %d, to %d: %v", n, text, len(laid.Items), laid.Bottom, err)
		}
		margin, right := tab.page.Numbers.Margin, tab.page.Sheet.Width-tab.page.Numbers.Margin
		for i, item := range laid.Items {
			left := margin
			if item.Number.Left+item.Number.Width == tab.second+tab.numbers {
				left = tab.second
			} else if item.Number.Left+item.Number.Width != margin+tab.numbers {
				t.Fatalf("item %d: the number %+v is not right in a column of the numbers", i, item.Number)
			}
			if item.Number.Text != strconv.Itoa(i+1)+"." {
				t.Fatalf("item %d is numbered %q", i, item.Number.Text)
			}
			for _, part := range item.Line {
				if !part.Box && part.Width != TextWidth(part.Text) {
					t.Fatalf("item %d: the text %q is %d wide, and TextWidth sets it %d", i, part.Text, part.Width, TextWidth(part.Text))
				}
				if part.Left < left+tab.indent || part.Left+part.Width > right || left == tab.second && part.Left+part.Width > tab.second+tab.column {
					t.Fatalf("item %d: %+v stands outside its column", i, part)
				}
			}
		}
	})
}
