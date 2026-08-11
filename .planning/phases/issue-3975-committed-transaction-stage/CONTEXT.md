# Context — Issue #3975: committed-transaction staging and durable receipts

**Gathered:** 2026-08-11
**Status:** Ready for TDD planning

## GSD execution note

`scripts/gsd prompt discuss-phase 3975 --auto` was resolved and executed
inline. `gsd-sdk query init.phase-op 3975` reports `phase_found: false` because
the active roadmap deliberately delegates connector work to issue-specific
artifacts rather than numbered roadmap phases. This is the repository-approved
manual-GSD fallback: these artifacts are the durable discuss/plan/execute/
verify/review record. No decision below reopens the issue or parent contracts.

## Phase Boundary

Implement the source-agnostic, private committed-transaction stage used by a
future database CDC reader. It owns begin/append/commit/abort boundaries,
bounded streaming persistence, recovery and orphan cleanup, immutable
whole-transaction receipts, and acknowledgement eligibility only after receipt
durability.

The owned production scope is a new private-stage implementation and tests
under `internal/connectors/database/`. It does not modify PostgreSQL decoding,
source LSN acknowledgement, polling, managed targets, destination DML, generic
SQL, CDC capabilities, or public CLI surfaces.

## Implementation Decisions

### Durability boundary

- **D-01:** A staged chunk is private and invisible until an explicit source
  transaction commit. A locally fsynced stage is recoverable storage only; it
  is never itself a delivery receipt or an acknowledgement fact.
- **D-02:** Commit hands the complete ordered transaction to a narrow,
  injected durable-receipt port. The port returns an immutable receipt only
  after its downstream transaction/materialization is durably accepted. The
  stage persists that receipt before it reports acknowledgement eligibility.
- **D-03:** Future source acknowledgement must consume the receipt-derived
  eligibility, not a caller-constructed boolean, a stage directory, a source
  commit observation, or a generic success result. This slice does not send an
  LSN acknowledgement.

### State, recovery, and cleanup

- **D-04:** Keep each transaction in a private transaction-specific stage.
  Append is streaming and ordered; it must not retain all chunks or payloads in
  memory. Payload/envelope/checkpoint/tombstone/history meanings remain those
  of `internal/synccontract`; this package creates no competing vocabulary.
- **D-05:** Recovery removes pre-commit/orphan temporary state without
  publication. It retains a sealed committed transaction that lacks a durable
  receipt so a caller can resume its delivery or replay from the prior durable
  source position. A persisted receipt makes a transaction eligible for
  cleanup; recovery cleans leftover sealed files only after validating that
  receipt.
- **D-06:** The receipt is immutable evidence of one whole transaction: stable
  transaction identity, ordered content digest, byte/record totals, and the
  downstream receipt identity/timestamp. It contains no raw payload and is
  defensively copied on all API boundaries.

### Failure containment and quotas

- **D-07:** Enforce finite byte, record, and elapsed-time limits before an
  incomplete transaction can consume unbounded memory or disk. A limit breach
  is the named `TransactionStageLimitExceeded` outcome, publishes nothing,
  creates no receipt/eligibility, leaves source progress untouched, and never
  triggers a cursor or polling fallback.
- **D-08:** Filesystem operations are injectable for tests. Every durable
  transition checks write, file sync, atomic rename, and parent-directory sync
  failures. Cancellation is checked before and during streaming and cannot
  produce a receipt or acknowledgement eligibility.
- **D-09:** Transaction identities are treated as opaque provider data and
  cannot become path traversal. Filesystem accounting includes recovered active
  stages so a restart cannot evade declared storage limits.

### TDD and verification

- **D-10:** The first RED test is exactly
  `TestCommittedTransactionStagePublishesOnlyAfterDurableReceipt`. It must fail
  before production implementation, then pass after GREEN.
- **D-11:** Tests use isolated temporary directories and failure-injecting
  filesystem/receipt fakes. They assert artifacts, ordered published chunks,
  durable receipts, acknowledgement eligibility, and zero temporary/orphan
  residue rather than only exit status.
- **D-12:** Required skills recorded for this phase: `golang-how-to`,
  `golang-design-patterns`, `golang-structs-interfaces`,
  `golang-error-handling`, `golang-security`, `golang-safety`,
  `golang-testing`, `golang-context`, `golang-concurrency`,
  `golang-database`, `golang-lint`, and `no-mistakes`.

### Agent discretion

- Use the smallest internal API that expresses the above proof. Reuse existing
  repository filesystem-durability helpers when their failure semantics match;
  do not add dependencies or a generic database/write interface.
