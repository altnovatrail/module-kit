# module-kit's own recipes. The ci- ones are what .github's shared go.yml runs, one job each, in
# the toolchain stage of .devcontainer/Dockerfile, and ci-lint, ci-vuln and ci-licenses in
# toolchain-core once toolchain holds more than it (R16 .github, R47 .github; R04): so a tool they
# run goes into the core, and one only the tests need, into toolchain. Their names are the same in
# every Go repository of NovaTrail, and what they run is this repository's own.

# The tools only the checks need, at exact versions. Each is a Go program run straight from its
# module, so nothing is installed for it: the Go module proxy serves that version and the checksum
# database vouches for it. Dependabot does not see them, so they are raised by hand (R04).
GOVULNCHECK := "golang.org/x/vuln/cmd/govulncheck@v1.9.0"
GO_LICENSES := "github.com/google/go-licenses/v2@v2.0.1"

# List the recipes.
default:
    @just --list

# Build the CLI into bin/.
build:
    go build -o bin/module-kit ./cmd/module-kit

# What .github's shared release.yml runs to build a release, in the toolchain-core stage of the
# commit it releases, from that commit's tree with no git in it, and then signs every file (R07;
# R18 .github, R47 .github). dist/ is made afresh and holds plain files alone: the CLI for each
# platform below, built without cgo and named module-kit_<version>_<os>_<arch>, the licenses a
# binary has to come with, and SHA256SUMS over the rest.
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
    # The six builds run at once over the one build cache, which go build shares safely between
    # processes: each target compiles packages of its own, and one build after another leaves cores
    # idle in each link. A binary is the same as one built alone. Each build writes what it says to
    # a log of its own, outside dist/, which a failure prints under its platform; a recipe that stops
    # before the builds end stops them too (R59).
    logs="$(mktemp -d)"
    pids=()
    trap 'kill "${pids[@]}" 2> /dev/null || true; rm -rf "$logs"' EXIT
    platforms=(linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64 windows/arm64)
    for i in "${!platforms[@]}"; do
        os="${platforms[$i]%/*}"
        arch="${platforms[$i]#*/}"
        name="module-kit_${version}_${os}_${arch}"
        if [ "$os" = windows ]; then
            name="${name}.exe"
        fi
        # No VCS stamp, so that a build from a checkout matches one from the release's tree.
        CGO_ENABLED=0 GOOS="$os" GOARCH="$arch" go build -trimpath -buildvcs=false \
            -ldflags "-s -w -X main.version=${version}" -o "dist/${name}" ./cmd/module-kit \
            > "$logs/$i" 2>&1 &
        pids+=("$!")
    done
    failed=0
    for i in "${!pids[@]}"; do
        if ! wait "${pids[$i]}"; then
            echo "release-files: the build for ${platforms[$i]} failed:" >&2
            cat "$logs/$i" >&2
            failed=1
        fi
    done
    if [ "$failed" -ne 0 ]; then
        exit 1
    fi
    # Apache-2.0 asks a copy of itself to go with the work, and Go's BSD license asks a binary
    # that holds its standard library to reproduce its notice; so do the licenses of the modules
    # the binaries are built from, which release-licenses names (R38).
    cp LICENSE dist/LICENSE
    cp "$(go env GOROOT)/LICENSE" dist/LICENSE-go
    just release-licenses dist
    (cd dist && sha256sum -- module-kit_* LICENSE LICENSE-* NOTICE-* > SHA256SUMS)

# What the job Binaries of ci.yml runs each binary of a release on (R41): an example module with
# templates and a solver, a test vector of a rule, the test vector of a package with a solver and
# the modules of the test vector of a first release, copied from spec's Go module at the version
# go.mod requires (R20), as CI checks out no submodule. The Go module cache holds its files
# read-only, and the copies are made writable.
# Copy the samples that the job Binaries runs the binaries on into a folder.
binaries-samples folder:
    #!/usr/bin/env bash
    set -euo pipefail
    folder={{ quote(folder) }}
    format="$(go list -m -f '{{{{.Dir}}' github.com/altnovatrail/spec)/format"
    rm -rf "$folder"
    mkdir -p "$folder/first-release"
    cp -R "$format/examples/module/valid/with-templates" "$folder/with-templates"
    cp -R "$format/vectors/rules/task.known-band/unknown-band" "$folder/unknown-band"
    cp -R "$format/vectors/packages/with-solver" "$folder/with-solver"
    cp -R "$format/vectors/indexes/first-release/modules" "$folder/first-release/modules"
    chmod -R u+w "$folder"

