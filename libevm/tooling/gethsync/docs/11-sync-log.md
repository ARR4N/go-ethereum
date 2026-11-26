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

## Group 2: v1.14.0 → v1.14.4

Upstream side: tag `v1.14.4` (`5550d8399`, on `master`), 122 commits. v1.14.2
and v1.14.3 were not merged separately; the upstream bug that left group 1
partly red is only fixed in v1.14.4.

### Reverts before the merge

Done in geth naming.

| Item | Result |
|---|---|
| #29644 (fee-history query limit) | Did not apply mechanically, because its line had been adapted in group 1. File restored to upstream's v1.14.0 contents |
| #29514 (ConfigCompatError) | One conflict on an import line next to the reverted one (libevm naming vs geth naming); resolved by removing the import |
| #110 (`WithBody(Body)`) | Hand-resolved in `ethclient.go`; two files already removed upstream; `beacon/types/exec_payload.go` restored to upstream's v1.14.0 contents. The tree does not build between this commit and the merge |

### The merge

12 conflicts. The substantial ones were upstream rewrites under libevm hooks:
the trie prefetcher (state-6, state-7), `StateDB.Commit` (state-5) and
`Block.WithBody` (types-1). See the catalogue entries for the re-seated forms
and the merge commit message for the full list.

### Maintainer decisions

| Question | Decision |
|---|---|
| When to release the worker pool now that prefetcher termination can be asynchronous | After every sub-fetcher finishes, in a goroutine when asynchronous. Approved because it mirrors upstream's new mechanism |
| Deep or shallow `copy()` on libevm's journal entry | Shallow for now. A known limitation, deferred because the only `StateAccountExtra` in production is a `bool`. To be addressed with the rest of the paused work on branch `arr4n/clonable-state-account-extras`; a TODO in the code says so |
| Should `NewBlock` carry the body's payload? | It MUST be cloned, to match `NewBlock`'s documented "the input data is copied". Done as a follow-up commit; test in `*.libevm_claude_test.go` |
| `stringer` is not on CI's PATH | Use `go run` in the directive, matching upstream's later line |

### Findings

- Upstream fixed its own v1.14.0 bug here, so
  `TestOverrideBlockHash_StandardTraceBlockToFile` passes again.
- A revert in geth naming can conflict on import lines adjacent to the
  reverted change, and on lines adapted in an earlier group. Both are easy to
  resolve by aiming for upstream's contents at the previous checkpoint.
- Upstream's freezer inspection creates a directory at whatever path it is
  given; test stubs must hand it a temporary one.
- Generated files must be regenerated with the tool version in `go.mod`;
  upstream's committed output can differ.
- `git show --remerge-diff --stat <merge>` is a reliable way to list every
  file a merge commit touched by hand.
- **Mistake, corrected:** upstream's `prefetch` began returning an error, and
  the linter flagged two libevm test calls that ignored it. They were first
  "fixed" by discarding the error with `_ =`, on the reasoning that adding an
  assertion would be a semantic change to a libevm test. The maintainer
  rejected that outright. Both now use `require.NoError`. Discarding an error
  is the worse semantic change, and doubt about a rule is a reason to ask.

## Group 3: v1.14.4 → v1.14.5

Upstream side: tag `v1.14.5` (`0dd173a72`, on `master`), 9 commits. No reverts
needed, no conflicts. The only hand edit is `LibEVMVersion`.

### Comparison with the old branch at v1.14.5 (post hoc)

Differences not explained by `main` having moved on since the old branch point:

| File | Old branch | This sync | Outcome |
|---|---|---|---|
| `core/vm/contracts.go`, `eth/tracers/native/prestate.go` | — | — | Order of declarations and a comment only |
| `core/state/statedb.go` | Commit options also passed to the internal `commit()`, unused there | Stop at `commitAndFlush` | Equivalent |
| `core/state/state.libevm.go` | Journal entry deep-copied via a `SelfCloner` constraint on `state.SetExtra` | Shallow, with a TODO | Maintainer chose shallow for the sync; see below |
| `libevm/legacy` | `PrecompileEnvironment.UseGas` gains tracer and reason parameters | Exported signature unchanged | Maintainer prefers this sync's |
| `log/value.libevm.go` | Still `golang.org/x/exp/slog` | `log/slog` | Maintainer prefers this sync's. The old one implements the wrong package's `LogValuer` for upstream's logger |
| `core/rawdb/examples.libevm_test.go` | Stub returns `""`, read by upstream as an in-memory freezer; expected output gains five state-freezer rows | Temporary directory; output unchanged | Maintainer is content with the temporary directory |

On cloning account payloads: the old branch's constraint was on
`state.SetExtra` only, which breaks that exported function for payload types
without a `Clone` method (the production payload is a `bool`) while
`StateAccount.Copy` stays shallow. The paused branch
`arr4n/clonable-state-account-extras` instead constrains
`types.RegisterExtras`, which lets both sites clone. Either is a breaking
change to libevm's API and is not part of the sync.

**Follow-up owed:** deep-copying account payloads, tracked on the back burner
in [09](09-open-items.md). The maintainer's ruling: keep the TODO during the
sync, then bring the change in as a follow-up commit using
`arr4n/clonable-state-account-extras` as the template.
