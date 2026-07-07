# TDD Ledger

## Red

- Added commandrunner tests for write-command planning, dry-run preview, non-mutating generic
  runner behavior, and invalid reverse-ETL metadata.
- Updated CLI expectations so `pm github issue create --title ...` creates a plan instead of being
  policy-blocked.
- Added CLI E2E test for `pm github issue close` plan, preview, invalid approval, and approved run.

## Green

- Added `commandrunner.BuildWriteCommand` for `record.*` flag mapping, validation, dry-run preview,
  mutation metadata, and redacted record projection.
- Added app-level `PlanConnectorCommand`, `PreviewConnectorCommandPlan`, and connector-command
  execution branch in `RunReverseETL`.
- Added CLI routing for provider-style write commands through reverse-plan creation, preview, and
  approved execution.
- Added initial GitHub command flag mappings for `issue create`, `issue close`, and
  `repo deploy-key add`.

## Refactor

- Split preflight command resolution from generic read execution so reverse-ETL commands can be
  recognized without making `commandrunner.Run` mutating.
- Centralized safe reverse-plan serialization in CLI output.
- Reused existing reverse-plan token hashing, expiry, replay protection, and plan hash validation.

## Evidence

- `go test ./internal/connectors/commandrunner ./internal/app ./internal/cli -run 'TestBuildWriteCommand|TestRunReverseETLCommandRemainsNonExecutableInGenericRunner|TestRunReverseETLRejectsMissingWriteAndUnsupportedFlagMapping|TestGitHubCommandSurfacePlansReverseETLCommand|TestGitHubCommandWriteUsesReversePlanApproval|TestRunReverseETLRejectsApprovalTokenReplay|TestRunReverseETLRejectsPlanHashMismatchWhenRowsChange'`
- `go run ./cmd/connectorgen validate internal/connectors/defs --json`
- `go test ./cmd/connectorgen ./internal/connectors/engine`
- `git diff --check`
