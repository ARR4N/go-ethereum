# 9. Open items

## Undecided

| Item | Notes |
|---|---|
| Where the skill finally lives | This branch is its seed; maintainer may also reuse the original conversation |

## To verify before or during the sync

| Item | Why |
|---|---|
| Fetch upstream tags beyond v1.16.2 | Readable directly from `ethereum/go-ethereum`; latest is v1.17.10 as of 2026-11-21. Twins for #243, #270, #286 are beyond v1.16.2 |
| Re-verify the tag → `master` commit table in [05](05-process.md) | Derived from `merge-base`; confirm tree equality per row |
| Both oracles give the same result in geth and libevm naming | Chained merges are tested in geth naming. Expected to hold; `accounts/abi/bind` and `rlp/rlpgen` are the known path-sensitive packages |
| Every hook has a test | The safeguard against a silently lost transform |
| SSH-signed commits show as verified on GitHub | No SSH-signed commit has ever landed on libevm `main`; all 211 first-parent commits since the base are GPG-signed |
| Upstream's twin for libevm's rlpgen alias support (#154) | Only candidate seen is geth #31148 ("implement package renaming support"); unconfirmed |
| Upstream version at which `IntrinsicGas()` changes | For #280 |

## To build (tooling)

- Group 1 pathspec file (single definition for reviewer and tools); to be
  produced together with the catalogue of libevm changes.
- Transform inventory: every libevm edit to a group 2 file on `main`, by shape
  and hook, derived from the diff against upstream after reverting
  cherry-picks. Reconciled after each merge.
- Hook-coverage check: for each wrapped or shadowed symbol, list uses in merged
  upstream code that bypass libevm's version.
- Revert-inverse check: each revert equals the inverse of its cherry-pick,
  after module-path normalisation.
- Upstream-test-diff check against the baseline in
  [07](07-baselines-and-facts.md#upstream-test-baseline).
- Oracle runner: `go test -short` as CI does, with sequential re-run of
  `FLAKY_REGEX` packages and reporting of any new flake.
- Old-branch comparison at matching versions.

<a id="cleanup"></a>

## Cleanup at the end of the first sync

This directory must end up sufficient for a human or agent with no context
beyond the repository and its GitHub artefacts. Before it lands on `main`:

- Remove rules that are about the planning session, not the task: the fork and
  push restrictions, "do not start", the cloud-environment notes, tone.
- Remove or relocate one-off state once redundant: anchors, the old branch's
  structure, the 1.13/1.14 history, cherry-pick snapshots, lessons.
- Remove references a cold reader cannot follow: hashes on branches that may be
  deleted, the proposal PDF (keep the summarised principles), undefined roles.
- Replace copied snapshots (flaky list, tracker tables) with how to derive them.
- Add what the sync itself teaches: the runbook, commit message templates, the
  group 1 pathspec and the catalogue of libevm changes.

## Back burner (remind the maintainer when wrapping up)

- **rlpgen alias testdata:** `rlp/rlpgen/testdata/alias.{in,out}.txt` and the
  `"alias"` entry in `gen_test.go` were sent upstream and can be deleted
  locally once the merged upstream version contains them. Probably outside
  this job.
- Marker-comment convention (`//libevm:start` etc.) was never formalised.
- Consumer API diff with attribution (nice to have).

## Lessons and corrections so far

Recorded because each was a wrong or over-confident claim made during planning.

| Claim | Reality |
|---|---|
| Reverts must happen in libevm naming to apply cleanly | 5 of 6 applied cleanly in geth naming. The round trip is a fallback |
| "64 upstream files carry hook edits" | Count included cherry-pick and pre-emption edits and was taken at the old branch point. Not a usable number; build the inventory properly |
| "Keep ours" for `go.mod` | `go.mod` is owned and merged, not ignored |
| Folding adaptations into the merge commit means no broken commits anywhere | Revert and rename commits between merges may not build. Green is per merge commit |
| Release tags before v1.15.0 are not on `master` | Only v1.13.15 and v1.14.6–12 are off `master`; two real break points |
| Upstream tests can be assumed to pass | Some packages are flaky; rule defined in [04](04-objectives-and-oracles.md) |
| "At most one upstream test was ever changed" | Three files carry libevm edits, all scaffold-level |
| The tracker's hash for #110 is the commit to revert | It is the upstream twin (`2e8e35f2a`); the libevm commit is `67f879ac1` |
| PR → commit mapping for #128 | Was inferred by counting; later confirmed from PR titles |
| Merge-base results in a fresh clone | Wrong while the clone was shallow |
| A fork clone alongside the libevm clone | Unnecessary; add the fork as a remote instead |
