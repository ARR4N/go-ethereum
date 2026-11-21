#!/usr/bin/env bash
#
# Flip the Go module path of the current tree between libevm naming
# (github.com/ava-labs/libevm) and geth naming (github.com/ethereum/go-ethereum).
# Direction is detected from go.mod. Changes the working tree only; committing
# is left to the caller.
#
# The procedure is copied verbatim from .github/workflows/rename-module.yml in
# ava-labs/libevm so results are byte-identical to existing rename commits. In
# particular the `find` expression is NOT "fixed": its final `-not -wholename`
# binds only to the `*.proto` clause. See docs/07-baselines-and-facts.md.
#
# Usage: rename-module.sh        (run from anywhere inside the tree to rename)

set -euo pipefail

cd "$(git rev-parse --show-toplevel)"

readonly LIBEVM="ava-labs/libevm"
readonly GETH="ethereum/go-ethereum"

current="$(sed -n 's|^module github\.com/||p' go.mod)"
case "${current}" in
  "${LIBEVM}") from="${LIBEVM}"; to="${GETH}" ;;
  "${GETH}") from="${GETH}"; to="${LIBEVM}" ;;
  *)
    echo "Unexpected Go module github.com/${current}" >&2
    exit 1
    ;;
esac

go mod edit -module "github.com/${to}"

find . \
  -iname '*.go' \
  -o -iname '*.txt' \
  -o -iname '*.go.tpl' \
  -o -iname '*.proto' \
  -not -wholename '*/libevm/tooling/*' | xargs \
  sed -i -E "s|([\"\`]github\.com/)${from}|\1${to}|g"

echo "Renamed github.com/${from} -> github.com/${to}" >&2
