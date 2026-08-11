---
phase: issue-4067-acknowledged-completion-rebase-r1
issue: 4067
status: r3_clean_after_disposition
depth: standard
mode: inline_manual_gsd_fallback
files_reviewed:
  - internal/app/app.go
  - internal/app/transport_dispatch.go
  - internal/app/transport_dispatch_test.go
findings:
  critical: 0
  warning: 0
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

## R3 code review record

**Review mode:** The resolved official `code-review` prompt still reports this named issue phase as absent from the numeric roadmap, and this custody forbids role spawning. The standard-depth review was therefore completed inline against the r3 committed source range and local evidence. It is neither independent review nor merge authorization.

### Scope and method

- Reviewed `24345ce1a0978a51bcc12b197f8d22c93cba1242..4f00ee8eb34ccf346d56125d6ad21d8031339880`, with particular attention to `RunETL → runTransportETL → failAcknowledgedTransportRun → failRun → updateState`.
- Checked that the new `errors.Is(errTransportStreamStateConflict)` branch is first, delegates only that typed condition to the established #4046 terminalizer, and leaves the acknowledged-witness finalizer available for every other error.
- Walked the committed two-App, two-page all-seven regression: page-one durable acknowledgement; unrelated writer; real winner checkpoint/run; stale page-two CAS; returned/reopened failed loser; typed identity; exact winner/unrelated preservation; and two loser applies with no replay.
- Rechecked the focused normal/race repeats, complete `internal/app`, GitHub definition/preflight/inspection/harness evidence, generator/quality gates, `git diff --check`, and JSON phase-state syntax. The credentialed smoke is intentionally unavailable rather than silently replaced.

### Findings and disposition

No r3 critical, warning, or informational finding. The early branch is neither a generic refresh nor a checkpoint retry: it reaches only #4046's typed-conflict `failRun`, whose callback edits only the target running run in the latest persisted state. The committed regression exercises the old page-one-witness mask directly and proves the direct path returns durable truth without overwriting the winner or unrelated state.

**R3 review disposition:** clean local review. Production registration remains unsupported/uncertified, the conditional credentialed smoke remains unavailable in this custody, and no #4015 wiring, provider, warehouse, or certification work is implied. Local-only no-mistakes loop 4/5 remains the next and final in-scope action before stopping for Firstmate.
