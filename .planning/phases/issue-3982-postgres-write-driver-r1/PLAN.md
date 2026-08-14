# PLAN — Issue #3982 PostgreSQL managed-table write driver

## GSD and skills record

- Adapter health and contract check passed before planning. All five required lifecycle prompts were resolved through `scripts/gsd prompt` and will be carried out inline under the canonical no-delegation contract.
- Required skills used: `golang-how-to`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-testing`, `golang-context`, `golang-concurrency`, and `golang-database`.
- Runtime/dbtest instructions, migration conventions, architecture design, connector canon, and the #3973/#3981 completed plans were read. CLI help/manual/website parity is not applicable: no CLI surface or connector definition changes are in scope.

## Foundation check

| Need | Proof before claim | Result |
| --- | --- | --- |
| Typed provisioning contract | `internal/connectors/database/managed_target_provisioning.go` | Present; PostgreSQL will implement its three methods only. |
| Typed write-session contract | `internal/connectors/database/database_write_session.go` | Present; PostgreSQL will implement `DatabaseWriteDriver` and `WriteSession` only. |
| Delivery ledger contract | `internal/connectors/database/managed_target_delivery_ledger.go` | Present; PostgreSQL will own durable backing storage behind its store port. |
| PostgreSQL live harness | `internal/connectors/native/dbtest` | Present; opt-in direct-local endpoint test will exercise state, not only errors. |
| Capability promotion | #3978 certification | Intentionally absent; `write=false` is preserved. |

## Discovered foundation gap — decision required

The current contract cannot express the input that the required PostgreSQL DDL
and five keyed modes must prove:

- `DatabaseWritePlan` binds only definition/control/mode/keys/count/batch/effect;
  it carries no target columns, logical types, mapping contract, or input schema.
- `ManagedTargetSchema` contains only a version and fingerprint. A fingerprint
  proves equality after an independently known layout, but cannot render the
  initial typed relation or encode a record value.
- `WriteBatch` contains `[]connectors.Record` only. It has no typed tombstone
  envelope, delete key payload, or managed-row schema; `synccontract.Tombstone`
  cannot reach `ApplyWriteBatch` through the current port.
- `defs/postgres/database.json` admits no target modes, so a sealed plan for
  any requested PostgreSQL mode is refused before a driver can run.

Adding a PostgreSQL-local map-key inference or JSONB fallback would violate the
issue's exact type-mapping, explicit-tombstone, and shared-contract
requirements. Adding fields or a second mapping/tombstone protocol in this
connector lane would be a shared foundation change, which the connector canon
requires to be split before continuing.

Decision options:

1. deliver/identify the missing shared typed mapping + managed-row/tombstone
   contract as a foundation issue, then resume #3982 against it;
2. explicitly authorize that foundation scope to be added here (requiring a
   scope/ownership exception); or
3. narrow #3982 to the PostgreSQL provisioning/control/ledger driver only and
   move typed five-mode writes to the new foundation-dependent issue.

## Decision resolved — partial execution

Firstmate assigned the missing contract back to #3973 as
`cli-3973-mapping-contract-r2`. This issue may now implement the native
provisioning and durability half without making a private mapping protocol.
The provisioning driver must refuse a create that needs a business relation
layout until the shared mapping contract supplies one; no placeholder relation
may be persisted because it would either lie about schema identity or force the
later slice to auto-evolve it. The non-mapping work is deliberately retained:

1. real advisory lock, target database/namespace/relation OID observation, and
   typed private control-record decoding;
2. private control-layout decoding and durable ledger storage exercised against
   independently seeded live state; first-create control DDL remains atomic
   with the mapping-derived business relation;
3. foreign/missing/unreadable owner, collision, database replacement, relation
   OID replacement, and schema-drift refusals with no target/control mutation;
4. PostgreSQL durability preflight, requiring `fsync=on` and a transaction
   that can establish `synchronous_commit=on` before a future write session;
5. compile-time port checks while leaving `DatabaseWriteDriver`, write-mode
   admission, record application, tombstones, receipts, and `write=false`
   capability behavior unchanged until #3973 is complete.

## Mapping contract landed — resumed execution

PR #4144 completed #3973's missing `MappingContractV1`, explicit
`TombstoneEnvelope`, and `DeliveryReceiptV1` contracts. The mapping is sealed
into `DatabaseWritePlan`, but the previously landed provisioning plan cannot
yet carry the exact mapping that PostgreSQL must render for first-create DDL.
This issue will make the smallest shared typed attachment: an optional,
defensively copied `MappingContractV1` on `ManagedTargetProvisioningPlan`.
The native PostgreSQL adapter will require it for first create, while existing
mapping-free provisioning callers remain valid and continue to fail closed at
the native adapter. This keeps all column/type authority in the shared contract
and avoids a driver-local mapping, placeholder relation, or schema evolution.

The resumed implementation therefore owns:

1. mapped, atomic PostgreSQL namespace/control/relation creation with relation
   and namespace OIDs re-observed by the existing provisioner;
2. closed PostgreSQL DDL/value encoding directly from `MappingContractV1`, with
   unsupported logical shapes and values refusing before mutation;
3. `DatabaseWriteDriver`/pinned transaction implementation for exactly the
   five phase-one modes, bounded batches, transactional overwrite, and explicit
   keyed tombstone deletes; and
4. live dbtest assertions for created rows/types/control state, all modes,
   physical absence retention, explicit deletes, rollback, durability and
   ownership refusals.

## TDD slices

1. **Red — concrete driver surface.** Add compile-time and real PostgreSQL test coverage for the existing descriptor-only driver to require provisioning, preview, session start, batch application, receipts, and ledger persistence. Preserve `write=false` / legacy `Connector.Write` fence assertions. Capture failing output.
2. **Green — private ownership observation and provisioning port.** Implement advisory locking, independent private-control decoding/ledger storage, database identity, namespace OID, relation OID, and schema fingerprint. First-create DDL remains coupled to the future mapping-derived relation layout, so the live truth table presently seeds its oracle state independently and asserts no driver mutation for foreign/unreadable ownership, collision, OID replacement, schema drift, and permissions.
3. **Red/green — type mapping and bounded transaction.** Establish a closed logical-type/value encoder and relation DDL from the sealed schema; then implement `PreviewDatabaseWrite` and one pinned session with bounded batch application. Live tests assert exact rows/types and zero mutation on unsupported values.
4. **Red/green — closed modes and deletion.** Implement canonical `full_overwrite`, `full_append`, `incremental_append`, `incremental_upsert`, and `incremental_dedupe` behavior through that session. `full_overwrite` publishes atomically; keyed tombstones explicitly delete; no dedupe-history or physical-absence deletion is accepted. Test output checks rows and counts.
5. **Red/green — certainty and durability.** Preflight `fsync` and `synchronous_commit`; test rollback for statement/batch/cancel failures, commit receipt durability, and an explicit unknown-commit outcome with no retry.
6. **Integration/refactor.** Exercise concurrent owner isolation and the built-binary approved API-to-PostgreSQL path with dbtest when the named direct endpoint is available. Run race, scoped static, connector-canon/preflight, GSD verification and review; use the gap loop only for a verified uncovered criterion.

## Guardrails

- Production scope is limited to `internal/connectors/native/postgres/**`, the
  corresponding PostgreSQL `database.json` admission/test, the narrow
  `ManagedTargetProvisioningPlan` mapping attachment required to avoid a
  driver-local DDL map, and this issue's planning evidence. No unrelated shared
  write semantics are changed.
- Use explicit driver-owned constant SQL identifiers only for fixed private control structures; render every derived identifier through a closed quoting helper and bind every value as an argument.
- No credentials, DSNs, raw records, or server error text enter error messages, plans, traces, or PR body.
- No new dependencies, generic SQL executor, connector registration, capability flip, schema evolution, unrelated source transport change, or fixes for #4125/#4136/#4090.

## Commit checkpoints

1. Plan/context/TDD evidence.
2. Preserved Red test output.
3. Green implementation and focused unit/live evidence.
4. Verification/review or gap-only fixes with focused proof.
