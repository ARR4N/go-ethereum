# 1. Context

## Goal

Merge upstream geth into libevm, from libevm's current base (geth v1.13.14) to
the latest upstream release tag, such that:

- the result lands in libevm's `main` as a **true merge commit** (never a
  squash), keeping both histories in parallel;
- upstream's side of the history stays on upstream's `master` line, so the merge
  base advances and **every future sync is a plain merge of the libevm and geth
  changes since the last one**;
- the result is provably still geth and provably still libevm
  (see [04](04-objectives-and-oracles.md)).

Then turn the process into a skill for all future syncs.

## Repositories and remotes

| Remote | Repository | Access | Use |
|---|---|---|---|
| `origin` | `ava-labs/libevm` | read only, by decision | Source of truth for libevm; **never push here** |
| `fork` | `ARR4N/go-ethereum` | push | All work is pushed here; the maintainer moves it to libevm |
| (to add) | `ethereum/go-ethereum` | read | Upstream release tags newer than those already fetched |

`ava-labs/libevm` contains two unrelated-looking lines:

- `main`: libevm production branch.
- `master`: mirror of upstream geth `master`.

`ARR4N/go-ethereum` is a GitHub fork of `ethereum/go-ethereum`; `libevm` is in
the same fork network. The maintainer cannot fork libevm separately (one fork
per network per account), hence reusing this one.

## Previous attempt (reference only)

`origin/arr4n/geth-merge-long-lived`: the maintainer's earlier, unfinished sync.
It reached geth v1.14.12 and stopped for unrelated priorities. It is **not**
continued; the sync is redone from the current libevm tip (reasons in
[05](05-process.md#why-redo)). It is kept as a post hoc oracle only: after a version group is
green, compare the new tree with it; never consult it while resolving.

Related older branches: `origin/arr4n/merge-resolve-helpers` (earlier approach:
bulk-accept "theirs" on rename-only conflicts), and
`origin/arr4n/experiment-merge-geth-1.14.11`.

## Design philosophy (from the 2024 "Proposal: libevm")

- libevm is the **EVM as a library**, guarded as if owned by an independent
  maintainer. Consumers are shims; libevm must not contain Ava-specific logic.
- **The geth team is a pseudo-user of libevm.** They change their code assuming
  the rest is intact. Any libevm change `C'` to geth code that a later geth
  feature `F` relies on forces `F'` too; incompatibility grows with its own
  size. So modifications must be minimally destructive and non-brittle.
- Guiding principles (numbering from the proposal):
  1. Independent code base; dependencies incoming only.
  2. Semantic changes MUST be compatible with geth; geth semantics SHOULD be the
     default. Geth tests MAY need minor syntactic changes, MUST pass without
     semantic changes.
  3. Syntactic changes SHOULD be compatible with geth: variadic options, config
     structs, extended ("optional") interfaces with internal type switches.
  4. Changes MUST have their own tests.
  5. Changes MUST have a concrete use by a consumer (no speculative hooks).
  6. How and why changes are made MUST be inspectable (deltas as text patches or
     AST modifications).
  7. Unneeded functionality MUST NOT be removed (removal is a modification);
     MAY be moved to `internal` or flag-disabled.

## Consumers

- `ava-labs/avalanchego` is the primary consumer; it sets the Go version
  (currently 1.26).
- `coreth` and `subnet-evm` will no longer exist by the time this sync lands.
- SAE is owned by the same team.
- Therefore consumer-visible API change being attributable is nice to have,
  not a gate.

## Glossary

| Term | Meaning |
|---|---|
| sync, geth sync | A merge of upstream geth into libevm |
| upstream | `ethereum/go-ethereum` |
| extra | Umbrella for a **hook** (behaviour modification) and a **payload** (new struct field, usually `pseudo.Type`). Usually one type is both |
| transform | One libevm edit to an upstream file, viewed as a mechanical operation (wrap a call, add an argument, ...) rather than as text |
| hook site | Place in an upstream file where a transform is applied |
| group (of files) | See [02](02-file-groups.md) |
| version group | All commits that take the sync from one upstream version to the next; the unit of review |
| checkpoint | The merge commit of a version group; must be green |
| green | Both oracles pass |
| oracle | A test suite that proves identity: upstream's tests ("still geth") and libevm's tests ("still libevm") |
| conjugation | Rename to geth naming, operate, rename back: `g · h · g⁻¹` |
| geth naming / geth space | Tree with Go module `github.com/ethereum/go-ethereum` |
| libevm naming / libevm space | Tree with Go module `github.com/ava-labs/libevm` |
| true cherry-pick | Upstream commit replayed on libevm `main` by fast-forward, preserving author, author date and message |
| pre-emption | A libevm PR that makes, early, a change upstream also makes later (not a cherry-pick; squashed) |
| foreshadowing | A libevm change in the direction upstream later takes, without being the same change (e.g. #280) |
| twin | The upstream commit equivalent to a cherry-pick or pre-emption |
| tracker | `ava-labs/libevm` issue #128: commits to revert before the next sync |
| break point | A release tag whose commits are not on upstream `master` |
