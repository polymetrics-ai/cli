# TDD ledger — Issue #3975: committed-transaction staging and durable receipts

Manual inline GSD TDD execution. Tests must assert durable artifacts and state
facts, never just an exit code. The first executable RED output is retained
under `traces/red-committed-transaction-stage.txt` before GREEN behavior is
committed.

| ID | Guarantee | RED assertion | GREEN proof |
|---|---|---|---|
| T1 | Private chunks | Appended chunks are visible to a receiver before commit. | Receiver observes no chunk before commit; active files are private. |
| T2 | Whole transaction | Commit can publish a subset, reorder chunks, or split receipts. | Receiver gets exactly one transaction in source append order and one receipt. |
| T3 | Durable receipt first | A stage/receiver success allows acknowledgement before a persisted receipt. | Receipt file + file sync + atomic rename + parent sync finish before eligibility is returned. |
| T4 | Abort | Abort can publish, retain a receipt, or leave active payload. | Abort calls receiver zero times and removes stage/temporary files. |
| T5 | Byte quota | A byte-over-limit stream continues, publishes, or becomes eligible. | `TransactionStageLimitExceeded` names bytes and leaves no output/eligibility. |
| T6 | Record quota | Record count over limit reaches receiver. | Named limit failure occurs before publication and active state is removed. |
| T7 | Time quota | An old transaction can append/commit after its limit. | Injected clock produces named time failure, cleanup, and no eligibility. |
| T8 | Bounded memory | One large reader is accumulated in a full transaction buffer. | Instrumented reader proves fixed bounded read buffer and staged files carry payload. |
| T9 | Cancellation | Cancellation after partial work creates a final chunk/receipt. | `context.Canceled` leaves only safe unrelated state; no final partial/temp residue. |
| T10 | Disk full | ENOSPC is masked or yields a deceptive final file/receipt. | Error is wrapped, active temp is removed, receiver/ack stay zero. |
| T11 | Durable I/O | File sync, rename, or parent directory sync failure becomes success. | Each injected failure blocks final transition and eligible acknowledgement. |
| T12 | Receiver/receipt failure | Receiver/receipt failure permits eligibility or cleanup as success. | Sealed work stays receipt-less/recoverable; receipt failure leaves no final receipt. |
| T13 | Restart recovery | Recovery publishes active data or treats sealed/no-receipt data as acknowledged. | Recovery removes incomplete/orphan state, retains resume work, and validates durable receipts. |
| T14 | Crash boundaries | A crash at any transition leaves misleading residue. | Transition matrix proves only active-cleaned, sealed-resumable, or receipt-valid states; zero temp residue. |
| T15 | Opaque identity/path safety | Transaction ID becomes a path or aliases another ID. | Traversal/control IDs stay inside root and distinct opaque IDs receive distinct stage components. |
| T16 | Concurrent isolation | Concurrent transactions race or leak chunks/receipts across IDs. | `-race` and isolation/order tests pass with no shared mutable payload. |
| T17 | Synccontract reuse | Stage introduces a forged acknowledgement/checkpoint envelope. | Returned eligibility adapts to `synccontract.NewDurableDownstreamAcknowledgement`; no duplicate checkpoint/tombstone/history type is added. |

## First RED command

```sh
go test -timeout 20m -count=1 ./internal/connectors/database \
  -run '^TestCommittedTransactionStagePublishesOnlyAfterDurableReceipt$'
```

**Required RED condition:** the test must demonstrate absent receipt-gated
transaction behavior before its GREEN implementation. A compile-only failure is
not accepted as final RED evidence; introduce the minimal declarations needed
to execute the test, then preserve the behavioral failing output.

## Green commands

```sh
go test -timeout 20m -count=1 ./internal/connectors/database
go test -timeout 20m -race -count=1 ./internal/connectors/database
go test -timeout 20m -count=1 ./internal/synccontract ./internal/app
```

## Refactor condition

Refactor only after all T1–T17 focused tests pass. Re-run every Green command
after any cleanup. Do not convert a failure into a skipped test, weaker
assertion, unbounded limit, or generic success result.
