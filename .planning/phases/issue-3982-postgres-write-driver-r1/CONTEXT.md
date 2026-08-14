# Context — Issue #3982 PostgreSQL managed-table write driver

## Task Delivery Header

- Issue: Refs #3982 — Postgres Parity: implement the managed-table write driver
- Base branch: `integration/4015-mvp-flat-r1`
- Merges into: `integration/4015-mvp-flat-r1 → main`
- Delivery: PR open against `integration/4015-mvp-flat-r1` with its checks green.
- Working branch: `fm/cli-3982-postgres-write-driver-r1`
- Task: Implement PostgreSQL's private managed-target namespace, typed control/ledger persistence, provisioning port, and bounded transactional write-session port behind the shared database contracts. Preserve the `write=false` capability fence.
- Verification: `go test -timeout 20m ./internal/connectors/native/postgres/... ./internal/connectors/database/...`, the opt-in dbtest PostgreSQL proof when its direct local endpoint is available, targeted race checks, and scoped repository verification gates.

## Evidence Table

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| First creation and repeat assertion create only an owned target | live | A dbtest PostgreSQL database shows the derived namespace, relation, and control row after first provision; repeating returns the same relation OID/control record without creating an additional relation. |
| Five allowed modes write exact typed values, composite keys, and atomic overwrites | live | Returned PostgreSQL rows and `COUNT(*)` prove final values/types for `full_overwrite`, `full_append`, `incremental_append`, `incremental_upsert`, and `incremental_dedupe`; a concurrent reader observes either old or full replacement data, never a partial overwrite. |
| Unowned, foreign, unreadable, colliding, replaced, or schema-drifted targets are refused | live | Each dbtest negative case re-queries relation/control state and proves its pre-existing rows and control records are unchanged. |
| Unsupported types/values, permission denial, and unsafe durability settings are refused before mutation | live | Each refusal re-queries target and delivery control state and proves zero changed rows/control records; permission and GUC cases use a restricted/live session. |
| Statement error, batch error, and cancellation roll back the whole transaction | live | Seeded target and ledger state remain unchanged after each forced failure/cancellation; a post-failure row query and control query prove zero partial writes. |
| A commit disconnect has an unknown outcome and is not retried | fake | A deterministic connection/transaction seam is necessary to fault only the commit acknowledgement boundary; it asserts one commit attempt, no retry/rollback claim, and no receipt/ledger mutation. Live tests cover confirmed commit durability. |
| Concurrent writers retain owner scope and no partial target provisioning | live | Two concurrent dbtest clients yield one fully asserted target/control pair for the same owner, while cross-owner attempts leave the other owner's namespace/relation/control state absent. |
| The approved reverse-ETL path reaches PostgreSQL through the typed contract | live | A built `pm` integration test executes the approved path and queries both target rows and the driver-owned control/delivery records. |
| Capability remains fenced until certification | live | Existing metadata/capability tests observe `write=false` after the driver becomes executable behind the typed contract. |

## Fixed decisions

- The only target connector is `postgres`; shared write-session and managed-target contracts from #3973/#3981 are consumed, not amended or duplicated.
- Names remain only the deterministic opaque `ManagedTargetRef` components. No DDL accepts a caller-provided schema, relation, SQL fragment, or destination display name.
- The driver must retain one session-bound transaction per approved write and must surface only typed shared errors at the boundary.
- `write` remains `false`; no registration, CLI surface, generic SQL, auto-evolution, physical-absence delete, or `incremental_dedupe_history` may be introduced.
- The execution is the GSD inline/manual fallback: the canonical worker contract disables role delegation and this issue is not a numbered roadmap phase. This records the generated `scripts/gsd prompt` lifecycle rather than waiving it.
