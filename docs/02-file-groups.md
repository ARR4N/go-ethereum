# 2. File groups

Every file in libevm belongs to exactly one of two groups (mutually exclusive,
collectively exhaustive). Ownership is by file, not by where libevm's changes
live: group 2 files are not pristine.

## Group 1: libevm-owned

| Sub-group | Identified by | Incoming upstream change |
|---|---|---|
| 1a. libevm additions | `libevm` in the path: `foo.libevm.go`, `foo.libevm_test.go`, anything under `./libevm/`, `*.libevm.*` testdata | None possible. Never conflicts textually; may break semantically when upstream APIs move |
| 1b. taken-over build and CI files | Explicit list (below) | Owned by libevm. Upstream's change is not applied as-is; libevm decides what to take |

Known 1b files:

- `go.mod`, `go.sum`: **owned, not ignored.** In practice a merge: upstream's
  requirements plus libevm's, libevm free to be ahead. The `go` directive is
  always avalanchego's version (1.26).
- `.github/workflows/go.yml`, `.golangci.yml`: keep ours (these are the two
  files the old `resolve.sh` checked out with `--ours`).
- libevm-added `.github` files: `lint.yml`, `labels.yml`, `rename-module.yml`,
  issue templates, PR template, `.yamllint.yml`.

Known exceptions to the naming rule (libevm-authored, no `libevm` in path):

- `rlp/rlpgen/testdata/alias.in.txt`, `alias.out.txt` (from #154). These were
  also sent upstream; see back burner in [09](09-open-items.md).

A pathspec file listing all of group 1 is to be created in this branch so the
reviewer and the tools share one definition (open item).

## Group 2: upstream-owned

Everything else. Two kinds:

| Kind | Handling in a merge |
|---|---|
| Untouched by libevm | Take upstream's change; merges automatically |
| Carrying libevm transforms (hook sites) | Take upstream's change, then re-seat the transform. **This is the real manual work** |

Further rules:

- A file **deleted upstream is deleted** in libevm too. It cannot be libevm's
  code, by the naming convention.
- A file not labelled `libevm`, not in group 1b, and absent from the merge
  ancestor came from a **cherry-pick of upstream**. Treat it as an upstream
  file: reverting the cherry-pick removes it and merging the twin restores it.
  Examples: `crypto/secp256r1/verifier.go` (#214), the `libsecp256k1` tree
  (#194).

## Comparison baselines differ by group

- Group 2 is judged against the **incoming** upstream version (the future).
- Group 1 (excluding 1b) is judged against **current** libevm `main` (the past).

See [04](04-objectives-and-oracles.md).
