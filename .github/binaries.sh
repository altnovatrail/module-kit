#!/usr/bin/env bash
# Runs one binary of a release on the runner of its own system, as the job Binaries of ci.yml does on
# each of the six (Д6 in RUN.md; R41 in docs/decisions.md). Bash on Linux, macOS and the Git Bash of
# Windows, with nothing but coreutils.
#
#   .github/binaries.sh <dist> <binary> <samples> <reference>
#
# <dist> is what `just release-files` wrote, <binary> the file of it to run, <samples> what `just
# binaries-samples` wrote: examples and test vectors of the module format from spec's Go module.
# With a <reference> folder that does not exist, a run that passes writes there what the binary
# built — the files of the package, each by its sha256, and the index; with one that exists, what
# the binary builds must be the same, byte for byte, as the reference's. The test vectors hold the
# package by its digest here, and the index by the tests of the commands (TestIndexVectors).
# The build job writes the reference by the binary of linux/amd64, and each runner holds its own to
# it: the package and the index depend on the module's files alone, wherever they are built
# (format/README.md, sections 25 and 26).
set -euo pipefail

if [ "$#" -ne 4 ]; then
    echo "usage: $0 <dist> <binary> <samples> <reference>" >&2
    exit 2
fi
dist=$1 binary=$2 samples=$3 reference=$4
work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT
failures=0

fail() {
    echo "::error::$*" >&2
    failures=$((failures + 1))
}

sha256() {
    if command -v sha256sum > /dev/null; then
        sha256sum "$@"
    else
        shasum -a 256 "$@"
    fi
}

# The sha256 of every file under a folder, by its path from the folder, in byte order.
manifest() {
    (cd "$1" && find . -type f | LC_ALL=C sort | while read -r f; do
        printf '%s  %s\n' "$(sha256 "$f" | cut -d' ' -f1)" "${f#./}"
    done)
}

# run <name> <status> <command...>: runs the binary and holds its exit status; its stdout is in
# $work/<name>.out.
run() {
    local name=$1 want=$2
    shift 2
    local status=0
    "$binary" "$@" > "$work/$name.out" 2> "$work/$name.err" || status=$?
    if [ "$status" -ne "$want" ]; then
        fail "$name: exit status $status, want $want: $(cat "$work/$name.out" "$work/$name.err")"
        return 1
    fi
}

name=$(basename "$binary")
chmod +x "$binary"

# 1. The binary is the one the build wrote.
want=$(grep "  $name\$" "$dist/SHA256SUMS" | cut -d' ' -f1)
got=$(sha256 "$binary" | cut -d' ' -f1)
[ -n "$want" ] && [ "$want" = "$got" ] || fail "$name has the sha256 $got, and SHA256SUMS names ${want:-none}"

# 2. It starts, and tells its release and its version of Unicode.
if run version 0 --version; then
    first=$(sed -n 1p "$work/version.out" | tr -d '\r')
    second=$(sed -n 2p "$work/version.out" | tr -d '\r')
    case $first in "module-kit v"*) ;; *) fail "--version says \"$first\" on its first line" ;; esac
    case $second in "Unicode "[0-9]*) ;; *) fail "--version says \"$second\" on its second line" ;; esac
fi

# 3. check: a module with templates and a solver, which the Starlark sandbox runs, is accepted; a
#    module that breaks a rule is refused by it.
run check-accepted 0 check "$samples/with-templates" || true
if run check-refused 1 check "$samples/unknown-band"; then
    grep -q ': task.known-band: ' "$work/check-refused.out" || fail "check names no task.known-band: $(cat "$work/check-refused.out")"
fi

# 4. build: the package of the vector with a solver, whose digest the vector gives.
digest=$(sed -n 's/^digest: "\{0,1\}\(sha256:[0-9a-f]\{64\}\)"\{0,1\}$/\1/p' "$samples/with-solver/package.yaml")
if run build 0 build "$samples/with-solver/module" "$work/package"; then
    grep -q "$digest" "$work/build.out" || fail "build gives no digest $digest: $(cat "$work/build.out")"
fi

# 5. index: the packages of the vector of a first release, built here, and their index.
packages=()
for module in "$samples/first-release/modules"/*/; do
    module=${module%/}
    out="$work/packages/$(basename "$module")"
    mkdir -p "$work/packages"
    run "build-$(basename "$module")" 0 build "$module" "$out" && packages+=("$out")
done
run index 0 index --first --tag v0.1.0 --no-withdrawals --out "$work/module-index.json" "${packages[@]}" || true

# 6. diff: a module of that release against its index and packages, which has not changed.
if [ -f "$work/module-index.json" ]; then
    published=()
    for p in "${packages[@]}"; do published+=(--published "$p"); done
    run diff 0 diff --index "$work/module-index.json" "${published[@]}" "$samples/first-release/modules/smallest" || true
fi

# 7. What the binary built, against the reference or as the reference.
mkdir -p "$work/built"
[ -d "$work/package" ] && manifest "$work/package" > "$work/built/package.sha256"
[ -f "$work/module-index.json" ] && cp "$work/module-index.json" "$work/built/module-index.json"
if [ -d "$reference" ]; then
    for f in package.sha256 module-index.json; do
        cmp -s "$work/built/$f" "$reference/$f" || fail "$f differs from the reference of linux/amd64: $(diff "$reference/$f" "$work/built/$f" || true)"
    done
elif [ "$failures" -eq 0 ]; then
    mkdir -p "$reference"
    cp "$work/built/package.sha256" "$work/built/module-index.json" "$reference/"
fi

if [ "$failures" -gt 0 ]; then
    echo "$name: $failures of its checks failed" >&2
    exit 1
fi
echo "$name: --version, check, build, index and diff as they ought"
