# #4067 — discussion log

> Audit trail only. Execution uses `CONTEXT.md` and `PLAN.md`.

**Date:** 2026-08-11
**Mode:** `discuss-phase --auto` inline/manual fallback
**Reason:** the Sol audit and controlling correction directive have already decided every material product and safety choice.

## Completion eligibility

| Option | Description | Selected |
|---|---|---|
| Refresh all stale Apps | Rebase any stale finalization on latest whole state. | |
| Acknowledged-target rebase | Rebase only a still-running run whose latest stream equals its acknowledged checkpoint. | ✓ |
| Retry checkpoint/destination work | Repeat previously acknowledged work to create a fresh revision. | |

**Decision:** strict acknowledged-target rebase only.

## #4046 boundary

| Option | Description | Selected |
|---|---|---|
| Broaden typed failure exception | Treat ordinary completion revision conflicts like `errTransportStreamStateConflict`. | |
| Separate #4067 completion rule | Preserve #4046's typed-conflict-only failure boundary. | ✓ |

**Decision:** separate #4067 child issue and completion-only rule.

## Evidence standard

| Option | Description | Selected |
|---|---|---|
| Error-code test | Assert only a revision conflict. | |
| Durable two-App witness | Assert returned/durable identity, reopen truth, target/winner and unrelated preservation, error chain, cancellation, all modes, and race. | ✓ |

**Decision:** behavioral RED/GREEN with real persisted state.

## Generated outputs

| Output | Decision |
|---|---|
| `website/lib/docs.generated.ts` | Regenerate only with its canonical generator. |
| `internal/connectors/certifications/flow-matrix.json` | Regenerate only with its canonical certification-matrix generator. |

## R2 Sol-audit continuation

**Mode:** `discuss-phase --auto` inline/manual continuation. The complete Sol r2 handoff and the amended #4067 acceptance contract lock the decisions; the official numeric-phase lookup reports `phase_found: false`, so no lifecycle role may be spawned in this custody lane.

| Area | Options considered | Selected |
|---|---|---|
| Post-ack error result | Discard the result and use ordinary `failRun`; retain a post-ack witness only for guarded failure finalization. | Retain only the acknowledged witness. |
| Failure rebase scope | Generic stale-state refresh; exact stream/run guarded failure finalizer. | Exact acknowledged stream plus still-running run only. |
| Missing target run | Return ordinary `run not found`; return typed revision conflict with no mutation. | Typed revision conflict and `Run{}`. |
| RED evidence | Error-code-only tests; persisted two-App all-mode witnesses with reopen and apply-count assertions. | Persisted all-mode witnesses. |

**Decision:** F1 and F2 are a narrow continuation of #4067. Preserve #4046, R7/R8, the winner checkpoint, unrelated state, and the existing #4059-only delivery lane.

## Deferred ideas

None. Generic state refresh, checkpoint/CAS redesign, #4046 behavior changes, provider work, generated-artifact expansion, and merge remain out of scope.

## R3 Sol-audit continuation

**Mode:** `discuss-phase --auto` inline/manual continuation. The r3 audit, amended #4067 issue, and updated loop-4 directive resolve the material choices; named-phase lookup remains absent and the custody contract forbids lifecycle-role spawning.

| Area | Options considered | Selected |
|---|---|---|
| Typed second-page conflict | Revalidate the stale page-one witness; send only the typed conflict to #4046 `failRun`. | Direct typed-conflict delegation before witness lookup. |
| State mutation | Retry or overwrite the losing checkpoint; terminalize only the matching losing run through the existing typed-conflict path. | Existing #4046 terminalization only. |
| RED evidence | Single-page error assertion; two-page persisted all-mode witness with a real winner, reopen, and apply counts. | Deterministic two-page all-seven witness. |
| Connector evidence | Treat a provider smoke as proof of core correctness; separate current-branch definition/preflight/inspection evidence from deterministic core proof. | Separate core and existing-connector gates. |
| Credentialed smoke | Request/copy a token or invent a run; use only an already-approved secret channel and bounded read-only operation. | Conditional sanctioned-channel smoke only. |

**Decision:** preserve r2 behavior for ordinary post-ack errors, preserve #4046's typed-conflict-only exception, and make no production-wiring or certification claim when transport roles are not registered at this branch.
