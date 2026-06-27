# TDD Ledger

Phase: flow-engine

Record failing test evidence before production code for every behavior-adding task.

---

All entries below are red-confirmed. Stubs return `errors.New("not implemented")` only.
Run command: `GOTOOLCHAIN=auto go test ./internal/flow/... 2>&1`

---

## T-01 — Manifest parse/validate

Status: red-confirmed

Files: `internal/flow/flow_test.go` (TestManifestParse, TestManifestValidate)

Failing output snippet:
```
--- FAIL: TestManifestParse/valid_two-step_manifest_round-trips
    flow_test.go:48: Received unexpected error: not implemented

--- FAIL: TestManifestValidate/valid_manifest_produces_no_errors
    flow_test.go:166: Should be empty, but was [not implemented]

--- FAIL: TestManifestValidate/empty_name_is_invalid
    flow_test.go:162: Should be true — expected ErrManifestInvalid, got not implemented
    (same pattern for version_2, unknown_kind, missing_connection, empty_streams,
     missing_sql, duplicate_IDs, bad_in_reference)
```

What's missing: `ParseManifest` must decode JSON; `ValidateManifest` must implement all
nine rules wrapping errors with `ErrManifestInvalid`.

---

## T-02 — DAG build + topological sort + cycle detection

Status: red-confirmed

Files: `internal/flow/flow_test.go` (TestDAGLinearChain, TestDAGIndependentSteps,
TestDAGTwoCycle, TestDAGThreeNodeCycle, TestDAGDiamond)

Failing output snippet:
```
--- FAIL: TestDAGLinearChain
    flow_test.go:188: Received unexpected error: not implemented

--- FAIL: TestDAGTwoCycle
    flow_test.go:212: Should be true — expected ErrCyclicDependency, got not implemented

--- FAIL: TestDAGDiamond
    flow_test.go:236: Received unexpected error: not implemented
```

What's missing: `BuildDAG` with Kahn's algorithm, adjacency list from `in`/`out` table names,
cycle detection, ordered `[]string` result.

---

## T-03 — Checkpoint store

Status: red-confirmed

Files: `internal/flow/flow_test.go` (TestCheckpointSetGet, TestCheckpointGetUnknown,
TestCheckpointClearFlow, TestCheckpointClearDoesNotAffectOtherFlow, TestCheckpointConcurrentSets)

Failing output snippet:
```
--- FAIL: TestCheckpointSetGet
    flow_test.go:253: Received unexpected error: not implemented

--- FAIL: TestCheckpointGetUnknown
    flow_test.go:262: Received unexpected error: not implemented

--- FAIL: TestCheckpointConcurrentSets
    flow_test.go:303: Received unexpected error: not implemented
```

What's missing: `FileCheckpointStore.Set/Get/Clear` backed by JSON file at
`<Dir>/flow-checkpoints.json` with mutex for concurrency safety.

---

## T-04 — Engine lease contention

Status: red-confirmed

Files: `internal/flow/flow_test.go` (TestEngineLockHeldReturnsErrLeaseHeld,
TestEngineLockReleasedAfterRun)

Failing output snippet:
```
--- FAIL: TestEngineLockHeldReturnsErrLeaseHeld
    flow_test.go:389: Should be true
        Messages: expected ErrLeaseHeld, got not implemented
```

What's missing: `Engine.Run` must call `state.FileLock.Lock()` at
`<LockDir>/flow-<name>.lock`, return `ErrLeaseHeld` on contention, defer release.

---

## T-05 — Engine dependency ordering

Status: red-confirmed

Files: `internal/flow/flow_test.go` (TestEngineDependencyOrderABeforeB,
TestEngineStepAFailsStepBNotCalled, TestEngineThreeStepsMiddleFailsThirdSkipped)

Failing output snippet:
```
--- FAIL: TestEngineDependencyOrderABeforeB
    flow_test.go:467: Received unexpected error: not implemented

--- FAIL: TestEngineStepAFailsStepBNotCalled
    flow_test.go:484: Not equal: expected "failed" actual ""
```

