# module-kit's decision log

> Why module-kit is built the way it is. Each entry: what was decided, why, what was rejected, and in its tail "— *Consequence:*" the files, by their paths, that a change of it edits.
>
> **Rule:** do not propose again what was rejected without new facts. If a decision has to change, ask the author and rewrite its entry in its place, and every entry the change touches: an entry is its text in force — what was decided, why, what was rejected and what a change of it edits —, git keeps the old text, and the commit says why. An entry's number stands in its heading alone: an entry cites no other entry and no section, says in words what it rests on, and carries no date of who decided it and no amendment; a claim about the outside world cites its one record of [sources.md](sources.md) or [measurements.md](measurements.md).
>
> The repository was set up 2026-10-19 from the template of `spec` v0.1.0. How it comes to differ from the template — tools in the devcontainer, rules in RUN.md — is recorded here as a decision; a move to a tag of `spec` is its commit, not an entry. Next number — R70.
>
> Open questions that concern several repositories — О-, П-, Т- — are in the registers of `spec`, in `spec/docs/product.md`, `spec/docs/privacy.md` and `spec/docs/repositories.md`. What is still to decide in the component itself is in [architecture.md](architecture.md).

## Documents, findings and the devcontainer

**R01. The documents of `module-kit` are checked by `just -f spec/justfile docs-check-repo`, the check of the `spec` its submodule is on, which CI runs by the shared `documents.yml` of `.github` as a required check. An example of a citation is written in inline code, so that the check does not take it for one.**

— *Why:* the shared `documents.yml` runs this recipe from the caller's submodule, in the toolchain stage of the caller's Dockerfile, and its check holds the merge. The check of the tag the submodule is on is the one whose rules the repository follows: a record of `spec` cited that does not exist at the tag, a gap in the log or a breach of the rule of citations does not pass unnoticed.

— *Rejected:* examples of citations in guillemets: the check reads a number in guillemets as a citation and checks it against the log, and a check that skipped guillemets would let real citations through unchecked.

— *Consequence:* `CLAUDE.md`; `RUN.md`; `docs-budgets.json`; `.github/workflows/ci.yml`.

**R02. RUN.md has the template's rule on the analysers' findings, in its words: before a new task, close the findings of SonarCloud, CodeQL and Scorecard on the previous one, by a change first, and as a false positive only as a last resort, with its reason in words. The merge waits for the checks that the ruleset `main` requires, and is made by a merge commit alone.**
- **The reason of a finding left open** — one shared by every repository with the shared checks is in `spec`'s decision log, `spec/docs/decisions.md`; one of this repository's own, in this log. The author commits the fixes of findings apart from the task.
- **The merge.** The author sets the required checks in the ruleset `main`; the findings of CodeQL hold the merge by a rule of the ruleset — errors and vulnerabilities of high and above. A merge is made by a merge commit: DCO checks the commits of a pull request, and a squash would put on `main` a commit whose sign-off nobody checked.
- **The CI files** for the repository's own code are written by its own sessions, by the calling files that `.github` gives for its shared workflows. Scorecard is the shared `scorecard.yml` of `.github`, whose result its server accepts from a shared workflow called from another repository, and the repository keeps no copy of it.

— *Why:* without the rule a next task starts without sorting out the findings of the previous one, and a finding left open without its reason cannot be told from one that was answered.

— *Rejected:* the rule in words of this repository's own: it is the template's text, and a difference from it would have to be recorded as a decision, with no reason for one.

— *Consequence:* `RUN.md`; `.github/workflows/`; [architecture.md](architecture.md).

**R03. `.devcontainer/` is what the template of `spec` builds for `module-kit` of the kind `go`, and a move to a tag of `spec` takes it from the template: in its Dockerfile each archive is checked against a sha256 checksum for each architecture, `amd64` and `arm64`, before it is unpacked, and Go is the Go of `spec`'s own stage `toolchain-core`. A difference of this container from `spec`'s is an entry of this log, and a fix of the template's files goes to `spec`, by a row «В реестр spec» of `run/QUESTIONS.md`.**
- A checksum comes from the publisher's own list, an `ARG` beside the version; a wrong one stops the build, and it changes with the version, by hand, in the same change: Dependabot does not see these versions.
- Go is raised by a move to a tag of `spec` that raises it in the template.

— *Why:* the container of `module-kit` is the container of `spec`, so every repository works in one environment and a fix of it is made once, in `spec`; an archive that is not checked could change under the same version.

— *Rejected:*
- to take from the template's Dockerfile only a part — the checksums, or a comment —, or to keep lines of this repository's own over the template's where Go is the same: the difference would have to be recorded as a decision, with no reason for one;
- to keep the Dockerfile as it was, so that a move to a tag gives no release: the container would diverge from `spec`'s, and the next edit of the Dockerfile, which is in the list `paths` of `release.yml`, would give a release anyway.

— *Consequence:* `.devcontainer/`; `CLAUDE.md`; `.claude/rules/ci.md`.

## Code and checks

**R04. The code of `module-kit` is Go: the module `github.com/altnovatrail/module-kit` and the CLI `cmd/module-kit`. `go.mod` names the first patch of the minor as its lower bound, and Go itself is the one of the stage `toolchain-core` of `.devcontainer/Dockerfile`, with `GOTOOLCHAIN=local`. The checks are recipes of its own, `ci-lint`, `ci-test`, `ci-vuln` and `ci-licenses`, which the shared `go.yml` of `.github` runs, and their tools run through `go run` at an exact version.**
- **The `justfile`:**
  - `build` builds the CLI into `bin/`, which `.gitignore` keeps out of git;
  - `ci-lint` — gofmt over the `.go` files that git tracks or would track, without `testdata/`: git names no file of the submodule `spec/` in this list; and `go vet`;
  - `ci-test` — the module's tests, with `-race`;
  - `ci-vuln` — govulncheck: known vulnerabilities in the code that the module reaches;
  - `ci-licenses` — go-licenses: the license of each package the module is built from must be in the list `ALLOWED_LICENSES`, which the shared `go.yml` gives (sources 3.2); without the list the recipe stops and says where to get it.

  A tool runs through `go run` at the module's exact version, which the `justfile` pins: nothing is installed into the image, the Go module proxy serves exactly this version, and Go's checksum database vouches for it, as the `justfile` of canvas-api does (sources 3.3).
- **`GOTOOLCHAIN=local`** keeps the command `go` from downloading another toolchain (sources 2.3).
- **`.github/dependabot.yml`** — the ecosystem `gomod` with a three-day pause. Dependabot sees neither Go in the Dockerfile nor the tools in the `justfile`: they are raised by hand.
- **`.github/workflows/ci.yml`** — the job "Go", which calls the shared `go.yml` of `.github`; the author puts its checks "Go / lint", "Go / test", "Go / vuln" and "Go / licenses" in the ruleset `main`.

— *Why:* the card of `module-kit` names a CLI for module authors, and the shared `go.yml` was tried on its skeleton ahead of the format's code; the skeleton stayed and grew into the CLI. The lower bound of `go.mod` is the one that `api` and tutor's eval runner, which import the library, must meet.

— *Rejected:*
- to wait for the format's code before the first Go code: the shared `go.yml` would have waited untried;
- a throwaway hello world to try it on: it would be removed later, while the CLI stays;
- gopls in the image through `go install …@v<version>`: SonarCloud flags it by the rule docker:S8545 as a dependency without a lock file, with `-mod=readonly` too, and its required check holds the merge (sources 3.4); gopls built inside a module of its own does not know its version and prints `(devel)`;
- the patch of Go, as `go 1.27.3`, in `go.mod`: each patch here would become a lower bound for `api` and the eval runner of `tutor`;
- golangci-lint in `ci-lint`: gofmt and go vet were enough for the skeleton, and golangci-lint is under GPL-3.0 (sources 3.1), so it could come only as a named exception;
- to install the checks' tools into the image: `go run` at an exact version does without it;
- a list of licenses of this repository's own: it is one for all repositories and comes from `go.yml`.

— *Consequence:* `go.mod`; `cmd/module-kit/`; `justfile`; `.gitignore`; `.github/dependabot.yml`; `.github/workflows/ci.yml`; `.devcontainer/Dockerfile`; [architecture.md](architecture.md).

## The review and CodeQL for Go

**R05. A task's review is one round of `/code-review high`, and the executor may hand in a small change without one; the rules of RUN.md and AGENTS.md are the template's, but for the values put into them.**

— *Why:* the rule of the review is the same in every repository and comes with the template, and so do RUN.md's other rules and AGENTS.md: rules of this repository's own would have to be recorded as decisions.

— *Rejected:* rules of RUN.md or an AGENTS.md that diverge from the template's: a difference would have to be recorded as a decision, with no reason for one.

— *Consequence:* `RUN.md`; `AGENTS.md`.

**R06. CodeQL reads Go: `.github/workflows/codeql.yml` calls the shared `codeql.yml` of `.github` with the input `languages: go`, and the author puts the check "CodeQL / analyze (go)" in the ruleset `main`.**
- **How it reads Go:**
  - by a build: the analysis runs with `build-mode: autobuild`, and the step analyze builds the database itself;
  - the Go for the build is that of `.devcontainer/Dockerfile`: the shared workflow takes the archive of go.dev that the stage `toolchain-core` installs, by the `ARG`s that pin it there, checked against the same sum, and sets `GOTOOLCHAIN=local` (sources 3.14). So the analysis runs with the Go of the checks "Go", and a raise of Go in the Dockerfile raises it too;
  - the queries are `security-extended`, the results are under the category `/language:go`, and the findings hold the merge by the ruleset's rule for CodeQL — errors and vulnerabilities of high and above. `actions` is always analysed too;
  - the autobuild of Go first runs `make`, `ninja`, `./build` and `./build.sh` in turn, until one of them and the `go list ./...` that follows it pass, and only then installs the dependencies by `go.mod` and extracts the code, as `go build ./...`; the tests, the files `_test.go`, are not analysed by default (sources 3.5). `module-kit` has neither a Makefile, nor `build.ninja`, nor `build` and `build.sh` in its root: the build is run by the `justfile`. If such a file appears, the analysis of Go needs the mode `manual`.

— *Why:* without `go` the Go code comes to `main` without CodeQL's analysis, and the ruleset's rule for CodeQL's findings has nothing to hold.

— *Rejected:*
- to leave the input to Dependabot: it raises the reference and adds no input, and without `go` Go goes without analysis;
- `go` with a release of the shared `codeql.yml` that does not accept it: the check "CodeQL / languages" would stop with its reason, and the required "CodeQL / analyze (actions)" would not come and would hold up every merge;
- CodeQL's default setup in the repository's settings: with it GitHub turns off the advanced setup — the shared workflow —, and Go would not be analysed with the Go of the Dockerfile.

— *Consequence:* `.github/workflows/codeql.yml`; `.claude/rules/ci.md`; [architecture.md](architecture.md).

## Release

**R07. The release of `module-kit` is put out by the shared `release.yml` of `.github`, which `.github/workflows/release.yml` calls, and the files for it are built by a recipe of its own, `release-files`: the CLI as one binary per platform — linux, darwin and windows, each on amd64 and arm64 —, the licenses the binaries must come with and `SHA256SUMS`. The number is the next patch of the line that `RELEASE_LINE` of `release.yml` names, and a release comes out only when something in its list `paths` has changed.**
- **The calling `.github/workflows/release.yml`** — the calling file that `.github` gives for its shared release:
  - it runs on `workflow_run` after CI on `main` and by the button; the shared job itself checks that CI on this commit is green and that it is a push from this repository;
  - `RELEASE_LINE`: by it the release recipes of `spec` from the submodule count the number;
  - `paths` — what the release is built from, by the list of the kind `go` that `.github` holds, with the files the binary embeds; `release.yml` is in it, so a change of `RELEASE_LINE` gives a release too. The submodule `spec` gets into no build: `git archive` does not take it, and the code takes the format from `spec`'s Go module;
  - the job has no `concurrency` of its own: the shared job itself stands in the queue `release`.
- **The recipe `release-files <version>`** — in the `justfile`, by the conditions of the shared workflow: the stage `toolchain-core` of the commit being released (sources 3.14), its tree without history and submodules, without git, under the runner's uid, with `HOME=/tmp`, without tokens (sources 3.7):
  - it takes the version as an argument and accepts only the form `v<major>.<minor>.<patch>`, with a suffix after a hyphen or without; otherwise it stops;
  - it deletes `dist/`, creates it anew and writes only regular files into it, with names by `^[A-Za-z0-9_+][A-Za-z0-9._+-]*$`, without `.sigstore.json`;
  - the CLI — `go build` for six platforms, `CGO_ENABLED=0`, `-trimpath`, `-buildvcs=false` and `-ldflags "-s -w -X main.version=<version>"`: `module-kit --version` answers with the release version. The name is `module-kit_<version>_<os>_<arch>`, for windows with `.exe`. The caller's Go environment does not get into the build: the recipe sets `GOENV=off`, empty `GOFLAGS` and `GOEXPERIMENT`, and unsets `GOAMD64`, `GOARM64`, `GOOS` and `GOARCH`;
  - the six builds run at once, a process each over the one build cache with a log of its own; the recipe waits for each and prints the log of a build that fails under its platform, and the binaries and `SHA256SUMS` are those that one build after another gives (measurements 3.1);
  - a CLI built without `-X` — for example by `go install github.com/altnovatrail/module-kit/cmd/module-kit@<tag>` — takes the version from the build information that Go puts into the binary, `runtime/debug.ReadBuildInfo`, and where there is none, answers `dev`;
  - `LICENSE` — the Apache-2.0 of `module-kit`, `LICENSE-go` — Go's license from `$(go env GOROOT)` of the same stage, and the licenses of the modules the CLI is built from, which `release-licenses` names;
  - `SHA256SUMS` — the checksums of all the files above, in the format of `sha256sum`, so that `sha256sum -c SHA256SUMS` checks the whole set at once;
  - `dist/` — in `.gitignore`.
- **What the shared workflow does next:** it signs each file with cosign without keys, puts its `.sigstore.json` beside it, checks each bundle before publishing, publishes the release as Latest and holds the published files to the release's attestation. Immutable releases are turned on for `module-kit`, and the tags `v*` are held by the ruleset "v-tags" — `update`, `deletion` and `non_fast_forward`, without bypass. `spec`'s recipe `release-verify` accepts only a signature of the shared `release.yml` at a release of `.github` no lower than the minimum that `spec/tools/release/verify.mjs` names, so the reference in `release.yml` stays at or above it.

— *Why:* through the shared `release.yml` the signature of the release is the shared workflow's, which `release-verify` checks where a file enters NovaTrail. What to build:
- six platforms — because the CLI is for module authors, and they work on macOS and Windows too, arm64 included;
- a binary without an archive: each file is signed and checked on its own, and whoever takes the CLI in CI downloads and checks one file. An archive is one more layer, which has to be unpacked, and inside it the bundle would sign the archive, not the binary;
- `CGO_ENABLED=0` — a static binary without the host's libc; `-trimpath` and `-buildvcs=false` — without build paths and without the VCS stamp, which the release tree does not have, so the build is reproducible: from the tree of `git archive` and from a working copy it gives the same checksums;
- the licenses in the release — because the binary is distributed without its sources: Apache-2.0 requires giving a copy of it, and Go's license requires reproducing its notice on redistribution in binary form (sources 3.6);
- `paths` as a list of what the release is built from — so that a change that does not change the release — a test, `.claude/`, `.gitignore` — gives none.

— *Rejected:*
- archives `.tar.gz` and `.zip` with the binary and the licenses — the reason is above; the licenses lie beside it anyway, and `SHA256SUMS` checks them;
- only linux amd64 — authors on macOS and Windows would build the CLI themselves, from the sources;
- goreleaser — one more tool with a configuration of its own, while the shared workflow signs and publishes: only `go build` is left to the recipe;
- cgo — the code does not need it, and with it the binary would depend on libc and each platform would need a C compiler of its own;
- the VCS stamp of `go build` by default — the release tree has no git, and a build from a working copy would differ from the release's build;
- a release without the licenses — it would breach a condition of Go's license;
- a `concurrency` of its own for the caller's job — the shared job and the caller's job would wait for each other;
- `paths` as the exclusions — everything but `docs/`, the files `*.md`, `.github/` and the submodule: a release would be given by changes that do not change the binary, as a test or a file of `.claude/`.

— *Consequence:* `.github/workflows/release.yml`; `justfile`; `.gitignore`; `.claude/rules/ci.md`; [architecture.md](architecture.md).

## Signatures and the work across repositories

**R08. The library of `module-kit` checks no signatures: a release file is checked by `spec`'s recipe `release-verify` where the file enters NovaTrail, and the library trusts a package by the chain of digests from the index that a pull request of `api` pinned. CI of `modules` and of `canvas` takes the CLI as a release binary at an exact tag and checks it with `release-verify` before its first run, not by `go install`. The work across repositories comes to `module-kit` by its rows of `spec`'s `COORDINATION.md`, as tasks of table 1 of RUN.md.**

— *Why:* a signature is checked once, at the border where the file enters NovaTrail, and from then on the digest the check pinned holds the file, so neither the library nor its consumers carry a check of their own. A release binary carries the shared workflow's signature, which `release-verify` checks; what `go install` builds from the sources has none to check.

— *Rejected:* the CLI for CI of `modules` and `canvas` by `go install` at a tag: nothing would check what it builds.

— *Consequence:* `CLAUDE.md`; `README.md`; `RUN.md`; [architecture.md](architecture.md).

## The module package

**R09. `module-kit` writes the canonical text of the package with its own code: one function over a tree of JSON values writes it as the [format text](../spec/format/README.md) sets the canonical serialization, and stops with an error on a value that cannot be in the format. The canonical view is built from the JSON value of each file of the module as it is. `encoding/json`, `jsontext` and the types of `formatv0` do not write the texts of the package.**
- **What is decided:**
  - the function of the canonical text gets a tree of JSON values — an object, an array, a string, an integer, `true`, `false`, `null` — and writes it by the format's canonical serialization:
    - the members of an object go in the order of the UTF-16 code units of their names, and there is no whitespace between tokens;
    - an integer is written in decimal digits, a negative one with the sign `-`;
    - in a string only `"` and `\` are escaped, the five control characters — `\b`, `\t`, `\n`, `\f` and `\r` —, and the other characters up to U+001F, as `\u00` and two lowercase hexadecimal digits. Everything else is written as it is, `<`, `>`, `&`, U+2028 and U+2029 too;
    - the text is UTF-8, with no BOM and no newline at the end;
  - on a value that cannot be in the canonical serialization, the function stops with an error rather than converting it. These are a number with a fractional part, an integer outside ±(2^53−1), a string that is not valid UTF-8 and a value that is not JSON;
  - an integer that the YAML reader gave as a double — `50.0` or `5e1` — comes into the tree as the integer `50`: so the format's canonical serialization says, and the build converts it, before the function;
  - a package is built only from a module that has passed the check. The schemas hold the numbers of the parts as integers in ±(2^53−1), and the rule `task.template-solver` holds the solver's results. So a refusal of the function on a checked module means a defect in the code of `module-kit`, not an error of the author;
  - the texts of the package — the canonical view, the listing of the device part and the inventory — and of the module index are written only by this function;
  - the canonical view is built from the JSON value of each file of the module "as it is": from what the YAML reader gave, without a pass through the types and without default values. How the parts, the tiers, the instances of templates, the inventory and the package as files are built, step by step, is in [architecture.md](architecture.md);
  - the tests:
    - CI passes all the vectors of `format/vectors/packages/` and `format/vectors/tiers/` from `spec.Format`: the digest and size of both parts and the package digest. CI computes the solver's results of the vector `with-solver` by a run of the solver of its own, and does not take them from the vector;
    - unit tests go through each refusal and the escaping table of the format's canonical serialization;
    - a fuzz test compares the function with `jsontext.Value.Canonicalize` of Go's standard library on values that the format allows. The reference lives only in the tests, in a file with `//go:build goexperiment.jsonv2`: with `GOEXPERIMENT=nojsonv2` this file is not built, and the other tests run.
- **The probe on the vectors** — in the session's folder, 2026-11-02, on Go 1.27.3; the prototype did not get into the repository. Six vectors of `format/vectors/packages/` of `spec` v0.6.3 were converted from YAML to JSON, and `server_part`, `device_part` and `inventory` of each were written in three ways. The digest and size of both parts and the package digest were checked:

  | Way | Package vectors | `1.5` | 2^53 | 2^63 | Broken UTF-8 |
  |---|---|---|---|---|---|
  | a function of its own, about a hundred lines | all six | refusal | refusal | refusal | refusal |
  | `json.Marshal`, then `jsontext.Value.Canonicalize` | all six | writes `1.5` | writes `9007199254740992` | writes `9223372036854776000` | writes U+FFFD |
  | `encoding/json` with `SetEscapeHTML(false)` | five: in `strings` the server part is 1418 bytes instead of 1415 | writes `1.5` | writes `9007199254740992` | writes `9223372036854775808` | writes U+FFFD |

— *Why:*
- the format text sets the serialization itself, so that an implementation does not depend on a library. On the format's data — ASCII keys, only integers, no surrogates — the serialization comes down to sorting the keys and escaping. The second implementation of `spec`, `tools/format-check/canonical.mjs`, is code of its own too, 44 lines, without libraries;
- a refusal is better than a silent conversion. The package digest pins the attempt, and a value converted silently would give a package that the second implementation of `spec` will not give;
- a function of its own pulls in neither a dependency nor a build constraint: anyone who imports the library builds it.

— *Rejected:*
- `jsontext.Value.Canonicalize` in the production code. In go1.27 the experiment `jsonv2` is on by default, and the package is imported without flags (sources 2.4). But:
  - it carries numbers through float64: `1.5` and 2^53 it writes, and 2^63 it rounds;
  - before it one more step is needed — `json.Marshal`, which silently replaces broken UTF-8 with U+FFFD;
  - whoever builds with `GOEXPERIMENT=nojsonv2` has no such package at all. The experiment is a property of the build of whoever imports the library, that is, of `api` and the eval runner of `tutor`;
- `encoding/json`: it always escapes U+2028 and U+2029, and `SetEscapeHTML(false)` removes only `<`, `>` and `&` (sources 2.4). This is what the vector `strings` catches;
- `github.com/gowebpki/jcs` v1.0.2 under Apache-2.0. This is a dependency, and with it testify in `go.mod`, for what a function of its own does in a hundred lines. It too carries numbers through float64, and it fixed the errors with surrogates and broken UTF-8 only 2026-09-20 (sources 4.1);
- `github.com/cyberphone/json-canonicalization`: it has no `go.mod`, the GitHub API does not recognize its license — `NOASSERTION` —, and `ci-licenses` would have to name it as an exception. The import as documented is broken (sources 4.1);
- to write the texts through the types of `formatv0` and `json.Marshal`. The fields of a struct go in the order of declaration, and the format's canonical serialization takes the JSON value of a file as it is. `spec` checks the writing back of the types only by value and only on its own corpus (`tools/types/roundtrip`).

— *Consequence:* `internal/canonical/`; `modulepackage/`; `moduleindex/`; [architecture.md](architecture.md).

## The generator

**R10. `module-kit` executes the generator and the places of randomness with its own code: a PCG stream with the DXSM output, a place from FNV-1a 64 and five functions — by the constants that the [format text](../spec/format/README.md) gives for its randomness, and on an argument outside the domain a function returns an error. Go's `math/rand/v2` and `hash/fnv` are only the reference in the tests. The generator lies in `internal/` and is not part of the library's API: what goes outside is the choice of an attempt's instance.**
- **What is decided:**
  - a stream is a generator with two words of state, high and low, taken as they are. A step is `state · M + C mod 2^128` through `bits.Mul64` and `bits.Add64`, the output is DXSM, and the constants are the numbers of the format text;
  - a place of randomness is a seed and a list of names. Its stream is started by two words: FNV-1a 64 of the UTF-8 text "the seed in decimal digits, then `|` and the name, for each name", and the seed itself. There is at least one name, and a name is a non-empty string of valid UTF-8 without `|`; otherwise the place is not built, and an error is returned;
  - the functions are five — `integer`, `below`, `between`, `fraction` and `permutation`, those that the vectors check, — and each draws as many integers as the format text says. Their domains and how they compute are in [architecture.md](architecture.md). On an argument outside the domain a function returns an error and draws no integers. There is no shuffle of a list (`Shuffle`);
  - the Go package is `internal/random`: it is not part of the library's API, and other Go modules do not import it. What goes outside is one function of the templates API — the instance of an attempt: the number `below(count)` of the place of the attempt's seed with the names `template`, the task's ID and `instance`, for count from 1 to 10,000. No other call is made on this place, and a module has no other places in format v0;
  - the generator has no shared state: a stream lives within one call, and a place is built anew for each. The code of `module-kit` has no randomness not from a seed — neither `math/rand`, nor the top-level functions of `math/rand/v2`, nor `crypto/rand`. `api` draws the attempt's seed from a cryptographic source, and `module-kit` gets it as a number;
  - the tests:
    - CI passes all the vectors of `format/vectors/random/` from `spec.Format`. A vector is read by the same YAML reader and the same schema validator as in the module check — against `random-vector.schema.json` — and by the type `formatv0.RandomVector`, and the numbers written as strings are parsed by `strconv`. When a seed and names are given, the derived words are compared with `state`. Then the calls go in order, and each value is compared, `fraction` as a double. The test fails if there are no vectors and if a call has no known field;
    - the place of a template is checked by the vectors of `format/vectors/templates/` too: for each seed they name the instance;
    - unit tests go through each refusal and the number of integers that each function draws: `below(1)` and `between(a, a)` draw one, `permutation(0)` and `permutation(1)` none;
    - the comparison with the standard library: the stream, the five functions and the words of a place are compared with `math/rand/v2` and `hash/fnv`. An ordinary test runs on every `go test`: on the words of all the vectors, on the bounds — 1, powers of two, 2^63 − 1, 2^63 + 1 and 2^64 − 1 — and on several thousand inputs that the generator itself gives from a constant seed. A fuzz test compares the same on any words, bounds, ranges and names, and permutations of up to 1024 numbers, and the same inputs lie in its corpus. These packages are in any build of Go, and the tests need no build tag. It is the second fuzz test, after that of the canonical text. If the comparison fails after Go is raised, Go has changed the stream of `math/rand/v2`: the format text names the generator as the generator of Go, and this is a question «в реестр spec», while its own code stays, as the format text says.
- **The probe** — in the session's folder, 2026-11-02, on Go 1.27.3; the prototype did not get into the repository. 12 vectors of `format/vectors/random/` of `spec` v0.6.3 were converted from YAML to JSON and read by the type `formatv0.RandomVector`:

  | Way | Generator vectors | Argument outside the domain | `GOARCH=386` |
  |---|---|---|---|
  | its own code, about a hundred lines | all 12 | an error | all 12 |
  | `math/rand/v2` and `hash/fnv` | all 12 | a panic: `Uint64N(0)`, `Int64N` of a negative, `Perm(-1)` — in `make` | all 12 — by the path of `uint32n` |

  The fuzz comparison of the two ways — 6,864,160 runs in 60 seconds on amd64 and another 20 seconds on 386 — found no divergences.

— *Why:*
- the stream of the library is set by the toolchain of whoever imports it. Which Go builds a program is decided by the lines `go` and `toolchain` of the main module (sources 2.7), and the main module is `api` or the eval runner of `tutor`. With the standard library the instance of an attempt would also depend on their Go, while the vectors pass in CI of `module-kit` only on its own. The stream of its own code is set by one version of `module-kit`, and the attempt keeps it, to repeat the verdict;
- Go promises that the streams of `math/rand/v2` will not change, but not in the package's documentation: in the blog and in a comment on a regression test (sources 2.5, sources 2.6). The format text sets the generator by numbers, "so that no implementation depends on a library", — like the canonical text, which `module-kit` also writes itself;
- an error is better than a panic: the methods of `math/rand/v2` panic on an argument outside the domain (sources 2.5), and a wrapper would check each argument itself anyway;
- its own code has one path on all platforms. In `math/rand/v2` on 32 bits `uint64n` goes through `uint32n`, and that the numbers are the same is held by the tests of Go;
- the reference in the tests catches a divergence from the standard library when Go is raised: the comparison runs on every `go test` in CI of `module-kit`, before the consumers.

— *Rejected:*
- `math/rand/v2` and `hash/fnv` in the production code — `rand.New(rand.NewPCG(…))` and the methods `Uint64`, `Uint64N`, `Int64N`, `Float64` and `Perm`. There is less code, standalone's simulator does so (sources 1.6), and `spec` calls them the normative functions of the format's randomness. But the stream would depend on the importer's Go, on an argument outside the domain the methods panic, and on 32 bits they take another path;
- an exported Go package of the generator, which `engine` or `sim` could take. This is a new link — an edit of the card «в реестр spec» — and a semver promise for one more API. The generator is one for the platform by its algorithm, not by its code: each component executes it by the format text and passes the same vectors, and `sim` reads packages without `module-kit`;
- to copy Go's references into the tests — `TestPCG` and `regressGolden` —, as the tests of `random.mjs` of `spec` do. The vector `go-pcg-1-2` takes the first six of the twenty values of `TestPCG`, and the comparison with the standard library on every `go test` checks the words 1 and 2 too, from which Go's references start, and much beyond them;
- the shuffle of a list, `Shuffle`: it has no vectors, and `module-kit` shuffles nothing. If it is needed, it will come with the code that needs it, by the format text.

— *Consequence:* `internal/random/`; `template/`; [architecture.md](architecture.md).

## The Starlark sandbox

**R11. `module-kit` executes Starlark in go.starlark.net of the version that the [format text](../spec/format/README.md) names, which `go.mod` pins and Dependabot `gomod` does not raise. The clock is a safeguard of `module-kit` with a margin over the budget of steps: 1 s for the evaluation of a predicate and 60 s for a template solver. An evaluation runs in the caller's thread and with the caller's `context`, and the library holds no slots. Until stage 4 Starlark runs in the process; from stage 4 the Starlark of a third-party module runs in a separate process with a memory limit, which `api` and `module-registry` provide.**
- **What is decided:**
  - the interpreter is go.starlark.net of the version that the format names, and raising it is an edit of the format, not of `module-kit` alone. In this version the steps are counted by backward jumps and calls, the depth of recursion is bounded, and the methods of `set` do not panic (sources 4.2). go.starlark.net is excluded from Dependabot `gomod`, like the module `spec`;
  - the dialect is the format's one for all its Starlark, a check predicate's too: `Set`, `While`, `TopLevelControl` and `GlobalReassign` are on, `Recursion` is off, `load` is forbidden, and the built-in names are the format's list of 33. The limit of steps fires when the count reaches it;
  - the clock — constants of the release of `module-kit`, a safeguard with a margin of no less than 10× over the time of the budget of steps on the machine of the measurement:
    - a predicate — 1 s for one evaluation. When it fires, the clock gives "unchecked";
    - a solver — 60 s per template: for the top level of the program and all the calls of `solve` together, like the budget of steps. When it fires, the clock gives a breach of `task.template-solver` with the numbers of the instance on which the solver stopped.

    The clock calls `Thread.Cancel` by a timer, and the interpreter sees a cancellation between instructions: it does not interrupt a built-in function (sources 4.2);
  - the caller gives the `context`. Its cancellation is `Cancel` too, but with another reason, and it is an error without a verdict: a predicate does not get "unchecked", and the module check does not get a breach;
  - the library holds no slots. An evaluation runs synchronously, in the caller's thread, on its own `starlark.Thread`, and the sandbox has no shared state. There is no waiting for a slot either, which would give "unchecked" from the load. How many evaluations run at once is held by the caller: `api` — no more than `GOMAXPROCS` at once, the CLI — template solvers one per CPU (`GOMAXPROCS`), each on its own thread. The CLI prints the summary in the order of the templates, not in the order of completion. The clock runs by wall time, so the caller holds no more concurrent evaluations than `GOMAXPROCS`: otherwise the clock of a predicate would depend on the load;
  - a panic of the interpreter is caught by `recover`: for a predicate it is "unchecked", for a solver a breach of `task.template-solver`. `recover` does not catch an overflow of the Go stack or a lack of memory. The recursion of Starlark functions is held by the recursion that is off. Built-in functions on deeply nested values go along the Go stack and for long — `str` of a list 100,000 deep takes 1.6 s on one step (measurements 1.1) —, and they, like memory, are held only by a separate process from stage 4;
  - go.starlark.net does not count memory: a built-in function in one step takes up to 1 GiB by repetition and any amount by `list(range(n))` (sources 4.2). At stages 1–3 the modules are only official (О-9), and Starlark runs in the process: the risk of memory is accepted — Д10. Third-party Starlark from stage 4 — the solvers in `module-registry` and the predicates in `api` — runs in a process with a memory limit, which is not available to the library, computation without input or output: how is for the tasks of `api` and `module-registry`.
- **The measurement** (measurements 1.1): on the solvers of standalone, by which the format's limits were calibrated, the latest version and standalone's counted about the same — a median ratio of steps of 0.96, the heaviest solver 2,725,355 steps against 2,413,617, the longest run 25.3 ms; a predicate of the whole budget of 1,000,000 steps spends 6 ms of its 1 s, and a solver of the whole budget of 100,000,000 steps 0.45 s of its 60; on standalone's version `set([1]).difference()` panics. A predicate that calls `sorted(range(1000))` on each turn reaches the limit of 1,000,000 steps at its 50,000th turn after 932 ms, and a solver of the same kind reaches 93,898,910 steps in 60 s, and the clock stops it.

