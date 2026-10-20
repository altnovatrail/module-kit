# module-kit

The NovaTrail module format in code: the one Go library that holds the format's semantics, and the authors' CLI on top of it. It reads an author's module and checks it — the schema, a skill graph without cycles, references, the shared taxonomy, keys by solver and fallback texts by leak check — builds the module package and the module index, and compares a new version of a module with the last by semver. It also fills seeded task templates, runs solvers in a Starlark sandbox and checks a package's signature, and for the platform's service it builds the open view of a task, runs the check kinds and admits a verdict only by the provenance of its key. It never calls an AI model: the library makes no network calls, the CLI only reads and writes a module's files, and neither holds learners' data.

A component of [NovaTrail](https://github.com/altnovatrail/spec/tree/v0.2.6), a universal open platform for adaptive learning, interactive lessons and AI tutoring. It belongs to stage 1, Format and core, with spec, modules, learner-model, engine and sim; together with modules it tries the module format v0 on two unlike subjects at once.

## Status

Detailing. The documents come first: the first epic of [RUN.md](RUN.md) decides what the component's card left open, and the code follows in the order RUN.md sets. There is no code here yet: it follows the module format v0, which spec has still to define.

## Documents

The documents are in Russian.

- [docs/architecture.md](docs/architecture.md) — the component: what it does and does not do, its data and links, and what is still to decide.
- [docs/decisions.md](docs/decisions.md) — the decision log: what was decided, why, and what was rejected.
- [RUN.md](RUN.md) — the work plan, in epics.

The working rules, for people and coding agents alike, are in [CLAUDE.md](CLAUDE.md) and [AGENTS.md](AGENTS.md). The platform's shared vocabulary, formats, contracts and decisions are in [spec](https://github.com/altnovatrail/spec/tree/v0.2.6), checked out here as a git submodule at `v0.2.6`: clone this repository with `git clone --recurse-submodules`.

## License

[Apache-2.0](LICENSE).
