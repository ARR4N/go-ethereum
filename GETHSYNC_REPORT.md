# gethsync report: v1.14.5 → v1.14.8

- **Date:** 2026-11-26
- **Status:** all checks green, with items 1–3 applied as I recommend. Not yet on `gethsync`; waiting for your answers to items 1–3.
- **Links are relative to the repository root at** `gethsync-blocked`, whose parent is `e2030f333`:
  - `b96980900` — the merge of geth v1.14.8 (upstream `880511dc3`). Still titled "WIP"; the final message is written once items 1–3 are answered.
  - `e2030f333` — one new test for item 1, in its own commit because it is my own work.
- **`gethsync` is at** `1c147d76f`: v1.14.5 plus this group's three reverts.

## How to read this

One list, in priority order.

| Items | Kind | What I need from you |
|---|---|---|
| 1–3 | **Decisions** | An answer. Each is applied as recommended so that you can review real code |
| 4–12 | **Alerts** | Nothing, unless you disagree. Each cites the rule or answer that settles it |
| 13–17 | **Informational** | Nothing |

Items 2, 3, 4, 11 and 12 are the differences from your old branch at v1.14.8 (`0f8a7d997`). Item 13 says how I found them.

## Decisions

1. **The prefetcher's worker pool broke upstream's stateless witness. Fix applied; please confirm both halves.**
   - **Symptom before the fix:** `TestBlockchain` failed about 45,000 subtests with "cross verification failed: stateless execution failed: invalid gas used", reported at [tests/block_test.go:89](tests/block_test.go#L89). Pristine upstream passes.
   - **Your old branch has the same failure.** Its v1.14.8 merge has the same always-copy code, and `go test -run TestBlockchain ./tests` fails there with the same message. I ran it after this group went green.
   - **How upstream's witness works (new in v1.14.8):**
     - The block importer creates a witness and hands it to the prefetcher: [core/blockchain.go:1813-1820](core/blockchain.go#L1813-L1820).
     - `StartPrefetcher` stores it and passes `noreads = (witness == nil)`: [core/state/statedb.go:211-223](core/state/statedb.go#L211-L223). So with a witness, reads are prefetched too.
     - At commit time, the state takes each trie *from the prefetcher* ([core/state/statedb.go:936](core/state/statedb.go#L936), [core/state/state_object.go:157](core/state/state_object.go#L157)), which is the sub-fetcher's own trie ([core/state/trie_prefetcher.go:187-207](core/state/trie_prefetcher.go#L187-L207), [core/state/trie_prefetcher.go:310-314](core/state/trie_prefetcher.go#L310-L314)).
     - The witness is the set of nodes those tries have accessed: [core/state/statedb.go:886-920](core/state/statedb.go#L886-L920) and [core/state/statedb.go:985-987](core/state/statedb.go#L985-L987), via [`Trie.Witness()`](trie/trie.go#L664-L674).
     - The block is then re-executed from the witness alone: [core/blockchain.go:1941-1942](core/blockchain.go#L1941-L1942), [core/stateless.go:42](core/stateless.go#L42).
   - **Why libevm broke it:** upstream's loop reads through `sf.trie` directly (`sf.trie.GetAccount` and `sf.trie.GetStorage` in pristine upstream). libevm replaces those two reads with [`sf.pool.GetAccount`](core/state/trie_prefetcher.go#L388) and [`sf.pool.GetStorage`](core/state/trie_prefetcher.go#L390), which before the fix always read through *copies* of the trie, pool or no pool. The sub-fetcher's own trie therefore recorded nothing beyond its root, and the witness was incomplete.
   - **Fix, part (a): with no worker pool, read through the sub-fetcher's own trie.** [core/state/trie_prefetcher.libevm.go:141-154](core/state/trie_prefetcher.libevm.go#L141-L154); the accessor is set at [core/state/trie_prefetcher.libevm.go:93](core/state/trie_prefetcher.libevm.go#L93). This is exactly what unmodified geth does, on the same goroutine, so it restores the default. I think this half is uncontroversial.
   - **Fix, part (b): a worker pool is ignored when a witness is being collected.** [core/state/trie_prefetcher.libevm.go:53-69](core/state/trie_prefetcher.libevm.go#L53-L69), documented on the exported option at [core/state/trie_prefetcher.libevm.go:42-47](core/state/trie_prefetcher.libevm.go#L42-L47).
     - The one call site changes from `newWorkerPool(opts...)` to [`newWorkerPool(!noreads, opts...)`](core/state/trie_prefetcher.go#L93). It relies on upstream passing `noreads = (witness == nil)`; that coupling is the weak point.
     - Reason: pool workers must use copies (a trie is not safe to share between goroutines), and copies cannot feed the witness. A silently incomplete witness fails block import.
     - Alternative: leave the pool on and document that `WithWorkerPool` and a witness are incompatible. Then anyone combining them gets the failure in the symptom above.
     - As far as I can tell this only matters to a consumer that passes both a witness and a pool. A caller that passes a `nil` witness is unaffected.
   - **New test (mine):** [core/state/trie_prefetcher.libevm_claude_test.go:29-92](core/state/trie_prefetcher.libevm_claude_test.go#L29-L92) covers the four combinations of witness and pool.
   - **Existing libevm tests:** unchanged by this fix and passing.

2. **Difference from your branch: you kept a wake-channel nudge in the sub-fetcher's stop case; I did not.**
   - **What you had:** after upstream's code in the stop case, a `//libevm:start` block doing a non-blocking send on `sf.wake`. It is in your merges at v1.14.5 and v1.14.8.
   - **What this sync has:** upstream's code unmodified, [core/state/trie_prefetcher.go:399-412](core/state/trie_prefetcher.go#L399-L412).
   - **Why I dropped it:** on libevm `main` that block carried the comment "copied, with alteration, from ethereum/go-ethereum#29519 and can be deleted once we update to include that change". #29519 arrived with v1.14.4, so I took upstream's version whole.
   - **My two mistakes:** I did not flag the removal at v1.14.4, and I missed the difference in the v1.14.5 comparison with your branch.
   - **Is it needed?** I believe not. `schedule` appends tasks and then always sends on `wake` ([core/state/trie_prefetcher.go:282-296](core/state/trie_prefetcher.go#L282-L296)), so when the stop case finds tasks pending, a wake signal is already queued or about to be. Without the nudge the loop can spin on the closed `stop` channel for that short interval; the nudge removes the spin.
   - **Options:** leave it out (recommended: it matches upstream and libevm's own comment), or restore your block.

3. **Difference from your branch: which arguments the libevm prefetcher tests pass for upstream's new parameters.**
   - **Yours:** `newTriePrefetcher(..., false /*noreads*/, opt)` and `prefetch(..., true /*read*/)`.
   - **This sync:** `true /*noreads*/` and `false /*read*/`: [core/state/trie_prefetcher.libevm_test.go:70-71](core/state/trie_prefetcher.libevm_test.go#L70-L71), [core/state/trie_prefetcher.libevm_test.go:124](core/state/trie_prefetcher.libevm_test.go#L124) and [core/state/trie_prefetcher.libevm_test.go:131](core/state/trie_prefetcher.libevm_test.go#L131).
   - **Why mine:** `noreads = true` is what upstream passes when there is no witness ([core/state/statedb.go:223](core/state/statedb.go#L223)), which is the configuration the pool is for.
   - **It now depends on item 1(b):** under 1(b), `noreads = false` means "collecting a witness", so the pool is ignored and your arguments would fail both tests. If you reject 1(b), either choice works.

## Alerts

4. **Decided by you and applied: `triedb.ReaderProvider` and the `triedb.DBOverride` interface are deleted. This now matches your branch.**
   - Upstream added `Reader` to its `backend` interface ([triedb/database.go:62-92](triedb/database.go#L62-L92)), so both duplicated `BackendDB`.
   - The injection point stays: [`Config.DBOverride`](triedb/database.go#L41), of type [`DBConstructor`](triedb/database.libevm.go#L32-L34), which returns `BackendDB`. The override must still satisfy `HashDB` or `PathDB` ([triedb/database.libevm.go:41-57](triedb/database.libevm.go#L41-L57)), as before the sync.
   - Consumers change their constructor's return type from `triedb.DBOverride` to `triedb.BackendDB`. The two libevm tests that build one changed in that way only: [triedb/database.libevm_test.go:31](triedb/database.libevm_test.go#L31), [core/state/statedb.libevm_test.go:57](core/state/statedb.libevm_test.go#L57).
   - Remaining differences from your branch: one doc-comment line of mine ([triedb/database.libevm.go:33](triedb/database.libevm.go#L33)), and the tests keep their existing parameter names where you dropped them.

5. **`triedb.PathDB.Recover` lost its `loader` parameter** ([triedb/database.libevm.go:78](triedb/database.libevm.go#L78)), mirroring upstream's `pathdb.Database.Recover` as called at [triedb/database.go:272](triedb/database.go#L272). Exported libevm interface; its own comment says to update it when upstream's type changes. Same as your branch.

6. **Prefetcher options re-seated after new upstream parameters.**
   - [`StartPrefetcher(namespace, witness, opts...)`](core/state/statedb.go#L211) and [`newTriePrefetcher(db, root, namespace, noreads, opts...)`](core/state/trie_prefetcher.go#L68).
   - Existing callers of `StartPrefetcher` must add a witness argument, as upstream's own does at [core/blockchain.go:1820](core/blockchain.go#L1820). That break is upstream's. Same as your branch.

7. **`core/rawdb` example: expected output updated.** Upstream prints two new Verkle rows ([core/rawdb/examples.libevm_test.go:141-142](core/rawdb/examples.libevm_test.go#L141-L142)) from its new case at [core/rawdb/database.go:561](core/rawdb/database.go#L561). You said at v1.14.5 that you do not mind what this example prints.

8. **`hashdb.New` lost its third argument** in one libevm test: [core/state/statedb.libevm_test.go:53](core/state/statedb.libevm_test.go#L53). Forced by upstream's signature.

9. **#323's cherry-pick was reverted only in part.** [crypto/signature_nocgo.go](crypto/signature_nocgo.go) was reverted and now equals upstream. `go.mod` and `go.sum` were left, since libevm already requires a later btcec ([go.mod:14](go.mod#L14)); `go.mod` is owned and merged.

10. **#155 was reverted by restoring two files to upstream's v1.14.5 contents**, as the `memsize` removal was libevm's only change to either: [cmd/geth/main.go](cmd/geth/main.go) and [internal/debug/flags.go](internal/debug/flags.go). After the merge both equal upstream.

11. **Difference from your branch: blank lines around libevm's cases in `InspectDatabase`.** You have a blank line before and after them; this sync has none ([core/rawdb/database.go:577-579](core/rawdb/database.go#L577-L579)), which is how they are on libevm `main`. Whitespace only; settled by the objective of keeping the difference from upstream minimal.

12. **Difference from your branch: `holiman/uint256` in the tooling module.** Yours lists `v1.3.1` as an indirect dependency; this sync has `v1.2.4` ([libevm/tooling/go.mod:24](libevm/tooling/go.mod#L24)), unchanged from libevm `main`. `go mod tidy` leaves it as it is, which is the CI gate that governs this file.

## Informational

13. **How the comparison with your branch was done.** For every file that differs between `0f8a7d997` and this sync, I compared that difference with the same file's difference at v1.14.5 (`291f8663b` against `bf9dccade`). Where the two are identical, the difference predates this group and was covered at v1.14.5. Fifteen files remained and I read each one.
    - They account for items 2, 3, 4, 11 and 12; the fix and test in item 1; version strings; and prefetcher structure that libevm `main` changed after your branch point (one shared pool, `WithWorkerPool`).
    - Caveat: this method cannot find a difference that was already present, unnoticed, at v1.14.5. Item 2 was one. I have not re-audited v1.14.5 for others; say if you want that before I go on.

14. **Check results on `e2030f333`.**
    - Test suite, full and from scratch (`-count=1`, no `-short`): green. 195 packages, no failures, nothing re-run.
    - CI gates, in libevm naming: all seven pass.
    - Upstream's own tests: 12 files differ from upstream v1.14.8, the same 12 as at v1.14.5, all previously explained.

15. **Files that equal upstream again:** the two `memsize` files and `crypto/signature_nocgo.go`. #30001 reverted cleanly. No new upstream path bypasses a hook.

16. **A mistake caught before it was pushed.** I typed a wrong full hash into one revert's message, corrected it, verified every "This reverts commit" hash on the branch, and now generate them with `git rev-parse`.

17. **What comes next, once items 1–3 are answered.** Final merge message, move to `gethsync`, sync-log and catalogue entries. Nothing needs reverting again until v1.15.0 (#31100), so v1.14.9 to v1.14.12 can then be taken in bigger steps.
