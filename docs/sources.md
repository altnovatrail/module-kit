# module-kit — sources

> The facts of the outside world the repository's documents under `docs/` rest on and no run of ours checks: a law, a regulator, a standard's normative text, a provider's terms and policies, a price, a license, a page that changes in place with no version, a text outside the repositories, and other projects' practice. Here are their records — the address, the date it was read and a quote in the source's language — and [architecture.md](architecture.md) and the decision log hold the claim and a citation of its one record by its number, `sources 1.1`, not a chain. A record's heading names its topic, not a task.
>
> No record is needed for a library's documentation or code at a version or commit — a link beside the claim —, for a behaviour a test or a measurement checks — they are its evidence —, or for spec's tags and the files of NovaTrail's and MathTrail's repositories — git, at a commit.
>
> A record is a subsection "N.M" of a topic: one source, or one reading of several sources for one decision. A record's number never changes and is never given again: a new one goes at the end of its topic, a new topic at the end of the file, and a record removed leaves a gap. A source read again with the same text changes the record's date; with another text, the claims and decisions on it are rewritten in the same change and the record holds the new reading alone — where a decision waits on the author, an open question, and both readings stand until it is closed. The model is `spec/docs/sources.md`. The results of measurements are not here but in [measurements.md](measurements.md).

## 1. MathTrail standalone

