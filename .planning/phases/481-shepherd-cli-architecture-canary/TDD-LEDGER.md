# TDD Ledger: #481

## Plan-first state

- Status: dependency-blocked by #480.
- Production edits: not started.
- GSD mode: existing manual programming-loop fallback.
- Review budget: one comprehensive Codex 5.6-sol xhigh round; at most one correction pass.

## Required RED

| ID | Failing behavior required before production edit |
|---|---|
| R1 | canary rejects moved/mismatched #397/#438 repository, branch, PR, or exact head evidence |
| R2 | deterministic topology runs two disjoint lanes concurrently and serializes dependency/collision lanes |
| R3 | restart preserves worktree/effect/human-gate ownership without duplicate publication or consumption |
| R4 | no path can mutate or merge #438 or expose parent/default-branch merge authority |
| R5 | secret-bearing/control/oversized fixture data is rejected and absent from persisted/output evidence |
| R6 | deprecation activation rejects a missing, stale, failed, or different-head canary receipt |

The worker records exact failing command/counts. Missing-module RED alone is insufficient; planned
assertions must execute against a compiling scaffold and fail for their intended behavior.

## GREEN / live canary / review

Pending #480 integration and worker handoff. The live read-only canary is supplementary exact-state
evidence, not a replacement for deterministic RED/GREEN or human approval.
