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

// Compile compiles an expression once on its names, in their order, and Call evaluates it on each
// binding of them, with the steps that Eval counts: the predicates of the vectors at the limit stop
// as they stop in Eval. A name that the program would assign is no name of the expression.
func TestCompile(t *testing.T) {
	ints := func(ns ...int) []starlark.Value {
		values := make([]starlark.Value, len(ns))
		for i, n := range ns {
			values[i] = starlark.MakeInt(n)
		}
		return values
	}
	for _, c := range []struct {
		src    string
		names  []string
		limits Limits
		args   []starlark.Value
		want   string
		err    string
	}{
		{"a * b + 1", []string{"a", "b"}, Limits{}, ints(6, 7), "43", ""},
		{"a - b", []string{"b", "a"}, Limits{}, ints(1, 10), "9", ""},
		{"_result + 1", []string{"_result"}, Limits{}, ints(1), "2", ""},
		{"_result + _result_", []string{"_result", "_result_"}, Limits{}, ints(1, 2), "3", ""},
		{"(a +\n b) # a comment", []string{"a", "b"}, Limits{}, ints(1, 2), "3", ""},
		{"\na", []string{"a"}, Limits{}, ints(1), "1", ""},
		{"# a comment\na", []string{"a"}, Limits{}, ints(1), "1", ""},
		{"a, b", []string{"a", "b"}, Limits{}, ints(1, 2), "(1, 2)", ""},
		{"a // b", []string{"a", "b"}, Limits{}, ints(1, 0), "", "sandbox: floored division by zero"},
		{"value in [n * n for n in range(99997)]", []string{"value"}, Predicate, ints(49), "True", ""},
		{"value in [n * n for n in range(99998)]", []string{"value"}, Predicate, ints(49), "", "sandbox: " + ErrSteps.Error()},
		{"a + b", []string{"a", "b"}, Limits{}, ints(1), "", "sandbox: 1 values for the 2 names [a b]"},
	} {
		f, err := Compile(c.src, c.names...)
		if err != nil {
			t.Fatalf("%q: %v", c.src, err)
		}
		got, err := f.Call(context.Background(), c.limits, c.args...)
		if c.err == "" && (err != nil || got.String() != c.want) || c.err != "" && (err == nil || err.Error() != c.err) {
			t.Errorf("%q on %v: %v, %v; want %s, %q", c.src, c.args, got, err, c.want, c.err)
		}
	}
	for _, c := range []struct{ src, err string }{
		{"c", "1:1: undefined: c"},
		{" a", "1:2: got indent, want primary expression"},
		{"a = 1", "1:4: got '=' after expression, want EOF"},
	} {
		if _, err := Compile(c.src, "a"); err == nil || err.Error() != c.err {
			t.Errorf("%q: %v, want %q", c.src, err, c.err)
		}
	}
	f, err := Compile(endless)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	if _, err := f.Call(ctx, Limits{}); !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("a deadline gives %v", err)
	}
}

// One Func is called from many goroutines at once, each with its own binding.
func TestConcurrentCall(t *testing.T) {
	f, err := Compile("a * a", "a")
	if err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	for n := range 64 {
		wg.Go(func() {
			got, err := f.Call(context.Background(), Limits{}, starlark.MakeInt(n))
			if err != nil || got.String() != fmt.Sprint(n*n) {
				t.Errorf("%d: %v, %v", n, got, err)
			}
		})
	}
	wg.Wait()
}

// A program of the dialect parses, holds no load, wherever it stands, and names only what it binds
// and the built-in names of the dialect: not one that a process adds to starlark.Universe.
func TestCheckProgram(t *testing.T) {
	starlark.Universe["gcd"] = starlark.NewBuiltin("gcd", nil)
	defer delete(starlark.Universe, "gcd")
	for _, c := range []struct{ src, err string }{
		{"def solve(numbers):\n    return {\"factor\": max(numbers.values())}\n", ""},
		{"x = [n for n in range(3)]\nx = set(x)\nfor n in x:\n    pass\n", ""},
		{"", ""},
		{"def solve(numbers)\n    return 1\n", "2:1: got newline, want ':'"},
		{"load(\"factors.star\", \"factor_of\")\n", "1:1: the dialect has no load"},
		{"def solve(numbers):\n    load(\"f.star\", \"f\")\n", "2:5: the dialect has no load"},
		{"def solve(numbers):\n    return gcd(1, 2)\n", "2:12: undefined: gcd"},
		{"def solve(numbers):\n    return solve2(numbers)\n", "2:12: undefined: solve2 (did you mean solve?)"},
	} {
		if err := CheckProgram(c.src); c.err == "" && err != nil || c.err != "" && (err == nil || err.Error() != c.err) {
			t.Errorf("%q: %v, want %q", c.src, err, c.err)
		}
	}
	if _, err := Exec(context.Background(), Limits{}, "def solve(numbers)\n"); err == nil || err.Error() != CheckProgram("def solve(numbers)\n").Error() {
		t.Errorf("Exec of no program gives %v", err)
	}
}

