---
paths:
  - "**/*.go"
  - "go.mod"
  - "go.sum"
  - "internal/schema/schemas/**"
  - "page/page.yaml"
---

# The Go library and its CLI

module-kit is one Go module, `github.com/altnovatrail/module-kit`, with `go 1.27.0` in `go.mod` as the floor for those who import it; the Go that builds and checks it is the `toolchain` stage's, with `GOTOOLCHAIN=local`, and it moves only with a move to a tag of spec (R04, R08, R54). Its packages, the CLI and how each is built are architecture 3; why — the decision log.

The shared `go.yml` of `.github` runs four recipes of the justfile, one job each, and a task that changes Go code, `go.mod` or `go.sum` ends with all four green here:

- `just ci-lint` — gofmt and `go vet`; `GOPLS_VERSION` of `.devcontainer/Dockerfile` against the gopls of `.devcontainer/gopls/go.mod` (R19, R51); `release-licenses`, so that a module the CLI comes to be built from without its license in `THIRD_PARTY` fails on its pull request (R38); and spec's Go modules `spec` and `spec/contracts` at the commit of the `spec/` submodule, one tag (R64 spec; R20).
- `just ci-test` — `go test -race` with `-coverpkg=./...`, its profile in `coverage/go.out`, which goes to Codecov; a merge waits for `codecov/patch`, at least 90% of the changed lines covered (R48 spec; R18). The test of vectors, `internal/vectors`, passes every test vector of spec's format and skips none (R20, R42).
- `just ci-vuln` — govulncheck at the version the justfile pins.
- `just ci-licenses` — go-licenses against `ALLOWED_LICENSES`, which CI takes from `go.yml` of `.github`; to run it here, set it to that list (R04).

The tools of the checks run by `go run` at exact versions and are raised by hand (R04). spec's Go modules and go.starlark.net are not raised by Dependabot: the modules move with a tag of spec, and the interpreter's version is the format's, which a test holds (R20, R28). The code's copies of the format's schemas, `internal/schema/schemas/`, and of `page.yaml`, `page/page.yaml`, are held to spec's by their tests (R23, R52). A new third-party module carries a license compatible with Apache-2.0 (RULES.md, Licenses) and a record of where it was read in `docs/sources.md`; a file the code embeds by `//go:embed` goes into the `paths` of the release (`.claude/rules/ci.md`).

A session gets gopls's diagnostics after each edit of a Go file, through the plugin `gopls-lsp` that `.claude/settings.json` turns on (R55); they do not stand in for the four checks.
