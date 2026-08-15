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
| Atomic first create, exact reassertion, and durable control ledger | live | dbtest PostgreSQL returns the newly created namespace/relation OIDs; the test reads one owner/control row and the mapped `tenant text`, `id bigint`, `value text` columns, then repeat assertion returns the same OID. One ledger store increases its private-table row count and re-reads the identifier. |
| Missing/foreign/tampered owner, collision, OID replacement, schema drift, and permissions | live | Every refusal re-queries namespace/relation OIDs, owner/control values, and ledger count after the call. The independently created ownerless namespace, restricted role, and tampered control assertions have no driver mutation. |
| Unsafe durability settings | live | A session visibly has `synchronous_commit=off`, preflight refuses it, then accepts the restored safe setting. |
| Private namespace/control DDL and first target creation | live | `MappingContractV1` travels on the typed provisioning plan; PostgreSQL creates namespace, three private controls, relation, OID records, and schema fingerprint in one transaction. Unsupported arrays leave no namespace/relation/control state. |
| Five typed modes and explicit tombstones | live | dbtest reads returned PostgreSQL rows after full/incremental append, dedupe, composite-key upsert, and overwrite. An absent row remains until a typed tombstone deletes its exact composite key. |
| Rollback and unknown commit | live + narrow timing fake | A PostgreSQL CHECK failure, invalid mapped value, and cancellation after an actual write batch leave the pre-overwrite rows intact. A wrapper closes a real second pgx connection after native apply and before commit; the shared executor reports `unknown` and opens exactly one session. |
| Capability remains fenced until certification | local | Existing metadata/capability tests observe `write=false` after the driver gains mapping-independent ports. |

## Fixed decisions

- The only target connector is `postgres`; shared write-session and managed-target contracts from #3973/#3981 are consumed. The provisioning plan gains only a defensively copied optional `MappingContractV1` attachment, because first-create DDL otherwise has no shared schema authority.
- Names remain only the deterministic opaque `ManagedTargetRef` components. No DDL accepts a caller-provided schema, relation, SQL fragment, or destination display name.
- The driver must retain one session-bound transaction per approved write and must surface only typed shared errors at the boundary.
- `write` remains `false`; no registration, CLI surface, generic SQL, auto-evolution, physical-absence delete, or `incremental_dedupe_history` may be introduced.
- The execution is the GSD inline/manual fallback: the canonical worker contract disables role delegation and this issue is not a numbered roadmap phase. This records the generated `scripts/gsd prompt` lifecycle rather than waiving it.

## Resumed mapped scope

#4144 landed the shared `MappingContractV1`, `TombstoneEnvelope`, and
`DeliveryReceiptV1` contracts. This driver carries the shared mapping from
provisioning DDL through preview and the session, with no PostgreSQL-local
field vocabulary. Its five internal admitted modes are deliberately distinct
from the connector's public `write=false` capability fence.
