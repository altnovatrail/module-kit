# module-kit — sources

> The external facts that the documents of `docs/` of this repository rest on: a provider's terms, a price, a law, a standard, a library's documentation. Here are their records: the address, the version or the date of the page, the date read and the quote in the language of the source, — and [architecture.md](architecture.md) and the decision log hold the claim and a reference to the record by its number, as to a section: `sources 1.1` (RULES.md, The bar; R143 spec).
>
> A record is a subsection "N.M" of a topic: one source, or one reading of several sources for the sake of one decision. The number of a record does not change: a new record goes at the end of its topic, a new topic at the end of the file. A source read again is an item with its own date in the same record, and the earlier text is not edited. A new version of a source — a release, a tag, a commit — is a record of its own: so Go 1.27.3 and 1.27.4 stand side by side (sources 2.1 and 2.2), the shared release v0.1.4 and v0.1.10 (sources 3.7 and 3.8), go-yaml v4.0.0-rc.9 and v4.0.0 (sources 4.3 and 4.4). The model is sources spec. The results of measurements are not here but in [measurements.md](measurements.md).
>
> Until T43 the records stood as items in [architecture.md](architecture.md), section 4. A pointer remains there: the decision log and the archives of `docs/plan/` refer to that section, and they are not rewritten. A record begins with the decision or the section it was read for, so the sources of a decision are found by its number: `grep -n 'R21' docs/sources.md`.

## 1. MathTrail standalone

### 1.1. The solver sandbox of standalone

