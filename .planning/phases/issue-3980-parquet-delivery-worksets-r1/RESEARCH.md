# Research — Issue 3980

## Evidence reviewed

- Live issue #3980 and the accepted PostgreSQL parity decomposition in
  `issue-3985-connector-canon-r1/POSTGRES-3972-BODY.md`.
- `internal/connectors/database/managed_target_delivery_ledger.go` and its
  tests: `ManagedTargetDeliveryLedgerKey` is the complete immutable target
  address and has no source artifact table/display dependency.
- `internal/connectors/database/managed_target.go`: a managed target derives
  from StreamID while the source `ArtifactRef.Table` remains provenance only.
- `internal/warehouse/parquet.go`: DuckDB is the repository's sole Parquet
  implementation and already provides real fixture materialization/readback.
- `internal/synctransport/types.go`: generic `WarehouseWorkset` is stage-owned
  dispatch data and intentionally lacks target identity, schema/key binding,
  baseline, and receipt semantics.

## Findings

- A database-package concrete value can consume `ManagedTargetControlRecord`
  and its ledger key without creating the warehouse-to-database import cycle.
- The implementation needs no new package: embedded DuckDB is already a
  declared dependency and is the required Parquet engine.
- A complete Parquet source snapshot must be copied into the workset. Holding a
  source path, input map, or caller-owned tombstone byte slice would make a
  supposedly sealed workset mutable after derivation.
- DuckDB can derive a bounded on-disk delta by joining source and baseline on
  explicitly validated key columns. The query must use only internally quoted
  column identifiers and escaped file literals; no caller SQL is accepted.
- Physical absence is deliberately not compared into tombstones. Only validated
  `synccontract.Tombstone` inputs are persisted in a deterministic order.

## Package legitimacy audit

No dependency is proposed or added.
