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
- [x] Direct PR #4173 is open and its API-backed `gh-axi pr list` query scoped
  by both head and `--base integration/4015-mvp-flat-r1` returned exactly that
  PR. The target base is therefore verified as
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
- After absorbing merged base `ec1f200c9`, `go test -count=1 -timeout 20m
  ./internal/cli -run '^TestFreshBinaryDeclarativeGitHubWarehouseFlowRoundTrip$'`
  passed.
- `go vet ./...`, `go build ./cmd/pm`, `make tidy-check`, `make lint`,
  `make docs-check`, `make smoke-no-build`, `make agent-contract-check`,
  `make connector-boundary`, `make release-workflow-check`,
  `connectorgen validate`, `surface-sync --check`, and `git diff --check`
  passed.
- CLI parity smoke: `./pm connectors`, `./pm help github`, and
  `./pm github issues list --help` all exited successfully. No docs/help
  artifact is changed because the command surface and output schema are
  unchanged.

## Resolved base dependency

`TestFreshBinaryDeclarativeGitHubWarehouseFlowRoundTrip` initially failed
before this phase's classification implementation at a valid flow-control step
with exit 3, before its provider-401 assertion. Merged PR #4174 replaced the
stale inline-action fixture. After fetching the updated integration base and
proving `ec1f200c9c9b11d1b2f54505bae2ea6c3a621f63` was contained in it, this
branch merged that base as `aa3704271`. The exact fresh-binary round-trip test
then passed locally without relaxing its assertion.

## Durable-parking CI triage (not a phase behavior change)

`TestCLIDurableParkingAdmissionAndResumeAcrossKilledProcess` failed in the PR
CI run at its concurrent `resume-race` assertion. A clean comparison showed a
failure at dispatch base `ef3c71caf` and a pass at the phase head, so the
phase's shared-window bounds were not causally attributable. Repeating the
same exact test at `ef3c71caf` produced six passes and one failure across the
seven directly observed runs. The host's one-minute load averaged 11.6–16.0 on
12 CPUs during that sample. This is evidence of a load-sensitive pre-existing
race, not a reason to relax its real-process assertion.

Commit `329699f2a` preserves the assertion and improves its failure evidence:
the parent now receives the losing helper's exit code and sanitized CLI output.
It adds no new behavior and is not a happy/bad/edge replacement for the three
issue-specific tests above. The underlying concurrency repair remains outside
this PR's scope.
