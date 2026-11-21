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
