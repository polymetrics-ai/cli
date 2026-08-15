# #4046 R9 — discussion log

> Audit trail only. Downstream execution uses `CONTEXT.md` and `PLAN.md`.

**Date:** 2026-08-11
**Mode:** `discuss-phase --auto` inline/manual fallback
**Reason for auto selection:** the authorized ship brief, complete R9 causal report, and amended #4046 acceptance addendum already lock all material product decisions.

## Typed stale-conflict terminalization

| Option | Description | Selected |
|---|---|---|
| General state refresh | Refresh any failed writer to the latest state before failure finalization. | |
| Typed-conflict-only rebase | On `errTransportStreamStateConflict`, mutate only the matching losing run in the latest locked state. | ✓ |
| Retry the checkpoint | Repeat the rejected checkpoint write. | |

**Decision:** use typed-conflict-only run terminalization.

**Rationale:** it makes the losing run truthful while retaining R7/R8's protection against stale checkpoint replacement. It neither promotes last-writer-wins behavior nor mutates the winner.

## Durable behavior proof

| Option | Description | Selected |
|---|---|---|
| Error-only assertion | Assert the stale-conflict error string. | |
| Durable two-App witness | Assert typed error, returned run identity/status, reopened loser state, winner state, and unrelated state. | ✓ |

**Decision:** retain and strengthen the deterministic two-App fixture as the RED/GREEN witness, then add restart, intervening-writer, cancellation, and seven-mode coverage.

## Scope fence

| Candidate | Decision |
|---|---|
| `internal/app/app.go` run finalization | In scope |
| `internal/app/transport_dispatch_test.go` focused test coverage | In scope |
| `internal/app/transport_dispatch.go` CAS logic | Read-only / out of scope |
| Provider, credential, network, warehouse, container, service, certification work | Out of scope |
| Old no-mistakes run `01KZQ0C1KEZRHNXX4WJFWXSCFB` | Immutable / never touch |

## Deferred ideas

None. The brief supplied strict non-goals and no unresolved product choice remains.
