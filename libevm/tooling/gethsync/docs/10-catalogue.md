# 10. Catalogue of libevm changes to upstream files

Every change libevm makes to a group 2 (upstream-owned) file, grouped by
package and then by logical change. For each: where, why, and the mechanical
steps to reproduce it on a fresh copy of the upstream file.

Use during a merge to check that each re-seated change still matches its
rationale and stays as close as possible to its original mechanical form.

**Status: inferred by the agent, awaiting maintainer review.** Rationale is
taken from PR descriptions where one exists and from reading the code
otherwise; steps are reverse-engineered from the diff.

## Scope and method

- Baseline: libevm `main` at `a335c2ecf` against its upstream base
  (`2bd6bd01d`, geth v1.13.14), compared in geth naming.
- True cherry-picks and tracker items #110, #155, #194 were reverted first, so
  they do not appear. Not reverted, and handled as noted: #286 (see
  [snapshot](#corestatesnapshot)), #323 (`crypto/signature_nocgo.go` and
  `go.mod` are entirely that cherry-pick), #232 (see [vet fixes](#vet)), #280
  (see [core-2](#core-2)).
- 58 upstream files differ. libevm-owned files (`*.libevm.go`, `libevm/…`) are
  not catalogued; they are named only as the destination of a transform.
- "(PR #n)" refers to `ava-labs/libevm`.

To regenerate the raw diff: revert the tracker items in a scratch worktree,
run `tools/rename-module.sh`, then
`git diff <upstream base> -- <files without "libevm" in the path>`.

## Shapes

Names used under "Steps".

| Shape | Meaning |
|---|---|
| rename-and-wrap | Lower-case the upstream function's name in place; a `*.libevm.go` file defines the original exported name and calls it |
| method-shadow | Turn a package-level function into a method on a libevm type with the same name, so call sites change only by a receiver |
| variadic option | Append `opts ...SomeOption` to a signature and thread it to where it is consumed |
| payload field | Append a field (usually `extra *pseudo.Type`) as the last field of a struct, after a blank line |
| hook call | Insert a single call into a libevm function at a fixed point |
| delegate return | Replace a final `return x` with a return that passes through a hook |
| interface widening | Replace a concrete type in a signature or type assertion with a libevm interface it satisfies |
| regenerate | Do not edit; re-run `go generate` after changing its input |

---

## Cross-package

<a id="vet"></a>

<a id="vet-1"></a>

### vet-1: fixes required by the Go directive (PR #232)

| File | Change |
|---|---|
| `accounts/abi/unpack_test.go` | `t.Fatalf("invalid hex: %s" + test.enc)` → `t.Fatalf("invalid hex: %s", test.enc)` |
| `cmd/evm/t8n_test.go` (×2) | `t.Logf(string(have))` → `t.Log(string(have))` |
| `cmd/rlpdump/main.go` | `fmt.Fprintf(out, ws(depth)+"[]")` → `fmt.Fprint(…)` |
| `core/rawdb/accessors_chain_test.go` (×2) | `t.Fatalf(err.Error())` → `t.Fatal(err.Error())` |
| `metrics/json_test.go` | `t.Fatalf(s)` → `t.Fatal(s)` |
| `core/asm/tokentype_string.go` | Regenerated `stringer` output (newer `stringer` emits an `idx := int(i) - 0` form) |

- **Rationale:** libevm's `go` directive is ahead of upstream's. From Go 1.24
  `vet` (run by `go test`) rejects a non-constant format string with no
  arguments. Not behaviour changes.
- **Steps:** build and `go vet ./...`; for each "non-constant format string"
  report, switch to the non-formatting variant or add the missing verb
  argument. Regenerate `stringer` files with `go generate`.
- **Added at geth v1.14.0:** `beacon/types/beacon_block.go` and
  `beacon/types/exec_header.go`, `fmt.Errorf("unsupported fork: " + forkName)`
  → `fmt.Errorf("unsupported fork: %s", forkName)`, matching upstream's own
  later fix.
- **At a merge:** upstream makes the same fixes when it raises its own Go
  version. Identical fixes merge cleanly; where they differ, take upstream's.
  New upstream code may need new fixes of the same kind; list them in the merge
  commit message.

<a id="readme-1"></a>

### readme-1: `README.md`

- **Change:** libevm title, badges and description prepended; geth's badges
  removed; a licensing paragraph for libevm additions and modifications added.
- **Rationale:** identify the fork and its licence terms.
- **Steps:** keep libevm's header block and licence paragraph; take upstream's
  changes to the rest of the file.

---

## `core`

<a id="core-1"></a>

### core-1: expose the header to the EVM block context (PR #27)

- **Where:** `core/evm.go`, `NewEVMBlockContext`.
- **Change:** add `Header: header,` as the last entry of the returned
  `vm.BlockContext` literal.
- **Rationale:** stateful precompiles reach block information through
  `vm.PrecompileEnvironment`; the header is carried on the block context
  (see [vm-5](#vm-5)).
- **Steps:** payload field population: one line in the struct literal.

<a id="core-2"></a>

### core-2: `IntrinsicGas` takes `params.Rules` (PR #280)

- **Where:** `core/state_transition.go` (definition and the call in
  `transitionDb`); callers in `cmd/evm/internal/t8ntool/transaction.go`,
  `core/bench_test.go`, `core/txpool/validation.go`,
  `tests/transaction_test_util.go`.
- **Change:** signature
  `(data, accessList, isContractCreation, isHomestead, isEIP2028, isEIP3860 bool)`
  → `(data, accessList, isContractCreation bool, rules params.Rules)`.
- **Rationale (PR):** the access-list gas hook ([core-3](#core-3)) lives on
  `params.RulesHooks`, and `Rules` carry the extras, so `IntrinsicGas` needs
  the `Rules`. **A foreshadowing:** upstream later moves this function towards
  carrying `params.Rules`. Keep for now; expect a conflict when that arrives,
  then decide between re-seating and adopting upstream's form.
- **Steps:**
  1. Replace the three trailing `bool` parameters with `rules params.Rules`.
  2. In the body: `isHomestead` → `rules.IsHomestead`; `isEIP2028` →
     `rules.IsIstanbul` (comment: Istanbul introduced EIP-2028); `isEIP3860` →
     `rules.IsShanghai` (comment: Shanghai introduced EIP-3860).
  3. Callers build a `Rules` in place of the booleans:
     - `transitionDb`: pass the existing `rules`.
     - t8ntool: `chainConfig.Rules(common.Big0, true, 0)`.
     - txpool: `opts.Config.Rules(head.Number, true, head.Time)`.
     - `bench_test.go`: `params.Rules{}`.
     - `tests`: `params.Rules{IsHomestead: isHomestead, IsIstanbul: isIstanbul}`.
- **Observation for review:** in `core/txpool/validation.go` upstream passed a
  literal `true` for `isHomestead`; the libevm form derives it from the chain
  config at the head block. These differ only for a config that is
  pre-Homestead at head.
- **Violates measure 2** (changes an upstream signature). Known and accepted.
- **Added at geth v1.14.0:** two calls in the upstream test
  `core/state_processor_test.go`,
  `IntrinsicGas(code, nil, true, true, true, true)` →
  `IntrinsicGas(code, nil, true, params.Rules{IsHomestead: true, IsIstanbul: true, IsShanghai: true})`.

<a id="core-3"></a>

### core-3: access-list gas hook (PR #282)

- **Where:** `core/state_transition.go`, `IntrinsicGas`, the
  `if accessList != nil` block.
- **Change:** the two upstream lines adding per-address and per-storage-key gas
  become the `else` branch of a check on `libevmAccessListGas(gas, accessList, rules)`;
  an error returns `0, err`; `override` adds the hook's gas instead.
- **Rationale:** let a chain price access lists differently
  (`params.RulesHooks.AccessListGas`). Default is the upstream calculation.
- **Steps:** wrap the upstream lines, unchanged and marked
  `//libevm: upstream original`, in the final `else`. Logic lives in
  `core/state_transition.libevm.go`.

<a id="core-4"></a>

### core-4: state-transition options (PR #305)

- **Where:** `core/state_transition.go`: `ApplyMessage`, `NewStateTransition`,
  `StateTransition` struct. Caller: `eth/gasestimator/gasestimator.go`.
- **Change:** both functions gain `opts ...StateTransitionOption`;
  `ApplyMessage` forwards them; the struct gains `opts []StateTransitionOption`
  and the constructor literal sets `opts: opts`. The gas estimator passes
  `core.WithoutMinimumGasConsumption()`.
- **Rationale (PR):** simulated calls (gas estimation) must not have the
  minimum gas consumption of [core-7](#core-7) applied, or estimates are
  skewed.
- **Steps:** variadic option on the two signatures; payload field on the
  struct; one option argument at the gas estimator's `core.ApplyMessage` call.
- **At a merge:** check for new upstream callers that simulate execution and
  should pass the same option.

<a id="core-5"></a>

### core-5: `TransitionDb` rename-and-wrap (PRs #208, #321)

- **Where:** `core/state_transition.go`.
- **Change:** `func (st *StateTransition) TransitionDb()` → `transitionDb()`.
  The exported method is in `core/state_transition.libevm.go`.
- **Rationale:** the wrapper snapshots state, runs the original, copies
  `gasExecuted` into the result ([core-8](#core-8)), and if a precompile called
  `vm.EVM.InvalidateExecution` it reverts to the snapshot and returns the
  error ([vm-6](#vm-6)).
- **Steps:** rename-and-wrap. The upstream doc comment stays on the lower-case
  function; the wrapper carries a copy plus the libevm-specific paragraph.

<a id="core-6"></a>

### core-6: optionally credit the base fee to the coinbase (PR #299)

- **Where:** `core/state_transition.go`, `transitionDb`, in the branch that
  pays the tip, immediately after
  `st.state.AddBalance(st.evm.Context.Coinbase, fee)`.
- **Change:** `st.maybeCreditBaseFeeToCoinbase() // libevm`.
- **Rationale:** `params.RulesHooks.ShouldCreditBaseFeeToCoinbase`; some chains
  do not burn the base fee. Default: not credited (upstream behaviour).
- **Steps:** hook call, one line, directly after the tip credit.

<a id="core-7"></a>

### core-7: post-refund gas adjustment (PRs #185, #288, #321)

- **Where:** `core/state_transition.go`, `refundGas`, immediately after
  `st.gasRemaining += refund` and before the remaining gas is converted back
  to balance.
- **Change:** `st.afterGasRefund(refund)`, with a comment pointing at the
  call-site requirements.
- **Rationale:** one hook point applies, in order: disabling refunds
  (`ShouldRefundGas`), recording `gasExecuted`, and enforcing
  `MinimumGasConsumption` unless disabled by [core-4](#core-4).
- **Placement rule:** MUST be after all code that modifies gas consumption and
  before the balance for remaining gas is returned.
- **Steps:** hook call, one line plus blank lines around it.

<a id="core-8"></a>

### core-8: `GasExecuted` (PR #321)

- **Where:** `core/state_transition.go` (`ExecutionResult`, `StateTransition`),
  `core/state_processor.go` (`applyTransaction`), `core/types/receipt.go`,
  `core/types/gen_receipt_json.go`, `core/chain_makers_test.go`.
- **Change:**
  - `ExecutionResult` gains `GasExecuted uint64`; `StateTransition` gains
    `gasExecuted uint64`.
  - `applyTransaction`: `receipt.GasExecuted = result.GasExecuted //libevm`
    directly after `receipt.GasUsed = result.UsedGas`.
  - `types.Receipt` gains ``GasExecuted uint64 `json:"-"` `` under a
    `// libevm modifications` comment; `gen_receipt_json.go` is regenerated.
  - The upstream test `TestGeneratePOSChain` zeroes the field on expected
    receipts: `r.GasExecuted = 0 //libevm: not persisted so absent from blockchain receipts`.
- **Rationale (PR):** minimum gas consumption skews throughput metrics, so the
  gas actually executed (after refunds, before the minimum) is recorded. The
  field is deliberately **not serialised**: adding it to any RLP encoding would
  cause significant upstream conflicts. That is why an upstream test had to
  change, libevm's first.
- **Steps:** payload fields; one assignment in `applyTransaction`; regenerate
  the receipt JSON codec; one line in the test's receipt-normalising loop.

<a id="core-9"></a>

### core-9: `core/gen_genesis.go` (PR #86)

- **Change:** formatting only (struct-tag column alignment).
- **Rationale:** the file was stale relative to its generator; libevm's CI
  checks that generated code is up to date.
- **Steps:** regenerate. Never hand-edit.
- **Gone at geth v1.14.0:** upstream's file now matches the generator's output.

---

## `core/tracing`

<a id="tracing-1"></a>

### tracing-1: `tracing.StateDB.GetState` accepts options (added at geth v1.14.0)

- **Where:** `core/tracing/hooks.go`, the `StateDB` interface (new upstream
  file in v1.14.0).
- **Change:** `GetState(common.Address, common.Hash)` gains
  `...stateconf.StateDBStateOption`; import `libevm/stateconf` under a
  `// libevm extra imports` comment.
- **Rationale:** upstream passes its state database to tracers as this
  interface (`EVM.GetVMContext`). `vm.StateDB` and `*state.StateDB` would
  satisfy it but for libevm's variadic `GetState` ([state-4](#state-4),
  [vm-9](#vm-9)). This is the inverse of adding an ignored variadic option to a
  concrete type such as `snapshot.Tree` ([snapshot-1](#snapshot-1)): here the
  upstream interface is widened so the libevm-widened types still fit.
  Confirmed by the maintainer.
- **Alternative rejected:** an adapter wherever upstream passes the database
  to a tracer. That grows with upstream; this is one fixed edit.
- **Steps:** variadic option on the one interface method.
- **At a merge:** any upstream interface that `vm.StateDB` or
  `*state.StateDB` must satisfy and that declares one of the three widened
  methods needs the same treatment.

libevm's own gas and balance adjustments are reported to these hooks with the
reasons listed in [11](11-sync-log.md#tracer-reasons); two libevm-specific
reason constants are in `core/tracing/hooks.libevm.go`.

---

## `core/rawdb`

<a id="rawdb-1"></a>

### rawdb-1: `InspectDatabase` options (PRs #111, #220)

- **Where:** `core/rawdb/database.go` (`InspectDatabase`),
  `core/rawdb/ancient_utils.go` (`inspectFreezers`).
- **Change:**
  - Both gain `opts ...InspectDatabaseOption`; import `libevm/options`.
  - `InspectDatabase`: first statement
    `libevmConfig := options.As[inspectDatabaseConfig](opts...)`.
  - In the key-classification `switch`, two cases inserted immediately before
    `default`: `case libevmConfig.recordStat(key, size):` and
    `case libevmConfig.isMetadata(key): metadata.Add(size)`.
  - `inspectFreezers(db)` → `inspectFreezers(db, opts...)`.
  - `table.AppendBulk(stats)` → `table.AppendBulk(libevmConfig.transformStats(stats))`.
  - `inspectFreezers`: early `return nil, nil` when the config says
    `skipFreezers`.
- **Rationale:** consumers store extra key ranges and have no freezers; they
  need to classify their own keys, reshape the report and skip freezer
  inspection.
- **Steps:** variadic option on both; the config line; two `case`s before
  `default`; wrap the stats argument; guard at the top of `inspectFreezers`.
- **At a merge:** new upstream `case`s go above the libevm ones.

---

## `core/state`

<a id="state-1"></a>

### state-1: intercept `Database` construction (PR #283)

- **Where:** `core/state/database.go`.
- **Change:** `NewDatabaseWithConfig` → `newDatabaseWithConfig`;
  `NewDatabaseWithNodeDB` → `newDatabaseWithNodeDB`. Exported versions in
  `core/state/database.libevm.go` pass the result through a registered
  `DatabaseInterceptor`.
- **Rationale (PR):** an alternative state backend (Firewood) needs a different
  `state.Database` wherever libevm constructs one, registered once.
- **Steps:** rename-and-wrap, twice.
- **At a merge:** any new upstream constructor of `state.Database` needs the
  same treatment.

<a id="state-2"></a>

### state-2: account payload in state objects (PRs #44, #62)

- **Where:** `core/state/state_object.go` (`empty`), `core/state/statedb.go`
  (`getDeletedStateObject`).
- **Change:**
  - `empty()` appends `&& s.data.Extra.IsZero()`.
  - The `types.StateAccount` literal built from a snapshot account gains
    `Extra: acc.Extra,`.
- **Rationale:** `StateAccount` carries a libevm payload
  ([types-2](#types-2)). An account with a non-zero payload is not empty, and
  the payload must survive a read from the snapshot.
- **Steps:** one conjunct; one struct-literal entry.
- **At a merge:** every upstream site that builds or copies a `StateAccount`
  must carry `Extra`.

<a id="state-3"></a>

### state-3: `SnapshotTree` interface (PR #77)

- **Where:** `core/state/statedb.go`.
- **Change:** field `snaps *snapshot.Tree` → `snaps SnapshotTree`; `New(...,
  snaps *snapshot.Tree)` → `snaps SnapshotTree`; first statement of `New`:
  `snaps = clearTypedNilPointer(snaps)`.
- **Rationale:** allow a drop-in replacement for the snapshot tree. The nil
  normalisation keeps upstream's `snaps != nil` checks correct when a typed nil
  pointer arrives through the interface.
- **Steps:** interface widening on the field and parameter; the normalising
  line. `*snapshot.Tree` still satisfies the interface, so upstream callers
  compile unchanged.

<a id="state-4"></a>

### state-4: state-key transformation options (PR #205)

- **Where:** `core/state/statedb.go`: `GetState`, `GetCommittedState`,
  `SetState`. Interface in [vm-9](#vm-9).
- **Change:** each gains `opts ...stateconf.StateDBStateOption`, and inside the
  `stateObject != nil` branch, before the call on the state object, the key is
  rewritten: `hash = transformStateKey(addr, hash, opts...)` (`key = …` in
  `SetState`). Import `libevm/stateconf` under a `// libevm extra imports`
  comment.
- **Rationale:** a registered hook may transform storage keys; an option lets a
  caller skip the transformation.
- **Steps:** variadic option ×3; one reassignment line ×3.
- **At a merge:** any other implementer of the widened interface, upstream test
  doubles included, needs the parameter added (known:
  `eth/tracers/logger/logger_test.go`).

<a id="state-5"></a>

### state-5: commit options (PR #190)

- **Where:** `core/state/statedb.go`, `Commit`.
- **Change:** gains `opts ...stateconf.StateDBCommitOption`; the snapshot
  update call appends `stateconf.ExtractSnapshotUpdateOpts(opts...)...`; the
  `TrieDB().Update` call appends `stateconf.ExtractTrieDBUpdateOpts(opts...)...`.
- **Rationale:** pass caller-supplied payloads through `Commit` to custom
  snapshot and trie-database backends ([snapshot-1](#snapshot-1),
  [triedb-2](#triedb-2)).
- **Steps:** variadic option; append the extracted options to the two calls.

<a id="state-6"></a>

### state-6: prefetcher worker pools (PRs #76, #307)

- **Where:** `core/state/statedb.go` (`StartPrefetcher`),
  `core/state/trie_prefetcher.go`.
- **Change:**
  - `StartPrefetcher(namespace)` and `newTriePrefetcher(...)` gain
    `opts ...PrefetcherOption`, forwarded.
  - `triePrefetcher` gains `workers WorkerPool`, set to `newWorkerPool(opts...)`
    in the constructor literal.
  - `close()`: `p.releaseWorkerPool()` before the fetchers are cleared.
  - `prefetch()`: `newSubfetcher(...)` gains a final
    `withWorkerPool(p.workers)` argument.
  - `subfetcher` gains `pool *subfetcherPool`; `newSubfetcher` gains
    `opts ...subfetcherPoolOption` and calls `sf.initPool(opts...)` before
    `go sf.loop()`.
  - `loop()`: `defer close(sf.term)` becomes a deferred closure that calls
    `sf.pool.wait()` first.
  - `loop()`: `sf.trie.GetAccount(...)` / `sf.trie.GetStorage(...)` →
    `sf.pool.GetAccount(...)` / `sf.pool.GetStorage(...)`.
- **Rationale (PR):** prefetch concurrently for parity with the consumers'
  earlier performance. The pool does the reads; the loop waits for them before
  signalling termination.
- **Steps:** variadic option threaded through three signatures; two payload
  fields; receiver swap on the two trie reads; wrap the deferred close.

<a id="state-7"></a>

### state-7: subfetcher termination, copied from upstream (PR #76)

- **Where:** `core/state/trie_prefetcher.go`, `subfetcher.loop`.
- **Change:** inside the task loop the `case <-sf.stop:` arm is removed
  (`for i, task` → `for _, task`); the outer `case <-sf.stop:` no longer
  returns at once but exhausts pending tasks first. Both are marked
  `//libevm:start` / `//libevm:end`.
- **Rationale:** copied, with alteration, from upstream geth #29519; needed so
  work handed to the pool is not abandoned.
- **Steps:** none to preserve. **Delete when the merged upstream version
  includes geth #29519**, then re-apply only [state-6](#state-6).
- **Not in tracker #128**; it is a partial pre-emption that cannot be reverted
  as a commit.

---

<a id="corestatesnapshot"></a>

## `core/state/snapshot`

`disklayer.go`, `journal.go`, `generate_test.go` and most of `generate.go` and
`snapshot.go` differ from the base only through #286, a squashed backport of
geth #33540 (generator shutdown race). That is a tracker item, not a libevm
change. The entries below are libevm's own and **sit on top of #286's code**,
so they are re-seated when the upstream twin arrives.

<a id="snapshot-1"></a>

### snapshot-1: `Tree.Update` accepts options (PR #77)

- **Where:** `core/state/snapshot/snapshot.go`.
- **Change:** `Update(...)` gains a final
  `_ ...stateconf.SnapshotUpdateOption`; doc comment gains
  "libevm: Options are ignored and only included to match an interface method."
- **Rationale:** `*snapshot.Tree` must satisfy `state.SnapshotTree`
  ([state-3](#state-3)), whose `Update` carries options for custom trees.
- **Steps:** variadic option, ignored.

<a id="snapshot-2"></a>

### snapshot-2: generator can stop mid-scan; read errors end the run (PR #310)

- **Where:** `core/state/snapshot/context.go`, `generate.go`.
- **Change:**
  - `generatorContext` embeds `generatorPausing //libevm`.
  - `newGeneratorContext` gains `opts ...generatorContextOption`, applied with
    `options.ApplyTo(ctx, opts...)` before the iterators are opened.
  - `openIterator`: each `ctx.db.NewIterator(...)` result is wrapped,
    `iter = newAbortableIterator(iter, ctx.cancel) //libevm`.
  - `reopenIterator`: after `hasNext := iter.Next()`, return early if
    `!hasNext && iter.Error() != nil`.
  - `proveRange`: after the iteration loop, return `iter.Error()` if set.
  - `checkAndFlush`: first statement `ctx.progress = common.CopyBytes(current)`.
  - `generateAccounts`: after each of `removeStorageBefore`, `removeStorageAt`
    (both sites) and `removeStorageLeft`, propagate `ctx.storage.Error()`.
  - `generate`: `newGeneratorContext(...)` gains
    `withCancelFromDiskLayer(dl)`; on error, `errLibEVMIteratorAborted` is
    converted via `dl.flushAfterIteratorAbort(ctx)`; a non-abort error is
    logged.
- **Rationale (PR):** on a hash-scheme database one iterator `Next` can step
  over every trie node sharing a snapshot prefix, but the generator only checks
  for a stop between batches, so a stop request can stall block processing for
  minutes. Iterators now stop when asked; every reader treats that as an abort,
  not as end of data; finished work is saved. A database read error also ends
  the run instead of reading as end of data.
- **Steps:** variadic option on the context constructor; wrap both iterators;
  add an iterator-error check after every loop or helper that advances an
  iterator; record progress at the top of `checkAndFlush`; convert the abort
  error in `generate`.
- **At a merge:** any new upstream code that advances these iterators needs the
  same error check.

---

## `core/types`

<a id="types-1"></a>

### types-1: header, body and block payloads (PRs #89, #94, #106, #130, #133)

- **Where:** `core/types/block.go`.
- **Change:**
  - `Header`, `Body`, `Block` and the unexported `extblock` each gain a last
    field `extra *pseudo.Type`. Import `internal/libevm/pseudo`.
  - `CopyHeader`: `h.hooks().PostCopy(&cpy)` immediately before `return &cpy`.
  - `Block.DecodeRLP`: `eb.extra = b.extraOrNil()` directly after
    `var eb extblock`.
  - `Block.EncodeRLP`: the `extblock` literal gains `extra: b.extraOrNil(),`.
  - `Block.Body()`: the positional `Body{...}` literal gains a final
    `b.cloneExtra()`.
  - `WithSeal`, `WithWithdrawals`: the `Block` literal gains
    `extra: b.cloneExtra(),`.
  - `WithBody`: the `Block` literal gains `extra: body.cloneExtra(),`.
- **Rationale:** consumers attach extra fields to headers, bodies and blocks,
  with their own RLP and JSON encodings. The payload must follow the value
  through every copy and codec path.
- **Steps:** payload field ×4; one hook call in `CopyHeader`; one extra entry
  in every struct literal of those types.
- **Dependency:** the `WithBody` entry needs the `Body`-taking signature, which
  libevm pre-empted in #110 and upstream introduces as geth #29482. Before that
  version is merged there is no `body` value in scope.
- **At a merge:** every new upstream literal or copy of `Header`, `Body`,
  `Block` or `extblock` must carry `extra`. This is the commonest way for the
  transform to be lost silently.

<a id="types-2"></a>

### types-2: `StateAccount` / `SlimAccount` payload (PRs #44, #79)

- **Where:** `core/types/state_account.go`, `core/types/gen_account_rlp.go`.
- **Change:**
  - `StateAccount` and `SlimAccount` gain an exported last field
    `Extra *StateAccountExtra`.
  - `Copy()`: `Extra: acct.Extra.clone(),`.
  - `SlimAccountRLP`: literal gains `Extra: account.Extra,`;
    `rlp.EncodeToBytes(slim)` → `rlp.EncodeToBytes(&slim)`.
  - `FullAccount`: `account.Extra = slim.Extra` after nonce and balance are
    copied.
  - New directive
    `//go:generate go run ../../rlp/rlpgen -type SlimAccount -out gen_slim_account_rlp.libevm.go`.
  - `gen_account_rlp.go` regenerated: encodes `obj.Extra` before `ListEnd`.
- **Rationale:** per-account payload for consumers (pseudo-generic, accessed
  only through `ExtraPayloads`). With no payload registered the encoding is
  unchanged. `&slim` is needed because the generated `SlimAccount` encoder has
  a pointer receiver.
- **Steps:** payload field ×2; carry it in the three conversion functions; add
  the generate directive; regenerate.

<a id="types-3"></a>

### types-3: overridable `Header` codecs (PRs #86, #89, #90, #94)

- **Where:** `core/types/block.go` (generate directives),
  `core/types/gen_header_json.go`, `core/types/gen_header_rlp.go`.
- **Change:**
  - After each existing `//go:generate` line for `Header`, a second one runs
    `libevm/cmd/internalise` on the output: JSON → `Header.MarshalJSON`,
    `Header.UnmarshalJSON`; RLP → `Header.EncodeRLP`.
  - Generated files therefore define `marshalJSON`, `unmarshalJSON`,
    `encodeRLP`. The exported methods are in `core/types/block.libevm.go` and
    dispatch to a hook or fall back to the generated one.
- **Rationale (PR):** to override `Header` encoding there must be only one
  exported method; renaming the generated one avoids hand-editing generated
  code.
- **Steps:** add the two directives; regenerate. Never hand-edit the `gen_*`
  files.
- **At a merge:** if upstream regenerates these files, take upstream's
  `block.go`, re-add the directives, regenerate.

Receipt changes are under [core-8](#core-8).

---

## `core/vm`

<a id="vm-1"></a>

### vm-1: precompile override (PRs #7, #151)

- **Where:** `core/vm/evm.go`, top of `(*EVM).precompile`.
- **Change:** before the upstream lookup, if
  `evm.chainRules.Hooks().PrecompileOverride(addr)` reports an override, log at
  debug level and `return p, p != nil`. Import `log`.
- **Rationale:** chains add, replace or disable precompiles per address. A nil
  override disables the address. No override falls through to upstream.
- **Steps:** early-return guard as the first statements of the function.

<a id="vm-2"></a>

### vm-2: active-precompiles override (PR #39)

- **Where:** `core/vm/contracts.go`.
- **Change:** `ActivePrecompiles` → `activePrecompiles`; exported version in
  `core/vm/contracts.libevm.go` passes a copy of the original result through
  `rules.Hooks().ActivePrecompiles`.
- **Rationale:** keep the reported set consistent with [vm-1](#vm-1) (it feeds
  access-list warming).
- **Steps:** rename-and-wrap.

<a id="vm-3"></a>

### vm-3: stateful precompiles via `evmCallArgs` (PRs #7, #20, #26, #40, #50, #73)

- **Where:** `core/vm/contracts.go` (`RunPrecompiledContract`),
  `core/vm/evm.go` (the four call types).
- **Change:**
  - `func RunPrecompiledContract(p, input, suppliedGas)` becomes the method
    `func (args *evmCallArgs) RunPrecompiledContract(...)` with the same
    parameters and results.
  - In its body, after `suppliedGas -= gasCost`:
    `args.gasRemaining = suppliedGas`; `p.Run(input)` → `args.run(p, input)`;
    the return uses `args.gasRemaining` in place of `suppliedGas`.
  - At each of the four call sites, one line is inserted before the call and
    the call gains the receiver:

    | In | Inserted line |
    |---|---|
    | `call` (was `Call`) | `args := &evmCallArgs{evm, Call, caller, addr, input, gas, value}` |
    | `CallCode` | `args := &evmCallArgs{evm, CallCode, caller, addr, input, gas, value}` |
    | `DelegateCall` | `args := &evmCallArgs{evm, DelegateCall, caller, addr, input, gas, nil}` |
    | `StaticCall` | `args := &evmCallArgs{evm, StaticCall, caller, addr, input, gas, nil}` |

- **Rationale:** a stateful precompile needs the full call context (EVM, call
  type, caller, self, value) and must be able to use and refund gas. Capturing
  the enclosing function's arguments positionally, and shadowing the function
  with a same-named method, keeps each call site a one-receiver change. A
  plain `PrecompiledContract` runs exactly as upstream.
- **Steps:** method-shadow the function; three edits in its body; at every call
  site build `args` from **that function's own parameters, in order**, with the
  call-type constant matching the enclosing function and `nil` for value where
  the function has no `value` parameter.
- **At a merge:** every upstream call of `RunPrecompiledContract` must go
  through an `evmCallArgs`; a new call type needs a new constant. If upstream
  changes the function's parameters, mirror them on the method.

<a id="vm-4"></a>

### vm-4: override `NewEVM` and `Reset` arguments (PRs #35, #36, #38)

- **Where:** `core/vm/evm.go`.
- **Change:**
  - `NewEVM`: immediately before the `evm := &EVM{...}` literal,
    `blockCtx, txCtx, statedb, chainConfig, config = overrideNewEVMArgs(blockCtx, txCtx, statedb, chainConfig, config)`.
  - `Reset`: the two assignments become
    `evm.TxContext, evm.StateDB = evm.overrideEVMResetArgs(txCtx, statedb)`.
- **Rationale:** a registered `vm.Hooks` may replace constructor arguments
  (e.g. wrap the `StateDB`). Pass-through when nothing is registered.
- **Steps:** argument pass-through: reassign all parameters from one call
  taking and returning them in declaration order.
- **At a merge:** if upstream changes either signature, the override's
  parameter list changes with it (in `core/vm/hooks.libevm.go`); `vm.Hooks`
  then changes for consumers, which is a decision to raise.

<a id="vm-5"></a>

### vm-5: `BlockContext.Header` (PR #27)

- **Where:** `core/vm/evm.go`, `BlockContext`.
- **Change:** last field
  `Header *types.Header // libevm addition; not guaranteed to be set`.
- **Rationale and population:** see [core-1](#core-1).
- **Steps:** payload field.

<a id="vm-6"></a>

### vm-6: execution invalidation (PR #208)

- **Where:** `core/vm/evm.go`: `EVM` struct, `Reset`.
- **Change:** `EVM` gains `executionInvalidated error` under a `// libevm`
  comment; `Reset` clears it as its first statement.
- **Rationale:** a precompile can mark the whole transaction's execution
  invalid; `core.TransitionDb` then reverts and errors
  ([core-5](#core-5)). It must not leak into the next transaction on a reused
  EVM.
- **Steps:** payload field; one reset line.

<a id="vm-7"></a>

### vm-7: preprocessing gas charge (PR #235)

- **Where:** `core/vm/evm.go`.
- **Change:** `(*EVM).Call` → `call`; `(*EVM).create` → `createCommon` (and its
  doc comment's first word). The original names are defined in
  `core/vm/evm.libevm.go`; each first spends
  `vm.Hooks.PreprocessingGasCharge` and then calls the renamed function.
- **Rationale (PR):** charge gas between intrinsic gas and execution (needed
  for Warp) without touching `core.IntrinsicGas`, at the two entry points to
  execution.
- **Steps:** rename-and-wrap, twice. Callers are untouched because the names
  they use are preserved.

<a id="vm-8"></a>

### vm-8: contract-creation allowlist (PRs #7, #28)

- **Where:** `core/vm/evm.go`, `createCommon`, after the address-collision
  check and before `evm.StateDB.Snapshot()`.
- **Change:** a `//libevm:start` … `//libevm:end` block calling
  `gas, err := evm.canCreateContract(caller, address, gas)` and returning
  `nil, common.Address{}, gas, err` on error.
- **Rationale:** `params.RulesHooks.CanCreateContract` gates deployment and may
  consume gas.
- **Placement rule (in the code):** MUST be after the caller's nonce is
  incremented but before all other state-modifying behaviour, even if those
  changes may be reverted to the snapshot.
- **Steps:** arbitrary placement with rationale; keep the markers and comment.

<a id="vm-9"></a>

### vm-9: `StateDB` interface (PRs #205, #235)

- **Where:** `core/vm/interface.go`.
- **Change:** `GetCommittedState`, `GetState`, `SetState` gain
  `...stateconf.StateDBStateOption`; the interface embeds `StateDBRemainder`
  as its last entry; import under `// libevm extra imports`.
- **Rationale:** the options are those of [state-4](#state-4).
  `StateDBRemainder` (in `interface.libevm.go`) exposes the `*state.StateDB`
  methods upstream left off the interface, which libevm code needs through it.
- **Steps:** variadic option ×3; one embedded interface.
- **At a merge:** upstream reshapes this interface often. Re-apply the option
  to whichever of the three methods remain; drop from `StateDBRemainder`
  anything upstream has added to the interface itself.

---

## `eth/filters`

<a id="filters-1"></a>

### filters-1: stop the timeout goroutine (PR #254)

- **Where:** `eth/filters/api.go`.
- **Change:** `FilterAPI` gains `quit chan struct{}`, made in `NewFilterAPI`;
  `timeoutLoop`'s bare `<-ticker.C` becomes a `select` that also returns on
  `<-api.quit`.
- **Rationale (PR):** the goroutine otherwise keeps the `filters.Backend` alive
  forever, failing GC tests downstream. Kept deliberately minimal.
- **Steps:** payload field; initialise it; replace the receive with a
  two-case `select`. The closer is in a libevm file.

<a id="filters-2"></a>

### filters-2: custom bloom source (PR #261)

- **Where:** `eth/filters/filter.go` (`blockLogs`, `pendingLogs`),
  `eth/filters/filter_system.go` (`lightFilterLogs`).
- **Change:** the first argument of each `bloomFilter(...)` call becomes
  `maybeOverrideBloom(<header>, <backend>)`: `header.Bloom` →
  `maybeOverrideBloom(header, f.sys.backend)`; `block.Bloom()` →
  `maybeOverrideBloom(block.Header(), f.sys.backend)`; in the event system the
  backend is `es.backend`.
- **Rationale (PR):** with delayed settlement the header's bloom is not the
  right one; a backend implementing an optional interface supplies it.
- **Steps:** wrap the bloom argument at every `bloomFilter` call site.
- **At a merge:** check for new `bloomFilter` calls or other reads of
  `header.Bloom` in this package.
- **Since geth v1.14.0 there is one site, `blockLogs`.** Upstream removed the
  other two functions:
  - `pendingLogs` (geth #29574, pending-log support removed). The libevm
    subtest `TestBloomOverride/pendingLogs` was deleted with it. Maintainer's
    rationale: pending logs do not make sense in SAE, so nothing is lost.
  - `lightFilterLogs` (light-client mode removed). The libevm subtest
    `TestBloomOverride/lightFilterLogs` was deleted as obsolete, because
    upstream removed the feature.

---

## `eth/tracers`

<a id="tracers-1"></a>

### tracers-1: overridable block hash (PR #295)

- **Where:** `eth/tracers/api.go`, six sites: `blockByNumberAndHash`,
  `traceChain` (×2), `traceBlock`, `traceBlockParallel`,
  `standardTraceBlockToFile`.
- **Change:** `block.Hash()` → `api.blockHash(block)` (likewise for
  `task.block`, `res.block`).
- **Rationale (PR):** consumers alter the header to return the right state root
  and base fee, which changes the computed hash; an optional backend interface
  returns the original hash.
- **Steps:** replace **every** `Hash()` call on a `*types.Block` in this file.
- **At a merge:** grep the merged file for remaining `.Hash()` calls on blocks.

<a id="tracers-2"></a>

### tracers-2: prestate tracer's scope-entry hook (PR #64; re-seated at geth v1.14.0)

- **Where:** `eth/tracers/native/prestate.go`, the `tracing.Hooks` literal in
  the tracer's constructor.
- **Change:** one entry, `OnEnter: t.OnEnter, // libevm: see prestate_libevm.go`.
- **Rationale (PR):** the prestate tracer assumes an account has been looked up
  before its storage is read, and maintains that only when it sees a
  scope-entering opcode. A stateful precompile calling another contract enters
  a scope with no such opcode, which panicked on a nil map. libevm's `OnEnter`
  (in `prestate_libevm.go`) restores the invariant.
- **History:** before v1.14.0 upstream's tracers were an interface, so libevm
  only had to define the method (`CaptureEnter`); no upstream file changed.
  v1.14.0 replaced the interface with a struct of hook functions, so the method
  must now be registered. Without the line the fix compiles and is silently
  dead; `TestPrecompileCallWithPrestateTracer` is what catches it.
- **Steps:** one entry in the struct literal. A compile-time assertion in the
  libevm file checks the method still matches `tracing.EnterHook`.

Test double: `eth/tracers/logger/logger_test.go` implements the widened
interface of [state-4](#state-4); see the upstream test baseline in
[07](07-baselines-and-facts.md#upstream-test-baseline).

---

## `ethclient`

<a id="ethclient-1"></a>

### ethclient-1: post-unmarshal hook (PR #316)

- **Where:** `ethclient/ethclient.go`, end of `getBlock`.
- **Change:** the single `return types.NewBlockWithHeader(head)…, nil` is split:
  assign to `block`, call `block.PostRPCUnmarshal(raw)` (returning its error),
  then `return block, nil`.
- **Rationale:** counterpart of [ethapi-3](#ethapi-3); lets body hooks restore
  payload fields from the raw RPC JSON.
- **Steps:** split the return; one hook call between.

---

## `internal/ethapi`

<a id="ethapi-1"></a>

### ethapi-1: `BlockChainAPI` options (PR #314)

- **Where:** `internal/ethapi/api.go`.
- **Change:** `BlockChainAPI` gains `libevm blockChainAPIConfig`;
  `NewBlockChainAPI(b)` gains `opts ...BlockChainAPIOption` and returns
  `&BlockChainAPI{b, *options.As[blockChainAPIConfig](opts...)}`. Import
  `libevm/options`.
- **Steps:** payload field; variadic option; extend the positional literal.

<a id="ethapi-2"></a>

### ethapi-2: call-result interceptor (PR #314)

- **Where:** `internal/ethapi/api.go`: `DoCall`, `(*BlockChainAPI).Call`.
- **Change:** `DoCall` gains `opts ...CallOption` and **named results**
  `(res *core.ExecutionResult, retErr error)`; its first statement is a
  deferred closure `res, retErr = interceptCallResult(res, retErr, opts...)`.
  `Call` passes `s.libevm.callOpts...`.
- **Rationale:** `ethapi.CallResultInterceptor` lets a consumer rewrite the
  outcome of `eth_call` without forking the function.
- **Steps:** variadic option; name the results; the deferred call as the first
  `defer`; forward options from the one upstream caller.

<a id="ethapi-3"></a>

### ethapi-3: post-marshal hooks (PR #285)

- **Where:** `internal/ethapi/api.go`: `RPCMarshalHeader`, `RPCMarshalBlock`.
- **Change:** `head.PostRPCMarshal(result) //libevm` and
  `block.PostRPCMarshal(fields) //libevm`, each immediately before the
  function's final `return`.
- **Rationale:** header and body hooks add or alter fields in the RPC
  representation.
- **Steps:** hook call before the final return, ×2.

---

## `metrics`

<a id="metrics-1"></a>

### metrics-1: enabled by default (PR #165)

- **Where:** `metrics/metrics.go`.
- **Change:** `var Enabled = false` → `true`; the doc comment's "If it is true,
  the metric returned is a stub" corrected to "false".
- **Rationale (PR):** `Enabled` gates metric *constructors*. Most metrics are
  package-level globals built during initialisation, so a consumer setting it
  later gets no-op metrics.
- **Steps:** flip the literal; fix the comment.
- **Note:** this changes a default relative to geth, so it is a deliberate
  exception to "geth semantics by default".

---

## `params`

<a id="params-1"></a>

### params-1: `ChainConfig` and `Rules` payloads and hooks (PRs #7, #29)

- **Where:** `params/config.go`.
- **Change:**
  - `ChainConfig` and `Rules` each gain a last field `extra *pseudo.Type`.
    Import `internal/libevm/pseudo`.
  - `Description()`: `return banner` → `return banner + c.Hooks().Description()`.
  - `CheckConfigForkOrder()`: final `return nil` →
    `return c.Hooks().CheckConfigForkOrder()`.
  - `checkCompatible()`: final `return nil` →
    `return c.Hooks().CheckConfigCompatible(newcfg, headNumber, headTimestamp)`.
  - `Rules()`: `return Rules{…}` → `r := Rules{…}`, then
    `c.addRulesExtra(&r, num, isMerge, timestamp)`, then `return r`.
- **Rationale:** the canonical extra. `ChainConfig` and `Rules` are plumbed
  everywhere, so they carry payloads and hooks to any site without further
  plumbing. With nothing registered, hooks are no-ops equal to geth.
- **Steps:** payload field ×2; delegate return ×3; in `Rules`, bind the literal
  to a variable, call the hook with the method's own parameters, return it.
- **At a merge:** if `Rules()` gains parameters, `addRulesExtra` and
  `Extras.NewRules` must mirror them.

---

## `rlp/rlpgen`

<a id="rlpgen-1"></a>

### rlpgen-1: alias and named-basic types (PRs #154, #236)

- **Where:** `rlp/rlpgen/gen.go`, `makeOp`; `rlp/rlpgen/gen_test.go`;
  `rlp/rlpgen/testdata/alias.{in,out}.txt`.
- **Change:**
  - `switch typ := typ.(type)` → `switch typ := types.Unalias(typ).(type)`.
  - In the `*types.Named` case, before the recursive `makeOp` call:
    `if hasBasicUnderlying(typ) { return bctx.makeNamedBasicOp(typ) }`.
  - The test-name list gains `"alias"` and `"named.libevm"`.
- **Rationale:** newer Go represents type aliases explicitly, which broke the
  generator on aliased types; libevm types also use named types over basic
  kinds.
- **Steps:** wrap the switch operand; one guard in the `Named` case; register
  the two test cases.
- **Back burner:** the alias part was sent upstream; delete locally once the
  merged version has it.

---

## `rpc`

<a id="rpc-1"></a>

### rpc-1: per-call timeout (PRs #313, #317)

- **Where:** `rpc/service.go`, `rpc/handler.go`, `rpc/http.go`.
- **Change:**
  - `serviceRegistry` gains `libevm registryExtras`.
  - `newHandler`: first statement
    `connCtx = context.WithValue(connCtx, registryContextKey{}, reg)`.
  - `ContextRequestTimeout`: after the context-deadline check,
    `if d, ok := callTimeout(ctx); ok { setTimeout(d) }`.
- **Rationale (PRs):** WebSocket calls descend from `context.Background()`, so
  upstream's per-call timeout never applies and a slow call runs on after the
  client gives up. `Server.SetCallTimeout` stores a timeout on the registry,
  reachable from each call's context. #317 put the registry on every handler's
  context so HTTP is covered too.
- **Steps:** payload field; one context line at the top of `newHandler`; one
  extra candidate in `ContextRequestTimeout`, placed with the others so the
  shortest wins.

---

## `trie`

<a id="trie-1"></a>

### trie-1: equal key lengths in range proofs (PR #176)

- **Where:** `trie/proof.go`, `VerifyRangeProof`, first statement of the
  monotonicity loop.
- **Change:** return `false, errKeysHaveDifferentLengths` if adjacent keys
  differ in length.
- **Rationale (PR):** the function assumes equal-length keys; a key that is a
  prefix of another can make it panic.
- **Steps:** one guard at the top of the loop body.
- **Note:** a behaviour change relative to geth for malformed input (error
  instead of possible panic). Check whether upstream has since fixed it; the
  cherry-pick of geth #33898 touches the same function.

---

## `triedb`

<a id="triedb-1"></a>

### triedb-1: arbitrary backend (PR #70)

- **Where:** `triedb/database.go`.
- **Change:**
  - `Config` gains `DBOverride DBConstructor`.
  - `NewDatabase`: after `db` is constructed and before the hash/path scheme
    checks, `if db.overrideBackend(diskdb, config) { return db }`.
  - `Reader`: a `case ReaderProvider: return b.Reader(blockRoot)` as the
    **first** case of the type switch.
  - Every type assertion to `*hashdb.Database` becomes `HashDB`, and to
    `*pathdb.Database` becomes `PathDB` (`Cap`, `Reference`, `Dereference`;
    `Recover`, `Recoverable`, `Disable`, `Enable`, `Journal`, `SetBufferSize`).
- **Rationale:** consumers supply their own trie database (e.g. Firewood). The
  interfaces are satisfied by the upstream concrete types, so default behaviour
  is unchanged.
- **Steps:** payload field on the config; early return after construction; one
  leading `case`; interface widening on each assertion.
- **At a merge:** any new method asserting a concrete backend type needs
  widening, and the interface in the libevm file needs the new method.

<a id="triedb-2"></a>

### triedb-2: update options (PR #190)

- **Where:** `triedb/database.go` (`backend` interface, `Database.Update`),
  `triedb/hashdb/database.go`, `triedb/pathdb/database.go`.
- **Change:** `Update` gains `opts ...stateconf.TrieDBUpdateOption` on the
  interface and on `Database` (which forwards them); `hashdb` and `pathdb`
  accept and ignore them (`_ ...`). Import `libevm/stateconf` in all three.
- **Rationale:** carry payloads from `StateDB.Commit`
  ([state-5](#state-5)) to a custom backend.
- **Steps:** variadic option on the interface method and its three
  implementations.

---

## Excluded, and why

| File(s) | Reason |
|---|---|
| `crypto/signature_nocgo.go`, `go.mod`, `go.sum` | Entirely cherry-pick #323 (geth #30181) |
| `core/state/snapshot/{disklayer,journal,generate_test}.go` and the rest of `generate.go`, `snapshot.go` | Backport #286 (geth #33540) |
| `crypto/secp256k1/libsecp256k1/**`, `crypto/secp256r1/**`, `core/vm/contracts.go` P256 parts | Cherry-picks #194, #214 |
| `.github/**`, `.golangci.yml`, `.yamllint.yml` | Group 1b |

## Things noticed while cataloguing

- **[state-7](#state-7) is a pre-emption that #128 does not list.**
- **[core-2](#core-2) changes what `txpool` passes for `isHomestead`.**
- **[metrics-1](#metrics-1) and [trie-1](#trie-1) change default behaviour**
  relative to geth, unlike every other entry.
- **Several entries depend on a tracker item's code**, so their order relative
  to that item's revert matters: [types-1](#types-1) on #110,
  [snapshot-2](#snapshot-2) on #286.
- **Silent-loss candidates** (a merge can drop them and still compile): every
  "carry `extra` in a new literal" case under [types-1](#types-1) and
  [state-2](#state-2); new `bloomFilter` or `block.Hash()` sites under
  [filters-2](#filters-2) and [tracers-1](#tracers-1); new iterator loops under
  [snapshot-2](#snapshot-2).
