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
| Exact existing target assertion and durable control ledger | live (dbtest; endpoint pending) | Independently seeded PostgreSQL owner/control rows reassert the same relation OID; one ledger store increases its private table count by one and re-reads the identifier. |
| Foreign/tampered owner, collision, OID replacement, schema drift, and permissions | live (dbtest; endpoint pending) | Each case re-queries namespace/relation OIDs, owner/control values, and ledger count after refusal; no driver mutation is accepted as evidence. |
| Unsafe durability settings | live (dbtest; endpoint pending) | A session visibly has `synchronous_commit=off`, preflight refuses it, then accepts the restored safe setting. |
| Private namespace/control DDL and first target creation | held for #3973 mapping | The target relation and its schema fingerprint must commit atomically; no placeholder or PostgreSQL-private mapping is legal. |
| Five typed modes, tombstones, rollback, and unknown commit | held for #3973 mapping/receipt | A later dbtest suite must assert persisted row/count/receipt outcomes rather than errors alone. |
| Capability remains fenced until certification | local | Existing metadata/capability tests observe `write=false` after the driver gains mapping-independent ports. |

## Fixed decisions

- The only target connector is `postgres`; shared write-session and managed-target contracts from #3973/#3981 are consumed, not amended or duplicated.
- Names remain only the deterministic opaque `ManagedTargetRef` components. No DDL accepts a caller-provided schema, relation, SQL fragment, or destination display name.
- The driver must retain one session-bound transaction per approved write and must surface only typed shared errors at the boundary.
- `write` remains `false`; no registration, CLI surface, generic SQL, auto-evolution, physical-absence delete, or `incremental_dedupe_history` may be introduced.
- The execution is the GSD inline/manual fallback: the canonical worker contract disables role delegation and this issue is not a numbered roadmap phase. This records the generated `scripts/gsd prompt` lifecycle rather than waiving it.

## Resumed partial scope

Firstmate confirmed that #3973 owns the missing `MappingContractV1` and
`DeliveryReceiptV1` completion lane. Until it lands, this issue proceeds only
with PostgreSQL-private control storage, advisory locking, durable identity
observation, fail-closed ownership/OID/schema refusal behavior, and durability
preflight. It does not infer a record layout, create a placeholder business
table, admit a write mode, or apply/delete records.
