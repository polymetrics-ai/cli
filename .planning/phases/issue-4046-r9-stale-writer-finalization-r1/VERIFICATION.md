# #4046 R9 verification checklist

**Status:** Behavioral RED observed at the pinned base; production change not started.

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

3. [ ] Intervening unrelated writer:

   ```sh
   go test -count=20 -timeout 20m ./internal/app -run '^TestFailRunTransportConflictPreservesLatestConcurrentState$'
   ```

4. [ ] Restart truth:

   ```sh
   go test -count=10 -timeout 20m ./internal/app -run '^TestRunETLTransportStaleWriterFailureSurvivesReopen$'
   ```

5. [ ] Cancellation:

   ```sh
   go test -count=10 -timeout 20m ./internal/app -run '^TestRunETLTransportStaleWriterFinalizesAfterCancellation$'
   ```

6. [ ] All seven modes:

   ```sh
   go test -count=1 -timeout 20m ./internal/app -run '^TestRunETLTransportStaleWriterFinalizesLosingRunForAllModes$' -v
   ```

7. [ ] Race/interleaving:

   ```sh
   go test -race -count=10 -timeout 20m ./internal/app -run '^(TestRunETLTransportStaleWriterFinalizesLosingRun|TestFailRunTransportConflictPreservesLatestConcurrentState|TestRunETLTransportStaleWriterFinalizesAfterCancellation)$'
   ```

8. [ ] Existing R7/R8 regression:

   ```sh
   go test -count=1 -timeout 20m ./internal/app -run '^(TestRunETLTransportRejectsAcknowledgedCheckpointWithIncompatibleResume|TestRunETLTransportPersistsActiveCheckpointBeforeSourceFailureForAllModes|TestRunETLTransportAdvancesInterimCheckpointAcrossPages|TestRunETLTransportPreservesUnrelatedStateDuringInterimCheckpointCommit|TestRunETLTransportRejectsStaleCheckpointWriter|TestRunETLTransportDistinguishesMissingAndPresentStreamState|TestRunETLTransportCommitsAcknowledgedPageBeforeCancellation|TestRunETLTransportRetainsInterimCheckpointWhenFinalStateSaveFails|TestRunETLTransportTreatsIndeterminateCheckpointPersistenceAsFailure)$'
   ```

9. [ ] Affected package:

   ```sh
   go test -count=1 -timeout 20m ./internal/app
   ```

10. [ ] Static/build:

    ```sh
    go vet ./internal/app/... && go build ./cmd/pm
    ```

11. [ ] Repository gates individually, not monolithically under a per-command timeout:

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

- [ ] Manual `execute-phase` record includes RED → GREEN commits and any deviation.
- [ ] Manual `verify-work` record includes the completed matrix and truthful non-certification boundary.
- [ ] Manual `code-review` record dispositions every finding.
- [ ] Fresh `no-mistakes axi run --intent <complete R9 intent> --skip=push,pr,ci` is started only after local review; it never uses `--yes` and never controls the old run.
- [ ] Pushed stacked PR targets `feat/3862-any-to-any-transport`, remains unmerged, and records review/CI status without a closing keyword.

## Explicitly out of scope

No provider, credential, network, warehouse, container, external service, reverse-ETL execution, generic write surface, or certification operation is evidence for this phase.
