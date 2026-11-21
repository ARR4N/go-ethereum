# 4. Objectives and oracles

The result is a deliberate hybrid. Two oracles prove it is still both things:

- **Geth oracle:** upstream's tests pass. "This is still geth."
- **libevm oracle:** libevm's tests pass, semantically unchanged. "This is still
  libevm."

Oracles only mean something where the two worlds join: at merge commits.

## Primary objectives

| | Files | Minimise | Compared against |
|---|---|---|---|
| A | Group 2 (upstream) | The diff from upstream ("disruption") | The incoming upstream version, i.e. the last upstream commit merged |
| B | Group 2 (upstream) | Changes to the set of transforms libevm applies | The transforms on current libevm `main` |
| C | Group 1 minus build/CI | Changes to non-test code; **zero semantic change to anything, tests or otherwise** | Current libevm `main` |

- **B is about transforms, not text.** A literal diff-of-diffs is meaningless
  because surrounding code moves. Every hook site after the sync should be the
  same transform (shape, hook, arguments) re-seated. Ideally none added,
  removed or altered.
- **C:** libevm files, tests and non-test code alike, may change only
  syntactically, to conform to upstream changes. **Any semantic change to
  anything libevm means stop and defer to the maintainer; never move on.**
  This is achievable because geth keeps backwards compatibility through
  `params.ChainConfig` and `params.Rules`.

## Priority when they conflict

1. **A is primary.** geth's future commits cannot know about libevm's edits.
2. **B next.** A stable transform is one upstream is least likely to collide with.
3. **C absorbs the remainder.** Changing libevm files is acceptable because
   future libevm commits will know about it. The test constraint still holds.

## Additional objectives

| # | Objective | Decision |
|---|---|---|
| 1 | **Hooks cover code paths upstream has added.** Neither oracle sees a new upstream call site that bypasses a hook. For every symbol a transform wraps or shadows, list its uses in the merged code and confirm each goes through libevm's version | Agreed. Done as **follow-up commits after the merge commit**, within the same version group, so the merge stays reviewable |
| 2 | **No transform silently lost** when upstream deletes or moves the function holding it | Keep an inventory of transforms and reconcile it, as a working aid. The real safeguard is that **every hook has a test**; verify that this is so |
| 3 | **The sync is reviewable** | See [05](05-process.md#merge-commit-contents) and [review](05-process.md#review) |
| 4 | Consumer-visible API change is attributable to an upstream change or a libevm adaptation | Nice to have, not a gate. Note attributions where free |
| 5 | **Tracker closed out**: every #128 item and true cherry-pick reverted and replaced by its twin, or still listed with a reason | Agreed |

## Definition of green

The oracle command is `go test ./...`, with or without `-short` (CI uses
`-short`). Upstream's and libevm's tests live in the same packages, so one run
exercises both oracles.

- **A test that did not run proves nothing.** Tests driven by downloaded
  fixtures (the `tests/testdata` submodules and the execution spec tests in
  `tests/spec-tests`) are **skipped, not failed**, when the fixtures are
  absent. Before accepting any result, confirm the fixtures were present and
  the tests actually ran. `tools/prepare-oracle.sh` fetches them;
  `tools/run-oracle.sh` refuses to run without them and checks its logs for
  the skip message. The same care applies to any other test that can skip.
- libevm tests: all pass.
- Upstream tests: all pass, unmodified apart from the known baseline
  ([07](07-baselines-and-facts.md#upstream-test-baseline)).
- Flaky tests:
  - Packages already in CI's `FLAKY_REGEX` (`.github/workflows/go.yml`): re-run
    sequentially; they pass eventually, usually on one repeat. Then continue.
  - A flaky test **not** in that list: **flag it to the maintainer.**
- Go directive is always 1.26, so `go vet` (run by `go test`) is stricter than
  upstream intended at older tags. Minimal vet fixes to upstream files are an
  accepted cost; see [05](05-process.md#go-version).

## Green rule

- **Both oracles pass at every upstream version merged**, patch versions
  included. No compile-only checkpoints.
- Patch versions may be skipped or combined in logical groups (as the old
  branch did, e.g. 1.14.2 → 1.14.5). Whatever is merged must be green before
  moving on.
- Green is guaranteed at every merge commit and at the end of every version
  group, **not at every commit**: revert and rename commits between merges may
  not build. Bisect over merge commits only.
- Throwaway work to make intermediate versions green is acceptable. Its cost is
  negligible given libevm's criticality.

## Final acceptance check on upstream tests

After the last merge, the diff of upstream test and testdata files between the
result (in geth naming) and the merged upstream commit must be exactly the
baseline in [07](07-baselines-and-facts.md#upstream-test-baseline), minus rows
upstream has since absorbed. Anything else needs a stated reason.