// solveTurns is a program whose solve(n) takes 20 + 10·n steps, as go.starlark.net at the version of
// the format counts them — 10 for the call, 10 for range and 10 for each turn —, after the 10 of its
// top level.
const solveTurns = "def solve(n):\n    for turn in range(n):\n        pass\n    return n\n"

// A program runs its top level once, and its functions are called on its thread: the steps of the top
// level and of every call count together, and the limit stops a call when the count reaches it, as
// the vectors within-the-budget and reaches-the-budget hold it at 100 000 000.
func TestProgramSteps(t *testing.T) {
	for _, c := range []struct {
		calls []int
		steps uint64
		err   error
	}{
		{[]int{0, 1, 3}, 110, nil},      // 10, then 30, 60 and 110: 10 below the limit of 120
		{[]int{0, 1, 4}, 120, ErrSteps}, // a turn more, and the count reaches it
		{[]int{100}, 120, ErrSteps},     // and so within one call
	} {
		p, err := Exec(context.Background(), Limits{Steps: 120}, solveTurns)
		if err != nil {
			t.Fatal(err)
		}
		if p.thread.starlark.Steps != 10 {
			t.Errorf("the top level takes %d steps", p.thread.starlark.Steps)
		}
		solve, _ := p.Global("solve")
		for _, n := range c.calls {
			var got starlark.Value
			got, err = p.Call(solve, starlark.MakeInt(n))
			if err == nil && got != starlark.MakeInt(n) {
				t.Errorf("solve(%d) gives %v", n, got)
			}
		}
		if !errors.Is(err, c.err) || p.thread.starlark.Steps != c.steps {
			t.Errorf("%v: %d steps and %v, want %d and %v", c.calls, p.thread.starlark.Steps, err, c.steps, c.err)
		}
		p.Close()
		p.Close()
		// After Close the program has no clock and no watch of ctx, so it is called no more.
		if _, err := p.Call(solve, starlark.MakeInt(0)); !errors.Is(err, ErrClosed) || err.Error() != "sandbox: the program is closed" {
			t.Errorf("a call after Close gives %v", err)
		}
	}
	if _, err := Exec(context.Background(), Limits{Steps: 1000}, "while True:\n    pass\n"); !errors.Is(err, ErrSteps) || err.Error() != "sandbox: "+ErrSteps.Error() {
		t.Errorf("an endless top level gives %v", err)
	}
}

// The globals of a program are frozen once its top level has run, so a call changes nothing that the
// next one reads; a call's own values are not. A global that the program binds is there, and one that
// it does not is not.
func TestProgramGlobals(t *testing.T) {
	p, err := Exec(context.Background(), Limits{}, "seen = []\nlimit = 3\ndef solve(n):\n    seen.append(n)\n\ndef fresh(n):\n    mine = []\n    mine.append(n)\n    return mine\n")
	if err != nil {
		t.Fatal(err)
	}
	defer p.Close()
	if v, ok := p.Global("limit"); !ok || v != starlark.MakeInt(3) {
		t.Errorf("limit is %v, %t", v, ok)
	}
	if v, ok := p.Global("solve2"); ok || v != nil {
		t.Errorf("solve2 is %v, %t", v, ok)
	}
	solve, _ := p.Global("solve")
	if _, err := p.Call(solve, starlark.MakeInt(1)); err == nil || err.Error() != "sandbox: append: cannot append to frozen list" {
		t.Errorf("a call that changes a global gives %v", err)
	}
	fresh, _ := p.Global("fresh")
	if v, err := p.Call(fresh, starlark.MakeInt(1)); err != nil || v.String() != "[1]" {
		t.Errorf("a call that changes its own list gives %v, %v", v, err)
	}
	if _, err := Exec(context.Background(), Limits{}, "fail(\"at the top\")\n"); err == nil || err.Error() != "sandbox: fail: at the top" {
		t.Errorf("a top level that fails gives %v", err)
	}
}