# The modules besides the standard library that the CLI is built from, on any of the six platforms
# of a release, each with the name its license takes in a release: LICENSE-<name>, and NOTICE-<name>
# beside it when the module has a NOTICE. spec's LICENSE is this repository's own, Apache-2.0 word
# for word, so LICENSE stands for it, and the recipe checks that it still is (R38).
THIRD_PARTY := "go.yaml.in/yaml/v4=go-yaml github.com/santhosh-tekuri/jsonschema/v6=jsonschema go.starlark.net=starlark golang.org/x/text=x-text golang.org/x/sys=x-sys golang.org/x/image=x-image"
SAME_LICENSE := "github.com/altnovatrail/spec"

# Copy into a folder the licenses and notices of the modules that the CLI is built from, and fail
# when the modules are not those THIRD_PARTY and SAME_LICENSE name.
release-licenses folder:
    #!/usr/bin/env bash
    set -euo pipefail
    folder={{ quote(folder) }}
    export GOENV=off GOFLAGS= GOEXPERIMENT= CGO_ENABLED=0
    unset GOAMD64 GOARM64
    built="$(for platform in linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64 windows/arm64; do
        GOOS="${platform%/*}" GOARCH="${platform#*/}" go list -deps -f '{{{{with .Module}}{{{{.Path}}{{{{end}}' ./cmd/module-kit
    done | sort -u | grep -v -x -F 'github.com/altnovatrail/module-kit' || true)"
    named="$(for entry in {{ THIRD_PARTY }} {{ SAME_LICENSE }}; do echo "${entry%%=*}"; done | sort -u)"
    if [ "$built" != "$named" ]; then
        echo "release-licenses: the CLI is built from these modules besides the standard library:" >&2
        echo "${built:-none}" >&2
        echo "and the justfile names these, whose licenses go with a release:" >&2
        echo "$named" >&2
        echo "Name each module the CLI is built from in THIRD_PARTY, with its license (R38)." >&2
        exit 1
    fi
    for module in {{ SAME_LICENSE }}; do
        if ! cmp -s LICENSE "$(go list -m -f '{{{{.Dir}}' "$module")/LICENSE"; then
            echo "release-licenses: the LICENSE of ${module} is no longer this repository's own; name it in THIRD_PARTY (R38)." >&2
            exit 1
        fi
    done
    mkdir -p "$folder"
    for entry in {{ THIRD_PARTY }}; do
        dir="$(go list -m -f '{{{{.Dir}}' "${entry%%=*}")"
        cp "$dir/LICENSE" "$folder/LICENSE-${entry#*=}"
        if [ -f "$dir/NOTICE" ]; then
            cp "$dir/NOTICE" "$folder/NOTICE-${entry#*=}"
        fi
    done

# Fail on code gofmt would rewrite, on what go vet reports, on a gopls lock file that names
# another version than GOPLS_VERSION in the Dockerfile, and on a version of spec's Go modules that
# is not the tag of the spec submodule.
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
    # gopls comes into the devcontainer image from its lock file, .devcontainer/gopls/, the go kind's
    # of spec's repo-template, and the image's build stops when the gopls it built does not name
    # GOPLS_VERSION (R102 spec; R19, R51). CI never builds that stage, so the two are held to each
    # other here: a pull request that raises one alone, as a security update of Dependabot may, fails
    # here.
    lock="$(sed -n 's/^[[:space:]]*\(require[[:space:]]\{1,\}\)\{0,1\}golang\.org\/x\/tools\/gopls[[:space:]]\{1,\}v\([^[:space:]]\{1,\}\).*$/\2/p' \
        .devcontainer/gopls/go.mod)"
    arg="$(sed -n 's/^ARG GOPLS_VERSION=\(.*\)$/\1/p' .devcontainer/Dockerfile)"
    if [ -z "$lock" ] || [ "$arg" != "$lock" ]; then
        echo "ci-lint: .devcontainer/gopls/go.mod names gopls ${lock:-at no version} and" \
            ".devcontainer/Dockerfile has GOPLS_VERSION ${arg:-unset}; raise them together, as spec's template has them (R51)" >&2
        exit 1
    fi
    # A module that the CLI comes to be built from, without its license in a release, fails here,
    # on its pull request, rather than at the release (R38).
    licenses="$(mktemp -d)"
    trap 'rm -rf "$licenses"' EXIT
    just release-licenses "$licenses"
    # The code takes spec's files from its Go modules, and the documents from the submodule spec/,
    # and the two move together, to one tag (R64 spec; R20): the commit the index records for the
    # submodule is the one the Go module proxy names as the origin of each module's version, and
    # contracts carries the number of spec's tag. The index is the commit's tree in CI, which checks
    # out no submodule, and here it holds a move to a new tag once spec/ is added, before a commit.
    submodule="$(git ls-files --stage -- spec | awk '$1 == "160000" { print $2 }')"
    spec_version="$(go list -m -f '{{{{.Version}}' github.com/altnovatrail/spec)"
    contracts_version="$(go list -m -f '{{{{.Version}}' github.com/altnovatrail/spec/contracts)"
    # The origin of a version is what proxy.golang.org gives with it; a proxy that gives none fails
    # the check, by name.
    origin() {
        go mod download -json "$1" | sed -n 's/^[[:space:]]*"Hash": "\([0-9a-f]\{40\}\)",\{0,1\}$/\1/p'
    }
    spec_origin="$(origin "github.com/altnovatrail/spec@${spec_version}")"
    contracts_origin="$(origin "github.com/altnovatrail/spec/contracts@${contracts_version}")"
    if [ -z "$submodule" ] || [ "$spec_origin" != "$submodule" ] || [ "$contracts_origin" != "$submodule" ] \
        || [ "$contracts_version" != "$spec_version" ]; then
        echo "ci-lint: the submodule spec/ is at ${submodule:-no commit}, while go.mod requires" \
            "github.com/altnovatrail/spec ${spec_version}, from ${spec_origin:-a commit the Go module proxy does not give}," \
            "and github.com/altnovatrail/spec/contracts ${contracts_version}, from" \
            "${contracts_origin:-a commit the Go module proxy does not give}; move them to one tag together" \
            "(R64 spec; R20)" >&2
        exit 1
    fi

