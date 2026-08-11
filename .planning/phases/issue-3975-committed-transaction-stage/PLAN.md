# PLAN — Issue #3975: committed-transaction staging and durable receipts

## GSD setup and manual fallback

- `scripts/gsd doctor` passed in the supplied isolated worktree.
- Resolved command sources and generated prompts for `discuss-phase`,
  `plan-phase --tdd`, `execute-phase`, `verify-work`, and `code-review`.
- `go run ./cmd/agentcontractgen check` passed before planning.
- `scripts/gsd prompt discuss-phase 3975 --auto` and
  `scripts/gsd prompt plan-phase 3975 --tdd --auto` were executed. The
  official runtime reports `phase_found: false` because #3975 is an
  issue-specific foundation rather than a numbered active-roadmap phase.
- The canonical delivery contract forbids spawning a researcher, planner,
  checker, executor, verifier, reviewer, Shepherd, or extra worker. This is
  therefore the permitted manual inline GSD fallback; CONTEXT/RESEARCH/PLAN,
  the ledger, run state, verification, UAT, and review files are the durable
  evidence.
- Required skills loaded: `golang-how-to`, `golang-design-patterns`,
  `golang-structs-interfaces`, `golang-error-handling`, `golang-security`,
  `golang-safety`, `golang-testing`, `golang-context`, `golang-concurrency`,
  `golang-database`, `golang-lint`, `no-mistakes`, `gsd-discuss-phase`,
  `gsd-plan-phase`, `gsd-execute-phase`, `gsd-verify-work`, and
  `gsd-code-review`.

## Goal

Add a source-agnostic committed-transaction stage that streams private source
chunks to bounded durable storage, exposes nothing before commit, produces one
immutable receipt only after downstream durability, and makes acknowledgement
eligible only from that receipt.

## Allowed implementation surface

- `internal/connectors/database/transaction_stage.go` — stage types, limits,
  receipt port, recovery, and filesystem transition implementation.
- `internal/connectors/database/transaction_stage_test.go` — public behavior,
  ordering, quota, cancellation, restart, and acknowledgement proofs.
- `internal/connectors/database/transaction_stage_fault_test.go` — injected
  durable-I/O and crash-boundary failure cases.
- `internal/connectors/database/database_test.go` only if a package-level
  foundation invariant belongs in the existing shared suite.
- `.planning/phases/issue-3975-committed-transaction-stage/**` — lifecycle,
  TDD, verification, Shepherd-compatible, no-mistakes, and review evidence.

## Explicit exclusions

- No PostgreSQL decoder, `pgoutput` version change, replication slot operation,
  source LSN acknowledgement, cursor/polling fallback, or CDC capability flip.
- No managed target, destination DML, generic SQL, database connection, reverse
  ETL execution, immutable workset, or Parquet/DuckDB materialization.
- No new module/dependency, credential, secret, public CLI/help/docs/website
  surface, or modification to `internal/connectors/native/postgres/cdc.go`.
- No replacement or duplicate of #3810 checkpoint/envelope/tombstone/history
  vocabulary or #3745/#3746 descriptor/admission truth.

## Design contract

### Public package boundary

Create an explicit `CommittedTransactionStage` constructor in package
`database` with a private stage root, finite `TransactionStageLimits`, injected
clock/filesystem durability seams, and a narrow committed-transaction receipt
port. Keep constructors concrete and interfaces consumer-owned/small.

The API must express:

1. `BeginTransaction(ctx, transactionID)` creates a private active stage.
2. `AppendChunk(ctx, transactionID, records, reader)` streams one opaque chunk
   to that transaction in source order.
3. `CommitTransaction(ctx, transactionID, receiver)` seals the stage, streams
   every complete chunk in order to the receiver, persists its receipt, and
   returns a receipt-derived acknowledgement eligibility.
4. `AbortTransaction(ctx, transactionID)` removes staged chunks and leaves no
   visible publication, receipt, or eligibility.
