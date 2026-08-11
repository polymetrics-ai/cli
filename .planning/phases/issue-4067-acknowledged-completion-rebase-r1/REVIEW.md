---
phase: issue-4067-acknowledged-completion-rebase-r1
issue: 4067
status: r2_clean_after_disposition
depth: standard
mode: inline_manual_gsd_fallback
files_reviewed:
  - internal/app/app.go
  - internal/app/transport_dispatch.go
  - internal/app/transport_dispatch_test.go
findings:
  critical: 0
  warning: 1
  info: 0
  unresolved: 0
---

# #4067 R2 code review record

**Review mode:** The official `code-review` prompt was resolved. `gsd-sdk` reports `phase_found: false` for this nonnumeric issue phase and custody forbids role spawning, so standard-depth review was completed inline. It is local review evidence only, not a certification, external CI, or merge authorization.

## Scope and method

- Reviewed the r2 diff from rejected `3f84693bfbc128523a66e22653db7227fb9c0869` through the current source checkpoint, following `RunETL → runTransportETL → failAcknowledgedTransportRun → failRunWithAcknowledgedTransportState → updateState` and the parallel acknowledged-completion rebase.
- Checked the acknowledgement witness origin, exact stream equality, running-run requirement, mutation footprint, missing-run typed chain, error joins, commit-outcome behavior, and no-replay boundary.
- Reviewed the persisted two-App tests, all-seven cancellation/missing-run table, source-error witness, outcome witness, and review-driven fail-closed cases. `git diff --check` is clean.

## Findings and dispositions

| ID | Severity | Finding | Disposition |
|---|---|---|---|
| CR-001 | Warning | C8/C9 proved success outcomes, but did not directly exercise the acknowledged-error finalizer when the exact stream is changed/removed or the exact run is terminal/removed. | Resolved in `e8a541a7` with `TestRunETLTransportAcknowledgedFailureFailsClosedWhenTargetChanges`; all four cases prove zero return, original source error plus typed conflict, one apply, and unchanged latest state. |

## Conclusions

- The post-ack error result is retained only after the real checkpoint callback completes. Pre-ack errors still use ordinary `failRun`.
- The dedicated finalizer changes only the exact still-running run's failed terminal fields after current state contains the exact acknowledged stream. It does not refresh broad App state, retry or overwrite a checkpoint, or replay source/destination work.
- A missing target run receives `errStateRevisionConflict` only in the acknowledged rebase branch; the ordinary missing-run behavior is unchanged.
- #4046's typed-conflict-only terminalization, R7/R8 source identity, and per-stream CAS remain outside the r2 production diff and pass fresh focused regressions.

**Review disposition:** no unresolved critical, warning, or informational finding. Fresh local-only no-mistakes loop 3/5 subsequently passed; its unrelated warehouse-documentation commit is cancelled in the scope-restoration follow-up. The #4059-only delivery route still requires Firstmate direction because current tool help has no safe existing-PR option.
