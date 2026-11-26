# 8. Rules of engagement

Standing instructions from the maintainer to whoever (human or agent) performs
the sync.

## Do not start

- The sync does not start until the maintainer says so, even if everything
  looks ready.

<a id="decide-or-stop"></a>

## Decide, or stop and ask

**Default: decide.** If these documents cover the situation, through a rule, a
priority order or a recorded precedent, use that context, proceed, and record
the decision where a reviewer will find it (the merge commit message).

**MUST stop and ask, always, when either holds:**

1. **It is a brand-new decision**, not covered anywhere in these documents.
2. **It involves a semantic change to anything libevm**, tests or otherwise.

Never work around either, and never just move on.

**Err on the side of caution.** Proceed without asking only when absolutely
sure a change to a libevm file is purely syntactic. In particular, alert the
maintainer when:

- **Upstream has rewritten the code under a hook**, so the libevm side must be
  restructured, even if the observable behaviour and the tests are unchanged.
- **libevm's exported API is forced to change**, e.g. a hook's signature must
  follow a changed upstream signature. Consumers break, so this is a stop.

Known instances of the two rules (not exhaustive):

- A libevm test cannot be kept semantically unchanged.
- An adaptation to a libevm non-test file would change what it does, not just
  how it is spelled.
- A transform at a hook site cannot be named (its shape or intent is unclear).
- Upstream now natively provides what a hook exists for.
- An upstream test would need changing beyond the baseline.

Covered by context, so decide and report:

- #280 reaching upstream's change to `IntrinsicGas()`: choose revert or
  re-seat, and say which and why.
- A flaky test in a package already in CI's `FLAKY_REGEX`: re-run and continue.

Flag without necessarily stopping:

- A **new flaky test**, i.e. one not in a package already in `FLAKY_REGEX`.

## Never silence a finding

A lint, vet or compiler finding is fixed at its cause, never suppressed.

- An error that a call newly returns is **checked** (`require.NoError` in a
  test), not discarded with `_ =` and not hidden with a `nolint` directive.
- If the proper fix seems to conflict with another rule here, for example it
  adds an assertion to a libevm test, that is a reason to **ask**, not to pick
  whichever option avoids the question.

This rule exists because of a mistake in group 2 of the first sync; see
[11](11-sync-log.md).

## Test failures must be useful

