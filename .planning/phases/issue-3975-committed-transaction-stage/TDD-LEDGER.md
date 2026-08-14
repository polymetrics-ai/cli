# TDD ledger — Issue #3975: committed-transaction staging and durable receipts

Manual inline GSD TDD execution. Tests must assert durable artifacts and state
facts, never just an exit code. The first executable RED output is retained
under `traces/red-committed-transaction-stage.txt` before GREEN behavior is
committed. It ran on this child branch and failed behaviorally at the absent
receipt-gated commit boundary.

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
| T13 | Restart recovery | Recovery publishes active data or treats sealed/no-receipt data as acknowledged. | Recovery removes incomplete/orphan state, holds sealed work for explicit admission, and validates durable receipts. |
| T14 | Crash boundaries | A crash at any transition leaves misleading residue. | Transition matrix proves only active-cleaned, sealed-recovery-held, or receipt-valid states; zero temp residue. |
| T15 | Opaque identity/path safety | Transaction ID becomes a path or aliases another ID. | Traversal/control IDs stay inside root and distinct opaque IDs receive distinct stage components. |
| T16 | Concurrent isolation | Concurrent transactions race or leak chunks/receipts across IDs. | `-race` and isolation/order tests pass with no shared mutable payload. |
| T17 | Synccontract reuse | Stage introduces a forged acknowledgement/checkpoint envelope. | Returned eligibility adapts to `synccontract.NewDurableDownstreamAcknowledgement`; no duplicate checkpoint/tombstone/history type is added. |
| T18 | Terminal discard recovery | A failed cleanup followed by restart revives deliberately discarded sealed work. | A faulted cancelled abort persists discard intent; recovery removes it, cannot commit it, and never calls a receiver. |
| T19 | Indeterminate discard recovery | A discard-control write failure plus cleanup failure replays a sealed transaction after restart. | Recovery holds every bare sealed transaction until explicit admission, so direct commit makes zero receiver, receipt, or acknowledgement calls. |
| T20 | Renamed discard marker | A parent-sync failure removes a newly renamed terminal marker. | The external marker remains observable across restart and recovery removes the matching stage without delivery. |
| T21 | Indeterminate cleanup | A cleanup parent-sync failure makes discarded work directly recoverable. | The marker remains terminal; restart has no committable stage or receiver call. |
| T22 | Failed discard intent with durable cleanup | A failed marker write leaves a removed transaction replayable after restart. | Durable stage-directory cleanup leaves no pending transaction or receiver call. |
| T23 | Bounded discard control | Repeated empty begin/abort leaves final controls outside the payload quota. | One required control slot is reserved per stage instance and released only after final retirement is directory-synced. |
| T24 | Owned temporary recovery | A crash leaves an owned discard-control temp forever, or cleanup deletes an operator artifact. | Recovery removes only exact regular legacy/new control temps, syncs the directory, and preserves unexpected artifacts fail-closed. |
| T25 | Final retirement ordering | A final is removed before matching transaction cleanup is durable, or survives a safe restart forever. | A valid final dominates matching recovery, then retires only after transaction cleanup and both directory sync boundaries. |
| T26 | Cleanup root poison | Failed final/control cleanup permits new begin, append, commit, or recovery admission. | Typed cleanup errors block all delivery-capable operations until cleanup-only reconciliation succeeds. |

## First RED command

```sh
go test -timeout 20m -count=1 ./internal/connectors/database \
  -run '^TestCommittedTransactionStagePublishesOnlyAfterDurableReceipt$'
```

**Required RED condition:** the test must demonstrate absent receipt-gated
transaction behavior before its GREEN implementation. A compile-only failure is
not accepted as final RED evidence; introduce the minimal declarations needed
to execute the test, then preserve the behavioral failing output.

**Recorded RED:** PASS — the exact command failed at
`CommitTransaction() error = committed transaction staging is not implemented`.

## Green commands

```sh
go test -timeout 20m -count=1 ./internal/connectors/database
go test -timeout 20m -race -count=1 ./internal/connectors/database
go test -timeout 20m -count=1 ./internal/synccontract ./internal/app
```

## Executed Green evidence

- [x] The named first test passes after durable implementation and reopens the
      root to prove the receipt artifact, not an in-memory value, survives.
- [x] The package suite covers T1–T26, including abort/quota/cancellation,
      streamed-buffer, recovery, opaque-key, receipt-forgery, and concurrent
      isolation cases.