# What ci-test runs after the rest (R60): the packages whose tests take Starlark to the clock of a
# second of sandbox.Predicate — a predicate to its million steps, 32 evaluations at once —, and the
# cases of TestVectors that take a predicate there. Each element of a -run pattern is one element
# of a test's name, and a case's name is its path in spec. A package whose tests come to do so
# too goes into CLOCK_PACKAGES.
CLOCK_PACKAGES := "./internal/sandbox ./verdict"
PREDICATE_CASES := '^TestVectors$/^format$/^vectors$/^verdicts$/^predicate$'

# The coverage goes into coverage/go.out, one Go profile, which the job coverage of .github's shared
# go.yml sends to Codecov, as "Go / coverage" (R48 spec; R18). Go counts statements and leaves the
# tests themselves out; a package that no test runs is in it too, its statements at zero. Every
# test counts for every package of the module, -coverpkg=./...: the test of vectors in
# internal/vectors runs the code of the other packages, and its lines count where they are (R20).
# The tests at the clock run after the rest, one package at a time with nothing beside them: under
# -race a Starlark step is several times slower, and packages running beside them pushed one past
# the clock (R11, R60). The run of PREDICATE_CASES prints each case with its time, the margin on
# this machine, and fails if it ran none, so that a folder spec renames cannot drop them unseen.
# The profiles of the later runs join coverage/go.out without their mode line, as go test joins
# the profiles of its packages.
# Run the tests, with the race detector, and write their coverage.
ci-test:
    #!/usr/bin/env bash
    set -euo pipefail
    mkdir -p coverage
    later=$(mktemp -d)
    trap 'rm -rf "$later"' EXIT
    flags=(-race -count=1 -covermode=atomic -coverpkg=./...)
    clock=$(go list {{ CLOCK_PACKAGES }})
    rest=$(go list ./... | grep -vxF "$clock")
    mapfile -t rest <<<"$rest"
    echo "ci-test: the module but for the tests at the clock of a predicate"
    go test "${flags[@]}" -skip '{{ PREDICATE_CASES }}' -coverprofile=coverage/go.out "${rest[@]}"
    echo "ci-test: {{ CLOCK_PACKAGES }}, one at a time"
    go test "${flags[@]}" -p 1 -coverprofile="$later/clock.out" {{ CLOCK_PACKAGES }}
    echo "ci-test: the verdicts of a predicate of TestVectors"
    go test "${flags[@]}" -run '{{ PREDICATE_CASES }}' -v -coverprofile="$later/predicate.out" \
        ./internal/vectors | tee "$later/predicate.log"
    if ! grep -q -- '--- PASS: TestVectors/' "$later/predicate.log"; then
        echo "ci-test: PREDICATE_CASES ran no case of TestVectors (R60)" >&2
        exit 1
    fi
    for profile in "$later/clock.out" "$later/predicate.out"; do
        tail -n +2 "$profile" >> coverage/go.out
    done

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