5. `AdmitRecoveredTransaction(ctx, transactionID)` is the explicit caller or
   operator decision required before a recovered sealed item may commit.
6. startup/recovery exposes sealed no-receipt work as held, removes
   incomplete/orphan temporary state safely, and never directly replays it.
7. `ReconcileDiscardControls(ctx)` retries only fail-closed discard cleanup;
   it never admits or delivers staged work.

No API accepts raw source paths or publishes a generic writer. Transaction
identifiers remain opaque data and are mapped to deterministic safe storage
components without becoming filesystem paths.

### Durable state machine

```text
absent --Begin--> active --Append*--> active --Commit seal--> sealed
  ^                   |                  |                      |
  |                   +--Abort/limit/cancel--> discard journal   |
  |                                                              |
  +--recovery cleanup incomplete <-- receipt durable <-- receiver succeeds
                                                 |
                                                 +--> acknowledgement eligible

restart sealed --verify--> recovery-held --AdmitRecoveredTransaction--> sealed
```

- `active` and all append-temporary files are never visible to a receiver.
- A chunk becomes part of an active stage only after the chunk file, its
  directory, and state update complete their durable transition; a failed write
  leaves no partial final chunk.
- `sealed` is restartable but not acknowledgement-eligible. A crash after
  sealing and before receipt persistence is an at-least-once replay case, not
  a success claim; after restart it is held until explicit admission.
- A receipt becomes valid only after file write, file sync, atomic rename, and
  required parent-directory sync have succeeded. Only then may the stage make
  a `synccontract.DownstreamAcknowledgement` available to the caller.
- Aborting, quota breaches, cancellation, or any receipt durability failure
  produce no receipt-derived eligibility. Source checkpoint/LSN interaction is
  deliberately outside this package.
- `MaxStagedTransactions` is a required finite control-slot budget independent
  of `MaxStagedBytes`: a slot is reserved before Begin creates durable state,
  carried through discard temporary/final control states, and released only
  after durable control retirement.
- Terminal discard evidence is per-stage-instance and outside the transaction
  directory. Mutation and cleanup outcomes are classified as not-applied,
  durable, or indeterminate. A renamed marker remains intact after a parent
  sync failure; it is retired only after matching generation cleanup has
  completed `RemoveAll` plus transactions-directory sync, followed by marker
  removal plus discards-directory sync.
- Recovery enumerates and validates controls before admission, removes only
  owned regular temporary controls, and fails the root closed while preserving
  corrupt, lookalike, symlink, directory, or unrelated artifacts. A cleanup
  error blocks Begin, Append, Commit, and recovery admission until the
  cleanup-only reconciliation succeeds.

### Receipt and existing sync contract

The receiver supplies a stable downstream receipt identifier, sink identity,
and durable timestamp after it has accepted the whole sealed transaction. The
stage persists an immutable receipt containing transaction identity, ordered
content digest, byte/record totals, and downstream receipt metadata, but never
raw payload. Returned/loaded receipts clone mutable fields and expose a
receipt-derived `synccontract.DownstreamAcknowledgement` only if their internal
durability marker came from validated persisted state.

The future source caller invokes `synccontract.CommitAfterDownstreamAcknowledgement`
with that acknowledgement. This issue never constructs a checkpoint, alters a
tombstone/history policy, or sends source feedback.

<threat_model>
## Threat model

