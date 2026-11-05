package template

import (
	"context"
	"io/fs"
	"strings"
	"testing"

	"go.starlark.net/starlark"

	"github.com/altnovatrail/spec"

	"github.com/altnovatrail/module-kit/internal/sandbox"
	"github.com/altnovatrail/module-kit/internal/yaml"
)

// expressionsOfTheFormat gives every expression of a template in spec.Format, its derived numbers' and
// its constraint, and every string of a task that a template holds.
func expressionsOfTheFormat(f *testing.F) (expressions, texts []string) {
	f.Helper()
	err := fs.WalkDir(spec.Format, ".", func(name string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() || !strings.HasSuffix(name, ".yaml") {
			return err
		}
		data, err := fs.ReadFile(spec.Format, name)
		if err != nil {
			return err
		}
		value, err := yaml.Read(data)
		if err != nil {
			return nil
		}
		var walk func(v any)
		walk = func(v any) {
			switch v := v.(type) {
			case map[string]any:
				if t, ok := v["template"].(map[string]any); ok {
					if d, ok := t["derived"].(map[string]any); ok {
						for _, e := range d {
							if s, ok := e.(string); ok {
								expressions = append(expressions, s)
							}
						}
					}
					if c, ok := t["constraint"].(string); ok {
						expressions = append(expressions, c)
					}
				}
				for _, member := range v {
					walk(member)
				}
			case []any:
				for _, element := range v {
					walk(element)
				}
			case string:
				if strings.ContainsAny(v, "{}") {
					texts = append(texts, v)
				}
			}
		}
		walk(value)
		return nil
	})
	if err != nil {
		f.Fatal(err)
	}
	return expressions, texts
}

// On any text, typeOf gives a type or an error, and never panics; and an expression that it takes,
// on the names a and b, go.starlark.net at the version of the format takes too, but for one that
// begins with a space or nests past the depth of its parser, and evaluates on any two ints to a value
// of the type that typeOf gives, or fails only by dividing by zero. So the grammar and the types of
// the subset are those of Starlark, which gives an expression its meaning (format/README.md, section
// 23), and the two ways go.starlark.net parses less are all (В28). The corpus is every expression of
// a template in spec.Format.
func FuzzTypeOf(f *testing.F) {
	expressions, _ := expressionsOfTheFormat(f)
	for _, e := range expressions {
		f.Add(e, int64(3), int64(-2))
	}
	f.Add("-a // b", int64(-7), int64(2))
	f.Add("b != 0 and a % b == 0", int64(5), int64(0))
	f.Add("not a < b or a == b", int64(1), int64(1))
	// The two ways go.starlark.net parses less than the subset.
	f.Add(" a + b", int64(1), int64(2))
	f.Add(strings.Repeat("(", 76)+"a"+strings.Repeat(")", 76), int64(1), int64(2))
	f.Fuzz(func(t *testing.T, src string, a, b int64) {
		got, err := typeOf(src, []string{"a", "b"})
		if err != nil {
			if got != "" || err.Error() == "" {
				t.Fatalf("%q: %q, %v", src, got, err)
			}
			return
		}
		fn, err := sandbox.Compile(src, "a", "b")
		if err != nil {
			if strings.HasPrefix(src, " ") && strings.Contains(err.Error(), "got indent") || strings.Contains(err.Error(), "excessive nesting") {
				return
			}
			t.Fatalf("%q is %s expression of the subset, and go.starlark.net does not take it: %v", src, got, err)
		}
		v, err := fn.Call(context.Background(), sandbox.Limits{}, starlark.MakeInt64(a), starlark.MakeInt64(b))
		if err != nil {
			if !strings.Contains(err.Error(), "by zero") {
				t.Fatalf("%q on %d and %d: %v", src, a, b, err)
			}
			return
		}
		if _, isInt := v.(starlark.Int); isInt != (got == intType) {
			t.Fatalf("%q is %s expression, and on %d and %d it gives the %s %v", src, got, a, b, v.Type(), v)
		}
		if _, isBool := v.(starlark.Bool); isBool != (got == boolType) {
			t.Fatalf("%q is %s expression, and on %d and %d it gives the %s %v", src, got, a, b, v.Type(), v)
		}
	})
}

// written writes pieces back as a string of a template: each brace of the text doubled, and each
// placeholder as {name}.
func written(pieces []piece) string {
	var b strings.Builder
	for _, p := range pieces {
		if p.placeholder != "" {
			b.WriteString("{" + string(p.placeholder) + "}")
			continue
		}
		b.WriteString(strings.NewReplacer("{", "{{", "}", "}}").Replace(p.text))
	}
	return b.String()
}

// On any string, piecesOf gives pieces or an error, and never panics; pieces that it gives, written
// back, are the string, no two pieces of text stand side by side and none is empty, and every
// placeholder names a name. The corpus is every string of spec.Format that holds a brace.
func FuzzPiecesOf(f *testing.F) {
	_, texts := expressionsOfTheFormat(f)
	for _, s := range texts {
		f.Add(s)
	}
	f.Add("as {{{a}, {b}}}.")
	f.Fuzz(func(t *testing.T, s string) {
		pieces, err := piecesOf(s)
		if err != nil {
			if pieces != nil || !strings.Contains(err.Error(), " at ") {
				t.Fatalf("%q: %v, %v", s, pieces, err)
			}
			return
		}
		if w := written(pieces); w != s {
			t.Fatalf("%q reads into %v, which is written back %q", s, pieces, w)
		}
		for i, p := range pieces {
			if p.placeholder == "" && (p.text == "" || i > 0 && pieces[i-1].placeholder == "") {
				t.Fatalf("%q reads into %v", s, pieces)
			}
			if p.placeholder != "" && !templateName(string(p.placeholder)) {
				t.Fatalf("%q reads into %v", s, pieces)
			}
		}
	})
}
