# Verification checklist — certify-harness cost foundation (#3795)

## Hosted final-topology calibration

- [x] The pre-refactor harness made 782 real CLI invocations. The final exact
  topology is 25 harness invocations plus 92 CLI invocations, or 117 total.
- [x] [Hosted Verify sample 1](https://github.com/polymetrics-ai/cli/actions/runs/31067299477)
  on 2026-08-06 measured the final 25/92 topology at 94.662s total elapsed
  and 125.655s wall elapsed.
- [x] [Hosted Verify sample 2](https://github.com/polymetrics-ai/cli/actions/runs/31068408270)
  on 2026-08-06 at `3332dc69378f39e10637a0d58777d4ddce6d51f4` measured the same
  final 25/92 topology at 124.013s total elapsed and 165.891s wall elapsed.
- [x] `CERTIFY_TIMING_MAX_DURATION` is `3m30s`. The maximum final wall
  time plus the observed spread is `165.891 + (165.891 - 125.655) =
  206.127s`; rounding up to the existing 15-second granularity yields
  210 seconds.
- [x] The harness retains exactly one real sample/outbox plan → preview →
  approval → execute → cleanup/replay proof,
  `TestSampleOutboxWriteLifecycleAgainstRealCLI`. The full source/read/flow/schedule
  proof remains coalesced in the real CLI route.

## CI remediation — coalesced full-route measurement

- [x] `TestCertifyCLISingleConnectorPassExitsZero` now runs
  `pm connectors certify sample --full --json` and checks the full source/read/
  flow/schedule report assertions alongside the CLI envelope and persistence.
  The separate direct full-sweep test was removed because it duplicated that
  executable topology.
- [x] The revised exact invocation caps are 25/25 in
  `internal/connectors/certify` and 92/92 in `internal/cli`.
- [x] The intermediate local `make certify-timing` run passed with harness elapsed/wall time
  15.252s/19.691s, CLI elapsed/wall time 77.330s/80.413s, and total
  elapsed/wall time 92.582s/100.104s. It is not the source of the enforced
  hosted final-topology duration limit.
- [x] `go test -count=1 ./internal/connectors/certify`,
  `go test -count=1 -run '^TestCertifyCLI' ./internal/cli`,
  `go test -count=1 ./internal/certifytiming ./cmd/certifytiming`, and
  `go test -count=1 ./internal/cli` passed. The complete CLI package took
  458.490s.

## Write-inventory contract

A green skip is indistinguishable from a check that ran and passed. Therefore
a missing `writes.json` remains a failing write-inventory stage; the retained
real lifecycle proof comes from the definition-backed sample/outbox setup
instead of silently passing a missing inventory. This avoids the
detector-disarming implemented-command failure pattern in which 213 declared
implemented commands passed checks but failed at runtime, alongside the public
485 generally-available/17 command-response disparity.

## Scoped verification record

- [x] timing parser/runner pass and controlled failure modes —
  `go test -count=1 ./internal/certifytiming ./cmd/certifytiming`
- [x] `go vet ./...`
- [x] `go build ./cmd/pm`
- [x] `make tidy-check`
- [x] `make lint`
- [x] `make docs-check`
- [x] `make smoke-no-build`
- [x] `make agent-contract-check`
- [x] `make connectorgen-validate`
- [x] `make connectorgen-surface-sync`
- [x] `make connector-boundary`
- [x] `make release-workflow-check`

A fresh hosted Verify run remains required for the final commit range. No
provider, credential, or runtime checks are run. Aggregate `go test ./...` and
`make verify` remain CI gates because their single-command duration exceeds the
agent timeout. No CLI documentation parity change is required because the
production CLI surface remains unchanged.
