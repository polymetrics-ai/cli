## Intent

Establish the typed, non-executing database connector foundation required by
the PostgreSQL parity programme.

Refs #3974

## Proof

- Strict `database.json` loading, closed logical types, bounded resources,
  structured identity/catalog/read-plan types, and defensive projections.
- Driver registration requires an exact registered driver and native admission;
  no declaration alone becomes executable.
- PostgreSQL is a compile-time reference seam only. Public `write` and `cdc`
  capability flags remain false.
- Captain mediation ruling: database contracts can form only one typed leg at a
  time — database source → shared warehouse artifact, or shared warehouse
  artifact → database target. They cannot encode a direct source/destination
  pair, and F1 adds no executor, write session, receipt, or zero-copy path.
- Layer one remains connector-agnostic: the existing warehouse materializer
  writes WAL then Parquet, the existing reverse path reads published Parquet,
  and `warehouse.ArtifactIdentity`/`ArtifactRef` provide their shared typed
  address. PostgreSQL owns neither of those values.
- A future MySQL author changes only layer two: its strict `database.json`,
  descriptor/conformance, and native extraction/apply code and tests. The
  `TestMySQLLayerTwoReferenceCompilesAgainstSharedWarehouseArtifact` proof uses
  no PostgreSQL import and requires no change to `internal/warehouse`, the
  generic app mediator, or the shared database contract. It is a type-level
  seam demonstration, not a MySQL capability declaration.
- GSD discuss → plan (TDD) → execute → verify → review evidence is under
  `.planning/phases/issue-3974-typed-database-foundation/`.
