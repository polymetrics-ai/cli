# Verification - #4043 post-budget control-slot reconciliation

**Status:** Implementation committed; verification pending
**Safety boundary:** local t.TempDir, injected filesystem operations, and fake
receivers only. No PostgreSQL service, Podman, provider, credentials, network,
warehouse mutation, target DML, or generic SQL.

## Focused behavior

1. Focused five-test Red and Green command from TDD-LEDGER.md.
2. Repeat:

   go test -timeout 20m -count=20 ./internal/connectors/database -run '^(TestCommittedTransactionStageDiscardControlTempCleanupRetainsSlotUntilReconciled|TestCommittedTransactionStageDiscardControlTempRemovalRequiresDirectorySync|TestCommittedTransactionStageRecoveredOverCapacityCannotClearPoisonOrDeliver|TestCommittedTransactionStageRecoveredControlReservationRestartMatrix|TestCommittedTransactionStageDiscardControlTemporaryCrashMatrix)$'

3. Race:

   go test -race -timeout 20m -count=10 ./internal/connectors/database -run '^(TestCommittedTransactionStageDiscardControlTempCleanupRetainsSlotUntilReconciled|TestCommittedTransactionStageDiscardControlTempRemovalRequiresDirectorySync|TestCommittedTransactionStageRecoveredOverCapacityCannotClearPoisonOrDeliver|TestCommittedTransactionStageRecoveredControlReservationRestartMatrix|TestCommittedTransactionStageDiscardControlTemporaryCrashMatrix)$'

4. Existing valid paths:

   go test -timeout 20m -count=1 -v ./internal/connectors/database -run '^(TestCommittedTransactionStageBoundsControlSlotsBeforeDurableBegin|TestCommittedTransactionStageDiscardControlRetentionIsBounded|TestCommittedTransactionStageRecoveryReapsOnlyOwnedDiscardTemps|TestCommittedTransactionStageDiscardFinalRetirementFailuresPoisonRoot|TestCommittedTransactionStageDiscardControlCleanupFailurePoisonsRoot|TestCommittedTransactionStageRestartRecoveryRetainsOnlySealedReceiptlessWork|TestCommittedTransactionStagePublishesOnlyAfterDurableReceipt|TestCommittedTransactionStagePostReceiptCleanupFailurePreservesOnlyDurableReceipt|TestCommittedTransactionStageDiscardIntentWriteFailureWithDurableCleanupNeverRecovers|TestCommittedTransactionStageDiscardIntentParentSyncFailureRetainsExternalMarker|TestCommittedTransactionStageDiscardCleanupParentSyncFailureNeverAdmitsRecovery)$'

## Package and repository gates in order

1. gofmt -w internal/connectors/database/transaction_stage.go internal/connectors/database/transaction_stage_fault_test.go internal/connectors/database/transaction_stage_test.go
2. go test -timeout 20m -count=1 ./internal/connectors/database
3. go test -race -timeout 20m -count=1 ./internal/connectors/database
4. go test -timeout 20m -count=1 ./internal/synccontract ./internal/app
5. go vet ./internal/connectors/database ./internal/synccontract ./internal/app
6. go build ./cmd/pm
7. git diff --check
8. go run ./cmd/agentcontractgen check
9. make tidy-check
10. make lint
11. make docs-check
12. make smoke-no-build
13. make agent-contract-check
14. make connectorgen-validate
15. make connectorgen-surface-sync
16. make connector-boundary
17. make release-workflow-check

The full repository Go suite remains CI-owned. Do not run go test ./... or
make verify as a single short-timeout agent command.

## Lifecycle and review

- scripts/gsd prompt execute-phase, verify-work, and code-review are resolved
  and executed inline under the named-phase manual fallback.
- Any GSD verification gap uses the named phase plan-gap/execute-gaps route
  documented in the ship brief.
- Every code-review finding receives an explicit disposition.
- CLI help/manual/website parity: Not Applicable; this repair changes no
  command, flag, output, connector surface, help topic, manual, or website
  contract.

## Results

Implementation and its focused regressions are committed. This document does
not assert focused, package, repository, or remote gate results until their
commands are recorded.
