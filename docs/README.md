# module-kit's documents

What is under `docs/` and beside it, and what README.md leaves out: how a release comes out and what it holds, and how a change comes in. README.md holds only what the component is and its place in NovaTrail.

- [architecture.md](architecture.md) — the component: what it does and does not do, its data and links, what is still to decide, and how it is built.
- [decisions.md](decisions.md) — the decision log, `R01`…: what was decided, why, and what was rejected.
- [sources.md](sources.md) — the records of the facts no run checks that the documents rest on, each with its address, the date read and a quote.
- [measurements.md](measurements.md) — the results of measurements: the sandbox, the sessions of Claude Code, the speed of CI.
- [plan/](plan/) — the archives of done epics, `plan/<epic>.md`, in Russian.
- [RUN.md](../RUN.md) — the work plan, in epics, with the current epic; `run/` — the plan's other tables: [the epics for later](../run/EPICS.md), [the open questions](../run/QUESTIONS.md) and [the technical debt](../run/DEBT.md).

The working rules, for people and coding agents alike, are in [CLAUDE.md](../CLAUDE.md) and [AGENTS.md](../AGENTS.md). The platform's shared vocabulary, formats, contracts and decisions are in [spec](https://github.com/altnovatrail/spec), checked out as the git submodule `spec/`: clone the repository with `git clone --recurse-submodules`.

## Releases

A release comes out by itself once the checks pass on `main`: the next patch of the current line, tagged `vX.Y.Z` with the release, and nobody tags by hand. The releases of the line 0.1 are steps of the code's first epic, and nobody takes them; from v0.2.0 on, each release is one others take, and a new version of the module format starts a new line. Releases are immutable, so a tag stays on its commit and no file changes once it is out.

- **The CLI** is one binary for each of linux, darwin and windows, on amd64 and arm64, `module-kit_<tag>_<os>_<arch>`, `.exe` on windows, beside the licenses of the modules it is built from and `SHA256SUMS`. Each file is signed with cosign by NovaTrail's shared release workflow, its bundle `<file>.sigstore.json` beside it. Take a binary at an exact tag with its bundle, `gh release download <tag> -R altnovatrail/module-kit -p 'module-kit_<tag>_<os>_<arch>*'`, and check it before its first run with spec's `just -f spec/justfile release-verify module-kit <tag> <file>`, which checks the signature against Sigstore's trust root and everything its certificate names. `module-kit` with no arguments prints how to call each command, and `module-kit --version` its release.
- **The library** is the Go module `github.com/altnovatrail/module-kit`, at the tag of a release: `go get github.com/altnovatrail/module-kit@<tag>`. Its packages are documented by their doc comments: [pkg.go.dev](https://pkg.go.dev/github.com/altnovatrail/module-kit).

## Contributing

Report a bug or propose a change in an [issue](https://github.com/altnovatrail/module-kit/issues), in English, and open one before a pull request unless the pull request fixes a typo or a broken link. How a change comes in, with every commit signed off, and what every change keeps to are in the organization's [CONTRIBUTING.md](https://github.com/altnovatrail/.github/blob/main/CONTRIBUTING.md); the rules every NovaTrail repository shares are in spec's [RULES.md](https://github.com/altnovatrail/spec/blob/main/RULES.md): code comes with its tests. Report a vulnerability privately, as the [security policy](https://github.com/altnovatrail/module-kit/security/policy) says, never in a public issue.

Work in the devcontainer: open the repository in VS Code and choose "Reopen in Container". It holds every tool at the version the repository pins. The checks are recipes of the [justfile](../justfile), and `just --list` names each with its arguments:

- `just ci-lint` — `gofmt` and `go vet`, and that the pinned versions agree: gopls with its lock file, spec's Go modules with the submodule, and the licenses a release carries with the modules the binary is built from;
- `just ci-test` — every test, with the race detector, among them the test of every example and test vector of the module format, which skips none; it writes the coverage to `coverage/go.out`;
- `just ci-vuln` — govulncheck; `just ci-licenses` — go-licenses, with the list of allowed licenses in `ALLOWED_LICENSES`;
- `just build` — the CLI into `bin/`; `just release-files <version>` — the files of a release into `dist/`.

A check exits with a non-zero status when it finds a problem. CI runs them all on every pull request and on `main`, runs each binary of a release on a runner of its own system, and a pull request merges only when they pass. Documents are checked by `just -f spec/justfile docs-check-repo`.
