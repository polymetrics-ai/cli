# #4063 TDD Ledger

**Status:** GREEN generated; correction 4 / 5 reserved and applied; focused gates pending.
**Exact start:** 002ddf674a447bf0872486aa979efdaa078f602c
**Required base:** feat/3988-github-certification

| Stage | Command / assertion | Expected result | Status |
|---|---|---|---|
| Precondition | #3897 authoritative TDD ledger + RUN-STATE | correction 3 / 5 | PASS |
| RED | go run ./cmd/connectorgen certification-matrix --check | exit 1 on stale flow matrix coordinate | PASS (expected exit 1) |
| GREEN | go run ./cmd/connectorgen certification-matrix | only source coordinate :20 -> :21 changes | PASS |
| GREEN check | go run ./cmd/connectorgen certification-matrix --check | exit 0 | PENDING |
| Focused regression | go test -timeout 20m ./cmd/connectorgen -run '^(TestCertificationDiscoverFunctionKindsFromRuntimeSource|TestCertificationGeneratedArtifactDriftFails)$' -count=1 | pass | PASS |
| Generator package | go test -timeout 20m ./cmd/connectorgen -count=1 | pass | PASS |
| Focused affected packages | go test -timeout 20m ./internal/flow ./internal/app ./internal/warehouse ./internal/cli -count=1 | pass | PASS |
| Refactor | none | generated-only metadata correction needs no refactor | NOT APPLICABLE |

## RED contract

The RED assertion is intentionally the canonical generated-artifact check,
not a new test file. Before generation, it must report flow-matrix.json drift.
The failure is valid only when the checked-out head equals the exact start and
the source / artifact coordinate mismatch is :21 / :20.

## GREEN contract

The canonical generator is the only writer. A valid GREEN result has exactly
one semantic JSON leaf change:

  workflow_kinds[id=flow_authoring].discovery_source
  internal/cli/flow_cli.go:20 -> internal/cli/flow_cli.go:21

No production code, test, capability, status, evidence, definition, or other
generated artifact may change.

## Observed RED

At exact HEAD 002ddf674a447bf0872486aa979efdaa078f602c, the canonical check
exited 1 with:

    connectorgen certification-matrix: generated artifact ".../flow-matrix.json"
    has drift; run go run ./cmd/connectorgen certification-matrix
    exit status 1

The inspected source marker/function coordinate is :20 / :21 while the
generated flow_authoring discovery_source is :20. The full non-secret command
record is traces/red-certification-matrix-check.txt.

## Observed GREEN

The canonical generator completed successfully. A recursive JSON comparison
against exact start 002ddf674a447bf0872486aa979efdaa078f602c found exactly
one leaf:

    workflow_kinds[1].discovery_source
    internal/cli/flow_cli.go:20 -> internal/cli/flow_cli.go:21

The only generated path changed after the pre-correction evidence commit was
internal/connectors/certifications/flow-matrix.json. The full non-secret
semantic-boundary record is traces/green-semantic-boundary.txt.

## Focused test evidence

The focused generator regression test, complete connectorgen package test, and
the required internal/flow, internal/app, internal/warehouse, and internal/cli
package run all passed. The command record is
traces/local-focused-gates.txt.