| Boundary | Threat | Mitigation and proof |
|---|---|---|
| Provider transaction ID → filesystem | Traversal, collision, or control characters redirect storage | Validate non-empty opaque identity; derive stage component from a stable digest; tests use traversal/control-character IDs and prove files stay below the configured root. |
| Provider chunk → local storage | Unbounded payload exhausts memory/disk or is exposed before commit | Fixed-size stream buffer, exact byte/record/time limits, no public reader until sealed, and quota tests that assert zero publication/eligibility. |
| Receiver success → source acknowledgement | Caller treats a local stage or ordinary success as durable source progress | Only validated/persisted receipt creates acknowledgement; fake source test proves it cannot acknowledge before receipt durability. |
| Process crash / filesystem failure | Partial files, stale state, discard intent, or receipt claim becomes misleading | Temp+sync+rename+parent-sync transitions; bounded external discard controls; recovery validates controls, retires only after durable generation cleanup, and holds bare sealed items for explicit admission. |
| Concurrent calls | Data race, out-of-order append, or cross-transaction interference | Per-stage synchronization with no lock held across receiver I/O; race suite and concurrent-order/isolation tests. |
</threat_model>

## TDD execution plan

### Slice 1 — observable commit/receipt boundary (RED → GREEN)

<read_first>
- `.planning/phases/issue-3975-committed-transaction-stage/CONTEXT.md`
- `internal/synccontract/commit.go`
- `internal/connectors/database/resources.go`
- `internal/connectors/database/registry.go`
- `data/cli-cdc-large-transaction-strategy-r1/report.md`
</read_first>

<action>
Write `TestCommittedTransactionStagePublishesOnlyAfterDurableReceipt` in
`transaction_stage_test.go` before behavioral production code. The test stages
two ordered chunks, asserts the receiver has observed zero chunks before
`CommitTransaction`, asserts no acknowledgement object is obtainable before
the receiver supplies a durable receipt, then asserts exactly one ordered
whole-transaction publication and acknowledgement eligibility only after the
receipt artifact is durable. Use `t.TempDir()` and a fake receiver that records
the event order. Introduce only the minimal concrete types/constructor needed
to compile the test, run the test to demonstrate its behavioural failure, then
implement the smallest private stage/receipt path that passes it.
</action>

<acceptance_criteria>
- The first named focused RED command contains
  `TestCommittedTransactionStagePublishesOnlyAfterDurableReceipt` and fails
  because the required receipt-gated behavior is absent, not because of a
  skipped test or malformed test setup.
- Before `CommitTransaction`, fake receiver calls are `0`, no receipt file
  exists, and no acknowledgement is eligible.
- After a successful durable receipt, receiver sees chunk sequence `0,1`, one
  transaction identity/digest, exactly one receipt artifact, and a valid
  `synccontract.DownstreamAcknowledgement` derived from the returned receipt.
- The test does not import a PostgreSQL package or call a database.
</acceptance_criteria>

### Slice 2 — abort, quotas, bounded stream, and cancellation (RED → GREEN)

<read_first>
- `internal/connectors/database/transaction_stage.go` (after Slice 1)
- `internal/connectors/database/transaction_stage_test.go` (after Slice 1)
- `internal/connectors/database/resources.go`
- `internal/synccontract/state.go`
- `data/cli-cdc-large-transaction-strategy-r1/report.md`
</read_first>

<action>
Add table-driven behavior tests for abort, duplicate/missing boundaries,
byte/record/time limit breaches, cancellation before/during append and commit,
bounded reader-buffer use, chunk order, and transaction isolation. Define the
typed named `TransactionStageLimitExceeded` error with the breached resource
and configured/observed values, usable through `errors.As`/`errors.Is` as
appropriate. Implement validation and streaming copy so the stage never keeps
the total transaction payload in memory and never emits or acknowledges a
limited/cancelled transaction.
</action>

<acceptance_criteria>
- Byte, record, and elapsed-time cases return the named typed limit outcome,
  call the receiver zero times, create no receipt, expose no acknowledgement,
  and remove incomplete active storage.
- Abort removes every stage file for that transaction and a subsequent begin
  with the same opaque ID starts cleanly.
- A reader larger than the fixed stage buffer demonstrates bounded reads and
  does not result in a full-payload in-memory copy.
- `context.Canceled` / deadline failure leaves no final partial chunk, receipt,
  or eligibility and preserves unrelated transaction state.
