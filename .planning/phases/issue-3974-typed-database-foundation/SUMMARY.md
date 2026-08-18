---
coverage:
  - id: D1
    description: Strict, secret-safe, versioned database.json loading and immutable definition projections.
    verification:
      - kind: unit
        ref: internal/connectors/database/database_test.go:TestDatabaseDefinitionStrictLoadAndDefensiveProjection
        status: pass
    human_judgment: false
  - id: D2
    description: Closed logical types and exact/lossless-only type planning without a text fallback.
    verification:
      - kind: unit
        ref: internal/connectors/database/database_test.go:TestLogicalTypeCompatibilityIsLosslessOrRejected
        status: pass
    human_judgment: false
  - id: D3
    description: Structured identity, catalog, fingerprint, and deterministic bounded read-plan contracts.
    verification:
      - kind: unit
        ref: internal/connectors/database/database_test.go:TestStructuredCatalogIdentityAndReadPlanAreStable
        status: pass
      - kind: unit
        ref: internal/connectors/database/database_test.go:TestResourcePolicyBoundsEveryDatabaseResource
        status: pass
    human_judgment: false
  - id: D4
    description: Registered-driver and shared native-admission boundary remains distinct from source execution.
    verification:
      - kind: unit
        ref: internal/connectors/database/database_test.go:TestDriverAdmissionRequiresRegisteredCompatibleNativeAdmission
        status: pass
      - kind: unit
        ref: internal/synccontract/native_admission_test.go:TestNativeAdmissionIsNotItselfASourceRunner
        status: pass
    human_judgment: false
  - id: D5
    description: PostgreSQL exposes only a compile-time database-driver seam and no write/CDC capability promotion.
    verification:
      - kind: unit
        ref: internal/connectors/native/postgres/database_driver_test.go:TestPostgresDatabaseDriverReferenceSeam
        status: pass
      - kind: unit
        ref: internal/connectors/engine/database_definition_test.go:TestBundleLoadPostgresDatabaseDefinitionWithoutCapabilityPromotion
        status: pass
    human_judgment: false
  - id: D6
    description: Connector-agnostic warehouse artifact identity and isolated database-to-warehouse / warehouse-to-database legs prevent a direct database pair; distinct native admissions cannot cross legs, and a MySQL-shaped layer-two implementation compiles against the shared seam.
    verification:
      - kind: unit
        ref: internal/warehouse/artifact_test.go:TestArtifactRefIsConnectorAgnosticAndStructurallyBound
        status: pass
      - kind: unit
        ref: internal/connectors/database/database_test.go:TestWarehouseMediationUsesSharedArtifactAndSeparateDatabaseLegs
        status: pass
      - kind: unit
        ref: internal/connectors/database/database_test.go:TestMySQLLayerTwoReferenceCompilesAgainstSharedWarehouseArtifact
        status: pass
    human_judgment: false
---

# Summary — Issue #3974: typed database connector foundation

Implemented the Wave A F1 typed database foundation in
`internal/connectors/database/`. PostgreSQL now supplies a strict,
policy-only `database.json` reference declaration that is loaded with its
existing bundle, but it does not register an executor, open a connection,
create a target, write data, emit a receipt, run CDC, or promote a public
capability.

## Delivered contracts

- A strict versioned definition loader and checked-in closed schema, with
  defensive projections and secret-safe errors.
- Closed logical/native type representations and a compatibility planner that
  permits only exact or proven-lossless mappings.
- Structured connection, catalog, relation, column, key, fingerprint, and
  non-executing read-plan values.
- Finite database resource policy for page, batch, pool, timeouts, and bind
  parameters.
- A shared native descriptor/evidence admission interface split from
  source-side `RunNativeSync`, plus an exact registered database-driver
  admission registry.
- PostgreSQL's compile-time reference driver and the optional bundle loading
  seam, with `capabilities.write=false` and `capabilities.cdc=false` retained.
- A captain-mandated two-layer mediator seam: neutral warehouse artifact/owner
  identity in `internal/warehouse`, plus database-only inbound and outbound
  legs that cannot contain a direct source/target pair. Per-leg native
  admissions prevent a single descriptor from standing in for both directions;
  the amendment has passed its refreshed validation. It does not add a
  warehouse executor or database I/O.

## TDD and lifecycle evidence

The initial focused Red run is retained in `traces/red-run.txt`; the package
and engine Green run is retained in `traces/green-run.txt`. The official GSD
prompts for `discuss-phase`, `plan-phase --tdd`, `execute-phase`,
`verify-work`, and `code-review` were generated through `scripts/gsd`. Because
this issue is not a numbered roadmap phase and the supplied runtime lacks the
required isolated worker role, the repository-approved manual inline fallback
is recorded in `PLAN.md`, `TDD-LEDGER.md`, `VERIFICATION.md`, and
`RUN-STATE.json`.

## Scope boundary

F2/F3/F4/F5 and all P-unit work remains absent: there is no managed target
ownership assertion/provisioning, write session, durable receipt/checkpoint,
CDC/changefeed, generic SQL executor, database-shaped REST operation, or
second sync-mode enum.