- Keep `internal/connectors/native/postgres/cdc.go` fail-closed. A later #3977
  owner wires this source-agnostic stage to pgoutput v2 and source feedback.

## Canonical References

Downstream agents must read these before planning or implementing.

### Issue, topology, and delivery contract

- `AGENTS.md` — mandatory lifecycle, storage durability, path-safety, and
  review gates.
- `.agents/agentic-delivery/canonical/delivery-contract.json` — single-worker
  lifecycle, no-mistakes argv, and stacked-sub-PR state vocabulary.
- `.agents/agentic-delivery/contracts/issue-agent-contract.md` — TDD and PR
  evidence requirements.
- `.agents/agentic-delivery/references/gsd-pi-adapter.md` — adapter command
  path and inline manual fallback requirement.
- `/Users/karthiksivadas/karthik-agent-workspace/data/cli-postgres-parity-topology-scout-r1/report.md`
  — accepted #3975 scope, #3974 dependency, first RED test, and forbidden
  PostgreSQL/CDC work.
- `/Users/karthiksivadas/karthik-agent-workspace/data/cli-connector-release-certification-r1/implementation-brief-template.md`
  — child branch/PR, correction-cap, Shepherd, and no-mistakes contract.

### CDC durability semantics

- `data/cli-cdc-large-transaction-strategy-r1/report.md` — bounded durable
  stage is not a receipt; StreamCommit/Abort, quota, and replay ordering.
- `data/cli-cdc-bidirectional-changefeed-design-r1/report.md` — warehouse
  extraction receipt boundary and receipt-before-source-acknowledgement rule.
- `data/cli-database-connector-framework-design-r1/report.md` — durable
  receipt construction and acknowledgement must be derived from the fact it
  proves.
- `/Users/karthiksivadas/karthik-agent-workspace/data/learnings.md` — dominant
  premature-acknowledgement defect class and required crash-durability review.
- `data/captain.md` — binding captain rulings and warehouse-only topology.

### Existing contracts and integration seam

- `internal/synccontract/commit.go` — unforgeable downstream acknowledgement
  shape and checkpoint ordering to reuse, not duplicate.
- `internal/synccontract/state.go` — canonical source/checkpoint/envelope
  vocabulary and defensive opaque-token handling.
- `internal/synccontract/tombstone.go` — canonical tombstone/history semantics.
- `internal/synccontract/recovery.go` — typed recovery/rebootstrap vocabulary.
- `internal/connectors/database/resources.go` — finite resource-policy pattern.
- `internal/connectors/database/registry.go` — #3974 closed descriptor and
  native-admission truth that this stage must not widen.
- `internal/connectors/native/postgres/cdc.go` — current fail-closed consumer
  boundary; do not enable or alter it in this issue.

## Existing Code Insights

### Reusable assets

- `synccontract.DownstreamAcknowledgement` and
  `synccontract.NewDurableDownstreamAcknowledgement` protect checkpoint
  persistence from caller-forged acknowledgement structs.
- `synccontract.CheckpointEnvelope`, `Tombstone`, and recovery outcomes already
  provide shared CDC state vocabulary with defensive opaque-byte cloning.
- `database.ResourcePolicy` demonstrates finite, validated limits and
  context-aware operation bounds.

### Established patterns

- The database foundation keeps concrete constructors, small consumer-owned
  interfaces, explicit resource bounds, and immutable defensive copies.
- Local warehouse durability uses fsync plus atomic replacement and must prove
  parent-directory durability; a stage transition must preserve the same
  crash-consistency discipline.
- Native PostgreSQL CDC stays unsupported before its transaction stage exists;
  this foundation must not flip a capability just because it compiles.

### Integration points

- A future #3977 PostgreSQL pgoutput-v2 reader will create/append/seal/abort
  source transactions and adapt a receipt-derived eligibility to existing
  checkpoint/standby-status machinery.
- A future warehouse materializer supplies the durable-receipt port. This
  issue defines no warehouse table, target DML, or source feedback call.

## Specific Ideas

The stage must make large transactions bounded without claiming exactly-once
delivery. A crash before its durable downstream receipt remains an at-least-once
replay from the prior durable source position.

## Deferred Ideas

- PostgreSQL pgoutput v2 protocol, decoder, source LSN feedback, slot health,
  and rebootstrap UX belong to #3977.
- Immutable Parquet/DuckDB delivery worksets belong to #3980.
- Snapshot/changefeed bootstrap belongs to #3979.
- Managed target sessions, commits, and destination DML belong to #3973/#3982.
