package module

import (
	"context"
	"io/fs"
	"path"
	"slices"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/altnovatrail/spec"

	"github.com/altnovatrail/module-kit/internal/yaml"
)

// warningsOf gives the warnings of a report each by its place, as Warning.String writes it, and the
// numbers its message names.
func warningsOf(r Report) []string {
	var warnings []string
	for _, w := range r.Warnings {
		numbers, _, _ := strings.Cut(strings.TrimPrefix(w.Message, "the text holds "), ":")
		warnings = append(warnings, w.Path+" "+w.Pointer+" "+numbers)
	}
	return warnings
}

// A number in a rung template, in the advice of a trap and in the title or the description of a
// technique gives a warning, at its place and in the order of the places, each number once and as it
// stands in the text's normal form; the warnings refuse nothing, and a module refused by a rule has
// them too. A part whose file its schema refuses gives none, and neither does a text of a task.
func TestWarnings(t *testing.T) {
	ctx := context.Background()
	if report, err := Check(ctx, base); err != nil || len(report.Breaches) > 0 || len(report.Warnings) > 0 {
		t.Fatalf("base gives %v, %v and %v", shortReport(report), report.Warnings, err)
	}

	numbered := module(fstest.MapFS{
		manifestPath:   file(replaced(manifest, "Which step of your work", "Which of your 1000 steps")),
		trapsPath:      file(replaced(string(base[trapsPath].Data), "Draw the posts as dots", "Draw 1000 posts as dots, ½ apart, 1000 in all,")),
		techniquesPath: file(replaced(string(base[techniquesPath].Data), "title: Start small", "title: Start with 1000 or 2000", "Lines and dots", "１２ lines and 08:30 dots")),
	})
	want := []string{
		`manifest.yaml #/rung_templates/0 the number "1000"`,
		`techniques.yaml #/techniques/draw/description the numbers "12", "08" and "30"`,
		`techniques.yaml #/techniques/small-numbers/title the numbers "1000" and "2000"`,
		`traps.yaml #/traps/off-by-one/trap_advice the numbers "1000" and "1/2"`,
	}
	report, err := Check(ctx, numbered)
	if err != nil || len(report.Breaches) > 0 || !slices.Equal(warningsOf(report), want) {
		t.Errorf("the numbered module gives %v, %q and %v; want no breach and %q", shortReport(report), warningsOf(report), err, want)
	}
	if w := report.Warnings[0].String(); w != "manifest.yaml #/rung_templates/0: warning: the text holds the number \"1000\": "+
		"a text of the whole module is searched for the keys of every task that can show it, so it SHOULD hold no number (format/README.md, section 20)" {
		t.Errorf("a warning is written %q", w)
	}

	refused := with(numbered, fstest.MapFS{"LICENSE": nil})
	if report, err := Check(ctx, refused); err != nil || !slices.Equal(shortReport(report), []string{". module.license"}) || !slices.Equal(warningsOf(report), want) {
		t.Errorf("the numbered module without LICENSE gives %v and %q; want module.license and %q", shortReport(report), warningsOf(report), want)
	}

	schemaRefused := with(numbered, fstest.MapFS{
		trapsPath: file(replaced(string(numbered[trapsPath].Data), "traps:\n", "traps:\n  bad trap:\n    description: A trap with no advice.\n")),
	})
	report, _ = Check(ctx, schemaRefused)
	if got := warningsOf(report); len(report.Breaches) == 0 || !slices.Equal(got, want[:3]) {
		t.Errorf("with traps.yaml refused by its schema, the module gives %v and %q; want breaches and %q", shortReport(report), got, want[:3])
	}

	if report, err := Check(ctx, module(nil)); err != nil || len(report.Warnings) > 0 {
		t.Errorf("the tasks of base, which hold numbers, give %q and %v", warningsOf(report), err)
	}
	if report, err := CheckFile(partManifest, numbered[manifestPath].Data); err != nil || len(report.Breaches) > 0 || len(report.Warnings) > 0 {
		t.Errorf("CheckFile gives %v, %q and %v; want neither a breach nor a warning", shortReport(report), warningsOf(report), err)
	}
}

// Each number is named once, in the order of the text.
func TestNumbersSaid(t *testing.T) {
	for _, c := range []struct {
		numbers []string
		want    string
	}{
		{[]string{"7"}, `the number "7"`},
		{[]string{"1", "3", "3", "5"}, `the numbers "1", "3" and "5"`},
		{[]string{"60", "60"}, `the number "60"`},
	} {
		if got := numbersSaid(c.numbers); !strings.HasPrefix(got, "the text holds "+c.want+": ") {
			t.Errorf("the numbers %q are said %q, want it to begin with %q", c.numbers, got, c.want)
		}
	}
}

// No example of a manifest, a catalog of traps or a catalog of techniques holds a number in a text of
// the whole module, as the format's text asks (format/README.md, section 20; the 2026-11-08
// amendments to R57 in spec): the examples of whole modules are held by TestCheckAccepts of the CLI.
func TestExamplesWarnOfNoNumber(t *testing.T) {
	read := 0
	for _, part := range []string{partManifest, partTraps, partTechniques} {
		root := path.Join("examples", part, "valid")
		entries, err := fs.ReadDir(spec.Format, root)
		if err != nil {
			t.Fatal(err)
		}
		for _, e := range entries {
			data, err := fs.ReadFile(spec.Format, path.Join(root, e.Name()))
			if err != nil {
				t.Fatal(err)
			}
			value, err := yaml.Read(data)
			if err != nil {
				t.Fatal(err)
			}
			p := newParts(false)
			p.add(path.Join(root, e.Name()), part, value)
			for _, w := range p.warn() {
				t.Errorf("%s: %s", path.Join(root, e.Name()), w)
			}
			read++
		}
	}
	if read == 0 {
		t.Fatal("no example was read")
	}
}
