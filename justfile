# module-kit's own recipes. The ci- ones are what .github's shared go.yml runs, one job each, in
# the toolchain stage of .devcontainer/Dockerfile (R16 .github; R04): their names are the same in
# every Go repository of NovaTrail, and what they run is this repository's own.

# The tools only the checks need, at exact versions. Each is a Go program run straight from its
# module, as canvas-api runs them (R48 canvas-api), so nothing is installed for it: the Go module
# proxy serves that version and the checksum database vouches for it. Dependabot does not see
# them, so they are raised by hand.
GOVULNCHECK := "golang.org/x/vuln/cmd/govulncheck@v1.9.0"
GO_LICENSES := "github.com/google/go-licenses/v2@v2.0.1"

# List the recipes.
default:
    @just --list

# Build the CLI into bin/.
build:
    go build -o bin/module-kit ./cmd/module-kit

# Fail on code gofmt would rewrite, and on what go vet reports.
ci-lint:
    #!/usr/bin/env bash
    set -euo pipefail
    # The directories of this module's own packages: spec/ is another repository.
    list="$(go list -f '{{{{.Dir}}' ./...)"
    readarray -t dirs <<< "$list"
    unformatted="$(gofmt -l "${dirs[@]}" | sort -u)"
    if [ -n "$unformatted" ]; then
        echo "ci-lint: gofmt would rewrite these files; run gofmt -w on them:" >&2
        echo "$unformatted" >&2
        exit 1
    fi
    go vet ./...

# Run the tests, with the race detector.
ci-test:
    go test -race -count=1 ./...

# Fail on a known vulnerability in code this module reaches.
ci-vuln:
    go run {{ GOVULNCHECK }} ./...

# Fail if a package this module is built from carries a license outside the list it is given:
# ALLOWED_LICENSES, SPDX identifiers separated by commas, which .github's go.yml sets.
ci-licenses:
    #!/usr/bin/env bash
    set -euo pipefail
    if [ -z "${ALLOWED_LICENSES:-}" ]; then
        echo "ci-licenses: ALLOWED_LICENSES is not set. CI takes it from .github's go.yml; to run this here, set it to the list there." >&2
        exit 1
    fi
    go run {{ GO_LICENSES }} check ./... --allowed_licenses="$ALLOWED_LICENSES"
