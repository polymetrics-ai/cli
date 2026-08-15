# #3864 discuss-phase record

## Inputs reviewed

- #3864 and #3862 issue/parent-PR metadata through `gh-axi`.
- `AGENTS.md`, the issue-first contract, parent topology contract, GSD Pi adapter,
  CLI parity reference, current connector canon and implementation procedure.
- #3810 plan/summary/verification artifacts and `internal/synccontract` contract
  source.
- `data/cli-sync-transport-topology-scout-r1/report.md` (2026-08-11).

## Resolved choices

| Area | Decision | Why |
| --- | --- | --- |
| Placement | `internal/connectors` owns descriptor/projection data; new `internal/synctransport` owns consumer interfaces, registry, preflight, and orchestration. | Avoids a `connectors` import cycle and keeps provider adapters out of this child. |
| Conformance | External verifier dependency; default unavailable verifier fails closed. | The existing #3810 evidence list is self-reported metadata, not an executed proof. |
| Pair model | Fake API/database transports pass only through a typed warehouse-stage seam. | Current canon forbids direct source → destination hops. |
| Mode ownership | Accept exactly `synccontract.Mode`; no conversion or second enum. | #3810 is the sole semantic owner. |
| Apply behavior | Per-mode declared, closed apply strategy selected in preflight before a source read. | Removes new-path hard-coded `upsert` dispatch without inventing generic writes. |
| Checkpoint | Commit only after a destination returns #3810's opaque durable acknowledgement. | Keeps the durable acknowledgment/checkpoint seam single-sourced. |
| Public surface | Manual and JSON inspection show descriptor-derived eligibility or explicit unsupported roles; every structurally valid destination remains declared, including `acknowledgement: none`. | Read-only inspect must be honest and machine-readable without admitting a non-durable destination to runtime. |

## Open dependency facts recorded, not solved here

- #3810 needs executable corpus/evidence binding before a real transport can be
  admitted truthfully.
- #3859 and provider adapters own real destination apply strategies.
- #3865, #3867, and #3866 remain separate issues. This child provides only the
  narrow fake-backed dispatch/preflight seams they will consume.
