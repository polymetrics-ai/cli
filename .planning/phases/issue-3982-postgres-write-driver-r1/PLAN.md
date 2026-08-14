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

## TDD slices

1. **Red — concrete driver surface.** Add compile-time and real PostgreSQL test coverage for the existing descriptor-only driver to require provisioning, preview, session start, batch application, receipts, and ledger persistence. Preserve `write=false` / legacy `Connector.Write` fence assertions. Capture failing output.
2. **Green — private ownership and provisioning.** Implement namespace/control/ledger DDL and observation under an advisory lock, with database identity, namespace OID, relation OID, and schema fingerprint. The live truth table must assert state is unchanged for missing/foreign/unreadable ownership, collision, OID replacement, schema drift, and permissions.
3. **Red/green — type mapping and bounded transaction.** Establish a closed logical-type/value encoder and relation DDL from the sealed schema; then implement `PreviewDatabaseWrite` and one pinned session with bounded batch application. Live tests assert exact rows/types and zero mutation on unsupported values.
4. **Red/green — closed modes and deletion.** Implement canonical `full_overwrite`, `full_append`, `incremental_append`, `incremental_upsert`, and `incremental_dedupe` behavior through that session. `full_overwrite` publishes atomically; keyed tombstones explicitly delete; no dedupe-history or physical-absence deletion is accepted. Test output checks rows and counts.
5. **Red/green — certainty and durability.** Preflight `fsync` and `synchronous_commit`; test rollback for statement/batch/cancel failures, commit receipt durability, and an explicit unknown-commit outcome with no retry.
6. **Integration/refactor.** Exercise concurrent owner isolation and the built-binary approved API-to-PostgreSQL path with dbtest when the named direct endpoint is available. Run race, scoped static, connector-canon/preflight, GSD verification and review; use the gap loop only for a verified uncovered criterion.

## Guardrails

- Production scope is limited to `internal/connectors/native/postgres/**` plus this issue's planning evidence. If the existing shared contracts prove insufficient, record `needs-decision` instead of editing them.
- Use explicit driver-owned constant SQL identifiers only for fixed private control structures; render every derived identifier through a closed quoting helper and bind every value as an argument.
- No credentials, DSNs, raw records, or server error text enter error messages, plans, traces, or PR body.
- No new dependencies, generic SQL executor, connector registration, capability flip, schema evolution, unrelated source transport change, or fixes for #4125/#4136/#4090.

## Commit checkpoints

1. Plan/context/TDD evidence.
2. Preserved Red test output.
3. Green implementation and focused unit/live evidence.
4. Verification/review or gap-only fixes with focused proof.
