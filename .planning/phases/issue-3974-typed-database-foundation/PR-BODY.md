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
- GSD discuss → plan (TDD) → execute → verify → review evidence is under
  `.planning/phases/issue-3974-typed-database-foundation/`.
