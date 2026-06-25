# TDD Ledger

Phase: agentic-etl-platform

## Baseline

- Command: `make verify`
- Result: passed before implementation edits.
- Notes: bundled GSD preflight ran from `/Users/karthiksivadas/.codex/skills/gsd-programming-loop/scripts/programming-loop.mjs`; repo-local `scripts/programming-loop.mjs` is absent.

## Red/Green/Refactor Entries

Entries are appended before each production-code slice.

## Gate Evidence Index

Status: red-confirmed

- record-baseline-verification-and-prompt-snapshot: baseline recorded before production edits.
- add-red-tests-for-structured-errors-and-json-error-contracts: covered by `TestUnknownCommandJSONErrorIsStructuredAndSanitized`.
- implement-typed-cli-errors-and-sanitized-stderr: implementation proceeded only after the structured error red test failed.
- add-red-tests-for-validators-and-terminal-sanitizer: covered by `internal/safety` tests.
- implement-shared-validation-sanitization-package: implementation proceeded only after the safety package red test failed.
- add-red-tests-for-connector-manifests-and-secret-redaction: covered by `TestConnectorInspectJSONIncludesManifest`.
- implement-connector-manifests-and-manifest-backed-inspection: implementation proceeded only after manifest red test failed.
- add-red-tests-for-generated-skills: covered by `TestSkillsGenerateWritesAgentSkills`.
- implement: implementation of `poly skills generate` proceeded only after the skills red test failed.
- add-red-tests-for-streaming-etl-batches: covered by `TestRunETLWritesBoundedBatches`.
- implement-bounded-etl-batch-writes-and-checkpoint-metadata: implementation proceeded only after streaming ETL red test failed.
- run-full-local-verification-and-update-phase-artifacts: final verification evidence is recorded in `VERIFICATION.md`.

### Red: Structured Errors, Safety, Manifests, Skills, Streaming ETL

- Command: `go test ./internal/safety ./internal/cli ./internal/app`
- Result: failed as expected.
- Evidence:
  - `internal/safety`: no non-test Go files.
  - `internal/app`: `RunETLRequest.BatchSize`, `Run.BatchCount`, and `Run.Checkpoint` missing.
  - `internal/cli`: unknown command with `--json` does not emit JSON error.
  - `internal/cli`: connector inspect output lacks `manifest`.
  - `internal/cli`: `skills generate` command missing.

### Green: Structured Errors, Safety, Manifests, Skills, Streaming ETL

- Command: `gofmt -w internal && go test ./internal/safety ./internal/cli ./internal/app`
- Result: passed.
- Implemented:
  - `internal/safety` sanitizer and validators.
  - Structured JSON CLI errors with stable categories.
  - Connector manifests and manifest-backed connector inspection.
  - `poly skills generate`.
  - `RunETLRequest.BatchSize`, `Run.BatchCount`, `Run.Checkpoint`, and bounded ETL batch writes.

### Red: CLI Parse-Boundary Validation

- Command: `go test ./internal/cli -run TestConnectorInspectRejectsUnsafeIdentifier`
- Result: failed as expected.
- Evidence: unsafe connector identifier was treated as `internal_error` with exit code 1 instead of `validation_error` with exit code 3.

### Green: CLI Parse-Boundary Validation

- Command: `gofmt -w internal/cli && go test ./internal/cli -run TestConnectorInspectRejectsUnsafeIdentifier`
- Result: passed.
- Implemented connector and endpoint identifier validation at CLI parse boundaries.
