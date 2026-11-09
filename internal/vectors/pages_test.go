package vectors

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/altnovatrail/spec/types/formatv0"

	"github.com/altnovatrail/module-kit/internal/schema"
	"github.com/altnovatrail/module-kit/page"
)

// pages passes a case of the workbook's page: the layout of its items by page.yaml is the vector's
// page and its bottom, whatever the order of keys, or has no room at the row of the item that
// no_room_at gives, counted from 0 (format/README.md, sections 9 and 14). The vector is read by its
// schema and its type of formatv0.
func pages(v vector) error {
	var pv formatv0.PageVector
	value, err := readVector(v.fsys, v.path, "page-vector.schema.json", &pv)
	if err != nil {
		return err
	}
	items := make([][]page.Piece, len(pv.Items))
	for i, item := range pv.Items {
		for _, part := range item.Line {
			if part.Binding != nil {
				items[i] = append(items[i], page.Piece{Box: true})
			} else {
				items[i] = append(items[i], page.Piece{Text: *part.Text})
			}
		}
	}
	laid, err := page.Lay(items)
	var noRoom *page.NoRoom
	if err != nil && !errors.As(err, &noRoom) {
		return err
	}
	switch {
	case noRoom != nil && pv.NoRoomAt == nil:
		return fmt.Errorf("the page has no room at item %d, from 0, and the vector gives its page", noRoom.Item)
	case noRoom == nil && pv.NoRoomAt != nil:
		return fmt.Errorf("the page has room for every item, and the vector gives no room at item %d", *pv.NoRoomAt)
	case noRoom != nil && int64(noRoom.Item) != *pv.NoRoomAt:
		return fmt.Errorf("the page has no room at item %d, from 0, and the vector gives item %d", noRoom.Item, *pv.NoRoomAt)
	case noRoom != nil:
		return nil
	}
	placed := make([]formatv0.PageItem, len(laid.Items))
	for i, item := range laid.Items {
		placed[i] = item.PageItem()
	}
	vector := value.(map[string]any)
	got := schema.ValueOf(map[string]any{"page": placed, "bottom": laid.Bottom})
	if want := map[string]any{"page": vector["page"], "bottom": vector["bottom"]}; !reflect.DeepEqual(got, want) {
		gotText, _ := json.MarshalIndent(got, "", "  ")
		wantText, _ := json.MarshalIndent(want, "", "  ")
		return fmt.Errorf("the page is\n%s\nand the vector gives\n%s", gotText, wantText)
	}
	return nil
}

// The runner of the page refuses a vector that any of its fields would make wrong: a place on the
// page, its bottom, the item at which it has no room, a page where it has none and no room where it
// has.
func TestPagesRefuse(t *testing.T) {
	const item = "items:\n  - line:\n      - text: \"7 ×\"\n      - binding: box\n"
	const laid = item + "bottom: 140\npage:\n" +
		"  - number: {text: \"1.\", left: 56, baseline: 100, width: 24}\n" +
		"    line:\n" +
		"      - {text: \"7 ×\", left: 96, baseline: 100, width: 41}\n" +
		"      - {left: 153, top: 40, width: 100, height: 100}\n"
	many := "items:\n" + strings.Repeat("  - line:\n      - binding: box\n", 21)
	for _, c := range []struct {
		name, data, says string
	}{
		{"a vector", laid, ""},
		{"no room where the vector says", many + "no_room_at: 20\n", ""},
		{"not of its schema", item, "page-vector.schema.json"},
		{"another bottom", strings.Replace(laid, "bottom: 140", "bottom: 141", 1), "the page is"},
		{"another left", strings.Replace(laid, "left: 153", "left: 152", 1), "the page is"},
		{"another number", strings.Replace(laid, `"1."`, `"2."`, 1), "the page is"},
		{"no room at another item", many + "no_room_at: 19\n", "no room at item 20, from 0, and the vector gives item 19"},
		{"a page where it has no room", many + "bottom: 140\npage: []\n", "the vector gives its page"},
		{"no room where it has", item + "no_room_at: 0\n", "the page has room for every item"},
	} {
		err := pages(vector{fsys: fstest.MapFS{"v.yaml": {Data: []byte(c.data)}}, path: "v.yaml"})
		if c.says == "" && err != nil || c.says != "" && (err == nil || !strings.Contains(err.Error(), c.says)) {
			t.Errorf("%s: %v, want %q", c.name, err, c.says)
		}
	}
	if err := pages(vector{fsys: fstest.MapFS{}, path: "none.yaml"}); err == nil {
		t.Error("a file that is not there passes")
	}
}
