# module-kit — measurements

> The results of the measurements of `module-kit`: what was measured, how, on what, and what numbers came out. The documents that describe how things are built — [architecture.md](architecture.md) and the decision log — refer here as to a section: `measurements 1.1` (R143 spec). An external fact a measurement rests on is a record of [sources.md](sources.md).
>
> Until T43 these measurements stood as items in [architecture.md](architecture.md), section 4, among the sources. The measurement of the tokens spent by an epic's sessions goes here when the epic is archived, as a row of a table on the model of measurements 1.9 spec (RUN.md, Epics; section 2).

## 1. Code and libraries

The measurements that the decisions on the Starlark sandbox, the YAML reader and the embedding of the format rest on.

### 1.1. The measurement of the sandbox

For R11 — 2026-11-02, Go 1.27.3, AMD Ryzen AI 9 365, 20 threads. The tests `TestEveryReferenceTaskProvesItsOwnAnswer` and `TestEverySolverTemplateProvesTheTaskItWasWrittenFor` of standalone — a clone at commit 89986bc in a temporary folder — ran on its go.starlark.net and with `go mod edit -replace` on 4f4267fb140d. Our own bench — in the session's folder, on both versions. The figures are R11's.

The results as R11 recorded them; an entry of the log is not rewritten, and here is their copy (T43):

- The solver tests of standalone — 1410 runs of reference tasks and 66 runs of templates — on its version and with `replace` to the latest:

  | | 89a6a09411d5 | 4f4267fb140d |
  |---|---|---|
  | the heaviest solver | 2,413,617 steps, `pig-34-d4-2` | 2,725,355 steps, `pig-34-d4-2` |
  | the heaviest template | 959,962 steps | 949,969 steps |
  | the longest run | 25.8 ms | 25.3 ms |
  | all steps together | 94,732,336 | 77,489,284 |

  Over 738 solvers the ratio of steps new/old has a median of 0.96, and for 80% it is from 0.59 to 1.53. More than 2× only for solvers of about 50 steps.
- Our own bench on both versions:

  | Program | 89a6a09411d5 | 4f4267fb140d |
  |---|---|---|
  | a turn of an empty `for` loop in a function | 6 steps | 10 steps |
  | one call of `solve` of the vector `past-the-budget-of-the-template` | 60,000,039 steps | 100,000,040 steps |
  | a predicate, 1,000,000 steps of brute force | 4.2 ms | 6.0 ms |
  | a predicate with a brute force of divisors up to 100,000 | 1,499,987 steps | 1,000,010 steps |
  | a simple predicate — `value % 3 == 0 and value % 4 == 0` | 19 steps | 20 steps |
  | a loop, 100,000,000 steps | 650 ms | 445 ms |
  | `len(list(range(50000000)))` with a clock of 100 ms | 763 MiB in 7 steps; a stop 217 ms after the clock | 763 MiB in 30 steps; after 209 ms |
  | `set([1]).difference()` | a panic: nil pointer dereference | the error "got 0 arguments, want 1" |
  | `(lambda f: f(f))(lambda f: f(f))` with `Recursion: false` | "function lambda called recursively" | the same |

  On the latest version a predicate that calls `sorted(range(1000))` on each of 100,000 turns fits in 1,000,000 steps in 588 ms. A solver of the same kind reaches 93,898,910 steps in 60 s, and the clock stops it. `str` of a list 100,000 deep takes 1.6 s on one step, and of one 1,000,000 deep more than 120 s, and the process does not crash, but the clock does not interrupt it either.

### 1.2. Probes of YAML readers

