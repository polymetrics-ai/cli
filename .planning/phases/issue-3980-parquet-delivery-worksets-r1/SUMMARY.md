---
phase: issue-3980-parquet-delivery-worksets-r1
status: complete
coverage:
  - id: D1
    description: immutable StreamID-keyed real-Parquet delivery workset
    verification:
      - kind: integration
        ref: internal/connectors/database/parquet_delivery_workset_test.go:TestDeriveChangeDeliveryWorksetImmutableIdentity
        status: pass
    human_judgment: false
  - id: D2
    description: keyed delta and explicit tombstone behavior
    verification:
      - kind: integration
        ref: internal/connectors/database/parquet_delivery_workset_test.go:TestDeriveChangeDeliveryWorksetDerivesRealParquetDeltaAndExplicitTombstones
        status: pass
    human_judgment: false
  - id: D3
    description: bounded, cancellation-safe artifact derivation
    verification:
      - kind: integration
        ref: internal/connectors/database/parquet_delivery_workset_test.go:TestDeriveChangeDeliveryWorksetCancellationCleansStagingArtifacts
        status: pass
    human_judgment: false
---

# Summary — Issue 3980: immutable Parquet delivery worksets

## Delivered

- Added a sealed `ChangeDeliveryWorkset` in `internal/connectors/database`.
  It derives its destination identity exclusively from the #3981
  `ManagedTargetDeliveryLedgerKey`, then binds target schema/key fingerprint,
  source/baseline versions, explicit tombstones, counts, and artifact hash.
- Materializes a complete immutable projection, keyed insert/update delta,
  explicit-tombstone stream, and unpromoted candidate baseline under a
  content-addressed directory. Source and supplied baseline files remain
  unchanged.
- Requires a finite per-artifact ceiling, validates inputs/outputs, uses
  bounded buffers, cleans staging on cancellation, fails closed on corrupt
  reuse, and recognizes the warehouse's zero-byte empty-table representation.

## Scope retained

- No target DML, database driver, receipt persistence, source checkpoint, or
  baseline promotion was added. Those operations remain for #3983/#3973 after
  a target receipt exists.
- No CLI, connector-definition, credentials, documentation, or capability
  surface changed.

## GSD and skills

- Inline/manual GSD lifecycle completed: `discuss-phase` → `plan-phase --tdd`
  → `execute-phase` → `verify-work` → `code-review`. The canonical
  single-worker contract prevents role spawning, so artifacts record the
  fallback.
- Skills used: `golang-how-to`, `golang-design-patterns`,
  `golang-structs-interfaces`, `golang-error-handling`, `golang-security`,
  `golang-safety`, `golang-testing`, `golang-database`, and `golang-lint`.
