# GSD Pi native runtime contract

## Ownership

- GSD Pi owns milestone, slice, task, session, worktree, internal validation, and recovery state.
- Go Shepherd reads GSD only through supported headless events and `headless query`.
- Go Shepherd owns admission, generations, fencing, approvals, attestations, external-effect outbox,
  liveness, and normalized telemetry.
- Markdown artifacts are projections, never a competing controller ledger.

## Identity

One implementation issue maps to one GSD milestone, isolated worktree, issue branch, and primary PR.
Parent issues coordinate child milestones and own parent artifacts and integration decisions.

## Effects and gates

GSD sessions run without GitHub credentials and emit typed effect intents. Shepherd authorizes and
publishes idempotently after a current fence and exact-head attestation. No autonomous merge
capability exists. A human gate can be satisfied only by a matching explicit Go-recorded decision;
defaults, answer files, prior conversation, or an agent inference never satisfy it.

## Evidence

Ratification binds repository, PR, base/head SHA, run generation, unit attempt, governance state,
contract/evidence hash, observed validator identity, thinking level, gates, and expiry. Shepherd
rechecks the remote head and fence immediately before effect delivery.

## Privacy

Persist only allowlisted lifecycle and numeric evidence. Never persist raw prompts, reasoning,
credentials, command arguments, unrestricted tool output, or full upstream event objects.

