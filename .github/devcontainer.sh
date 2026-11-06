#!/usr/bin/env bash
# Checks the image of the devcontainer that the job Devcontainer of devcontainer.yml built, on the
# runner's own architecture (Д12 and Д14 in RUN.md; R44 in docs/decisions.md): the image is of the
# architecture given, and each tool in it answers with the version that its ARG in the Dockerfile
# pins. The build itself has checked each archive against its sha256 and gopls against its lock file;
# this holds that what came out is what the Dockerfile names.
#
#   .github/devcontainer.sh <image> <dockerfile> <architecture>
set -euo pipefail

if (($# != 3)); then
    echo "usage: $0 <image> <dockerfile> <architecture>" >&2
    exit 2
fi
image=$1 dockerfile=$2 arch=$3
failures=0

fail() {
    echo "::error::$*" >&2
    failures=$((failures + 1))
}

# The value of an ARG of the Dockerfile.
arg() {
    sed -n "s/^ARG $1=//p" "$dockerfile"
}

# check <ARG> <command...>: the first line of the command's output in the image names the version
# that the ARG pins, not as part of a longer number.
check() {
    local name=$1
    shift
    local want got pattern
    want=$(arg "$name")
    if [[ -z $want ]]; then
        fail "the Dockerfile has no ARG $name"
        return
    fi
    got=$(docker run --rm "$image" "$@" 2>&1 | head -n 1) || true
    pattern="(^|[^0-9.])${want//./\\.}([^0-9.]|$)"
    if [[ $got =~ $pattern ]]; then
        echo "$name $want: $got"
    else
        fail "$*: \"$got\", not $name $want"
    fi
}

got=$(docker run --rm "$image" dpkg --print-architecture)
if [[ $got != "$arch" ]]; then
    fail "the image is of $got, not $arch"
fi

check NODE_VERSION node --version
check JUST_VERSION just --version
check GO_VERSION go version
check BUF_VERSION buf --version
check GH_VERSION gh --version
check JUST_LSP_VERSION just-lsp --version
check CLAUDE_CODE_VERSION claude --version
check GOPLS_VERSION gopls version

where=$(docker run --rm "$image" sh -c 'command -v gopls') || true
if [[ $where != /usr/local/bin/gopls ]]; then
    fail "gopls is at \"$where\", not /usr/local/bin/gopls"
fi

if ((failures > 0)); then
    echo "$arch: $failures of the checks of the image failed" >&2
    exit 1
fi
echo "$arch: the image is of $arch, and every tool is at the version of its ARG"
