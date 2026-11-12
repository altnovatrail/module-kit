---
paths:
  - ".github/**"
  - ".devcontainer/**"
  - "justfile"
  - ".sonarcloud.properties"
  - ".bestpractices.json"
---

# CI, the release and the devcontainer

CI runs every check on every pull request and on `main`; what a change here does is seen on its pull request, and architecture 3.1 holds the detail.

- **Checks.** `.github/workflows/ci.yml` calls the shared workflows of `.github` by SHA — `secrets.yml`, `dco.yml`, `documents.yml`, the documents check, and `go.yml`, the four recipes of `.claude/rules/go.md` (R18) — and runs the jobs Binaries: `just release-files v0.0.0-ci` and `just binaries-samples <folder>` in the `toolchain` stage, then `.github/binaries.sh <dist> <binary> <samples> <reference>` runs each of the six binaries on a runner of its own system and holds what it builds to what linux/amd64 built, byte for byte (R41). A change to a command of the CLI or to its samples changes `binaries.sh` with it. `codeql.yml` calls the shared `codeql.yml` with `languages: go` (R06); `scorecard.yml` runs OpenSSF Scorecard.
- **The devcontainer.** `.devcontainer/` is byte for byte what spec's `tools/new-repo/` builds for the kind `go` at the submodule's tag, and a move to a tag takes it afresh (R51, R54); gopls comes from its lock file, `.devcontainer/gopls/`. `devcontainer.yml` builds the `devcontainer` stage on amd64 and arm64 when `.devcontainer/**` changes, and `.github/devcontainer.sh <image> <dockerfile> <architecture>` checks each tool against its `ARG` (R44).
- **The release.** `release.yml` calls the shared `release.yml` of `.github` once CI is green on `main`: the next patch of `RELEASE_LINE`, and only when something in its `paths` changed — what a release is built from, listed in `release.yml` itself; a file the code comes to embed goes into that list (R18, R23, R52). `just release-files <version>` builds `dist/` from the commit's tree — six binaries, the licenses `release-licenses` names, `SHA256SUMS` — reproducibly, and the shared workflow signs each file (R07, R38; R77 spec). A new version of the format is a new release line (R52).
- **The rest.** `.github/dependabot.yml` says what Dependabot raises; `.github/codecov.yml` holds `codecov/patch` to 90% (R48 spec); `.sonarcloud.properties` is SonarCloud's scope; `.bestpractices.json` holds the answers of project 16385 on bestpractices.dev, and a change that changes an answer changes the file (R48 spec). The tools the justfile pins, govulncheck and go-licenses, are raised by hand (R04).