What's missing: `Engine.Run` must iterate topo-sorted steps, dispatch `runSync`/`runQuery`,
stop on first failure, populate `RunResult.Status`.

---

## T-06 — Engine checkpoint/resume

Status: red-confirmed

Files: `internal/flow/flow_test.go` (TestEngineSkipsPreseededStep, TestEngineForceReclears,
TestEngineCheckpointsPersistedAfterRun)

Failing output snippet:
```
--- FAIL: TestEngineSkipsPreseededStep
    flow_test.go:525: Received unexpected error: not implemented

--- FAIL: TestEngineForceReclears
    flow_test.go:554: Received unexpected error: not implemented
```

What's missing: Checkpoint read before dispatch; `Force` clears checkpoints;
`StepResult.Status="skipped"` when pre-seeded; `Set("success")` after successful dispatch.

---

## T-07 — Engine ledger writes

Status: red-confirmed

Files: `internal/flow/flow_test.go` (TestEngineLedgerEntriesSuccessfulRun,
TestEngineLedgerEntriesFailedRun)

Failing output snippet:
```
--- FAIL: TestEngineLedgerEntriesSuccessfulRun
    flow_test.go:605: Received unexpected error: not implemented

--- FAIL: TestEngineLedgerEntriesFailedRun
    flow_test.go:636: Not equal: expected "failed" actual ""
    flow_test.go:643: Not equal: expected 1 actual 0
```

What's missing: `Engine.Run` must call `LedgerAdapter.Append` with status=running before
dispatch and status=success/failed after; one additional flow-level record on completion.

---

## T-08 — CLI flow subcommand

Status: red-confirmed (package build failure)

Files: `internal/cli/flow_cli_test.go` (TestFlowList, TestFlowPlanValid, TestFlowPlanCyclic,
TestFlowStatusMissing, TestFlowPreviewValid)

Failing output snippet:
```
# polymetrics.ai/internal/cli [polymetrics.ai/internal/cli.test]
internal/cli/runtime_record_test.go:21:11: undefined: runtimeETLLeaseRequest
internal/cli/runtime_record_test.go:25:12: undefined: runtimeETLRunRecord
FAIL    polymetrics.ai/internal/cli [build failed]
```

What's missing: `runtimeETLLeaseRequest` and `runtimeETLRunRecord` are pre-existing stubs not
yet implemented (tracked separately). Once that blocker is resolved, the flow CLI tests will
fail for the right reason: `runFlow` stub returns `errors.New("not implemented")`.

---

## T-09 — Integration smoke

Status: red-confirmed

Files: `internal/flow/flow_test.go` (TestEngineIntegrationSyncQueryChain)

Failing output snippet:
```
--- FAIL: TestEngineIntegrationSyncQueryChain
    flow_test.go:659: Received unexpected error: not implemented
```

What's missing: Full `Engine.Run` implementation (gated on T-04 through T-07 green).

---

## Summary table

| Task | Test function(s) | Red reason |
|------|-----------------|------------|
| T-01 | TestManifestParse, TestManifestValidate | `not implemented` / wrong error sentinel |
| T-02 | TestDAGLinearChain…TestDAGDiamond | `not implemented` / wrong error sentinel |
| T-03 | TestCheckpointSetGet…TestCheckpointConcurrentSets | `not implemented` |
| T-04 | TestEngineLockHeldReturnsErrLeaseHeld, TestEngineLockReleasedAfterRun | wrong error sentinel |
| T-05 | TestEngineDependencyOrderABeforeB…TestEngineThreeStepsMiddleFailsThirdSkipped | `not implemented` |
| T-06 | TestEngineSkipsPreseededStep…TestEngineCheckpointsPersistedAfterRun | `not implemented` |
| T-07 | TestEngineLedgerEntriesSuccessfulRun, TestEngineLedgerEntriesFailedRun | `not implemented` |
| T-08 | TestFlowList…TestFlowPreviewValid | package build failure (pre-existing sibling stub) |
| T-09 | TestEngineIntegrationSyncQueryChain | `not implemented` |
