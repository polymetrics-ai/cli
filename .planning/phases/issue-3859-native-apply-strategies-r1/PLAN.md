# Plan — #3859 native database apply strategies

## Lifecycle and skills

Resolved and executed inline: `discuss-phase`, `plan-phase --tdd`; later
artifacts will record `execute-phase`, `verify-work`, and `code-review`.
Required skills loaded: `golang-how-to`, `golang-design-patterns`,
`golang-structs-interfaces`, `golang-naming`, `golang-error-handling`,
`golang-security`, `golang-safety`, `golang-testing`, `golang-context`,
`golang-concurrency`, and `golang-database`.

The CLI help/manual/website parity reference was read. No command, flag, help,
manual, website, or public connector capability changes are planned; the
implementation remains a private typed target path. The final diff must
re-check this exemption.

## Design

1. Extend the registered native polling target boundary with a sealed,
   descriptor-bound `ApplyPage` request and durable result. It carries bounded
   records, explicit tombstones, an exact mapping/control binding, and
   source-order evidence; it exposes no raw write primitive.
2. Make admission fail before any target session when the descriptor lacks the
   selected strategy, stable keys/order fence, history close-and-insert
   guarantee, or declared batch limits. Cancellation is checked before opening
   and between bounded apply units.
3. Reuse `DatabaseWritePlan`, `DatabaseWriteExecutor`, mapping, receipt, and
   delivery ledger for the transaction/acknowledgement boundary. Strengthen the
   sealed input as needed to retain source-order/idempotency data without
   exposing it as caller-controlled SQL or a generic write API.
4. Implement PostgreSQL's private target application using only verified
   managed relations and parameterized values. A newer source tuple wins for
   keyed modes; replay or older tuples cannot regress the retained current row.
   History applies a retry-safe close-current/insert-next transition and an
   explicit tombstone closes, rather than deletes, the current interval.
5. Declare only modes whose live driver proof exists. Keep public connector
   write capability false and do not add an API-surface operation.

## Acceptance evidence

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| `full_append` is bounded and preserves explicit duplicate semantics | live + fake | Reapplying the same two input rows produces the expected two additional durable rows, with fake batches never exceeding the declared record/byte limit. |
| `full_overwrite` publishes only a complete generation | live + fake | A bad/cancelled replacement preserves the old rows; a successful replacement leaves exactly the new input rows. |
| `incremental_append` does not pretend replays are deduplicated | live | Replayed input leaves both independently observable rows. |
| `incremental_upsert` has a stable-key source-order fence | live + fake | A newer tuple updates the keyed row; a late older tuple leaves that newer durable value unchanged. Physical absence leaves unrelated rows intact; an explicit tombstone is the only deletion. |
| `incremental_dedupe` retains the declared raw/current contract | live + fake | Replays do not create a second current identity and an older tuple cannot replace the current source-ordered record. |
| `incremental_dedupe_history` closes validity windows safely | live + fake | A newer record atomically closes the prior `_valid_to`/`_is_current` interval and inserts one current row; a soft-delete tombstone closes it without a physical delete. Late/replayed history events leave valid intervals intact. |
| Ack/partial/rollback/cancellation behavior is truthful | fake + live | Failed batch, rollback, missing receipt/ledger, cancellation, and unknown commit return no acknowledgement and leave the source checkpoint outside this slice unchanged. |
| Unsupported declarations are safe | fake | Each refusal asserts zero session opens, batch calls, commits, ledger writes, and target-row mutations. |

## TDD slices and checkpoints

1. **Plan checkpoint:** commit this delivery evidence before production edits.
2. **RED — closed page contract:** add target-apply tests that need the
   registered `ApplyPage`/ack path, descriptor batch bytes, order fence, and
   all six mode admissions. Assert every admission refusal has zero mutations.
3. **GREEN — shared sealed dispatch:** implement only the typed target request,
   immutable copies, exact registration, context/bound checks, and receipt gate.
   Run focused engine/database/synctransport tests.
4. **RED — PostgreSQL strategy state:** add the live-gated strategy scenario
   for the same mapping/input set. It must observe the six resulting target
   states, stale-order fence, atomic overwrite rejection, physical-absence
   retention, and tombstone history close before driver changes can pass.
5. **GREEN — private PostgreSQL strategies:** implement parameterized,
   transaction-bound application and reassert target/ledger evidence. No
   generic command or capability promotion.
6. **REFACTOR and review:** reduce duplication while retaining defensive
   copies, context checks, opaque errors, bounded resource use, and the public
   write fence. Record actual Red/Green output and independent review findings.

## Verification plan

- `go test -timeout 20m -count=1 ./internal/connectors/... ./internal/synctransport/...`
- focused race runs for the changed shared/database/PostgreSQL packages
- `POLYMETRICS_DATABASE_INTEGRATION=1 POLYMETRICS_CONTAINER_RUNTIME=docker POLYMETRICS_CONTAINER_ENDPOINT=unix:///Users/karthiksivadas/.colima/default/docker.sock go test -tags=databaseintegration -count=1 -timeout 20m -v ./internal/connectors/native/postgres`
- individual `make verify` gates: tidy-check, lint, docs-check, smoke-no-build,
  agent-contract-check, connectorgen-validate, connectorgen-surface-sync,
  connector-boundary, and release-workflow-check.
