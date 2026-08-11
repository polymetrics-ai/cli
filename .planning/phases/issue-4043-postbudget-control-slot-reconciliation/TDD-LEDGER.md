# TDD Ledger - #4043 post-budget control-slot reconciliation

**Starting commit:** 6a82f3650ab4be0b511541f91721ce7cefe08762
**Branch:** fix/4043-postbudget-control-slot-reconciliation
**Status:** Plan checkpoint committed before production edits

## Required skills

golang-how-to; golang-design-patterns; golang-structs-interfaces;
golang-error-handling; golang-security; golang-safety; golang-testing;
golang-context; golang-concurrency; golang-database; golang-troubleshooting;
golang-lint; github-issue-first-delivery; no-mistakes.

## Red tests

| Test | Behavioral Red to capture | Status |
|---|---|---|
| TestCommittedTransactionStageDiscardControlTempCleanupRetainsSlotUntilReconciled | cap one, write plus temp-remove fault: typed cleanup-required, one retained slot, root poison, Begin blocked, no more than one owned temp. | Pending |
| TestCommittedTransactionStageDiscardControlTempRemovalRequiresDirectorySync | pre-rename write fault plus successful unlink plus discards sync fault: sync call is observed and release/admission remain blocked. | Pending |
| TestCommittedTransactionStageRecoveredOverCapacityCannotClearPoisonOrDeliver | cap two creation then cap one recovery: repeated reconcile remains cleanup-required; unreserved work has receiver zero, absent receipt, unavailable ack. | Pending |
| TestCommittedTransactionStageRecoveredControlReservationRestartMatrix | 1 to 1, 2 to 2, 2 to 1, 2 to 1 to 2, receipt residue, ordering, repeated reopen/reconcile exact mappings. | Pending |
| TestCommittedTransactionStageDiscardControlTemporaryCrashMatrix | create/write/file-sync/close/rename/remove/temp-sync/stage-remove/stage-sync/final-remove/final-sync durable and indeterminate images. | Pending |

## Focused command

go test -timeout 20m -count=1 -v ./internal/connectors/database -run '^(TestCommittedTransactionStageDiscardControlTempCleanupRetainsSlotUntilReconciled|TestCommittedTransactionStageDiscardControlTempRemovalRequiresDirectorySync|TestCommittedTransactionStageRecoveredOverCapacityCannotClearPoisonOrDeliver|TestCommittedTransactionStageRecoveredControlReservationRestartMatrix|TestCommittedTransactionStageDiscardControlTemporaryCrashMatrix)$'

## Expected genuine Red evidence

- The old code permits three owned discard temporaries with
  MaxStagedTransactions equal to one, no cleanup poison, and a subsequent Begin.
- The old code performs zero recorded discards-directory sync attempts after a
  successful pre-rename temporary removal.
- The old code lets ReconcileDiscardControls return nil with an unreserved
  retained entry, then admits it, calls the receiver once, persists a receipt,
  and exposes acknowledgement eligibility.

Compile failure, skipped tests, missing symbols, or an exit-status-only
assertion is invalid Red evidence.

## Green gates

| Gate | Required assertion | Status |
|---|---|---|
| Slice A | One unresolved Temporary retains one slot and blocks delivery; durable removal plus discards sync permits release. | Pending |
| Slice B | Every retained entry has an exact Reserved control before poison clears; unreserved work cannot reach receiver. | Pending |
| Repeat | Focused five tests pass with count 20. | Pending |
| Race | Focused five tests pass with race and count 10. | Pending |
| Restart | Cap and temporary crash matrices assert exact artifact and entry/control mapping after reopen. | Pending |
| Valid paths | Final retirement, owned temp recovery, within-capacity recovery, receipt-before-ack stay green. | Pending |

## Evidence log

### Red

Red: Pending. The five named tests will be added and the focused command will
be run before any production edit. This section will then contain literal
durable artifact and receiver-boundary failure output.

### Green

Green: Pending. After both repair slices, this section will contain literal
focused, repeat, race, restart, valid-path, and package evidence.