For R21 — 2026-11-04, in the session's folder, over 2302 YAML files of `spec.Format` v0.9.2 outside `vectors/rules/yaml.*`:
- `go.yaml.in/yaml/v3` v3.0.6 (`@latest` of 2026-10-13) refuses 195, all with "did not find expected ',' or ']'" on the line `rung_templates: [Which step of your work gave you this answer?, …]`; on `\/` — "found unknown escape character";
- `github.com/goccy/go-yaml` v1.19.2 (`@latest` of 2026-01-08, MIT) reads all, but on `? {a: b}` and `? [a, b]` answers "found an invalid key for this map", reads a Latin-1 file, and counts a directive and a file of a single comment as documents;
- `go.yaml.in/yaml/v4` v4.0.0-rc.9 reads all. On edge inputs: `%YAML 1.2` — "found incompatible YAML document"; `a: x<U+2028>  y` — `"x\u2028y"`; `\ud800` — "found invalid Unicode character escape code" with `ContextMark` at the opening quote; an alias without an anchor — "unknown anchor 'x' referenced"; a control character — "control characters are not allowed" without a place; the tag `!` — `Tag == "!"` without `TaggedStyle`; an implicit key of 1100 characters it accepts.

### 1.3. The size of `spec.Format`

For R23 — `find spec/format -type f -printf '%s\n'` at v0.9.2, 2026-11-04: `format/` — 5,682,239 bytes in 2821 files, of them `vectors/versions/` — 4,072,532, and `schemas/` — 111,654 in 20; `admission/` — 66,210. `spec.go` at the same tag — `//go:embed all:format all:admission` into one `embed.FS`, from which `fs.Sub` gives `Format` and `Admission`.

## 2. Claude Code sessions

How `module-kit` keeps what R143 spec and R145 spec saved, by the pattern of measurements 1.9 spec (R55): budgets of size in `docs-check` on every pull request, by the block `repository` of `spec/tools/docs-check/budgets.json` and [docs-budgets.json](../docs-budgets.json), and a measurement of the sessions at each epic's archive (RUN.md, Epics).

### 2.1. The measurement of sessions

`just -f spec/justfile tokens-report --since … --row` prints one row of the table below, for the epic's window, at its archive. The first row is epic E3, the first archived after RUN.md, Epics, named the measurement; E1 and E2 were archived before it.

| Project | Window, UTC | Sessions | Calls, main / subagents | Context of a call, k, main / subagents | Compactions | Cost at API prices |
|---|---|---|---|---|---|---|
| `-workspaces-module-kit` — epic E3: T45, T46, measured in T46 before its pull request | 2026-11-16T00:00 … 2026-11-16T03:26 | 2 | 134 / 47 | 120.7 / 60.1 | 0 | $9.15 |

## 3. The speed of CI

What R59 changed in CI, measured before and after, by the pattern of measurements 3.1 spec.

### 3.1. The parallel builds and tests of T47, 2026-11-16

- **Locally, not as a comparison of times.** The devcontainer on a host of 20 cores whose load average was 63–79 while it ran, from the other containers on it: one build of `release-files v0.0.0-ci` with an empty `GOCACHE` took 52.6–81.1 s one build after another and 58.5–81.0 s at once on four cores (`taskset -c 0-3`), 68.6–92.8 s and 32.7–33.7 s on twenty, two runs each — too unsteady to give a gain. Both gave `SHA256SUMS` with the same sha256, `e9b0a52b8d56…`.
- **The clock of a predicate.** `go test -race ./internal/vectors` with `t.Parallel()` in every case of `TestVectors` failed `format/vectors/verdicts/predicate/steps-below-the-limit.yaml`: "unchecked, and the vector gives correct", in 1.04 s, over the clock of a second of `sandbox.Predicate`. Alone, under `-race`, the case took 0.72 s and `steps-reach-the-limit.yaml` 0.66 s, the two other predicate cases 0.15 s. With the cases of `verdicts` one at a time, `just ci-test` passed six times of six on four cores and eight of nine on twenty; the one failure, on twenty, was run with its output thrown away, and its cause is not known. On four cores, with `TestChecksTheFormat` of `module` running its parallel subtests beside it, `steps-below-the-limit.yaml` took 0.31–0.42 s and `steps-reach-the-limit.yaml` 0.24–0.26 s, five runs each, built as `ci-test` builds them, with `-covermode=atomic -coverpkg=./...`. That these are shorter than the 0.72 s alone is not explained by the runs themselves, taken minutes apart on the same busy host: the host's load is the likely difference, and the margin on a runner of CI is not measured (Д20 in [run/DEBT.md](../run/DEBT.md)).
