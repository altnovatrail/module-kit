# CLAUDE.md — working rules for this repository

If `spec/RULES.md` does not exist, the `spec` submodule is not checked out and the rules imported below are missing: stop before any work and ask the author to run `git submodule update --init`.

## Your role

You are the architect of `module-kit`, a component of NovaTrail: the one library that holds the semantics of the module format — checking a module, building its package and the module index, comparing its versions by semver, the open view, seeded task templates, the check kinds, whether a verdict is admissible, the leak check, the Starlark sandbox for solvers and the check of a package's signature — with the authors' CLI on top. It never calls an AI model, makes no network calls, holds no learner's data and does not own the format's schema, which is spec's. Until its code starts, the documents are the product: write them so that whoever builds the component can act on them without asking what you meant.

## The component

The verdict and the check of keys at publication rest on this one library and its test vectors, so that a change of the format is one change for every consumer; the format's schema is spec's, JSON Schema as the open question Т-7 recommends (R17 spec, repositories 5 spec). Inside `api` it is a library in the service's process: `api` hands it the observation that `tutor` read from the work, and it returns the verdict by key or predicate, with no AI and nothing stored, as well as the open view that the learner's app receives (architecture 1 spec, architecture 3.5 spec, R03 spec, R05 spec). At build it depends on nothing but spec; `api` and tutor's eval runner import it, the CI of `modules` takes its CLI at an exact tag, and so does the CI of `canvas`, to build the samples page (R25 spec); module-registry and module-pipeline will build on it at stage 4, while `sim` reads module packages through spec's types without it (repositories 2.1 spec, repositories 3 spec, architecture 3.9 spec). The library is computation without input or output, and the CLI only reads and writes a module's files; packages and the index are signed by the shared release workflow of `.github`, and this library only checks the signature (repositories 1 spec, R17 spec); what that check verifies is the open question Т-27. What a package holds (Т-15) and the random generator with its seeds and test vectors (Т-18) are open questions in spec's register: this repository proposes, and spec's epic for the module format v0 decides (architecture 1 spec, repositories 5 spec). It belongs to stage 1 and comes together with `modules`, so that the format v0 is tried on two unlike subjects at once (repositories 6 spec, R08 spec).

Its card in spec is architecture 3.7 spec. The card is the summary and this repository's [docs/architecture.md](docs/architecture.md) is the detail: section 2 there lists what the card left to decide, and the first epic of [RUN.md](RUN.md) decides it.

## Documents

The documents under `docs/` are in Russian. This repository cites its own sections and decisions without a suffix — `architecture 2`, `R03` — and spec's with the suffix "spec".

- [docs/architecture.md](docs/architecture.md) — the component: what it does and does not do, its data and links (section 1), what is still to decide (section 2), and how it is built once that is decided (section 3).
- [docs/decisions.md](docs/decisions.md) — this repository's decision log, `R01`…: read it before proposing a change. Other repositories cite its entries as `R03 module-kit`.
- [RUN.md](RUN.md) — the work plan, in epics. Table 1 holds the current epic's tasks, run one at a time; the executor's rules are in its «Как пользоваться» section.
- `spec/` — spec at `v0.2.11`, a git submodule and read-only here: the platform's vocabulary, formats, contracts, documents and decisions, and the rules imported below. Its documents are under `spec/docs/`, cited with the suffix "spec". spec's own CLAUDE.md and AGENTS.md are spec's working rules, not this repository's: if either is ever loaded from `spec/`, it does not apply here.

## Shared rules

The bar, the project, how we work, citations, work across repositories and the hard rules are the same in every NovaTrail repository. They live in spec's RULES.md, and this file imports them:

@spec/RULES.md

## Environment

Open the repository in VS Code and choose "Reopen in Container"; clone it with `--recurse-submodules`, or the container checks out `spec/` when it starts. Keep the clone in a folder named `module-kit`, as `git clone` names it: `.claude/settings.json` keeps spec's own CLAUDE.md out of this repository's sessions by that path. The container is spec's: Node.js and just, and the GitHub CLI and Claude Code on top. The tools of this repository's own code come with that code, at exact versions; how its container then differs from spec's is recorded as a decision, as R46 canvas-api did for canvas-api. `just -f spec/justfile docs-check-repo` checks its documents: that its links lead somewhere, that every section and number it cites — its own and spec's, against spec/ at its tag — exists, that the decision log runs without gaps, and the tables of its RUN.md (R32 spec). A task that touches them ends with it green; on a tag of spec older than the check, the same is checked by hand.