</acceptance_criteria>

### Slice 3 — recovery and crash-consistency failures (RED → GREEN → REFACTOR)

<read_first>
- `internal/connectors/database/transaction_stage.go` (after Slices 1–2)
- `internal/connectors/database/transaction_stage_test.go` (after Slices 1–2)
- `internal/synccontract/commit.go`
- `internal/synccontract/recovery.go`
- `internal/app/local_warehouse.go` and its durability tests for repository
  fsync/atomic-replacement conventions
- `data/cli-cdc-bidirectional-changefeed-design-r1/report.md`
- `/Users/karthiksivadas/karthik-agent-workspace/data/learnings.md`
</read_first>

<action>
Add a test-only fault-injecting filesystem/clock/receipt receiver and cover
every persistence transition: begin directory creation, append write/file sync/
rename/parent sync, seal write/sync/rename/parent sync, receiver failure,
receipt write/file sync/rename/parent sync, and post-receipt cleanup. Add
restart tests that reopen the same root, discard incomplete/orphan temporary
stages, report recovery-held sealed receipt-less work without eligibility,
require explicit admission before its resume, recover a valid receipt, and
clean post-receipt sealed residue. Refactor the production
path only after all tests are green, preserving small interfaces, defensive
copies, checked errors, and no lock across receiver I/O.
</action>

<acceptance_criteria>
- Disk-full, file-sync, rename, parent-sync, receiver, and receipt failures
  leave source acknowledgement ineligible and never produce a deceptive final
  receipt.
- Restart at each named transition yields either no transaction residue for
  incomplete work, a recovery-held sealed no-receipt transaction, or a
  validated durable receipt; no `.tmp`/partial final artifacts remain.
- Recovery accounting includes retained sealed bytes so restart cannot bypass
  limits.
- `go test -race` reports no race and concurrent independent transactions do
  not share chunks, counters, or receipts.
</acceptance_criteria>

### Refactor checkpoint

After all slices pass, remove only duplication revealed by the tests. Keep
filesystem state names/private details unexported, errors wrapped with `%w`,
payload bytes defensively isolated, and all durable-operation errors observed.
Run focused tests after refactor; create a refactor commit only if source
changes are needed.

## Commit and push checkpoints

1. `docs(3975): capture transaction-stage context` — completed prior to code.
2. `docs(3975): plan committed transaction stage` — plan/TDD/verification
   checkpoint; pushed before implementation began.
3. `test(3975): add receipt-gated transaction-stage red test` — first named
   RED evidence captured before behavioral implementation; push once the red
   evidence and planning files are coherent.
4. `feat(3975): stage committed transactions with durable receipts` — GREEN
   implementation plus focused/race evidence. Per the user-directed audit
   hold after `2afa128e`, keep this and all later local commits unpushed until
   verify/review, Shepherd-compatible evidence, and no-mistakes are complete.
5. `refactor(3975): harden transaction-stage recovery` — only if cleanup is
   warranted after green, followed by review/no-mistakes fixes as needed.

Every commit includes body trailers `Refs #3975`, `Refs #3972`, `Refs #2986`,
and `Refs #2988`. Never commit or push to `main`.

## Manual plan-checker self-review

The single-worker contract prohibits calling the installed planner/checker role.
Before execution, manually verify this plan covers every CONTEXT decision:

- D-01 through D-03: receipt boundary and acknowledgement eligibility — Slice 1.
- D-04 through D-06: private state, recovery, receipt immutability — Slices 1 and 3.
- D-07 through D-09: finite quotas, failure injection, path/accounting safety — Slices 2 and 3.
- D-10 through D-12: named first RED, artifact proofs, required skills — all slices and final gates.

The plan remains one independent child slice, stays entirely within
`internal/connectors/database/**`, preserves #3974's closed descriptor truth,
and contains no external dependency/human-gate action. Manual plan-check result:
**passed — ready for execute-phase inline TDD.**
