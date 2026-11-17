# module-kit

[![CI](https://github.com/altnovatrail/module-kit/actions/workflows/ci.yml/badge.svg)](https://github.com/altnovatrail/module-kit/actions/workflows/ci.yml)
[![Release](https://github.com/altnovatrail/module-kit/actions/workflows/release.yml/badge.svg)](https://github.com/altnovatrail/module-kit/actions/workflows/release.yml)
[![CodeQL](https://github.com/altnovatrail/module-kit/actions/workflows/codeql.yml/badge.svg)](https://github.com/altnovatrail/module-kit/actions/workflows/codeql.yml)
[![Latest release](https://img.shields.io/github/v/release/altnovatrail/module-kit)](https://github.com/altnovatrail/module-kit/releases/latest)
[![Go version](https://img.shields.io/github/go-mod/go-version/altnovatrail/module-kit)](go.mod)
[![Coverage](https://codecov.io/gh/altnovatrail/module-kit/branch/main/graph/badge.svg)](https://codecov.io/gh/altnovatrail/module-kit)
[![OpenSSF Best Practices](https://www.bestpractices.dev/projects/16385/badge)](https://www.bestpractices.dev/projects/16385)
[![OpenSSF Scorecard](https://api.scorecard.dev/projects/github.com/altnovatrail/module-kit/badge)](https://scorecard.dev/viewer/?uri=github.com/altnovatrail/module-kit)
[![License](https://img.shields.io/github/license/altnovatrail/module-kit)](LICENSE)

[![Code Smells](https://sonarcloud.io/api/project_badges/measure?project=altnovatrail_module-kit&metric=code_smells)](https://sonarcloud.io/summary/new_code?id=altnovatrail_module-kit)
[![Reliability Rating](https://sonarcloud.io/api/project_badges/measure?project=altnovatrail_module-kit&metric=reliability_rating)](https://sonarcloud.io/summary/new_code?id=altnovatrail_module-kit)
[![Security Rating](https://sonarcloud.io/api/project_badges/measure?project=altnovatrail_module-kit&metric=security_rating)](https://sonarcloud.io/summary/new_code?id=altnovatrail_module-kit)
[![Maintainability Rating](https://sonarcloud.io/api/project_badges/measure?project=altnovatrail_module-kit&metric=sqale_rating)](https://sonarcloud.io/summary/new_code?id=altnovatrail_module-kit)
[![Vulnerabilities](https://sonarcloud.io/api/project_badges/measure?project=altnovatrail_module-kit&metric=vulnerabilities)](https://sonarcloud.io/summary/new_code?id=altnovatrail_module-kit)

The NovaTrail module format in code: the one Go library that holds the format's semantics, and the authors' CLI on top of it. It reads an author's module and checks it — the schema, a skill graph without cycles, references, the shared taxonomy, keys by solver and fallback texts by leak check — builds the module package and the module index, and compares a new version of a module with every published one by the format's rules of versions. It also fills seeded task templates and runs their solvers in a Starlark sandbox, and for the platform's service it builds the open view of a task, joins two readings of a learner's work into one observation and runs the check kinds on it, admits a verdict only by the provenance of its key and trusts a package by the chain of digests from a pinned index; signatures it leaves to spec's `release-verify`. It never calls an AI model: the library makes no network calls, the CLI only reads and writes a module's files, and neither holds learners' data.

A component of [NovaTrail](https://github.com/altnovatrail/spec), a universal open platform for adaptive learning, interactive lessons and AI tutoring. It belongs to stage 1, Format and core, with spec, modules, learner-model, engine and sim; together with modules it tries the module format v0 on two unlike subjects at once.

## Releases

A release comes out by itself once the checks pass on `main`: the next patch of the current line, tagged `vX.Y.Z` with the release, and nobody tags by hand. The releases of the line 0.1 were steps of the code's first epic, and nobody takes them; the line 0.2, from v0.2.0, is the one others take (R16, R42). Releases are immutable, so a tag stays on its commit and no file changes once it is out.

- **The CLI** is one binary for each of linux, darwin and windows, on amd64 and arm64, `module-kit_<version>_<os>_<arch>`, beside the licenses of the modules it is built from and `SHA256SUMS`. Each file is signed with cosign by NovaTrail's shared release workflow, its bundle `<file>.sigstore.json` beside it. Take a binary at an exact tag, and check it before its first run with spec's `just -f spec/justfile release-verify module-kit <tag> <file>`, which checks the signature against Sigstore's trust root and everything its certificate names (R77 spec). `module-kit` with no arguments prints how to call each command, and `module-kit --version` its release.
- **The library** is the Go module `github.com/altnovatrail/module-kit`, at the tag of a release: `go get github.com/altnovatrail/module-kit@v0.2.1`. Its packages are documented by their doc comments: [pkg.go.dev](https://pkg.go.dev/github.com/altnovatrail/module-kit).

## Contributing

Report a bug or propose a change in an [issue](https://github.com/altnovatrail/module-kit/issues), in English or Russian, and open one before a pull request unless the pull request fixes a typo or a broken link. How a change comes in, with every commit signed off, and what every change keeps to are in the organization's [CONTRIBUTING.md](https://github.com/altnovatrail/.github/blob/main/CONTRIBUTING.md); the rules every NovaTrail repository shares are in spec's [RULES.md](https://github.com/altnovatrail/spec/blob/main/RULES.md): code comes with its tests. Report a vulnerability privately, as the [security policy](https://github.com/altnovatrail/module-kit/security/policy) says, never in a public issue.

Work in the devcontainer: open the repository in VS Code and choose "Reopen in Container". It holds every tool at the version the repository pins. The checks are recipes of the [justfile](justfile), and `just --list` names each with its arguments:

- `just ci-lint` — `gofmt` and `go vet`, and that the pinned versions agree: gopls with its lock file, spec's Go modules with the submodule, and the licenses a release carries with the modules the binary is built from;
- `just ci-test` — every test, with the race detector, among them the test of every example and test vector of the module format, which skips none; it writes the coverage to `coverage/go.out`;
- `just ci-vuln` — govulncheck; `just ci-licenses` — go-licenses, with the list of allowed licenses in `ALLOWED_LICENSES`;
- `just build` — the CLI into `bin/`; `just release-files <version>` — the files of a release into `dist/`.

A check exits with a non-zero status when it finds a problem. CI runs them all on every pull request and on `main`, runs each binary of a release on a runner of its own system, and a pull request merges only when they pass. Documents are checked by `just -f spec/justfile docs-check-repo`.

## Documents

- [docs/architecture.md](docs/architecture.md) — the component: what it does and does not do, its data and links, and what is still to decide.
- [docs/decisions.md](docs/decisions.md) — the decision log: what was decided, why, and what was rejected.
- [RUN.md](RUN.md) — the work plan, in epics.
- `run/` — the plan's other tables: [the epics for later](run/EPICS.md), [the open questions](run/QUESTIONS.md) and [the technical debt](run/DEBT.md).

The working rules, for people and coding agents alike, are in [CLAUDE.md](CLAUDE.md) and [AGENTS.md](AGENTS.md). The platform's shared vocabulary, formats, contracts and decisions are in [spec](https://github.com/altnovatrail/spec), checked out here as a git submodule: clone this repository with `git clone --recurse-submodules`.
