# #4046 R9 TDD ledger

**Base:** `e5a55c68003e63860e8213b509a18643d5f4e3d6`
**Old exhausted run (immutable):** `01KZQ0C1KEZRHNXX4WJFWXSCFB`
**Fresh delivery budget:** 0/5 before no-mistakes begins

| ID | Requirement | Red | Green | Status |
|---|---|---|---|---|
| R9-T1 | A stale CAS loser is observably returned as zero/non-terminal and durably reopened as `running` at the pinned base while winner and unrelated state are preserved. | **Red:** observed with `go test -json -count=1 -timeout 20m ./internal/app -run '^TestRunETLTransportStaleWriterFinalizesLosingRun$'` (exit 1): zero returned loser, durable `running` loser, and zero completion timestamp; typed conflict and winner/unrelated assertions passed first. | **Green:** observed with `go test -count=1 -timeout 20m ./internal/app -run '^TestRunETLTransportStaleWriterFinalizesLosingRun$' -v` (exit 0): non-zero failed returned run matched durable state and retained the typed conflict. | Green |
| R9-T2 | Only a typed `errTransportStreamStateConflict` rebases terminalization to current locked state; ordinary failures retain the revision guard. | **Red:** covered by R9-T1's stale revision failure path. | **Green:** `TestFailRunTransportConflictPreservesLatestConcurrentState`, `TestFailRunRetainsRevisionGuardWithoutTransportConflict`, and `TestFailRunTransportConflictRequiresRunningTarget` pass: typed conflict finalizes only its matching running run, ordinary failure retains `errStateRevisionConflict`, and a terminal target is not replaced. | Green |
| R9-T3 | An unrelated project write between conflict observation and loser terminalization survives unchanged. | **Red:** covered by the pinned R9-T1 state leak. | **Green:** `go test -count=20 -timeout 20m ./internal/app -run '^TestFailRunTransportConflictPreservesLatestConcurrentState$'` exit `0`; the stale app first receives the real typed CAS conflict, a second writer then persists unrelated stream/checkpoint/run data, and typed finalization retains it with the winner checkpoint unchanged. | Green |
| R9-T4 | Restart does not reload a false-running loser. | **Red:** R9-T1's reopen assertion captures the original leak. | **Green:** `go test -count=10 -timeout 20m ./internal/app -run '^TestRunETLTransportStaleWriterFailureSurvivesReopen$'` exit `0`. | Green |
| R9-T5 | Cancellation after acknowledgement does not erase ordering, typed conflict, or terminalization. | **Red:** R9-T1 establishes the common path is broken at the base. | **Green:** `go test -count=10 -timeout 20m ./internal/app -run '^TestRunETLTransportStaleWriterFinalizesAfterCancellation$'` exit `0`; cancellation occurs after acknowledgement and the typed conflict remains detectable. | Green |
| R9-T6 | The shared path terminalizes stale losers for every canonical mode. | **Red:** R9-T1 establishes the common path is broken at the base. | **Green:** `go test -count=1 -timeout 20m ./internal/app -run '^TestRunETLTransportStaleWriterFinalizesLosingRunForAllModes$' -v` exit `0` for `full_overwrite`, `full_append`, `incremental_append`, `incremental_upsert`, `incremental_dedupe`, `incremental_dedupe_history`, and `change_capture`. | Green |
| R9-T7 | A definite pre-rename finalization failure never reports a speculative terminal loser, while committed and indeterminate outcomes do. | **Red:** review traced `JSONStore.Update` returning callback-mutated `next` when `saveNoLock` fails before rename; the prior typed-conflict branch returned that speculative run. | **Green:** focused race-enabled app tests prove a real stale CAS returns zero with the durable loser still `running`, and committed/indeterminate finalizations return the terminal loser. | Green |
| R9-R7/R8 | Resume identity and target-entry CAS continue to reject an invalid/stale checkpoint without overwriting winner state. | **Red:** prior #3864 T20/T21 evidence is linked, not reused as this phase's RED. | **Green:** the exact nine-test R7/R8 suite in `VERIFICATION.md` passed unchanged with exit `0`. | Green |

## TDD gate commitments

