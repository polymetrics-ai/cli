# TDD ledger — #3792 operation direct-read preflight and surface reconciliation

| ID | Contract | Red evidence | Green evidence |
| --- | --- | --- | --- |
| R1 | A generic `OperationDirectReader` alone cannot make an implemented command preflight. | Existing unmerged proof `0efd5ddff` demonstrated the current runner returns nil for an executor-rejected operation. Recreate the focused test on this branch before implementation. | Missing metadata, unsupported kind, missing/invalid cap, endpoint mismatch, and policy mismatch return `BlockedCommandError`; no execution dispatch occurs. |
| R2 | Metadata is the same static operation admission used by the engine. | A provider-search/rest-read table is absent and an unsupported operation is not rejected by a no-network metadata call. | Engine tests prove `rest_read` and `provider_search` summaries, cap and endpoint failures, and accepted policy behavior. Existing bounded provider-search execution tests remain green. |
| R3 | The global runtime sweep observes the stronger rule. | The existing sweep does not consult read metadata before this change. | `TestEveryImplementedCommandPassesRuntimePreflight` remains unchanged and passes against all declared implemented commands. |
| R4 | An `api_surface` row becomes covered only through real runtime preflight. | A stale operation-row fixture has no reconciliation command. | A matching executable command produces `covered_by`; planned/failing/unknown candidates remain blocked or are refused, and check mode writes nothing. |
| R5 | The 574 stale #2985 rows are reported without a ledger write. | No deterministic reporting command exists. | Check-only filtered run reports exact proposed covered/blocked/refused counts for Zendesk Support, HubSpot, Asana, Bitbucket, Freshchat, and YouTube Analytics. |
| R6 | A shipped operation-backed direct read cannot bypass API-surface provenance because `defs.FS` omits raw `api_surface.json`. | With no compact projection, the old surface-optional helper accepted the operation after every other static check. | The generated root projection contains only method/path/kind/max_bytes; loader and preflight reject missing, unresolved, incomplete, and malformed projections. |

## Commands

```sh
go test ./internal/connectors/engine -run 'TestOperationDirectReadMetadata|TestOperationDirectReadExecutesProviderSearch' -count=1
go test ./internal/connectors/commandrunner -run 'TestPreflightOperationDirectRead|TestEveryImplementedCommandPassesRuntimePreflight' -count=1
go test ./cmd/connectorgen -run 'Test.*SurfaceReconcile|TestRun.*SurfaceReconcile' -count=1
```

## Evidence

- **R1 red:** `go test ./internal/connectors/commandrunner -run
  '^TestPreflightOperationDirectReadRejectsNonExecutableOperationMetadata$'
  -count=1` failed before implementation: `Preflight error = nil, want loaded
  operation rejection`. The test also pins that preflight cannot dispatch
  `OperationDirectRead`.
- **R1/R3 green:** `go test ./internal/connectors/commandrunner -run
  '^(TestPreflightOperationDirectReadRejectsNonExecutableOperationMetadata|TestEveryImplementedCommandPassesRuntimePreflight)$'
  -count=1` passed after the runner began calling the real metadata preflight.
  The unchanged global sweep initially exposed Amazon SQS and Ashby, whose
  existing direct-read runtimes now expose their closed metadata contracts.
- **R2 green:** `go test ./internal/connectors/engine -run
  '^TestPreflightOperationDirectReadValidatesDeclaredContract$' -count=1`
  passed for `provider_search` and `rest_read`, including kind, method, path,
  cap, and output-policy rejection cases.
- **R4 green:** `go test ./cmd/connectorgen -run 'Test.*SurfaceReconcile'
  -count=1` passed. The fixture is covered only after real `commandrunner`
  preflight; planned and failing commands stay blocked, and an unknown model
  is refused without writing the file.
- **R5 green:** the check-only `#2985` report is recorded in
  `RECLASSIFICATION-REPORT.md`; it writes no connector definition.
- **R6 green:** `go test ./internal/connectors/engine
  ./internal/connectors/commandrunner ./cmd/connectorgen -count=1` passed. The
  new shipped-registry test proves all 247 implemented operation-backed
  direct-read commands pass with the embedded projection and reject when it is
  removed; the unchanged global commandrunner sweep passes in the same run.
