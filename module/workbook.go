package module

import (
	"errors"
	"fmt"
	"slices"
	"strconv"

	"github.com/altnovatrail/spec/types/formatv0"

	"github.com/altnovatrail/module-kit/page"
)

// The rules of a workbook's page and of a worked example on a workbook (format/README.md, sections
// 14 and 21).
const (
	RuleTaskWorkbookPage   = "task.workbook-page"
	RuleLessonExampleFills = "lesson.example-fills"
	RuleLessonWorkbookPage = "lesson.workbook-page"
)

// A box of a worked example's lines, by its line and its place in the line, each from 1, as a fill
// of the demo names it.
type exampleBox struct {
	line, box int64
}

// exampleFills checks lesson.example-fills on each worked example on a workbook of a lesson: every
// fill of its demo names a box of its lines, and the demo fills each box of its lines once
// (format/README.md, section 21). A fill of no box and a second fill of a box stand where the demo
// fills it, and a box that no fill fills where the line holds it.
func (c *checker) exampleFills(path string, lesson formatv0.Lesson) {
	for i, segment := range lesson.Segments {
		example := segment.WorkedExample
		if example == nil || example.Canvas.Kind != formatv0.CanvasKindWorkbook {
			continue
		}
		place := []string{"segments", strconv.Itoa(i), "worked_example"}
		var boxes []exampleBox
		for l, line := range example.Lines {
			box := int64(0)
			for _, part := range line.Line {
				if part.Binding != nil {
					box++
					boxes = append(boxes, exampleBox{int64(l) + 1, box})
				}
			}
		}
		filled := map[exampleBox]int{}
		for k, part := range example.Demo {
			if part.Fill == nil {
				continue
			}
			b := exampleBox{part.Fill.Line, part.Fill.Box}
			where := at(slices.Concat(place, []string{"demo", strconv.Itoa(k), "fill"})...)
			switch first, twice := filled[b]; {
			case !slices.Contains(boxes, b):
				c.add(RuleLessonExampleFills, path, 0, 0, where,
					fmt.Sprintf("the demo fills box %d of line %d, and the example's lines have no such box", b.box, b.line))
			case twice:
				c.add(RuleLessonExampleFills, path, 0, 0, where,
					fmt.Sprintf("the demo fills box %d of line %d again: the part at %s filled it first, and it fills each box once", b.box, b.line, at(slices.Concat(place, []string{"demo", strconv.Itoa(first)})...)))
			default:
				filled[b] = k
			}
		}
		for _, b := range boxes {
			if _, ok := filled[b]; !ok {
				c.add(RuleLessonExampleFills, path, 0, 0, at(slices.Concat(place, []string{"lines", strconv.FormatInt(b.line-1, 10)})...),
					fmt.Sprintf("no part of the demo fills box %d of line %d, and the demo fills each box of the lines", b.box, b.line))
			}
		}
	}
}

// workbookPage checks task.workbook-page on a task on a workbook that is no template, or on an
// instance of one, whose texts hold the numbers of the instance: its items fit its page as the page
// of a workbook lays them out, every row with room and no more items than most_items
// (format/README.md, section 14). The breach stands at the first item of the row that has no room,
// as canvas refuses to draw the page there.
func (c *checker) workbookPage(path string, task formatv0.Task) {
	if task.Canvas.Kind != formatv0.CanvasKindWorkbook {
		return
	}
	items := make([][]page.Piece, len(task.Items))
	for i, item := range task.Items {
		for _, part := range item.Line {
			items[i] = append(items[i], piece(part.Text, part.Binding != nil))
		}
	}
	if n, ok := noRoom(items); ok {
		c.add(RuleTaskWorkbookPage, path, 0, 0, at("items", strconv.Itoa(n)), roomSaid(n, "item"))
	}
}

// exampleOnPage checks lesson.workbook-page on each worked example on a workbook of a lesson: its
// lines fit the page as a task's lines do, and the value of each box is no wider, as a text of the
// line, than the box (format/README.md, section 21).
func (c *checker) exampleOnPage(path string, lesson formatv0.Lesson) {
	for i, segment := range lesson.Segments {
		example := segment.WorkedExample
		if example == nil || example.Canvas.Kind != formatv0.CanvasKindWorkbook {
			continue
		}
		place := []string{"segments", strconv.Itoa(i), "worked_example", "lines"}
		lines := make([][]page.Piece, len(example.Lines))
		for l, line := range example.Lines {
			for k, part := range line.Line {
				lines[l] = append(lines[l], piece(part.Text, part.Binding != nil))
				if part.Value != nil && !page.Fits(string(*part.Value)) {
					c.add(RuleLessonWorkbookPage, path, 0, 0, at(slices.Concat(place, []string{strconv.Itoa(l), "line", strconv.Itoa(k), "value"})...),
						fmt.Sprintf("the value %s is %d units wide on the page, wider than its box of %d", *part.Value, page.TextWidth(string(*part.Value)), page.Numbers().BoxWidth))
				}
			}
		}
		if n, ok := noRoom(lines); ok {
			c.add(RuleLessonWorkbookPage, path, 0, 0, at(slices.Concat(place, []string{strconv.Itoa(n)})...), roomSaid(n, "line"))
		}
	}
}

// piece gives a piece of a line as the layout reads it: a box, or else a text, as the schema of the
// part has a piece one or the other.
func piece(text *formatv0.Text, box bool) page.Piece {
	if box {
		return page.Piece{Box: true}
	}
	return page.Piece{Text: string(*text)}
}

// noRoom lays out lines of a workbook and gives the line, from 0, at whose row the page has no room,
// and whether there is one: page.Lay refuses a page only so.
func noRoom(lines [][]page.Piece) (int, bool) {
	_, err := page.Lay(lines)
	var full *page.NoRoom
	if errors.As(err, &full) {
		return full.Item, true
	}
	return 0, false
}

// roomSaid writes why a page refuses its lines at the line n: its row has no room, or it comes after
// the most_items of page.yaml, which, with the numbers of page.yaml, rows run out long before.
func roomSaid(n int, what string) string {
	return fmt.Sprintf("the row of %s %d has no room on the page of a workbook, which lays the lines out as format/README.md, section 14, says, holds %d at most and shrinks nothing", what, n+1, page.MostItems())
}