Its files are git: a claim links to [mathtrail/mathtrail-standalone](https://github.com/mathtrail/mathtrail-standalone) at its commit.

## 2. Go and its toolchain

Go's documentation and sources at the version `.devcontainer/Dockerfile` pins are links beside the claims, and `go help` of that Go; what remains here is a promise of Go's no run checks.

### 2.6. Go's promise about the streams

The Go blog, ["Evolving the Go Standard Library with math/rand/v2"](https://go.dev/blog/randv2), 1 May 2024, read 2026-11-02. On `math/rand`: "At that point, we decided Go 1 compatibility included the specific random outputs for a given seed and added a test". On `math/rand/v2` — after `IntN` and the related functions got Lemire's method: "The concrete rand.Rand API is now locked in to that value stream, so we will not be able to take advantage of any optimizations yet to be discovered".

## 3. Checks, CI and release

The files of `.github` and of other NovaTrail repositories are git, at the release or commit a claim links to; what remains here is licenses, terms and the answers of services that no run of ours checks.

### 3.1. Tools

The licenses of the tools of the checks, the GitHub API, `repos/<owner>/<repository>/license`, read 2026-10-21: golang/vuln — BSD-3-Clause, google/go-licenses — Apache-2.0, golangci/golangci-lint — GPL-3.0.

### 3.4. SonarCloud, docker:S8545

The rule "Go dependencies should be locked to verified versions" (the SonarCloud API, `api/rules/show?key=docker:S8545`, 2026-10-21). On pull request #8 it flagged the line `go install` of gopls in the Dockerfile: "Dependency versions are not predictable. Use a lock-file enforcing command instead". In canvas-api such findings are open on two lines of its Dockerfile, 198 and 199, although `-mod=readonly` stands there (the SonarCloud API, the project MathTrail_canvas-api, 2026-10-21).

### 3.6. Licenses in the release

The two licenses:
- Apache-2.0 — [LICENSE](../LICENSE), section 4 (a): "You must give any other recipients of the Work or Derivative Works a copy of this License";
- Go's license — `LICENSE` at the root of Go 1.27.4 of the stage `toolchain`, `$(go env GOROOT)/LICENSE`, read 2026-11-24: "Redistributions in binary form must reproduce the above copyright notice, this list of conditions and the following disclaimer in the documentation and/or other materials provided with the distribution". A Go binary carries the standard library.

### 3.10. gopls from a lock file

Read 2026-11-04: `tutor`, which builds gopls from a lock file, has no findings of docker:S8545 (the SonarCloud API, `api/issues/search?componentKeys=altnovatrail_tutor&rules=docker:S8545`), and its Scorecard Pinned-Dependencies flags in the Dockerfile only `npmCommand not pinned by hash` (the GitHub API, `code-scanning/alerts`); the license of golang/tools is BSD-3-Clause (the GitHub API, `repos/golang/tools/license`).

### 3.11. When Dependabot proposes a version

GitHub Docs, read 2026-11-04: ["Configuring Dependabot version updates"](https://docs.github.com/en/code-security/dependabot/dependabot-version-updates/configuring-dependabot-version-updates) — "When this file is checked in, Dependabot checks the manifest files on the default branch for outdated dependencies"; ["Dependabot options reference"](https://docs.github.com/en/code-security/dependabot/working-with-dependabot/dependabot-options-reference), `cooldown` — "A new version is not considered for a version update until 3 days after its release", and from what instant and how the days are counted it does not say.

### 3.17. GitHub-hosted runners for the binaries

GitHub Docs, "GitHub Actions billing", read 2026-11-05: "Standard GitHub-hosted runners in public repositories" are among what "remains free".

## 4. Go libraries

The code and documentation of a library at a version or commit are links beside the claims; its license `ci-licenses` checks, its vulnerabilities `ci-vuln`. What remains here is the state of issues and pull requests, reports that change in place, and the libraries not taken.

### 4.1. JCS libraries for Go

The Go module proxy and the GitHub API, read 2026-11-02:
- `github.com/gowebpki/jcs` — `@latest` v1.0.2 of 2026-09-21, the license Apache-2.0. In `go.mod` — `require github.com/stretchr/testify v1.12.1`. Issue #5 "RFC 8785: reversed surrogate pair becomes U+FFFD and invalid UTF-8 passes through" was closed 2026-09-20;
- `github.com/cyberphone/json-canonicalization` — there are no version tags, `@latest` is the pseudo-version `v0.0.0-20261011152517-32a48195b9b2`. The license by the GitHub API is `NOASSERTION`, and the repository has no `go.mod`. Issue #33 "Go import (as documented) broken" has been open since 2025-03-11.

### 4.5. The JSON Schema validator

Read 2026-11-04:
- Bowtie, [bowtie.report](https://bowtie.report), the run of 2026-11-04 06:21 UTC: on draft 2020-12, go-jsonschema (v6.0.2), `github.com/santhosh-tekuri/jsonschema/v6`, passes 1300 tests of 1300;
- issue [#276](https://github.com/santhosh-tekuri/jsonschema/issues/276) "propertyNames and contentSchema errors: InstanceLocation is overwritten by later siblings" has been open since 2026-09-24; pull request #278 was closed without a merge on 2026-10-02, "Closing in favor of earlier community PR #269", and #269 is open;
- other libraries: `github.com/kaptinlin/jsonschema` — `v0.9.10` of 2026-09-13, MIT, Bowtie does not check it; `github.com/google/jsonschema-go` — `v0.4.3` of 2026-04-17, MIT, validation stops at the first error; Corvus Go (`corvus-dotnet/Corvus.JsonSchema`, `src-go/`) — v0.1.4 of 2026-11-03, the first release v0.1.0 of 2026-10-08, Apache-2.0, patterns by ECMA-262; `xeipuuv/gojsonschema` v1.2.0 — draft-04, -06 and -07; `qri-io/jsonschema` v0.2.1 — draft 2019-09.

### 4.11. go-yaml: an anchor whose name ends with a colon

yaml/go-yaml#109, "Anchor map key with no space before colon works in v3; breaks in v4", read 2026-11-06: closed on 2025-11-15 as not planned.

## 5. Standards and specifications

### 5.1. Semantic Versioning 2.0.0

[semver.org/spec/v2.0.0.html](https://semver.org/spec/v2.0.0.html), read 2026-11-02:
- item 3: "Once a versioned package has been released, the contents of that version MUST NOT be modified";
- item 4: "Major version zero (0.y.z) is for initial development. Anything MAY change at any time";
- item 8: "Major version X (X.y.z | X > 0) MUST be incremented if any backward incompatible changes are introduced to the public API. It MAY also include minor and patch level changes" — the author may raise the version part that the comparison found to a higher one.

### 5.2. Cargo on 0.y.z

[The Cargo Book, "SemVer Compatibility"](https://doc.rust-lang.org/cargo/reference/semver.html), read 2026-11-02: "Initial development releases starting with "0.y.z" can treat changes in "y" as a major release, and "z" as a minor release. "0.0.z" releases are always major changes. This is because Cargo uses the convention that only changes in the left-most non-zero component are considered incompatible". The same in oasdiff: "Below `1.0.0`, semver gives the minor version the major's role" ([`docs/VERSIONING.md`](https://github.com/oasdiff/oasdiff/blob/2daca9b/docs/VERSIONING.md), commit 2daca9b, read 2026-11-02).

### 5.3. Tools that check semver by the content

Read 2026-11-02:
- gorelease — [pkg.go.dev/golang.org/x/exp/cmd/gorelease](https://pkg.go.dev/golang.org/x/exp/cmd/gorelease): "If no version is proposed with -version, gorelease suggests the lowest version consistent with semantic versioning". It finds the lowest part and proposes a version;
- apidiff — [pkg.go.dev/golang.org/x/exp/apidiff](https://pkg.go.dev/golang.org/x/exp/apidiff): "No tool can detect behavioral changes; and even if it could, whether a behavioral change is a breaking change or not depends on many factors, such as whether it closes a security hole or fixes a bug". The behavior of a module is the check of a task, and the comparison sees it;
- cargo-semver-checks, the license Apache-2.0 — [README](https://github.com/obi1kenobi/cargo-semver-checks/blob/161450e/README.md), commit 161450e: "By default, `cargo-semver-checks` uses crates.io to look up the previous version of the crate, which is used as the baseline for semver-checking the current version of the crate", and the previous version can also be given as a folder, `--baseline-root`. And: "A design goal of `cargo-semver-checks` is to not have false positives". It requires no less than what it found, and the author sets the version themselves;
- oasdiff — the same `docs/VERSIONING.md`: "a breaking change requires a major version increase", and under "What is not checked" — "Which version bump a non-breaking change deserves. Only breaking changes are checked";
- `elm bump` — [`terminal/src/Main.hs`](https://github.com/elm/compiler/blob/1c278bf/terminal/src/Main.hs) of elm/compiler, commit 1c278bf: "so there cannot be MAJOR changes hiding in PATCH releases in Elm!"; the rule — [`builder/src/Deps/Diff.hs`](https://github.com/elm/compiler/blob/1c278bf/builder/src/Deps/Diff.hs), lines 308–312: `addMag = if null added then M.PATCH else M.MINOR`, `removeMag = if null removed then M.PATCH else M.MAJOR` and the largest of the parts. A removal is major, an addition is minor, and Elm requires exactly the version it found;
- a version of data — the recipe Frictionless Data Package Version, [specs.frictionlessdata.io/patterns](https://specs.frictionlessdata.io/patterns/): major — "Combine, split, delete or change the meaning of data that is referenced by another data resource", patch — "Correct errors in existing data". Here the author declares the part, and there is no tool;
- [Conventional Commits 1.0.0](https://www.conventionalcommits.org/en/v1.0.0/): "fix: a commit of the type fix patches a bug in your codebase (this correlates with PATCH in Semantic Versioning)". The author declares the part — by the type of the commit, — and the tool does not look at the content.

### 5.4. YAML 1.2.2

[yaml.org/spec/1.2.2](https://yaml.org/spec/1.2.2/), read 2026-11-04, and 6.9.2 on 2026-11-23:
- 5.1: "To ensure JSON compatibility, YAML processors must allow all non-C0 characters inside quoted scalars";
- 5.2: "Byte order marks may appear at the start of any document" and "To allow for JSON compatibility, byte order marks are also allowed inside quoted scalars"; `nb-char ::= c-printable - b-char - c-byte-order-mark`;
- 6.1: "To maintain portability, tab characters must not be used in indentation";
- 6.3: `s-flow-line-prefix(n) ::= s-indent(n) s-separate-in-line?`;
- 7.4.2: "the ":" indicator must appear at most 1024 Unicode characters beyond the start of the key";
- 8.1.1.1: "It is an error for any of the leading empty lines to contain more spaces than the first non-empty line";
- 8.2.2: implicit keys "are limited to a single line and must not span more than 1024 Unicode characters";
- 6.9.2 "Node Anchors": `[91] ns-anchor-char ::= ns-char - c-flow-indicator` and `[92] ns-anchor-name ::= ns-anchor-char+`, so `:` is a character of an anchor's name, and the name of `&a:` is `a:`.

### 5.5. The YAML test suite

[yaml/yaml-test-suite](https://github.com/yaml/yaml-test-suite), the data release `data-2022-01-17`, commit 5f49729, the one go-yaml v4 takes too; the license is MIT, the file `License` at the tag `v2022-01-17`, commit 6d48918 (the GitHub API, read 2026-11-04).

### 5.6. The signatures of media files

[WHATWG MIME Sniffing](https://mimesniff.spec.whatwg.org/), read 2026-11-04: 6.2.3 "Signature for MP3 without ID3", "To match an mp3 header", step 1 — "If length is less than 4, return false".

### 5.7. Unicode for normalization

[UAX #15](https://www.unicode.org/reports/tr15/), section 13 "Stream-Safe Text Format", version 18.0.0, revision 58, of 2026-08-12, read 2026-11-04: "The Stream-Safe Text Process is the process of producing a Unicode string in Stream-Safe Text Format by processing that string from start to finish, inserting U+034F COMBINING GRAPHEME JOINER (CGJ) within long sequences of non-starters" and "The Stream-Safe Text Process preserves all of the four normalization forms defined in this annex (NFC, NFD, NFKC, NFKD). However, normalization and the Stream-Safe Text Process do not commute".

### 5.8. Strings in proto3 and ProtoJSON

Read 2026-11-05:
- [Language Guide (proto 3)](https://protobuf.dev/programming-guides/proto3/), the table "Scalar Value Types", `string`: "A string must always contain UTF-8 encoded or 7-bit ASCII text, and cannot be longer than 2³²";
- [ProtoJSON Format](https://protobuf.dev/programming-guides/json/), the row `enum`: "The name of the enum value as specified in proto is used. Parsers accept both enum names and integer values".

### 5.9. RFC 9649, WebP

[RFC 9649](https://www.rfc-editor.org/rfc/rfc9649.html), "WebP Image Format", November 2024, read 2026-11-23: of the lossless bitstream, "bits of each byte are read in least-significant-bit-first order"; of the flag Animation of the chunk VP8X, "Set if this is an animated image. Data in 'ANIM' and 'ANMF' Chunks should be used to control the animation."

## 6. The second implementation of `spec`

It is spec's `tools/format-check/`, git at a tag of spec: a claim names its file in `spec/` or links to the tag.
