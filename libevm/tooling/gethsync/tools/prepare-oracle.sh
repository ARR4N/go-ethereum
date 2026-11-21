#!/usr/bin/env bash
#
# Fetch the test fixtures CI uses, so run-oracle.sh exercises the same tests:
#   - git submodules (tests/testdata, tests/evm-benchmarks);
#   - execution spec tests, extracted into tests/spec-tests.
#
# The spec-test version and file name are read from the tree's own
# .github/workflows/go.yml; the checksum from build/checksums.txt. Idempotent.
#
# Usage: prepare-oracle.sh <tree>

set -euo pipefail

if [[ $# -ne 1 ]]; then
  echo "usage: $0 <tree>" >&2
  exit 2
fi
cd "$1"

git submodule update --init --recursive --depth 1

yml=".github/workflows/go.yml"
version="$(sed -n 's/^ *EXECUTION_SPEC_TESTS_VERSION: *//p' "${yml}")"
file="$(sed -n 's/^ *EXECUTION_SPEC_TESTS_FILE: *//p' "${yml}")"
if [[ -z "${version}" || -z "${file}" ]]; then
  echo "spec-test version or file not found in ${yml}" >&2
  exit 1
fi

if [[ -d tests/spec-tests/fixtures ]]; then
  echo "tests/spec-tests already populated; skipping download" >&2
  exit 0
fi

tmp="$(mktemp -d)"
trap 'rm -rf "${tmp}"' EXIT
curl -fsSL -o "${tmp}/${file}" \
  "https://github.com/ethereum/execution-spec-tests/releases/download/${version}/${file}"
(cd "${tmp}" && sha256sum --ignore-missing --check "${OLDPWD}/build/checksums.txt" | grep "${file}: OK")
mkdir -p tests/spec-tests
tar -xz -f "${tmp}/${file}" -C tests/spec-tests
echo "spec tests ${version} extracted" >&2
