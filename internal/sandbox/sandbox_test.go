package sandbox

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"maps"
	"os"
	"regexp"
	"slices"
	"sync"
	"testing"
	"time"

	"github.com/altnovatrail/spec"
	"go.starlark.net/starlark"
	"go.starlark.net/syntax"
)

// formatText gives the text of the format, format/README.md.
func formatText(t *testing.T) string {
	t.Helper()
	data, err := fs.ReadFile(spec.Format, "README.md")
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

// The dialect is the format's: the options of section 24, and as the built-in names the 33 that it
// lists, in its order, which are those of starlark.Universe at the version that go.mod pins.
func TestDialect(t *testing.T) {
	want := syntax.FileOptions{Set: true, While: true, TopLevelControl: true, GlobalReassign: true}
	if got := *options(); got != want {
		t.Errorf("the options are %+v, want %+v", got, want)
	}
	if universe, names := slices.Sorted(maps.Keys(starlark.Universe)), slices.Sorted(slices.Values(builtins)); !slices.Equal(universe, names) {
		t.Errorf("starlark.Universe holds %q, and the dialect %q", universe, names)
	}
	list := regexp.MustCompile("the 33 names of that version's `starlark.Universe`: ([^.]*)\\.").FindStringSubmatch(formatText(t))
	if list == nil {
		t.Fatal("format/README.md lists no built-in names")
	}
	var text []string
	for _, name := range regexp.MustCompile("`([A-Za-z]+)`").FindAllStringSubmatch(list[1], -1) {
		text = append(text, name[1])
	}
	if !slices.Equal(text, builtins) {
		t.Errorf("format/README.md lists %q, and the dialect %q", text, builtins)
	}
}

// go.mod pins go.starlark.net at the version that the format names (format/README.md, section 24;
// R88 spec).
func TestVersionIsTheFormats(t *testing.T) {
	named := regexp.MustCompile("go\\.starlark\\.net at the version `([^`]+)`").FindStringSubmatch(formatText(t))
	if named == nil {
		t.Fatal("format/README.md names no version of go.starlark.net")
	}
	goMod, err := os.ReadFile("../../go.mod")
	if err != nil {
		t.Fatal(err)
	}
	pinned := regexp.MustCompile(`(?m)^\s*go\.starlark\.net (\S+)$`).FindSubmatch(goMod)
	if pinned == nil || string(pinned[1]) != named[1] {
		t.Errorf("go.mod pins go.starlark.net at %q, and the format names %s", pinned, named[1])
	}
}

// CheckExpr takes an expression of the dialect on its names, the names it binds itself and the
// built-in names, and nothing else.
func TestCheckExpr(t *testing.T) {
	for _, c := range []struct{ src, want string }{
		{"value % 3 == 0 and value % 4 == 0", ""},
		{"value in [n * n for n in range(10)]", ""},
		{"(lambda x, y: x < y)(0, value)", ""},
		{"value in set([1, 2]) and {1: value}[1] == value", ""},
		{"value > 0 # a comment", ""},
		{"(value\n> 0)", ""},
		{"value, value", ""},
		{"value = 12", "1:8: got '=' after expression, want EOF"},
		{"value\n> 0", "2:2: got '>' after expression, want EOF"},
		{"if value: True", "1:1: got if, want primary expression"},
		{"", "1:1: got end of file, want primary expression"},
		{"number % 12 == 0", "1:1: undefined: number"},
		{"n * n for n in range(10)", "1:10: got for after expression, want EOF"},
		{"[n for n in range(10)] and n", "1:28: undefined: n"},
		{"permutations([value])", "1:1: undefined: permutations"},
		{"json.encode(value)", "1:1: undefined: json"},
		{"value.bit_length() > 2", ""},
	} {
		err := CheckExpr(c.src, "value")
		if got := fmt.Sprint(err); err == nil && c.want != "" || err != nil && got != c.want {
			t.Errorf("CheckExpr(%q) gives %v, want %q", c.src, err, c.want)
		}
	}
	for _, name := range builtins {
		if err := CheckExpr(name); err != nil {
			t.Errorf("the built-in name %s: %v", name, err)
		}
	}
}

// Eval gives what the expression gives on env, or its error: as it runs, or as it is none of the
// dialect, which a name of starlark.Universe beyond the dialect is not either.
func TestEval(t *testing.T) {
	starlark.Universe["added"] = starlark.True
	t.Cleanup(func() { delete(starlark.Universe, "added") })
	for _, c := range []struct {
		src  string
		env  starlark.StringDict
		want string
	}{
		{"value % 3 == 0", starlark.StringDict{"value": starlark.MakeInt(12)}, "True"},
		{"value + 1", starlark.StringDict{"value": starlark.MakeInt(12)}, "13"},
		{"chr(ord('a')) + str(len(sorted(set([3, 1, 3]))))", nil, `"a2"`},
		{"12 % value == 0", starlark.StringDict{"value": starlark.MakeInt(0)}, "error: sandbox: integer modulo by zero"},
		{"(lambda f: f(f))(lambda f: f(f))", nil, "error: sandbox: function lambda called recursively"},
		{"fail('no')", nil, "error: sandbox: fail: no"},
		{"value", nil, "error: sandbox: 1:1: undefined: value"},
		{"value = 1", nil, "error: sandbox: 1:8: got '=' after expression, want EOF"},
		{"added", nil, "error: sandbox: 1:1: undefined: added"},
	} {
		got, err := Eval(context.Background(), Predicate, c.src, c.env)
		text := fmt.Sprint(got)
		if err != nil {
			text = "error: " + err.Error()
		}
		if text != c.want {
			t.Errorf("Eval(%q) gives %s, want %s", c.src, text, c.want)
		}
	}
}

// The count of steps stops an evaluation when it reaches the limit, as go.starlark.net stops it:
// the predicates of the vectors steps-below-the-limit.yaml and steps-reach-the-limit.yaml take
// 999 990 steps and would take 1 000 000.
func TestSteps(t *testing.T) {
	for _, c := range []struct {
		src   string
		steps uint64
		err   error
	}{
		{"value in [n * n for n in range(99997)]", 999_990, nil},
		{"value in [n * n for n in range(99998)]", 1_000_000, ErrSteps},
	} {
		var steps uint64
		_, err := run(context.Background(), Predicate, func(thread *starlark.Thread) (starlark.Value, error) {
			defer func() { steps = thread.Steps }()
			expr, err := checked(c.src, []string{"value"})
			if err != nil {
				return nil, err
			}
			return starlark.EvalExprOptions(options(), thread, expr, starlark.StringDict{"value": starlark.MakeInt(49)})
		})
		if !errors.Is(err, c.err) || steps != c.steps {
			t.Errorf("%q: %d steps and %v, want %d and %v", c.src, steps, err, c.steps, c.err)
		}
	}
	if _, err := Eval(context.Background(), Limits{Steps: 1000}, endless, nil); err == nil || err.Error() != "sandbox: "+ErrSteps.Error() {
		t.Errorf("an endless loop at 1000 steps gives %v", err)
	}
}

// endless is a loop that keeps no list and has no end of its own within a test: only a limit or
// ctx stops it.
const endless = "[n for n in range(1 << 62) if False]"

// The clock stops an evaluation that runs too long, and so does ctx, which gives its own error;
// with neither, the count of steps does.
func TestClockAndCancel(t *testing.T) {
	start := time.Now()
	if _, err := Eval(context.Background(), Limits{Clock: 20 * time.Millisecond}, endless, nil); !errors.Is(err, ErrClock) {
		t.Errorf("the clock gives %v", err)
	}
	if took := time.Since(start); took > 10*time.Second {
		t.Errorf("the clock of 20 ms stopped the evaluation after %v", took)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	if _, err := Eval(ctx, Limits{Clock: time.Hour}, endless, nil); !errors.Is(err, context.DeadlineExceeded) || errors.Is(err, ErrClock) {
		t.Errorf("a deadline gives %v", err)
	}
	cancelled, cancel := context.WithCancel(context.Background())
	cancel()
	ran := false
	_, err := run(cancelled, Predicate, func(*starlark.Thread) (starlark.Value, error) { ran = true; return starlark.None, nil })
	if !errors.Is(err, context.Canceled) || ran {
		t.Errorf("a context done before gives %v, and the evaluation ran: %t", err, ran)
	}
	if err.Error() != "sandbox: context canceled" {
		t.Errorf("a context done before says %q", err)
	}
}

// The first cause to stop the thread is the one that the error names: ctx done after the clock
// does not change it. A cause that comes once the evaluation has failed by itself does not
// change its error either.
func TestFirstCauseWins(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	_, err := run(ctx, Limits{Clock: time.Nanosecond}, func(thread *starlark.Thread) (starlark.Value, error) {
		time.Sleep(10 * time.Millisecond)
		cancel()
		expr, err := parse("None")
		if err != nil {
			return nil, err
		}
		return starlark.EvalExprOptions(options(), thread, expr, nil)
	})
	if !errors.Is(err, ErrClock) || errors.Is(err, context.Canceled) {
		t.Errorf("the clock and then ctx give %v", err)
	}
	ctx, cancel = context.WithCancel(context.Background())
	defer cancel()
	_, err = run(ctx, Predicate, func(thread *starlark.Thread) (starlark.Value, error) {
		expr, err := parse("1 // 0")
		if err != nil {
			return nil, err
		}
		result, err := starlark.EvalExprOptions(options(), thread, expr, nil)
		cancel()
		time.Sleep(10 * time.Millisecond)
		return result, err
	})
	if err == nil || err.Error() != "sandbox: floored division by zero" {
		t.Errorf("an error and then ctx give %v", err)
	}
}

// A panic of the interpreter, or of a built-in function it calls, is recovered: the evaluation
// stops with ErrPanic, and the caller goes on.
func TestPanic(t *testing.T) {
	boom := starlark.NewBuiltin("boom", func(*starlark.Thread, *starlark.Builtin, starlark.Tuple, []starlark.Tuple) (starlark.Value, error) {
		panic("boom")
	})
	_, err := Eval(context.Background(), Predicate, "[boom() for n in range(3)]", starlark.StringDict{"boom": boom})
	if !errors.Is(err, ErrPanic) || err.Error() != "sandbox: the interpreter panicked: boom" {
		t.Errorf("a panic gives %v", err)
	}
	if got, err := Eval(context.Background(), Predicate, "1 + 1", nil); err != nil || got != starlark.MakeInt(2) {
		t.Errorf("after a panic, 1 + 1 gives %v and %v", got, err)
	}
}

// print writes nothing: not to the standard error, where go.starlark.net writes it by default.
func TestPrintWritesNothing(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	stderr := os.Stderr
	os.Stderr = w
	got, err := Eval(context.Background(), Predicate, "print('written') == None", nil)
	os.Stderr = stderr
	w.Close()
	written, _ := io.ReadAll(r)
	if err != nil || got != starlark.True || len(written) != 0 {
		t.Errorf("print gives %v and %v, and writes %q", got, err, written)
	}
}

// Evaluations share nothing: many at once each give their own result.
func TestConcurrentEval(t *testing.T) {
	var wg sync.WaitGroup
	errs := make([]error, 32)
	for i := range errs {
		wg.Go(func() {
			got, err := Eval(context.Background(), Predicate, "len([n for n in range(value)])", starlark.StringDict{"value": starlark.MakeInt(i * 1000)})
			if err == nil && got != starlark.MakeInt(i*1000) {
				err = fmt.Errorf("evaluation %d gives %v", i, got)
			}
			errs[i] = err
		})
	}
	wg.Wait()
	for _, err := range errs {
		if err != nil {
			t.Error(err)
		}
	}
}

// FuzzCheckExpr checks any text: without a panic, the same way twice, and a text that it takes is
// an expression that the parser of the dialect takes. It evaluates nothing: a built-in function
// takes a gigabyte in a step, which no limit of the sandbox stops (RUN.md, Д10).
func FuzzCheckExpr(f *testing.F) {
	for _, src := range []string{
		"value % 3 == 0 and value % 4 == 0", "value in [n * n for n in range(99997)]", "12 % value == 0",
		"value = 12", "number % 12 == 0", "(lambda f: f(f))(lambda f: f(f))", "{1: 2}", "'\\x00'", "value if value else -value",
	} {
		f.Add(src)
	}
	f.Fuzz(func(t *testing.T, src string) {
		err := CheckExpr(src, "value")
		if again := CheckExpr(src, "value"); fmt.Sprint(again) != fmt.Sprint(err) {
			t.Fatalf("CheckExpr gives %v, and then %v", err, again)
		}
		if _, parseErr := options().ParseExpr("", src, 0); err == nil && parseErr != nil {
			t.Fatalf("CheckExpr takes what the parser does not: %v", parseErr)
		}
	})
}
