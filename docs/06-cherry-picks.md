# 6. Cherry-picks and tracker items

Changes on libevm `main` that upstream also makes later. Each is **reverted
immediately before merging the upstream version containing its twin**, so the
upstream version lands cleanly and libevm's delta contains only what is really
libevm's. Any file the change added disappears with the revert and returns with
the merge.

## Three kinds

| Kind | How it landed on `main` | How to detect |
|---|---|---|
| True cherry-pick | Upstream commit replayed, **fast-forwarded** (no squash, no merge commit). Author, author date and message preserved; subject keeps upstream's PR number; only committer and commit date differ | `tools/find-cherry-picks.sh`: match author + author date + subject against upstream |
| Pre-emption / backport | Squashed libevm PR (committer `GitHub`), rewritten subject, libevm author | Tracker only |
| Foreshadowing | Squashed libevm PR in the direction upstream later takes | Tracker / maintainer |

Sources of truth: metadata matching for true cherry-picks; **issue #128** for
everything else. (#286 should have been a true cherry-pick; it was squashed.)

## Inventory as of `main` = `a335c2ecf` (2026-11-11)

True cherry-picks (14), by tracker item:

| Tracker item | libevm commit | Upstream twin | Twin first in | Subject |
|---|---|---|---|---|
| #187 | `858038850` | `99bbbc027` | v1.14.0 | internal/build, rpc: add missing HTTP response body Close() calls (#29223) |
| #187 | `9f0e1f8da` | `1e9bf2a09` | v1.14.0 | core/state: fix bug in statedb.Copy and remove unnecessary preallocation (#29563) |
| #187 | `b570d5ee0` | `69f815f6f` | v1.14.2 | params: print time value instead of pointer in ConfigCompatError (#29514) |
| #187 | `7c5a8086a` | `e4b8058d5` | v1.14.2 | eth/gasprice: add query limit for FeeHistory to defend DDOS attack (#29644) |
| #187 | `7688f988c` | `34b46a2f7` | v1.14.6 | core/state/snapshot: add a missing lock (#30001) |
| #187 | `955616d7c` | `fa9a2ff86` | v1.15.0 | crypto: add IsOnCurve check (#31100) |
| #187 | `7b6ff3e0f` | `da71839a2` | v1.15.1 | internal/ethapi: fix panic in debug methods (#31157) |
| #191 | `ab5ad25dc` | `0ce13346c` | v1.16.0 | crypto/bn256/cloudflare: pull in upstream fix for R27 and R29 usage (#32057) |
| #194 | `8ae463196` | `5606cbc71` | v1.15.6 | crypto/secp256k1: update libsecp256k1 (#31242). Tracker: revert the pick, **not** the linter fix in the same PR |
| #199 | `08e2b6da8` | `d5bacfa4d` | v1.14.0 | crypto/kz4844: pass blobs by ref (#29050) |
| #214 | `9e4c147bf` | `b3131f00a` | v1.16.2 | core/vm: implement EIP-7951 - precompile for secp256r1 (#31991) |
| #243 | `bb1e421e2` | `d8f980130` | after v1.16.2 (unresolved) | rpc: avoid unnecessary RST_STREAM, PING frames sent by client (#33122) |
| #270 | `a7c86d251` | `be92f5487` | after v1.16.2 (unresolved) | trie: error out for unexpected key-value pairs preceding the range (#33898) |
| #323 (first commit) | `97406059c` | `df3f0a81a` | v1.14.8 | go.mod: upgrade to btcsuite/btcd/btcec v2.3.4 (#30181) |

Other tracker items:

| Tracker item | libevm commit | Kind | Twin | Twin first in | Note |
|---|---|---|---|---|---|
| #110 | `67f879ac1` | Pre-emption; **treat as a cherry-pick** | `2e8e35f2a` (geth #29482) | v1.14.2 (old branch reverted it there) | `Block.WithBody(Body)` signature |
| #155 | `24d7bdeae` | Pre-emption | `e4675771e` (geth #30253) | v1.14.8 | Remove `memsize` |
| #156 | `9f410ecc7` | Go version | — | — | `go@1.23`. Handling undecided now the directive is fixed at 1.26 |
| #232 | `51afedec9` | Go version + vet fixes | — | — | Go 1.24; includes vet fixes to upstream tests. Handling undecided |
| #280 | `9c6496377` | **Foreshadowing: keep, do not revert pre-emptively** | upstream later moves `IntrinsicGas()` towards carrying `params.Rules` | unknown | Expect a conflict at that version; then decide revert vs re-seat and report why. Tracker still lists it for reverting |
| #286 | `2e3b1cb8d` | Squashed backport | `bc1967f08` (geth #33540) | unresolved | Snapshot generator shutdown race |

"Twin first in" comes from `git describe --contains` with tags up to v1.16.2.
Re-run after fetching newer tags.

## Reverts that currently conflict against `main`

Tried in libevm naming at `a335c2ecf`, newest first:

| Item | Conflicting file |
|---|---|
| #323 `97406059c` | `go.mod`, `go.sum` |
| #286 `2e3b1cb8d` | `core/state/snapshot/generate.go` |
| #194 `8ae463196` | `crypto/secp256k1/libsecp256k1/.github/workflows/ci.yml` |
| #155 `24d7bdeae` | `libevm/tooling/release/cherrypicks` |
| #110 `67f879ac1` | `ethclient/ethclient.go` |

All others reverted cleanly. These need hand resolution at their turn.

## Risk to watch

A cherry-pick whose twin is far ahead (e.g. #214 at v1.16.2) sits on top of
every intermediate upstream version until then, and may conflict with upstream
changes before its twin arrives.

## Closing out

At the end every item above is either reverted and replaced by its twin, or
still listed with a reason (e.g. twin is later than the final version).
Suggest the maintainer strikes #280 from #128 or annotates it.
