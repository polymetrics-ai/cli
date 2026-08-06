# Verification checklist — issue-3755-rate-limit-operator-output-r1

## Behavior

- [x] A declared test-only bundle shows policy ID, subject kind, selection reason, selected policy count, local pacing duration, provider remaining budget when reported, provider 429 status/wait, and separate request latency. Evidence: `TestRateLimitReportShowsDeclaredPolicyPacingAndProviderPushbackWithoutSecrets`; `TestRunETLPersistsDeclaredRateLimitSummaryFromTestBundle`.
- [x] An absent bundle declaration reports `undeclared`, never unlimited, and does not attach a policy. Evidence: `TestRateLimitReportCallsAbsentDeclarationUndeclared`; `TestETLRunRateLimitOutputIsStructuredHumanReadableAndSecretFree`.
- [x] A valid provider reset/Retry-After remains honored exactly; the observation path only reports it. Evidence: existing connsdk suite passed unchanged, including typed `RateLimitError` / retry timing coverage.
- [x] A long run has a bounded summary: coalesced policies and scalar counters/durations, no per-request output. Evidence: `TestRateLimitReportCoalescesLongRunsIntoBoundedPolicySummary`.
- [x] A report never contains credentials, secret map values, token-derived values, raw bindings, opaque scope key, runtime subject, selector runtime values, raw headers/bodies/URLs, or `CredentialRevision`. Evidence: declared-bundle sentinel regression checks both JSON and `HumanLines`.

## CLI parity

- [x] `pm help etl` and bare `pm etl` show the updated output contract; no new command or flag exists. The pre-existing `pm etl run --help` route attempts a run (even in an initialized temp project) rather than rendering leaf help, so it is unchanged and documented here as not applicable to this output-only change.
- [x] `pm etl run` shows a concise rate-limit breakdown after the existing completion line. Evidence: `TestETLRunRateLimitOutputIsStructuredHumanReadableAndSecretFree`.
- [x] `pm etl run --json` carries the same summary as `run.rate_limit`. Evidence: `TestRunETLPersistsDeclaredRateLimitSummaryFromTestBundle`; CLI JSON regression.
- [x] `docs/cli/**`, relevant website docs, and generated help/manual output explain the structured summary and `undeclared` state. Evidence: docs generator, golden transcript update, website data generation, and grep verification.

## Local gates

- [x] focused core, engine, app, and CLI tests pass; full changed-package tests passed separately for connectors, coordination, connsdk, engine, app, and CLI.
- [x] focused race test passes: `go test -race ./internal/connectors ./internal/connectors/engine ./internal/app -run 'Test.*RateLimit' -count=1`.
- [x] `gofmt -w internal cmd` left the worktree clean.
- [x] targeted `go vet` and `go build ./cmd/pm` pass.
- [x] `go run ./cmd/connectorgen validate` and `surface-sync --check` pass (550 connectors; zero findings/drift).
- [x] individual project gates pass: `tidy-check`, lint, docs-check, smoke-no-build, agent-contract-check, connector-boundary, and release-workflow-check.
- [x] generated `verify-work` and `code-review` prompts are applied inline; no gap plan is required.
