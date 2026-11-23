# module-kit — measurements

> The results of the measurements of `module-kit`: what was measured, how, on what, and what numbers came out. The documents that describe how things are built — [architecture.md](architecture.md) and the decision log — cite a result by its record's number, `measurements 1.1`. An external fact a measurement rests on is a record of [sources.md](sources.md).
>
> The measurement of the tokens spent by an epic's sessions goes here when the epic is archived, as a row of the table of Claude Code sessions below.

## 1. Code and libraries

The measurements that the decisions on the Starlark sandbox, the YAML reader, the embedding of the format and the open view rest on.

### 1.1. The measurement of the sandbox

2026-11-02, Go 1.27.3, AMD Ryzen AI 9 365, 20 threads. The tests `TestEveryReferenceTaskProvesItsOwnAnswer` and `TestEverySolverTemplateProvesTheTaskItWasWrittenFor` of standalone — a clone at commit 89986bc in a temporary folder — ran on its go.starlark.net and with `go mod edit -replace` on 4f4267fb140d. Our own bench — in the session's folder, on both versions. The figures are those the decision on Starlark recorded.

The results as the decision on Starlark recorded them, copied here:

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

2026-11-04, the same machine, with the sandbox of `module-kit` on 4f4267fb140d: a turn with `sorted(range(1000))` costs 20 steps, and `[n for n in range(100000) if sorted(range(1000)) == None]` reaches the limit of 1,000,000 at its 50,000th turn, after 932 ms, rather than fitting in it in 588 ms as above.

The same day, the same sandbox, a probe in the session's folder, with `value` = 49:

| Expression | Steps | Time | With `-race` |
|---|---|---|---|
| `value in [n * n for n in range(99997)]` | 999,990, the result `True` | 13–15 ms | not measured |
| `value in [n * n for n in range(99998)]` | a stop at 1,000,000; without a limit — exactly 1,000,000 | 12–13 ms | not measured |
| `[n for n in range(1 << 62) if False] == []` | a stop at 1,000,000, 10 steps per turn | 3.6 ms | 21 ms |
| `[n for n in range(100000) if sorted(range(1000)) == None] == []` | a stop at 1,000,000, 20 steps per turn: at the 50,000th turn | 932 ms | 8.6 s |

### 1.2. Probes of YAML readers

2026-11-04, in the session's folder, over 2302 YAML files of `spec.Format` v0.9.2 outside `vectors/rules/yaml.*`:
- `go.yaml.in/yaml/v3` v3.0.6 (`@latest` of 2026-10-13) refuses 195, all with "did not find expected ',' or ']'" on the line `rung_templates: [Which step of your work gave you this answer?, …]`; on `\/` — "found unknown escape character";
- `github.com/goccy/go-yaml` v1.19.2 (`@latest` of 2026-01-08, MIT) reads all, but on `? {a: b}` and `? [a, b]` answers "found an invalid key for this map", reads a Latin-1 file, and counts a directive and a file of a single comment as documents;
- `go.yaml.in/yaml/v4` v4.0.0-rc.9 reads all. On edge inputs: `%YAML 1.2` — "found incompatible YAML document"; `a: x<U+2028>  y` — `"x\u2028y"`; `\ud800` — "found invalid Unicode character escape code" with `ContextMark` at the opening quote; an alias without an anchor — "unknown anchor 'x' referenced"; a control character — "control characters are not allowed" without a place; the tag `!` — `Tag == "!"` without `TaggedStyle`; an implicit key of 1100 characters it accepts.

### 1.3. The size of `spec.Format`

`find spec/format -type f -printf '%s\n'` at v0.9.2, 2026-11-04: `format/` — 5,682,239 bytes in 2821 files, of them `vectors/versions/` — 4,072,532, and `schemas/` — 111,654 in 20; `admission/` — 66,210. `spec.go` at the same tag — `//go:embed all:format all:admission` into one `embed.FS`, from which `fs.Sub` gives `Format` and `Admission`.

### 1.4. The cost of the open view of a task

2026-11-04, `go test -bench`, AMD Ryzen AI 9 365: the open view of a task through `openview.Task`, together with the check against the schema of a task, — 0.30 ms for the task `zones-and-picture` of the vector of the open view, and 0.13 ms for the task `gaps-ladder-1` of the example `with-lessons` of `spec.Format` v0.9.2.

### 1.5. The cost of a template of 10,000 instances

