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

## Supplementary live canary preflight

A clean persistent checkout was created at `/private/tmp/shepherd-481-cli-architecture-canary`
because the preserved historical #397 worktree contains globally ignored `.DS_Store` files and
was not mutated or cleaned. Exact target evidence passed at #438 head
`21d195aff0c7bd60b3bf54f14b1ce165cec9e03f` (open/draft, GitHub `DIRTY`, visible checks green).

Read-only generation 1 (`run-689079d0-089c-4d72-ab78-3dd1b213923e`) supplied useful behavior RED:
the scout lane completed exact reconciliation, while the concurrent validator halted because the
extension independently initialized a Pi OAuth runtime for each embedded session. Hard gates were
`lane_execution_failed,run_halted`; score was 0.000. #438 was not mutated or merged, no credential
value was requested/read/printed, and deprecation remains inactive. The parent owns the bounded
single-flight runtime correction because it blocks Shepherd dispatch itself; this does not count as
#481 production implementation or satisfy R1-R6.

## Parent-preflight GREEN / live canary generation 2

Parent preflight tests passed 5/5 focused and 168/168 affected with strict pinned-Pi typecheck,
family verification, and offline RPC. Generation 2
(`run-ae39456f-e034-4204-b23b-bd8e076b251e`) then completed both model lanes without the OAuth
initialization error. Both independently classified exact #438 as not merge-ready because it is
still draft and GitHub `DIRTY`; the run failed closed with score 0.000 rather than halting. It did
not mutate #438 and it is not a passing canary/deprecation receipt.

## GREEN / live canary / review

Pending #480 integration and worker handoff. Keep both durable generations as evidence. The live
read-only reconciliation is supplementary exact-state evidence, not a replacement for deterministic
R1-R6 RED/GREEN, a passing bound synthetic canary receipt, or human approval.
