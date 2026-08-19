---
phase: cli-3987-four-path-conformance-r1
issue: 3987
status: complete
coverage:
  - id: D1
    description: Four generated direction IDs resolve their exact production GitHub/PostgreSQL source and destination descriptors.
    verification:
      - kind: unit
        ref: internal/app/warehouse_flow_conformance_test.go:TestWarehouseMediatedFourPathConformance
        status: pass
    human_judgment: false
  - id: D2
    description: Each direction stages a connection-owned source page, reopens a sealed workset, applies it, independently reads it back, and checkpoints last.
    verification:
      - kind: unit
        ref: internal/synctransport/transport_test.go:TestOrchestratorDispatchesFourClosedPairingsWithoutPairBranches
        status: pass
    human_judgment: false
  - id: D3
    description: Closed mode evidence follows the current branch, including executable dedupe history and PostgreSQL-only change-capture source admission.
    verification:
      - kind: unit
        ref: internal/app/warehouse_flow_conformance_test.go:TestWarehouseMediatedModeConformance
        status: pass
    human_judgment: false
---

# Summary — #3987 four-path warehouse conformance

## Delivered

- Added an explicit four-direction GitHub/PostgreSQL conformance matrix derived from the generated `certificationcatalog.FlowKinds` catalog. Each named case resolves a distinct production source/destination executor and a persisted dispatch selection.
- Strengthened the shared orchestration proof to show the required owner-bound, sealed warehouse receipt/workset path and its execution order for every generated flow kind.
- Covered every closed `synccontract.Mode` according to current behavior: six executable transports pass; PostgreSQL `change_capture` is an implemented CDC source with a durable workset path, not a PostgreSQL destination mode.
- Recorded a schema-valid API→API source-binding defect that made only `api_to_api` fail, then restored the declaration and passed the targeted proof.

## Intentional divergence from stale issue text

Issue #3987 predates merged capability work. It asks for `incremental_dedupe_history` to be rejected, but PostgreSQL history support (PR #4187) and GitHub dedupe/history source support (PR #4188) are now executable. This gate proves that current behavior and does not unbuild it.

## Boundaries

This is deterministic CI conformance evidence, not a new live certification claim. It preserves the existing separate fresh-binary/live route evidence for API→API, API→database, database→API, and database→database. It does not change connector definitions, generic registration, certification profiles, source-stage roll-up, CLI dispatch/help, or final #3978 publication.