- [x] `transaction_stage_fault_test.go` injects begin, chunk, manifest, seal,
      receiver, receipt, and post-receipt-cleanup faults; every failed crash
      boundary is reopened and shown to have only a safe state.
- [x] Package `-race`, `internal/synccontract`, and `internal/app` checks pass.

The sole implementation-review correction is bounded/saturating quota
arithmetic for extreme untrusted record counts. It is a source hardening
correction inside #3975, not a newly discovered repository-gate defect, and is
covered by `TestCommittedTransactionStageSaturatesUntrustedRecordQuotaDiagnostics`.

## Correction #4043 — discarded sealed recovery

- Red: `traces/discarded-sealed-recovery-red.txt` records the focused test
  failure where a restart exposed a failed-cleanup sealed transaction as
  pending.
- Green: `traces/discarded-sealed-recovery-green.txt` records the focused pass
  after durable discard intent makes recovery remove the transaction before it
  can resume or deliver.

## Correction #4043 — indeterminate discard intent

- Red: `traces/discard-intent-write-and-cleanup-failure-red.txt` records the
  deterministic dual-failure replay where restart directly delivers the
  sealed transaction.
- Green: `traces/discard-intent-write-and-cleanup-failure-green.txt` records
  the focused package pass after the external marker and recovery-hold
  transition prevent direct replay.

### Discard durability matrix

| Discard mutation outcome | Cleanup outcome | Crash/restart treatment | Proof |
|---|---|---|---|
| not-applied | durable | The transaction directory is absent, so there is nothing to replay. | T22 write fault with successful cleanup. |
| not-applied | indeterminate | No on-disk distinction is claimed; a surviving bare sealed stage is recovery-held. | T19 dual write/cleanup fault. |
| durable | indeterminate | The retained external marker causes recovery cleanup before delivery. | T18 failed cleanup and T21 parent-sync cleanup fault. |
| indeterminate | indeterminate | A visible renamed marker is retained and dominates recovery; if a crash loses visibility, the bare sealed fallback is held. | T20 marker parent-sync fault. |

### Bounded discard-control matrix

| Control mutation | Generation cleanup | Final retirement | Restart/retry result | Proof |
|---|---|---|---|---|
| not-applied | durable | not applicable | No generation remains; the reserved control slot is released. | T22 write failure with durable cleanup. |
| not-applied | indeterminate | not applicable | Bare sealed residue is recovery-held; the current root is cleanup-poisoned. | T19 dual write/cleanup fault. |
| durable | durable | durable | Final is removed and `discards/` is synced; the slot is released. | T23 bounded repeated Begin/Abort. |
| durable | indeterminate | not attempted | Final remains and the root is poisoned until cleanup-only reconciliation. | T21 and T26 transaction cleanup failures. |
| indeterminate | durable | not attempted | The renamed final remains; the root is poisoned until reconciliation validates and retires it. | T20 parent-sync failure. |
| durable | durable | not-applied or indeterminate | The root is poisoned; retry scans the validated final before admitting work. | T25 final remove/directory-sync faults. |
| crash before final rename | not applicable | not applicable | Only exact owned control temporaries are reaped; lookalikes and operator artifacts remain fail-closed. | T24 recovery temp tests. |

Every bare recovered sealed item requires `AdmitRecoveredTransaction` before
delivery. A discard marker is per stage instance, so retaining old terminal
evidence cannot discard a later reuse of the same opaque transaction identity.

## Correction #4043 — bounded discard controls

- Red: `traces/discard-control-final-correction-red.txt` records the four
  deterministic behavioural failures in
  `TestCommittedTransactionStageDiscardControlRetentionIsBounded`,
  `TestCommittedTransactionStageRecoveryReapsOnlyOwnedDiscardTemps`,
  `TestCommittedTransactionStageDiscardRetirementCrashMatrix`, and
  `TestCommittedTransactionStageDiscardControlCleanupFailurePoisonsRoot`.
  The pre-fix root retained three final controls under a one-byte payload
  limit, retained a real pre-rename temporary file after reopen, retained a
  final after transaction cleanup, and continued admitting work after failed
  control cleanup.
- Green: `traces/discard-control-final-correction-green.txt` records the
  focused matrix, full database package, database race run, and
  synccontract/app regressions after the bounded one-control-slot admission,
  exact temporary recovery, durable final retirement, typed root poison, and
  cleanup-only reconciliation implementation.

## Refactor condition

Refactor only after all T1–T26 focused tests pass. Re-run every Green command
after any cleanup. Do not convert a failure into a skipped test, weaker
assertion, unbounded limit, or generic success result.
