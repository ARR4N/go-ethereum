# 5. Process

## The pattern: conjugation

libevm and geth differ in Go module path, so a naive merge conflicts on import
lines in ~850 files. Instead:

1. **Rename** the tree to geth naming (`tools/rename-module.sh`).
2. **Merge** upstream there; only real divergence conflicts.
3. **Rename back** to libevm naming.

`g · h · g⁻¹`. The inverse is not applied after every merge: several merges may
be chained in geth space (`g · h₁ · h₂ · h₃ · g⁻¹`). Rename back only when
something needs libevm space, and once at the very end.

The rename is reproducible and exactly invertible (verified, see
[07](07-baselines-and-facts.md#rename)).

## Starting point

<a id="why-redo"></a>

**Redo from the current libevm tip; do not continue or rewind the old branch.**

- The old branch's first merge is v1.13.15, the first break point (below), so
  fixing the sequence means going back to the start anyway.
- Starting from today's tip avoids merging newer `main` commits mid-way, and
  lets every cherry-pick be reverted before its twin (e.g. #323, picked onto
  `main` after the old branch had already passed its twin's version).
- Running the whole process end to end is what produces the skill.
- The old branch remains a cross-check: at each matching version compare trees;
  any difference is an error in one of them. It need not be fully verified.

The maintainer permits rewinding/redoing anything to keep history logical.
Report what changed and why.

## Choosing the upstream sequence (general rule)

1. **Target:** the latest upstream release tag.
2. **The upstream side of every merge is a commit on upstream `master`.**
   - Release tag on `master`: merge the tag.
   - Release tag off `master`: merge the `master` commit whose tree is
     identical to the tag's (confirm with `git diff --quiet`), and say so in
     the merge message.
   - No such commit: skip that release. Its changes arrive with the next
     release that is on `master`; confirm that they do.
3. **One release at a time, oldest first.** Patch releases may be skipped or
   grouped logically. Every release that is merged must be green.
4. **No tags of our own.** The tree states its version (upstream's
   `params/version.go`) and the merge message names the release.

In the steady state libevm merges each geth release as soon as it is out, so
the sequence is a single merge of the new tag.

## Version sequence for the first sync

How the rule above was arrived at. Upstream's release process changed over
time:

| Stretch | Relation to upstream `master` |
|---|---|
| v1.13.15 | On `release/1.13`: v1.13.14 + four fixes cherry-picked from `master` under new hashes. **No `master` commit has this tree** |
| v1.14.0 – v1.14.5 | On `master` |
| v1.14.6 – v1.14.11 | One version-bump commit off `master`; tree otherwise identical to a `master` commit |
| v1.14.12 | 101 commits cherry-picked onto the release branch; tree **identical** to `master` commit `c64cf28f9` |
| v1.15.0 – v1.16.2 | On `master` |
| v1.16.3 – v1.16.7 | One version-bump commit off `master`; tree identical to a `master` commit (v1.16.3 differs in `version/version.go` only) |
| v1.16.8, v1.16.9 | Hotfix releases on the release branch (txpool, ecies, secp256k1 fixes). **No `master` commit has these trees** |
| v1.17.0 – v1.17.10 | On `master` |

**Problem at a break point:** merging an off-`master` tag and later a `master`
descendant shows git the same upstream change twice as unrelated commits.
Where `master` later touched those lines it conflicts upstream-against-upstream
(pure v1.13.15 merged with pure v1.14.0 conflicts in
`eth/protocols/snap/gentrie.go` and `params/version.go`), the duplicates stay
in history, and tricks like `-s ours` merges become necessary.

**Rule: the upstream side of every merge is a commit on upstream `master`.**

- **Skip v1.13.15.** Its four fixes (#29313, #29535, #29534, #29378) are all in
  v1.14.0 (verified). Nothing is lost; the final diff against upstream would
  expose it otherwise.
- **v1.14.6 – v1.14.12:** merge the `master` commit whose tree matches the tag.
  v1.14.12 is kept as a checkpoint via `c64cf28f9` (last stop before v1.15.0).
- **v1.15.0 – v1.16.2 and v1.17.0 onward:** the tag itself.
- **v1.16.3 – v1.16.7:** the matching `master` commit. **Skip v1.16.8 and
  v1.16.9**, which have none; confirm their fixes are in v1.17.0.
- **No tags of our own.** The version relation is carried by (i) the tree
  itself, since upstream's `params/version.go` states the version, and (ii) the
  merge commit message, which names the tag and states the trees are identical
  (checkable with one `git diff`).
- Cost: the v1.13.15, v1.14.6–12 and v1.16.3–9 tags are not ancestors of the
  result.

Candidate `master` commits (from `git merge-base <tag> master`; **re-verify
tree equality with `git diff --quiet` before use**):

| Tag | `master` commit | Note |
|---|---|---|
| v1.14.0 | `87246f3cb` | on master |
| v1.14.2 | `35b2d07f4` | on master |
| v1.14.3 | `ab48ba42f` | on master |
| v1.14.4 | `5550d8399` | on master |
| v1.14.5 | `0dd173a72` | on master |
| v1.14.6 | `09056601d` | differs from tag in `params/version.go` only |
| v1.14.7 | `0aafbb31a` | tree identical |
| v1.14.8 | `880511dc3` | tree identical |
| v1.14.9 | `f321dfa82` | tree identical |
| v1.14.10 | `269551876` | tree identical |
| v1.14.11 | `f14f13bac` | tree identical |
| v1.14.12 | `c64cf28f9` | tree identical (verified) |
| v1.15.0 | `756cca7c6` | on master |
| v1.16.3 | `ffe758c7a` | differs from tag in `version/version.go` only |
| v1.16.4 | `b964b6574` | tree identical |
| v1.16.5 | `32ccb548d` | tree identical |
| v1.16.6 | `044828e66` | tree identical |
| v1.16.7 | `07129d21c` | tree identical |
| v1.16.8, v1.16.9 | none | skip; confirm their fixes are in v1.17.0 |

- There was no v1.14.1 tag. Grouping/skipping patch versions is allowed.
- v1.13.14 → v1.14.0 is 209 `master` commits. Do it as one merge; add an
  intermediate stop at a non-release `master` commit only if the changes prove
  too significant to take together.
- **End point:** the latest upstream release tag (v1.17.10 as of 2026-11-22).
  Upstream `master` between v1.13.14 and v1.17.10 is 2,576 commits.
- The sync branch is named `gethsync`.

## One version group

1. **Revert** the cherry-picks and tracker items this version supersedes
   ([06](06-cherry-picks.md)), just in time, never earlier (earlier would
   break the green rule at intermediate versions).
2. **Rename to geth naming**, if not already there.
3. **Merge commit** (the checkpoint; must be green).
4. **Follow-up commits**: hook coverage for upstream's new code paths. Green
   again afterwards.
5. Push to the fork.

Then straight to the next group. Many merges in a row is normal when nothing
needs reverting.

### Reverts

- **Done in geth space**, to avoid two no-information 850-file rename commits
  per round trip. Tested: 5 of 6 reverted cleanly there; the failure was a hunk
  adjacent to an import line.
- **Fallback:** if a revert conflicts in geth space, rename back, revert in
  libevm space, rename forward.
- **Check (to be written as a reviewer script):** each revert is the exact
  inverse of its cherry-pick. In geth space, pass the original commit's patch
  through the rename `sed` first, then compare.
- Reverts of non-cherry-picks (pre-emptions, backports) may conflict and need
  hand resolution; several already do against today's `main`
  ([06](06-cherry-picks.md)).

<a id="diff3"></a>

### diff3 conflict style is mandatory

Every merge, revert and cherry-pick in a sync is run with
`merge.conflictStyle=diff3` (set it in the repository config before starting:
`git config merge.conflictStyle diff3`). Each conflict then shows three
sections, ours, the common ancestor and theirs, not two.

The ancestor is what makes a hook-site conflict resolvable by intent: comparing
it with "ours" shows exactly which transform libevm applied, and comparing it
with "theirs" shows what upstream changed underneath. With two-way markers
those have to be guessed.

### Merge commit contents

<a id="merge-commit-contents"></a>

One atomic commit containing:

- **conflict resolutions** in group 2 files (re-seated transforms), and
- **adaptations** to group 1 files forced by upstream changes.

Because the groups are mutually exclusive,
`git show --remerge-diff <merge>` filtered by the group 1 pathspec separates
the two exactly. The merge commit therefore builds and passes by itself.

- Conflict resolutions must be **obvious**. They are the agent's decisions, to
  be owned and explained.
- **Edits to upstream files that merged without conflict** (e.g. a hook site
  calling something upstream renamed elsewhere; an interface implementer
  needing a variadic parameter; a `vet` fix) are distinguishable in
  `--remerge-diff` (no conflict markers removed). Keep them rare and **list
  them in the commit message**.
- The commit message body lists, grouped by file: each adaptation and the
  upstream change that forced it; each non-conflict upstream-file edit; the tag
  and tree-equality statement.

### go.mod and Go version

<a id="go-version"></a>

- `go` directive: **always 1.26** (avalanchego's version), at every checkpoint.
  Using upstream's incoming version per checkpoint was considered and rejected:
  libevm's dependencies require Go >= 1.24 (verified by a failed build at 1.22).
- Consequence: `vet` rejects some older upstream code, e.g. non-constant format
  strings (`fmt.Errorf("… " + s)`), which it only flags when `go.mod` declares
  ≥ 1.24. Fix minimally in the upstream file; list in the commit message. The
  edit disappears at the version where upstream makes the same fix.
- `go.mod` / `go.sum`: take what upstream needs, keep what libevm needs.

## Review

<a id="review"></a>

- A human reviews, not necessarily in the GitHub UI.
- **Every version group must be reviewable in isolation.** In practice the last
  group is reviewed in full and a random part of a random subset of interim
  groups is spot-checked. Build each as if it will be picked.
- Mechanical steps (renames, reverts) should be reproducible by script so a
  reviewer can compare tree hashes. Nice to have; unlikely to be run.

## Final step

Merge the sync branch into libevm `main` as a true merge commit. Never squash.

`main` does not allow merge commits through the GitHub UI. The maintainer
instead fast-forwards `main` locally and pushes, which is permitted provided
there is an associated PR with all requirements met.

This applies to the sync branch only. The `sync-tooling` branch will also land
on `main` eventually, but as an ordinary squashed PR: its history is
irrelevant.
