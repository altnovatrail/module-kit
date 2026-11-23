---
paths:
  - "**/*.go"
  - "go.mod"
  - "go.sum"
  - "internal/schema/schemas/**"
  - "page/page.yaml"
---

# The Go library and its CLI

module-kit is one Go module, `github.com/altnovatrail/module-kit`, with the `go` line of `go.mod` as the floor for those who import it; the Go that builds and checks it is the `toolchain` stage's, with `GOTOOLCHAIN=local`, and it moves only with a move to a tag of spec. Its packages, the CLI and how each is built are in `docs/architecture.md`; why — in the decision log, `docs/decisions.md`.

The shared `go.yml` of `.github` runs four recipes of the justfile, one job each, and a task that changes Go code, `go.mod` or `go.sum` ends with all four green here:

- `just ci-lint` — gofmt and `go vet`; `GOPLS_VERSION` of `.devcontainer/Dockerfile` against the gopls of `.devcontainer/gopls/go.mod`; `release-licenses`, so that a module the CLI comes to be built from without its license in `THIRD_PARTY` fails on its pull request; and spec's Go modules `spec` and `spec/contracts` at the commit of the `spec/` submodule, one tag.
- `just ci-test` — `go test -race` with `-coverpkg=./...`, the tests at the clock of a predicate after the rest, with nothing beside them — a package whose tests come to take a predicate to its clock goes into `CLOCK_PACKAGES`, since beside other packages a busy machine can stop it by that clock —, its profile in `coverage/go.out`, which goes to Codecov; a merge waits for `codecov/patch`, at least 90% of the changed lines covered. The test of vectors, `internal/vectors`, passes every test vector of spec's format and skips none.
- `just ci-vuln` — govulncheck at the version the justfile pins.
- `just ci-licenses` — go-licenses against `ALLOWED_LICENSES`, which CI takes from `go.yml` of `.github`; to run it here, set it to that list.

The tools of the checks run by `go run` at exact versions and are raised by hand. spec's Go modules and go.starlark.net are not raised by Dependabot: the modules move with a tag of spec, and the interpreter's version is the format's, which a test holds. The code's copies of the format's schemas, `internal/schema/schemas/`, and of `page.yaml`, `page/page.yaml`, are held to spec's by their tests. A new third-party module carries a license compatible with Apache-2.0 and a record of where it was read in `docs/sources.md`; a file the code embeds by `//go:embed` goes into the `paths` of the release (`.claude/rules/ci.md`).

A session gets gopls's diagnostics after each edit of a Go file, through the plugin `gopls-lsp` that `CLAUDE_CODE_PLUGIN_DIRS` of `.devcontainer/devcontainer.json` loads; they do not stand in for the four checks.
