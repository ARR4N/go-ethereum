# 11. Log of the first sync

One section per version group. Records what the rules alone do not explain:
decisions the maintainer made, exceptions and their reasons, and findings worth
knowing next time. The commits themselves are the worked example.

Exceptions recorded here are **not precedent**; see
[08](08-rules-of-engagement.md).

## Group 1: v1.13.14 → v1.14.0

Upstream side: tag `v1.14.0` (`87246f3cb`, on `master`), 209 commits.

### Reverts before the merge

Done in geth naming.

| Cherry-pick | Result |
|---|---|
| #29050 (kzg4844 blobs by ref) | Clean |
| #29563 (statedb.Copy) | Clean |
| #29223 (HTTP body Close) | **Partial.** Its line in `rpc/http.go` was later rewritten by the still-live cherry-pick of geth #33122 (twin in v1.17.0), so it stays as #33122 left it; the merge keeps it too |

### The merge

15 conflicts, all in group 2 files or `go.mod`/`go.sum`.

| File | Resolution |
|---|---|
| `core/vm/evm.go`, `core/vm/contracts.go` | Re-seated vm-3 (upstream added a `logger` parameter to `RunPrecompiledContract`; mirrored on the method and passed at all four call sites), vm-7 (rename-and-wrap on upstream's new bodies) and kept both sides' new precompile maps |
| `core/vm/interface.go` | vm-9 plus upstream's new `GetStorageRoot` |
| `core/state/statedb.go` | state-3 beside upstream's new `logger` field; state-5 on upstream's changed `triestate.New` call |
| `core/state_transition.go` | core-6 after upstream's tip credit, which gained a tracing reason |
| `eth/tracers/api.go` | tracers-1 where upstream deleted the neighbouring line |
| `eth/filters/filter.go`, `filter_system.go` | Took upstream's deletion of `pendingLogs` and light-client mode; filters-2 is down to one site |
| `eth/handler_eth_test.go`, `eth/fetcher/block_fetcher.go` | Deleted upstream; our only change was a #110 call site |
| `cmd/geth/main.go` | Upstream's new signature, without the `memsize` line (#155 still live) |
| `rpc/http.go` | Kept #33122's line (see above) |
| `crypto/secp256k1/.../secp256k1.h` | Kept the newer library from cherry-pick #194 |
| `go.mod`, `go.sum` | Upstream's requirements plus libevm's, higher version wins, then `go mod tidy` |

Edits to upstream files that did not conflict:

| File | Why |
|---|---|
| `core/tracing/hooks.go` | tracing-1 |
| `eth/tracers/native/prestate.go` | tracers-2 |
| `core/vm/evm.go` | vm-8: `:=` became `=` because upstream gave `createCommon` named results |
| `beacon/types/beacon_block.go`, `exec_header.go` | vet-1 |
| `beacon/types/exec_payload.go` | A `WithBody` call adapted to the `Body`-taking signature, because #110 is live until v1.14.2 |
| `core/state_processor_test.go` | core-2: two `IntrinsicGas` calls |
| `eth/gasprice/feehistory.go` | Live cherry-pick #29644 brought to its upstream twin's form (two more return values) |

Adaptations to libevm files, all syntactic unless noted:

- `log/slog` replaces `golang.org/x/exp/slog` (five files); upstream switched.
- `params.BeaconRootsStorageAddress` → `BeaconRootsAddress`.
- New arguments upstream now requires: balance-change reasons on
  `SetBalance`/`AddBalance`, `NewFilterAPI` without its `lightMode` flag,
  `vm.Config.Tracer` as `*tracing.Hooks`, `StateAtTransaction`'s new return
  type, the test shim for `RunPrecompiledContract`.
- `LibEVMVersion` → `1.14.0-0.4.0.beta`, as its own rule requires.
- `libevm/tooling` module tidied (in libevm naming).
- **Rewritten, not just respelled:** the test tracer in
  `eth/tracers/api.libevm_test.go`, rebuilt on upstream's new `Tracer` struct.
  Assertions unchanged.
- **Restored:** the prestate-tracer fix (#64); see tracers-2. The test now
  calls `OnTxStart` first, because the tracer takes its state database from
  there.

### Maintainer decisions

| Question | Decision |
|---|---|
| `TestBloomOverride/pendingLogs` tests code upstream removed | Delete. Pending logs do not make sense in SAE |
| `TestBloomOverride/lightFilterLogs` likewise | Delete as obsolete, because upstream removed the feature |
| Restore #64 with a one-line edit to upstream's `prestate.go` | Yes |
| Widen `tracing.StateDB.GetState` | Yes; same rationale as ignored options on `snapshot.Tree` |
| Should tracers see libevm's gas and balance adjustments? | Yes, everything, with an appropriate reason where one exists; done now, not deferred |

<a id="tracer-reasons"></a>

### Tracer reasons for libevm adjustments

A follow-up commit after the merge; novel work, tests in
`*.libevm_claude_test.go`.

| Adjustment | Reason |
|---|---|
| Gas a stateful precompile consumes via `env.UseGas` | `GasChangeCallPrecompiledContract` |
| Gas for a precompile's outbound call | `GasChangeCallOpCode` |
| Gas returned from that call | `GasChangeCallLeftOverRefunded` |
| Base fee credited to the coinbase | `BalanceIncreaseRewardTransactionFee` |
| Refund reversed when `ShouldRefundGas` is false | `GasChangeTxRefunds` |
| Minimum gas consumption | `GasChangeLibEVMMinimumConsumption` (= unspecified, TODO) |
| Preprocessing gas charge | `GasChangeLibEVMPreprocessing` (= unspecified, TODO) |

Balance changes reach a tracer only if it is attached to the state database
with `SetLogger`; the EVM's `Config.Tracer` alone does not see them.

### Exception: this checkpoint is not fully green

`TestOverrideBlockHash_StandardTraceBlockToFile` (libevm, `eth/tracers`) fails
at v1.14.0.

- **Cause:** an upstream bug. `standardTraceBlockToFile` calls the tracer's
  `OnTxEnd` hook without checking it is set, and the JSON logger does not set
  it. Upstream fixed it in v1.14.4 (geth #29795). No upstream test covered the
  function until much later, so upstream's suite never saw it.
- **Not fixed early:** the fix cannot be cherry-picked at v1.14.0; the commit
  also uses hooks added in v1.14.2.
- **Maintainer's ruling:** keep the v1.14.0 checkpoint anyway, (a) as an audit
  trail of the work and (b) because of these specific circumstances.
  **This MUST NOT be treated as precedent.** Circumstance (b) may not be mapped
  onto any future situation; always ask.

### Findings

- A reference merge by someone else is a post hoc oracle only. It was
  consulted once during this merge, for three compile errors; the maintainer
  ruled that out.
- A libevm fix can be disconnected without any compile error (tracers-2).
  Only its test caught it.
- libevm tests can fail because of upstream bugs that upstream's own tests do
  not exercise.
- CI gates need libevm naming; the oracles do not.
