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
