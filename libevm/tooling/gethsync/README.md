# libevm ⇄ geth sync: tooling

Everything needed to merge upstream `go-ethereum` ("geth") into `libevm`: the
agreed objectives, the process, the facts established so far, and the tools.
This branch is the seed of a future skill for all later syncs.

Status (2026-11-21): **agreements recorded; the sync itself has not started.**
Do not start it without an explicit go-ahead from the maintainer.

## How to use this directory

These documents give **context, not an algorithm**. They say what a correct
sync is, why each rule exists, and what was decided before. Use that context to
make decisions. The rule for when to decide and when to stop and ask is in
[docs/08-rules-of-engagement.md](docs/08-rules-of-engagement.md#decide-or-stop);
read it first.

The commit history of the first sync done this way is the worked example.

This directory currently also holds material specific to the first (catch-up)
sync and to the session that planned it. It is to be cleaned up at the end; see
[docs/09-open-items.md](docs/09-open-items.md#cleanup).

## Read in this order

| File | Contents |
|---|---|
| [docs/01-context.md](docs/01-context.md) | Goal, repositories, branches, glossary, design philosophy |
| [docs/02-file-groups.md](docs/02-file-groups.md) | The two file groups and how each is treated in a merge |
| [docs/03-extras-and-transforms.md](docs/03-extras-and-transforms.md) | How libevm hooks into upstream files, and the rules for doing so |
| [docs/04-objectives-and-oracles.md](docs/04-objectives-and-oracles.md) | What a correct sync is, and how it is proven |
| [docs/05-process.md](docs/05-process.md) | The procedure: sequence, renames, reverts, merge commits, green rule |
| [docs/06-cherry-picks.md](docs/06-cherry-picks.md) | Commits to revert before their upstream twin is merged |
| [docs/07-baselines-and-facts.md](docs/07-baselines-and-facts.md) | Measured baselines and verified facts, with how each was obtained |
| [docs/08-rules-of-engagement.md](docs/08-rules-of-engagement.md) | When to stop and ask; access, signing and pushing rules |
| [docs/09-open-items.md](docs/09-open-items.md) | Undecided questions, unverified assumptions, back burner, lessons |
| [docs/10-catalogue.md](docs/10-catalogue.md) | Every libevm change to an upstream file: where, why, and how to reproduce it |

## Tools

| Tool | Purpose | State |
|---|---|---|
| [tools/rename-module.sh](tools/rename-module.sh) | Flip the Go module path between libevm and geth naming | Verified, see docs/07 |
| [tools/find-cherry-picks.sh](tools/find-cherry-picks.sh) | List true cherry-picks on libevm and their upstream twins | Verified, see docs/07 |

## Branch names (in the working fork)

| Name used in discussion | Ref | Meaning |
|---|---|---|
| libevm | `fork/libevm` | Copy of `ava-labs/libevm@main` |
| geth, master | `fork/master` | Upstream `ethereum/go-ethereum@master` |
| merge, sync | `fork/gethsync` (not yet created) | The new long-lived merge branch |
| tooling | `fork/sync-tooling` | This branch: libevm plus `libevm/tooling/gethsync/` |

"A sync" or "a geth sync" means a merge of upstream into libevm.