— *Why:*
- the format's limits of steps are set "as go.starlark.net counts them", and the count of go.starlark.net changed 2026-10-05 (sources 4.2): the same limit in its two versions is different work, and two implementations of the format would diverge on one module. The latest version — because on the solvers of standalone the measurement showed no big difference, for predicates the new count is no dearer than the old, and the panic of the methods of `set` is fixed in it;
- a verdict is repeated by the stored observation by the release of `module-kit` that gave it, and the acceptance of a template must not depend on the machine of CI. So the clock is a safeguard far above the time of the budget. The clock fires before the steps only for a program whose steps call heavy built-in functions, and only for it does the outcome depend on the machine: a predicate with `sorted(range(1000))` on each turn reaches the clock before the limit already on a machine slightly slower than the machine of the measurement. This is the accepted price of the safeguard: a predicate of one integer does not need such steps;
- the library is computation without input or output and may run on the device. It has no shared semaphore and no waiting: concurrency is a property of the process that the caller owns. Waiting for a slot, as in standalone (sources 1.1), would give "unchecked" from the load;
- the format leaves the clock and concurrent runs to `module-kit`;
- one dialect for all the Starlark of the format: the author writes a predicate and a solver in one language, and the recursion that is off protects the Go stack for a predicate too (sources 1.1);
- a separate process is needed when Starlark is written by a third-party author — from stage 4 (О-9). Until then the project checks the modules, and a process costs a start, an exchange and an implementation of its own for each OS of the CLI.

— *Rejected:*
- to pin the version of standalone, 89a6a09411d5 (sources 1.4). The limits would stay as calibrated, but the panic of the methods of `set` would stay too, and `module-kit` would fall behind the fixes;
- to leave the format without a named version: the same limit in two implementations of the format would mean different work;
- `module-kit` holds the slots, like the sandbox of standalone, `Limits{Steps, Timeout, Concurrency, Wait}`. This is shared state in the library and "unchecked" from the load; standalone itself rejected four slots with a safety clock: "Verdicts depend on the load" (sources 1.3);
- `api` holds everything, and `module-kit` has no clock: the format names the clock of `module-kit` and leaves it to it;
- 2 s per run, as in standalone (sources 1.1). A solver with heavy steps went through 93,898,910 steps in 60 s on the machine of the measurement, and in 2 s it would go through about 3,000,000: the clock would stop it deep inside the budget, and the acceptance of a template would depend on the machine;
- without a clock: the built-in functions on each step would run without a time limit;
- a separate process in the CLI now: this is an implementation of its own of the memory limit on linux, darwin and windows, and the predicates in `api` stay in the process anyway;
- zero options for a predicate: two dialects for one format, and a predicate would not know `set`, which a solver knows;
- `GOMEMLIMIT` or `debug.SetMemoryLimit` from the library: this is a setting of the caller's whole process, not of the sandbox, and the limit is soft — the garbage collector will not stop one built-in function. Whether to set it is decided by the process, that is, by `api`;
- a price in bytes for built-in functions, as for the helpers of standalone: the built-in functions are the code of go.starlark.net, not of `module-kit`, and the methods of the values themselves bypass it — "only a limit on the memory of the process can" (sources 1.2).

— *Consequence:* `internal/sandbox/`; `go.mod`; `.github/dependabot.yml`; [architecture.md](architecture.md).

## A move to a tag of `spec` before the task that needs it

**R12. When a task needs what a newer tag of `spec` has decided, the move to that tag comes first, as a task of its own, and takes the latest release of `spec`, checked by `gh release verify`; the task then goes on the new tag. A task does not run on the older tag with the newer one's decisions taken as external facts, and the move is not made inside the task that needs it.**

— *Why:* a repository reads `spec` at the tag its submodule is on, and `docs-check-repo` checks a citation of `spec` only against that tag. A task on the older tag would decide anew what the newer tag has already decided, and its decision would have to be brought into agreement with `spec`'s later. The latest release brings with it the other answers of `spec` that came since, and a release of `spec` is immutable, so its attestation pins what the submodule takes.

— *Rejected:*
- the task on the older tag, with the newer tag's decisions as external facts with a source: the documents check would not check the citations of a tag the submodule is not on;
- a tag below the latest that already holds what the task needs: the move would leave behind the answers that came after it, and the next task would move again;
- the move inside the task that needs it: a repository moves to a newer tag in a task of its own, whose commit says why.

— *Consequence:* `spec/`; `go.mod`; `CLAUDE.md`; `RUN.md`.

## Version comparison

**R13. `module-kit` compares a new version of a module by packages, not by the author's YAML, with each published version below it by semver, withdrawn ones too, by the rules of version comparison of the [format text](../spec/format/README.md), `module-version.*`, and has no rules and no table of its own. The comparison itself finds the lowest version part that has to be raised, and the author may raise a higher one, but not a lower one. Major — when something of what the learner had in the previous version is gone or means something else; a transition to a new format version, while the format is v0, is major. The CLI has a command of its own for the comparison — `diff`.**
- **What is decided:**
  - the principle of the version parts, the table "change — version part", the rule of 0.y.z, the comparison of a check by meaning, the instances of a template by the values of the parameters and the rules of comparison are the format's, with the vectors `vectors/versions/`. The format took them from this decision: major — an entry with a full ID is gone, the audience has narrowed, a skill no longer teaches an anchor skill, the check of a task that stayed has changed, the module has moved to a format version that may change verdicts, a new entry of `retired_skills`; minor — everything previous is in place and means the same, and the module gives more or leads otherwise; patch — the same entries, checks and path, other words or pictures; and a part may also be "none". How `module-kit` reads them is in [architecture.md](architecture.md);
  - the author declares only the version number, and a change of an entry's meaning — by a new ID: for a skill the format says so, for a task, a lesson, a trap and a technique this is a removal and a new entry;
  - packages are compared — the canonical view and the listing of the device part. The new version is the module's folder, which `diff` checks and builds in memory, or a package built by `build`; the published ones are packages checked by the chain of digests from the index. `eval/`, comments and the order of keys do not change the version: the package is the same;
  - a slot check is compared by meaning, as its check kind reads it: for `number`, `fraction` and `choice` — the key, `units`, `form`, the options and the trap values after normalization, and the predicate of `predicate` — as a string. For a template the instances are matched by the values of the parameters, and the checks of each instance with its numbers are compared;
  - the rules are eight, `module-version.*`, the eighth holding the bands, `module-version.bands-kept`; the same version with another digest is a refusal of `module-version.above`. `diff` requires the package of each version that the index names: the rules of the whole history — by all of them, and the part found — by the newest. A withdrawn version is a published one, and the comparison does not read `withdrawn`;
  - a transition to a new format version: while the format is v0 — major, from v1 — the same part as that of the format's step. Across a transition `diff` checks the skills and the IDs of entries by the places of the package that the format keeps in all versions, and reads the inventory, `languages`, the skills and `retired_skills` of a package of a previous format version by them;
  - in 0.y.z, as in Cargo, y plays the major, and z the minor and the patch. In 0.0.z each version is a major of its own: any edit raises z. The first version may be any;
  - `diff` is a command of the CLI over one function of the library: a module or a package, a checked index or an explicit "none published", the folders of the published packages. Code 0 — accepted, 1 — a refusal with the IDs of the rules, 2 — no comparison is possible; the report — the part found and the changes by paths. The CLI does not go to the network and checks no signatures.

— *Why:*
- the part is found by the comparison, as with gorelease and cargo-semver-checks (sources 5.3): a part chosen not by the author but by the content will not diverge from the content. Semver lets the author go higher: a major "MAY also include minor and patch level changes" (sources 5.1);
- major is "is gone or means something else": within one major nothing disappears for the learner, and `api` can move them to a new version itself, and `api` decides which version a new attempt gets. In the same way a removal is counted as major by semver, Frictionless — "delete or change the meaning of data" — and Elm, in which what is removed gives `M.MAJOR` (sources 5.3);
- a command of its own: the comparison needs other inputs than the check — the packages of the published versions, which CI downloads and checks — and a report of its own, by which the author chooses the version. `spec` calls the comparison `diff` too;
- a format transition is major while the format is v0: a step of 0.y promises nothing, and a transition that changes verdicts is akin to a change of the key of a published task. The chain of comparisons must not break at a transition: `learner-model` transfers nowhere a skill that is in neither `skills` nor `retired_skills` of the current version, and there is no error in that;
- packages, not YAML: the canonical view is the JSON values of the files as they are, and everything that does not go into the package does not concern the learners;
- the comparison of a check by meaning: the verdict depends on the check only through these normal forms, and `module-kit` holds the normalization for the verdict anyway. Only the predicate stays a string: the equality of two Starlark programs cannot be decided in the general case;
- the instances by the values of the parameters: the learner got a task, not an instance, and the attempt keeps its package and seed;
- all the published versions: the comparison with the newest suffices by induction only if each publication passed `diff`, and the IDs of tasks, lessons, traps and techniques have no table of their own, like `retired_skills` for skills, so a rule over the whole history sees by the newest alone only the IDs gone in it;
- 0.y.z by Cargo: semver says of 0.y.z "Anything MAY change at any time" (sources 5.1), and Cargo and oasdiff give y the role of the major (sources 5.2, sources 5.3). `api` and `diff` must read "the same major" alike, so the rule is the format's.

— *Rejected:*
- the author declares the part themselves, and the comparison checks only `retired_skills` and the growth of the version, as in Frictionless and Conventional Commits (sources 5.3): the part could diverge from the content;
- exactly the part found, as in Elm: the author could not raise the major for the sake of what the comparison does not see, although semver allows it;
- major only from what breaks progress — the removal of a task, a lesson, a trap or a technique would be minor. Then `api` could not move a learner within a major itself: the lesson in which they stand could disappear;
- the comparison as part of `check`: `check` checks one module, and the comparison needs the packages of releases, which CI downloads and checks separately;
- the comparison only at the index build: the author would learn of an error at the release, after the merge;
- across a format transition to check only the growth of the version: nobody would check the retired skills in the pull request of the transition;
- a format transition is minor: in v0 a step of the format may change verdicts;
- to compare the author's YAML: a comment or the order of keys would require a new version, although the package is the same;
- strict equality of the fields of a check: `"2.50"` and `"2.5"` or rearranged trap values would give a major where no verdict changes;
- to count an added unit or trap value as minor, as an extension: some answer changes its verdict or its trap, and the author chose "the check has changed — major". Besides, an exact check of an extension is subtle: in `units` the longest unit at the end wins;
- to match the instances by numbers: a new constraint of a template shifts the numbers, and an edit that changes the check of no instance would become major;
- to compare only with the newest version: see "Why";
- fixes of an old line, below the newest version: the module's folder has one line of versions, and a withdrawn newest one is fixed by a version above it. A version below the newest would have to be checked against the versions above it too, and against the rule that each version above names its skills. It will be needed if `api` starts keeping a learner on an old major: then it is an edit of the format's rule that a new version stands above all the published ones;
- for 0.y.z — the same rules as for X.Y.Z: a module in development would go to 1.0.0 with the first retirement. Or 0.y.z without rules, by semver: then `api` and `diff` would read "the same major" differently.

— *Consequence:* `moduleversion/`; `cmd/module-kit/`; [architecture.md](architecture.md).

## The package as files

**R14. `build` writes a package as files, and `module-kit` builds no archive of it: the package's archive for a release is put by a recipe of `modules`. The withdrawals that the index build reads come as the file `withdrawals.yaml` of `modules`, by `$defs/withdrawals` of the index's schema. The container's buf, which the template of `spec` brings into every repository, stays in the Dockerfile, though `module-kit` has no contracts to build with it.**

— *Why:* the card of `module-kit` in `spec` says so: "build" writes the package as files, and "build the index" reads the withdrawals from the file of `modules`. The name of a package's file in a release of `modules` is derived from the entry of the index, so the release's form is `modules`' to keep, and `module-kit` only reads and writes a module's files. buf comes into the copy of each repository, like Go, and the container of `module-kit` is the template's.

— *Rejected:*
- an archive of the package written by `build`: a second writer of the release's form, which would have to agree with the recipe of `modules`;
- the withdrawals as arguments of the command: the index's schema holds their form, and the file of `modules` is checked by it;
- buf removed from the Dockerfile: the container would diverge from `spec`'s, and the difference would have to be recorded as a decision, with no reason for one.

— *Consequence:* `modulepackage/`; `moduleindex/`; `cmd/module-kit/`; `.devcontainer/Dockerfile`; [architecture.md](architecture.md).

## The command `open`, the exit codes and the version of an attempt

**R15. The CLI of `module-kit` has a command of its own for open views, `open`, for the samples page build of `canvas`. It reads the index of a release of `modules`, the index of the newest release — for the marks of withdrawal — and the packages. It writes `open-views.json`, for each task an entry with its open view and the digest of it, and for a template an instance for each seed it is given; beside it, the files of the device part of each version — the media files of the tasks, `LICENSE` and `NOTICE` —, checked against the inventory; and it skips withdrawn versions.**
- **What is decided:**
  - the inputs — the index of the release, which CI of `canvas` checked with the recipe `release-verify`, the index of the newest release and the folders of the packages, which CI downloaded by the entries of the index. The command knows a package by the digest of the bytes of `inventory.json` and checks it by the chain, and the media files — by the listing of the device part;
  - for each module — the newest by semver version within the format's window that the newest index has not marked withdrawn. The rule is the same as for the version of a new attempt in `api`, and the library holds it;
  - the output — `open-views.json` with `format`, `version` and the modules, each with its ID, version, `module_format_version`, package digest, the `release` of its entry in the index and the open view of the manifest, and for each task an entry `{"open_view_digest", "task"}`, with `"seed"` for an instance of a template, as a canonical text; the files of the device part — in `modules/<ID>/<version>/` with the layout `device/`, `LICENSE` and `NOTICE` among them: the device part carries them to everyone who receives the content;
  - for a template the command gives the instance by each seed it is given, with the numbers substituted, and the digest of an open view is `sha256:` and the SHA-256 of its canonical text, as the eval protocol sets them: by this digest the eval runner knows the task of a sample;
  - each module gets one version — the newest unwithdrawn one: taking all the unwithdrawn versions would mean repeating the module's tasks on the page.

— *Why:* the samples page build takes the tasks and the media files from a release of `modules` through the CLI of `module-kit`. The open view is the format's semantics, and `module-kit` builds it by the marks `x-open`: `canvas` does not repeat it on its side, nor the rule of withdrawal. `canvas` expects from the CLI the open views, the media files, the ID and the version of the module and the package digest. The form of the output follows `tutor-open-tasks` v0, which CI of `canvas` already reads, and the types are the same, `types/formatv0/open` of the module `spec`. How a sample names a task is the eval protocol's, so the instance by the seed and the digest are its.

— *Rejected:*
- a narrow command — the open view of one part of one package to stdout: the loop, the withdrawals and the check of the media files `canvas` would repeat on its side;
- `canvas` builds the open views itself by `server.json`: the format's semantics would diverge from `module-kit`;
- all the instances of a template: up to 10,000 open views per template;
- the instances by numbers: the eval protocol names an instance by the task's ID and the seed.

— *Consequence:* `cmd/module-kit/`; `openview/`; `moduleindex/`; [architecture.md](architecture.md).

**R16. The exit code has one meaning for all the commands of the CLI: 0 — accepted, 1 — a refusal by the format rules, with their IDs, 2 — the command could not do its work. What is absent — a previous index, withdrawals, published versions — a command is told explicitly, not by a missing file. The version of a module for a new attempt — the newest by semver unwithdrawn one within the format's window — is chosen by a function of the library, which `api` and the command `open` use. A library that the code takes is chosen with its options, its license and its exact version named.**

— *Why:*
- one meaning of the exit codes: CI of `modules` and `canvas` tell a refusal of a module from a failure without reading the output. An explicit "none" — as with the build of an index without a previous one: a missing file will not pass silently;
- the version of a new attempt — by a rule of the library: the semver rule is one, and `module-kit` holds it, as it holds `diff`;
- a library's options, license and exact version — because a dependency must carry a license compatible with Apache-2.0 and stand at an exact version, and what the library takes, every consumer of it builds.

— *Rejected:*
- an exit code of its own for each command and an empty path as "none": CI would have to know each command, and a missing file would pass silently;
- the version of a new attempt chosen by the code of `api`: then there would be two semver rules.

— *Consequence:* `cmd/module-kit/`; `moduleindex/`; [architecture.md](architecture.md).

## One definition, in the format

**R17. Where the format text or the contracts of `spec` define a rule, a table or a vector, `module-kit` executes it and keeps no definition of its own beside it: version comparison, the dialect and version of Starlark, the observation and the join of two readings. The code is written for the format version of the tag that the submodule is on. What the platform gives to another component, `module-kit` does not do: the check of a sample file is `spec`'s recipe `sample-check`, and the admission table for AI is executed by the admission function of `tutor`.**
- **The observation** is the contract `novatrail.observation.v1` of the nested Go module `spec/contracts`, at the tag `contracts/vX.Y.Z`. `module-kit` joins two readings into one, by the slot check, and gives the verdict by the joined observation, by the vectors `contracts.Vectors`. With the same code it gives the eval runner of `tutor` the verdict by the truth of a sample's markup, and for the admission of a model only a report with the verdicts of `module-kit` is valid.

— *Why:* one definition is in the format, and the vectors hold it: a table of `module-kit`'s own beside it would diverge from it, and two implementations of the format would give different verdicts. The window of versions holds only the current format version while the format is v0, so code for an earlier one would be code of a format that the platform no longer reads.

— *Rejected:*
- keeping a table of parts and rules of comparison of its own in [architecture.md](architecture.md) beside the format text;
- writing the code of an earlier format version and moving to the current one later: the window rejects a module of an earlier version, and `modules` writes modules of the current one;
- the check of a sample file by the CLI of `module-kit`: the eval protocol gives it to `spec`'s recipe.

— *Consequence:* [architecture.md](architecture.md); `observation/`; `verdict/`.

## CI for the format's code

**R18. Test coverage goes to Codecov from the shared `go.yml` of `.github`: the job "Go" calls it with `id-token: write`, and `ci-test` writes the profile `coverage/go.out`, every test counting for every package of the module. `.github/codecov.yml` is the text of the template of `.github`. The release is built by the list `paths` of the kind `go`. Every reference to `.github` stands at one release, by the SHA of its tag, which `ci.yml` pins, and Dependabot raises them together, by the group `shared-workflows`; a reference that a task needs before Dependabot proposes it is raised by hand, in the task's own pull request.**
- **Coverage:**
  - `ci-test` runs `go test -race -count=1 -covermode=atomic -coverpkg=./...` three times: the module but for the tests at the clock of a predicate, then `internal/sandbox` and `verdict` one at a time, then the verdicts of a predicate of `TestVectors` alone; `coverage/go.out` is the first profile with the blocks of the other two after it, their mode lines dropped — a profile of several packages, as `go test` writes it. A package without tests is in it too, with zeros, and the lines that the vector test in `internal/vectors` reaches count for the Go package where they lie (sources 2.9). `.gitignore` holds `coverage/`;
  - the job `test` of the shared `go.yml` gives the profile as an artifact, and its job `coverage` — the only one with `id-token: write` — sends it to Codecov by OIDC, without a secret (sources 3.9). "Go / coverage" does not hold the merge; what holds it is `codecov/patch`, which the author puts in the ruleset `main`, with `integration_id` 254, as for `spec`;
  - `.github/codecov.yml` — the threshold of changed lines is 90%, `codecov/project` only reports, and a comment comes only when the coverage changes. There is one upload, so `after_n_builds` is not needed. Codecov itself shortens the import path to the path in the repository;
  - the coverage badge is in the README, after the badge of the Go version.
- **The list `paths`** of `release.yml` is the list of the kind `go` that `.github` holds: `*.go` without `*_test.go`, `go.mod`, `go.sum`, `justfile`, `LICENSE`, `.devcontainer/Dockerfile`, `release.yml` itself and the files the binary embeds.
  - A release is given by the code without tests, `go.mod` and `go.sum`, the recipes, the license, the Dockerfile and `release.yml`, the change of `RELEASE_LINE` included.
  - None is given by the tests, `devcontainer.json`, `.gitignore`, `.claude/`, the documents, the rest of CI and `.devcontainer/gopls/` without the Dockerfile: gopls is raised together with `GOPLS_VERSION` in the Dockerfile, and the release comes with it.

— *Why:* the format's code comes with its tests, and the merge with `codecov/patch`, at least 90% of the changed lines covered, and this stands before the first line of the format's code. The list is needed so that the change of `RELEASE_LINE` gives a release and so that edits that do not change the release give none: Dependabot would propose each such tag to `api` and `tutor`. The references to `.github` at one release: its shared workflows and its composite action are made to work together at one release, and `ci.yml` fails when they stand at more than one SHA. A reference that a task needs is raised by hand, because the task's pull request would otherwise show neither its checks nor their report, and a pull request of its own carries both the reference and what it needs at once.

— *Rejected:*
- waiting for Dependabot for a reference that a task needs: the argument above. The order of the template — first the report, then the link — is needed when the link is raised by Dependabot: a pull request of the repository's own carries both steps at once;
- an older release of `.github` with the same file: the template says to take the latest release of `.github`;
- the badge after the first upload on `main`, by a separate pull request: one more pull request for a few minutes of "unknown" after the merge;
- excluding `testdata/` from the list in advance: there is no code there, and the exclusion will be added by the task that puts code there.

— *Consequence:* `justfile`; `.gitignore`; `.github/codecov.yml`; `.github/workflows/ci.yml`; `.github/workflows/release.yml`; `.github/dependabot.yml`; `README.md`; [architecture.md](architecture.md).

