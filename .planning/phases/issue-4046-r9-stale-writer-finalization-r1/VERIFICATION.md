# #4046 R9 verification checklist

**Status:** Local implementation and ordered matrix passed; manual GSD verify/code-review records are complete; fresh no-mistakes and GitHub delivery remain pending.

## Safety preconditions

- [x] #4046 was amended and read back with the R9 acceptance addendum before local project edits.
- [x] Live #4046, #3864, #3862, #4015, #4059, and #4019 were re-verified through `gh-axi`.
- [x] Branch starts at `e5a55c68003e63860e8213b509a18643d5f4e3d6`.
- [x] `.codegraph/` is absent; CodeGraph is not available.
- [x] `scripts/gsd doctor`, required `sources`, and `go run ./cmd/agentcontractgen check` passed.
- [x] The planning and behavioral RED checkpoints contain no production edit.

## Ordered local matrix

1. [x] Behavioral RED (before `internal/app/app.go` changes):

   ```sh
   go test -count=1 -timeout 20m ./internal/app -run '^TestRunETLTransportStaleWriterFinalizesLosingRun$' -v
   ```

   Observed as `go test -json ...` with exit `1`: the typed conflict remained detectable, winner/unrelated preservation assertions passed, and the test reported a zero returned loser plus a durable reopened `running` loser.

2. [x] Matching narrow GREEN: same command, passing (exit `0`); returned and reopened losing runs match, are terminal `failed`, and retain the typed conflict.

3. [x] Intervening unrelated writer (exit `0`, 20 repetitions; preserves winner and unrelated stream/checkpoint/run):

   ```sh
   go test -count=20 -timeout 20m ./internal/app -run '^TestFailRunTransportConflictPreservesLatestConcurrentState$'
   ```

4. [x] Restart truth (exit `0`, 10 repetitions):

   ```sh
   go test -count=10 -timeout 20m ./internal/app -run '^TestRunETLTransportStaleWriterFailureSurvivesReopen$'
   ```

5. [x] Cancellation (exit `0`, 10 repetitions; cancellation is issued after acknowledgement, and the typed conflict remains detectable):

   ```sh
   go test -count=10 -timeout 20m ./internal/app -run '^TestRunETLTransportStaleWriterFinalizesAfterCancellation$'
   ```

6. [x] All seven modes (exit `0`: `full_overwrite`, `full_append`, `incremental_append`, `incremental_upsert`, `incremental_dedupe`, `incremental_dedupe_history`, `change_capture`):

   ```sh
   go test -count=1 -timeout 20m ./internal/app -run '^TestRunETLTransportStaleWriterFinalizesLosingRunForAllModes$' -v
   ```

7. [x] Race/interleaving (exit `0`, 10 repetitions; no race report):

   ```sh
   go test -race -count=10 -timeout 20m ./internal/app -run '^(TestRunETLTransportStaleWriterFinalizesLosingRun|TestFailRunTransportConflictPreservesLatestConcurrentState|TestRunETLTransportStaleWriterFinalizesAfterCancellation)$'
   ```

8. [x] Existing R7/R8 regression (exit `0`):

   ```sh
   go test -count=1 -timeout 20m ./internal/app -run '^(TestRunETLTransportRejectsAcknowledgedCheckpointWithIncompatibleResume|TestRunETLTransportPersistsActiveCheckpointBeforeSourceFailureForAllModes|TestRunETLTransportAdvancesInterimCheckpointAcrossPages|TestRunETLTransportPreservesUnrelatedStateDuringInterimCheckpointCommit|TestRunETLTransportRejectsStaleCheckpointWriter|TestRunETLTransportDistinguishesMissingAndPresentStreamState|TestRunETLTransportCommitsAcknowledgedPageBeforeCancellation|TestRunETLTransportRetainsInterimCheckpointWhenFinalStateSaveFails|TestRunETLTransportTreatsIndeterminateCheckpointPersistenceAsFailure)$'
   ```

9. [x] Affected package (exit `0`):

   ```sh
   go test -count=1 -timeout 20m ./internal/app
   ```

10. [x] Static/build (exit `0`):

    ```sh
    go vet ./internal/app/... && go build ./cmd/pm
    ```

11. [x] Repository gates individually, not monolithically under a per-command timeout (all exit `0`):

    ```sh
    make tidy-check
    make lint
    make docs-check
    make smoke-no-build
    make agent-contract-check
    make connectorgen-validate
    make connectorgen-surface-sync
    make connector-boundary
    make release-workflow-check
    ```

## Lifecycle and delivery checks

- [x] Manual `execute-phase` record includes RED → GREEN commits and the non-numbered-phase fallback in `SUMMARY.md`.
- [x] Manual `verify-work` record includes the completed matrix and truthful non-certification boundary in `UAT.md`.
- [x] Manual `code-review` record dispositions every finding in `REVIEW.md`.
- [ ] Fresh `no-mistakes axi run --intent <complete R9 intent> --skip=push,pr,ci` is started only after local review; it never uses `--yes` and never controls the old run.
- [ ] Pushed stacked PR targets `feat/3862-any-to-any-transport`, remains unmerged, and records review/CI status without a closing keyword.

## Explicitly out of scope

No provider, credential, network, warehouse, container, external service, reverse-ETL execution, generic write surface, or certification operation is evidence for this phase.

## Completed local evidence

- Review-fix focused verification (exit `0`):

  ```sh
  go test -race -count=1 -timeout 20m ./internal/app -run '^(TestRunETLTransportStaleWriterDoesNotReportUncommittedFinalization|TestFailRunTransportConflictRequiresRunningTarget|TestFailRunTransportConflictReturnsMayHaveCommittedFinalization)$'
  ```

  It proves a real stale CAS does not return a speculative failed run after a definite pre-rename state failure, retains the typed conflict, leaves the durable loser `running` after reopen, preserves winner/unrelated state, and still returns terminal runs for committed and indeterminate outcomes.

- The race command passed with the platform linker warning emitted by the macOS toolchain but no Go race detector or test failure.
- `go vet ./internal/app/... && go build ./cmd/pm` passed.
- `make tidy-check`, `make lint`, `make docs-check`, `make smoke-no-build`, `make agent-contract-check`, `make connectorgen-validate`, `make connectorgen-surface-sync`, `make connector-boundary`, and `make release-workflow-check` each passed individually. The smoke gate used its hermetic temporary sample project only; no live provider, credential, or external service was contacted.
