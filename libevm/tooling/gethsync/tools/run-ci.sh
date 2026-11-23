#!/usr/bin/env bash
#
# Run the CI gates other than the main test run (which is run-oracle.sh) on a
# tree, as .github/workflows/{go,lint}.yml do. Every checkpoint of a sync must
# pass these as well as the oracles, so that a break can be pinned to one
# upstream version.
#
#   nocgo     CGO_ENABLED=0: go build ./... and go test ./crypto/...
#   tooling   go test ./... in the libevm/tooling module
#   generate  go generate leaves the tree unchanged
#   tidy      go mod tidy leaves go.mod/go.sum unchanged (root and tooling)
#   golangci  golangci-lint run (root), version pinned in lint.yml
#   yamllint  yamllint -c .yamllint.yml on tracked YAML files
#   shell     shellcheck on scripts under ./libevm
#
# `generate` and `tidy` modify the tree to see whether anything changes, then
# restore it. Run this on a clean tree; it refuses otherwise.
#
# Usage: run-ci.sh <tree> <log-dir> [gate ...]     (default: all gates)
# Env:   TARGET_BRANCH  passed to the tooling tests (default: main)
#        PATH           must provide golangci-lint (CI's version), yamllint
#                       and the shellcheck binary
# Exit:  0 if every gate run passed.

# Gates are invoked indirectly, as "gate_${g}".
# shellcheck disable=SC2317

set -uo pipefail

if [[ $# -lt 2 ]]; then
  echo "usage: $0 <tree> <log-dir> [gate ...]" >&2
  exit 2
fi
tree="$(cd "$1" && pwd)"
mkdir -p "$2"
logs="$(cd "$2" && pwd)"
readonly tree logs
shift 2
gates=("$@")
if [[ ${#gates[@]} -eq 0 ]]; then
  gates=(nocgo tooling generate tidy golangci yamllint shell)
fi

cd "${tree}" || exit 2
if [[ -n "$(git status --porcelain --untracked-files=no)" ]]; then
  echo "tree has uncommitted changes; generate and tidy need a clean tree" >&2
  exit 2
fi
module="$(go list -m)"
readonly module

restore() { git checkout -q -- . && git clean -fdq -- . ':!tests/spec-tests'; }

gate_nocgo() {
  CGO_ENABLED=0 go build ./... && CGO_ENABLED=0 go test ./crypto/...
}

gate_tooling() {
  git show-ref --quiet --verify refs/heads/main || echo "note: no local 'main' branch" >&2
  (cd libevm/tooling && go test ./... --target_branch="${TARGET_BRANCH:-main}")
}

gate_generate() {
  local raw exclude rc
  raw="$(sed -n 's/^ *EXCLUDE_REGEX: *"\(.*\)"$/\1/p' .github/workflows/go.yml)"
  exclude="${module}/${raw#*/*/}"
  go list ./... | grep -Pv "${exclude}" | xargs go generate
  rc=$?
  if [[ ${rc} -eq 0 ]]; then
    git diff --exit-code --stat
    rc=$?
  fi
  restore
  return "${rc}"
}

gate_tidy() {
  local rc=0
  go mod tidy && (cd libevm/tooling && go mod tidy) || rc=1
  if [[ ${rc} -eq 0 ]]; then
    git diff --exit-code --stat
    rc=$?
  fi
  restore
  return "${rc}"
}

gate_golangci() { golangci-lint run; }
# Tracked files only: CI lints a checkout without submodules or fixtures.
gate_yamllint() {
  git ls-files -z '*.yml' '*.yaml' | xargs -0 yamllint -c .yamllint.yml
}

gate_shell() {
  local files
  mapfile -t files < <(git ls-files 'libevm/*.sh' 'libevm/**/*.sh')
  [[ ${#files[@]} -eq 0 ]] && return 0
  shellcheck "${files[@]}"
}

echo "module: ${module}"
echo "commit: $(git rev-parse --short=9 HEAD)"
bad=0
for g in "${gates[@]}"; do
  start="${SECONDS}"
  if "gate_${g}" >"${logs}/ci-${g}.log" 2>&1; then
    result="pass"
  else
    result="FAIL"
    bad=1
  fi
  printf '%-9s %s (%ss)\n' "${g}" "${result}" "$((SECONDS - start))"
done
[[ "${bad}" -eq 0 ]] && echo "RESULT: CI gates green" || echo "RESULT: CI gates NOT green"
exit "${bad}"
