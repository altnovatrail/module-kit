// Package page lays out the page of a workbook: where each item of a task on a workbook stands, its
// number, the texts of its line and its boxes, in whole units of the page, from the items alone, by
// the numbers and the table of widths of format/page.yaml (format/README.md, section 14; the
// 2026-11-01 amendment to R51 spec). A box so stands where it stands on every device, whatever its
// fonts: the canvas fits each text to the width the layout gives it (R34 canvas, R35 canvas).
//
// Lay gives the Layout of a task's items, each a line of Pieces — a text, or a box —, or NoRoom,
// the item at whose row the page has no room. TextWidth gives the width of a text as the layout sets
// it on one line, and Fits whether a value written in a box is no wider than the box, Numbers'
// BoxWidth. The test vectors of the page, format/vectors/pages/, pin every place (section 9), and
// internal/vectors passes them all.
//
// The check of a module holds task.workbook-page by it, and lesson.workbook-page, whose values fit
// their boxes by Fits (section 21); canvas lays its pages out by the same numbers, from the npm
// package of spec. The package holds a copy of page.yaml, byte for byte that of spec.Format at the
// tag go.mod requires, and does not import spec itself: a consumer of the library carries the one
// file, not every file of the format.
package page