For architecture 3.4 — [mathtrail/mathtrail-standalone, `SPEC.md`](https://github.com/mathtrail/mathtrail-standalone/blob/c6cbeb2/SPEC.md), section 6 "The Starlark solver", branch `main`, commit c6cbeb2, read 2026-10-19:
- 6.1: "the program is written by the same model that wrote the task";
- 6.4, recursion is turned off: "A recursive Starlark function recurses on the Go stack, and a stack overflow in Go cannot be recovered: it would take the whole instance down, not the request";
- 6.5: "Every helper that builds a list is capped at 1,000,000 elements", "a step for every sixteen bytes"; "There is no random number generator, seeded or otherwise. A brute force that samples proves nothing about uniqueness, and a verdict that depends on a seed is a verdict nobody can reproduce from the task alone";
- 6.6, the table "The limits": Steps — 25,000,000; Wall clock — 2 seconds per run; Memory — "What the built-ins keep: 381 MiB at the ceiling above; nine tenths of the instance for the runtime"; Source size — 8 KB; Result — 5 letters; Concurrent runs — 1 per vCPU; Wait for a slot — "3 seconds per run", and a run that did not get a slot in time "does not start, and the task is told it was not checked and cost no attempt". On the literal: "`[{} for i in range(n)]` keeps 512 bytes for every nine steps — 2.6 million dictionaries at the ceiling, and an instance of 1 GiB was killed for its memory by them". On the calibration: "The bench measured the worst of the 450 at 2,081,362 steps, which is what moved the ceiling from 10,000,000 to 25,000,000" (R53 standalone) and "the worst of all 603 is 2,413,617 steps". On the memory gap: "a solver written that way takes the instance down rather than the request" and "The fix, if the gap ever bites, is a solver in a process of its own with a memory limit on it — a change of shape rather than a setting, and not one v1 pays for".

### 1.2. The memory gap is wider than in SPEC

For architecture 3.4 — [`docs/decisions.md`](https://github.com/mathtrail/mathtrail-standalone/blob/c6cbeb2/docs/decisions.md) of standalone, R53 standalone, commit c6cbeb2, read 2026-10-19: "the methods of the values themselves take a lazy range of any length in one step — `x.extend(range(3000000))` is nineteen steps", "so no cap on the helpers can close the gap, and only a limit on the memory of the process can"; on the time: "the clock is seen only between instructions" and "a process of its own would bound its time as well as its memory".

### 1.3. The price in bytes, the slots and the memory of an instance

For architecture 3.4 — R125 standalone, the same `docs/decisions.md`, commit c6cbeb2, read 2026-10-19: "an instance runs one solver for each vCPU, and it has 1 GiB, nine tenths of it told to the runtime"; rejected: "Four slots with the clock as insurance. Verdicts depend on the load" and "The process of its own now. A change of shape, and the author's decision"; open: "One account handing in solvers that spend their whole clock, for requests nobody opened, keeps an instance's single slot busy" and "Such a hand-in spends no attempt, so the account pays nothing for it. That stays open".

### 1.4. The Starlark interpreter of standalone

For architecture 3.4 — [`go.mod`](https://github.com/mathtrail/mathtrail-standalone/blob/c6cbeb2/go.mod), commit c6cbeb2, read 2026-10-19: `go.starlark.net v0.0.0-20260908191801-89a6a09411d5`.

### 1.5. Randomness in the code of standalone

For architecture 3.3 — a search of `internal/` and `cmd/` of the same commit, 2026-10-19:
- outside tests, only `internal/store/drive/calls.go` and `internal/transport/mcp/afresh.go` import `math/rand/v2`, and both calls give a random length of a pause (`time.Duration`); there is no generator of tasks from a seed there;
- in tests, PCG generators get their seed as numbers, without a hash: `rand.NewPCG(uint64(grade), uint64(child))` in `internal/domain/tutor/simulation_test.go`, `rand.NewPCG(7, 7)` in `pause_test.go`, `rand.NewPCG(d.Dice, 0)` in `property_test.go` and others.

### 1.6. The derivation of seeds in standalone's simulator

For architecture 3.3 — [`tools/learners/child.go`](https://github.com/mathtrail/mathtrail-standalone/blob/c6cbeb2/tools/learners/child.go), commit c6cbeb2, read 2026-10-19: the function `seeded(unit, purpose string) *rand.Rand` with the comment "seeded is the random stream of one unit of the experiment for one purpose"; it writes to `fnv.New64a()` the string `"%d|%s|%s|%s"` of `masterSeed`, `experiment`, `unit` and `purpose`, and returns `rand.New(rand.NewPCG(hash.Sum64(), masterSeed))` of `math/rand/v2`. The same function is in `research/experiments/` of the same commit.

## 2. Go and its toolchain

### 2.1. Go 1.27.3

For R04 — the list of go.dev's releases, `https://go.dev/dl/?mode=json`, read 2026-10-21: the newest stable release is `go1.27.3`; the sha256 of the archive `go1.27.3.linux-amd64.tar.gz` is `65efbc80fdc1fa2cd2631cad5b481777a68e8725f34e89f50fc6ac714e379b40`, of `go1.27.3.linux-arm64.tar.gz` — `a06e99309dbcbdca1a8ccd2282c3ee146e55930886e26b4f1193269587165f49`. Both archives were downloaded and checked with `sha256sum` on 2026-10-21.

### 2.2. Go 1.27.4

For R54 — the list of go.dev's releases, `https://go.dev/dl/?mode=json&include=all`, read 2026-11-12: the sha256 of the archive `go1.27.4.linux-amd64.tar.gz` is `7226736c2339719cda28fbf3a54d71aa63bcefe376e98cef00034268dab24840`, of `go1.27.4.linux-arm64.tar.gz` — `3392582988a273a87ab5043ee1b2c9c17a2c3c00580bf9326c47a2e0aaaeac92`, the same as in `.devcontainer/Dockerfile` of `spec` v0.16.3. The amd64 archive was downloaded and checked with `sha256sum` on 2026-11-12.

### 2.3. `GOTOOLCHAIN=local`

For R04 — [Go Toolchains](https://go.dev/doc/toolchain), read 2026-10-21: "When GOTOOLCHAIN is set to local, the go command always runs the bundled Go toolchain".

### 2.4. JSON in Go 1.27.3

For R09 — the sources in `$(go env GOROOT)` of the stage `toolchain`, read 2026-11-02:
- `src/internal/buildcfg/exp.go`, line 87: `JSONv2: true` in the set `baseline`, that is, the experiment `jsonv2` is on by default. The files of `encoding/json/jsontext` begin with `//go:build goexperiment.jsonv2`. `go list encoding/json/jsontext` without flags finds the package, and with `GOEXPERIMENT=nojsonv2` answers "build constraints exclude all Go files";
- `go doc encoding/json/jsontext.Value.Canonicalize`: "Canonicalize canonicalizes the raw JSON value according to the JSON Canonicalization Scheme (JCS) as defined by RFC 8785" and "JSON numbers are formatted as double precision numbers";
- `encoding/json` in go1.27 is built from the files `v2_*.go` with `//go:build goexperiment.jsonv2`, and `encode.go` with `//go:build !goexperiment.jsonv2`. In `internal/jsonflags/flags.go`, `DefaultV1Flags` includes `EscapeForHTML` and `EscapeForJS`, and in `internal/jsonwire/encode.go`, line 97, there is `case (r == '\u2028' || r == '\u2029') && flags.Get(jsonflags.EscapeForJS):`. `SetEscapeHTML(false)` removes only `EscapeForHTML`.

### 2.5. `math/rand/v2` and `hash/fnv` in Go 1.27.3

For R10 — the sources in `$(go env GOROOT)` of the stage `toolchain`, read 2026-11-02:
- `src/math/rand/v2/pcg.go`: `mulHi = 2549297995355413924`, `mulLo = 4865540595714422341`, `incHi = 6364136223846793005`, `incLo = 1442695040888963407` and the DXSM output with `cheapMul = 0xda942042e4dd58b5` — the same numbers as in the [format text](../spec/format/README.md), section 22;
- `src/math/rand/v2/rand.go`: `Uint64N` has "It panics if n == 0", `Int64N` — "It panics if n <= 0", `Shuffle` — "Shuffle panics if n < 0". `Perm` says nothing of a panic, but `Perm(-1)` fails before `Shuffle`, in `make([]int, n)`: "panic: runtime error: makeslice: len out of range", line 235 (a probe, 2026-11-02). On 32 bits `uint64n` goes through `uint32n`, and the comment on it says: "On 32-bit systems, here we need to implement that same logic in 32-bit math, both to preserve the exact output sequence observed on 64-bit machines and to preserve the optimization that the unbiasing loop almost never runs". The package's documentation says nothing of the streams staying unchanged: `go doc -all math/rand/v2` finds no such phrase;
- `src/math/rand/v2/regress_test.go`: "Test that random number sequences generated by a specific seed do not change from version to version" and "Do NOT make changes to the golden outputs. If bugs need to be fixed in the underlying code, find ways to fix them that do not affect the outputs";
- `src/hash/fnv/fnv.go`: "Package fnv implements FNV-1 and FNV-1a, non-cryptographic hash functions", `offset64 = 14695981039346656037` and `prime64 = 1099511628211`.

### 2.6. Go's promise about the streams

For R10 — the Go blog, ["Evolving the Go Standard Library with math/rand/v2"](https://go.dev/blog/randv2), 1 May 2024, read 2026-11-02. On `math/rand`: "At that point, we decided Go 1 compatibility included the specific random outputs for a given seed and added a test". On `math/rand/v2` — after `IntN` and the related functions got Lemire's method: "The concrete rand.Rand API is now locked in to that value stream, so we will not be able to take advantage of any optimizations yet to be discovered".

### 2.7. The toolchain of the main module

For R10 — [Go Toolchains](https://go.dev/doc/toolchain), read 2026-11-02: "The choice of Go toolchain being used depends on the GOTOOLCHAIN environment setting and the go and toolchain lines in the main module's go.mod file or the current workspace's go.work file" and "The go line for each module sets the language version the compiler enforces when compiling packages in that module".

### 2.8. The modules of `spec` at the Go proxy

For R20 — `https://proxy.golang.org/github.com/altnovatrail/spec/@v/v0.9.2.info` and `…/spec/contracts/@v/v0.9.2.info`, read 2026-11-04: both have `"Hash":"6f86ff50c76dacb79b7faf673713ddf02e42f531"`, and `Ref` is `refs/tags/v0.9.2` and `refs/tags/contracts/v0.9.2`. `go help mod download`, Go 1.27.3: `-json` has the field `Origin any // provenance of module`; `go mod download -json <module>@<version>` prints it, and without a version it does not (a probe, 2026-11-04).

### 2.9. Coverage from the tests of other packages

For R20 — `go help testflag`, Go 1.27.3, read 2026-11-04: "-coverpkg pattern1,pattern2,pattern3 Apply coverage analysis in each test to packages whose import paths match the patterns. The default is for each test to analyze only the package being tested".

## 3. Checks, CI and release

### 3.1. Tools

For R04 — the Go module proxy, `https://proxy.golang.org/<module>/@latest`, read 2026-10-21: `golang.org/x/vuln` — v1.9.0 (2026-10-11), `github.com/google/go-licenses/v2` — v2.0.1 (2025-09-08). The licenses — the GitHub API, `repos/<owner>/<repository>/license`, 2026-10-21: golang/vuln — BSD-3-Clause, google/go-licenses — Apache-2.0.

### 3.2. `go-licenses check`

For R04 — [README v2.0.1](https://github.com/google/go-licenses/blob/v2.0.1/README.md), section Usages, Check, read 2026-10-21: "Allow only specific license names" — `go-licenses check <package> [package...] --allowed_licenses=<comma separated license names>`. A probe on 2026-10-21: with `spec`'s list the recipe `ci-licenses` passes, and with a list of MIT alone it answers "Not allowed license 'Apache-2.0' found for library 'github.com/altnovatrail/module-kit/cmd/module-kit'".

### 3.3. A model for the recipes

For R04 — [the `justfile` of canvas-api](https://github.com/MathTrail/canvas-api/blob/0901af0439280334b3da33ffc4cc5c853e4403ff/justfile), commit 0901af0, read 2026-10-21: `ci-lint`, `ci-test`, `ci-vuln` and `ci-licenses`, and on the tools — "Each is a Go program run straight from its module, so none of them is installed anywhere".

### 3.4. SonarCloud, docker:S8545

For R04 — the rule "Go dependencies should be locked to verified versions" (the SonarCloud API, `api/rules/show?key=docker:S8545`, 2026-10-21). On pull request #8 it flagged the line `go install` of gopls in the Dockerfile: "Dependency versions are not predictable. Use a lock-file enforcing command instead". In canvas-api such findings are open on two lines of its Dockerfile, 198 and 199, although `-mod=readonly` stands there (the SonarCloud API, the project MathTrail_canvas-api, 2026-10-21).

### 3.5. Go autobuild in CodeQL

For R06 — [GitHub Docs, "CodeQL build options and steps for compiled languages"](https://docs.github.com/en/code-security/reference/code-scanning/codeql/build-options-for-compiled-languages), section "Autodetection for Go", the source in `github/docs` at commit c008b66, read 2026-10-21: "Invoke `make`, `ninja`, `./build` or `./build.sh` (in that order) until one of these commands succeeds and a subsequent `go list ./...` also succeeds"; "By default, test code (code in files ending in `_test.go`) is not analyzed".

### 3.6. Licenses in the release

For R07:
- Apache-2.0 — [LICENSE](../LICENSE), section 4 (a): "You must give any other recipients of the Work or Derivative Works a copy of this License";
- Go's license — `LICENSE` at the root of Go 1.27.3 of the stage `toolchain`, `$(go env GOROOT)/LICENSE`, read 2026-10-21: "Redistributions in binary form must reproduce the above copyright notice, this list of conditions and the following disclaimer in the documentation and/or other materials provided with the distribution". A Go binary carries the standard library, and `module-kit` has no third-party modules (`go.mod`).

### 3.7. The shared release

For R07 — [`release.yml`](https://github.com/altnovatrail/.github/blob/11435dbe61d3d4885c9a8dd30686e8111e26cabb/.github/workflows/release.yml) and R18 .github, the release of `.github` v0.1.4, commit 11435db, read 2026-10-21: the conditions for the recipe `release-files` — the stage `toolchain` of the commit being released, its tree from `git archive`, a run under the runner's uid with `HOME=/tmp` and without tokens, in `dist/` only regular files with names by `^[A-Za-z0-9_+][A-Za-z0-9._+-]*$` and without `.sigstore.json`. The release is immutable: `gh release verify v0.1.4 -R altnovatrail/.github` — "Release v0.1.4 verified!".

### 3.8. The shared release v0.1.10

For R07 and T08 — a comparison of `release.yml` of `.github` v0.1.4 (11435db) and v0.1.10 (f3addb7), `gh api repos/altnovatrail/.github/compare/…`, read 2026-11-02. The release runs the recipes of `spec` with the script `spec/tools/toolchain/toolchain.sh` from the caller's submodule, and before the build it checks: "that spec/ be a submodule of altnovatrail/spec, checked out at the commit the caller pins and at a tag of spec" and "that the tag hold the script, which spec ships from v0.2.17; anything else stops it before anything is built, with the reason". It checks the script's sum after the build too: "spec/tools/toolchain/toolchain.sh changed in the copy of the checkout while the release ran". The conditions for the recipe `release-files` above were read at v0.1.4: at v0.1.10 the first release will check them.

### 3.9. Coverage and the list `paths`

For R18 — [docs/architecture.md of `.github`](https://github.com/altnovatrail/.github/blob/89d3ceee0b9977b3334488a4133d084d76474e0d/docs/architecture.md), section 3.3 — «Покрытие», the job "Go" and the list of the kind `go`, — and [`go.yml`](https://github.com/altnovatrail/.github/blob/89d3ceee0b9977b3334488a4133d084d76474e0d/.github/workflows/go.yml), the release of `.github` v0.1.15, commit 89d3cee, read 2026-11-04. Of the shared workflows, between v0.1.10 and v0.1.15 only `go.yml`, `release.yml` and `typescript.yml` changed (`gh api repos/altnovatrail/.github/compare/v0.1.10...v0.1.15`, 2026-11-04).

### 3.10. gopls from a lock file

For R19 — R12 tutor and its [`.devcontainer/Dockerfile`](https://github.com/altnovatrail/tutor/blob/be607e56f87bf94a93b1d318f9cd578a3e3c221e/.devcontainer/Dockerfile) and `.devcontainer/gopls/`, `main` be607e5, read 2026-11-04. `tutor` has no findings of docker:S8545 (the SonarCloud API, `api/issues/search?componentKeys=altnovatrail_tutor&rules=docker:S8545`, 2026-11-04). Its Scorecard Pinned-Dependencies flags in the Dockerfile only `npmCommand not pinned by hash` (the GitHub API, `code-scanning/alerts`, 2026-11-04). gopls v0.24.0 of 2026-10-14 is the latest release (`https://proxy.golang.org/golang.org/x/tools/gopls/@latest`, read 2026-11-04); the license of golang/tools is BSD-3-Clause (the GitHub API, `repos/golang/tools/license`, 2026-11-04).

### 3.11. When Dependabot proposes a version

For the 2026-11-04 amendment to R18 — GitHub Docs, read 2026-11-04: ["Configuring Dependabot version updates"](https://docs.github.com/en/code-security/dependabot/dependabot-version-updates/configuring-dependabot-version-updates) — "When this file is checked in, Dependabot checks the manifest files on the default branch for outdated dependencies"; ["Dependabot options reference"](https://docs.github.com/en/code-security/dependabot/working-with-dependabot/dependabot-options-reference), `cooldown` — "A new version is not considered for a version update until 3 days after its release", and from what instant and how the days are counted it does not say. The release of `.github` v0.1.15 was published 2026-11-02 at 03:40:33 UTC (`gh release view v0.1.15 -R altnovatrail/.github`, 2026-11-04).

## 4. Go libraries

### 4.1. JCS libraries for Go

For R09 — the Go module proxy and the GitHub API, read 2026-11-02:
- `github.com/gowebpki/jcs` — `@latest` v1.0.2 of 2026-09-21, the license Apache-2.0. In `go.mod` — `require github.com/stretchr/testify v1.12.1`. Issue #5 "RFC 8785: reversed surrogate pair becomes U+FFFD and invalid UTF-8 passes through" was closed 2026-09-20. It reads a number as a double: `jcs.go` at the tag v1.0.2, lines 530 and 539, — `strconv.ParseFloat(value, 64)`, then `NumberToJSON(ieeeF64)`;
- `github.com/cyberphone/json-canonicalization` — there are no version tags, `@latest` is the pseudo-version `v0.0.0-20261011152517-32a48195b9b2`. The license by the GitHub API is `NOASSERTION`, and the repository has no `go.mod`. Issue #33 "Go import (as documented) broken" has been open since 2025-03-11.

### 4.2. go.starlark.net

For R11 — [google/starlark-go](https://github.com/google/starlark-go), the license BSD-3-Clause (the GitHub API), read 2026-11-02:
- there are no version tags. `https://proxy.golang.org/go.starlark.net/@latest` is `v0.0.0-20261021141710-4f4267fb140d`, commit 4f4267fb140d of 2026-10-21; standalone pinned `v0.0.0-20260908191801-89a6a09411d5` (`go.mod` of standalone, commit 89986bc);
- pull request #663 "starlark: check steps/cancellation only on back-edges", merged 2026-10-05 by commit de228787c8fb: "We only need to increment the step counter and check for cancellation in loops (back edges) and calls. We increment the counter by a constant scaled by the number of instructions between such checks" and "Clients may need to adjust their SetMaxSteps parameter to taste". In `starlark/interp.go` at 4f4267fb140d there is `const stepCost = 10 // empirical average number of instructions per back-edge or call`, and at 89a6a09411d5, line 109, `thread.Steps++` on every instruction;
- `starlark/eval.go` at 4f4267fb140d. On `Cancel`: "There may be a delay before the interpreter observes the cancellation if the thread is currently in a call to a built-in function". On the repetition of a list and of a string: "It's always possible to overeat in small bites but we'll try to stop someone swallowing the world in one gulp" and `const maxAlloc = 1 << 30`;
- `starlark/interp.go` at 4f4267fb140d on recursion. When it is off: "We look for the same function code, not function value, otherwise the user could defeat the check by writing the Y combinator". When it is on — `if len(thread.stack) > 100_000`, since pull request #619 of 2025-10-27: "Exceeding this limit should not crash the Go program";
- pull request #672, merged 2026-10-21: "calling one of them without an argument dereferenced a nil `Iterable` and panicked, e.g. `set([1]).difference()`".

### 4.3. go-yaml v4

For R21 — the Go module proxy, `https://proxy.golang.org/go.yaml.in/yaml/v4/@latest`, read 2026-11-04: `v4.0.0-rc.9` of 2026-11-02, commit 2141fcc, the tag `refs/tags/v4.0.0-rc.9`. The license by the GitHub API (`repos/yaml/go-yaml/license`) is Apache-2.0, and the module's `NOTICE`: "The following files were ported to Go from C files of libyaml, and thus are still covered by their original MIT license". The module's `go.mod` is `go 1.18`, without `require`. The README at the same tag: "v4.0.0 *is not released yet.* It is intended to be released in October 2026", "v2 and v3 remain stable" and "The library supports most of YAML 1.2 and retains selected YAML 1.1 behavior".

### 4.4. go-yaml v4.0.0

For R45 — `https://proxy.golang.org/go.yaml.in/yaml/v4/@latest`, read 2026-11-06: `v4.0.0` of 2026-11-04T18:36:35Z, commit f2f9563, the tag `refs/tags/v4.0.0`. Five commits since rc.9, `gh api repos/yaml/go-yaml/compare/v4.0.0-rc.9...v4.0.0`; `yts/known-failing-tests` of v4.0.0 still names 2SXE and W5VH.

### 4.5. The JSON Schema validator

For R23 — read 2026-11-04:
- `github.com/santhosh-tekuri/jsonschema/v6` — `https://proxy.golang.org/github.com/santhosh-tekuri/jsonschema/v6/@latest`: `v6.0.3` of 2026-06-28, commit b0fc661, `refs/tags/v6.0.3`. The license by the GitHub API (`repos/santhosh-tekuri/jsonschema/license`) is Apache-2.0. Its `go.mod` is `go 1.21`, `require github.com/dlclark/regexp2 v1.11.0 // used for testing` and `golang.org/x/text v0.14.0`. Bowtie, [bowtie.report](https://bowtie.report), the run of 2026-11-04 06:21 UTC: on draft 2020-12, go-jsonschema (v6.0.2) passes 1300 tests of 1300;
- its source at v6.0.3, `validator.go`: for `propertyNames` — `verr.InstanceLocation = vd.vloc`, without a copy, while `instanceLocation()` of the other errors is `slices.Clone(vd.vloc)`. Issue [#276](https://github.com/santhosh-tekuri/jsonschema/issues/276) "propertyNames and contentSchema errors: InstanceLocation is overwritten by later siblings" has been open since 2026-09-24; pull request #278 was closed without a merge on 2026-10-02, "Closing in favor of earlier community PR #269", and #269 is open. A probe on 2026-11-04 on `spec.Format` v0.9.2: `examples/task/invalid/parameter-named-if.yaml` got `propertyNames at #/items/0` instead of `#/template/parameters`. Commit 9ae86900d of 2026-10-31, after v6.0.3, — "validator/fix: big.Rat fails to parse large exponents causing panic": before it, `numValidate` reads a number with `new(big.Rat).SetString(fmt.Sprintf("%v", v))` and does not check the result. How a vocabulary is turned on — `Compiler.RegisterVocabulary` and `AssertVocabs`, and an extension's error is given by `ValidatorContext.AddError` (`vocab.go`);
- `golang.org/x/text` — `@latest` `v0.43.0` of 2026-10-09; govulncheck v1.9.0 on v0.14.0: GO-2026-6629 "Panic parsing crafted input in x/text/secure/precis", fixed in v0.41.0, and GO-2026-5970 "Infinite loop on invalid input", fixed in v0.39.0, — "your code doesn't appear to call these vulnerabilities"; on v0.43.0 — "No vulnerabilities found". go-licenses v2.0.1: santhosh — Apache-2.0, x/text — BSD-3-Clause;
- other libraries: `github.com/kaptinlin/jsonschema` — `v0.9.10` of 2026-09-13, MIT, Bowtie does not check it; `github.com/google/jsonschema-go` — `v0.4.3` of 2026-04-17, MIT, validation stops at the first error; Corvus Go (`corvus-dotnet/Corvus.JsonSchema`, `src-go/`) — v0.1.4 of 2026-11-03, the first release v0.1.0 of 2026-10-08, Apache-2.0, patterns by ECMA-262; `xeipuuv/gojsonschema` v1.2.0 — draft-04, -06 and -07; `qri-io/jsonschema` v0.2.1 — draft 2019-09.

### 4.6. go.starlark.net in the sandbox

For R28 — the sources of `go.starlark.net` `v0.0.0-20261021141710-4f4267fb140d` from the module cache, read 2026-11-04:
- `starlark/library.go`, on `Universe`: "The Go application may add or remove items from the universe dictionary before Starlark evaluation begins"; it holds 33 names — those that the [format text](../spec/format/README.md), section 24, lists;
- `starlark/eval.go`: on `Thread.Print` — "If nil, fmt.Fprintln(os.Stderr, msg) is used instead"; on `OnMaxSteps` — "The default behavior is to call thread.Cancel("too many steps")"; `Cancel` — "Atomically set cancelReason, preserving earlier reason if any"; on `EvalExprOptions` — "Resolving an expression mutates it. Do not call EvalExprOptions more than once for the same expression", and `makeExprFunc` resolves the names with `resolve.ExprOptions(opts, expr, env.Has, Universe.Has)`;
- `starlark/interp.go`: `checkSteps` — `thread.Steps += stepCost` and `thread.Steps >= thread.maxSteps`, and `stepCost = 10`;
- `starlark/int_posix64.go`, the build for linux and darwin on 64 bits: "It reserves a 4GB portion of the address space using mmap", `unix.Mmap(-1, 0, len, unix.PROT_READ, unix.MAP_PRIVATE|unix.MAP_ANON)`, and on an error — `log.Printf("Starlark failed to allocate 4GB address space: %v. Integer performance may suffer.", err)`;
- the module's `go.mod` — `go 1.25.0`, and `golang.org/x/sys v0.42.0` among `require`; the license of google/starlark-go by the GitHub API is BSD-3-Clause.

### 4.7. `golang.org/x/sys`

For R28 — read 2026-11-04: `https://proxy.golang.org/golang.org/x/sys/@latest` is `v0.49.0` of 2026-10-06; the license of golang/sys by the GitHub API is BSD-3-Clause. govulncheck v1.9.0 on v0.42.0: GO-2026-5024 "Invoking integer overflow in NewNTUnicodeString in golang.org/x/sys/windows", fixed in v0.44.0, the platform windows, — "your code doesn't appear to call these vulnerabilities"; on v0.49.0 — "No vulnerabilities found". go-licenses v2.0.1: go.starlark.net and `golang.org/x/sys/unix` — BSD-3-Clause.

### 4.8. protobuf-go

For R29 — the sources of `google.golang.org/protobuf` v1.36.12 from the module cache, read 2026-11-05:
- `internal/errors/errors.go`, `InvalidUTF8`: "field %v contains invalid UTF-8" — this is how `proto.Marshal` refuses a string not in UTF-8, and `TestNotUTF8IsNoMessage` holds it on an observation;
- `internal/filedesc/desc.go`, `EnforceUTF8`: "If this method does not exist, the default is to enforce valid UTF-8";
- `encoding/protojson/decode.go`, `unmarshalEnum`: a JSON number becomes `protoreflect.EnumNumber(n)` if it fits in 32 bits, without a lookup in the enum. So `protojson.Unmarshal` reads a vector's cause `1000`;
- the license — the module's `LICENSE`, BSD-3-Clause, as go-licenses v2.0.1 names it (`ci-licenses`).

## 5. Standards and specifications

### 5.1. Semantic Versioning 2.0.0

For R13 and architecture 3.5 — [semver.org/spec/v2.0.0.html](https://semver.org/spec/v2.0.0.html), read 2026-11-02:
- item 3: "Once a versioned package has been released, the contents of that version MUST NOT be modified";
- item 4: "Major version zero (0.y.z) is for initial development. Anything MAY change at any time";
- item 8: "Major version X (X.y.z | X > 0) MUST be incremented if any backward incompatible changes are introduced to the public API. It MAY also include minor and patch level changes" — the author may raise the version part that the comparison found to a higher one.

### 5.2. Cargo on 0.y.z

For R13 and architecture 3.5 — [The Cargo Book, "SemVer Compatibility"](https://doc.rust-lang.org/cargo/reference/semver.html), read 2026-11-02: "Initial development releases starting with "0.y.z" can treat changes in "y" as a major release, and "z" as a minor release. "0.0.z" releases are always major changes. This is because Cargo uses the convention that only changes in the left-most non-zero component are considered incompatible". The same in oasdiff: "Below `1.0.0`, semver gives the minor version the major's role" ([`docs/VERSIONING.md`](https://github.com/oasdiff/oasdiff/blob/2daca9b/docs/VERSIONING.md), commit 2daca9b, read 2026-11-02).

### 5.3. Tools that check semver by the content

For R13, read 2026-11-02:
- gorelease — [pkg.go.dev/golang.org/x/exp/cmd/gorelease](https://pkg.go.dev/golang.org/x/exp/cmd/gorelease): "If no version is proposed with -version, gorelease suggests the lowest version consistent with semantic versioning". It finds the lowest part and proposes a version;
- apidiff — [pkg.go.dev/golang.org/x/exp/apidiff](https://pkg.go.dev/golang.org/x/exp/apidiff): "No tool can detect behavioral changes; and even if it could, whether a behavioral change is a breaking change or not depends on many factors, such as whether it closes a security hole or fixes a bug". The behavior of a module is the check of a task, and the comparison sees it;
- cargo-semver-checks, the license Apache-2.0 — [README](https://github.com/obi1kenobi/cargo-semver-checks/blob/161450e/README.md), commit 161450e: "By default, `cargo-semver-checks` uses crates.io to look up the previous version of the crate, which is used as the baseline for semver-checking the current version of the crate", and the previous version can also be given as a folder, `--baseline-root`. And: "A design goal of `cargo-semver-checks` is to not have false positives". It requires no less than what it found, and the author sets the version themselves;
- oasdiff — the same `docs/VERSIONING.md`: "a breaking change requires a major version increase", and under "What is not checked" — "Which version bump a non-breaking change deserves. Only breaking changes are checked";
- `elm bump` — [`terminal/src/Main.hs`](https://github.com/elm/compiler/blob/1c278bf/terminal/src/Main.hs) of elm/compiler, commit 1c278bf: "so there cannot be MAJOR changes hiding in PATCH releases in Elm!"; the rule — [`builder/src/Deps/Diff.hs`](https://github.com/elm/compiler/blob/1c278bf/builder/src/Deps/Diff.hs), lines 308–312: `addMag = if null added then M.PATCH else M.MINOR`, `removeMag = if null removed then M.PATCH else M.MAJOR` and the largest of the parts. A removal is major, an addition is minor, and Elm requires exactly the version it found;
- a version of data — the recipe Frictionless Data Package Version, [specs.frictionlessdata.io/patterns](https://specs.frictionlessdata.io/patterns/): major — "Combine, split, delete or change the meaning of data that is referenced by another data resource", patch — "Correct errors in existing data". Here the author declares the part, and there is no tool;
- [Conventional Commits 1.0.0](https://www.conventionalcommits.org/en/v1.0.0/): "fix: a commit of the type fix patches a bug in your codebase (this correlates with PATCH in Semantic Versioning)". The author declares the part — by the type of the commit, — and the tool does not look at the content.

### 5.4. YAML 1.2.2

For R21 — [yaml.org/spec/1.2.2](https://yaml.org/spec/1.2.2/), read 2026-11-04:
- 5.1: "To ensure JSON compatibility, YAML processors must allow all non-C0 characters inside quoted scalars";
- 5.2: "Byte order marks may appear at the start of any document" and "To allow for JSON compatibility, byte order marks are also allowed inside quoted scalars"; `nb-char ::= c-printable - b-char - c-byte-order-mark`;
- 6.1: "To maintain portability, tab characters must not be used in indentation";
- 6.3: `s-flow-line-prefix(n) ::= s-indent(n) s-separate-in-line?`;
- 7.4.2: "the ":" indicator must appear at most 1024 Unicode characters beyond the start of the key";
- 8.1.1.1: "It is an error for any of the leading empty lines to contain more spaces than the first non-empty line";
- 8.2.2: implicit keys "are limited to a single line and must not span more than 1024 Unicode characters".

### 5.5. The YAML test suite

For R21 — [yaml/yaml-test-suite](https://github.com/yaml/yaml-test-suite), the data release `data-2022-01-17`, commit 5f49729, the one v4 takes too (`YTS-TAG ?= data-2022-01-17` in `GNUmakefile` of v4.0.0-rc.9); the license is MIT, the file `License` at the tag `v2022-01-17`, commit 6d48918 (the GitHub API, read 2026-11-04). The release has 402 cases, 94 of them with a file `error`. v4 rc.9 on them, without the passes of `internal/yaml`: it lets through five invalid ones — 9C9N, QB6E, DK95/01, Y79Y/003 and S98Z, — and rejects 14 valid ones, 12 of them directives and anchors.

### 5.6. The signatures of media files

For R22 — [WHATWG MIME Sniffing](https://mimesniff.spec.whatwg.org/), read 2026-11-04: 6.2.3 "Signature for MP3 without ID3", "To match an mp3 header", step 1 — "If length is less than 4, return false". The same — `isMp3Header` in `spec/tools/format-check/module.mjs` at v0.9.2: `bytes.length >= 4`.

### 5.7. Unicode for normalization

For R27 — read 2026-11-04:
- Go 1.27.3, `src/unicode/tables.go`: `const Version = "17.0.0"`; `CaseRanges` — "the table describing case mappings for all letters with non-self mappings", and `unicode.ToLower` by it is a simple mapping: a probe gives for U+0130 the character U+0069, a single one, and U+0085 is White_Space, U+FEFF is not;
- `golang.org/x/text` v0.43.0, `unicode/norm`: `tables17.0.0.go` — `//go:build go1.27`, `Version = "17.0.0"`, and `tables15.0.0.go` — `//go:build !go1.27`, `Version = "15.0.0"`; the same for `cases` and `rangetable`. `normalize.go`, `decomposeSegment`: "It inserts 0x034f (Grapheme Joiner) when it encounters a sequence of more than 30 non-starters", and `composition.go` — `maxNonStarters = 30`. A probe: `"a"` and 31 characters U+0301 — the output of `norm.NFKC` has U+034F;
- [UAX #15](https://www.unicode.org/reports/tr15/), section 13 "Stream-Safe Text Format", version 18.0.0, revision 58, of 2026-08-12: "The Stream-Safe Text Process is the process of producing a Unicode string in Stream-Safe Text Format by processing that string from start to finish, inserting U+034F COMBINING GRAPHEME JOINER (CGJ) within long sequences of non-starters" and "The Stream-Safe Text Process preserves all of the four normalization forms defined in this annex (NFC, NFD, NFKC, NFKD). However, normalization and the Stream-Safe Text Process do not commute";
- the second implementation of `spec` — `spec/tools/format-check/verdicts.mjs` at v0.9.2: NFKC — `text.normalize("NFKC")` of Node 24.21, ICU 78.3, Unicode 17.0 (R56 spec), without the Stream-Safe Text Process; a lowercase letter — `toLowerCase` with a special case of U+0130, "as Go's unicode.ToLower".

### 5.8. Strings in proto3 and ProtoJSON

For R29 — read 2026-11-05:
- [Language Guide (proto 3)](https://protobuf.dev/programming-guides/proto3/), the table "Scalar Value Types", `string`: "A string must always contain UTF-8 encoded or 7-bit ASCII text, and cannot be longer than 2³²";
- [ProtoJSON Format](https://protobuf.dev/programming-guides/json/), the row `enum`: "The name of the enum value as specified in proto is used. Parsers accept both enum names and integer values".

## 6. The second implementation of `spec`

### 6.1. The second implementation of `spec`

For R21 — `spec/tools/format-check/yaml.mjs` at v0.9.2, npm `yaml` 2.9.1, a probe on 2026-11-04 through its `readModuleFile`: it reads `a: x<U+2028>  y` as `"x\u2028  y"`; `%YAML 1.2` — `yaml.no-directives`; a BOM after `...` — `yaml.encoding`; `&x.y` — `yaml.no-anchors`; the empty key `? ` twice — `yaml.plain-strings` twice and `yaml.unique-keys`; "not YAML" — `\U00110000`, a control character in a comment, an implicit key longer than 1024 characters and under-indented continuation lines. It rejects a file with CR line breaks, counts `...` at the start of a stream as a document, and accepts U+FEFF in a plain and in a block scalar and in a comment.

### 6.2. The second implementation of `spec` and the schemas

For R23 — `spec/tools/format-check/` at v0.9.2, read 2026-11-04: `schemas.mjs`, lines 368–375, — `new Ajv2020({allErrors: true, validateFormats: false, …})` and `ajv.addKeyword("x-open")`; `cases.mjs`, lines 144–156, — an example passes if `errors.some((error) => error.keyword === header.keyword && error.instancePath === header.pointer)`, and `expectedError`, line 828, decodes the place with `decodeURIComponent`; `cases.mjs`, lines 219–239, — `module.file-ids` "read only from a file that passed its schema, so that a file the schema refuses breaks that and nothing more", and the rules of the whole module — `if (values.partsOk)`. The open view is built by `openview.mjs` with a walk of its own, and ajv answers only whether a subschema passes (`passes`, lines 116–121).
