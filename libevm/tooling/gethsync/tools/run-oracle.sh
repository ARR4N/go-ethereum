#!/usr/bin/env bash
#
# Run both oracles (upstream's tests and libevm's tests) on a tree, the way CI
# does: `go test -short`, non-flaky packages concurrently and flaky packages
# one at a time. Works in either module naming.
#
# The flaky-package list is read from the tree's own .github/workflows/go.yml
# (FLAKY_REGEX), which is the source of truth, plus the short `extra_flaky`
# list below for packages not yet added there.
#
# A failing package is re-run on its own, up to RETRIES more times:
#   - passes on a re-run and is in FLAKY_REGEX  -> "flaky (known)"; carry on
#   - passes on a re-run and is not in the list -> "FLAKY (NEW)"; flag it
#   - never passes                              -> "FAIL"
#
# Usage: run-oracle.sh <tree> <log-dir>
# Env:   RETRIES (default 2), GOTEST_FLAGS (default "-short")
# Exit:  0 if nothing is "FAIL" or "FLAKY (NEW)", else 1.
#
# Prerequisites, as in CI: submodules checked out (tests/testdata) and the
# execution spec tests extracted into tests/spec-tests. See prepare-oracle.sh.
# Without them the fixture-driven tests are skipped, not failed, so this script
# refuses to run and also checks afterwards that none were skipped for that
# reason.

set -uo pipefail

if [[ $# -ne 2 ]]; then
  echo "usage: $0 <tree> <log-dir>" >&2
  exit 2
fi
tree="$(cd "$1" && pwd)"
mkdir -p "$2"
logs="$(cd "$2" && pwd)"
readonly tree logs
readonly retries="${RETRIES:-2}"
read -r -a flags <<<"${GOTEST_FLAGS:--short}"

cd "${tree}"

# Missing fixtures make tests SKIP, not fail, so a run without them proves
# nothing. Refuse to start; and after the run, check the logs for the skip.
for d in tests/testdata/BlockchainTests tests/testdata/GeneralStateTests tests/spec-tests/fixtures; do
  if [[ ! -d "${d}" ]]; then
    echo "missing test fixtures: ${d} (run prepare-oracle.sh first)" >&2
    exit 2
  fi
done

module="$(go list -m)"
readonly module

# FLAKY_REGEX is written against libevm naming; re-anchor it on this module.
raw="$(sed -n 's/^ *FLAKY_REGEX: *"\(.*\)"$/\1/p' .github/workflows/go.yml)"
if [[ -z "${raw}" ]]; then
  echo "FLAKY_REGEX not found in .github/workflows/go.yml" >&2
  exit 2
fi
# Packages seen to flake that CI's list does not yet include. Each entry is a
# reminder to add it to FLAKY_REGEX in go.yml, after which it can go from here.
#   cmd/devp2p/internal/ethtest: TestSnapSuite/GetStorageRanges, 2026-11-21.
readonly extra_flaky='cmd/devp2p/internal/ethtest'
readonly flaky_regex="${module}/${raw#*/*/}|${module}/(${extra_flaky})\$"

go list ./... >"${logs}/packages.txt"
grep -P "${flaky_regex}" "${logs}/packages.txt" >"${logs}/flaky.txt" || true
grep -Pv "${flaky_regex}" "${logs}/packages.txt" >"${logs}/nonflaky.txt" || true

echo "module:     ${module}"
echo "commit:     $(git rev-parse --short=9 HEAD)$(git diff --quiet HEAD || echo ' (dirty)')"
echo "packages:   $(wc -l <"${logs}/packages.txt") ($(wc -l <"${logs}/flaky.txt") known flaky)"
echo "go:         $(go version | awk '{print $3}')"

start="${SECONDS}"

# 1. Non-flaky packages, concurrently.
xargs go test "${flags[@]}" <"${logs}/nonflaky.txt" >"${logs}/nonflaky.log" 2>&1
grep -E '^(FAIL|---)?\s*FAIL\s+\S+' "${logs}/nonflaky.log" |
  awk '{for (i = 1; i <= NF; i++) if ($i ~ /^github\.com\//) print $i}' |
  sort -u >"${logs}/failed-first-pass.txt"
t_nonflaky=$((SECONDS - start))

# 2. Known-flaky packages, sequentially.
: >"${logs}/flaky.log"
while read -r pkg; do
  if ! go test "${flags[@]}" "${pkg}" >>"${logs}/flaky.log" 2>&1; then
    echo "${pkg}" >>"${logs}/failed-first-pass.txt"
  fi
done <"${logs}/flaky.txt"
t_flaky=$((SECONDS - start - t_nonflaky))

# 3. Re-run every failure on its own.
: >"${logs}/summary.txt"
bad=0
sort -u "${logs}/failed-first-pass.txt" | while read -r pkg; do
  [[ -z "${pkg}" ]] && continue
  safe="${pkg//\//_}"
  verdict="FAIL"
  for ((i = 1; i <= retries; i++)); do
    if go test "${flags[@]}" -count=1 "${pkg}" >"${logs}/retry-${safe}-${i}.log" 2>&1; then
      if grep -qxF "${pkg}" "${logs}/flaky.txt"; then
        verdict="flaky (known), passed on re-run ${i}"
      else
        verdict="FLAKY (NEW), passed on re-run ${i}"
      fi
      break
    fi
  done
  echo "${verdict}: ${pkg}" >>"${logs}/summary.txt"
done
t_retry=$((SECONDS - start - t_nonflaky - t_flaky))

echo "time:       non-flaky ${t_nonflaky}s, flaky ${t_flaky}s, re-runs ${t_retry}s, total $((SECONDS - start))s"
echo "no tests:   $(grep -c 'no test files' "${logs}/nonflaky.log" "${logs}/flaky.log" | awk -F: '{s += $2} END {print s}') packages"
if [[ -s "${logs}/summary.txt" ]]; then
  echo "first-pass failures:"
  sed 's/^/  /' "${logs}/summary.txt"
  grep -qE '^(FAIL|FLAKY \(NEW\))' "${logs}/summary.txt" && bad=1
else
  echo "first-pass failures: none"
fi
if grep -lE "can't find test files|missing test files" "${logs}"/*.log >/dev/null 2>&1; then
  echo "fixture-dependent tests were SKIPPED for want of fixtures; see logs"
  bad=1
fi
[[ "${bad}" -eq 0 ]] && echo "RESULT: green" || echo "RESULT: NOT green"
exit "${bad}"
