# 8. Rules of engagement

Standing instructions from the maintainer to whoever (human or agent) performs
the sync.

## Do not start

- The sync does not start until the maintainer says so, even if everything
  looks ready.

## Stop and ask; never just move on

- A libevm test cannot be kept **semantically unchanged**.
- A transform at a hook site **cannot be named** (its shape or intent is
  unclear).
- Upstream now **natively provides** what a hook exists for.
- A **new flaky test** appears, i.e. one not in a package already in CI's
  `FLAKY_REGEX`.
- Anything would require changing an **upstream test** beyond the baseline.
- #280 reaches the upstream change to `IntrinsicGas()`: decide, then report
  which way and why.

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
