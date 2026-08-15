# Context — Issue 3983: deliver immutable worksets to managed PostgreSQL targets

## Task Delivery Header

- Issue: Refs #3983 — Postgres Parity: deliver derived worksets to managed targets
- Base branch: `integration/4015-mvp-flat-r1`
- Merges into: `integration/4015-mvp-flat-r1` → `main`
- Delivery: Pull request open against `integration/4015-mvp-flat-r1` with green checks; its API-reported base is read back after creation.
- Working branch: `fm/cli-3983-workset-delivery-r1`
- Task: Consume a sealed `ChangeDeliveryWorkset` through the shared mapping/write contracts and the PostgreSQL managed-target driver. Seal destination identity and workset into a previewable, explicitly approved delivery; apply only its keyed delta and explicit tombstones; persist its per-destination baseline only after a durable receipt.
- Verification: Targeted `go test -timeout 20m ./internal/connectors/native/postgres/... ./internal/connectors/database/... ./internal/warehouse/...`, the supplied Docker dbtest command, scoped repository gates, and green PR CI.

## Evidence Table

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| Sealed workset delivery applies changed rows by composite key. | live | A real Parquet workset reaches the dbtest PostgreSQL target; queries observe inserted and updated keyed rows while unchanged rows retain their value. |
| Physical absence never deletes. | live | The first real target write inserts a row; a later source projection omits it without a tombstone; a query still returns that row. |
| Explicit tombstones alone delete. | live | A sealed tombstone removes exactly its keyed target row; the query no longer returns it. |
| Approval binds the exact workset/destination plan. | fake | A narrow driver/session fake counts session opens and writes. A changed workset/destination approval is rejected with both counters zero; a fake is necessary to prove pre-mutation refusal deterministically. |
| Receipt, baseline, replay and unknown commit semantics are explicit. | fake + live | A narrow receipt/baseline fake proves failed, receipt-store-failed, and unknown commit paths retain the previous baseline and use the same workset identity on replay. The live test queries receipt/ledger state after a committed delivery. |
| Destinations cannot cross ownership boundaries. | fake | Two independently sealed controls/baseline stores prove distinct ownership keys and baseline entries; invalid cross-owner/control drift is rejected before the driver is called. |
| Real built delivery path is bounded. | live | The PostgreSQL dbtest uses the real managed driver and real Parquet input with the sealed finite batch size; its queries observe the resulting target rows and durable receipt. |

## Decisions

- #3980's `ChangeDeliveryWorkset`, #3973's `MappingContractV1`/`DeliveryReceiptV1`, and #3982's managed PostgreSQL driver are authoritative. This slice adds no private parallel representation and no generic SQL/direct connector route.
- The target connector is exactly native PostgreSQL. Any work beyond the shared consumer seam, native PostgreSQL driver integration, their tests, and this evidence is out of scope; #4125, #4136, and #4090 are explicitly excluded.
- The delivery controller remains synchronous and reuses the existing `DatabaseWriteExecutor` for the single approved session. It materializes only the sealed workset's delta and explicit tombstones into `DatabaseWriteInput`; a missing projection row is never turned into a delete.
- The approval is generated from the native preview of a complete immutable plan that includes workset identity/content and the asserted control/mapping/key bindings. An altered workset, owner, target OID, schema, or key order is stale and must be refused before session opening.
- A candidate baseline is promoted by a per-destination store only after `DatabaseWriteExecutor` returns a ledger-backed `DeliveryReceiptV1`. Batch failure, receipt/ledger failure, and unknown commit leave the prior baseline unchanged and surface a replay-required result; no blind retry occurs after unknown commit.

## GSD execution and required skills

- Passed `scripts/gsd doctor`; resolved prompts for `discuss-phase`, `plan-phase --tdd`, `execute-phase`, `verify-work`, and `code-review`; passed `go run ./cmd/agentcontractgen check`.
- The canonical single-worker/Firstmate delivery contract forbids spawning lifecycle roles, so the generated GSD prompts are executed inline. These phase files record the required manual fallback.
- Loaded: `golang-how-to`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-testing`, `golang-database`, `golang-context`, `golang-concurrency`, and `golang-lint`.
- No user-facing command, help topic, generated connector surface, or website documentation is planned to change. CLI help/manual/website parity is therefore not applicable unless tracing identifies an existing CLI surface that needs a behavior change.
