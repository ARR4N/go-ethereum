#!/usr/bin/env bash
#
# List true cherry-picks on libevm: commits whose author, author date and
# subject all match an upstream commit that libevm does not yet contain.
#
# This finds only true cherry-picks (fast-forwarded, metadata preserved).
# Squashed pre-emptions and backports are NOT found; those are tracked in
# ava-labs/libevm issue #128. See docs/06-cherry-picks.md.
#
# Requires full history: run `git fetch --unshallow` first in a shallow clone.
#
# Usage:  find-cherry-picks.sh <libevm-ref> <upstream-ref>
# Output: <libevm commit> <upstream twin> <first upstream tag or -> <subject>
#         oldest libevm commit first.

set -euo pipefail

if [[ $# -ne 2 ]]; then
  echo "usage: $0 <libevm-ref> <upstream-ref>" >&2
  exit 2
fi
readonly libevm="$1" upstream="$2"

if [[ "$(git rev-parse --is-shallow-repository)" == "true" ]]; then
  echo "shallow clone: history analysis would be wrong; unshallow first" >&2
  exit 1
fi

base="$(git merge-base "${libevm}" "${upstream}")"
readonly base

tmp="$(mktemp)"
trap 'rm -f "${tmp}"' EXIT

# Tab-separated: author name, author timestamp, subject, hash.
readonly fmt='%an%x09%at%x09%s%x09%H'
git log --format="${fmt}" "${base}..${upstream}" >"${tmp}"

git log --reverse --format="${fmt}" "${base}..${libevm}" |
  awk -F'\t' '
    NR == FNR { up[$1 FS $2 FS $3] = $4; next }
    ($1 FS $2 FS $3) in up { print $4 "\t" up[$1 FS $2 FS $3] "\t" $3 }
  ' "${tmp}" - |
  while IFS=$'\t' read -r pick twin subject; do
    tag="$(git describe --contains --match 'v[0-9]*' "${twin}" 2>/dev/null | sed 's/[~^].*//' || true)"
    printf '%s %s %s %s\n' "${pick:0:9}" "${twin:0:9}" "${tag:--}" "${subject}"
  done
