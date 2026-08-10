# Context — Issue #3974: typed database connector foundation

## Discuss-phase record

`scripts/gsd prompt discuss-phase 3974 --auto` was resolved and reviewed before
planning. The issue is a shared foundation rather than a numbered roadmap phase,
so the installed GSD phase runtime cannot create its normal numbered artifacts.
This single-worker issue therefore uses the repository-approved manual inline
fallback. This context, the TDD ledger, and the verification checklist are the
durable GSD evidence.

## Contract inputs

- GitHub issue #3974, especially its **Acceptance and proof** section.
- `data/cli-postgres-parity-issue-tree-r1/report.md`: F1 is Wave A and unblocks
  F2 (#3981), F4 (#3975), and P1 (#3976).
- `data/cli-database-connector-framework-design-r1/report.md`: the typed
  `internal/connectors/database` foundation, closed definitions, no generic SQL,
  and the split between native admission and source execution.
- `data/cli-cdc-bidirectional-changefeed-design-r1/report.md` and
  `data/learnings.md`: no acknowledgement or durable-receipt claim may be
  created ahead of the committed fact it would prove.
- `data/captain.md`: target identity is always workspace + connector + connection
  ID; it never includes a display name or credentials.

## Captain warehouse-mediation amendment

The captain's 2026-08-10 ruling is binding on this foundation: every future
flow is source → warehouse → destination. The warehouse is the durable record;
zero-copy and direct source-to-destination paths are prohibited.

- **Layer one is shared and connector-agnostic.** The existing app-owned
  `runWarehouseETL` is the durable inbound materializer (connection-owned WAL,
  then published Parquet); the existing reverse path reads published Parquet.
  F1 does not recreate either operation or claim a new receipt/write boundary.
  `warehouse.ArtifactIdentity` and `warehouse.ArtifactRef` give that shared
  layer a typed, structural owner/table address, and `warehouse.Owner` delegates
  identity comparison to the same triple.
- **Layer two is database-specific.** A database can construct only a
  `WarehouseInboundRef` (`database source → warehouse artifact`) or a
  `WarehouseOutboundRef` (`warehouse artifact → database target`). The sealed
  database admission command contains exactly one of those legs; no value or
  executor in F1 receives both a source and target. A native-admitted driver
  supplies separate #3810 descriptor/evidence values for its distinct legs, so
  one descriptor cannot silently represent both directions.
- **MySQL seam.** A future MySQL author supplies only its own strict definition,
  driver descriptor/per-leg conformance, and native extraction/apply mechanics
  in the MySQL layer. The conformance test constructs both legs through the
  neutral artifact and database contracts without a PostgreSQL import or change
  to `internal/warehouse`, the generic app mediator, or the shared `database`
  package. This is illustrative type-level proof only; it does not declare a
  MySQL capability or executor.

## Scope boundary

This issue creates only the typed, non-executing database foundation:

1. strict optional `database.json` loading and immutable `Definition` projections;
2. closed logical types, lossless compatibility classification, structured catalog
   and read-plan values;
3. source/target references with the full owner identity;
4. a driver registry that separates a driver declaration from native executor
   admission; and
5. bounded resource policy validation and compile-time PostgreSQL driver seam;
   and
6. warehouse-bound database leg values that preserve the two-layer mediator
   without implementing any database or shared warehouse I/O.

It does **not** create a generic SQL API, database-shaped REST operation, target
DDL, write session, receipt, changefeed, polling executor, or PostgreSQL query.
`metadata.json` remains authoritative for public capability claims; PostgreSQL
`write` and `cdc` remain false.

## Coordination points

- #3810 supplies `synccontract.Mode` and native command/evidence semantics.
- #3857 owns polling/watermark execution and is not recreated here.
- #3864 owns transport dispatch and is not recreated here.
- F2 builds managed-target ownership on the identity/reference/catalog boundary.
- F4 builds committed transaction staging and receipts after this foundation;
  this issue must not fabricate a durability acknowledgement seam.