2026-11-05, `go test`, AMD Ryzen AI 9 365: `module.CheckFile` checks a template of 10,000 instances — the example `missing-factor` with domains from 1 to 100 — in 1.5 s: the schema of the instances takes 0.6 s, the substitution and the round trips of JSON 0.5 s, and the count of the expressions in the sandbox less than 0.1 s. The template `heads-and-legs` of 81 combinations is counted in 2 ms.

### 1.6. The solver at the budget

2026-11-05, Go 1.27.3, AMD Ryzen AI 9 365, `go test`: a solver at the budget — the loop `for turn in range(n): pass` — takes 0.10–0.16 s per 100,000,000 steps, and the template `number-thought-of` of 35 instances 7–11 ms. Under `-race`, each package separately, with the solver in the code against `main` without it: `module` — 13.5 s against 9.2 s, `internal/vectors` — 9.1 s against 3.2 s; `FuzzCheck` with the vectors of `task.template-solver` in its corpus ran 10 s, and without them 4.3 s against 3.6 s. The whole `ci-test` took 23 s.

## 2. Claude Code sessions

How `module-kit` keeps the savings of a session's tokens, by the pattern of measurements 1.9 spec: budgets of size in `docs-check` on every pull request, by the block `repository` of `spec/tools/docs-check/budgets.json` and [docs-budgets.json](../docs-budgets.json), and a measurement of the sessions at each epic's archive.

### 2.1. The measurement of sessions

`just -f spec/justfile tokens-report --since … --row` prints one row of the table below, for the epic's window, at its archive. The first row is epic E3, the first archived after the rules of RUN.md named the measurement; E1 and E2 were archived before it. The last column, the context of a main call among its session's first 80, `tokens-report` prints since a later version of it, and the rows before it have "—" there.

| Project | Window, UTC | Sessions | Calls, main / subagents | Context of a call, k, main / subagents | Compactions | Cost at API prices | Context of a main call among its session's first 80, k |
|---|---|---|---|---|---|---|---|
| `-workspaces-module-kit` — epic E3, measured before its pull request | 2026-11-16T00:00 … 2026-11-16T03:26 | 2 | 134 / 47 | 120.7 / 60.1 | 0 | $9.15 | — |
| `-workspaces-module-kit` — epic E4, measured with the next epic, since its archive had not | 2026-11-16T03:27 … 2026-11-16T22:13 | 6 | 427 / 77 | 124.1 / 54.6 | 0 | $24.51 | — |
| `-workspaces-module-kit` — epic E5, measured before its pull request | 2026-11-16T22:13 … 2026-11-16T22:25 | 1 | 60 / 26 | 101.0 / 59.0 | 0 | $3.85 | — |
| `-workspaces-module-kit` — epic E6, measured before its pull request; the window from the end of E5 | 2026-11-16T22:26 … 2026-11-17T00:19 | 2 | 40 / 0 | 103.4 / — | 0 | $2.57 | — |
| `-workspaces-module-kit` — epic E7, measured before its pull request; the window from the end of E6 | 2026-11-17T00:20 … 2026-11-17T05:27 | 2 | 50 / 12 | 122.7 / 49.7 | 0 | $3.31 | — |
| `-workspaces-module-kit` — epic E8, measured before its pull request; the window from the end of E7 | 2026-11-17T05:28 … 2026-11-17T15:02 | 2 | 48 / 0 | 89.5 / — | 0 | $2.18 | — |
| `-workspaces-module-kit` — epic E9, measured before its pull request; the window from the end of E8 | 2026-11-17T15:03 … 2026-11-17T19:27 | 3 | 83 / 20 | 100.3 / 51.9 | 0 | $4.40 | — |
| `-workspaces-module-kit` — epic E10, measured before its pull request; the window from the end of E9 | 2026-11-17T19:28 … 2026-11-18T05:34 | 2 | 60 / 0 | 108.7 / — | 0 | $3.04 | 108.7 |
| `-workspaces-module-kit` — epic E11, measured before its review; the window from the end of E10 | 2026-11-18T05:35 … 2026-11-18T13:07 | 7 | 300 / 564 | 91.2 / 65.1 | 0 | $36.56 | 91.2 |
| `-workspaces-module-kit` — epic E12, measured before its pull request; the window from the end of E11 | 2026-11-18T13:08 … 2026-11-20T21:15 | 2 | 54 / 22 | 96.8 / 90.8 | 0 | $3.95 | 96.8 |
| `-workspaces-module-kit` — epic E13, measured before its pull request; the window from the end of E12 | 2026-11-20T21:16 … 2026-11-23T18:33 | 11 | 574 / 118 | 120.5 / 45.0 | 0 | $41.27 | 118.7 |