libevm and SAE follow the Go Style Decisions
(https://google.github.io/styleguide/go/decisions.html), in particular its
section on useful test failures. For any test written or edited in a sync:

- **Identify the function.** A failure message names what was called,
  qualifying a method with its receiver's type via `%T`:

  ```go
  if err := prefetcher.prefetch(owner, root, addr, keys, false); err != nil {
  	t.Fatalf("%T.prefetch() error %v", prefetcher, err)
  }
  ```

  not a bare `"prefetch()"`.
- **Identify the input where it helps.** Inputs are not always easy to
  communicate or relevant; include them when they are.
- **`t.Error` and friends are welcome.** The maintainer often uses them
  directly because `testify` is unnecessarily verbose. They are typically
  coupled with a diff, reported only when non-empty: `cmp.Diff()` for values
  and `testerr.Diff()` for errors.

  ```go
  if diff := cmp.Diff(want, got); diff != "" {
  	t.Errorf("%T.Foo() diff (-want +got):\n%s", x, diff)
  }
  if diff := testerr.Diff(err, tt.wantErr); diff != "" {
  	t.Fatalf("%T.Foo() %s", x, diff)
  }
  ```

  Copy the exact message form from a neighbouring test. Follow the style of
  the file being edited; do not convert between the two.

A failure must tell the reader what went wrong without opening the test.

## Novel work must be identifiable

A sync adapts existing code; it should rarely add behaviour. When it does, on
the maintainer's instruction, the new work must be obviously the performer's
own and not an adaptation of existing libevm code:

- tests written for it go in `*.libevm_claude_test.go` files (or the
  equivalent for another performer), never in existing test files;
- it is a separate commit after the merge commit, whose message says it is
  novel work and who chose the behaviour.

## Exceptions are never precedent

When the maintainer allows a departure from these rules, record it in
[11](11-sync-log.md) with the reasons given. **It applies to that case only.**
Do not map its circumstances onto a later situation, however similar: always
ask again.

## Sharing a conflict that cannot be resolved

When a conflict needs the maintainer, share it by whichever of these fits, in
this order of preference:

1. **In conversation.** Usually the blocker is a decision, not the text. Give
   the conflicting hunk with all three sides (base, upstream's, ours), the
   [catalogue](10-catalogue.md) entry it belongs to, and the options with a
   recommendation.
2. **The recipe to reproduce it.** A conflict is fully determined by its two
   inputs: the tip of `gethsync` and the upstream commit being merged. Checking
   out the first and merging the second gives the identical conflict locally.
3. **A branch with the conflict committed as-is**, for anything large. Commit
   the tree with the conflict markers left in and push it to
   `gethsync-blocked` on the fork. Never to `gethsync` itself. Reuse that one
   branch name and overwrite it each time.

The answer comes back either as a resolution stated in conversation, or as a
branch the maintainer pushes to the fork. Either way it goes into the real
merge commit, whose message says the resolution was the maintainer's decision.

Conflict markers are always diff3 style; see
[05](05-process.md#diff3).

## How to report

Every status report to the maintainer SHOULD take this form. No stream of
consciousness, and no partial reports while work is still in flight: wait
until the report can be complete.

- **One numbered list, in priority order.**
- **Ranges labelled by kind**, stated up front: which items are *decisions*
  (an answer is needed, and they block), which are *alerts* (done under
  existing rules; speak up only to disagree) and which are *informational*.
- **Markdown.**
- **Every reference to code is a link relative to the repository root, with
  line numbers**: `[core/state/statedb.go:211](core/state/statedb.go#L211)`,
  or a range `#L211-L223`. State which commit the line numbers refer to.
- **Delivered as a file, not as text to copy.** Publish it with
  `tools/publish-report.sh`, which puts it at the root of `gethsync-blocked`
  as `GETHSYNC_REPORT.md`, on top of the commit it describes, overwriting the
  previous report. Pulling that branch makes every link open in an IDE. No
  separate download is wanted.
- The report commit is never part of `gethsync`.
- **A difference from a reference merge is never informational.** Where the
  result differs from the maintainer's own earlier merge (the post hoc oracle;
  see [05](05-process.md)), in either direction, the difference is a deliberate
  choice by one party that the other did not make. Each one is its own item,
  as a *decision* (default) or an *alert* (only if an existing rule or an
  earlier answer already settles it, which the item must cite). Each states
  what the reference did, what this sync did, why, and what follows from
  either choice. They are never grouped under one item or summarised.

In conversation, say only that the report is ready, where it is, and how many
decisions it contains.

## Report, don't hide

- Any rewind or redo of history: say what changed and why, before replacing
  anything that already exists.
- Conflict resolutions are the performer's own decisions: make them obvious and
  own them.
- State what was verified and what was only inferred.
- Assume neither party's logic is correct; re-check claims that the plan rests
  on.

## Access and pushing

- **Never push to `ava-labs/libevm`.** Read only.
- Push only to the fork (`ARR4N/go-ethereum`); blast radius there is accepted.
- Push after every green checkpoint, so at most one version group can be lost.
- Keep durable knowledge on the `sync-tooling` branch (`libevm/tooling/gethsync/`), not only in conversation.

## Commits

- **Every commit MUST be signed** (branch rule set). SSH signing is assumed
  acceptable. Fallback if signatures do not verify: the maintainer re-signs the
  whole branch once, at the end (trees and structure unchanged, hashes change).
- Merges are real merge commits. Never squash, never linearise.
- The commit message of a merge carries the lists described in
  [05](05-process.md#merge-commit-contents).

## Priorities when in doubt

1. Minimise disruption to upstream files.
2. Keep transforms unchanged.
3. Absorb the rest in libevm files.
4. Both oracles green at every merge.

## Tone of the working relationship

- Ask clarifying questions when necessary; challenge a faulty premise, but only
  when the whole premise is faulty.
- Succinct answers; acknowledge with a bare ACK when asked.
