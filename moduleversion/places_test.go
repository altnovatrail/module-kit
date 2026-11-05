package moduleversion

import (
	"errors"
	"io/fs"
	"slices"
	"testing"
	"testing/fstest"

	"github.com/altnovatrail/module-kit/modulepackage"
)

// A published package of an earlier version of the format is read by its places, and refused when it
// breaks the chain of digests or does not hold them where the format keeps them; one of a newer
// version of the format is refused by its version.
func TestReadPublished(t *testing.T) {
	good := asPublished(t, "vectors/versions/format-transition/published/1.0.0")
	files := func() fstest.MapFS {
		folder := fstest.MapFS{}
		for _, name := range []string{modulepackage.InventoryFile, modulepackage.ServerFile} {
			data, err := fs.ReadFile(good.Files, name)
			if err != nil {
				t.Fatal(err)
			}
			folder[name] = &fstest.MapFile{Data: data}
		}
		return folder
	}
	if v, err := readPublished(good); err != nil || v.pkg != nil || v.places.formatVersion != "0.2" || v.places.version.String() != "1.0.0" {
		t.Errorf("readPublished = %+v, %v", v, err)
	}
	inventory := func(text string) Published {
		folder := files()
		folder[modulepackage.InventoryFile] = &fstest.MapFile{Data: []byte(text)}
		return Published{Digest: digestOf([]byte(text)), Files: folder}
	}
	noServer := files()
	delete(noServer, modulepackage.ServerFile)
	otherServer := files()
	otherServer[modulepackage.ServerFile] = &fstest.MapFile{Data: []byte(`{}`)}
	for _, c := range []struct {
		name   string
		p      Published
		reason error
	}{
		{"another digest", Published{Digest: digestOf([]byte("x")), Files: files()}, modulepackage.ErrChain},
		{"no inventory", Published{Digest: good.Digest, Files: fstest.MapFS{}}, modulepackage.ErrChain},
		{"no server part", Published{Digest: good.Digest, Files: noServer}, modulepackage.ErrChain},
		{"another server part", Published{Digest: good.Digest, Files: otherServer}, modulepackage.ErrChain},
		{"no JSON", inventory(`{`), modulepackage.ErrNotCanonical},
		{"a newer format", inventory(`{"module_format_version":"0.4"}`), modulepackage.ErrFormatVersion},
		{"no format version", inventory(`{"id":"example/versions"}`), modulepackage.ErrFormatVersion},
		{"a server part named otherwise", inventory(`{"module_format_version":"0.2","server_part":{"digest":"sha256:0","size":1}}`), modulepackage.ErrChain},
		{"a 0.3 package not of its schema", inventory(`{"module_format_version":"0.3"}`), modulepackage.ErrNotCanonical},
	} {
		if _, err := readPublished(c.p); !errors.Is(err, c.reason) {
			t.Errorf("%s: readPublished = %v; want %v", c.name, err, c.reason)
		}
	}
}

// placesOf reads the places alone, and refuses a package that does not hold one where the format
// keeps it.
func TestPlacesOf(t *testing.T) {
	inventory := map[string]any{"module_format_version": "0.2", "id": "m", "version": "1.0.0"}
	manifest := func() map[string]any {
		return map[string]any{"languages": map[string]any{"explanation": "en"}, "bands": []any{map[string]any{"id": "b"}}}
	}
	view := map[string]any{
		"manifest": manifest(),
		"skill_graph": map[string]any{"skills": map[string]any{"s": map[string]any{}},
			"retired_skills": map[string]any{"r": map[string]any{"replaced_by": []any{"s"}}}},
		"tasks": map[string]any{"t": map[string]any{}},
		"traps": map[string]any{"traps": map[string]any{"p": map[string]any{}}},
	}
	p, err := placesOf(inventory, view)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(p.bands, []string{"b"}) || !slices.Equal(p.ids["skills"], []string{"s"}) || !slices.Equal(p.ids["traps"], []string{"p"}) ||
		len(p.ids["lessons"]) != 0 || !slices.Equal(p.retired["r"], []string{"s"}) {
		t.Errorf("placesOf = %+v", p)
	}
	with := func(name string, value any) map[string]any {
		return map[string]any{"manifest": manifest(), name: value}
	}
	for _, c := range []struct {
		name            string
		inventory, view any
	}{
		{"no inventory", nil, view},
		{"no id", map[string]any{"module_format_version": "0.2"}, view},
		{"no version", map[string]any{"module_format_version": "0.2", "id": "m"}, view},
		{"no manifest", inventory, map[string]any{}},
		{"no languages", inventory, map[string]any{"manifest": map[string]any{}}},
		{"a band without its id", inventory, map[string]any{"manifest": map[string]any{"languages": map[string]any{}, "bands": []any{1}}}},
		{"retired skills not an object", inventory, with("skill_graph", map[string]any{"retired_skills": []any{}})},
		{"no replaced_by", inventory, with("skill_graph", map[string]any{"retired_skills": map[string]any{"r": map[string]any{}}})},
		{"replaced_by not of strings", inventory, with("skill_graph", map[string]any{"retired_skills": map[string]any{"r": map[string]any{"replaced_by": []any{1}}}})},
		{"skills not an object", inventory, with("skill_graph", map[string]any{"skills": []any{}})},
		{"tasks not an object", inventory, with("tasks", []any{})},
		{"a catalog not an object", inventory, with("traps", "x")},
	} {
		if _, err := placesOf(c.inventory, c.view); !errors.Is(err, modulepackage.ErrNotCanonical) {
			t.Errorf("%s: placesOf = %v; want ErrNotCanonical", c.name, err)
		}
	}
}
