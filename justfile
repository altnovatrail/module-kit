# module-kit's own recipes. The ci- ones are what .github's shared go.yml runs, one job each, in
# the toolchain stage of .devcontainer/Dockerfile (R16 .github; R04): their names are the same in
# every Go repository of NovaTrail, and what they run is this repository's own.

# The tools only the checks need, at exact versions. Each is a Go program run straight from its
# module, as canvas-api's justfile runs them, so nothing is installed for it: the Go module proxy
# serves that version and the checksum database vouches for it. Dependabot does not see them,
# so they are raised by hand.
GOVULNCHECK := "golang.org/x/vuln/cmd/govulncheck@v1.9.0"
GO_LICENSES := "github.com/google/go-licenses/v2@v2.0.1"

# List the recipes.
default:
    @just --list

# Build the CLI into bin/.
build:
    go build -o bin/module-kit ./cmd/module-kit

# What .github's shared release.yml runs to build a release, in the toolchain stage of the commit
# it releases, from that commit's tree with no git in it, and then signs every file (R07; R18
# .github). dist/ is made afresh and holds plain files alone: the CLI for each platform below,
# built without cgo and named module-kit_<version>_<os>_<arch>, the licenses a binary has to come
# with, and SHA256SUMS over the rest.
# Build a release's files into dist/, for a version such as v0.1.0.
release-files version:
    #!/usr/bin/env bash
    set -euo pipefail
    version={{ quote(version) }}
    if ! [[ "$version" =~ ^v[0-9]+\.[0-9]+\.[0-9]+(-[0-9A-Za-z.-]+)?$ ]]; then
        echo "release-files: the version is v<major>.<minor>.<patch>, as v0.1.0, not ${version}" >&2
        exit 1
    fi
    # Nothing of the caller's Go environment reaches the build: the binaries are the same wherever
    # the recipe runs, and a release can be built again here and compared with its SHA256SUMS.
    export GOENV=off GOFLAGS= GOEXPERIMENT=
    unset GOAMD64 GOARM64 GOOS GOARCH
    rm -rf dist
    mkdir dist
    for platform in linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64 windows/arm64; do
        os="${platform%/*}"
        arch="${platform#*/}"
        name="module-kit_${version}_${os}_${arch}"
        if [ "$os" = windows ]; then
            name="${name}.exe"
        fi
        # No VCS stamp, so that a build from a checkout matches one from the release's tree.
        CGO_ENABLED=0 GOOS="$os" GOARCH="$arch" go build -trimpath -buildvcs=false \
            -ldflags "-s -w -X main.version=${version}" -o "dist/${name}" ./cmd/module-kit
    done
    # Apache-2.0 asks a copy of itself to go with the work, and Go's BSD license asks a binary
    # that holds its standard library to reproduce its notice.
    cp LICENSE dist/LICENSE
    cp "$(go env GOROOT)/LICENSE" dist/LICENSE-go
    (cd dist && sha256sum -- module-kit_* LICENSE LICENSE-go > SHA256SUMS)

# Fail on code gofmt would rewrite, on what go vet reports, and on a gopls lock file that names
# another version than GOPLS_VERSION in the Dockerfile.
ci-lint:
    #!/usr/bin/env bash
    set -euo pipefail
    # The Go files git tracks or would, outside testdata/: spec/ is a submodule, another
    # repository's, and git lists none of its files.
    unformatted="$(git ls-files -z --cached --others --exclude-standard -- '*.go' ':!:**/testdata/**' \
        | xargs -0 -r gofmt -l)"
    if [ -n "$unformatted" ]; then
        echo "ci-lint: gofmt would rewrite these files; run gofmt -w on them:" >&2
        echo "$unformatted" >&2
        exit 1
    fi
    go vet ./...
    # gopls comes into the devcontainer image from its lock file, .devcontainer/gopls/, and the
    # image's build stops when the gopls it built does not name GOPLS_VERSION (R19). CI never
    # builds that stage, so the two are held to each other here: a pull request that raises one
    # alone, or a move to a new tag of spec that drops the step from the Dockerfile, fails here.
    lock="$(sed -n 's/^[[:space:]]*\(require[[:space:]]\{1,\}\)\{0,1\}golang\.org\/x\/tools\/gopls[[:space:]]\{1,\}v\([^[:space:]]\{1,\}\).*$/\2/p' \
        .devcontainer/gopls/go.mod)"
    arg="$(sed -n 's/^ARG GOPLS_VERSION=\(.*\)$/\1/p' .devcontainer/Dockerfile)"
    if [ -z "$lock" ] || [ "$arg" != "$lock" ]; then
        echo "ci-lint: .devcontainer/gopls/go.mod names gopls ${lock:-at no version} and" \
            ".devcontainer/Dockerfile has GOPLS_VERSION ${arg:-unset}; raise them together (R19)" >&2
        exit 1
    fi

# The coverage goes into coverage/go.out, one Go profile, which the job coverage of .github's shared
# go.yml sends to Codecov, as "Go / coverage" (R48 spec; R18). Go counts statements and leaves the
# tests themselves out; a package that no test runs is in it too, its statements at zero.
# Run the tests, with the race detector, and write their coverage.
ci-test:
    mkdir -p coverage
    go test -race -count=1 -covermode=atomic -coverprofile=coverage/go.out ./...

# Fail on a known vulnerability in code this module reaches.
ci-vuln:
    go run {{ GOVULNCHECK }} ./...

# The list is ALLOWED_LICENSES, SPDX identifiers separated by commas, which .github's go.yml sets.
# Fail if a package this module is built from carries a license outside that list.
ci-licenses:
    #!/usr/bin/env bash
    set -euo pipefail
    if [ -z "${ALLOWED_LICENSES:-}" ]; then
        echo "ci-licenses: ALLOWED_LICENSES is not set. CI takes it from .github's go.yml; to run this here, set it to the list there." >&2
        exit 1
    fi
    go run {{ GO_LICENSES }} check ./... --allowed_licenses="$ALLOWED_LICENSES"
