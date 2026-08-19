---

## Classification

**Cross-flow integration and conformance gate** for the PostgreSQL parity tree. This issue owns the
missing proof that a connector implements only its own side of a persisted warehouse flow. It does
not create a direct connector-to-connector transport, generic API executor, generic SQL executor, or
second mode vocabulary.

## Scope

Make the warehouse-mediated flow and mode matrix executable before final PostgreSQL certification:

- persist/resolve the selected warehouse table or immutable workset before a source or destination
  connector is invoked; a source may write only to its connection-owned warehouse state and a target
  may read only a sealed warehouse workset;
- exercise the four canonical paths as distinct contracts: API read → warehouse → API write, API
  read → warehouse → PostgreSQL write, PostgreSQL read → warehouse → API write, and PostgreSQL read
  → warehouse → PostgreSQL write;
- cover the closed `internal/synccontract.Mode` vocabulary without inventing support: prove the five
  phase-one target modes (`full_overwrite`, `full_append`, `incremental_append`,
  `incremental_upsert`, and `incremental_dedupe`); prove `incremental_dedupe_history` is rejected as
  non-executable; and prove `change_capture` is admitted only as the PostgreSQL source contract and
  reaches a target through its derived warehouse workset, initially with `incremental_upsert`;
- preserve plan → preview → explicit approval → execute for every mutating target and when a
  persisted schedule invokes an approved flow; a schedule repeats a flow and cannot form a direct
  source-to-target hop;
- integrate the existing one-engine dispatch seam in #3864 rather than creating a PostgreSQL-only
  router, and give the final certification gate machine-readable flow/mode evidence.

API sides are admitted only where their own typed connector contracts exist. This issue does not
claim API CDC or API target support merely because the database matrix names an API leg.

## Dependencies

Hard dependencies: #3976, #3982, #3977, #3979, and #3983, plus their shared foundations. Coordinate
with #3864; it supplies dispatch, while this issue supplies the missing warehouse-only conformance
matrix. This issue gates #3978.

## Acceptance and proof

- [ ] Tests fail if a source receives a destination connector reference, a target receives a live
      source connector/credential, or a route skips the named warehouse table/workset.
- [ ] The four canonical paths prove record counts, provenance, approval state, receipts, and retry
      behavior at the warehouse boundary; an API → API proof does not promote PostgreSQL capability.
- [ ] The mode matrix records an executable result for each of the five supported target modes and a
      typed, specific non-executable result for `incremental_dedupe_history`; it never treats a
      filename, declaration, or exit status as proof.
- [ ] `change_capture` proof uses PostgreSQL 14+ `pgoutput` v2 through the bounded stage and a
      receipt-before-acknowledgement warehouse contract. API polling or a cursor is refused as its
      fallback.
- [ ] Flow plan/preview/approval state stays valid across scheduler invocation, restart, target
      retry, unknown commit, and stale workset/owner/schema changes.
- [ ] Focused contract/integration tests, GSD/TDD Red/Green evidence, and a reviewed production PR
      are green. Live container certification remains #3978's responsibility.
