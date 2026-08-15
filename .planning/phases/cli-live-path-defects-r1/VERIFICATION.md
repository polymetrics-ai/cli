# Verification: live-path defects r1

## Required checks

- [x] Focused #4119 requester/rate-admission tests, including zero-send refusal
  and formatter typed-error preservation.
- [x] `go test -timeout 20m ./internal/coordination/...` for #4125 bounds.
- [x] Focused #4169 real CLI construction-path and classification tests.
- [!] `go test -timeout 20m ./internal/cli/...` was started, but the harness
  detached the package process without a recoverable result. Its previously
  isolated pre-existing failure is recorded below; all modified behavior has
  focused package and real-binary green evidence.
- [x] `go vet ./...`.
- [x] `go build ./cmd/pm`.
- [x] `go run ./cmd/connectorgen validate internal/connectors/defs` and
  `go run ./cmd/connectorgen surface-sync --check` when the final diff touches
  connector definitions or generator-derived surfaces; otherwise record N/A.
- [x] `make tidy-check`, `make lint`, `make docs-check`, `make smoke-no-build`,
  `make agent-contract-check`, `make connector-boundary`, and
  `make release-workflow-check` individually, per the scoped verification rule.
- [x] `git diff --check`.
- [x] Inline `verify-work` and `code-review` evidence with all actionable
  findings fixed or explicitly dispositioned.
- [ ] Push direct PR and confirm the API-reported base is
  `integration/4015-mvp-flat-r1`.

## CLI parity disposition

- N/A: no command, flag, help topic, output schema, manual, website page, or
  generated help artifact changes. The typed provider-error classification is
  proven via focused CLI/binary tests instead.

## Completed local evidence

- `go test -timeout 20m ./internal/connectors/engine -run
  '^(TestEndpointSharedRateLimitAdmissionUsesRedirectDestination|TestEndpointLocalRateLimitAdmissionAllowsRedirectDestination|TestEndpointSharedRateLimitAdmissionCanonicalizesBasePrefixedRedirectDestination)$'`
  passed.
- `go test -timeout 20m ./internal/coordination` passed.
- `go test -timeout 20m ./internal/connectors/connsdk` and
  `go test -timeout 20m ./internal/connectors/engine` passed.
- `go test -timeout 20m ./internal/cli -run
  '^(TestClassifyErrorProvider401IsCredentialError|TestClassifyErrorInternalFailureRemainsInternal|TestWriteErrorProvider401RedactsCredential)$'`
  passed.
- `go test -timeout 20m ./internal/cli -run
  '^TestFreshBinaryProvider401IsCredentialErrorWithoutWritesOrCheckpointAdvance$'`
  passed: a freshly built binary emitted `auth/credential_error`, sent one
  provider read, sent zero writes, and did not advance its checkpoint.
- `go vet ./...`, `go build ./cmd/pm`, `make tidy-check`, `make lint`,
  `make docs-check`, `make smoke-no-build`, `make agent-contract-check`,
  `make connector-boundary`, `make release-workflow-check`,
  `connectorgen validate`, `surface-sync --check`, and `git diff --check`
  passed.
- CLI parity smoke: `./pm connectors`, `./pm help github`, and
  `./pm github issues list --help` all exited successfully. No docs/help
  artifact is changed because the command surface and output schema are
  unchanged.

## Independent pre-existing failure

`TestFreshBinaryDeclarativeGitHubWarehouseFlowRoundTrip` failed before this
phase's classification implementation at a valid flow control step with exit
3, before its provider-401 assertion. The test is intentionally retained with
the corrected auth expectation; this phase neither relaxes it nor changes the
unrelated flow-control path.