**R19. gopls stands in the devcontainer image at an exact version, `GOPLS_VERSION` of the Dockerfile, and is built from its own lock file `.devcontainer/gopls/`: the step and the lock file are those of the kind `go` of `spec`'s template, and come with the rest of its devcontainer on a move to a tag of `spec`. `ci-lint` holds `GOPLS_VERSION` and the lock file to each other. The same gopls serves the sessions of Claude Code, through the plugin `gopls-lsp` that `CLAUDE_CODE_PLUGIN_DIRS` of `devcontainer.json` loads from `spec/claude-plugins/`.**
- **What is set up:**
  - `.devcontainer/gopls/go.mod` — the module `devcontainer/gopls` with `tool golang.org/x/tools/gopls`, and the `go.sum` beside it holds the hash of each module of the build; they are raised by `go get -tool golang.org/x/tools/gopls@v<version>` in that folder, together with `GOPLS_VERSION`;
  - in the stage `devcontainer` of the Dockerfile:
    - `ARG GOPLS_VERSION`, a `COPY` of the lock file and `go install -mod=readonly tool` into `/usr/local/bin`. A module without a hash or with another hash stops the build, as a wrong sum stops an archive;
    - `gopls version` must name `GOPLS_VERSION` as a whole line, `grep -xF`, so `v0.24.10` will not pass for `v0.24.1`;
    - everything the step writes stays in one temporary folder, and the same layer removes it: the cache of modules and of the build, and through `XDG_CONFIG_HOME` the telemetry counters of go and gopls too;
    - the stage `toolchain` does not change: the checks build it without a context and do not reach the `COPY`. So `ci-lint` checks `GOPLS_VERSION` in the Dockerfile against the version of gopls in the lock file. A pull request that raised one without the other, and a move to a tag of `spec` that lost the step, go red on "Go / lint", not on Rebuild Container;
  - `devcontainer.json` — the Go extension does not install or update its tools and does not lint on save: `go.toolsManagement.autoUpdate`, `go.toolsManagement.checkForUpdates` and `go.lintOnSave`. Linting is done by `just ci-lint`. The comment to `context` says that the image takes only `gopls/` from its folder;
  - `dependabot.yml` — `.devcontainer/gopls/` is not in the ecosystem `gomod`: Dependabot would raise gopls without `GOPLS_VERSION`, and the check would stop the build. Dependabot alerts and security updates are on for the repository, and the dependency graph reads each `go.mod`, this one too. A vulnerability in a module of the lock file is a finding before the next task; its pull request is checked as the repository's own. One that raises a dependency of gopls is merged; one that raises gopls itself goes red on `ci-lint` until `GOPLS_VERSION` is raised in it too;
  - the license of gopls is BSD-3-Clause ([LICENSE golang/tools](https://github.com/golang/tools/blob/master/LICENSE); the GitHub API, `repos/golang/tools/license`, read 2026-11-04). gopls is a development tool: it is shipped nowhere and puts none of its own code into what ships;
  - the plugin `gopls-lsp` starts the gopls of `PATH` and after each edit of Go code gives the session its diagnostics.

— *Why:* without gopls in the image the Go extension installs the latest version itself, not an exact one. `go install …@v<version>` in the Dockerfile SonarCloud counts as the finding docker:S8545 (sources 3.4), and the lock file fixes it by code. `tutor`, with the same step, has no findings S8545 (sources 3.10). A second `go.mod` without packages CodeQL autobuild passes, writing "No packages found.", and the `./...` of the recipes does not go into it. The step is the template's, so every repository in Go has it from one place.

— *Rejected:*
- installing the lock module by `postCreateCommand`: each creation of the container would go to the network, and gopls would land in `~/go/bin`, which is not in `PATH`;
- leaving gopls to the Go extension — the argument above;
- `go install …@v<version>`, with `-mod=readonly` or without: S8545 marks both;
- marking the finding S8545 as false: a false positive is the last resort, and the lock file fixes it by code.

— *Consequence:* `.devcontainer/Dockerfile`; `.devcontainer/devcontainer.json`; `.devcontainer/gopls/`; `.github/dependabot.yml`; `justfile`; [architecture.md](architecture.md).

## The modules of `spec` and the vector test

**R20. The code takes the format and the vectors as the Go modules of `spec` — `github.com/altnovatrail/spec` and `github.com/altnovatrail/spec/contracts` —, at the tag of the submodule and outside Dependabot, and `ci-lint` checks them against the submodule. The library is Go packages by the parts of the format, and the mechanisms are in `internal/`. The vector test is one for all the cases, `internal/vectors`; it fails on a kind that it does not know and skips no case.**
- **The modules of `spec`:**
  - `go.mod` requires both at the version of the submodule's tag; the nested module's tag is `contracts/v<version>` at the same commit, and Go finds the tag with the folder's prefix by the version itself;
  - Dependabot `gomod` does not raise them: `ignore` of both by name, as `tutor` excludes `spec`;
  - `ci-lint` checks three values: the commit that the git index recorded for `spec/` (`git ls-files --stage -- spec`), and the commit that the Go module proxy names as the origin of the version of each module (`Origin.Hash` in `go mod download -json <module>@<version>`, sources 2.8), — and that the version of `contracts` is the same as that of `spec`. The submodule is not needed for the check: in CI the index is the tree of the commit, and the checkout of the shared `go.yml` does not take the submodule. Here the index holds the move to the tag as soon as `spec/` is added, even before the commit. A move that shifted one side without the other goes red on "Go / lint", and a proxy that gave no origin the recipe names. A probe: with `spec` or `contracts` a tag lower in `go.mod`, and also with the submodule at that lower tag added to the index, the recipe fails and names all three values;
  - a move to a new tag of `spec` moves the submodule, both modules and the copy of the schemas `internal/schema/schemas/` — from the Go module cache — together, and the new vectors of the tag come with their code: otherwise the vector test, or `TestSchemasAreSpecs`, is red.
- **Go packages** — each with a `doc.go`, which says what it holds and who calls it; the packages, what each holds and who calls it are the table of [architecture.md](architecture.md). A shared internal Go package — the parsing of semver, the parsing of numbers — is started by the task that needs it a second time. `package` is a Go keyword, so the module package is named `modulepackage`, and with it `moduleindex` and `moduleversion`, like the prefixes of the rules `module-index.*` and `module-version.*`.
- **The vector test** — `TestVectors` in `internal/vectors/vectors_test.go`:
  - it sorts each file of `spec.Format` and `contracts.Vectors` into cases by the kinds that the format text and the contracts text set, each folder with exactly what the format requires, nothing extra and nothing missing. The text, `VERSION` and the schemas are not cases, `eval/v1/` are vectors of `canvas` and `tutor`, and a file that no kind takes is a refusal with its path. A new kind, a new part or a new check kind comes with a move to a tag, and the test goes red until the task of the move knows it;
  - a case is a subtest named by its path in `spec`: `format/vectors/rules/yaml.encoding/byte-order-mark.yaml`, `contracts/vectors/observation/v1/boxes.json`. Each kind is passed by its runner from the table `runners`, and a case that does not pass fails the test;
  - the cases run as parallel subtests, but for the verdicts of a predicate, `format/vectors/verdicts/predicate/`, which run one at a time and before the rest; `ci-test` runs those in a `go test` of their own, after the rest of the module's tests;
  - the tests of the test itself — `inventory_test.go`, on `fstest.MapFS`: each kind and the path of its case, the refusal on a file of each unknown kind and on a folder that lacks something.
- **Coverage** — `ci-test` with `-coverpkg=./...`. By default Go counts for a package only its own tests (sources 2.9). Without it, `codecov/patch` would count as uncovered the lines of `verdict` that the vector test reached. The profile stays one file, `coverage/go.out`: a block that the tests of two packages covered is in it twice, and `go tool cover -func` reads it.

— *Why:*
- Go packages by parts: `api`, the eval runner of `tutor` and the CLI call different things, and each takes its own part, and the names follow the format's terms and the prefixes of the rules. Mechanisms that have no consumer of their own are in `internal/`, as the generator;
- one test for all the cases and `-coverpkg`: the test itself sees each case and each kind, and coverage counts where the code lies;
- no skips: each release passes all the vectors of its tag, and a vector that does not pass holds its pull request;
- the check in `ci-lint`: the modules and the submodule move by one move, and the check catches a move that shifted one of them. In the same way `ci-lint` holds `GOPLS_VERSION` with the lock file of gopls.

— *Rejected:*
- one Go package at the root of the module, `modulekit`, with the parts in `internal/`: a facade of wrappers and type aliases for each part, and a consumer takes one of its parts anyway;
- the vectors in the tests of each Go package through a shared helper, without `-coverpkg`: the central test would not see whether a package executed its cases, and a forgotten kind would go unpassed silently;
- checking the version of the modules by a Go test: the checkout of the shared `go.yml` does not take the submodule, and the code does not read `spec/`;
- not checking, but holding it by the move to a tag and the review: a move that forgot one of the sides would pass green.

— *Consequence:* `go.mod`; `go.sum`; `justfile`; `.github/dependabot.yml`; `internal/vectors/`; `internal/schema/schemas/`; [architecture.md](architecture.md).

## The YAML reader

**R21. A file of a module is read by `go.yaml.in/yaml/v4` into a tree of nodes, and the package `internal/yaml` goes around it in passes: they read as YAML 1.2.2 where v4 keeps to YAML 1.1 or lets through what YAML 1.2.2 rejects, and they check the rules `yaml.*`. The value is a JSON value with an integer as `int64` and a double as `float64`, and a refusal names each breach with its rule ID and place, or says that the file is not YAML. The vectors `rules/yaml.*` are passed, all the files of the format are read, and the YAML test suite lies in `testdata` and holds the reader together with a fuzz test.**

- **What to read with:**
  - `go.yaml.in/yaml/v4` — the YAML organization, Apache-2.0, and the files carried over from libyaml are MIT, by its `NOTICE`; it has no dependencies (sources 4.3, sources 4.4);
  - on the YAML files of `spec.Format` outside `rules/yaml.*`, each of which the format reads without breaches, v4 refuses not once. v3 refuses on 195 of them: a plain scalar with `?` in a flow collection it cuts off, by YAML 1.1. goccy/go-yaml refuses on the valid complex keys `? {a: b}` and `? [a, b]`, and does not check UTF-8 (measurements 1.2).
- **Where v4 diverges from YAML 1.2.2 — and which pass closes it** (YAML 1.2.2 — sources 5.4). Each pass is in its own file of the package, and a test holds each row:

  | Where v4 diverges | Pass | Test |
  |---|---|---|
  | a BOM at the start it lets through silently, after `...` it is a parse error; `%YAML 1.2` it rejects: its parser expects 1.1 | the level of the stream (`text.go`): a BOM is `yaml.encoding`, a line beginning with `%` is `yaml.no-directives`; the BOM is removed, and `%` becomes `#` of the same length, and a directive not followed by `---` is "not YAML" | `TestRead`, "byte order mark…", "directive…" |
  | invalid UTF-8 and control characters it rejects with no place | the bytes (`text.go`): `yaml.encoding` at the first invalid byte; a character outside the printable set of YAML 1.2.2 is `yaml.printable` with a place, in a comment and in quotes too; a file not in UTF-8 is not read further: keys that differ only in invalid bytes would merge | "not UTF-8", "control character…" |
  | U+0085, U+2028 and U+2029 it counts as a line break, as YAML 1.1 does: `a: x<U+2028>  y` it reads as `"x y"`, and YAML 1.2.2 as `"x   y"` | replacement by characters of the private use area that are neither in the text nor among its escapes, and back in the values (`text.go`) | "line separator…", "next line…", "a stand-in written as an escape" |
  | an escaped surrogate and a code above U+10FFFF are a scanner error; an anchor `&😀`, an alias without an anchor and a tag with an unknown handle are parse errors | by `LoadError` (`errors.go`): a surrogate is `yaml.characters`, an anchor and an alias are `yaml.no-anchors`, a tag is `yaml.no-tags`; an anchor without a name, a tag outside the grammar of YAML 1.2.2 and a handle that `%TAG` did not give are "not YAML" | "escaped surrogate…", "8XYN…", "alias with no anchor", "tag …" |
  | an anchor or an alias whose name ends with a colon, `&a:` and `&:@*!$"<foo>:` of 2SXE and W5VH, its scanner does not read: it leaves the last colons out of the name and refuses further on | `colonAnchor` (`errors.go`): at such a name the refusal is `yaml.no-anchors` | "2SXE…", "an alias whose name ends in colons…" |
  | a plain scalar it resolves by its own schema; `\/` it silently reads as `/`; a repeated key in the tree it does not reject | the walk of the tree (`walk.go`), as `Walker` in `spec/tools/format-check/yaml.mjs` | all the rules `yaml.*` |
  | under-indented lines of a quoted string and of a flow collection, tabs in indentation (9C9N, QB6E, DK95/01, Y79Y/003), leading empty lines of a block scalar (S98Z), an implicit key longer than 1024 characters, U+FEFF outside quotes | `layout.go`: each is "not YAML", as YAML 1.2.2 rejects it | "9C9N…", "QB6E…", "S98Z…", "an implicit key beyond the limit", "a byte order mark in…" |

  The messages of v4 by which `errors.go` reads are pinned by these tests: a version that says otherwise will fail them.
- **The second implementation of `spec` in the same places.** The reader reads as YAML 1.2.2 and `yaml.mjs` do, CR and CR LF as line breaks, `...` without a document and U+FEFF too. A character outside the printable set inside quotes YAML 1.2.2 accepts, and the format rejects as `yaml.printable`: the reader does as the format says.
- **API** — `yaml.Read(data []byte) (any, error)`. The value is `map[string]any`, `[]any`, `string`, `int64` for a number with neither a fraction nor an exponent, `float64` for the others, `bool` or `nil`: the format tells the integer and the double apart, and `types/formatv0` holds integers as `int64`. A refusal is a `*Refusal` with the breaches `Breach{Rule, Line, Column, Message}` in the order of places: the line and the column from 1, the column in characters of the file as it is written, and a BOM is a character too. `Rule` is empty when the file is not YAML: such a refusal has no ID in the format. A panic of v4 is caught and becomes the refusal "not YAML". On a parse error, the documents read before it are checked the same way, except for a BOM outside quotes: where quotes stand beyond the error is not known.
- **Tests** (`internal/yaml`):
  - `TestRead` — a table: each rule, the places, the types of the value and each edge case of the table above;
  - `TestReadsTheFormat` — every file of `spec.Format` outside `rules/yaml.*` is read without breaches. It also guards each raise of v4;
  - `TestYAMLTestSuite` — the YAML test suite, the data release `data-2022-01-17`, MIT, in `testdata/yaml-test-suite/`, in two files — `valid.json` with the valid cases and `error.json` with the invalid ones (sources 5.5). An invalid one is rejected, and a valid one is read into the value of its `in.json` or is rejected by a rule, not as "not YAML". The list `suiteUnpassed` is empty, and a case put on it that passes fails the test;
  - `FuzzRead` — the corpus: the vectors of `rules/yaml.*`, the YAML test suite and the cases of `TestRead`. On any input — a value or a refusal, without a panic, the value within the domain of the format, and its JSON text, written by a writer of its own, reads back into the same value. `json.Marshal` does not do for this: `1e20` it writes as an integer, and U+007F–U+009F and U+FEFF as they are;
  - the vector test `internal/vectors`: the runner of the kind `rules` reads a case `yaml.*` and passes when there is a refusal and each breach is of the rule of the folder.

— *Why:*
- v4 and the passes: v4 reads all the files of the format, its scanner is libyaml, and each of its divergences from YAML 1.2.2 is bounded and closed by a pass with a test;
- what v4 lets through and YAML 1.2.2 rejects is closed in the reader itself, together with the YAML test suite: a reader that let it through would accept a module that the format rejects.

— *Rejected:*
- a parsing of YAML 1.2.2 of our own — the most exact, but it is thousands of lines, and a risk of its own in code that reads others' input;
- `go.yaml.in/yaml/v3` and goccy/go-yaml — by the probe above;
- leaving what v4 lets through as a debt, to close later — until then the reader would accept files that the format rejects;
- refusing on U+0085, U+2028 and U+2029 instead of the replacement — the reader would reject a file that YAML 1.2.2 and the second implementation of `spec` read;
- `json.Marshal` in the fuzz test — for the reason above;
- parsing the errors of v4 by their sense rather than by their message: its errors have no other sign, and the tests catch a new message.

— *Consequence:* Dependabot `gomod` raises go-yaml, and a pull request with a new v4 passes the corpus test, the YAML test suite and the pinned messages. — `go.mod`, `go.sum`; `internal/yaml/`; `internal/vectors/`; [architecture.md](architecture.md).

## The module check

**R22. The module check is `module.Check` over the folder of a module as an `fs.FS` and `module.CheckFile` over a file of a part. It reads the folder by the format's layout, holds the rules `module.*`, reads each file of a part with the YAML reader and checks it against the schema of its part, and takes the format version from the manifest before anything else: the window is one version, `FormatVersion`. A module of a version outside the window gets this one breach. The report is the breaches with the rule ID or the schema's keyword and the place, in the order of places, not of the listing. The layout, the forms of names and the version the code holds itself, and the tests check them against the format text, the schema and `VERSION` of `spec.Format`.**

- **API** (`module`):
  - `Check(ctx, fsys fs.FS) (Report, error)` — the module whose folder is the root of `fsys`; the solvers of its templates run with `ctx`. An error comes only when the folder or a file cannot be read, or `ctx` is done: that is code 2 of the CLI commands, and a report with breaches is code 1;
  - `CheckFile(part string, data []byte) (Report, error)` — a file of a part by itself, as an example of a part and a vector of a rule of a part laid out as a file are checked: the YAML rules, the schema of the part, the rules within the part, and for the manifest the version first. On a version outside the window it gives this one breach, as `Check` does: otherwise a manifest of a newer version would get breaches of the schema of this one. A part outside the layout is an error;
  - `Report{Breaches, Warnings}`, `Breach{Rule, Keyword, Path, Line, Column, Pointer, Message}`. A warning refuses nothing. `Keyword` is the JSON Schema keyword that failed: such a breach has an empty `Rule`, no line and no column, and its place is `Pointer`; both are empty only for a file that is not YAML, and `String` writes `schema <keyword>` in place of the rule. `Path` is the path from the module's folder with `/`, as the listing gives it; `.` is the folder itself, as `fs.FS` names its root, and a file of `CheckFile` has no path, and `Path` is empty. The line and the column are in the file of the part, as with `yaml.Breach`, and in `LICENSE` and `NOTICE` only LF ends a line, as with the format's plain text; 0 is a breach of the whole file or of a value in it. `Pointer` is the value, a JSON Pointer in URI fragment form, as the format writes a place: `#/module_format_version`, `#/id`. The order is the path, the line, the column, the pointer by tokens — the number of an element as a number and before any name —, the rule, the keyword, the message;
  - `FormatVersion` and `ReadsFormat(v string) bool`: while the format is v0, the library reads one version.
- **The order of the check** — the format's rules of a module's folder, in order, and the version before them:
  1. the listing. The names are as the folder gives them, letter case and all. A file or folder beginning with `.` is not the module's, with everything inside it. A folder without files holds nothing. A link the listing does not follow: it is `module.layout`, as is everything that is neither a file nor a folder — `format-check` of `spec` decides the same way;
  2. the manifest. No file `manifest.yaml` is `module.manifest`. If there is one, it is read first, and a version outside the window, named by a string of the form MAJOR.MINOR, is `manifest.format-version`, and only it stays in the report. Without a manifest, or with a manifest that the YAML rules rejected, the module is checked by the reader's version: `Manifest.yaml` in place of `manifest.yaml` breaks both `module.manifest` and `module.layout`, as the format text itself says;
  3. the layout. A file not at a path of the layout is `module.layout`. A file of a part is read by the YAML reader and checked against the schema of its part, and its breaches go into the report with the file's path. In a file at a path with `<id>` that both the YAML rules and the schema accepted, a string `id` not equal to the name before the extension is `module.file-ids`;
  4. `media/` — `module.media`: a path of the form `media_path` from `media/` and the signature of the format by the extension, from the first 14 bytes of the file;
  5. `LICENSE` and `NOTICE` — `module.license` and `module.plain-text`. Each of the three faults of plain text — not UTF-8, a BOM and CR — gives one breach, at its first place, and for CR the message also says how many there are.
- **What is decided where the format text is silent:**
  - a manifest that the YAML rules rejected names no version: such a file has no value. The module is checked by the reader's version, as without a manifest, and the report holds the YAML breaches of the manifest and everything else. A manifest of a newer version with such an error gets breaches by the reader's version, and `manifest.format-version` only when the author fixes the YAML: the test "a manifest of 0.5 that the YAML rules refuse…" holds this;
  - a version that is not a string of the form MAJOR.MINOR, a manifest without one and a manifest that is not an object are not `manifest.format-version`, but a breach of the manifest's schema by the reader's version. So says the example `examples/manifest/invalid/format-version-as-number.yaml` — `type at #/module_format_version` —, and so does `format-check` of `spec`: "A version that is no MAJOR.MINOR string is the schema's to refuse";
  - `module.file-ids` is only a string `id` not equal to the name, in a file that the schema accepted: so `id` has the form `local_id` too. No `id`, or one that is not a string, is the schema's: a rule says what the schema cannot say. A file that the schema rejected breaks it and nothing more;
  - the header of an MP3 frame is no shorter than four bytes, as the format says and as WHATWG MIME Sniffing and `format-check` read it (sources 5.6).
- **The code against the format text.** The layout is the table `layout` in `layout.go`; the forms `local_id` and `media_path` are a pattern, a length and `not`; the version is `FormatVersion`. The forms stay Go data beside the schema validator: the layout reads the names of files, not values, and before the schemas. The tests of `format_test.go` check the table against the format's table of the layout — `README.md` from `spec.Format` —, the rules `module.*` against its table of rules, the forms against the `$defs` of `common.schema.json`, where nothing but `type`, `pattern`, `maxLength`, `not` and `description` should be, the extensions of `media_path` against the signatures, and `FormatVersion` against `VERSION`. A move to a tag that changes any of this turns red until the code catches up.
- **Tests** (`module`):
  - `TestCheck` and `TestCheckFile` — tables of modules on `fstest.MapFS` and of files: each rule, names beginning with `.`, empty folders, links and pipes, letter case, paths with `<id>` and `media_path` at the bounds of length and of the names of Windows devices, signatures, versions in the window and outside it; `TestMessages`, `TestBreachString` and `ExampleCheck` — the text of the report;
  - `TestOrderOfListing` — the same report when the folder is listed in reverse order; `TestReadErrors` — an error, not a report, at each read; `TestSymbolicLinkOnDisk` — links to a file and to a folder on disk, through `os.DirFS`;
  - `TestChecksTheFormat` — every folder of `spec.Format` laid out as a module — the examples of a whole module, the folder cases of `rules/`, the modules of the vectors of packages, of the index and of versions —, and every file of a part — the examples, the file cases of `rules/` and `part.yaml` of the open view —, each in a parallel subtest named by its path, gives not one breach but of the rule, or the keyword, of its case;
  - `FuzzCheck` — any file at any path of a module and any file of each part: a report without an error and without a panic, in form and in order;
  - the vector test: a folder case is passed by `Check`, and a file case of a rule of a part by `CheckFile` of the part from the rule ID.

— *Why:*
- two functions: `check` of the CLI and `module-registry` check a module whole, and an example of a part and a vector of a rule in a file are a file by itself;
- the error apart from the report: the exit code of the CLI tells a refusal by the rules from a failure of input and output;
- the window as a constant with a test: a move of `spec` to a new format version turns red on the test and the vectors, rather than silently reading a module of the new version with the code of the old one;
- one breach for a module of another version: "A module is checked, built and run by the version it names", the format says, and its layout and rules the reader does not know;
- the layout and the forms as Go data with a test: `format-check` of `spec` reads the table of the layout by parsing markdown, and the forms with the schema validator; parsing a normative text at each check is one more dependency;
- the pointer: the format writes the place of a value as a JSON Pointer, and a breach of the schema has no other place.

— *Rejected:*
- the window from `VERSION` of `spec.Format` at run time — a move to a tag with a new format version would silently shift the window of the code of the old one;
- checking the rest of the module even with a version outside the window — by the layout and the rules of one version, a module of another would get breaches that are not its own;
- stopping at a manifest that the YAML rules rejected: the author of a module would see one breach per round, and without a manifest the check goes by the reader's version anyway;
- searching for the version in a manifest that the YAML rules rejected: the value of such a file is not defined — which of two equal keys, which number outside ±(2^53−1);
- `manifest.format-version` also for a version that is not a MAJOR.MINOR string — two breaches of one error, and the format's example names the schema here;
- reading the layout and the forms from `README.md` and the schema at run time — above;
- the report as an error, like `*yaml.Refusal`: warnings go into the report of `check` too — about a picture with transparent pixels —, and an empty report is not an error;
- one breach `module.plain-text` per file, as with `format-check`: the BOM and the CRLF of a file from Notepad the author fixes in one go only if they see both.

— *Consequence:* `module/`; `internal/vectors/`; [architecture.md](architecture.md).

## Schemas

**R23. A file of a part is checked by JSON Schema 2020-12 with the library `github.com/santhosh-tekuri/jsonschema/v6` — by the schemas of the version that the library reads: this is a copy of `schemas/` from `spec.Format`, embedded in `internal/schema` and checked byte for byte by a test. A breach of a schema names the keyword and the place of the value, as an example of the format names it, and a file that the schema rejected breaks it and nothing more: the rules of its part do not read it, and the rules between parts run only when all the files of parts have passed the schema. The examples `examples/` are passed, and the files of vectors the test reads by the schema and a type of `formatv0`.**

- **The validator** (sources 4.5):
  - `github.com/santhosh-tekuri/jsonschema/v6`, Apache-2.0, passes the tests of 2020-12 at Bowtie in full. All the errors it gives as a tree with the place in the value, patterns it reads with Go's `regexp`, and numbers it takes of any Go type, through `big.Rat`, so the value of the YAML reader goes as it is. Annotations it does not collect;
  - it requires `golang.org/x/text` (BSD-3-Clause) — the messages. The version from the library's `go.mod` govulncheck marks with two findings, though the code does not reach them, and `go.mod` raises x/text past it. `dlclark/regexp2` is only the library's tests and does not enter the build;
  - kaptinlin/jsonschema is MIT, but Bowtie does not check it, and it has eight direct dependencies. google/jsonschema-go stops at the first error and gives no place. Corvus Go collects annotations, but reads patterns by ECMA-262 rather than as Go's `regexp`, takes input only as JSON text, and is new. xeipuuv/gojsonschema and qri-io/jsonschema are not 2020-12.
- **Schemas** — `internal/schema/schemas/*.schema.json`, `//go:embed`: a copy of `schemas/` from `spec.Format` at the tag of `go.mod`, taken from the Go module cache, not from the submodule. `TestSchemasAreSpecs` checks the set of files and the bytes against `spec.Format`, and the `$id` of each against `VERSION`. The validator gets each schema by its `$id`, references between schemas resolve by it, and the default loader reads only `file:` addresses, so it does not go to the network.
- **API** (`internal/schema`) — `Validate(location string, value any) []Error`. `location` is the file name of a schema, `task.schema.json`, or that name with a pointer to a subschema in the fragment, `task.schema.json#/$defs/check`: so the walk of the open view asks whether a subschema passes. `Error{Keyword, Location []string, Message}`, and `Pointer` writes the place. A schema is compiled once and kept by its place, and `Validate` is safe from several goroutines: the compiled schema it reads without a lock, and the lock is held only by the compiler. An unknown place, a schema that does not compile, `false` anywhere but in `unevaluatedProperties`, and a value not from JSON are a defect of `module-kit`, and `Validate` panics, as `regexp.MustCompile` does: the error of `module.Check` stays only with reading. `Keywords` is the list of keywords that the format allows; a test holds it.
- **Which keywords are named** — as the format names them, from the validator's tree:
  - `$ref`, `allOf`, `properties`, `items` and `if`/`then`/`else` are never named: the errors of `properties`, `items`, `then` and `else` the validator puts right at their place, and those of `$ref` and `allOf` under a node of their own, through which the collection descends, as it does through a group of several errors of one place and through the root;
  - every other keyword is named where it failed, and what failed inside it is not: the branches of `anyOf` and `oneOf`, the subschema of `not`, the elements that `contains` tried, and the names that `propertyNames` read. So the format's example names `anyOf` and `oneOf`, "when none of their branches passes";
  - `unevaluatedProperties` the format holds only as `false`, and `false` stands nowhere else. The validator puts a false schema on the field that it rejects, and the keyword is named at the object, as the example names it. It is recognized by the end of the place of the false schema, `/unevaluatedProperties`: a name of a field or of a definition of the format does not end so, because it is in snake_case;
  - a subschema that failed gives no annotations, by JSON Schema 2020-12, and the fields that its `$ref` declares become "unevaluated" beside it. The report holds these `unevaluatedProperties` too, as ajv does in `format-check`: the example allows it — "reports one with this keyword at this place".
- **The workaround of a bug of the validator.** v6.0.3 gives an error of `propertyNames` not a copy of the place but a live slice of the validator, and the check of the next neighbour overwrites it: the error names another place, and which one depends on the order of the walk of a Go map (santhosh-tekuri/jsonschema#276; sources 4.5). A vocabulary of its own, `propertyNames` (`names.go`), checks the names by the same subschema at the same place and adds an error through `ValidatorContext.AddError`, which copies the place. The error of the validator itself the collection passes over: both fail on the same names, and the value passes or not, as before. A fixed release is a reason to remove the vocabulary.
- **What v6.0.3 does not touch.** The panic on a number that `big.Rat` will not read — `fmt.Sprintf("%v", v)` with a huge exponent — is fixed after v6.0.3 (sources 4.5). Only a `json.Number` of such a text gives it, and the YAML reader gives `int64` and a finite `float64`, which `big.Rat` reads; `FuzzValidate` holds this.
- **The place** — a JSON Pointer in a URI fragment (RFC 6901): `#`, then `/` and a token, where `~` is written `~0` and `/` is written `~1`, and percent-encoding of what a fragment does not hold: `#/modules/example~1smallest/0.1.0/withdrawn`. The header of an example and the place of a report are compared by the pointer that they encode: `format-check` decodes the header too, and a fragment can write one place in different ways. The report of `check` writes a breach of a schema at its pointer, without a line and a column.
- **The module check** (`module`). A file of a part is checked against the schema of its part after the YAML rules, and the manifest after the version. A breach of a schema is a `Breach` with `Keyword` and `Pointer`. A file that both the YAML rules and the schema accepted the check holds by its path, and "all passed" it counts by the listing and these files, without storing it: `module.file-ids` and the rules within a part read only such a file, and a rule between parts runs only when all passed. So does `format-check` of `spec`: "so that a file the schema refuses breaks that and nothing more" (sources 6.2).
- **The vector test** (`internal/vectors`):
  - the runners `examples/valid`, `examples/invalid` and `examples/module`. A file of a part of the layout is checked by `module.CheckFile`, and `inventory` and `module-index` — they are no parts of a module — by the YAML reader and the root schema. A valid example gives not one breach. An invalid one breaks only the schema, and one of the breaches is the keyword at the place that its first line names. An example of a whole module `module.Check` does not reject;
  - `readVector` reads the file of a vector as the format says: the YAML rules, the schema of the vector, then strictly, without unknown fields, a type of `formatv0`. The runners of the vectors read with it, `part.yaml` of the open view by its schema and the type of the part. The packages of a previous format version in `versions/` are read only by the places, as the format reads them, and `open-view.yaml` by the type of `open`.
- **Tests:** in `internal/schema` — `TestSchemasAreSpecs`, `TestKeywordsAreTheFormats`, `TestEverySchemaCompiles`, `TestFaults`, `TestCollect` — on small schemas of its own, each collection and `propertyNames` before a neighbour, as in #276 —, `TestPointer`, `TestOrderOfPointers`, `TestConcurrentValidate` under `-race` and `FuzzValidate`: any text through the YAML reader against each schema, the keyword from the format's list, the place within the value, the corpus the examples. In `module` — fixtures from the examples of a whole module of `spec.Format`, valid by the schemas and by the rules between parts, `TestPartsThatPassed`, `TestChecksTheFormat` by the pair rule — keyword, and `FuzzCheck` with the examples of parts in the corpus.

— *Why:*
- the library: a validator of the subset that the format allows is about a thousand lines of our own code, with `unevaluatedProperties`, and the JSON-Schema-Test-Suite in `testdata`. santhosh passes 2020-12 at Bowtie in full, and a probe on all the invalid examples of the format it passed with one workaround. The open view needs no annotations: `format-check` builds it by its own walk of the schemas and asks ajv only whether a subschema passes, and `module-kit` does the same;
- the copy of the schemas: the code holds the schemas of the version that it executes, as the module check holds the window of versions as a constant. A move to a tag with new schemas turns red on a test, rather than silently starting to check by them. And `spec.Format` at run time would carry into every binary with `module` — `api`, the eval runner of `tutor`, six binaries of the CLI — all of `format/` and `admission/`, megabytes that grow with each vector (measurements 1.3);
- a file that the schema rejected: `format-check` does so, and the reports of the two implementations on one module agree, and `formatv0` says to check a value against the schema before reading it with a type. A rule between parts on a file that the schema did not accept would give false breaches: without a trap catalog, each trap of a task is "unknown".

— *Rejected:*
- a validator of the subset of our own — above;
- kaptinlin/jsonschema, google/jsonschema-go, Corvus Go, xeipuuv/gojsonschema and qri-io/jsonschema — by the facts above;
- `spec.Format` at run time — above;
- `module.*` on any file that YAML read, and the rules of parts on those that passed the schema: the report would diverge from `format-check`, and the boundary would fall between the rules `module.*` and the rules of parts;
- a rule between parts when the files that it reads itself have passed the schema: the author would see more per round, but the report would diverge from `format-check`, and each rule would hold a list of its inputs and a test of it;
- naming also the errors inside the branches of `anyOf` and `oneOf`: no example names them, the format names here `anyOf` and `oneOf` themselves, and at a branch of a common name, a `oneOf` with four `const`, the report would get five lines instead of one;
- cleaning out the "unevaluated" fields beside a subschema that failed: so 2020-12 says, and so ajv gives;
- an error of `Validate` instead of a panic — `module.Check` would have to return it as code 2, though it is a defect of `module-kit`, not of input and output;
- waiting for a release of santhosh with the fix of #276 — its date is unknown, and the vocabulary closes the bug on v6.0.3;
- a textual order of pointers — `#/items/10` would come before `#/items/2`, and a mixed one, where numbers are compared as numbers and names as text, is not transitive: `2 < 10`, `10 < 1a`, `1a < 2`.

— *Consequence:* Dependabot `gomod` raises santhosh and x/text; a pull request with a new santhosh passes the vector test, `TestCollect` and `FuzzValidate`, and with the fix of #276 it removes the vocabulary `propertyNames`. A move to a new tag of `spec` updates the copy of the schemas. — `go.mod`, `go.sum`; `internal/schema/`; `module/`; `internal/vectors/`; [architecture.md](architecture.md).

## The open view

**R24. The open view is built by a walk of the schemas by the format's rules of the open view — `schema.Open` in `internal/schema`: the mark `x-open` counts only in a subschema that applies to the place and passes, and whether a branch of `anyOf`, `oneOf` or `if` passes the walk asks of the validator. The package `openview` gives the open view of each part of the layout — `Manifest`, `SkillGraph`, `Traps`, `Techniques`, `SolvingSteps`, `Lesson` and `Task`: it takes the part as a type of `formatv0`, and gives the view as a type of `formatv0/open`. The vectors `open-view/` are passed.**

- **The walk** (`internal/schema/open.go`) — as `openViewOf` in `tools/format-check/openview.mjs` of `spec`:
  - at a place there apply its subschema and what that subschema brings there: the subschema of its `$ref` beside the other keywords, each branch of `allOf`, the branches of `anyOf` and `oneOf` that pass, `then` when `if` passes, and `else` when it does not, and the branches of `dependentSchemas` whose key the object has. The subschemas of a field are under `properties`, under each key of `patternProperties` that matches it, and under `additionalProperties` where none of them matched in the same subschema; the subschema of an element of an array is under `prefixItems` at its number, or else under `items`;
  - a field is open when at least one of its applied subschemas carries the mark, an element of an open array is open, and the root is always open. An open object without open fields stays an empty object;
  - whether a branch passes is decided by `Validate` of the same subschema by a pointer, compiled once: `task.schema.json#/$defs/answer_slot/allOf/2/if`. The validator collects no annotations, and the walk replaces them, as `format-check` replaces them for ajv. A subschema that the walk reached by two paths — two references to one definition, or references in a circle — is taken once: there is no repeated work, and a circle does not loop the walk;
  - a boolean `if` the walk reads as 2020-12 does: `true` always passes, and `then` applies. `openViewOf` takes `if` only when it is an object, but the format's schemas have no boolean `if`, and the implementations have nothing to diverge on;
  - `patternProperties` and `prefixItems` are not in the format's schemas, and `dependentSchemas` is not even among its keywords. The format's rules of the open view name all three, and the walk holds them, and a test checks them on schemas of its own;
  - `Open(location string, value any) any`, `location` as with `Validate`. The value does not change, and the view is new maps and slices. The format defines the view only of a value that the schema accepts; for another, `Open` gives what the same walk gives. A place or a reference that names nothing is a panic, as with `Validate`.
- **API** (`openview`). The part is a value of a type of `formatv0`, as the consumer holds it after reading the package. The function writes its JSON and reads it into a value of the YAML reader — with `int64` and `float64` —, checks it against the schema of the part and, if the schema does not accept it, gives an error with all the breaches. Then comes the walk, and the view is read strictly into the type of `open`: a field that the type does not know is a panic, a defect, because the type is given by the same marks. Together with the check against the schema, the open view of a task costs a fraction of a millisecond (measurements 1.4). The open view of an instance of a template is `Task` of the instance, and `TaskView` and `TaskDigest` give the view as a JSON value and the digest of its canonical text.
- **The vector test.** The runner `open-view` reads `part.yaml` by the YAML rules, the schema of the part and strictly a type of `formatv0` (`readVector`), and `open-view.yaml` by the YAML rules and strictly the type of `open` (`readOpenView`); both give a JSON value too. There are two comparisons, and the order of keys matters in neither. The view that the walk gives is compared with `open-view.yaml` as a JSON value, as in `format-check`: so a field that is absent differs from a field with a zero. The view that the function of the part gives is compared with the value of the type of `open`: so the library gives the view that the vector holds, and in its type.
- **Tests:** `TestOpen` — on schemas of its own, a case for each rule of the open view, for `true` in place of a subschema, for a circle of references and for one definition by two paths; `TestOpenAtSubschema`, `TestOpenFaults`, `TestTokensOf` and `TestConcurrentOpen` under `-race`; `FuzzOpen` — any text through the YAML reader and each schema of the format: the view is built without a panic and lies within the value. In `openview` — each valid example of each part, an invalid part, a type that does not hold an open field, and `jsonValue`.

— *Why:*
- the walk is in `internal/schema`: it reads the schemas as they are written and asks the validator about subschemas, and both are held by this package. `openview` is a layer of types for consumers;
- types at the input and the output: `api`, the eval runner of `tutor` and the CLI hold the parts as types of `formatv0` and their open views as types of `formatv0/open`, and the type of `open` will not read a closed field;
- the check against the schema at the input: a type of `formatv0` holds neither `required`, nor `pattern`, nor conditions between fields, and the branches of the walk are chosen by the validator. The format does not define the view of an invalid part, and the function does not invent it;
- a panic on a defect — as with `Validate`.

— *Rejected:*
- the open view by a list of fields of its own: the marks in the schemas hold it, and a list would be a second definition of it;
- at the input, a JSON value of the YAML reader: the consumer holds the parts as types, and the value of the reader is the library's internal matter;
- at the output, a JSON value: each consumer would have to read it into a type itself;
- without the check against the schema: a value outside the schema would silently get a view that the format does not define;
- the walk in `openview`, and from `internal/schema` the schemas as JSON: the parsing of references and pointers would be split between two packages;
- caching the answers of the validator for the branches of `anyOf`, `oneOf` and `if`, or checking a branch only by the keywords that choose it, though the whole part has already been checked: the view of a task costs a fraction of a millisecond, which is no price for a second parsing of the schemas, and a cache by value is a key for each place.

— *Consequence:* `internal/schema/`; `openview/`; `internal/vectors/`; README; [architecture.md](architecture.md).

## The rules of parts

**R25. The module check holds the rules of the manifest, of the skill graph and of the catalogs — `manifest.dependencies`, `manifest.bands`, `manifest.rung-templates`, `skill-graph.acyclic`, `skill-graph.known-skills`, `skill-graph.retired-ids`, `skill-graph.own-anchors`, `traps.known-skills`, `task.known-traps` and `task.known-techniques` — and computes the tiers of skills, `module.Tiers`. A rule reads a part as a type of `formatv0` and names the place of the value that breaks it. A rule within one part runs on each file that passed the schema — both in `Check` and in `CheckFile` —, and a rule between parts only when all the files of parts have passed the schema. The vectors of these rules and of `tiers/` are passed.**

- **Within a part and between parts.** Within one part — `manifest.bands` and four rules of the skill graph: their vectors lie as files, and `CheckFile` passes them, as an example of a part. The other five read two parts or more — the manifest and the graph, the manifest and the tasks, a catalog and the graph, a task and a catalog — and run as the rules between parts. A part that the module does not have the rule does not find, so a trap of a task in a module without a catalog is unknown, as in the vector `task.known-traps/no-trap-catalog`.
- **Types.** A value that the schema accepted is read strictly into the type of the part of `formatv0` — `schema.Decode`, shared with `openview` —, once per file, and the rules within a part and between parts take the same values: `formatv0` says to read a value from outside so, and a rule is written by fields, not by maps. A value that the schema accepted and the type does not read is a panic, a defect of `module-kit`.
- **What is decided where the format text is silent:**
  - the place of a breach is the pointer of the value that breaks the rule: a repeated ID of a band is `#/bands/1/id`, a grade of the audience outside the bands or in two of them is `#/audience/grades/1`, a band without grades is `#/bands/1`, a grade of a band that is not in the audience is `#/bands/0/grades/1`; an unknown prerequisite and an anchor that the module itself teaches are `#/skills/<ID>/prerequisites/<number>`, and a skill on a cycle is `#/skills/<ID>/prerequisites`; a retired ID that a skill also bears is `#/retired_skills/<ID>`; an extra anchor of the manifest is `#/dependencies/<number>`, and a missing one is `#/dependencies`, or `#` when there is no field. For `manifest.rung-templates` the place is `#` of the manifest, and the message names the first task by path with an item with a verdict;
  - a cycle is each skill that stands on it: a member of a strongly connected component of two skills or more, or a skill that names itself. The components are found by Tarjan's algorithm, in time linear in the size of the graph. The message names the other skills of the cycle, and a skill that also names itself is told this too, in the same breach. A skill that only rests on a cycle gets no breach, as in the vector `three-skills.yaml`;
  - an item with a verdict is an item at least one answer slot or box of a line of which is checked by a kind with a verdict. So `format-check` reads it too — `hasVerdict` in `tools/format-check/leaks.mjs`;
  - an anchor is a prerequisite of the form `core.<name>`: a local ID has no dot, and an anchor that is not in the shared taxonomy the schema rejects.
- **Tiers** — `Tiers(formatv0.SkillGraph) (formatv0.SkillTiers, error)`, as the format computes them: a skill without prerequisites among the module's skills is tier 1, the others are one tier above their highest prerequisite, and an anchor gives no tier. A graph whose prerequisite names no skill of the module, and a graph with a cycle, have no tiers — this is an error, as `problem` of `tiersOf` of `format-check` is. The package build takes the tiers into the canonical view; the module check does not compute them: there is no rule about them.
- **The vector test.** The runner `tiers` reads a vector by the schema and the type `formatv0.TierVector` and compares the `Tiers` of the graph with `skill_tiers`.
- **Tests** (`module`): `TestPartRules` — each rule on a whole module, at each place of a breach, and that the rules between parts wait until all the files pass, and the rules within a part do not; `TestPartRulesOfAFile` — `CheckFile` runs the rules of its part and not one between parts; `TestPartRuleMessages`, `TestSelfAndCycle`, `TestHasVerdict`, `TestTiers`, `TestCycles` and `TestTypedFault`, and in `internal/schema` `TestDecode`. `TestRulesAreTheFormats` holds each rule in the tables of the format text, and `TestChecksTheFormat` that neither an example nor a vector of another rule breaks these rules. In the corpus of `FuzzCheck` are the vectors of all the rules that the check holds, except `yaml.*`: `FuzzRead` takes them.

— *Why:*
- a rule within a part also in `CheckFile`: an example of a part "meets every format rule that can be checked within its part, and `module-kit`'s CI runs it too", the format says, and the vectors of `manifest.bands` and of the rules of the graph lie as files;
- the rules between parts only when all the files have passed the schema: so the report agrees with `format-check`;
- types: a rule by fields is checked by the compiler, and `formatv0` is given by the same schemas;
- the place is the value, not the file: the author fixes the value, and the report writes a breach of a value at its pointer, as it writes one of the schema.

— *Rejected:*
- rules on a JSON value, by maps `map[string]any`: dozens of conversions whose errors the compiler does not see;
- one breach `skill-graph.acyclic` per cycle: the author who fixes one skill would not see which other skills the cycle runs through;
- tiers in the report of the check: there is no rule about tiers, and the package build computes them.

— *Consequence:* the rules of a task and of a lesson go into the same `withinPart` and `betweenParts`. — `module/`; `internal/vectors/`; README; [architecture.md](architecture.md).

## The rules of a task and of a lesson

**R26. The module check holds the rules of a task and of a lesson that need neither a check kind, nor the leak check, nor templates: `task.known-skills`, `task.skill-roles`, `task.own-anchors`, `task.known-band`, `task.known-pictures`, `task.age-categories`, `task.zones` and `task.assessability`, and for a lesson `lesson.known-skills`, `lesson.known-tasks`, `lesson.task-skills`, `lesson.known-pictures`, `lesson.areas`, `lesson.age-categories` and `lesson.order`. They run as the rules of the manifest and the graph do: within one part — `task.skill-roles`, `task.zones` and `lesson.areas`, on each file that passed and in `CheckFile` —, and the others between parts, when all the files of parts have passed the schema. The vectors of these rules are passed.**

- **Within a part and between parts.** The boundary is the same as for the rules of the manifest and the graph: a rule within a part reads only its own file. `task.known-pictures` and `lesson.known-pictures` read one part, but also the listing of the module — the files of `media/` —, and `CheckFile` has no listing. This is a rule that is checked only on a whole module, and it runs between parts.
- **What is decided where the format text is silent:**
  - the skills of an item are its own, or else the task's. `task.skill-roles` checks the skills of the task only when at least one item takes them: the rule speaks of the skills of an item — "of the task, for an item that names no skills of its own". `task.known-skills` and `task.own-anchors` check each place where the task names skills;
  - `task.assessability`: an item with a rubric requires `rubric` of its primary skills, and the other items `verdict`. A rubric stands only on an answer slot `sheet`, and such an answer slot is the only one in the task, so the kinds do not mix within one item. A skill that is not in the graph the rule skips — this is `task.known-skills` —, and a skill of the task that breaks the rule in several items is named once;
  - zones overlap when their intersection has an area, and touching at an edge is no overlap, as the format says. The breach stands at the zone that comes later by items and answer slots, and names the earlier one;
  - `lesson.age-categories` gives two breaches: a category outside the audience — at the category itself, `#/age_categories/<number>`, and a lesson that, together with its tasks, no category fits — at `#` of the lesson. A task that the module does not have takes no part in this intersection: it breaks `lesson.known-tasks`;
  - `lesson.order`: lesson A comes after lesson B when a skill of A requires, through the prerequisites of the module's skills, a skill of B that is not among the skills of A, as the format orders lessons. A cycle in this order is found by the same search for strongly connected components as for `skill-graph.acyclic`, and the breach stands at each lesson on the cycle, `#/skills`. The order is derived from the skill graph, and a graph with a cycle of prerequisites gives no order: then the rule waits until the graph passes `skill-graph.acyclic`, as a rule between parts waits for a file that the schema rejected — otherwise one cycle of the graph would repeat as a breach at each lesson of its skills;
  - `lesson.known-tasks` and `lesson.task-skills` — at the segment, `#/segments/<number>/task`, and a picture at its own `picture`.
- **Tests** (`module`): `TestTaskRules` — each rule on a whole module, at each place of a breach, and lessons in order; `TestTaskRulesOfAFile` — the rules within a task and a lesson in `CheckFile`, where zones that touch at an edge do not overlap; `TestTaskRuleMessages` and `TestReviewedCases` — the boxes of lines with a verdict without rung templates, and a cycle of the graph under two lessons. Each case of `TestPartRules` touches only its own rule.

— *Why:*
- one boundary for all the rules of parts — the report agrees with `format-check`, and `CheckFile` checks an example of a part by what "can be checked within its part";
- `task.skill-roles` — by the letter of the rule: the skills of the task are skills of an item only when the item names none of its own;
- one breach for a skill of the task in `task.assessability`: the place is one, and the author fixes it once.

— *Rejected:*
- `task.known-pictures` and `lesson.known-pictures` within a part: `CheckFile` has no files of the module, and an example of a part would be rejected for a picture that only a whole module holds;
- checking by `task.skill-roles` the skills of the task when all the items name their own: such skills belong to no item;
- one breach `lesson.order` per cycle — as for `skill-graph.acyclic`: the author who fixes one lesson would not see which others the cycle runs through;
- `lesson.order` on a graph with a cycle too: lessons in which there is no error would get a breach that goes away together with the cycle of the graph.

— *Consequence:* `module/`; `internal/vectors/`; README; [architecture.md](architecture.md).

## The verdict of a key

**R27. The verdict of an answer slot and of an item is given by the package `verdict`: the kinds with a key — `number`, `fraction`, `choice` and `time` — by the format's normalization, the notation of numbers, exact comparison and the trap of a wrong value, and a rubric has no verdict. NFKC is `golang.org/x/text/unicode/norm`, and the version of Unicode the release names by the constant `verdict.UnicodeVersion` and by the second line of `module-kit --version`. The module check holds the rules `task.choice-key`, `task.choice-options`, `task.fraction-key` and `task.trap-values` within a part, by the verdicts of the same check. The vectors `verdicts/` and the vectors of the four rules are passed.**

- **API** (`verdict`):
  - `Verdict` — `correct`, `not_yet`, `unreadable` and `unchecked`, and `Reason` — `blank` and `notation`, and the reasons of reading of an observation: types of its own with the words of the format;
  - `AnswerSlot` — the binding, as the format writes it, the check, and the options of an answer slot bound `choice`. `AnswerSlots(item)` gives the answer slots of an item in order, and `ChoiceOptions` the options of a choice: the answer slot's for the binding `choice`, otherwise the check's own;
  - `OfItem(ctx, item, values)` and `OfSlot(ctx, slot, value)` — the verdict of an item and of an answer slot by the values as they are written, and `nil` is nothing read. `SlotVerdict` holds the reason for `unreadable`, and for `not_yet` a copy of the trap value that matched, with the ID of the trap and the text for the learner. `ItemVerdict` is the verdict of an item and of its answer slots;
  - `Same(slot, a, b)` — whether two values are equal as the check compares them: for a choice the normal forms, an option or not, for a number and a fraction the values in the notation of the kind, for a predicate the integers;
  - an error gives no verdict: `context` is cancelled; an item has no answer slots, or the values are not one per answer slot; a value is not in UTF-8; a check that the format does not accept — a kind not among the format's, no key, a key not in the notation of the kind, a choice without options, and a placeholder of a template in the key or in a trap value of `number` and `fraction` is `ErrPlaceholder`; a rubric is `ErrNoVerdict`. A check that breaks one of the four rules does give a verdict: by it the rules read the check.
- **Normalization** — the format's steps 1–7 and the form of a choice, and at the edge cases as `verdicts.mjs` of `spec` does; the steps are in `internal/normal`, which the leak check shares:
  - at step 1 the neighbours of a vulgar fraction are characters of the original string, so `½½` and `²½` are both replaced; superscript and subscript characters are the whole block U+2070–U+209F;
  - steps 3 and 4 come after NFKC and catch what it gives: U+2011 becomes U+2010, then `-`;
  - only spaces are trimmed, and between steps 5 and 6 there is no trimming: `= rungs` and `7 rungs .` with the unit `rungs` are not in notation, and `7 .` is, by the letter of steps 5 and 6, as `verdicts.mjs` and the vector `verdicts/number/units.yaml` read them;
  - a unit is compared by the lowercase letter of each character, and its length is counted in characters: `İ` is two bytes, `i` one. The lowercase letter is `unicode.ToLower`, the simple mapping of UnicodeData, field 13 (sources 5.7);
  - the key is read as written, and the read value and the trap value after normalization, for a number with the units.
- **Numbers** — the notation by the regular expressions of the format text, no longer than 16 characters. The value is a `big.Rat`, and irreducibility is the GCD of the numerator and the denominator as written: `big.Rat` reduces them.
- **Rules** (`module`, `checks.go`) — within a part, on each task file that passed and in `CheckFile`: they need only the file itself.
  - `task.choice-key` — a key not among the options as written;
  - `task.choice-options` — an option that is `Same` with an earlier one;
  - `task.fraction-key` — a key to which its own check with `form: reduced` does not give `correct`: the key is in notation, and this is exactly "reducible";
  - `task.trap-values` — a trap value that is not `not_yet`, for a choice not an option as written, or a value that is `Same` with an earlier one.

  A task with `template` is checked by its instances, each read as a task by the same function, `slotRules`. A check to which `OfSlot` answers `ErrPlaceholder` — `{` in the key or in a trap value of `number` or `fraction`, as with `holdsPlaceholder` of `format-check` — in a task that is no template is the rule `task.placeholders`, not these four: so a placeholder is recognized in one place. An error of `OfSlot` on a check that the schema accepted is a defect of `module-kit`, a panic, as with the reading of a part into its type.
- **The version of Unicode.** `verdict.UnicodeVersion` is the version of the NFKC tables with which the library is built, and `TestUnicodeVersion` holds it equal to `norm.Version` and `unicode.Version`, at the version the test names. `module-kit --version` writes `Unicode` and `verdict.UnicodeVersion` as its second line, and the test of `cmd` expects it.
- **Where it diverges from the second implementation of `spec`** — on a run of more than 30 non-starters: `norm.NFKC` goes together with the Stream-Safe Text Process and inserts U+034F, and ICU in `verdicts.mjs` does not (sources 5.7). The format keeps such runs out of its vectors, and verdicts are hardly touched by them: such a value is not in the notation of a number and is not equal to an option of a choice.
- **The vector test.** The runner `verdicts` reads a vector by the schema and a type of `formatv0` and checks the verdict, the reason and the trap of each answer slot, and the verdict of the item.
- **Tests:**
  - `verdict` — `TestOfSlot`: the steps, notation, units, fractions, traps, choice and the predicate. Also `TestOfSlotErrors`, `TestTrapIsTheTrapValue`, `TestAnswerSlots`, `TestOfItem`, `TestItemOf`, `TestSame`, `TestUnicodeVersion` and `FuzzOfSlot`; in `internal/normal` — `TestSteps`, `TestNFKCIsStreamSafe` and `FuzzNotation`;
  - `module` — `TestCheckRules`, `TestCheckRulesOfAModule`, `TestCheckRuleMessages` and `TestVerdictOfFault`;
  - `cmd/module-kit` — the output of `--version`: its second line the test checks against `verdict.UnicodeVersion`.

— *Why:*
- `x/text/unicode/norm` is a direct dependency already, for the messages of the schema validator, BSD-3-Clause, and with `go1.27` its tables are of the version of Unicode of `unicode` of Go (sources 5.7). NFKC, White_Space, the category N and the lowercase letter are of one version. NFKC tables of our own are thousands of lines of code and data;
- the format has the reader's release name its version of Unicode, and the description of the release the shared `release.yml` assembles itself. The constant is for `api` and the eval runner of `tutor`, which keep the version of `module-kit` with an attempt and with a report, and `--version` is for the module author and CI of `modules`;
- types of its own: `formatv0.SlotReadingReason` has no reasons of an observation, which the join of two readings needs, and a trap by its ID alone would not give `api` a text after a wrong value. `context` and `error` from the start: a cancellation is an error without a verdict, and the predicate's sandbox needed no change of signatures;
- the rules within a part: they read only the task itself, as `task.zones` does.

— *Rejected:*
- NFKC tables of our own from the UCD;
- the version of Unicode only as a constant, or only in the documents;
- the types of `formatv0` in the API — `SlotReading` and `SlotReadingReason`; an API without `context` and `error`;
- irreducibility by `big.Rat`: `SetString` stores `6/8` as 3/4;
- the rules on the template itself: the valid `examples/task/valid/fraction-template.yaml` with the trap `"{whole}/{part}"` would give a false `task.trap-values`.

— *What is not closed:* `unicode` and `norm` will diverge on a Go with a new version of Unicode, if `norm` has not taken it yet — `//go:build go1.27` is true on Go 1.28 too —, and at a consumer that raises x/text by MVS. `TestUnicodeVersion` sees this only on the toolchain of `module-kit`, and the NovaTrail repositories take Go from the container of `spec`.

— *Consequence:* `verdict/`; `internal/normal/`; `module/checks.go`; `cmd/module-kit/main.go`; `internal/vectors/`; README; [architecture.md](architecture.md).

## The check predicate

**R28. The Starlark of the format is executed by the package `internal/sandbox` on go.starlark.net, pinned in `go.mod` and excluded from Dependabot `gomod`: the dialect of the format, its 33 built-in names by a list of its own, the limit of steps "the count reaches", the clock, the cancellation of a `context` and `recover`. The verdict of a predicate is given by `verdict` in the sandbox: `True` is `correct`, `False` is `not_yet`; an error, a result that is not a `bool`, the limit of steps, the clock and a panic are `unchecked`; a cancellation of the `context` is an error without a verdict. The module check holds `task.predicate-expression` within a part. The vectors `verdicts/predicate/` and `rules/task.predicate-expression/` are passed.**

- **The sandbox** (`internal/sandbox`):
  - `CheckExpr(src, names...)` — an expression of the dialect that names only `names`, the names that it binds itself and the built-in names of the dialect. The parsing is `syntax.FileOptions.ParseExpr` with the options of the dialect, the names are `resolve.ExprOptions` with its own list of 33 names, not with `starlark.Universe`. An error is the line and the column where go.starlark.net finds them, and its words: `1:8: got '=' after expression, want EOF`, `1:1: undefined: number`;
  - `Eval(ctx, limits, src, env)` — an expression on its own thread: first `CheckExpr` on the names of `env`, then the parsing again and `starlark.EvalExprOptions`. The parsing is done twice because the resolver writes into the expression, and `EvalExprOptions` resolves it once more (sources 4.6); ≤ 1000 characters of a predicate are parsed in microseconds;
  - `Limits{Steps, Clock}`, zero being no limit; `Predicate` and `Solver` — the format's steps and the clock of `module-kit` for each;
  - the thread — `SetMaxExecutionSteps` and `OnMaxSteps`, the clock — `time.AfterFunc`, the cancellation — `context.AfterFunc`. Each calls `Thread.Cancel` with its own reason, and the error names the first: `ErrSteps`, `ErrClock` or the error of the `context`, wrapped. A panic of the interpreter or of a built-in function — `recover` and `ErrPanic`. `print` writes nothing: without a `Thread.Print` of its own, go.starlark.net writes to stderr. There is no `Thread.Load`: `load` cannot stand in an expression, and the format forbids it to a solver;
  - a `context` done before the evaluation is an error at once, without a thread. An evaluation that ended before the interpreter saw the cancellation returns its result or its error, even if the clock or the `context` fired while it was ending: the reason is taken only by an error in which the interpreter named it a cancellation, "Starlark computation cancelled: <reason>" (`interp.go`, `checkStepsSlow`). So `12 % value == 0` on `0` stays `unchecked`, and not an error of a `context` that was cancelled at the same moment.
- **The verdict of a predicate** (`verdict`):
  - `OfSlot` and `OfItem` pass the `context` to the evaluation. Empty and not in notation — before the sandbox, as for the other kinds. An integer in notation goes into `value` as a number, `starlark.MakeBigInt`, exactly: `value == 9007199254740993` on `9007199254740993` is `correct`;
  - `True` is `correct`, `False` is `not_yet`, and an error, a result that is not a `bool` — even a truthy one, `[value]` —, the limit, the clock and a panic are `unchecked`, without a reason: the `unchecked` of a predicate has none, neither in the format nor in `SlotVerdict`. A cancellation of the `context` is an error: `errors.Is` with `context.Canceled` or `context.DeadlineExceeded`;
  - a predicate that breaks `task.predicate-expression` gives `unchecked`: the format says "an error … give `unchecked`" of an expression, and a check that breaks a rule that it shows by itself does give a verdict;
  - `CheckPredicate(expression)` — the rule `task.predicate-expression` in one place: only `verdict` knows the name `value`.
- **The rule** (`module`) — in `slotRules`, within a part, on each file that passed and in `CheckFile`, at `#/items/<number>/answer_slots/<number>/check/predicate`. In a template the predicate is a string of the task, where `{name}` is a placeholder and `{{` is `{`, and the rule reads each instance, by the same `slotRules`.
- **Dependencies.** go.starlark.net — BSD-3-Clause, direct. It takes `golang.org/x/sys/unix`, BSD-3-Clause: the version that its `go.mod` requires govulncheck marks with a finding on `windows` that the code does not reach, and `go.mod` raises `x/sys` past it, so govulncheck writes "No vulnerabilities found" (sources 4.7). `ci-licenses` sees go.starlark.net and `x/sys/unix` as BSD-3-Clause.
- **The address reserve.** The package `go.starlark.net/starlark` on 64-bit linux and darwin reserves 4 GiB of address space at the start of the process — `mmap` with `PROT_READ` — for small integers, and if that fails, it writes "Starlark failed to allocate 4GB address space" to `log` and computes integers more slowly (sources 4.6). Every process that imports `verdict` inherits this — `api` and the eval runner of `tutor`: the reserve takes no memory, but a limit of the process's address space, `ulimit -v`, sees it.
- **Dependabot:** `go.starlark.net` is in `ignore`, beside the modules of `spec`: its version is part of the format.
- **The vector test.** The runner `verdicts` takes the predicate too.
- **Tests:**
  - `internal/sandbox` — `TestDialect`: the options, the 33 names — the keys of `starlark.Universe` and the list of the format text, in order; `TestVersionIsTheFormats`: the version in `go.mod` is the one that the format text names; `TestCheckExpr`, `TestEval` — also that a name that the process added to `starlark.Universe` is not a name of the dialect; `TestSteps`: a stop exactly at the limit; `TestClockAndCancel`, `TestFirstCauseWins`, `TestPanic`, `TestPrintWritesNothing`, `TestConcurrentEval` and `FuzzCheckExpr`;
  - `verdict` — the cases of a predicate in `TestOfSlot`, `TestPredicateStops` — the limit, the clock and the `context`, through `OfItem` too —, and `TestCheckPredicate`; `FuzzOfSlot` runs predicates too, and `unchecked`;
  - `module` — the cases of the rule in `TestCheckRules` and its messages in `TestCheckRuleMessages`.
- **The measurement** (measurements 1.1): a predicate of 999,990 steps gives its result, one of 1,000,000 stops at the limit, both in 12–15 ms; a predicate that calls `sorted(range(1000))` on each turn stops at the limit at its 50,000th turn, after 932 ms.

— *Why:*
- one sandbox for the predicate, the template expressions and the solver: `Limits` and the thread are what they share;
- a list of names of its own in the resolver: the dialect is the 33 names that the format lists, while `starlark.Universe` is a variable of which go.starlark.net says "The Go application may add or remove items from the universe dictionary" (sources 4.6). A name that the process of `api` added does not become a name of a predicate, and the rule and the verdict read the same;
- `EvalExprOptions` on an environment with `value`: so the vector counts the steps — 999,990 for `steps-below-the-limit` —, and a wrapper `lambda value: …` would add the steps of a call and take the vector past the limit;
- the first reason — as the thread itself has it: `Cancel` keeps it, "preserving earlier reason if any" (sources 4.6);
- `unchecked` without a reason — the format and the contract do not name one, and the same release repeats the verdict on the stored value;
- `x/sys` past the version go.starlark.net requires, as x/text past santhosh's: it is simpler to close a finding of govulncheck by a raise than to keep it open.

— *Rejected:*
- `starlark.Universe.Has` instead of a list of its own: a process that adds a name to `Universe` would widen the dialect, the rule and the verdict;
- the predicate as `lambda value: …` or `ExprFunc` — one parsing per answer slot: other steps, and the vectors at the limit diverge;
- an error without a verdict on a predicate that breaks `task.predicate-expression`: the format gives an error of an expression `unchecked`, and a check that breaks a rule that it shows by itself gives a verdict;
- a reason of `unchecked` as a field of `SlotVerdict`: neither the format nor the contract of the observation names it, and the eval report names the cause of a predicate's `unchecked` on a valid observation by its own field, `UNCHECKED_CAUSE_CHECK`;
- a fuzz test of the evaluation: a built-in function takes a gigabyte per step (Д10), and the worker of the fuzz test would fall from memory. `FuzzCheckExpr` checks only the parsing and the names, and the evaluation is run by `FuzzOfSlot` on the predicates of its own cases;
- `x/sys` at the version go.starlark.net requires: govulncheck marks it.

— *What is not closed:* memory — Д10.

— *Consequence:* `internal/sandbox/`; `verdict/`; `module/checks.go`; `internal/vectors/`; `.github/dependabot.yml`; `go.mod`; README; [architecture.md](architecture.md).

## The observation

**R29. The validity of an observation `novatrail.observation.v1`, the join of two readings and the verdict by an observation are held by the package `observation` over `verdict`, with four functions — one per rule of the contract of the observation and per field of its vector: `Validate`, `Join`, `Verdicts` and `Essay`. An invalid observation is the error `ErrNotValid`, and that means there is no observation: each answer slot that AI reads is `unchecked`. So an `unchecked` in the verdict by a valid observation is always the `unchecked` of a predicate. `verdict` stays without protobuf and gives the join its normal form of a value. The vectors `observation/v1/` of `contracts.Vectors` are passed.**

- **The package `observation`** imports `verdict`, `formatv0` and `observationv1`, the generated contract package of the module `spec/contracts`. The protobuf runtime comes only with it: `verdict`, `module` and the CLI do not take it.
- **`verdict` gives the join** four reasons of reading — `Illegible`, `Ambiguous`, `NoFinalAnswer` and `SeveralFinalAnswers`, beside `Blank` and `Notation` —, `NormalForm(slot, value)` and `ItemOf(slots)`:
  - `NormalForm` — the value as the kind normalizes it, without the key and without Starlark, and whether it is in notation. For a choice — the normal form, in notation when it is the form of the key or of an option; for the other kinds — steps 1–7, with the units of a number, in notation when the kind takes what is left. An empty form is empty. A rubric normalizes nothing;
  - `ItemOf` — the verdict of an item from the verdicts of its answer slots, and without answer slots `unchecked`: nothing is checked;
  - a value not in UTF-8 `NormalForm` gives as it is, not in notation, as `Same` equates it to nothing.
- **`Validate(items, o)`** — the conditions of the contract: each answer slot that AI reads — one not bound `choice` — once, in the order of the task, and no others; a value that is not empty, not only White_Space and in UTF-8; a reason that is not `UNREADABLE_UNSPECIFIED` and is in the table of `module-kit`; `UNREADABLE_NO_FINAL_ANSWER`, `UNREADABLE_SEVERAL_FINAL_ANSWERS` and `crop_box` only for `anywhere`, and `crop_box` only beside a value, from 0 to 1000, `ymin < ymax` and `xmin < xmax`; for `sheet` — a value, `UNREADABLE_BLANK` or `UNREADABLE_ILLEGIBLE`. The error names each condition broken, as the `breaches` of the second implementation do.
- **The table of reasons** — an explicit mapping `Unreadable` → `verdict.Reason`. A reason that the package adds later — even at a consumer that raises `spec/contracts` higher by MVS — is not valid to this release, as the contract requires: "A reason added to the package after a release of `module-kit` is not valid to that release". `TestReasonsAreThePackages` holds the table equal to the enumeration of the pinned tag.
- **`Join(items, first, second)`** — without `context` and without Starlark:
  - an essay is an error, it has one reader; a check that the verdict rejects is an error, whatever was read; if either is invalid — `ErrNotValid`, and the error names which;
  - a value with an empty normal form is `UNREADABLE_BLANK` before the pairing;
  - two values join when both are not in notation with one normal form, or both are in notation, equal by `verdict.Same`, and their verdicts by `verdict.OfSlot` are the same — the verdict, the reason and the trap's ID. For a predicate the verdicts are not compared: one integer gives one verdict;
  - when they join — the value of the first, the `crop_box` of the first or the second; two same reasons — that reason; the rest — `UNREADABLE_AMBIGUOUS`, without `crop_box`;
  - the joined observation is a new message: the places and the `crop_box` are copied, unknown fields of the input do not pass over.
- **`Verdicts(ctx, items, o, tapped)`** — the verdict of each item in the order of the task. `o` is what `Join` gave, or what `Validate` accepts: the reading of one reader, the truth of the markup. `nil` means there is no observation: each answer slot of AI is `unchecked`, and the answer slots `choice` get a verdict. A value — `verdict.OfSlot`, a predicate — in the sandbox with `ctx`; a reason — `unreadable` with the reason by the table; a tap — `OfSlot` on the text of the option, and without a tap — `blank`; an item — `verdict.ItemOf`. The items of an instance of a template are those of the instance, with its numbers: a task that holds a placeholder is `verdict.ErrPlaceholder`. The errors are a `ctx` that is done, a task with a rubric (`verdict.ErrNoVerdict`), an item without answer slots, an invalid observation, a tap not at its own place and a check that `OfSlot` rejects.
- **One reader.** In the reading mode `single` one reader reads a task that is not an essay: `api` calls `Verdicts` on that observation without `Join`, or on `nil` when `Validate` gives `ErrNotValid` or AI gave none, and keeps the observation only when it is valid, as it keeps the joined one. The API is the same.
- **The checks of the answer slots come before the reading.** `Join` and `Verdicts` first call `verdict.OfSlot` without a value on each answer slot that AI reads: so the check alone is read — the kind, the key, the options, the placeholder and the predicate —, without Starlark. A check that the verdict rejects is an error, whatever was read and whether there is an observation at all. Inside the join an error of `OfSlot` is therefore a defect of `module-kit` and a panic.
- **`Tapped{Place, Value}`** — as `api.v1.TappedChoice` (`attempt.proto`): only the taps, in the order of the task. A place not bound `choice`, one out of order or twice, or a tap without a place is an error.
- **`Essay(items, o)`** — `Submission`: `SubmissionAssessment` — a value, the text goes to the assessment by the rubric; `SubmissionUnreadable` — `UNREADABLE_BLANK` or `UNREADABLE_ILLEGIBLE`; `SubmissionUnchecked` — `nil`. An invalid one is `ErrNotValid`. A task that is not an essay is an error: an essay is a task whose only answer slot is bound `sheet`.
- **The `unchecked` of a predicate.** Invalidity is named by `Validate`, `Join` and `Essay` with the error `ErrNotValid`: the eval runner puts `UNCHECKED_CAUSE_INVALID` by it, knows `FAILURE` and `LIMIT` itself, and an `unchecked` in `Verdicts` on a valid observation is the `unchecked` of a predicate, the cause `UNCHECKED_CAUSE_CHECK`.
- **The dependency.** `google.golang.org/protobuf`, BSD-3-Clause, is direct, at the version that `spec/contracts` requires: the tests take `protojson` and `proto`, and `observation` the generated package. connect-go is not imported and does not go into the build. Dependabot `gomod` raises protobuf: it is not part of the format. `go list -deps ./module ./cmd/module-kit` is without protobuf.
- **The vector test.** The runner `observations` is in `internal/vectors/observation_test.go`. The file is strictly JSON, `DisallowUnknownFields`; the canvas and the items — by the schema of a task, `task.schema.json#/properties/canvas` and `#/$defs/item`, as in `observation.mjs`, and strictly by the types of `formatv0`; the observations — by `protojson.Unmarshal`, which refuses an unknown field and reads a number outside the enumeration, `1000` (sources 4.8). `valid` — by `Validate`, `joined` — `proto.Equal` with `Join`, `verdicts` — by `Verdicts`, `submission` — by `Essay`, where `ErrNotValid` is `unchecked`. `TestObservationsRefuse` holds that the runner refuses on a corruption of each field.
- **Tests:**
  - `observation` — `TestReasonsAreThePackages`, `TestValidate` — a case per condition, `TestJoin`, `TestJoinSharesNothing`, `TestJoinErrors`, `TestVerdicts`, `TestVerdictsErrors`, `TestEssay` and `TestNotUTF8IsNoMessage`;
  - fuzz tests, the observation being foreign input: `FuzzJoin` — two readings of one answer slot of each kind and binding: there is a joined observation only when both are valid, the joined one is valid and gets a verdict, in the reverse order — the same "value or reason", and a reading joined with itself is itself, except an empty normal form, which is `UNREADABLE_BLANK`; `FuzzObservation` — any bytes through `proto.Unmarshal` on the tasks of the vectors: nothing panics, and a valid one joins with itself and gets a verdict or the outcome of an essay. The seed corpus is the observations of the vectors;
  - `verdict` — `TestNormalForm`, a value not in UTF-8 too, and `TestItemOf` — on `ItemOf`, without answer slots too.

— *Why:*
- a separate package: the format and the contract are split in `spec` the same way — the root module `spec` requires nothing, protobuf only the nested `contracts`, and the second implementation takes the key of the join from `verdicts.mjs` in `tools/contracts/observation.mjs`. `verdict` without protobuf is needed by the module check and the CLI, which read no observations;
- four functions — a rule per function: each is held by its own field of the vector, the repeat of a verdict in `api` and the truth of the markup at the runner are the same `Verdicts` on one observation, and the `unchecked` of an invalid observation and the `unchecked` of a predicate are told apart by the form of the API, without a field of the reason in `SlotVerdict`;
- `Join` without `context` and without Starlark: verdicts are compared only for the kinds with a key, and a predicate on one integer gives one verdict. Otherwise the clock of `module-kit` would make the join depend on the machine;
- a value not in UTF-8 is invalid, not an error: proto3 requires of a string "A string must always contain UTF-8 encoded or 7-bit ASCII text", and protobuf-go neither writes nor reads such a string (sources 5.8, sources 4.8). An invalid observation is "as a failure of AI", and the learner gets `unchecked`, not a refusal of the server;
- the table of reasons is its own, not the enumeration of the package: so a new reason at a consumer with a new `contracts` too is invalid until `module-kit` names it.

— *Rejected:*
- everything in `verdict`: protobuf would come into `module` and into the CLI's binary;
- one function `Read` on 0–2 observations with the result `{Joined, Invalid, Items, Submission}`: two forms of the result, and the meaning depends on the length of the slice;
- the reason of `unchecked` as a field of `SlotVerdict` — as for the predicate, neither the format nor the contract names it;
- the taps as one value for each answer slot `choice`, `nil` without a tap, as the vector writes them: `api` gets only the taps (`api.v1.TappedChoice`), and the form is its own;
- comparing in `Join` the verdicts of a predicate too: Starlark on each pair and the clock in the join;
- a value not in UTF-8 as an error without a verdict, as for `verdict.OfSlot`: the observation comes from AI, and an invalid one is its outcome, not a defect of the caller.

— *Consequence:* `api` and the eval runner of `tutor` take `observation`, and the types of the verdict from `verdict`. — `observation/`; `verdict/`; `internal/vectors/`; `go.mod`; README; [architecture.md](architecture.md).

## The leak check

**R30. The leak check is held by the package `leak`: `For(task)` builds the search of a task — the keys of its checks `number`, `fraction`, `choice` and `time` that the texts of their item do not show, and its leak phrases —, and `Search.In(text)` gives what of them was found in the text, each a `Leak` with its place in the task. A template is the error `ErrTemplate`: it is searched by its instances. Normalization and the notation of numbers are `internal/normal`, shared by the verdict and the leak. The module check holds `task.no-leaks` — within a part the texts of the task itself, between parts the texts of the catalogs, the solving steps and the rung templates — and `lesson.no-leaks` between parts, by the keys and phrases of each task and of each instance of a template that a lesson names. The vectors `leaks/` and those of these rules are passed.**

- **API** (`leak`), by the [format text](../spec/format/README.md):
  - `For(task) (*Search, error)` — what is searched for: the key of each check `number`, `fraction`, `choice` and `time` of each item, except a key that the same search finds in a text that the learner sees with its item — in the paragraphs of the statement and the `alt` of its pictures, in the `text` of the item, in the `text` pieces of its line and in the options of its answer slots `choice` —, and each leak phrase. A key of a number, a fraction or a time not in the notation of its kind is not searched for, as in `leaks.mjs` of `spec`: a placeholder outside a template breaks `task.placeholders`. A template is `ErrTemplate`;
  - `Search.In(text) []Leak` — what was found, in the order of the task: the keys by items and answer slots, then the phrases. `Leak{Kind, Written, Item, Slot, Phrase}` — a key or a phrase as the task writes them; for a key — the item and the answer slot from 0, the answer slots in the order of `verdict.AnswerSlots`, for a phrase — its place in `leak_phrases`. A key of one magnitude in two answer slots is found twice. The search changes nothing after `For`, and one `Search` searches from several goroutines at once;
  - `Numbers(text)` — the numbers that the search reads in a text, each as it stands in the normal form, in the order of the text: the module check warns of a number in a text of the whole module by it;
  - `api` builds the search on the instance of an attempt — a task with numbers, without `template` — and searches every one of the tutor's words with it: if anything was found, the tutor's words are replaced by the fallback text.
- **The search:**
  - the normal form — step 1, NFKC, steps 3 and 4 and the simple lowercase mapping: `normal.Search`. A character of a word is one of the categories L, M and N of the package `unicode`; U+FFFD is not a character of a word;
  - a phrase — a key of a choice and a leak phrase — is searched for as a substring of the normal form at each place where it stands, and is found where an edge that is itself a character of a word has no character of a word outside it: so `odd` is found in "oddly odd". The substring is searched for by bytes, `strings.Index`: both forms are UTF-8, and a match of bytes begins and ends on a boundary of characters. An empty form is found nowhere;
  - numbers — by the characters of the normal form, from left to right: at a digit 0–9 — the longest run of the form of a number, a fraction, a decimal or an integer; the run is a number when it has no character of a word outside it, and the search goes on after the whole run. The value is exact, a `big.Rat`, leading zeros and all. A key is compared by magnitude, and the text gives no sign;
  - a time — a run of digits, `:` and digits in the notation of a time with no character of a word at its edges; a key of a time is found where a time of the text has the same hours and minutes, `normal.Clock`;
  - the numbers and the times of a text are read once, and only when a key of their kind is searched for: for a task with phrases alone no `big.Rat` is built;
  - a byte not in UTF-8 in a text is U+FFFD: step 1 reads the text by the characters of Go. The rest of the text around such a byte is searched as usual.
- **Normalization** — `internal/normal`: step 1 — `Scripts`, steps 2–4 — `Text`, `Lower`, the form of a choice — `Choice`, steps 1–7 — `Number`, the form of a unit — `Unit`, the notation of a number — `Notation`, `Written{Value, Lowest}`, the form of the search — `Search`, — and `UnicodeVersion`. `verdict` calls it, and `verdict.UnicodeVersion` is `normal.UnicodeVersion`; `unitsOf` stays in `verdict`: it reads the check.
- **The rules** (`module`, `leaks.go`):
  - `task.no-leaks` within a part — the texts of the task itself: `givens`, `unknown`, `leading_question`, the `text` of each trap value and the `simpler_case` of each item, and of an instance of a template the picture path of each part of its statement and its solution, whose names the numbers fill. The picture path of a task without a template is its author's name, and is not searched. `CheckFile` holds them too: `format-check` checks an example of a part the same way, without other parts;
  - `task.no-leaks` between parts — the title and description of each technique that the task names, the title and description of each solving step, which every task goes through, the `trap_advice` of each trap that it names — in `traps` and in the trap values —, and the rung templates of the manifest, when the task has an item with a verdict;
  - a breach of `task.no-leaks` is in the file of the task, one per text, at the place of the text; for a text of another part — at the place where the task names it: `#/techniques/<number>`, the first place of the trap, and for a solving step and a rung template — `#`. The message names a text of another part by its file and place, `techniques.yaml#/techniques/draw/title`, each key found by the place of its answer slot, and a phrase by its place;
  - `lesson.no-leaks` between parts — the title of the lesson, the paragraphs, the `alt` of the pictures and the texts of the numbered lists of its explanations, the paragraphs and the `alt` of the pictures of the statements and the demos of its worked examples, by the keys and phrases of each task of the module that the lesson names; in a worked example on a workbook, the leak phrases in the text pieces of its lines, but not the keys, and the line of the example that solves a line of the task — the same pieces in the same order, the texts in the normal form, the value of each box equal to its key as the check of the box compares it. A breach is at the place of the text in the lesson, one per text and task, and names the segment of the task and the file of the task — the path by which the lesson found it, not its `id`: a file with another's `id` breaks `module.file-ids`, and the rules between parts run all the same;
  - the search of a task is built once per its path, `checker.searchOf`, and serves both rules: within a part, between parts and each lesson that names it; an instance of a template is checked by a `checker` of its own, so its search is built per instance;
  - a template both rules read by its instances, the texts of the module and of a lesson by the keys and phrases of each instance. A task, a technique and a trap that the module does not have the rules skip: that is `lesson.known-tasks`, `task.known-techniques` and `task.known-traps`.
- **The vector test.** The runner `leaks` reads a vector by the schema and the type of `formatv0` and checks the set of what was found in each text against it.
- **Tests:**
  - `leak` — `TestIn`: texts that free a key, a choice written by hand, two keys of one magnitude, the order of what was found, a byte not in UTF-8 and a placeholder; `TestFor`, `TestNumbersIn`, `TestHolds`, `TestConcurrentIn`, `TestKeyOf` and `FuzzNumbers`. The fuzz test writes a number of each form with leading zeros: between characters that are not of a word the search finds it, by itself and once, and with a character of a word beside it neither it nor its tail; on any text the numbers are the same as those the model reads — the regular expression of the longest run;
  - `module` — `TestLeakRules`: each place of a breach of both rules on a whole module; `TestLeakRulesOfAFile`, `TestLeakRuleMessages` and `TestLeakRuleNamesTheTasksFile`. `TestChecksTheFormat` holds that neither an example nor a vector of another rule breaks these rules;
  - `internal/normal` — `TestSteps`, `TestNFKCIsStreamSafe` and `FuzzNotation`.

— *Why:*
- the search is built once per task: `api` searches every one of the tutor's words on one attempt, the module check dozens of texts of one task, and the free keys are computed from the texts of the item once;
- `ErrTemplate`, not a search by the template as it is: the phrase `{a} rabbits` would be found nowhere, and the tutor's words that give the answer away would pass silently;
- normalization is in `internal/`, shared by the verdict and the leak: neither `verdict` nor `leak` gives it out, since the forms are the format's business, not the API's;
- the answer slots of an item are `verdict.AnswerSlots`: the order of answer slots is one in the verdict, the observation and the leak, and `Leak.Slot` is the same place as in the verdict;
- `task.no-leaks` within a part and between parts: a rule within a part reads its own file, and `CheckFile` checks an example of a part, as `format-check` does, which has no other parts in it;
- the breach is in the file of the task: the rule is a rule of the task, and the report puts `task.*` in the file of the task, as with `task.known-traps`. Where the text of another part stands, the message says;
- a byte not in UTF-8 is U+FFFD, not an error: the leak check is a safeguard, and a text with such a byte is searched whole, rather than passing silently. An error of `In` would add a branch to each call of `api`, and would find no leak.

— *Rejected:*
- one function `Find(task, text)`: the free keys and the forms of the phrases would be computed for each text;
- the search of a phrase by comparing characters at each place, as in `leaks.mjs`: `strings.Index` over the bytes of UTF-8 finds the same places faster;
- normalization as exported functions of `verdict`: the forms of the leak would stand in the API of the verdict;
- one breach per task, as in `format-check`: the author would not see which text to fix;
- a breach of a text of another part in that part's file: the rule is the task's, and for a technique shared by several tasks each has its own leak.

— *Consequence:* `api` takes `leak` on the tutor's words, and the verdict and the leak normalize with one code. — `leak/`; `internal/normal/`; `verdict/`; `module/leaks.go`; `internal/vectors/`; README; [architecture.md](architecture.md).

## The generator and the instance of an attempt

**R31. The generator is the package `internal/random`: `Stream` with five functions, `New(high, low)` — a stream from two words, `Words(seed, names)` and `Place(seed, names...)` — a place of randomness, and on an argument outside the domain — the error `ErrDomain`, and the function draws no integers. What goes outside is `template.AttemptInstance(seed, task, count)` — the number of the instance of an attempt: `below(count)` of the place with the names `template`, the task's ID and `instance`, for count from 1 to `MaxInstances`, 10,000, and a refusal is `template.ErrDomain`. The vectors `random/` are passed, and the comparison with `math/rand/v2` and `hash/fnv` runs as an ordinary test on every `go test` and as a fuzz test with the same corpus.**

- **API** (`internal/random`), by the randomness of the [format text](../spec/format/README.md):
  - `New(high, low uint64) *Stream` — a stream whose state is the two words as they are; `Words(seed uint64, names []string) (high, low uint64, err error)` — the words of a place; `Place(seed uint64, names ...string) (*Stream, error)` — its stream. The constants are the numbers of the format text, with the hexadecimal notation beside them;
  - `Integer() uint64`, `Below(b uint64) (uint64, error)`, `Between(a, c int64) (int64, error)`, `Fraction() float64` and `Permutation(n int) ([]int, error)`, and `MaxPermutation` is 2^31 − 1. The order is an `[]int`, as with Go's `Perm`;
  - `ErrDomain` — one error for all the refusals, and what is wrong the message says: `below(0)`; `between` with a > c and with a count above 2^63 − 1 — first a ≤ c, then `uint64(c) − uint64(a)` no greater than 2^63 − 2; `permutation` outside 0…2^31 − 1; a place without names, and a name that is empty, with `|` or not in UTF-8. A refusal of a function draws no integers, a refusal of a place builds no stream;
  - a `Stream` is not for two goroutines at once, and the package has no shared state: a stream lives in the call that built it.
- **The instance of an attempt** (`template/attempt.go`):
  - `AttemptInstance(seed uint64, task formatv0.TaskID, count int) (int, error)` — the number from 0, in the order of the counter. A count outside 1…`MaxInstances` and an ID that is not a name of a place are the error `template.ErrDomain`, and with the ID of a valid task this never happens (`common.schema.json`). `template.ErrDomain` is `random.ErrDomain`: `api` and the eval runner of `tutor` do not import the package `internal/random`, and `errors.Is` finds it, whichever refusal gave it;
  - the function takes no template and counts no instances: their numbers, in order, the package carries, and count is their number. The instance itself is `template.Instance` by these numbers.
- **The vector test.** The runner `random` is `randomDraws` in `internal/vectors`: a vector is read by the YAML reader, the schema `random-vector.schema.json` and the type `formatv0.RandomVector`, and the numbers written as strings by `strconv`. If a seed and names are given, `Words` is checked against `state`; then the stream from the words of `state` gives the calls in order, and a fraction is compared as a double. A call without a known field fails the case, though the schema would not let it pass anyway. `TestRandomDrawsRefuse` holds that the runner catches a wrong field of each kind and an integer that a word of 64 bits does not hold.
- **Tests:**
  - `internal/random` — `TestWords`: the words of the vectors, each refusal of a place, wherever the invalid name stands, and é of one character and of two — different places; `TestDraws`: the number of integers of each function, in steps from state to state — `below(1)` and `between(a, a)` draw one, `permutation(0)` and `permutation(1)` none, `below(2^63 + 1)` from the words 0 and 2 — five, as in `below-refusals.yaml`; `TestRefusals`: on each refusal — `ErrDomain`, the zero value and the previous state; `TestEdges`;
  - the comparison with the standard library (sources 2.5) — `TestReference` and `FuzzReference` with a shared `compare`: the words of a place against `hash/fnv`, the names as one string, joined by `|`, as the text of a place writes them: there is no `|` in a name, so the fuzz test reaches any list of valid names, with a line break too, and then, on one stream, `integer`, `below`, `between`, `fraction` and `permutation` against `Uint64`, `Uint64N`, `Int64N`, `Float64` and `Perm`, and after each function the state of the stream is compared with `PCG.MarshalBinary`. So the number of integers that a function draws is checked too. `TestStateAsPCGWritesIt` holds that `MarshalBinary` writes the state as the test writes it;
  - the inputs of `TestReference` — the words of 12 vectors and the words 0 and 2^64 − 1, each with each bound of `below` — 0, 1, powers of two and their neighbours, 2^63 ± 1, 2^64 − 1 — and in turn with the edges of `between`, counts of permutations up to 1024 and names of places, about 2,700 inputs, and 5,000 more inputs that the generator itself gives from the seed 2026. The corpus of `FuzzReference` is the same inputs, except those that the generator gives: they are not data but a consequence of the code. A permutation longer than 1024 the comparison only rejects or skips;
  - `template` — `TestAttemptInstanceOfTheVectors`: the seeds of all the vectors of `templates/`. The number of the instance is derived from the numbers of the vector in the order of the counter and written in the test, and the test reads the vectors from `spec.Format` and holds that they give the same seeds and the same numbers, and that `Instances` at the number of an attempt gives its parameters: on a new tag of `spec` the copy will not fall behind the vectors. A template with a solver is counted with a solver that returns 0, not with the solver of the vector, which may run up to the budget: the number of an instance does not depend on the solver. `TestAttemptInstanceIsItsPlace`: on each count from 1 to 10,000 the number is below count and is the one that the place `template`, ID, `instance` gives; `TestAttemptInstanceRefuses`.

— *Why:*
- `Words` and `New` are exported by the package: the vector test compares the derived words with `state`, and the stream from the words with the calls. `Place` is for a template, which needs no words;
- one `ErrDomain`: the caller needs to know that the argument is outside the domain, and which one and why is the message's. `internal/random` does not go outside, and `template` gives the error out under its own name;
- `AttemptInstance` takes count, not the template: `api` counts no instances and executes no expressions — the package carries their numbers;
- the task's ID is not checked against the form `local_id`: a place is defined on any name, the ID comes from a checked package, and the form is held by the module check. A name that is not a name of a place is an error, not a panic;
- the state after each function is compared with `PCG.MarshalBinary`, not by the next integer: so the number of integers drawn is checked exactly, not by a coincidence of numbers.

— *Rejected:*
- a function of the instance of an attempt from the template, `formatv0.Task`, with a count of the instances: it would need the counter of the expressions, and `api` knows the number of instances from the package;
- a check of the ID against `common.schema.json` in `AttemptInstance`: a second place for the form `local_id`, and it does not change the number;
- an error for each function: the caller would tell apart five errors of one meaning;
- a check of the number of integers by the next integer of two streams: two different states can give one next integer.

— *Consequence:* the vectors `templates/` are passed by seeds through `AttemptInstance`: the instance of a seed is the number among the instances that the template builds. — `internal/random/`; `template/`; `internal/vectors/`; README; [architecture.md](architecture.md).

## Task templates

**R32. Task templates are the package `template`: `Instances(ctx, task)` checks a template by the rules that it breaks by itself — `task.template-names`, `task.template-expressions`, `task.placeholders`, `task.template-instances` and `task.template-solver` — and counts its instances in the order of the counter, with the results of its solver, and `Instance(task, numbers)` substitutes the numbers of an instance in a single pass, without expressions and Starlark. An expression of the subset is read and typed by a parsing of its own, as in `format-check` of `spec`; beyond it, go.starlark.net of the format must parse it, and the sandbox computes it: `sandbox.Compile` compiles an expression once, and `Func.Call` evaluates it on each combination. The module check checks a template by its instances: the rules of a task that read the strings of the task — on each instance, and a breach names the numbers of the first instance that breaks the rule at that place; the rules that read no strings — on the template as it is written; `task.instances` holds each instance by the schema of a task. The vectors `templates/` and the vectors of the rules of templates are passed.**

- **API** (`template`), by the templates of the [format text](../spec/format/README.md):
  - `Instances(ctx, task) ([]formatv0.Numbers, []Breach, error)` — the numbers of each instance in the order of the counter: each parameter, each derived number and each result of the solver by its name. A template that breaks a rule of templates gives breaches and not a single instance; a task without a template — neither. The count looks at `ctx` before each combination, and the solver runs with the same one; the error is only a `ctx` that is done, and it is never a breach;
  - `Instance(task, numbers) (formatv0.Task, error)` — the instance: each placeholder of each string outside `template` is its number, in decimal digits and with `-` before a negative one, and a doubled brace is one, in a single pass; there is no `template`, and the `id` is the template's. The function executes no expressions and no Starlark: `api` substitutes the numbers that the package carries, by the number of `AttemptInstance`, and the open view of the instance is `openview.Task` of it. An error is for a task without a template, for a string that breaks `task.placeholders`, and for a placeholder whose number is absent; the instance is not checked against the schema of a task — that is `task.instances` of the module check;
  - `Breach{Rule, Pointer, Message}` — a breach of a rule of templates with the place of the value, and `RuleNames`, `RuleExpressions`, `RulePlaceholders`, `RuleInstances` and `RuleSolver` are the IDs of these rules: the rules of templates are the validity of the template itself, and the package build needs this same count. `module` gives them under its own names, `RuleTaskTemplateNames` and the others, and adds `RuleTaskInstances`;
  - `Shown(template, numbers)` — the numbers of an instance, as a breach names them: the parameters in the order of the names, `a = 1, b = 4`, as `instanceShown` in `templates.mjs` of `spec`.
- **The order of the rules** — as in `format-check`: the names, the expressions, the placeholders and the program of the solver are all checked, and if at least one is broken, the combinations are not counted. Then — an empty domain, at each of its parameters, `#/template/parameters/<name>`; the product of the sizes of the domains is a `big.Int`, and more than 10,000 is at `#/template/parameters`, with the exact number; the first combination on which an expression fails or a derived number goes beyond ±(2^53 − 1) — one breach at that expression, and the count stops; not a single instance — at `#/template/constraint`; then the solver on the instances. The names are at `#/template/derived/<name>` and `#/template/solver/results/<number>`, the expressions at their own place.
- **Expressions:**
  - the parsing — its own tokens, grammar and types of the subset (`template/expression.go`, `typeOf`), as `expressions.mjs` of `spec`: tokens are separated by spaces U+0020 or by nothing, a name or an integer is the longest run of `[a-z0-9_]`, an integer has no leading zero, comparisons do not chain, and the types `int` and `bool` follow the rules of the format text. An expression begins with a token, not a space, and nests no deeper than 64 levels — each parenthesis, unary `-` and `not` is a level: the dialect reads a space before the first token as an indent, "got indent", and go.starlark.net parses no deeper than 75 parentheses or 987 unary minuses in a row, "excessive nesting", so every expression of the subset is one that the dialect parses. An error names the token and its place;
  - the dialect — an expression must also be accepted by `sandbox.CheckExpr`, as a check predicate is: the subset is Starlark, and it runs in the dialect of the format. An expression of the subset that the dialect still does not parse is a breach, not a panic;
  - the evaluation — `sandbox.Compile(src, names...)` once: a program of one assignment `_result = <expression>` with the names as predeclared, built from the tree of the parsed expression, not from the text — it compiles everything that `CheckExpr` takes, a comment and a line break too, is as deep for the parsing as the expression itself, and makes the same steps as `Eval` —, and `Func.Call` on each combination — `Program.Init` on its own thread, with its own values of the names. An expression has no limit of steps, no clock and no `context`: it calls no functions and holds no loops, and its steps are bounded by its length. Division by zero is a breach of `task.template-instances` with the words of Starlark: "floored division by zero", "integer modulo by zero". The derived numbers are computed in the order of the names, then the constraint — on the parameters and derived numbers in the same order; `and` and `or` do not compute the right operand when the left one has decided;
  - the order of the names is by code units, and a name is ASCII, so this is the byte order of Go.
- **Placeholders** (`template/placeholders.go`) — a string is read from left to right: `{{` is `{`, `}}` is `}`, `{name}` is a placeholder, and any other brace is an error at the place of its character. The strings of a task outside `template` are the strings of its JSON value, `schema.ValueOf`, the picture paths too. A breach — one per string: a brace, a name that does not exist, or a result of the solver outside the key and the trap value of `number` and `fraction`, the solution and the leak phrase — in a picture path, of the solution too, a result never stands. Outside a template, a brace in a picture path, and a key or a trap value of `number` and `fraction` with a placeholder, are `task.placeholders`, the latter at `…/check`, one per answer slot: a placeholder is still recognized in one place, `verdict.ErrPlaceholder`.
- **The module check** (`module/templates.go`):
  - a template is expanded once, for all the rules that read its instances: the breaches of its rules, and without them each instance, its numbers and which keywords of the schema of a task it does not pass;
  - within a part — the rules of a template, `task.instances` and, on each instance that the schema accepts, the rules that the slot check shows by itself, `task.known-pictures` on its picture paths and `task.no-leaks` on the texts of the task; between parts — `task.no-leaks` on the texts of other parts and `lesson.no-leaks` of a lesson that names the template, by the keys and phrases of each instance. Each instance is checked by its own `checker`, so the search of the leak is built per instance, not per path;
  - the author fixes the place where a rule is broken: the report keeps the breaches at that place of the first instance that breaks the rule there, and each begins with "on the instance first = 2, second = 7:". 10,000 instances whose trap value is their key give one breach, not 10,000. An instance that breaks the rule at the same place otherwise — a text with the key of one instance and the key of another — is seen when the place is fixed;
  - the rules that do not read the strings of a task — `task.skill-roles`, `task.zones`, `task.known-*` but `task.known-pictures`, `task.own-anchors`, `task.assessability`, `task.age-categories`, `manifest.rung-templates`, the rules of a lesson except the leak — read the template as it is written, and a breach stands once, without numbers. Their fields — IDs, words of enumerations, numbers and paths — hold no braces by the schema, so for each instance they are the same as for the template;
  - `task.instances` — each keyword of the schema at each place, by its first instance that does not pass it: different keywords are different errors —, except `unevaluatedProperties`, if there are others: the keys of an instance are the same as those of the template, which passed the schema, so a field stays "unevaluated" only beside a subschema that did not pass — and that is a keyword of its own. Otherwise one long digit of a key would give seven breaches.
- **The cost** (measurements 1.5): a template of 10,000 instances `CheckFile` checks in 1.5 s, most of it the schema of the instances and the round trips of JSON, and the count of the expressions in the sandbox takes less than 0.1 s.
- **The vector test.** The runner `templates` reads a vector by the schema and the type `formatv0.TemplateVector`; the template is checked by `module.CheckFile` — its JSON, and JSON is YAML —, and the runner expects an empty report; then the number of instances, and for each attempt the number of `AttemptInstance` by the seed, its numbers, the instance, equal to the vector's as a JSON value, and its open view.
- **Tests:**
  - `template` — `TestTypeOf`: each rule of the tokens, the grammar and the types, and a space at the start; `TestPiecesOf`, `TestStringsOf`, `TestInstances`: the order of the counter, a constraint with lazy `and` and `or`, floored division; `TestInstancesBreaches`: each rule of templates at each place, an expression that nests 65 levels and the deepest that the subset takes; `TestInstance`, `TestShown`, `TestTakesResults`. `FuzzTypeOf` checks the parsing against go.starlark.net: an accepted expression it accepts too, and on any two integers it gives a value of the same type or fails with division by zero. `FuzzPiecesOf`: what was read and written back is the same string;
  - `internal/sandbox` — `TestCompile`: the order of the names, the name `_result` among the names, a comment and a line break — in parentheses, at the start and before the expression —, an error of the expression, the steps of the vectors at the limit — the same as those of `Eval` —, and a cancellation of the `context`; `TestConcurrentCall`;
  - `module` — `TestTemplateRules`, `TestTemplateRulesOfAModule`, `TestTemplateRuleMessages` and `TestInstancesByKeyword`; `TestLeakRulesOfAFile` — a template is searched by its instances; `TestChecksTheFormat` holds that neither an example nor a vector of another rule breaks these rules;
  - `internal/schema` — `TestValueOf`.

— *Why:*
- the meaning of a template is in `template`, together with the IDs of its rules: the rules of templates speak of the template itself, and the package build needs the same count; the rules of a task on instances are rules of `module`;
- a parsing of its own of the subset, and beyond it the parsing of go.starlark.net: the messages are in the words of the subset and about its tokens, the parsing is the same as that of the second implementation, and that the dialect takes less is seen as a separate breach, not silently;
- an expression is compiled once: a template has up to 10,000 combinations, and parsing with compilation on each would repeat the same work;
- the `context` — before each combination, not in each expression: an expression of the subset is bounded by its length, so the count sees a cancellation through the combination, and a template without expressions sees it too;
- the rules without strings — by the template: 10,000 identical breaches "task.known-band" do not help the author, and the result is the same;
- the breaches of the first instance at a place, not one per rule, as in `format-check`: the author sees each place to fix and the numbers on which it breaks.

— *Rejected:*
- computing the expressions with arithmetic of its own on `big.Int`, as `format-check` does: the format requires the meaning of Starlark and the dialect of the format, and arithmetic of its own would be a second implementation inside one;
- reading the subset by a walk of the tree of go.starlark.net: its parsing takes more — unary `+`, `in`, tuples, the conditional expression —, and the messages would be in the words of Starlark, not of the subset;
- `lambda <names>: (<expression>)` in `Compile`: the wrapper spends levels of parsing, and an expression at the limit that `CheckExpr` accepts would not compile — the fuzz test found this;
- the `context` in each expression: a template without expressions and a solver would not see it at all;
- checking all the rules of a task on each instance and naming the numbers at each breach;
- keeping in the report each distinct breach at a place: for a template where the trap value is the key, there would be as many of them as instances, "the trap value N gets correct" for each N;
- in `task.instances` — all the keywords of the schema, with the cascade of `unevaluatedProperties`;
- reading a template once for all its instances by a separate API: 1.5 s per template of 10,000 instances, and the API would grow for the sake of one module check.

— *Consequence:* the package build takes the numbers of `Instances` into `instances` of the canonical view, and the digest of the open view of an instance from `openview.Task(Instance(…))`; `api` builds the instance of an attempt so: `AttemptInstance`, `Instance` by the numbers from the package, `openview.Task`. — `template/`; `internal/sandbox/`; `internal/schema/`; `openview/`; `module/templates.go`; `internal/vectors/`; README; [architecture.md](architecture.md).

## Template solvers

**R33. The solver of a template is executed by the package `template` inside `Instances(ctx, task)`: after the count of the instances the program of the solver runs once — `sandbox.Exec` on its own thread, with the format's budget of 100,000,000 steps and the clock of `module-kit` of 60 s, `sandbox.Solver` —, and `solve` is called on each instance in order on the same thread, with a frozen dict of the numbers that the open fields of the template show, `Inputs(task)`. The global values of the program after its top level are frozen, as Starlark freezes a module. The results go into the numbers of the instance, and the rules of a task, `task.instances` and the leak read the instances of a template with a solver as they read those without one. A solver that fails is one breach of `task.template-solver`, on the first instance where it fails, with its numbers, or before the first, and the count stops; a `context` that is done is an error without a breach. `module.Check(ctx, fsys)` runs the solvers one per CPU. The vectors `rules/task.template-solver/` and `templates/` with a solver are passed.**

- **The sandbox** (`internal/sandbox`):
  - `CheckProgram(src)` — a program of the dialect: the parsing by `syntax.FileOptions.Parse` with the options of the dialect; `load`, wherever it stands, is "the dialect has no load" at its place; the names — `resolve.File` without predeclared names and with its own list of the 33 built-in ones, not with `starlark.Universe`, as for `CheckExpr`. An error is the line, the column and the words of go.starlark.net: `2:1: got newline, want ':'`, `2:45: undefined: gcd`;
  - `Exec(ctx, limits, src)` — the `context`, then `CheckProgram`, the parsing again and `starlark.FileProgram` outside the thread: a program that `CheckProgram` accepted and the compiler did not is a defect of `module-kit`, a panic, not the `ErrPanic` of a solver. Then `Program.Init` on its own thread and the freezing of the global values, as the Starlark specification requires and `starlark.ExecFileOptions` does, while `Program.Init` itself does not freeze (sources 4.9). An error is as for `Eval`: the `context`, `ErrSteps`, `ErrClock`, `ErrPanic` or an error of the top level; then there is no `Program`;
  - `Program.Global(name)` — the value of a global name; `Program.Call(fn, args...)` — a call on the thread of the program, within what the top level and the earlier calls have left of the steps and the clock, with the same errors; after a panic `Call` calls nothing and returns the same error, and after `Close` it returns `ErrClosed`: the program no longer has the clock and the watching of the `context`. `Program.Close()` stops the clock and the watching of the `context`, and it can be called again;
  - `Solver` — `Limits{Steps: 100_000_000, Clock: time.Minute}`;
  - the thread is a type of its own: `start` — a thread with the limit of steps, the clock and the watching of the `context`, `do` — one evaluation on it with `recover` and the first reason of the stop, `end` — the end of the clock and of the watching. `run` is `start`, `do` and `end` for one evaluation, and `Program` is `start` in `Exec`, `do` for the top level and for each `Call`, `end` in `Close`. So the predicate, the template expressions and the solver stop by one code;
  - the count is checked against the vectors: the top level of `within-the-budget` is 10 steps, the call of `solve` with a loop of 9,999,996 turns is 99,999,980, 99,999,990 in all, and there is a result; one turn more — `reaches-the-budget` — and the count reaches 100,000,000, and the call stops. The steps are 10 per entry into a Starlark function, per return of each call in it, of a built-in function too, and per each backward jump, and the thread sees a cancellation there too (sources 4.9): a clock that fired between calls stops the next call at its entry;
  - which error is the `context`'s is said by `sandbox.ContextDone`, one for the sandbox: the verdict of a predicate takes it too.
- **The solver** (`template/solver.go`), by the solvers of the [format text](../spec/format/README.md):
  - `Inputs(task)` — the names of the numbers that `solve` gets: the parameters and derived numbers whose placeholders stand in the strings of the open view of the template as it is written, `schema.Open`, in the order of code units — as `solverInputsOf` in `templates.mjs` of `spec`. The openness of the fields is held by the schema of a task, and the list of the format text is its retelling: the statement, and the `alt` of its picture too, the givens and the unknown, the text of the item, the pieces of its line and the options of an answer slot bound `choice`. The key, the trap value and the trap text, the solution, the leak phrase and `simpler_case` are closed. A name in doubled braces is not a placeholder, and a string that breaks `task.placeholders` gives no names of its own;
  - together with the rules of the names, the expressions and the placeholders — `CheckProgram`: a program that does not parse, holds `load` or names a name outside the dialect breaks the rule "before the first instance", however the instances would be counted, and the combinations are then not counted;
  - after the count — `Exec`, and `solve` must be a function of the program — `def` or `lambda`, `*starlark.Function` — with one parameter, `NumParams() == 1`, which may have a default value, — without keyword-only parameters, `*args` and `**kwargs`: its only parameter gets the dict positionally. go.starlark.net counts `def solve(*, numbers)` as a function of one parameter, and a call with one argument fails on it; the words of the breach write the parameters as `def` writes them: "a, *more, b, **rest". A built-in function is not a function of the program, and `*args` would get a tuple;
  - on each instance — a new `starlark.Dict` with the numbers of `Inputs`, inserted in order, each a `starlark.MakeInt64`, frozen, and a call on the same thread. The result is a `*starlark.Dict` whose keys are strings, the names of the results, exactly those, and whose values are `starlark.Int` within ±(2^53 − 1): `bool` and `float` are not `int`. The order of the check of the result: not a dict; the first key of the dict that names no result; the first result in the order of `results` that is absent, is not an `int` or is out of bounds;
  - a breach is one, at `#/template/solver/program`: "the solver fails before the first instance: …" or "on a = 2, b = 7, …", the numbers being `Shown`, as for `task.template-instances`. An error of the program is the words of go.starlark.net and the place of the innermost call of the program, as its backtrace ends: "solve fails: fail: no factor fits, at 5:9 in solve"; the steps — "the count of steps reaches 100000000, the budget of the template"; the clock — "module-kit's clock stops it after 60 s"; a panic — "the interpreter panicked: …";
  - the caller's `context` the count looks at before each combination, and the solver runs with it: `Exec` and each `Call`. Its cancellation is an error of `Instances`, without numbers and breaches: "template: the task times: context canceled" before the count or in it, "…: sandbox: context deadline exceeded" in the solver.
- **The module check** (`module/templates.go`) — `module.Check(ctx, fsys)` expands the templates of the module in a pool of `runtime.GOMAXPROCS(0)` goroutines with the caller's `context`, each by `Instances`; `module.CheckModule` gives the numbers of the instances too, and `modulepackage.Build` takes them from it and does not run the solver a second time. Only `CheckFile` — a file on its own — runs with `context.Background()`, and an error that such a `context` does not give is a panic of a defect. Each rule that reads the instances of a template reads those with the results too. `RuleTaskTemplateSolver` is in the list of rules.
- **The vector test.** The runner `templates` checks the `inputs` of a vector — in any order — against `Inputs`, and fails a vector without a solver that has `inputs`; the numbers of each attempt are checked against the numbers of the instance that its own run of the solver gave.
- **Tests:**
  - `internal/sandbox` — `TestCheckProgram` — also that a name that the process added to `starlark.Universe` is not a name of the program; `TestProgramSteps` — a budget of 120 on a program whose top level is 10 steps and `solve(n)` 20 + 10·n: a result at a count of 110 and a stop when the count reaches 120, in one call and across several, — the same limit as `within-the-budget` and `reaches-the-budget` hold at 100,000,000, — and `ErrClosed` after `Close`; `TestProgramGlobals` — the freezing; `TestProgramClockAndCancel` — the clock between calls, the `context` and `ContextDone`; `TestProgramPanic`, `TestConcurrentExec` and `FuzzCheckProgram`, which executes nothing: `x = x + x` doubles a string without a loop or a call (Д10), so it holds the parsing, `load` and that what is accepted compiles;
  - `template` — `TestInputs`: each open and closed field; `TestSolve`: the dict in order, `int`, the bounds, `lambda` and a default value; `TestSolverBreaches`: each breach with its own words; `TestSolverStops`: the budget of a template, of an instance and of the top level, and the clock on lowered `solverLimits`, and 1300 at a count of 1290 and 1290 at that count — as `within-the-budget` and `reaches-the-budget`; `TestInstancesContext` and `TestConcurrentInstances`;
  - `module` — the cases of `templateRuleCases` with a solver: its own rules, a solver that fails on an instance, `task.trap-values` and `task.no-leaks` on the results, — and their words in `TestTemplateRuleMessages`.
- **`FuzzCheck`** takes a corpus without the vectors of `task.template-solver`: four of them run up to the budget, about a second under `-race`, and each run of the fuzz test checks a file twice — `Check` and `CheckFile`. A solver comes into the corpus with the example `number-thought-of-template`.
- **The cost** (measurements 1.6): a solver at the budget takes 0.10–0.16 s, and the vectors of the solver multiply the time of `module` and `internal/vectors` under `-race` by about 1.5 and 3.

— *Why:*
- the solver inside `Instances`: the format names `task.template-solver` a rule of a template, and the numbers of an instance "every parameter, every derived number and every result of the solver" (`formatv0.Numbers`), and both the module check and the package build await them — one call gives them all. This is a function for one template: how many templates run at once, the caller decides;
- the solvers one per CPU: the clock runs by wall time, so there are no more concurrent evaluations than `GOMAXPROCS`. The pool is in the library, not in the CLI: otherwise the CLI would have to expand the templates itself, and `Check` gives no API for this, and `module-registry` needs the same pool;
- a second run of the solver in `Build` would cost one more run up to the budget per template — up to 60 s for the worst solver, — and the check has already computed the numbers of the instances;
- the thread is in the sandbox, not in `template`: the limits, the clock, the `context`, `recover` and the first reason are the same for the solver as for the predicate and the expressions;
- the freezing: the format requires it, as the language does, and without it a solver could remember earlier instances, and the result of an instance would depend on the order not only through the steps;
- `Inputs` — by the open view of the template as it is written, as in the second implementation, not by a list of fields: so both read one schema;
- the program is checked before the count: it does not depend on the instances, and the author sees it beside the other breaches of the template;
- one breach and a stop at the first failure: the budget and the clock are shared, and after a failure they say nothing about the following instances, and the format requires naming the instance "on which the solver fails".

— *Rejected:*
- `Solve(ctx, task, instances)` apart from `Instances`: the module check and the package build would call two functions, and numbers without results are not the numbers of an instance of the format;
- not to freeze the global values, like `Program.Init`: the language requires freezing;
- to check `solve` only by a call: `def solve(*args)` would pass, although its parameter gets a tuple, not a dict;
- a breach on each instance where the solver fails: after a failure of the budget they cannot be counted, and identical failures do not help the author;
- the whole backtrace in the breach: the line is longer than the report, and the place where the program failed is the last frame of the program;
- the pool only in the CLI: the library has no API to expand templates from outside;
- the vectors of `task.template-solver` in the corpus of `FuzzCheck`: above; a fuzz test that executes programs — a built-in function takes a gigabyte per step (Д10), and the worker of the fuzz test would fall from memory.

— *What is not closed:* memory — Д10.

— *Consequence:* the package build takes the numbers of the instances, already with the results, into the `instances` of the canonical view; `api` does not execute the solver: the package carries the results. — `template/`; `internal/sandbox/`; `module/`; `modulepackage/build.go`; `verdict/`; `internal/vectors/`; README; [architecture.md](architecture.md).

## The module package in code

**R34. The canonical text is written by `internal/canonical.Text` — the function of the canonical serialization over a tree of JSON values, with a refusal `*canonical.Error` at the value's place. The package is built by `modulepackage.Build(ctx, fsys)`: first `module.CheckModule`, and a module with breaches is a `*Refusal` with the report, not a package; then the canonical view from the JSON value of each file of a part as it is, integers given as doubles — as integers, `skill_tiers` by `module.Tiers` and `instances` by the numbers that the check gave, with the results of the solver; the listing of the device part, the inventory and the digest. The package is the bytes of its files, `Package.Files`, and the CLI's `build` writes them to disk. `modulepackage.Read(fsys, digest)` reads a package as files by the chain: the inventory by the package digest, then the parts by the inventory and the files of `device/` by the listing, and checks each text against the schema and against its canonical text. The digest of the open view of a task and of an instance is `openview.TaskDigest`, `sha256:` of the canonical text of the open view. The vectors of `packages/` are passed.**

- **The canonical text** (`internal/canonical`), by the canonical serialization of the [format text](../spec/format/README.md):
  - `Text(value)` — an object, an array, a string, `int64`, `bool` and `nil`; the members of an object go in the order of UTF-16 code units, `utf16.Encode`, so a character outside the BMP comes before U+E000–U+FFFF; a string — `\"`, `\\`, `\b`, `\t`, `\n`, `\f`, `\r`, the other characters up to U+001F — `\u00` and two lowercase digits, the rest as it is;
  - a refusal — `*Error{Pointer, Message}` at the first value that the text does not hold: a fractional number, an integer given as `float64` — the build converts it, before the function —, an integer outside ±(2^53 − 1), a string or a name not of UTF-8, a value of another Go type, `int` and `map[string]string` too. The place is a JSON pointer, as with the breaches of the check;
  - `Digest(bytes)` — `sha256:` and 64 lowercase digits, `Descriptor(bytes)` — the digest and the size, as types of `formatv0`;
  - `Parse` — JSON with an integer as `int64` and another number as `float64`, shared by reading a package and the index;
  - the tests — `TestEscapes`, `TestEveryControlCharacter`, `TestText` — the order, nesting and a surrogate pair against U+FB01, — `TestRefusals`, `TestDigest` and `TestParse`. `FuzzText` builds a value of the format from its input and checks `Text` against `jsontext.Value.Canonicalize` (sources 2.4); the file is `//go:build goexperiment.jsonv2`, and with `GOEXPERIMENT=nojsonv2` the Go package builds and the tests run without it.
- **The build** (`modulepackage/build.go`, `view.go`, `device.go`):
  - the module's files — a walk of the folder without names beginning with `.` and the folders under them, and without `eval/`; a member of the view — by the layout: `manifest`, `skill_graph`, `traps`, `techniques`, `solving_steps`, and `tasks/<id>.yaml` and `lessons/<id>.yaml` — by the local ID from the file's name, which the check has already checked against `id` (`module.file-ids`);
  - the value is what `internal/yaml` gave, and `integers` turns a `float64` with an integer value within ±(2^53 − 1) into `int64`: the `5e1` of the vector `strings` is `50`. Another `float64` stays, and `Text` rejects it — on a checked module there is none, the schemas hold the numbers to integers;
  - `instances` — by the templates in the order of IDs, the numbers of the instances that `module.CheckModule` gave with the caller's `context`: the solver runs once, in the check. A cancellation of the `context` is an error of `Build`, not a refusal;
  - the inventory — `module_format_version`, `id` and `version` from the manifest and the descriptors of the two texts; the package digest is the digest of the text of the inventory. `Package` also carries the inventory, the view and the listing, read strictly into the types of `formatv0`, — what `api` takes;
  - a refusal of the text on a module that has passed the check is a defect of `module-kit`: an error with these words, and a value that the schema accepted and the type does not read is a panic, as in `module`.
- **Reading** (`modulepackage/read.go`, `files.go`):
  - the inventory — by the package digest, it has no size; then its format version by the place that all versions keep — another version is `ErrFormatVersion` —, the schema of the inventory and the file being equal to its canonical text. The parts — by the descriptors of the inventory, with the schemas `canonical-view.schema.json` and `$defs/device_listing`; the manifest of the view names the same `module_format_version`, `id` and `version` as the inventory. Then the files of `device/` by the listing, in the order of paths. `ReadParts` reads the same without `device/`;
  - the refusals — `ErrChain`: there is no file, or the digest and size are not the right ones; `ErrFormatVersion`; `ErrNotCanonical`: not JSON, something extra after the value, not by the schema, not the canonical text — `1.0` or a space —, an inventory of another module. Each — with the file's path. A file of the folder that the chain does not name is not read;
  - JSON is read by `canonical.Parse`, and the check against the canonical text rejects a number that is not an integer. The check of the text makes a separate check of digits and escaping needless: a package read is byte for byte the one that `Build` built;
  - the tests — `TestRead` — a package as files is read back into the same `Package`, with extra files beside it, — `TestReadParts`, `TestReadRefuses`, a case for each refusal, and `TestReadUnreadable`. `FuzzRead` gives three arbitrary texts, reads them by the digest of the inventory with the files of `device/` of the vector `with-media-and-notice`, and holds that `Read` does not panic, an error carries one of the three reasons, and an accepted package is exactly the files read. The corpus is the three texts of the package of each vector of `packages/`.
- **The digest of an open view** (`openview.TaskDigest`): the task — as a value of `formatv0`, its JSON value is checked against the schema of the task, the open view — the walk of the schemas `schema.Open`, as with `Task`, but before the type `formatv0/open`, so that the text has exactly what is in the view, — and `sha256:` of its canonical text. For a template it is the digest of the open view of the template itself, with its placeholders in place; for an instance — of the open view of `template.Instance`, with the numbers substituted, which a sample names. The vectors of `open-view/task/` and of `templates/` hold the digests; `TestTaskDigest` and `TestInstanceDigest` hold those that the second implementation of `spec` gives.
- **The vector test** — the runner `packages`: `package.yaml` — by the schema and the type of the vector; the package of `module/` — by `Build`; its three texts are byte for byte equal to the canonical texts `server_part`, `device_part` and `inventory` of the vector, the inventory and the digest — to the values of the vector, and the package as files is read by `Read` back into the same one.

— *Why:*
- the check inside `Build`: a package is built only from a checked module, a refusal of the function of the text then means a defect, and the caller cannot build a package bypassing the check. The cost is a second parsing of the files of the parts, milliseconds;
- the canonical view — from the values of `internal/yaml`, not from the types of `formatv0`: the canonical serialization takes the value "as it is", without default values and the order of a struct's fields. The types — only for `skill_tiers`, `instances` and what `Package` gives;
- the package as bytes, not a write to disk: the library has no input or output, and `api` does not keep packages as a folder;
- reading checks each text against its canonical text, not only the digest: the digest holds the bytes, and the schema and the canon — that these bytes are a package of the format, so the caller gets a view that it trusts, and an inventory of another module will not pass;
- `TaskDigest` — by the JSON value of the walk, not by `open.Task`: so the digest does not depend on how the type of `spec` writes empty fields, and matches the second implementation by construction.

— *Rejected:*
- `Build` without the check, by the caller's ready report: a package from an unchecked module would become possible by one error of the call;
- `Read` that checks only the digests: a file with foreign JSON and a right digest is impossible, but an inventory whose parts name a package of another module would pass;
- the digest of an open view through `json.Marshal(open.Task)` and `Text`: it would depend on the tags `omitzero` of the types of `spec`;
- a function `Write(dir)` in `modulepackage`: writing is the CLI's business, and Go's `fs.FS` does not write.

— *Consequence:* the module index writes its text by `canonical.Text` and names the packages by `Package.Digest`; the command `open` writes `open-views.json` with the same text, and the digest of each task by `openview.TaskDigest`. — `internal/canonical/`; `modulepackage/`; `openview/digest.go`; `internal/vectors/`; README; [architecture.md](architecture.md).

## The module index

**R35. The module index is built by `moduleindex.Build(Release)`: the previous index — by its file or by an explicit `First`, the release's packages — `*modulepackage.Package`, the tag and the withdrawals by `$defs/withdrawals`. The entries of the previous index stand as they were, a package of a version that the previous index names leaves that version's entry, the others get a new one with the release's tag, and the withdrawals are marked last. The build is rejected by a `*Refusal` with all the breaches of the rules `module-index.*`, and the index's file is the canonical text `canonical.Text`. The build reads the previous index as the build writes it: by the schema of its own version, by the canonical text and with `teaches` in order, — and it does not read an index that a newer format version built. `moduleindex.Read` reads an index of any format version by the places that all versions keep, gives each entry, whatever format version it is of, and counts a version as withdrawn for any reason. `moduleindex.Newest` chooses the version of a new attempt — the newest by semver unwithdrawn one within the format's window. The parsing and the order of a module's versions are `internal/semver`, shared with version comparison, and the parsing of JSON is `canonical.Parse`, shared with reading a package; it rejects a text not in UTF-8 and a number outside double. The vectors of `indexes/` are passed, and the valid examples of the index hold what the schema does not say.**

- **The build** (`moduleindex/build.go`), by the module index of the [format text](../spec/format/README.md):
  - `Release{Previous, First, Tag, Packages, Withdrawals}`. "No previous one" is `First`, and an empty `Previous` is not an index: a file that was not read will not start a new index. `First` together with `Previous` is an error too;
  - the previous index — JSON; then its format version by the place that all versions keep — a version newer than that of `module-kit`, `module.NewerFormat`, is a refusal; then the schema `module-index.schema.json` of the version of `module-kit`, the file being equal to its canonical text, and the `teaches` of each entry in the order of code units, which the schema cannot say. Only then — the type `formatv0.ModuleIndex`. An index that did not pass is `ErrPrevious`. The build holds each entry "as it was" in an index of its own version, which passes its schema, so an index with a member or a reason that its schema does not know cannot be read, and an index of a newer format version means that the CLI in CI of `modules` was rolled back, and the build does not read it either, as the format text requires;
  - the packages are checked before the rules: each is not `nil`, of the format version of `module-kit`, with the module's ID and version by their forms and an entry that `$defs/entry` accepts, — as `modulepackage.Build` and `Read` give them. Otherwise — `ErrRelease`, and before any refusal: the rules are read only for a release where everything is in form. Then the packages go in the order of the module's ID, the version by semver and the digest, so neither the index nor a refusal depends on the order of the input, and the version of each is parsed once. A package of the same version with the same digest a second time changes nothing;
  - a new entry — the `module_format_version` of the inventory, `Package.Digest`, the tag and `teaches`: the anchor skills from the `teaches` of the skills of the graph of the canonical view, each once, in the order of bytes. The ID of an anchor skill is ASCII, so this is also the order of UTF-16 code units. A module without a graph and skills without anchors have no `teaches`;
  - the rules: `module-index.one-package` — two packages of the release of one version with different digests, or a package and an entry of the previous index with another digest; `module-index.version-above` — a new version below some version of its module from the previous index, withdrawn ones too, and the entry stands all the same, so that its withdrawal gives no second breach, as with `index.mjs` of `spec`; `module-index.known-withdrawals` — a withdrawal of a version that is not in the new index; `module-index.entries-kept` — a withdrawal with another reason than that of the withdrawn version. The build has no other ways to break `entries-kept`: it does not touch the entries of the previous index. A repeated withdrawal with the same reason changes nothing, and a version that comes out in this same release can be withdrawn too;
  - a refusal — `*Refusal{Breaches}`, where `Breach{Rule, Module, Version, Message}`, in the order of the module, the version, the rule and the words: "module-index.version-above: example/times-tables 0.1.1: it is new in the release, and the previous index names 0.1.10 and 0.2.0 above it: a new version stands above every published one, withdrawn ones too". A build with a breach is rejected as a whole, and does not build an index without what breaks the rule;
  - the error `ErrRelease` — a tag not of the form `vX.Y.Z`, withdrawals not by `$defs/withdrawals` — an empty object too —, a package not of the form above, and a first release without packages: an index names at least one version (`modules`, `minProperties: 1`);
  - the result — `module_format_version` `module.FormatVersion` and the entries, as canonical text. Its bytes the CLI's `index` writes to `module-index.json`.
- **Reading** (`moduleindex/read.go`), by the format text, which keeps some places of the index in all its versions:
  - `Read(data)` takes from the root only `module_format_version` and `modules`, and from an entry — `module_format_version`, `package_digest`, `release`, `teaches` and `withdrawn`; a member at another place, which the next version may add, is not read. `withdrawn` is any string, and the version is withdrawn whatever the reason: `Withdrawn` carries it as written. The other members are checked by the index's schema by their forms: the forms and meanings of these places are also the same in all versions, and a member not in its form is `ErrNotIndex`;
  - an entry of a format version that `module-kit` does not read, `Read` gives together with the others: version comparison needs each published version of the module, and it is the choice of a version, `Newest`, that skips it. The index is not rejected because of it;
  - `Read` does not require the canonical text: it is part of the format version, and the next major may change it. A name that stands in an object twice — no build writes it — is taken by the last member, as `encoding/json` reads it;
  - `Newest(index, id)` — the newest by semver version of the module among those that are not withdrawn, for any reason, and whose format version is within the window, `module.ReadsFormat`. The same one `api` gives a new attempt, by it the module teaches the anchor skills of its entry, and `open` takes it, with the marks of withdrawal of the newest index.
- **Shared code:**
  - `internal/semver` — `Parse` of a module version by the form `common.schema.json#/$defs/module_version` and `Version.Compare` by the precedence of Semantic Versioning 2.0.0, its item 11. The numbers have no bound, as in the format: they are held as text and compared first by length, so 1.10.0 is above 1.9.0 and no version overflows;
  - `module.NewerFormat(v)` — whether the format version MAJOR.MINOR is newer than `module.FormatVersion`, by the values of the numbers: 0.10 is newer than 0.9. It stands beside `ReadsFormat`, at the window of versions, and compares the version MAJOR.MINOR as the module version MAJOR.MINOR.0: the forms of their numbers are the same;
  - `canonical.Parse` (`internal/canonical/json.go`) — parsing JSON with integers as `int64`, in the Go package whose tree of values it builds; `modulepackage` and `moduleindex` call it. It rejects a text not in UTF-8, which `encoding/json` would read with U+FFFD in place of each bad byte, — so `Read` will not accept a withdrawal reason with broken bytes, — and a number outside double: `encoding/json` gives `1e700` as `+Inf`, and the schema validator panics on it. `FuzzBuildPrevious` found such an input — the input lies in `moduleindex/testdata/fuzz/` —, and `TestReadRefuses` of `modulepackage` holds the case of an inventory whose `id` is `1e700`.
- **Tests:**
  - `moduleindex`: `TestRulesAreTheFormats` — `Rules` are equal to the list of refusals of `index-vector.schema.json`; `TestBuildFirst`, `TestBuildNext` — the previous entries as they were, an entry of an earlier format version, a repeated package, a withdrawal of a new version and a repeated withdrawal; `TestBuildRefuses` — each rule with its words and several breaches in order; `TestBuildErrors` — a case for each error, a previous index of a newer format version, `teaches` out of order and a package not in form beside a breach of a rule; `TestTeachesOf`; `TestReadExamples` — the valid examples of the index are read as by the schema and the type, the invalid ones are `ErrNotIndex`, except `unknown-reason`, which `Read` reads as a withdrawal; `TestReadLaterVersion` — a member and a reason of the next version, an entry of version 1.4, a text that is not canonical; `TestReadRefuses`; `TestNewest`. In `module` — `TestNewerFormat`, in `internal/semver` — `TestParse`, `TestParseIsTheSchemas`, `TestCompare` and `FuzzCompare` against `math/big`;
  - the fuzz tests: `FuzzRead` holds that `Read` does not panic, an error is `ErrNotIndex`, the canonical text of an accepted index is read into the same index, and `Newest` takes an unwithdrawn version of the window no lower than the others; `FuzzBuildPrevious` gives an arbitrary previous index to the release of the vector `next-release` and holds that the build does not panic, gives an index, a refusal or `ErrPrevious`, and a built index holds each entry of the previous one as it was. The corpus of both is the indexes of the vectors of `indexes/` and the examples of the index;
  - the vector test — the runner `indexes`: `index.yaml` — by the schema and the type of the vector; the package of each module of `modules/` — by `modulepackage.Build`; the previous index — by its canonical text, as the release of `modules` holds it; the built index is byte for byte equal to the canonical text of the vector's `index`, and a refusal — only by the rule of `refusal`. The previous and the built indexes `Read` reads back into the values of the vector;
  - the valid examples of the index the runner `examples/valid` also holds to what the schema cannot say, as `indexFormProblems` of `format-check` of `spec` does: the root names the version of `format/VERSION`, and the `teaches` of each entry is in order (`indexForm`, `TestIndexForm`).

— *Why:*
- an explicit `First`, not `nil` in place of the previous index: `os.ReadFile` of an empty file gives an empty slice, not a `nil` one, and an empty file read by mistake would start a new index silently — what the format text forbids;
- the previous index — as the build writes it, not by places: the build writes an index of its own version, which its schema accepts, and the entries in it must stand "as they were", byte for byte. The previous index is signed, and `module-kit` built it, so in fact it is the canonical text with `teaches` in order; anything else would mean that something is wrong in the chain;
- a previous index of a newer format version is a refusal, even without new members: a release whose CLI is older than the CLI of the previous release is a rollback, and it must fail loudly, not rewrite the root of the index to the old version;
- the packages — by form before the rules: otherwise a package not in form beside a breach of a rule would give a refusal by the rule, and the author would not learn that the package was not built by `modulepackage`;
- `Read` — by places and without the canonical text: so the format text requires of a reader of other versions, and `api`, `diff` and `open` read indexes that other versions built too;
- `Read` gives the entries of other format versions too, and `Newest` skips them: `diff` needs each published version of the module and its format version, and not rejecting the index because of them the text requires of any reader;
- a refusal with all the breaches at once, in the order of the module and the version: CI of `modules` shows the author everything in one run, and the order does not depend on the order in which the CLI read the package folders;
- parsing versions without `math/big`: the length and the order of digits give the order of values when there are no leading zeros, and the form forbids them.

— *Rejected:*
- `Build(previous []byte, …)` with `nil` in place of "no previous one": an empty file read by mistake would start a new index;
- to read the previous index by places and silently drop the members that the schema does not know: an entry would become other than it was, and `entries-kept` would not notice it;
- to accept a previous index of a newer format version when the schema lets it pass, and write the root with its own version, as `index.mjs` of `spec` does: a rollback of the CLI would pass silently;
- `Read` that drops the entries of format versions outside the window: `diff` would have to read the index once more, with code of its own;
- `Read` that requires the canonical text: an index of the next major version with another canonical text would not be read, although its places are the same;
- comparing versions through `math/big`, like `BigInt` in `versions.mjs` of `spec`: for the sake of the comparison the numbers would have to be parsed, while the length and the digits give the order;
- a parser of JSON in `modulepackage` and a second one in `moduleindex`: two parsers with one job would diverge.

— *Consequence:* `moduleversion` takes `internal/semver` and `moduleindex.Read`: all the versions of the module and their format versions. — `moduleindex/`; `internal/semver/`; `internal/canonical/json.go`; `module/check.go`; `modulepackage/files.go`; `internal/vectors/`; README; [architecture.md](architecture.md).

## Version comparison in code

**R36. A new version of a module is compared with the published ones by `moduleversion.Compare(next, published)`: the published ones — packages as files without `device/` with their digests, and each is read by `modulepackage.ReadParts` by the chain of digests; the newest — by semver, withdrawn ones too. `Changes` finds the version part of two packages of one format version by the table of version comparison of the [format text](../spec/format/README.md), and gives a report by places: each row of the table — at its own exact place, and everything else — a patch at each place where the views diverged. `Steps`, `StepPart` and `SameMajor` hold the rule of 0.y.z. The vectors of `versions/` with a part are passed.**

- **The report** — by places: a table row — at the place that it names, an answer slot — `answer_slots/<j>` or `line/<k>` in the item, a change of an instance — at the place in the task, with the numbers of its parameters; everything else, except `manifest.version`, — a patch at each place where the JSON values of `server.json` and `device.json` diverge, by members and elements, and the `instances` of a template — as a whole. A difference at the place of a table row or under it, by tokens, is not written, and under `skill_tiers` and `instances/<ID>` — when the skill graph or the task itself diverged: they follow them. The order — by part from major, then by place. `Comparison.String()` writes the report as lines, like the breaches of the module check, and `--json` — the same `Comparison` with JSON tags; the command `diff` prints it.
- **The input** — the published versions as packages as files with their digests, read by the chain of digests: CI of `modules` gives `diff` the packages without `device/`.
- **The meaning of a check** is compared with the same code by which the verdict reads it: `Compared` and `Units` in `internal/normal`, and the public API of `verdict` does not carry them. For a choice the meaning is held by `notation` too — `options` without it —, and by the options in the normal form of that notation; for a time, by the key and the traps by hours and minutes. The leading question and the description of a solving step are a patch by the rule "anything else".
- **Shared code:** the order of places — `schema.ComparePointers`; raising a version — `semver.Version.Raise`; `SameMajor` — for `api`, which moves a learner within the same major.
- **Tests** — `moduleversion`: `TestChanges`, `TestCompare`, `TestComparisonString`, `TestPartsAreTheFormats`, `TestSteps`, `TestStepPart`, `TestSameMajor`, `TestMeaningOfNotationAndTime` and `FuzzDiff`; `internal/semver` — `TestRaise`.

— *Why:*
- the code repeats the meaning of `foundPartOf` of the second implementation of `spec`, `versions.mjs`, which holds all the vectors; the report is more exact: in `versions.mjs` everything beyond the table rows is one line of patch `#`, and the author would not see what gave the patch. The parts do not change from this: each table row is no lower than patch;
- `skill_tiers` and `instances` stay in the comparison: `Read` does not recompute them, and a package of another builder may diverge only in them — then it is a patch, as in `versions.mjs`. Only a difference that follows another one is not written;
- packages as files: that is what CI of `modules` gives `diff`, and the published versions are trusted only by the chain of digests;
- `Compared` — in an internal Go package, not in the API of `verdict`: the meaning of a check is read by the verdict and the comparison, and `api` does not need it.

— *Rejected:*
- one line of patch, as in `versions.mjs`: the author would not see what gave the patch;
- the input — `*modulepackage.Package` in memory: a published package would not be checked by its chain of digests;
- to exclude `skill_tiers` and `instances` from the comparison: a package of another builder that diverged only in them would get "none" where `versions.mjs` gives patch;
- a second `comparedBy` in `moduleversion`: two codes of one meaning would diverge.

— *Consequence:* `moduleversion/`; `internal/semver/`; `internal/normal/compared.go`; `internal/schema/errors.go`; `modulepackage/read.go`; `internal/vectors/`; README; [architecture.md](architecture.md).

## The rules of versions and the format transition

**R37. The rules `module-version.*` are held by `moduleversion.Compare`: each breach is a `Breach{Rule, Message}` in `Comparison.Breaches`, and `Refused()` says that the new version is rejected. `above` — for each published version, withdrawn ones too; `step` — for the newest, only when `above` let it pass; the other six — by the places that the format keeps in all versions, for each published one, so they hold across a format transition too. A published package of an earlier format version is read by the chain of digests and only by these places: the part is major while the format is v0. All the vectors of `versions/` are passed.**

- **The places** (`moduleversion/places.go`) — the format version, the module's ID and version from the inventory; the languages of the manifest, the IDs of the bands in order, the keys of skills, tasks, lessons, traps and techniques, and `retired_skills` with `replaced_by` from the canonical view. They are read by one function over JSON values, both for a package of the current version and for an earlier one: the rules see the packages the same way. A package without a place where the format keeps it is `ErrNotCanonical`.
- **Reading a published package** (`moduleversion/published.go`): `inventory.json` by the package digest and the format version by the place that all versions keep. A version that `module-kit` reads — `modulepackage.ReadParts`. An earlier one — `server.json` by the digest and size that the inventory gives, without the schema and without the canonical text: they are set by the version of the package, not of the reader. A newer one and one not of the form MAJOR.MINOR — `ErrFormatVersion`.
- **The part across a transition** — `Major` by one change `inventory.json #/module_format_version: major: the module moves from version <earlier> of the format to <current>`. The table does not compare packages of different format versions: so does `versions.mjs` too (`movePart`). From v1 on, the format text gives no part of a transition, and there is no question yet: the window of reading is the one current version, and earlier ones are read only by places.
- **The order of breaches** — `above`, `step`, then `retired-complete`, `retired-kept`, `retired-known`, `ids-kept`, `languages-kept` and `bands-kept`; the words — as in `placeBreaches` and `compareVersions` of `versions.mjs`. The report writes each as a line `refused: <ID>: <words>` before the changes, and `--json` as `breaches`; `diff` gives the exit code 1 by them.
- **The vector test:** a refusal of a vector — exactly one rule, its own; a vector with a part — not a single breach and the same part.
- **Tests** — `moduleversion`: `TestBreaches`, `TestRulesAreTheFormats`, `TestPlacesOf` and `TestReadPublished`.

— *Why:*
- the format text requires reading a package of an earlier version by its places alone, not by the schema of its own version, so the schema of the current version is not applied to an earlier package, and the digests of the chain are;
- the six rules by places hold across a transition — the format text: "the rules hold through those places alone"; a shared function of places rules out that a rule sees a package of the current version otherwise than an earlier one;
- `step` only without `above`: so the table of rules requires ("applies to a new version that `module-version.above` lets pass"), and the vector `below-the-newest` would otherwise breach two rules.

— *Rejected:*
- a refusal of a package of an earlier format version by an error: the format promises to read it, and the vectors of a transition with a part expect this;
- to read an earlier package through `modulepackage`: that one checks the schema and the canonical text of its own version, while the places are all that is promised;
- a breach of `step` together with `above`: a refusal vector would breach two rules, and the format text requires one.

— *Consequence:* `moduleversion/`; `internal/vectors/`; README; [architecture.md](architecture.md).

## The commands `check` and `build`

**R38. The CLI has the commands `check <module>` and `build <module> <folder>`. Their exit code is shared: 0 — accepted, 1 — a refusal by the format rules, 2 — the command could not do its work. The report is text to stdout, a line per breach and per warning and a line of the summary, and there is no `--json`. The template solvers run one per CPU inside `module.Check(ctx, fsys)`, and `build` runs each solver once. The package is written to a temporary folder beside the target one and moved to its place. The release carries the `LICENSE` and `NOTICE` of each third-party module that the binary is built from, and `ci-lint` checks their list.**

- **The report**: a breach — as the line `Breach.String()` with the path from the folder that the command line gave, — `<path>[:line:column] [#/pointer]: <rule ID>: <message>`; a breach of a schema — `schema <keyword>`, a file that is not YAML — `not YAML`. Then the warnings, the numbers in a text of the whole module and then the pictures, which do not change the exit code. The last line is the summary: "<module>: the module breaks no rule of the format <version>" or "<module>: the format <version> refuses the module: N breaches". The report goes to stdout, an error of the call and of input or output — to stderr, as one line `module-kit <command>: …`.
- **The solvers** — `check` calls `module.Check` and `build` `modulepackage.Build`, which takes the numbers of the instances from `module.CheckModule`, with a `context` that SIGINT or SIGTERM makes done.
- **The folder of `build`**: it does not exist or is empty, or else code 2 and nothing is touched. The package is written to the folder `package` inside a temporary `.<name>.<random>` beside it and moved into the target one, and the temporary folder is removed whatever the outcome. The permissions of folders are 0755, of files 0644, as the process's umask adjusts them. The temporary folder of `os.MkdirTemp`, accessible only to its owner, the command does not turn into a shared one: an `os.Chmod(…, 0o755)` on it is what SonarCloud's rule `go:S2612` marks. A refusal by the rules is code 1, the report and nothing on disk; a path of a file of the package outside its folder is an error.
- **`--version`** names the release on its first line and `verdict.UnicodeVersion` on the second.
- **The licenses in the release**: the recipe `release-licenses` in the justfile collects `go list -deps` of the binary on the six platforms of the release and checks the list of modules against `THIRD_PARTY` and `SAME_LICENSE`, which name each module with the name its license takes in the release, `LICENSE-<name>`, and `NOTICE-<name>` beside it when the module has a `NOTICE`. `golang.org/x/sys` does not go into the binary for Windows, and its file goes with all of them. The `LICENSE` of `github.com/altnovatrail/spec` is byte for byte our Apache-2.0, and the `LICENSE` of the release replaces it; the recipe checks this too. `release-files` puts the files into `dist/` and into `SHA256SUMS`, and `ci-lint` calls the recipe: a module without its license fails the pull request, not the release.

— *Why:*
- the text report is read by CI of `modules` and the author; nobody has asked for `--json` yet, and adding a flag is an extension that breaks nothing;
- a temporary folder and a rename will not leave half a package if the writing breaks off, and the command does not touch a non-empty folder, because someone else's files may lie in it;
- the list of licenses in the justfile, checked against `go list -deps`, catches a new dependency in its own pull request: otherwise a module without its license would reach a release unnoticed.

— *Rejected:*
- overwriting a non-empty folder: it would erase what lay there;
- `LICENSE-spec` beside `LICENSE`: the same text twice. When the `LICENSE` of `spec` changes, the recipe will fail, and it will be entered into `THIRD_PARTY`.

— *Consequence:* `cmd/module-kit/`; `module/check.go`, `module/templates.go`, `module/rules.go`; `modulepackage/build.go`, `modulepackage/view.go`; justfile; README; [architecture.md](architecture.md).

## The warning of a transparent picture

**R39. The Go package `picture` warns of a picture of the module in which even one pixel is not fully opaque: `Warnings(ctx, fsys)` reads each PNG and WebP in `media/`, and `Check(name, data)` one picture. PNG is decoded by `image/png`, WebP by `golang.org/x/image/webp`, BSD-3-Clause. A picture larger than `MaxPixels`, 2^25 pixels, the Go package does not decode, but warns that it did not check it; the same for a picture that cannot be read. `check` prints the warnings as lines `<path>: warning: …` after the breaches and before the summary, and the exit code does not change from them.**

- **The Go package** — `picture`, outside `module`: `go list -deps` of the Go packages `module`, `modulepackage`, `moduleindex`, `moduleversion`, `observation`, `openview`, `verdict`, `leak` and `template` does not name `x/image`, so `api`, which takes `modulepackage`, and with it `module`, does not carry the decoder. It is called by `check` of the CLI, and from stage 4 by `module-registry`.
- **What to read with** (sources 4.10): `x/image` reads `ALPH` and VP8L, and rejects an animated WebP with the error `webp: invalid format`; it requires the versions of `x/sys` and `x/text` that `go.mod` already holds. JPEG has no transparency.
- **What is transparent** — a pixel with an alpha less than 0xFFFF, the first by rows. The message names its place and the alpha out of 65535 and repeats why the format needs a light opaque background. A picture that itself knows that it is opaque — `Opaque()` of the types of `image` — is checked at once. A PNG with an alpha channel where all the pixels are opaque, and a palette with `tRNS` whose transparent entry no pixel takes, get no warning.
- **The limit** — `MaxPixels = 1 << 25`, 33,554,432 pixels, for example 5792 × 5792. A decoded pixel is no more than 8 bytes (`image.NRGBA64`), so the picture will take no more than 256 MiB. The size is read by `DecodeConfig` — the header, without decoding —, and a file whose header promises more is not decoded: so a picture bomb will not eat the memory. The pictures of the examples and vectors of `spec` are 1 × 1, so the limit is chosen by memory, not by them: a picture that is shown in a task statement with a width in CSS px is much smaller.
- **An unreadable picture** — `DecodeConfig` or `Decode` give an error, as for an animated WebP, — this is a warning "module-kit cannot read / decode it as …, so it could not check", not code 2: `module.media` checks only the signature, and the module is accepted. A file without the signature of its format gets no warning: `module.media` rejects it, and the report would say one thing twice.
- **What is read** — the files of `media/` with the extension `.png` and `.webp`, in the order of paths; files and folders with a name beginning with "." — not of the module — are skipped. MP3 is not a picture. A module without `media/` has no warnings; a folder that cannot be read and a done `ctx` `Warnings` gives as an error — code 2 of the CLI.
- **The tests** — the PNG files the test makes itself with `image/png`: opaque, with an alpha channel but opaque, with one transparent pixel, semi-transparent, a palette with `tRNS` — with the transparent entry in use and without —, gray. The WebP files the test builds by RFC 9649, without files in `testdata/`: VP8L of one color — one simple prefix code per channel, — a simple file and an extended one with VP8X, opaque and transparent, and an extended one with the animation flag, ANIM and ANMF. `TestTestFiles` holds that `x/image` decodes them as the tests think. `TestFormatPictures` — none of the pictures of the examples and vectors of `spec`, except the vectors of `module.media`, is transparent, and each is read. `FuzzCheck` checks `Check` against the decoder with the limit lowered to 2^16.

— *Why:*
- the format requires warning of a picture "with a pixel that is not fully opaque", and in WebP transparency can be in `ALPH` with compression 1 and in VP8L — without a VP8L decoder it cannot be seen;
- a package of its own, not `module.Report`: `api` must not carry the decoder, and `module-registry` needs the same warning;
- the limit is needed because anyone writes a picture, and `DecodeConfig` gives the size cheaply; 256 MiB is a ceiling that the CLI bears on the author's machine and in CI;
- an unreadable picture is also something the author needs to know of: a browser may not show it either. But the format does not refuse for it, so it does not change the exit code either;
- WebP files that the test builds by the specification are repeatable and carry no foreign files with their licenses; the test files of `x/image` would do — `gopher-doc.with-alpha.lossless.webp` and `yellow_rose.lossy-with-alpha.webp` are transparent, — but the module does not record their provenance.

— *Rejected:*
- a parsing of WebP of its own: `ALPH` with compression 1 and VP8L require a VP8L decoder, and a copy of its own is more code and risk; WebP without a check: the format requires warning of each picture;
- the warning in `module.Report`: `x/image` would go into the binary of `api`; only in `cmd/module-kit`: `module-registry` would not get it;
- the warning at the first opaque place of each part of the module, by the references `picture`: the format speaks of each picture of the module, and a file without a reference to it is no more the concern of `module-kit` than a file with a reference;
- to count as transparent a pixel with an alpha less than 0xFF00 — a tolerance for rounding: the format says "not fully opaque", and a semi-transparent shadow on dark paper is seen just as badly;
- to decode a picture of any size: the memory is that of the caller's process, and Go has no limit for `image`;
- to warn of a file without a signature: `module.media` already rejects it.

— *Consequence:* `golang.org/x/image` is a direct dependency in `go.mod`, raised by Dependabot `gomod`; `THIRD_PARTY` of the justfile names it, and the release carries `LICENSE-x-image`. — `picture/`; `cmd/module-kit/main.go`, `cmd/module-kit/build.go`; `go.mod`, `go.sum`; justfile; README; [architecture.md](architecture.md).

## The commands `index` and `diff`

**R40. The CLI has the commands `index` and `diff` over the library. `index` builds the index of a release of `modules` from the previous index or an explicit `--first`, from the package folders, the tag and the withdrawals as a file or an explicit `--no-withdrawals`, and writes `--out`, which does not exist yet. `diff` compares a new version — a module's folder or a package — with each published version that `--index` names, by their packages from `--published`, or is told `--first` explicitly.**

- **The call of `index`** — `module-kit index (--previous <module-index.json> | --first) --tag vX.Y.Z (--withdrawals <withdrawals.yaml> | --no-withdrawals) --out <file> <package folder>...`. Exactly one flag of each pair. An empty `--previous`, a flag twice or neither of a pair — code 2: a recipe of `modules` to which the shared release did not give the previous index fails, and does not start a new one. The recipe `release-files VERSION PREVIOUS` of `modules` needs only `if [ -z "$PREVIOUS" ]` and a check of whether `withdrawals.yaml` exists.
- **The inputs of `index`** — the package folders that `build` wrote: each is read by `modulepackage.Read` by the chain from the digest of the bytes of its `inventory.json`, with `device/`. The withdrawals — YAML by the rules of the YAML reader, then `$defs/withdrawals` of the index's schema. An empty file of withdrawals is not withdrawals, but a breach of form: "no withdrawals" is said by a flag.
- **The output of `index`** — `--out`, which does not exist yet: the file is written to a temporary folder beside it and moved to its place, as a package by `build`. To stdout — a line with the tag and the number of packages.
- **The codes of `index`** — 0: the index is written; 1: `*moduleindex.Refusal`, a line `<rule ID>: <module> <version>: <words>` per breach and a line of the summary; 2: `ErrPrevious` — including an index of a newer format version —, `ErrRelease`, withdrawals not in their form, a folder that is not a package, input or output.
- **The call of `diff`** — `module-kit diff [--json] (--index <file> | --first) [--published <package folder>]... <module>`. A module that the index does not name is the first version, code 0. `--first` — when there is no index at all yet. The package of each version of the module from the index, withdrawn ones too, is looked for by the `package_digest` of the entry among `--published`; without it — code 2 with the version and the digest. Extra folders are skipped.
- **The new version in `diff`** — a folder with `inventory.json` at the root is a package, `modulepackage.ReadParts` by its digest. With `manifest.yaml` — a module, `module.CheckModule` and `modulepackage.Build`, the solvers one per CPU. Neither — code 2. A module that the check rejected — the report as from `check` and code 1, without a comparison.
- **The report of `diff`** — `Comparison.String()`: the lines `refused: <ID>: <words>`, as the library writes them, and the changes. With `--json` — `Comparison` through `encoding/json`, and for a module that the check rejected — `{"check": [breaches]}`, with the path from the folder of the command line.
- **The flags** — a `flag.FlagSet` of the standard library per command, without the output of the Go package itself: a usage of its own, types of its own for a flag given or not, and for a repeated one.

— *Why:*
- the format text requires that the build be told "no previous one" explicitly, and that an unreadable previous index give a refusal, not a new index; a flag for each "no" makes a missing file an error of the call;
- packages, not modules: the recipe of `modules` builds them anyway — for the archives `sha256-<hex>.tar.gz` of its release, — and the solvers do not run a second time;
- "not in the index — the first version" is needed by the first release of a new module in a repository where others are already published. That the index is of the newest release is held by CI of `modules`, which checks it with `release-verify`. A version below the published ones is also caught by `module-index.version-above` at the index build.

— *Rejected:*
- an empty `--previous` as "no previous one" — so a missing file would silently start a new index;
- module folders at the input of `index` — the solvers would run a second time, and the recipe needs the packages anyway;
- the index to stdout — on a failure half a file would be left in `dist/`;
- to require `--first` for a module that is not in the index too — it breaks the first release of a module beside published ones;
- the text report of `check` with `--json` too — CI that reads JSON would stumble on the parsing;
- a third-party library of flags — a new license in the release for the sake of two commands.

— *Consequence:* `cmd/module-kit/`; README; [architecture.md](architecture.md).

## Binaries on each OS

**R41. Each of the six binaries of the release is run in CI on a runner of its own OS, on each pull request and push to `main`. "Binaries / build" builds them with the recipe `release-files`, as the release does, and runs the linux/amd64 binary — its result becomes the reference. "Binaries / \<platform>" on six standard GitHub runners runs `.github/binaries.sh`: the sum by `SHA256SUMS`, `--version`, `check`, `build`, `index`, `diff` and `open` on samples of the format and a byte-for-byte check of what was built against the reference. "Binaries" waits for both and stands in the ruleset `main`.**

- **The runners** — `ubuntu-24.04` (linux/amd64), `ubuntu-24.04-arm`, `macos-15-intel` (darwin/amd64), `macos-15` (darwin/arm64, M1), `windows-2025` (windows/amd64) and `windows-11-arm`: the labels of the table of standard runners for public repositories (sources 3.17).
- **The build** (`ci.yml`, the job `binaries-build`):
  - «As in the pinned .github» — every reference to `.github` is at one SHA, the mount of the Go caches in «Release files and samples» is by the lines of the header of the composite action `toolchain` of that release, and the awk of «The stage of the release» is that of the step «Toolchain stages» of the shared `go.yml`; a move of `.github` that changes either fails here, and the pull request takes it by hand;
  - «The stage of the release» fails the job once the stage `toolchain` of `.devcontainer/Dockerfile` holds more than `toolchain-core`, in which the shared release builds its files: the job builds in `toolchain`, and the two hold the same files only while `toolchain` is `toolchain-core` with its `SHELL` alone;
  - the image of `toolchain` and the Go caches — the composite action `toolchain` of `.github`, by the SHA of its tag, with `check: binaries`, twice: before the build, and at the end of the job with `save: "true"`, so a push to `main` saves the Go caches once the binaries and their reference are built;
  - «Release files and samples» — `docker run` as the runner's user, `HOME=/tmp`, the Go caches mounted at `/go-cache`, `just release-files v0.0.0-ci` and `just binaries-samples`. The recipe `binaries-samples` copies the samples from the Go module `spec` of the version in `go.mod` — CI does not take the submodule — and makes the copies writable: the Go module cache holds them read-only.
- **What the run checks** (`.github/binaries.sh`, bash: Git Bash on Windows, `shasum -a 256` on macOS):
  - the binary's sum is the one in `SHA256SUMS`;
  - `--version` — two lines, `module-kit v…` and `Unicode …`;
  - `check` of the example `with-templates` — code 0: the Starlark sandbox and the solver, and on linux and darwin the mmap of `x/sys`, by which go.starlark.net holds integers;
  - `check` of the vector `task.known-band/unknown-band` — code 1 and the rule's ID in the report;
  - `build` of the vector `packages/with-solver` — its digest: the canonical text and the paths on Windows;
  - `build` of the modules of the vector `indexes/first-release`, `index --first` over their packages and `diff` of one of them — code 0;
  - `index --first` over the packages of `indexes/first-release` and the package of `packages/with-solver`, then `open` over them with the seeds 1 and 2^64 − 1;
  - the files of the built package and of the open views by sha256 and the index are byte for byte the same as those of the linux/amd64 reference: they depend only on the module's files.
- **One required check** — the job "Binaries" with `if: always()`, which fails if the build or even one run is not `success`. It alone stands in the ruleset `main`, as a check of GitHub Actions, not the six names of the matrix, which would change with the labels of the runners.
- **The permissions** — `contents: read` for the whole `ci.yml`, `persist-credentials: false`. The matrix takes only the script from the repository, `sparse-checkout`. The artifact lives 3 days.
- **The actions** — `actions/checkout`, `actions/upload-artifact` and `actions/download-artifact`, pinned by SHA with the version beside it, as in `ci.yml` of `spec`.

— *Why:*
- the binary carries go.starlark.net with `x/sys`, which has code of its own on each OS, and writes files by the paths of the OS: a binary that was never run on its OS may not start;
- the byte-for-byte check against the reference catches what the exit code will not tell: another canonical text or other paths in the package on Windows;
- the image and the Go caches by the action of `.github`: the entry of the image is the one that the jobs of `go.yml` fill and read, and the six cross-builds do not compile with an empty `GOCACHE` on every run;
- the cost is $0: standard GitHub-hosted runners in public repositories are free (sources 3.17).

— *Rejected:*
- a run by hand at the release: a binary that does not start would be found by the author, not by CI;
- six required checks by the names of the matrix — the ruleset would have to be edited with each change of a runner's label;
- a path filter on the job — a skipped required check would hang on a pull request without code, and the minutes are free;
- QEMU on one runner — checks neither macOS nor Windows;
- a full checkout and Go on each runner — the binary is checked as the release will build it, not by a build on the spot.

— *What is not closed:* GitHub may reconfigure or remove the labels `macos-15-intel` and `windows-11-arm`. Then the job will fail with the label's name, and a task will change it.

— *Consequence:* `.github/workflows/ci.yml`; `.github/binaries.sh`; justfile; [architecture.md](architecture.md).

## The vector test without a skip list

**R42. The vector test skips not a single case: there is no skip list, and a case that does not pass fails `TestVectors`. So each release passes every vector of the tag of `spec` that it is built on. The releases that others take are of the line that `RELEASE_LINE` of `.github/workflows/release.yml` names: CI of `modules` and `canvas` take the binary of the CLI that `release-verify` accepted, and `api` and the eval runner of `tutor` — the Go module at a tag.**

- **The test** (`internal/vectors`): the sorting into cases, the refusal on a file of a kind that the test does not know, and the refusal on a kind without a single case. Each case is a subtest, and its runner from `runners` must return `nil`.
- **Who takes the release and how** — [architecture.md](architecture.md). CI of `modules` and `canvas` take the binary at an exact tag and check it with `release-verify` before the first run. `api` and the eval runner of `tutor` take the Go module: `go get github.com/altnovatrail/module-kit@<tag>` and `go mod tidy`. The Go module is not a release file: the Go module proxy builds its zip at the tag, and `go.sum` holds its hash, as for the modules of `spec`. The `go.mod` of `module-kit` has no `replace` and `tool` that would hinder the import.

— *Why:* every release passes all the vectors, and the test does not let a skip list grow: while a list and its parsing stood in the code, any pull request could add a line to it, and only the review held this. Without the list a case that does not pass cannot be got around by anything but code, and the test holds the rule. The parsing of a list that does not exist and its helpers would be code that no run executes.

— *Rejected:*
- to keep the file and fail the test when it is not empty: the outcome is the same, but the parsing, the helpers and their tests would remain code that is not executed on an empty list, and "the list must be empty" and "there is no list" are one rule;
- to keep the list for a move to a tag: the new vectors of the tag would stand in it until their code comes, and a release could come out without passing the vectors. A move to a tag that brings vectors without code stays red until the code comes.

— *Consequence:* a move to a new tag of `spec` that brings new vectors brings their code in the same task. — `internal/vectors/`; `.github/workflows/release.yml`; README.md; [architecture.md](architecture.md).

## The command `open`

**R43. The command `open` writes the open views of a release of `modules` for the samples page of `canvas`: `module-kit open --index <release index> --latest <newest release index> --seed N [--seed N]... --out <folder> <package folder>...`. For each module it takes the version of a new attempt by the index of the release, with the marks of withdrawal from the newest index. In `open-views.json` each task has an entry with its open view and digest, and a template an instance for each seed. Beside it are the files of the device part in `modules/<ID>/<version>/`. The marks are combined by `moduleindex.WithLatest`, and the open view as a JSON value is given by `openview.TaskView` and `ManifestView`.**

- **The call** (`cmd/module-kit/open.go`):
  - `--index` is the index of the release from which the page is built; `--latest` is the index of the newest release, the same file when the release is the newest. CI of `canvas` has checked both with `release-verify`;
  - `--seed` is a seed of the format: unsigned 64-bit, in decimal digits, with no leading zero. It repeats; the same seed twice, another spelling or no seed at all — usage and code 2. The page build chooses the seeds;
  - the package folders are positional, as with `index` and `diff`, each by the digest of the bytes of its `inventory.json`. An extra folder does no harm, a missing needed one is code 2, and a release with nothing to open needs no folders;
  - `--out` — the folder does not exist or is empty, or else code 2. It is written through a temporary folder beside it, as `build` writes a package.
- **The choice of versions:**
  - `moduleindex.WithLatest(release, latest)` gives the index of the release where each entry is withdrawn if the newest index withdrew it, with its reason. An entry of the release that the newest one lacks, holds otherwise or whose withdrawal it does not hold is `ErrLatest`: such an index is not the newest, and the command exits with code 2;
  - for each module of the release's index, in the order of IDs, — `moduleindex.Newest`: the newest by semver unwithdrawn version within the format's window, as for a new attempt in `api`. Each newer version that it passed over is named on stdout — a withdrawn one with its reason, one of another format version with its number —, and a module without a version by a line `<ID>: no version to open`;
  - the package of the chosen version is read by `modulepackage.Read` by the chain of digests, with `device/`. The ID, the version and the format version of the inventory must be those that the entry names, or else code 2.
- **`open-views.json`** is a canonical text:
  - the root is `format: "module-kit-open-views"`, `version: 0`, `module_format_version` and `modules` in the order of IDs;
  - a module has `id`, `version`, `module_format_version`, `release` — the tag of the entry —, `package_digest`, `manifest` — the open view of the manifest — and `tasks` in the order of task IDs;
  - a task has `{"open_view_digest", "task"}`, and an instance of a template also `seed`, as a string of decimal digits. The instances go in the order of the numbers of the seeds. The instance — the number `template.AttemptInstance(seed, ID, count)` among the `instances` of the package, with the numbers substituted, `template.Instance`, — is the one that an attempt with this seed would get. No solvers run: the package carries the numbers;
  - `task` is the JSON value of the open view before a type took it, `openview.TaskView`, and `open_view_digest` is the digest of exactly its canonical text. So whoever reads the file gets the same digest as the sample has.
- **The files of the device part** — `modules/<ID>/<version>/LICENSE`, `NOTICE` and `media/<path>`, byte for byte those that the chain of digests checked against the listing of the part. The layout is the `device/` of the package: the path of a media file from the open view is found from `media/`, as in the module. The ID is `<publisher>/<name>`, and the pattern of the ID already excludes the names of Windows devices.
- **Codes** — 0 and 2. The command checks no format rules: the packages are read by the chain of digests, so it has no code 1.
- **Tests:**
  - `open_test.go`: `TestOpenPackages` — all the vectors of `packages/`; `TestOpenInstances` — the instances of `missing-factor` by the three seeds of the vector `templates/missing-factor.yaml`; `TestOpenMedia`, `TestOpenWithdrawn`, `TestOpenFails` — the cases of code 2 —, `TestOpenFailsLater`, `TestOpenFormatOfThePackage`, `TestOpenNoModule`, `TestModuleViewsFails`, `TestOpenHelpers`;
  - `moduleindex/latest_test.go`: `TestWithLatest` and `TestWithLatestRefuses` on the vectors of `indexes/`, and `FuzzWithLatest` over a pair of texts through `Read`;
  - `openview`: `TestViews` — the value of `TaskView` and `ManifestView` is equal to the open view of the vectors of `open-view/`, and the digest to the text of that value;
  - `.github/binaries.sh` runs `open` on each OS and checks what it wrote against the linux/amd64 reference.

— *Why:*
- `--seed`: a sample names an instance by the task's ID and the seed (`TaskReference` in `sample.proto` of `spec`). Which seeds the page needs, its build knows; an explicit "at least one" is like the other explicit "none" of the commands;
- a wrapping entry: the element of `canvas` rejects a field that the open view does not have, and the seed and the digest are not fields of a task;
- positional folders — as with `index` and `diff`, and CI of `canvas` names the packages by the entries of the index;
- `modules/<ID>/<version>/` — the same layout as that of `device/`, and the names `LICENSE` and `NOTICE` do not stand beside the media files;
- the value before a type, `TaskView`: the digest is computed from the value before a type, and the text in the file is the same value, so the match is held by code, not by chance. On the vectors of `open-view/task/` and on the examples of tasks the value of the type `open.Task` is the same, but only the agreement of the type generator of `spec` with the marks `x-open` holds this: an empty open object, which the open view keeps, a type with `omitzero` could lose;
- `WithLatest` is in the library: the rule "the newest index knows of all the withdrawals" is a rule of the format, not of the CLI, and `Newest` stays one rule of the choice of a version.

— *Rejected:*
- `--seeds <file>` with seeds by module and task — the form of the file would be invented before `canvas` asked for it;
- one seed without a flag — the page would not show different instances;
- all the instances of a template — up to 10,000 open views per template;
- `seed` and `open_view_digest` right in the object `open.Task` — `canvas` would cut them out before the element;
- `--packages <folder of folders>` — one more form of input beside `index` and `diff`;
- `media/<ID>/<version>/<path>` with `LICENSE` beside it — a layout that is neither in the package nor in the module;
- code 1 for a module without a version — this is not a refusal of the format, but what the release is: stdout names it, and the page is built from the others.

— *Consequence:* `cmd/module-kit/open.go`, `main.go`, `published.go`, `build.go`; `moduleindex/latest.go`; `openview/digest.go`, `openview.go`; `.github/binaries.sh`; README.md; [architecture.md](architecture.md).

## The devcontainer image on arm64

**R44. The devcontainer image is built whole by a workflow of its own, "Devcontainer", `.github/workflows/devcontainer.yml`: the stage `devcontainer` on `ubuntu-24.04` and on `ubuntu-24.04-arm`, as "Rebuild Container" builds it, and then `.github/devcontainer.sh` checks that the image is of its own architecture and that each tool answers with the version of its `ARG` in the Dockerfile. The workflow runs on pull requests and pushes to `main` that change `.devcontainer/**`, the workflow or the script, and by the button; it is not in the ruleset `main`, and neither file is in the list `paths` of the release.**
- **The jobs** "Devcontainer / amd64" and "Devcontainer / arm64" run `docker build --target devcontainer --tag devcontainer .devcontainer`: the context and the stage that `devcontainer.json` names. The build itself checks each archive against its sha256, and gopls against its lock file `.devcontainer/gopls/`.
- **`.github/devcontainer.sh <image> <Dockerfile> <architecture>`** — `dpkg --print-architecture` in the image is the one given, and the first line of `node`, `just`, `go`, `buf`, `gh`, `just-lsp`, `claude` and `gopls` in the image names the version of its `ARG`, and not as part of a longer number; `gopls` is in `/usr/local/bin`.
- **The permissions** — `contents: read`, `persist-credentials: false`; the checkout takes only `.devcontainer` and the script, by `sparse-checkout`, with the checkout action of `ci.yml`, by the same SHA.
- **From the network:** the image is built with no cache. What the workflow checks is that the image builds, on amd64 and arm64, with the tools at their `ARG`s, and a cache would hide a broken download.

— *Why:* the checks of `ci.yml` build only the stage `toolchain`, on amd64, so without this workflow a wrong arm64 sum or a broken step of the stage `devcontainer` would be found by whoever next opened the project, not by CI. By paths, because the image changes only with what it is built from, and a run on every pull request would spend minutes for nothing; standard runners of a public repository are free (sources 3.17).

— *Rejected:*
- one run by the button: the next raise of Go, buf or gopls in the Dockerfile nobody would check on arm64 again;
- a row «В реестр spec», so that CI of `spec` built the image on arm64: `spec` builds its own image, not that of the kind `go` with its gopls step;
- putting "Devcontainer" in the ruleset `main`: a check that the path filter skipped would never report, and a pull request without an edit of `.devcontainer/` would wait for it forever;
- checking the versions inside the Dockerfile: only the gopls step compares its version with its `ARG`, and the script holds every tool by one rule.

— *What is not closed:* the workflow does not replace a rebuild of the image on a Mac with Apple silicon, where Docker Desktop and the Dev Containers extension are their own; a wrong arm64 sum, though, is seen in CI.

— *Consequence:* `.github/workflows/devcontainer.yml`; `.github/devcontainer.sh`; `.claude/rules/ci.md`; [architecture.md](architecture.md).

## go-yaml v4.0.0

**R45. `go.yaml.in/yaml/v4` is at v4.0.0, its first release that is no release candidate, raised by hand before the `cooldown` pause of Dependabot let it through. The raise passed what stands in the pause's place:**
- the test of the format's corpus, the YAML test suite and every test of the module; the fuzz test of the reader, `FuzzRead` of `internal/yaml`, for 60 s without findings;
- the two cases of the suite that the release still names in its `yts/known-failing-tests`, 2SXE and W5VH, pass by a pass of the reader's own;
- the code takes no API that the release removed — the Limit plugin —, and `LICENSE` and `NOTICE` are the same, which `release-licenses` puts as before;
- `ci-lint`, `ci-test`, `ci-vuln` — no vulnerabilities — and `ci-licenses` are green; in `go.mod` and `go.sum` only go-yaml changed.

— *Why:* the API of a release candidate could change before the release, and the reader stood on it. The pause of three days (sources 3.11) is set against releases that are retracted or substituted soon after they come out; here the checks above replace it, and the Go proxy holds the release's hash (sources 4.4).

— *Rejected:* waiting for the pull request of Dependabot: the pause would hold the release, and the reader would stay on a release candidate's API for nothing.

— *Consequence:* `go.mod`, `go.sum`; [architecture.md](architecture.md).

## An anchor with a colon at the end of its name

**R46. An anchor or an alias whose name by YAML 1.2.2 ends with a colon — `&a:`, `*a:` — the YAML reader names a breach of `yaml.no-anchors` by its own pass, `colonAnchor` in `internal/yaml/errors.go`, rather than wait for go-yaml. The cases 2SXE and W5VH of the YAML test suite pass with it.**
- **Why go-yaml does not read them.** Its scanner takes the ASCII characters ns-anchor-char into the name of an anchor, but cuts the trailing colons off on purpose, and go-yaml closed the issue as not planned (sources 4.11). So go-yaml reads `&a: key: &a value` as the anchor `a` with a separator and fails at the second colon, "mapping values are not allowed in this context", 1:8 — a reason with no anchor in it. By YAML 1.2.2 the name of the anchor is `a:`, and the file breaks `yaml.no-anchors`.
- **What the pass does.** When go-yaml has refused with an error that `parseError` did not turn into a rule, `colonAnchor` looks at the lines before the place of the refusal, and on the line of the refusal up to its column. It searches for `&` or `*` at the start of a node: first on the line, but for spaces and tabs, after `[`, `{` or `,`, after `-`, `?` or `:` with a space or a tab, and not in a comment. If the name after it by YAML 1.2.2 ends with a colon, this place gets `yaml.no-anchors`, and the message has the words of go-yaml. Otherwise the refusal is "not YAML", as it was.
- **Tests:** `TestRead` — 2SXE and W5VH, and `&` with a colon inside a plain scalar and in a comment, still "not YAML". `TestYAMLTestSuite` passes both cases; without the pass both tests fail them. `FuzzRead` runs over the pass.

— *Why:* go-yaml will not fix this, so a wait for it would not end. The sign of the divergence is narrow and exact: the name by YAML 1.2.2 ends with a colon exactly where go-yaml cuts it. The pass runs only when go-yaml has already refused, so it does not change the verdict on the file, only the reason named, and the module author sees "the format has no anchors", not someone else's markup error.

— *Rejected:*
- accepting it as a known limitation and leaving the cases unpassed: the module author would see a reason that is silent about the anchor;
- a scanner of anchors of its own, before go-yaml, over the whole file: it would have to repeat the parsing of quotes, block scalars and flow collections, and it is needed only where go-yaml has already refused;
- proposing another reading to go-yaml: it resolved the issue as not planned, and v4 holds this on purpose.

— *Consequence:* `internal/yaml/errors.go`, `read_test.go`, `suite_test.go`; [architecture.md](architecture.md).

## A row «В реестр spec» closed by a move to a tag

**R47. A row «В реестр spec» of `run/QUESTIONS.md` leaves the table by the move to the tag of `spec` that records its answer, and the work the answer sets becomes a row of `run/DEBT.md` or a task, not an open question. A move to a tag that changes only documents is made at once, with no planning.**

— *Why:* this repository reads `spec` at the tag its submodule is on, so an answer that `spec` recorded at a later tag is not there for its sessions until the move; once it is there, the question is answered, and what is left is work. A tag that changes neither the format, nor the vectors, nor the schemas, nor the template leaves nothing to decide.

— *Rejected:*
- closing the rows without a move, by `main` of `spec`: a session here would not find the answer at the submodule's tag;
- waiting for the tag that brings the code of the answers and closing the rows with it: the rows are questions, and their answers are already recorded.

— *Consequence:* `run/QUESTIONS.md`; `run/DEBT.md`; `spec/`; `go.mod`, `go.sum`.

## Several tags in one move

**R48. A move to a tag of `spec` goes to the latest tag at once, over every tag between, when none of them changes what the code reads — the schemas and vectors of `format/`, the template, and what `module-kit` takes from the contracts —; the code, the copy of the schemas and `.devcontainer/` then stay as they were.**

— *Why:* a tag that does not touch what the code reads asks nothing of the code, and a move for each such tag would be the same move again.

— *Rejected:*
- one tag at a time: as many moves, each the same;
- waiting for the tag that brings code: until then the sessions here would read an older `spec`.

— *Consequence:* `spec/`; `go.mod`, `go.sum`; CLAUDE.md; RUN.md; [architecture.md](architecture.md).

## New code that a tag asks for, in a task of its own

**R49. When a tag of `spec` asks for new code of `module-kit` that no vector of the tag holds — a warning, which the format gives neither an ID nor vectors —, the move to the tag holds only that everything that was green is green on the new tag too, and the code comes by a task of its own, with a row of `run/DEBT.md` until then.**

— *Why:* the code comes with its tests and their review, and a move to a tag holds one thing. A module is valid without the code, since the format refuses nothing for what it is missing, so the move need not wait for it.

— *Rejected:*
- writing the code in the same move: the move would hold two things, and its review would read both;
- not moving while there is no code: until then the sessions here would read an older `spec`.

— *Consequence:* `run/DEBT.md`; RUN.md.

## The warning about a number in a text of the whole module

**R50. Of a number in a text of the whole module — a rung template `rung_templates`, a trap advice `trap_advice`, the title and the description of a technique, the title and the description of a solving step — the module check itself warns: `module.Report` has the field `Warnings`, a warning per text in which the leak check reads at least one number, and the numbers of a text are given by the function `Numbers` of the package `leak`. `check` prints them as lines `<path> #/pointer: warning: …` after the breaches, before the warnings about pictures, and the exit code does not change from them.**

- **What is warned of** — each text that the [format text](../spec/format/README.md) calls a text of the whole module: the rung templates of the manifest, the advice of each trap of the catalog, the title and the description of each technique, and those of each solving step, which the check reads as a part of their own. The texts of a task — `givens`, `unknown`, the texts of trap values, `simpler_case` — are not: they are for one task, and `task.no-leaks` holds them by the keys of its task. The texts of a lesson are not either: a lesson names its tasks, and `lesson.no-leaks` holds them.
- **What a number is** — what the leak check reads: `leak.Numbers` gives the numbers of a text by the same steps as the search for a key — `runsIn`, on which `numbersIn` stands too —, so the warning and `task.no-leaks` never diverge. A number is written as it stands in the normal form of the text: `½` as `1/2`, full-width digits as ASCII, `07:30` as two numbers, `07` and `30`, and "seven", `7th` and `x7` are not numbers. The message names each number once, in the order of the text.
- **Where** — in the package `module`, not in a package of its own, as `picture` is: the texts are parts that the check has already read into the types of `formatv0`, and the warning carries no new third-party module, so `api` takes nothing extra. The warning reads each part whose file passed the schema, in a module that the rules rejected too: the author needs to know of a number in a text in any case. A part that the schema rejected gives no warning, as it gives no breaches of rules either.
- **Order and form** — by the path of the file, then by the place in it, as with breaches; the place is a JSON pointer, without a line and a column, as with the breaches of `task.no-leaks`. The line is `<path> #/pointer: warning: <message>`, and the message repeats why the format wants a text without numbers.
- **Who gets them** — `Check` and `CheckModule`. `CheckFile` does not: it checks the examples and the vectors of parts, and the warning has no vectors. `build` and `diff` print the report of a refusal without the warnings, as they do without those of pictures.
- **Tests** — `TestNumbers` and a check in `FuzzNumbers`: `Numbers` gives the same numbers that the search reads, each in the normal form of the text with the same value; `TestWarnings` — each kind of text, the order, a rejected module, a part that the schema rejected, the texts of tasks and `CheckFile`; `TestNumbersSaid`; `TestExamplesWarnOfNoNumber` — no example of a manifest, of traps or of techniques of `spec` holds a number; in the CLI, `TestCheckWarnsOfNumbers`. `TestCheckRefuses` reads a line of a warning apart: the vectors of `task.no-leaks` `key-in-a-rung-template`, `key-in-a-technique` and `key-in-a-trap-advice` hold a key in a text of the whole module, which is also a number in it.

— *Why:* the format asks for a warning here, not a refusal: `task.no-leaks` finds a number in a text of the whole module only once it is already the key of a task that shows this text, so a number in a trap's advice is found only when some task with that key comes, later. Reading a number by the same code as the search means warning of exactly what may one day become a leak.

— *Rejected:*
- a package of its own, as `picture`: that one stands apart for the sake of the picture decoder, which `api` must not carry, and here there is no new code of dependencies, and the parts have already been read by the check;
- a separate function over `fs.FS`, as `picture.Warnings`: it would read and parse the YAML of all the parts a second time;
- warning of a number in words — "seven", «семь»: the format does not read it as a number, and a task whose answer is written so holds it in `leak_phrases`;
- a warning of each number by a separate line: an advice with the multiplication table has eight of them, and the place is one.

— *Consequence:* `leak/leak.go`, `leak/doc.go`, their tests; `module/warnings.go`, `module/check.go`, `module/doc.go`, their tests; `cmd/module-kit/main.go` and its tests; [architecture.md](architecture.md).

## A move across a new version of the format, in two

**R51. A move to a tag of `spec` that crosses a new version of the format goes in two moves, each with its own pull request: first to the last tag before the version, then to the tag with it. The template's files that a move takes — `.devcontainer/` and `.claude/settings.json` — are built by `just new-repo-draft module-kit go <card> <tag>` or the generator `tools/new-repo/` on a clone of `spec` at the tag, in a temporary folder, never into the submodule.**

— *Why:* a new version of the format is work of another size than a tag of documents and vectors that the code already passes, and the review of one large change is weaker than the review of two; the move before it is checked by itself. The generator at the tag gives exactly the template's files, while `spec/` is only read here, and `new-repo-draft` in the submodule would write into its `.repos/`.

— *Rejected:*
- straight to the tag with the new version, by one pull request: the two kinds of change would be reviewed in one diff;
- the generator run in the submodule: it writes into `spec/.repos/`, and `spec/` is only read here.

— *Consequence:* `spec/`; `go.mod`, `go.sum`; `.devcontainer/`; `.claude/settings.json`.

## One version of the format, and a release line for each

**R52. The library reads one version of the module format, the one `module.FormatVersion` names, and only it: `ReadsFormat` accepts no other, and a module of another version is refused by `manifest.format-version`. The code holds everything the version has and passes every vector of the tag without skips. A new version of the format is a new release line: `RELEASE_LINE` of `.github/workflows/release.yml` takes the next minor. The layout of a workbook page is the package `page`, with its own copy of `page.yaml` embedded and held byte for byte to `spec.Format` by a test.**
- **What the code holds of version 0.4:**
  - **`yaml.printable`** — a character outside the printable set of YAML 1.2.2, wherever it stands, in quotes too, is a refusal that names the rule;
  - **template expressions** begin with a token and are nested no deeper than 64 levels — a parenthesis, a unary `-` and `not`. This is a rule of the subset, in `typeOf`: the dialect parses every expression of the subset, and the check against go.starlark.net stays as a second line, so an expression of the subset that the dialect still does not parse is a breach, not a panic of `compiled`;
  - **the kind `time`** — `normal.Time` and `normal.TimeNotation` read a time by the steps 1–5 of a number, `H:MM` or `HH:MM` with minutes of two digits, or of three that begin with no zero, and `normal.Clock` compares the hours and the minutes as numbers, without a sum of minutes: `5:60` is not `6:00`. The verdict `timeVerdict`, the traps, `NormalForm` and the join of two readings go by the same. The leak check finds a key of a time where a time of the text has the same hours and minutes, and a time of a text is a run of digits, `:` and digits in the notation of a time with no character of a word at its edges;
  - **`notation: word`** of a choice written by hand — `normal.ChoiceOf`, in `internal/normal/choice.go`, also drops `.`, `!`, `?` and spaces at the end, and one word of letters that is not among the options is `not_yet`; `normal.Word` recognizes it: a letter of the category L, then letters and marks M. `task.choice-key` requires of the key one word, its end as written, and `task.choice-options` no option that normalization would leave empty;
  - **the picture paths of a template** hold placeholders: an instance substitutes the numbers into all the strings, the path too, `task.known-pictures` checks the path of each instance, and in a task that is not a template a brace in a path is a breach of `task.placeholders`, as is a placeholder of a solver's result in any path, the path of a picture of the solution too: `takesResults` lets a result only into the text of the solution. The picture path of an instance is among the texts of `task.no-leaks`, the path of a task without a template is not;
  - **the leak check** also searches the leading question `leading_question`, in the file of the task; the title and the description of each solving step, a text of the whole module, for each task; the texts of the numbered list of a lesson; and in a worked example on a workbook the leak phrases in the texts of its lines, but not the keys, and the line of the example that solves a line of the task: the same pieces in the same order, the texts in the normal form, the value of each box equal to the key, as the check of the box compares it. The normal form of the search is `normal.Search`: `leak` searches by it, the module check compares the texts of lines by it too, and `leak` gained no exported name;
  - **version comparison** — for a choice the meaning is held by `notation` too — `options` without it —, and by the options in the normal form of that notation; for a time, by the key and the traps by hours and minutes. The leading question and the description of a step are a patch, by the general rule "anything else";
  - **the vector test** knows the kind `pages/` — `pages_test.go` —, the verdict kind `time`, `page.yaml` as a file that is no case, and `page-vector.schema.json` of the page vectors.
- **`page`** — `Lay` lays out the item lines by the numbers and the width table of `page.yaml` and gives the places of the number, the texts and the boxes, or `*NoRoom` with the item whose row has no room; `TextWidth` and `Fits` give the width of a text and whether a value fits in a box. On it stand `task.workbook-page`, on a task without a template and on each instance, and `lesson.workbook-page` with `lesson.example-fills`, in `module/workbook.go`. The message about a lack of room is one: with the numbers of `page.yaml` the rows end at about the twentieth item, long before `most_items`, 99. `page/page.yaml` is in the list `paths` of the release, since the binary embeds it.

— *Why:* while the format is v0, the platform reads only its current version. In versions 0.y of semver an incompatible change is carried by the minor number (sources 5.2), and a release that refuses every module of the previous version is an incompatible change for `modules`, `canvas`, `api` and the eval runner of `tutor`. `page` stands apart because others need it too — `api`, the check of instances —, and its own copy of `page.yaml` keeps a consumer from carrying every file of the format, which `spec.Format` embeds.

— *Rejected:*
- staying on the line of the previous version: its next patch would refuse the modules that its previous patch reads — an incompatible change in a patch;
- an epic of tasks for a version of the format: the vector test without skips holds anyway that the whole format is in the code, and there are no skips to merge in parts;
- the page layout in `module`: the package is needed by others too, and holds its own copy of `page.yaml` and its own vectors, as `picture` holds its decoder;
- taking `page.yaml` from `spec.Format` at run time: the package `spec` embeds all the files of `format/`, and `api` would carry them;
- a window of two versions at once: while the format is v0, the platform reads only the current version.

— *Consequence:* `module/`; `page/`; `internal/normal/`; `internal/yaml/`; `verdict/`; `leak/`; `template/`; `moduleversion/`; `internal/schema/schemas/`; `internal/vectors/`; `.github/workflows/release.yml`; [architecture.md](architecture.md).

## A move to a tag that changes nothing the code reads

**R53. A move to a tag of `spec` is made though the format, the code and the copies of the format's files stay the same, and its merge gives a release of the line, of the same code: `go.mod` and `go.sum` are in the list `paths`.**

— *Why:* this repository reads `spec` at the tag its submodule is on, and the vector test on the new tag holds that the library passes its vectors and the format's examples as the tag has them.

— *Rejected:* staying on the older tag because the format is the same: the sessions here would read an older `spec`, and an example the tag changed would go unchecked.

— *Consequence:* `spec/`; `go.mod`, `go.sum`.

## spec's own rules kept out of the sessions

**R54. `.claude/settings.json` keeps spec's own CLAUDE.md files and its rules of parts, `spec/.claude/rules/`, out of this repository's sessions by `claudeMdExcludes`, under the path of a clone in a folder named `module-kit`.**

— *Why:* Claude Code loads a rule of a part into a session as soon as the session reads a file of that part, in `spec/` too, and spec's rules speak of spec's recipes, not of `-f spec/justfile`; spec's CLAUDE.md is spec's working rules, not this repository's.

— *Rejected:* letting spec's rules in: a session here would act on recipes that run from spec's root.

— *Consequence:* `.claude/settings.json`; CLAUDE.md.

## The settings of sessions, the rules of parts and the budgets

**R55. `.claude/settings.json` is the template's file of the kind `go`, byte for byte: it compacts a session at 400 thousand tokens, fetches no claude.ai connector and runs the hook `UserPromptSubmit` from `spec/tools/session/`; the plugin `gopls-lsp` loads by `CLAUDE_CODE_PLUGIN_DIRS` of `.devcontainer/devcontainer.json` and runs the gopls of `PATH`, the one the stage `devcontainer` builds from its lock file. The checks of each part are its rule of `.claude/rules/`: `go.md` — the Go code, `go.mod`, `go.sum` and the embedded copies of the format's files, with `ci-lint`, `ci-test`, `ci-vuln` and `ci-licenses` —, and `ci.md` — CI, the release and the devcontainer. The external facts are records of `docs/sources.md`, the measurements those of `docs/measurements.md`, and `docs-budgets.json` names a limit above spec's block `repository` only for what is larger, at its size plus 10–20% (measurements 1.9 spec).**

— *Why:* a rule is loaded only when the session reads or edits a file of its part, so its checks reach the session that needs them and cost the others nothing; without them the checks stood scattered over the architecture and the comments of the justfile. The settings are the template's, so every repository's sessions run alike, and a difference would have no reason of this repository's. The template loads the plugin from the container's own path, whether or not the folder is trusted.

— *Rejected:*
- one rule `go.md`, without `ci.md`: the checks of CI and of the release would stay scattered over the architecture and the justfile;
- the plugin by a marketplace in `extraKnownMarketplaces` and `enabledPlugins`: the template no longer has them, and the plugin would hang on the `~/.claude` that the containers share.

— *Consequence:* `.claude/settings.json`; `.claude/rules/go.md`, `ci.md`; `.devcontainer/devcontainer.json`; CLAUDE.md; `docs-budgets.json`; [sources.md](sources.md), [measurements.md](measurements.md).

## The short RUN.md

**R56. RUN.md keeps only what a session acts on and what is said nowhere else, and has no tasks outside an epic and no `run/TASKS.md`: a task that does not come from an epic — a move to a tag of `spec`, an analysis of the sessions by spec's threshold — goes into table 1 of the current epic, and when there is none, it is the first task of the next and starts it. The rules keep their numbers.**

— *Why:* RUN.md is read by every task's session, so each byte of it is paid for in every session. What CLAUDE.md and RULES.md say is loaded into each session anyway, and the history of epics is in their archives. With no task outside an epic there is no case of where its body goes while no epic is current, and no file such tasks would fill.

— *Rejected:*
- renumbering the rules after cutting some: the archives cite the rules by number;
- keeping «Задачи вне эпика» empty, as in the template: one more rule and section for a case this repository does not have;
- an empty `run/TASKS.md` by the template: a file no task here fills; the template moves the bodies out of RUN.md because in `spec` they pile up between epics, and here none piles up.

— *Consequence:* RUN.md; `run/`.

## The work plan and its differences from the template

**R57. The work plan is RUN.md and `run/`: table 1, the current epic, in RUN.md, and tables 2, 3 and 4 in `run/EPICS.md`, `run/QUESTIONS.md` and `run/DEBT.md`, each opening with its rules; the rules and tables in English, terse, for the session, and the epic's description, table 1, the bodies of tasks and the archives `docs/plan/` in Russian. RUN.md differs from the template's: no `run/TASKS.md` and no «Citations»; rules 3 and 5 are a line with a pointer; «Epics» says that a task started between epics starts the next one; rule 4 names that the exception for a task's own checks in a browser or with a real call of AI does not arise, since `module-kit` calls no AI and has no browser; rule 12 names this repository's SonarCloud and CodeQL; the statuses keep their short form.**

— *Why:* the orchestrator of `spec` and `just inbox` read the plan in this layout. The session reads the rules, and the author writes and reviews the tasks and the archives, so each is in the language of its reader. Each difference from the template is what this repository does not have: no task outside an epic, no AI and no browser.

— *Rejected:* `run/TASKS.md` by the template, empty: a file no task here fills.

— *Consequence:* RUN.md; `run/`; CLAUDE.md; `.bestpractices.json`.

## The shared checks of `.github` at one release

**R58. CI of `module-kit` calls the shared checks of `.github` at one release, by the SHA of its tag with the version beside it, in every `uses:` of `.github/workflows/`, the composite action `toolchain` among them; Dependabot raises every reference to `.github` in one pull request, the group `shared-workflows`, `patterns: ["altnovatrail/.github/*"]` under `github-actions`. The input `same-bits` of the shared `go.yml` stays off. `devcontainer.yml` builds its image from the network, with no cache.**

— *Why:* the references name one release, so one pull request moves them all, the action with them: dependabot-core names it by its path, which the pattern matches (sources 3.15). The jobs of `go.yml` and `documents.yml` and "Binaries / build" take the image of `toolchain` from one entry of the cache of GitHub Actions and build it only on a miss, so an outage of what the Dockerfile downloads from does not fail the checks while the Dockerfile stays the same (sources 3.12). `same-bits` is for the libraries whose bits must be the same on every build, and `module-kit` is not one of them; "Binaries" already holds that its binaries build to the byte on six systems.

— *Rejected:*
- `same-bits: true`: four jobs more on every pull request for a property this library does not need;
- the cache in `devcontainer.yml`: what it checks is that the image builds from the network, and a cache would hide a broken download;
- a path or key of the image of its own for "Binaries / build": a second entry of the same image in the cache;
- waiting for Dependabot without the group: it would raise each reference in a pull request of its own.

— *Consequence:* `.github/workflows/`; `.github/dependabot.yml`; `.claude/rules/ci.md`; [architecture.md](architecture.md).

## The speed of CI

**R59. `release-files` runs its six cross-builds at once; `TestVectors` runs its cases as parallel subtests, but for the verdicts of a predicate, `verdicts/predicate/`, which run one at a time and apart from the rest; `TestChecksTheFormat` checks each module and each file of a part in a parallel subtest of its own. What the checks prove is the same.**
- **The builds.** Each `go build` of `release-files` is a process of its own in the background, over the one build cache, which Go shares between processes; the recipe waits for each and fails after all six have ended, printing the log of a build that failed under its platform — each build writes to a log of its own, outside `dist/` —, and a recipe that stops before the builds end stops them by its `trap`. A build compiles its own packages for its own target, so the six do not wait on each other, and the binaries are those that one build after another gives: `SHA256SUMS` is the same byte for byte. On the four cores of a runner of CI the builds at once gain nothing that the spread of runs shows, since one `go build` already keeps four cores busy; they are for a runner of more cores and for a build by hand (measurements 3.1).
- **The tests.** Every case of `TestVectors` shares nothing but the read-only file systems `spec.Format` and `contracts.Vectors` and the tables of the test, and every module and file of `TestChecksTheFormat` nothing but `spec.Format`, so each is a subtest with `t.Parallel()`; the race detector, which `ci-test` runs, would find what they shared. The verdicts of a predicate take a predicate to its million steps, which under `-race` takes a good part of the clock of a second that `sandbox.Predicate` gives it, and with other cases beside them the clock stopped one, "unchecked" where the vector gives "correct" (measurements 3.1); so `ci-test` runs them apart, after the rest.

— *Why:* on the four cores of a standard runner `go test ./...` already runs the packages of the module side by side, and what is left is the tail of the two longest packages, `module` and `internal/vectors`, each a test of one goroutine. Only what changes no proof: the same cases, the same binaries, `-race` and the coverage.

— *Rejected:*
- a longer clock for the predicates in the test: the test would pass the vectors with limits other than those the library runs with;
- `ci-test` without `-race`, or a run of the vectors outside it: what the check proves would change;
- `t.Parallel()` in `FuzzCheck` and in the other tests of `module`: `FuzzCheck` runs its corpus in 12 s under `-race` with the coverage on four cores, and is a change of its own;
- the kind `verdicts` whole one at a time: its other cases, of `choice`, `number`, `fraction` and the rest, run no predicate, and the clock is what the exception is for.

— *Consequence:* `justfile` (`release-files`, `ci-test`); `internal/vectors/vectors_test.go`; `module/format_test.go`; [architecture.md](architecture.md); [measurements.md](measurements.md).

## The tests at the clock of a predicate apart

**R60. `ci-test` runs the tests that take Starlark to the clock of a second of `sandbox.Predicate` after the rest of the module's tests, with nothing beside them: first the module but for them, then the packages `internal/sandbox` and `verdict` one at a time, `-p 1`, then the verdicts of a predicate of `TestVectors`, `format/vectors/verdicts/predicate/`, alone and with `-v`, so that the log prints each case with its time. The three profiles are one `coverage/go.out`. The recipe fails if the last run ran no case. What the check proves is the same — the same tests, under `-race`, with the same coverage —, and a plain `go test ./...` runs them as before.**
- **What is at the clock.** Under `-race` a Starlark step is several times slower, and the clock of a predicate, 1 s, is the library's. With the predicate cases alone after the rest, `TestConcurrentEval` — 32 evaluations at once —, still beside the other packages, was stopped by the clock (measurements 3.2). The tests of `internal/sandbox` and `verdict` take a predicate to its million steps too, `TestCompile` and "the limit of steps", so the packages go out whole: their tests share the package's process, and naming each test would miss the next one written. No other package evaluates a predicate near its steps.
- **How.** `CLOCK_PACKAGES` and `PREDICATE_CASES` are variables of the `justfile`: the first run is `go list ./...` without the first, with `-skip` by the second; an element of a `-run` pattern is an element of a test's name, and a case's name is its path in `spec`. A folder that `spec` renames would leave the last run without cases, and the recipe then fails rather than drop them. The later profiles join the first without their mode lines, as `go test` joins those of its packages: a block covered twice is twice in it, as with `-coverpkg=./...` already.

— *Why:* side by side with the other tests, a case at the clock can turn "unchecked" where the vector gives "correct", at random on a slower or busier runner. Running them after the rest takes the contention away from these tests alone and keeps the parallel packages for the rest; on a runner of CI the cases at the clock then take a small part of the second (measurements 3.2).

— *Rejected:*
- `ci-test` with `-p` lower: every package waits for the others, and with any `-p` above 1 the tests at the clock still have a package beside them;
- a longer clock in the tests: the tests would pass with limits other than those the library runs with;
- only the verdicts of a predicate apart: `TestConcurrentEval` then failed by the clock beside the other packages (measurements 3.2);
- `-v` for `internal/sandbox` and `verdict` too: 556 lines in the log for two tests whose times the measurement holds;
- waiting until CI fails: a failure at night stops the orchestrator's merge.

— *Consequence:* `justfile` (`ci-test`); `internal/vectors/vectors_test.go`; [.claude/rules/go.md](../.claude/rules/go.md); [architecture.md](architecture.md); [measurements.md](measurements.md).

## The Go caches of "Binaries / build"

**R61. "Binaries / build" takes the caches of Go — the modules and the builds — with the image of `toolchain`, under the check `binaries` in the key of the Go caches, a key of its own; it takes the image from the one entry that the jobs of `go.yml` and `documents.yml` read, by itself, without waiting for `go.yml`. A push to `main` alone saves the Go caches.**

— *Why:* the time of the builds goes to compiling with an empty `GOCACHE`: locally a filled cache took `release-files` from 78 s to 3 s, with the same `SHA256SUMS` (measurements 3.3). A key of its own, because the builds of six targets with `-trimpath` are those of no check of `go.yml`. By itself, because two called workflows of one run, and a job of the caller, see each other's jobs only through the caller's `needs` (sources 3.13), and `needs: go` would hold "Binaries" until the last check of `go.yml`.

— *Rejected:*
- `needs: go` and the artifact of the image of `go.yml`: "Binaries" would wait for every check of `go.yml`, and the artifact is an inner detail of `go.yml`, not its interface;
- "Binaries / build" without the Go caches: the six builds would compile with an empty `GOCACHE` on every run;
- the key of a check of `go.yml`, as `test`: another job's builds, and two saves of one key on `main`.

— *Consequence:* `.github/workflows/ci.yml`; `.claude/rules/ci.md`; [architecture.md](architecture.md).

## The stages of the Dockerfile

**R62. `.devcontainer/Dockerfile` has the template's three stages, `toolchain-core`, `toolchain` and `devcontainer`: the shared release builds its files in `toolchain-core`, and the checks run in `toolchain`, which here is the core with its `SHELL` alone, the checks without a browser and tests in the core only where `toolchain` holds more than it (sources 3.14). "Binaries / build" builds in `toolchain`, as the job `arm64` of the shared `go.yml` does, and its step «The stage of the release» fails the job once `toolchain` holds more than the core — an instruction but `SHELL`, or a `FROM` of another base —, or if awk cannot read the file; its awk is that of the step «Toolchain stages» of the job `toolchain` of `go.yml`, to which «As in the pinned .github» holds it. A tool that only the tests need goes into `toolchain`; one that a release or `ci-lint`, `ci-vuln` and `ci-licenses` run, into the core.**

— *Why:* the release of `module-kit` gains no time by the core, since its `toolchain` adds nothing to it, but the stages are the template's, which a move to a tag takes byte for byte, and a tool added later goes to the stage that keeps the release and the checks without tests small. "Binaries / build" builds the release's files to prove the binaries of each system equal: in `toolchain` it takes the entry of the image that `go.yml` fills on every run, and the two stages hold the same files while `toolchain` adds nothing, which the step holds.

— *Rejected:*
- "Binaries / build" in `toolchain-core`: the shared checks keep no entry of the core for a caller whose `toolchain` adds nothing to it (sources 3.14), so the job would build the core on every pull request, or keep an entry of 382 MiB of its own (measurements 3.1 spec) for the same files; the step gives the same proof while the stages are one, and fails once they are not;
- no such step: a tool added to `toolchain` that `release-files` came to need would pass here and fail the release on `main`;
- a reading of the stages of its own, not held to `go.yml`: a change of that reading in `.github` would leave the two disagreeing on whether the core is a stage of its own;
- the image built with `STAGES: toolchain-core` under the key of `toolchain`: a file of one stage read as the other.

— *Consequence:* `.devcontainer/Dockerfile`; `.github/workflows/ci.yml`; `.claude/rules/ci.md`; [architecture.md](architecture.md).

## The container kept running when its window closes

**R63. `.devcontainer/devcontainer.json` sets `"shutdownAction": "none"`, as the template's does, so that closing the window of VS Code leaves the container running under an executor of the orchestrator; the default for a Dockerfile is `stopContainer` (sources 9.4 spec). A change of `.devcontainer/Dockerfile` gives a release of the same code: the list `paths` of the shared `release.yml` names files, not lines, and the Dockerfile is in it because the release builds its files in its stage `toolchain-core`.**

— *Why:* a container that stops when a window closes loses a running executor's attempt, which the next run pays for again. The file is the template's, taken whole, so that the next move again compares byte for byte. A release of the same code costs a run of the shared workflow and nothing else.

— *Rejected:*
- keeping `stopContainer` as a difference of this repository: nothing here needs it, and the orchestrator runs here too;
- taking the tag without `.devcontainer/`: the copy would no longer be the template's, which `.claude/rules/ci.md` holds it to;
- keeping a change of the Dockerfile from starting a release: `paths` names files, and the Dockerfile is what the release is built in.

— *Consequence:* `.devcontainer/devcontainer.json`; `.github/workflows/release.yml`.

## The composite action `toolchain` of `.github`

**R64. "Binaries / build" takes the image of the stage `toolchain` and the caches of Go by the composite action of `.github`, `altnovatrail/.github/.github/actions/toolchain`, at the SHA of the shared workflows, called twice with the check `binaries`: «Toolchain image and Go caches» before «Release files and samples», and «Go caches into the cache» with `save: "true"` after the upload of the artifact. Dependabot moves the action with the shared workflows, in the group `shared-workflows`. The step «As in the pinned .github» holds what the job still has of `.github`'s: every reference to it at one SHA, the mount of the Go caches to the lines of the action's header, and the awk of «The stage of the release» to `go.yml`.**
- **The key** of the Go caches is `go-<OS>-<architecture>-binaries-…`, so the entries saved on `main` serve on. «Release files and samples» mounts the folder and points Go into it by the lines the action's header names.
- **«As in the pinned .github»** fails the job, naming the reason, if the references to `altnovatrail/.github` in `.github/workflows/` are not at one SHA, so that the action is `go.yml`'s of the same release, with its file, image, folder and keys; if a line of the action's header at that SHA that names `/go-cache` — the three variables and the mount, which the action fixes for a caller's `docker run` — is not a line of «Release files and samples», indentation aside; and if the awk of «The stage of the release» is not that of «Toolchain stages» of the job `toolchain` of `go.yml` at that SHA.
- **`.github/dependabot.yml`** — the comment of the group `shared-workflows` names the action; its pattern `altnovatrail/.github/*` matches the action's name in dependabot-core (sources 3.15), so the group needs no pattern of its own for it.

— *Why:* with a copy of the steps of `go.yml` in the job, each move of `.github` that changed them turned "Binaries", a required check, red until a session edited the copy, and a merge by green checks waited for it. With the action, the steps are one file in `.github`, which the SHA that Dependabot raises brings, and a pull request of Dependabot that moves `.github` passes untouched unless it changes the mount or the reading of the stages.

— *Rejected:*
- dropping the guard whole: a reading of the stages of its own not held to `go.yml` would leave the two disagreeing on whether the core is a stage of its own; a mount that the action no longer fills would build cold and save an empty folder on `main` with "Binaries" green; and a reference to the action at another release than `go.yml` would read or save entries `go.yml` no longer names. A pull request of Dependabot that changes none of the three passes untouched; one that does fails, and takes the change by hand;
- the action with `check: arm64` or another check of `go.yml`: the entry of another job's builds;
- the save right after «Release files and samples»: the job would save its caches before `binaries.sh` proved the binaries; the save stands after the artifact;
- holding the mount of «Release files and samples» to the step `ci-test` of the job `arm64` of `go.yml`: the action's header is the interface it names for a caller, and `go.yml`'s own recipes are its inner detail;
- the mount taken by the job from the action by an output: the action has none for it (sources 3.15).

— *Consequence:* `.github/workflows/ci.yml`; `.github/dependabot.yml`; `.claude/rules/ci.md`; [architecture.md](architecture.md).

## README without its status, its license and the tag of `spec`

**R65. README.md has no section «Status» and no section «License», no sentence on the languages of the documents and the plan, and its links to `spec` name no tag: the platform and the submodule lead to `https://github.com/altnovatrail/spec`, and spec's RULES.md to its `main`. The badge of the license stays in the header. A task, a move to a tag of `spec` among them, does not edit README for a status or a tag.**

— *Why:* the status and the tag went out of date with every task and every move to a tag, and each move spent an edit on them; the license is named by LICENSE, NOTICE and the badge, and the languages by spec's rules.

— *Rejected:*
- the link to RULES.md at a tag: a tag in it goes out of date as the status did;
- moving the status into [architecture.md](architecture.md): each fact of the status already stands where it is decided or built.

— *Consequence:* README.md.

## The reading mode `single`

**R66. The reading mode `single`, in which one reader reads a task that is not an essay and its observation goes to the verdict as it is, needs no function of its own: `Verdicts` gives the verdict on one reader's valid observation and refuses one that is not valid with `ErrNotValid`, so `api` calls it with no `Join`. The documents and the comments of `observation` name the mode. `TestVerdicts` gives verdicts on one reader's observations, none of them joined, and `TestVerdictsErrors` refuses one that is not valid; spec's vectors of `observation/v1/` hold no case of `single` yet, Д22.**

— *Why:* the mode is a rule of the [contracts text](../spec/contracts/README.md), which this library implements, so its documents name how `api` calls it in that mode.

— *Rejected:*
- a function of its own for the mode `single`: `Verdicts` is it already, and a second name would be two ways to one verdict;
- a test case of `single` of its own here before spec's vector: `TestVerdicts` holds the same verdicts on one reader's observations, and `TestVerdictsErrors` refuses one that is not valid with `ErrNotValid`, after which `nil` gives `unchecked`, as the case "no observation" of `TestVerdicts` holds; the case belongs to the vectors;
- leaving the missing vector to spec's table alone: the work it returns to is this repository's runner of the vectors, which reads a vector strictly and may need code.

— *Consequence:* `observation/doc.go`, `observation/observation.go`; `run/DEBT.md`; [architecture.md](architecture.md).

## README without commands

**R67. README.md holds only what the component is and its place in NovaTrail, under its title alone: the header and its badges, what the library and the CLI do, and, as paragraphs and a list, that the format is spec's, who takes the library and the CLI and how, its card in spec's architecture and a link to `docs/`. It holds no heading below its title, no command and no block of code, no section «Releases», «Contributing» or «Documents», and no status, stage, version, date or language. [README.md](README.md) of `docs/` holds the table of contents of `docs/` and of the plan, and the sections «Releases» and «Contributing», which CLAUDE.md names; «Contributing» asks for issues in English. The answers of `.bestpractices.json` on how to take a release, contribute and run the checks point to it, and `english_justification` asks for issues and pull requests in English.**

— *Why:* README is read by a person on GitHub, and nobody runs its commands by hand: agents do, from the documents they read, and the detail went out of date in README.

— *Rejected:*
- a short «Releases» of two sentences and a block of commands: commands that nobody runs by hand, in a text a person reads;
- the checks in a document of their own, `docs/build.md`: the list is four lines, and «Contributing» is where a contributor looks for them;
- the detail of releases in [architecture.md](architecture.md): it holds how CI builds a release, not how one takes it.

— *Consequence:* README.md; [README.md](README.md); CLAUDE.md; `.bestpractices.json`.

## The Go caches by a prefix and the template's comments

**R68. CI calls the shared checks of `.github` at a release whose Go caches, on a miss of the exact key, come by its prefix without the hash of `go.sum`, their modules dropped and their builds kept, in every `uses:` of `.github/workflows/` and in the composite action `toolchain` (sources 3.16): `go mod verify` runs on an exact hit alone, and the builds of an entry by the prefix are held by the scope of `main`, as those of a hit. A move to a tag of `spec` takes `.devcontainer/` from the template but for a change of the template's comments alone, which waits for the next change of `.devcontainer/` beyond comments and comes with it.**

— *Why:* every move to a tag of `spec` changes `go.sum` here, and with the exact key alone the Go caches were cold after each. A change of the Dockerfile changes the key of the image of `toolchain` and of every Go cache and asks for a rebuild of the container, which a comment does not repay.

— *Rejected:* the template's comment of the Dockerfile taken with the move: a rebuild of the container and cold caches for a comment.

— *Consequence:* `.github/workflows/ci.yml`; `.devcontainer/`; `.claude/rules/ci.md`; [architecture.md](architecture.md).

## The key of a task from paper

**R69. The key of a task from paper is the package `paper`: the solver of a formula, with a parser of its own and exact rationals of `math/big`; the rules of an answer of AI; the key with its provenance; and the task's one item and the values read brought to the key, by which the package `observation` joins two readings and gives the verdict, as on a task of a module. It passes the vectors `paper/v1/` of `contracts.Vectors`, by the rules of the [contracts text](../spec/contracts/README.md).**

- **`Solve(formula)`** — `*paperv1.Number`, its value and the letter of an equation's unknown, or the `Reason` it is not taken. The reasons are those of the contracts text, the first that applies: the grammar, then the letters — two unknowns, an unknown in an expression, an equation with none —, then the operations as they are evaluated, each part before its operation and the left before the right — not linear, a division by zero, a part of more than 100 digits, the two sides of an equation brought together too —, then no root and every root, and last an answer longer than 16 characters. Each part is linear, a·x + b, its coefficients `*big.Rat`; the answer is an integer or p/q in lowest terms.
- **`Breach(answer)`** — the first rule an answer of AI breaks, or none: no kind, a value not in the notation of `number`, units empty or the same after the steps of units, an unknown not one letter, the count and length of options, options the same as a choice normalizes them, a key that is no option.
- **`Key(solved, ofAI)`** — one task's key, a `TaskKey` whose statement `api` sets: the solver's answer with `PROVENANCE_SOLVER` when there is one, or else AI's with `PROVENANCE_AI` when it breaks no rule, and with neither no answer and no provenance, as `key.proto` has it: a key with no answer says the task has none, its item unchecked at `api`, and `CAUSE_NO_KEY` when no task has one.
- **`Items(answer)`** — one item with one answer slot bound `anywhere`, its check `number` with the key and units, or `choice` with the options, the key and `notation: options`, and no trap values: what `observation.Validate`, `Join` and `Verdicts` take as they take a module's. An answer of no kind, a task with no key, gives no item: no check can be written of it, and its item is unchecked at the caller.
- **`Brought(answer, o)` and `BroughtValue`** — a copy of an observation whose values are brought to the key: a letter and `=` before the value are dropped when the letter, after steps 1 to 4, is the key's unknown, and the brackets round a label of a choice — `(C)`, `[C]`, `C)`. Any other value stays as it was read.
- **The vectors** — `internal/vectors/paper_test.go`: three kinds, `paper/v1/solver`, `paper/v1/answers` and `paper/v1/verdicts`, one per file, so that a file gone is a kind without a case. Each is read strictly; a value of the verdicts goes into the one answer slot, `null` as `UNREADABLE_BLANK`, through `Brought`, `Validate` and `Verdicts`. **Tests** of `paper` itself: `TestKey`, `TestItemsOfNoKey`, `TestBreachOrder`, `TestSolveBounds`, `TestJoinOfBroughtValues`, `TestBroughtCopies`, `TestParse`, and fuzz tests, the formula and the value read being foreign input: `FuzzSolve` — an answer or a reason, never both, and an answer is one `Breach` takes — and `FuzzBroughtValue`.

— *Why:*
- exact rationals: `1/3+1/7` is `10/21`, and a key that a float rounds would give a correct answer "not yet". The bound of 100 digits on the way keeps `(9^10)^10` from growing without end;
- a parser of its own, not the Starlark sandbox: the grammar is narrow, ASCII, and the solver must tell a linear equation from one that is not, which an evaluator of Starlark does not say; it needs no clock;
- the verdict through `observation`: the readers of a task from paper read as those of a module, and the join, the reasons of "unreadable" and the verdict are one code. Bringing the values first makes `x = 7` and `7` one value, so two readers who wrote them differently join;
- the item is built from the key, not kept: `api` keeps `paperv1.Key` on the attempt, and the verdict repeats from it and the observation.

— *Rejected:*
- `float64`: it loses exactness;
- the solver in Starlark: its sandbox brings a clock and limits that a formula of 200 bytes does not need, and no answer of linearity;
- a check kind of its own for paper: `number` and `choice` already compare, and a new kind is a new version of the format;
- bringing the values in `verdict`: that is the normalization of the [format text](../spec/format/README.md), and the letter of the unknown and the brackets of a label are rules of the contracts text for a task from paper alone.

— *Consequence:* `paper/`; `internal/vectors/paper_test.go`; [architecture.md](architecture.md).
