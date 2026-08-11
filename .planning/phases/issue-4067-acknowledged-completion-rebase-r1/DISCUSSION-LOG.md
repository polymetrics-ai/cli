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

## Deferred ideas

None. Generic state refresh, checkpoint/CAS redesign, #4046 behavior changes, provider work, and merge remain out of scope.