// One clock runs for a program from Exec to Close, over its calls and between them, and ctx stops it
// with its own error, done before Exec or after.
func TestProgramClockAndCancel(t *testing.T) {
	p, err := Exec(context.Background(), Limits{Clock: 20 * time.Millisecond}, solveTurns)
	if err != nil {
		t.Fatal(err)
	}
	solve, _ := p.Global("solve")
	if _, err := p.Call(solve, starlark.MakeInt(1)); err != nil {
		t.Fatal(err)
	}
	time.Sleep(40 * time.Millisecond)
	if _, err := p.Call(solve, starlark.MakeInt(1)); !errors.Is(err, ErrClock) {
		t.Errorf("a call after the clock gives %v", err)
	}
	p.Close()
	if _, err := Exec(context.Background(), Limits{Clock: 20 * time.Millisecond}, "while True:\n    pass\n"); !errors.Is(err, ErrClock) {
		t.Errorf("an endless top level gives %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	p, err = Exec(ctx, Limits{Clock: time.Hour}, solveTurns)
	if err != nil {
		t.Fatal(err)
	}
	defer p.Close()
	solve, _ = p.Global("solve")
	cancel()
	// ctx stops the thread from a goroutine of its own, so the call ends once the interpreter sees it.
	if _, err := p.Call(solve, starlark.MakeInt(1<<62)); !errors.Is(err, context.Canceled) || errors.Is(err, ErrClock) {
		t.Errorf("a call after ctx is done gives %v", err)
	}
	if _, err := Exec(ctx, Limits{}, solveTurns); !errors.Is(err, context.Canceled) || err.Error() != "sandbox: context canceled" {
		t.Errorf("Exec with ctx done gives %v", err)
	}
	deadline, cancel := context.WithTimeout(context.Background(), 0)
	defer cancel()
	_, deadlineErr := Eval(deadline, Limits{}, "1", nil)
	for err, done := range map[error]bool{
		fmt.Errorf("sandbox: %w", context.Canceled): true, deadlineErr: true, ErrClock: false, ErrSteps: false, errors.New("context canceled"): false,
	} {
		if ContextDone(err) != done {
			t.Errorf("ContextDone(%v) is %t", err, !done)
		}
	}
}

// A panic of the interpreter, or of a built-in function that a call reaches, stops the call with
// ErrPanic, and the program is not called again.
func TestProgramPanic(t *testing.T) {
	calls := 0
	boom := starlark.NewBuiltin("boom", func(*starlark.Thread, *starlark.Builtin, starlark.Tuple, []starlark.Tuple) (starlark.Value, error) {
		calls++
		panic("boom")
	})
	p, err := Exec(context.Background(), Limits{}, solveTurns)
	if err != nil {
		t.Fatal(err)
	}
	defer p.Close()
	for range 2 {
		if _, err := p.Call(boom); !errors.Is(err, ErrPanic) || err.Error() != "sandbox: the interpreter panicked: boom" {
			t.Errorf("a panic gives %v", err)
		}
	}
	solve, _ := p.Global("solve")
	if _, err := p.Call(solve, starlark.MakeInt(1)); !errors.Is(err, ErrPanic) || calls != 1 {
		t.Errorf("after a panic, solve gives %v, and boom ran %d times", err, calls)
	}
}

// Programs share nothing: many at once, each on its own thread, count their own steps.
func TestConcurrentExec(t *testing.T) {
	var wg sync.WaitGroup
	for n := range 32 {
		wg.Go(func() {
			p, err := Exec(context.Background(), Limits{Steps: uint64(30 + 10*n + 1)}, solveTurns)
			if err != nil {
				t.Error(err)
				return
			}
			defer p.Close()
			solve, _ := p.Global("solve")
			if got, err := p.Call(solve, starlark.MakeInt(n)); err != nil || got != starlark.MakeInt(n) {
				t.Errorf("%d: %v, %v", n, got, err)
			}
		})
	}
	wg.Wait()
}

// FuzzCheckProgram checks any text: without a panic, the same way twice, and a text that it takes is
// a program that the parser of the dialect takes, with no load, and that compiles as Exec compiles it.
// It runs nothing: a built-in function takes a gigabyte in a step, and x = x + x doubles a string
// with no loop, which no limit of the sandbox stops (RUN.md, Д10).
func FuzzCheckProgram(f *testing.F) {
	for _, src := range []string{
		solveTurns, "def solve(numbers)\n    return 1\n", "load(\"f.star\", \"f\")\n", "def solve(numbers):\n    return gcd(1, 2)\n",
		"x = [1]\nx = set(x)\nif x:\n    y = {1: 2}\n", "def f():\n    return f()\n", "solve = lambda numbers: {\"factor\": 1}\n", "",
	} {
		f.Add(src)
	}
	f.Fuzz(func(t *testing.T, src string) {
		err := CheckProgram(src)
		if again := CheckProgram(src); fmt.Sprint(again) != fmt.Sprint(err) {
			t.Fatalf("CheckProgram gives %v, and then %v", err, again)
		}
		if err != nil {
			return
		}
		file, parseErr := options().Parse("", src, 0)
		if parseErr != nil {
			t.Fatalf("CheckProgram takes what the parser does not: %v", parseErr)
		}
		syntax.Walk(file, func(n syntax.Node) bool {
			if _, ok := n.(*syntax.LoadStmt); ok {
				t.Fatal("CheckProgram takes a load")
			}
			return true
		})
		if _, err := starlark.FileProgram(file, none); err != nil {
			t.Fatalf("CheckProgram takes what does not compile: %v", err)
		}
	})
}
