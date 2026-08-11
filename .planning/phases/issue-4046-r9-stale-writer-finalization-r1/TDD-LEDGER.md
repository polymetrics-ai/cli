# #4046 R9 TDD ledger

**Base:** `e5a55c68003e63860e8213b509a18643d5f4e3d6`
**Old exhausted run (immutable):** `01KZQ0C1KEZRHNXX4WJFWXSCFB`
**Fresh delivery budget:** 0/5 before no-mistakes begins

| ID | Requirement | Red | Green | Status |
|---|---|---|---|---|
| R9-T1 | A stale CAS loser is observably returned as zero/non-terminal and durably reopened as `running` at the pinned base while winner and unrelated state are preserved. | **Red:** observed with `go test -json -count=1 -timeout 20m ./internal/app -run '^TestRunETLTransportStaleWriterFinalizesLosingRun$'` (exit 1): zero returned loser, durable `running` loser, and zero completion timestamp; typed conflict and winner/unrelated assertions passed first. | **Green:** same command must return a non-zero `failed` run matching durable state and retain `errors.Is(err, errTransportStreamStateConflict)`. | Red observed |
| R9-T2 | Only a typed `errTransportStreamStateConflict` rebases terminalization to current locked state; ordinary failures retain the revision guard. | **Red:** covered by R9-T1's stale revision failure path. | **Green:** focused test asserts the matching running run is changed, and non-conflict behavior remains guarded. | Planned |
| R9-T3 | An unrelated project write between conflict observation and loser terminalization survives unchanged. | **Red:** planned `go test -count=20 -timeout 20m ./internal/app -run '^TestFailRunTransportConflictPreservesLatestConcurrentState$'`. | **Green:** same command passes repeatedly, including unrelated stream/checkpoint/run values. | Planned |
| R9-T4 | Restart does not reload a false-running loser. | **Red:** R9-T1's reopen assertion captures the original leak. | **Green:** `go test -count=10 -timeout 20m ./internal/app -run '^TestRunETLTransportStaleWriterFailureSurvivesReopen$'` passes. | Planned |
| R9-T5 | Cancellation after acknowledgement does not erase ordering, typed conflict, or terminalization. | **Red:** planned `go test -count=10 -timeout 20m ./internal/app -run '^TestRunETLTransportStaleWriterFinalizesAfterCancellation$'`. | **Green:** same command passes with the persisted loser terminal and the conflict detectable. | Planned |
| R9-T6 | The shared path terminalizes stale losers for every canonical mode. | **Red:** R9-T1 establishes the common path is broken at the base. | **Green:** `go test -count=1 -timeout 20m ./internal/app -run '^TestRunETLTransportStaleWriterFinalizesLosingRunForAllModes$' -v` passes for all seven named modes. | Planned |
| R9-R7/R8 | Resume identity and target-entry CAS continue to reject an invalid/stale checkpoint without overwriting winner state. | **Red:** prior #3864 T20/T21 evidence is linked, not reused as this phase's RED. | **Green:** the exact nine-test R7/R8 suite in `VERIFICATION.md` remains green unchanged. | Planned |

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