## 3. The speed of CI

What the changes of CI did to its speed, measured before and after, by the pattern of measurements 3.1 spec.

### 3.1. The parallel builds and tests, 2026-11-16

- **Locally, not as a comparison of times.** The devcontainer on a host of 20 cores whose load average was 63–79 while it ran, from the other containers on it: one build of `release-files v0.0.0-ci` with an empty `GOCACHE` took 52.6–81.1 s one build after another and 58.5–81.0 s at once on four cores (`taskset -c 0-3`), 68.6–92.8 s and 32.7–33.7 s on twenty, two runs each — too unsteady to give a gain. Both gave `SHA256SUMS` with the same sha256, `e9b0a52b8d56…`.
- **The clock of a predicate.** `go test -race ./internal/vectors` with `t.Parallel()` in every case of `TestVectors` failed `format/vectors/verdicts/predicate/steps-below-the-limit.yaml`: "unchecked, and the vector gives correct", in 1.04 s, over the clock of a second of `sandbox.Predicate`. Alone, under `-race`, the case took 0.72 s and `steps-reach-the-limit.yaml` 0.66 s, the two other predicate cases 0.15 s. With the cases of `verdicts` one at a time, `just ci-test` passed six times of six on four cores and eight of nine on twenty; the one failure, on twenty, was run with its output thrown away, and its cause is not known. With only `verdicts/predicate/` one at a time, as the change has it after the review, `go test -race` of `internal/vectors` and `module` passed three times of three on four cores, and CI twice. On four cores, with `TestChecksTheFormat` of `module` running its parallel subtests beside it, `steps-below-the-limit.yaml` took 0.31–0.42 s and `steps-reach-the-limit.yaml` 0.24–0.26 s, five runs each, built as `ci-test` builds them, with `-covermode=atomic -coverpkg=./...`. That these are shorter than the 0.72 s alone is not explained by the runs themselves, taken minutes apart on the same busy host: the host's load is the likely difference, and the margin on a runner of CI is not measured (Д20 in [run/DEBT.md](../run/DEBT.md)).
- **In CI.** The steps of the two changed jobs, from `actions/runs/<id>/jobs`, and the times of the two longest packages, from the log of "Go / test"; read 2026-11-16. Before — the eleven last successful runs of CI before the change, of `main` and of pull requests, 2026-11-12 … 2026-11-16, all of the same Go and the same tests but for the change's; after — the two runs of the change's pull request #79, 326a1be and 3f80b07. Standard runners `ubuntu-24.04`, four cores.

  | | Before, 11 runs: range, median | After, 2 runs |
  |---|---|---|
  | "Binaries / build", the step «Release files and samples» | 44–72 s, 52 s | 47 s, 70 s |
  | "Go / test", the step `ci-test` | 65–95 s, 94 s | 65 s, 66 s |
  | `internal/vectors` in `ci-test` | 28.8–44.2 s, 43.0 s | 22.1 s, 22.2 s |
  | `module` in `ci-test` | 40.4–57.2 s, 55.2 s | 37.5 s, 37.9 s |

  The runs before fall in two groups, the same runner being faster or slower: `internal/vectors` 28.8–32.5 s or 43.0–44.2 s. The two packages after are below either group; `ci-test` is at the faster group's floor, since it also compiles the module under `-race` and waits for `module`, whose `FuzzCheck` stays a test of one goroutine. The builds at once gave no gain that the spread shows: four cores are what one `go build` already keeps busy.

### 3.2. The tests at the clock of a predicate apart, 2026-11-16

- **Locally.** The devcontainer on a host of 20 cores whose load average was 74–109 while it ran, from the other containers on it; every run on four cores, `taskset -c 0-3`, with the flags of `ci-test` and `-v`, eight runs each.
  - Before, as the tests were then laid out — one `go test ./...`: 93–135 s, all green; the last six ran beside another `go test` on four other cores. `steps-below-the-limit.yaml` took 0.33–0.87 s of the clock of a second, `steps-reach-the-limit.yaml` 0.27–0.86 s, `TestConcurrentEval` of `internal/sandbox` 0.40–0.82 s and `TestCompile` 0.29–0.68 s.
  - The first layout of the change, the verdicts of a predicate alone after the rest, on the other four cores beside those runs: one run of three failed in `TestConcurrentEval` — 32 evaluations at once, each "sandbox: the clock stopped the evaluation", in 1.17 s. So the predicate cases were not the only tests at the clock, and the layout adopted takes `internal/sandbox` and `verdict` out too.
  - After, in the layout adopted, alone on the four cores: the rest 81–91 s, `internal/sandbox` and `verdict` one at a time 4 s, the verdicts of a predicate 2 s; all green. `steps-below-the-limit.yaml` took 0.26–0.32 s, `steps-reach-the-limit.yaml` 0.22–0.26 s, `TestConcurrentEval` 0.23–0.39 s and `TestCompile` 0.24–0.31 s.
  - The joined profile gives the same coverage as one run, function by function: `go tool cover -func` 98.1% for both, and no line of the two lists differs.
