# 7. Baselines and verified facts

Each fact states how it was obtained, so it can be re-checked. Hashes refer to
`ava-labs/libevm` unless noted.

## Anchors

| What | Commit |
|---|---|
| libevm `main` tip when recorded | `a335c2ecf` (#327, 2026-11-11); also `fork/libevm` |
| libevm's upstream base (v1.13.14 on `release/1.13`) | `2bd6bd01d` |
| Merge base of `main` and upstream `master` | `9038ba694` ("params: release Geth v1.13.14") |
| Upstream `master` in `origin` when recorded | `0d061403c` (2026-11-03) |
| Upstream `master` in `fork` when recorded | `2b7829e` (2026-11-20) |
| Old branch tip | `6a455df03` (2025-11-09) |
| Where the old branch left `main` | `1bccf4f2d` (#238, 2025-10-16) |
| Upstream tags present locally | up to v1.16.2 |

**Shallow-clone trap:** in a shallow clone `git merge-base` silently returns
wrong or empty results. Run `git fetch --unshallow` before any history
analysis.

## Rename

<a id="rename"></a>

- Procedure: `tools/rename-module.sh`, copied from
  `.github/workflows/rename-module.yml`.
- Verified at `1bccf4f2d`: forward rename gives a tree identical to the
  maintainer's own rename commit `c16d43565`; forward then back gives the
  original tree.
- Scope: quoted or backticked `github.com/<module>` in `*.go`, `*.txt`,
  `*.go.tpl`, `*.proto`; plus `go mod edit -module`.
- Quirk kept deliberately: the `find` expression's
  `-not -wholename '*/libevm/tooling/*'` binds only to the last `-o` clause, so
  it does not exclude tooling `.go` files. Changing it would break
  reproducibility against existing rename commits.
- Old-branch rename commit messages say `ava-libs/libevm` (typo for
  `ava-labs`); contents are correct.

## Upstream test baseline

<a id="upstream-test-baseline"></a>

libevm's own edits to upstream test files, i.e. not explained by a cherry-pick
or tracker item. **This is the acceptance baseline.**

| File | Change | From | Nature |
|---|---|---|---|
| `eth/tracers/logger/logger_test.go` | `dummyStatedb.GetState`/`SetState` gain `...stateconf.StateDBStateOption`; one import under `// libevm extra imports`; `gofmt` realignment of the neighbouring `GetRefund` line | #205 | Syntactic: test double must keep implementing an interface whose methods went variadic. No test function, assertion or expected value changed |
| `core/chain_makers_test.go` | One line: `r.GasExecuted = 0 //libevm: not persisted so absent from blockchain receipts` | #321 | New libevm receipt field that is deliberately not serialised (serialising would cause far worse upstream conflicts). Described in the PR as libevm's first ever change of an upstream test |
| `rlp/rlpgen/gen_test.go` | Test-name list gains `"alias"`, `"named.libevm"` | #154, #236 | Registers new cases in upstream's table |

Also: `rlp/rlpgen/testdata/alias.{in,out}.txt` added by #154 without `libevm`
in the name (sent upstream; back burner).

No upstream test has had an assertion about geth behaviour changed.

Differences explained by tracker items (not baseline; disappear when the twin
is merged):

| Files | Cause |
|---|---|
| `accounts/abi/unpack_test.go`, `cmd/evm/t8n_test.go`, `metrics/json_test.go`, part of `core/rawdb/accessors_chain_test.go` | #232 (vet fixes to `t.Fatalf`/`t.Logf`) |
| `eth/catalyst/api_test.go`, `eth/handler_eth_test.go`, rest of `accessors_chain_test.go` | #110 |
| `core/state/snapshot/generate_test.go` | #286 (162 changed lines, matching exactly) |
| `core/vm/contracts_test.go`, `params/config_test.go`, `internal/ethapi/api_test.go`, kzg4844 / blobpool / `tx_blob` tests | Cherry-picks (#214, #187, #199) |

Method:

1. Worktree at `main`; revert every true cherry-pick and tracker item that
   reverts cleanly, newest first.
2. Test surface = files at the upstream base matching `_test.go` or
   `/testdata/`, without `libevm` in the path (793 files).
3. Normalise the module path in the worktree; `git diff --stat <base>` over the
   surface.
4. For items that would not revert cleanly, attribute by `git log -- <file>`.

Checks made: no upstream test deleted or renamed; choice of base
(`9038ba694` vs `2bd6bd01d`) makes no difference; only one non-`_test`
test-adjacent file differs (`cmd/evm/internal/t8ntool/block.go`, one line).

## Oracle baseline on untouched libevm

`tools/run-oracle.sh` (`go test -short`, as CI) on `a335c2ecf`, 2026-11-21,
4 cores, Go 1.26.10, fixtures present (`tests/testdata` submodules, execution
spec tests v2.1.0):

| Naming | Result | Time | First-pass failures |
|---|---|---|---|
| libevm | green | 542s (non-flaky 426s, known-flaky 111s) | `eth/catalyst` (known flaky), passed on first re-run |
| geth | green | 450s (non-flaky 358s, known-flaky 92s) | none |

- 193 packages: 129 with tests, 64 without. Identical per-package outcome in
  both namings, so **the oracles can be run in either naming**.
- The `tests` package took about 200s in each run and no fixture-skip message
  appeared, i.e. the fixture-driven tests ran.
- No test failed for environmental reasons in the cloud session.
- A checkpoint therefore costs roughly 8 to 9 minutes of test time here.

## CI gate baseline on untouched libevm

`tools/run-ci.sh` on `a335c2ecf`, 2026-11-23, with `golangci-lint` v1.64.8
built with Go 1.26.10, `yamllint` 1.38.0, `shellcheck` 0.10.0:

| Gate | libevm naming | geth naming |
|---|---|---|
| nocgo | pass (39s) | pass |
| tooling | pass (5s) | **fail** |
| generate | pass (37s) | pass |
| tidy | pass (7s) | **fail** (tooling module) |
| golangci | pass (51s) | pass |
| yamllint | pass | pass |
| shell | pass | pass |

All gates pass in libevm naming, in about two and a half minutes.

**The two failures in geth naming are caused by the rename, not by the code.**
The rename rewrites import paths in `libevm/tooling/**/*.go` but not
`libevm/tooling/go.mod`, which still requires and replaces
`github.com/ava-labs/libevm`. The tooling module therefore cannot build in geth
naming. This is the `find` quirk described under [Rename](#rename): the
exclusion of `libevm/tooling` was evidently intended but does not take effect
for `.go` files.

Consequence: **run the CI gates in libevm naming.** At a checkpoint that is in
geth naming, rename a scratch copy back and run them there. The oracles can
still be run in either naming.

## With and without `-short`

Same tree and environment as above, libevm naming, `go test -count=1 -json
./...` as a single invocation (no flaky/non-flaky split), with and without
`-short`. The full run used `-timeout 90m`; it did not need it.

| | `-short` | full | Difference |
|---|---|---|---|
| Wall time | 376s | 463s | +87s (+23%) |
| `tests` package | 214s | 343s | +129s |
| Test cases run (subtests included) | 384,298 | 412,054 | +27,756 |
| Top-level tests | 1,983 | 1,983 | same set |

Extra test cases that pass in the full run and do not run under `-short`
(27,935):

| Where | Count | What |
|---|---|---|
| `tests`: `TestState`, `TestLegacyState` | 27,862 | Fixture categories marked slow: `stPreCompiledContracts` (13,659), `stTransactionTest` (6,742), `stBadOpcode` (5,138), `stQuadraticComplexityTest` (1,289), `stStaticCall` (630), `stSystemOperationsTest` (350), `stAttackTest` (54) |
| `tests`: `TestBlockchain/ValidBlocks` | 71 | Slow block fixtures |
| `metrics` | 1 | `TestUniformSampleConcurrentUpdateCount` |
| `cmd/devp2p/internal/ethtest` | 1 | `TestEthSuite/LargeTxRequest` |

Still skipped without `-short`: 295 cases, of which 277 are `TestDifficulty`
fixtures; the rest are individually skipped upstream tests.

Failures in these two single-invocation runs, each re-run alone:

| Run | Package | In CI's `FLAKY_REGEX` | Re-run alone |
|---|---|---|---|
| short | `miner` (`TestBuildPayload`) | yes | passed first time |
| full | `accounts/keystore` (`TestUpdatedKeyfileContents`) | yes | failed once, then passed |
| full | `eth/catalyst` (`TestEth2NewBlock`) | yes | passed first time |
| full | `cmd/devp2p/internal/ethtest` (`TestSnapSuite/GetStorageRanges`) | **no** | passed first time |

- The `ethtest` failure is a **flake outside the known list**. The same test
  passed in the `-short` pass and in both earlier baseline runs, so it is not
  specific to the full mode. Flagged to the maintainer.
- Running everything in one invocation produces more flaky failures than CI's
  split; `run-oracle.sh` keeps the split.

## CI facts (`.github/workflows/go.yml` on `main`)

- `go test -short`, split into flaky and non-flaky jobs by package.
- `FLAKY_REGEX`: `ava-labs/libevm/(triedb/pathdb|eth|eth/tracers/js|eth/tracers/logger|eth/tracers/internal/tracetest|accounts/abi/bind|accounts/keystore|eth/downloader|miner|ethclient|ethclient/gethclient|eth/catalyst)$`
- Execution spec tests are downloaded and unpacked into `tests/spec-tests`.
- Separate jobs: no-cgo build (`go test ./crypto/...`), tooling module tests,
  `go generate` cleanliness, `go mod tidy` cleanliness.
- The rename workflow's smoke test: `go build ./...` and
  `go test ./accounts/abi/bind ./rlp/rlpgen` (both depend on the module path).

## Old branch: structure

First-parent history from `1bccf4f2d`, oldest first. G = rename to geth,
L = rename to libevm, M = merge, V = revert.

| Step | Commit(s) |
|---|---|
| G | `c16d43565` |
| M v1.13.15 | `3f7490701` |
| L, V×3 (#29563, #29223, #29050), G | `331726b39`, `0dcc0bda9`, `a16ae8fab`, `270fc3cf9`, `b9b356326` |
| M v1.14.0 | `9179a9fe8` |
| L, V×3 (#29514, #29644, #110), G | `21daaf7d4`, `848398d4b`, `d3211958b`, `b3df56ae4`, `23652fa3e` |
| M v1.14.2, M v1.14.5 | `c567ba0a9`, `291f8663b` |
| L, V (#30001), G | `61f7ee882`, `2f703980e`, `a1c4ee716` |
| M v1.14.6 | `828fb1380` |
| L, V (#155), G | `1eaf09fe0`, `9e3939a97`, `9ddd4b9d0` |
| M v1.14.8, .10, .11, .12 | `0f8a7d997`, `5a852fe36`, `bb398ebf7`, `61b423c28` |
| M `c64cf28f9` with `-s ours` | `473896833` |
| L | `6a455df03` |

- The maintainer believes compilation and tests held at every merge; not
  independently verified, and not required to be.
- Adaptations seen at its v1.14.0 merge in upstream files `main` never touched:
  `fmt.Errorf` vet fixes in `beacon/types/*`; variadic option on `GetState` in
  upstream's new `core/tracing/hooks.go`; an `OnEnter` line in
  `eth/tracers/native/prestate.go`.
- Its `go.mod` at the tip is upstream's requirements plus libevm's, with some
  versions ahead of upstream (e.g. `golang.org/x/crypto`).

## Environment (cloud session)

- Go toolchain 1.26.10 (downloaded per `go.mod`), 4 cores.
- Container is ephemeral: unpushed commits and scratch files are lost on
  reclaim; the conversation survives but is summarised as it grows.
- Commits are signed automatically with an SSH key (no GPG key available).