- A test-only commit must precede the production `app.go` commit.
- The first RED must observe a persisted `running` loser and/or zero returned loser after real `RunETL -> failRun`; an error-code-only test is rejected.
- No test receives a live credential, provider, network, warehouse, container, or external service.
- Every result appended below records its exact command, exit status, and whether it ran before or after the production change.

## Execution record

### R9-T1 Red: durable stale-writer leak (before production change)

- Command: `go test -json -count=1 -timeout 20m ./internal/app -run '^TestRunETLTransportStaleWriterFinalizesLosingRun$'`
- Exit: `1` (expected RED)
- Observed failure: `RunETL returned zero losing Run`; returned status was empty; reopened durable loser was `running` with a zero completion timestamp; the durable ID was `run_94e7f2d8862e3416` while the returned ID was empty.
- The test had already proved `errors.Is(losingErr, errTransportStreamStateConflict)`, retained the winner checkpoint/run identity, and retained the unrelated stream/checkpoint before it emitted the aggregate durable-symptom failure.
- This is test and evidence only; `internal/app/app.go` remains unchanged at this checkpoint.

### R9-T1 Green: typed-conflict-only terminalization

- Command: `go test -count=1 -timeout 20m ./internal/app -run '^TestRunETLTransportStaleWriterFinalizesLosingRun$' -v`
- Exit: `0`
- Result: the matching returned/durable loser run is non-zero and `failed`; its completion timestamp is non-zero; `errors.Is(losingErr, errTransportStreamStateConflict)` remains true; winner stream state and unrelated state remain unchanged after reopen.
- Production change: `failRun` bypasses the stale whole-state revision guard only for the typed transport stream-state conflict, requires the matching current run to still be `running`, and otherwise retains the existing revision-conflict behavior.

### R9-T2 through R9-T6 Green: focused expansion (after production change)

- `go test -count=20 -timeout 20m ./internal/app -run '^TestFailRunTransportConflictPreservesLatestConcurrentState$'` — exit `0`; the unrelated write is explicitly sequenced after the stale app observes its actual typed CAS conflict and before `failRun` terminalizes the loser.
- `go test -count=10 -timeout 20m ./internal/app -run '^TestRunETLTransportStaleWriterFailureSurvivesReopen$'` — exit `0`.
- `go test -count=10 -timeout 20m ./internal/app -run '^TestRunETLTransportStaleWriterFinalizesAfterCancellation$'` — exit `0`.
- `go test -count=1 -timeout 20m ./internal/app -run '^TestRunETLTransportStaleWriterFinalizesLosingRunForAllModes$' -v` — exit `0` for all seven canonical modes.
- Adjacent focused tests prove ordinary errors still return the revision conflict and a conflict cannot replace a completed target run.

### R9-R7/R8 regression (after production change)

- Command: `go test -count=1 -timeout 20m ./internal/app -run '^(TestRunETLTransportRejectsAcknowledgedCheckpointWithIncompatibleResume|TestRunETLTransportPersistsActiveCheckpointBeforeSourceFailureForAllModes|TestRunETLTransportAdvancesInterimCheckpointAcrossPages|TestRunETLTransportPreservesUnrelatedStateDuringInterimCheckpointCommit|TestRunETLTransportRejectsStaleCheckpointWriter|TestRunETLTransportDistinguishesMissingAndPresentStreamState|TestRunETLTransportCommitsAcknowledgedPageBeforeCancellation|TestRunETLTransportRetainsInterimCheckpointWhenFinalStateSaveFails|TestRunETLTransportTreatsIndeterminateCheckpointPersistenceAsFailure)$'`
- Exit: `0`
- Result: the established identity, source-generation, acknowledgement ordering, target-entry CAS, unrelated-state, cancellation, and state-store-outcome protections remain unchanged.

### R9-T7 Green: finalization commit-outcome truth

- Command: `go test -race -count=1 -timeout 20m ./internal/app -run '^(TestRunETLTransportStaleWriterDoesNotReportUncommittedFinalization|TestFailRunTransportConflictRequiresRunningTarget|TestFailRunTransportConflictReturnsMayHaveCommittedFinalization)$'`
- Exit: `0`
- Result: a real two-app stale CAS followed by a pre-rename temporary-file creation failure returns `Run{}` and reopens the loser as `running`; the winner and unrelated stream/checkpoint/run remain unchanged, while committed and indeterminate finalization outcomes return a terminal loser and preserve the typed conflict chain.