- **In CI.** The run of the change's pull request #80, a8a3231, read 2026-11-16 from `actions/jobs/<id>` and the log of "Go / test", a standard runner `ubuntu-24.04` of four cores: the step `ci-test` 71 s, against 65 s and 66 s on pull request #79 of the parallel tests (measurements 3.1); in it, `internal/vectors` 21.3 s and `module` 37.3 s in the first run, `internal/sandbox` 1.5 s and `verdict` 1.2 s one at a time, and the verdicts of a predicate 1.4 s, `steps-below-the-limit.yaml` 0.15 s and `steps-reach-the-limit.yaml` 0.12 s of the clock of a second. One run: the price of the two later runs is some 5 s, within the spread of measurements 3.1.

### 3.3. The shared checks v0.1.30, 2026-11-16

- **Locally.** The devcontainer on a host of 20 cores whose load average was 72–92 while it ran, from the other containers on it: `just release-files v0.0.0-ci` with `GOMODCACHE` and `GOCACHE` in an empty folder took 78 s, and again over the folder it had filled 3 s; both gave `SHA256SUMS` with the same sha256, `e9b0a52b8d56…`, that of measurements 3.1, and `just binaries-samples` over the same folder wrote its four samples. One run each.
- **In CI.** The jobs and their steps, from `actions/runs/<id>/jobs`, read 2026-11-16, standard runners `ubuntu-24.04`. Before — the four last runs of CI before the move, on `.github` v0.1.24: `main` 1e9c772 and eebc3e2, and pull request #80 of the tests at the clock apart, a8a3231 and 645319a; for "CodeQL / analyze (go)", the runs of `main` 1e9c772 and eebc3e2 and of 645319a. After — the first run of the move's pull request #81, 9950769. In it the image of `toolchain` came from the cache — «Cache restored from key: toolchain-Linux-X64-f85d62b0…», «Toolchain image: loaded from its file.» — in every job, from the entry of `main` that `documents.yml` of v0.1.24 had saved under the same path, 409,094,492 B of 2026-11-16 03:41 UTC (`gh cache list`); the caches of Go missed in every job — «Cache not found for input keys: go-Linux-X64-binaries-…», «Go caches: none in the cache; the job starts with empty ones.» —, since only a push to `main` saves them and none has run on v0.1.30 yet.

  | | Before, 4 runs: range | After, 1 run |
  |---|---|---|
  | "Binaries / build", the job | 77–98 s | 88 s |
  | — the step «Toolchain image» | 12–18 s | 12 s |
  | — the step «Release files and samples» | 47–64 s | 47 s |
  | "Go / test", the job | 103–141 s | 100 s |
  | — the step `ci-test` | 71–103 s | 72 s |
  | "Go / lint", the job | 55–133 s | 43 s |
  | "Go / vuln", the job | 48–62 s | 44 s |
  | "Go / licenses", the job | 43–68 s | 67 s |
  | "Documents / docs-check-repo", the job | 43–61 s | 49 s |
  | "Go / toolchain", the job, new | — | 10 s |
  | "CodeQL / analyze (go)", the job, 3 runs before | 76–87 s | 50 s |
  | — the step «Go of the caller's toolchain stage» | 23–29 s | 2 s |
  | CI, the run from its start to its end | 141–235 s | 187 s |

  Without the caches of Go, the jobs of v0.1.30 are within the spread of those of v0.1.24: what the run gains is CodeQL's, whose Go now comes from the archive in 2 s in place of a build of the stage. The time of the builds and tests with the caches of Go is what the first pull request after the first push to `main` on v0.1.30 shows; locally a filled cache took `release-files` from 78 s to 3 s.

### 3.4. The stage `toolchain-core` and the shared checks v0.1.31, 2026-11-16

- **Locally.** In the devcontainer of `spec` v0.18.3, whose tools are those of the core of v0.18.4: `just ci-lint`, `ci-test`, `ci-vuln` and `ci-licenses` pass on `spec` v0.18.4, and `just release-files v0.0.0-ci` builds `dist/`. One run each.
