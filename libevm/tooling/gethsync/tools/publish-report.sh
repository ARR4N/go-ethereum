#!/usr/bin/env bash
#
# Publish a status report to the root of the `gethsync-blocked` branch, as
# GETHSYNC_REPORT.md, overwriting the previous one.
#
# The report is a single extra commit on top of the tree it describes, so its
# relative links (path#Lnn) resolve for anyone who pulls the branch. The commit
# is made with plumbing: the working tree, index and current branch are not
# touched, and the report never reaches the `gethsync` branch itself.
#
# Usage: publish-report.sh <report.md> [<tree-ish to describe, default HEAD>]
# Env:   REMOTE (default "fork"), BRANCH (default "gethsync-blocked")

set -euo pipefail

if [[ $# -lt 1 || $# -gt 2 ]]; then
  echo "usage: $0 <report.md> [<commit>]" >&2
  exit 2
fi
readonly report="$1"
readonly remote="${REMOTE:-fork}"
readonly branch="${BRANCH:-gethsync-blocked}"
readonly name="GETHSYNC_REPORT.md"

base="$(git rev-parse --verify "${2:-HEAD}^{commit}")"
readonly base

index="$(mktemp)"
trap 'rm -f "${index}"' EXIT

blob="$(git hash-object -w "${report}")"
GIT_INDEX_FILE="${index}" git read-tree "${base}"
GIT_INDEX_FILE="${index}" git update-index --add --cacheinfo "100644,${blob},${name}"
tree="$(GIT_INDEX_FILE="${index}" git write-tree)"

commit="$(git commit-tree -S "${tree}" -p "${base}" -m "report: status of gethsync at $(git rev-parse --short=9 "${base}")

Not part of the sync. Overwritten with every report; never merged into
gethsync.")"

git push --quiet --force "${remote}" "${commit}:refs/heads/${branch}"
echo "published ${name} as ${commit} on ${remote}/${branch} (describes ${base})"
