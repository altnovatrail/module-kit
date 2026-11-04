# module-kit

[![CI](https://github.com/altnovatrail/module-kit/actions/workflows/ci.yml/badge.svg)](https://github.com/altnovatrail/module-kit/actions/workflows/ci.yml)
[![Release](https://github.com/altnovatrail/module-kit/actions/workflows/release.yml/badge.svg)](https://github.com/altnovatrail/module-kit/actions/workflows/release.yml)
[![CodeQL](https://github.com/altnovatrail/module-kit/actions/workflows/codeql.yml/badge.svg)](https://github.com/altnovatrail/module-kit/actions/workflows/codeql.yml)
[![Latest release](https://img.shields.io/github/v/release/altnovatrail/module-kit)](https://github.com/altnovatrail/module-kit/releases/latest)
[![Go version](https://img.shields.io/github/go-mod/go-version/altnovatrail/module-kit)](go.mod)
[![Coverage](https://codecov.io/gh/altnovatrail/module-kit/branch/main/graph/badge.svg)](https://codecov.io/gh/altnovatrail/module-kit)
[![OpenSSF Scorecard](https://api.scorecard.dev/projects/github.com/altnovatrail/module-kit/badge)](https://scorecard.dev/viewer/?uri=github.com/altnovatrail/module-kit)
[![License](https://img.shields.io/github/license/altnovatrail/module-kit)](LICENSE)

[![Code Smells](https://sonarcloud.io/api/project_badges/measure?project=altnovatrail_module-kit&metric=code_smells)](https://sonarcloud.io/summary/new_code?id=altnovatrail_module-kit)
[![Reliability Rating](https://sonarcloud.io/api/project_badges/measure?project=altnovatrail_module-kit&metric=reliability_rating)](https://sonarcloud.io/summary/new_code?id=altnovatrail_module-kit)
[![Security Rating](https://sonarcloud.io/api/project_badges/measure?project=altnovatrail_module-kit&metric=security_rating)](https://sonarcloud.io/summary/new_code?id=altnovatrail_module-kit)
[![Maintainability Rating](https://sonarcloud.io/api/project_badges/measure?project=altnovatrail_module-kit&metric=sqale_rating)](https://sonarcloud.io/summary/new_code?id=altnovatrail_module-kit)
[![Vulnerabilities](https://sonarcloud.io/api/project_badges/measure?project=altnovatrail_module-kit&metric=vulnerabilities)](https://sonarcloud.io/summary/new_code?id=altnovatrail_module-kit)

The NovaTrail module format in code: the one Go library that holds the format's semantics, and the authors' CLI on top of it. It reads an author's module and checks it — the schema, a skill graph without cycles, references, the shared taxonomy, keys by solver and fallback texts by leak check — builds the module package and the module index, and compares a new version of a module with every published one by the format's rules of versions. It also fills seeded task templates and runs their solvers in a Starlark sandbox, and for the platform's service it builds the open view of a task, joins two readings of a learner's work into one observation and runs the check kinds on it, admits a verdict only by the provenance of its key and trusts a package by the chain of digests from a pinned index; signatures it leaves to spec's `release-verify`. It never calls an AI model: the library makes no network calls, the CLI only reads and writes a module's files, and neither holds learners' data.

A component of [NovaTrail](https://github.com/altnovatrail/spec/tree/v0.9.2), a universal open platform for adaptive learning, interactive lessons and AI tutoring. It belongs to stage 1, Format and core, with spec, modules, learner-model, engine and sim; together with modules it tries the module format v0 on two unlike subjects at once.

## Status

Detailing is done. The documents came first: the first epic of [RUN.md](RUN.md), now archived in [docs/plan/E1.md](docs/plan/E1.md), decided what the component's card left open, and the code follows in the order RUN.md sets. The one code here so far is the skeleton of the CLI, `module-kit --version`: the author had it written ahead of the format, so that the shared Go checks of .github have code to be tried on (R04). The code of the format comes in the current epic, E2 in RUN.md, on the module format 0.3 that spec defines in its `format/`, and its first release that others take is the first of the line 0.2.

## Documents

The documents are in Russian.

- [docs/architecture.md](docs/architecture.md) — the component: what it does and does not do, its data and links, and what is still to decide.
- [docs/decisions.md](docs/decisions.md) — the decision log: what was decided, why, and what was rejected.
- [RUN.md](RUN.md) — the work plan, in epics.

The working rules, for people and coding agents alike, are in [CLAUDE.md](CLAUDE.md) and [AGENTS.md](AGENTS.md). The platform's shared vocabulary, formats, contracts and decisions are in [spec](https://github.com/altnovatrail/spec/tree/v0.9.2), checked out here as a git submodule at `v0.9.2`: clone this repository with `git clone --recurse-submodules`.

## License

[Apache-2.0](LICENSE).
