# Verification checklist — certify-harness cost foundation (#3795)

## Pre-remediation 93/61-call measurement

- [x] Two final-topology `make certify-timing` measurements held the 93/93
  harness and 61/61 CLI caps. The first had harness elapsed/wall time
  79.229s/84.769s, CLI elapsed/wall time 70.000s/76.467s, and total
  elapsed/wall time 149.229s/161.236s. The focused review verification had
  harness elapsed/wall time 81.733s/86.645s, CLI elapsed/wall time
  54.825s/60.665s, and total elapsed/wall time 136.558s/147.309s.
- [x] The harness measurement retains exactly one real sample/outbox
  plan → preview → approval → execute → cleanup/replay lifecycle proof,
  `TestSampleOutboxWriteLifecycleAgainstRealCLI`, alongside the real full
  source/read/flow/schedule sweep.
- [x] The raw final-topology measurements historically produced the superseded
  `180s` result: the maximum wall measurement is 161.236s and the observed
  spread is `161.236 - 147.309 = 13.927s`; `161.236 + 13.927 = 175.163s`,
  rounded up to the next 15-second boundary, gives 180s.
- [x] `CERTIFY_TIMING_MAX_DURATION` is `3m30s` (210 seconds), the active
  hosted-measurement-derived cap; the raw 180-second calculation above is
  retained only as superseded history.

The pre-refactor harness made 782 real invocations. That pre-remediation
topology recorded the superseded 3-minute wall cap. The active
hosted-measurement-derived cap remains 3m30s (210 seconds).

## CI remediation — coalesced full-route measurement

- [x] `TestCertifyCLISingleConnectorPassExitsZero` now runs
  `pm connectors certify sample --full --json` and checks the full source/read/
  flow/schedule report assertions alongside the CLI envelope and persistence.
  The separate direct full-sweep test was removed because it duplicated that
  executable topology.
- [x] The revised exact invocation caps are 25/25 in
  `internal/connectors/certify` and 92/92 in `internal/cli`.
- [x] `make certify-timing` passed with harness elapsed/wall time
  15.252s/19.691s, CLI elapsed/wall time 77.330s/80.413s, and total
  elapsed/wall time 92.582s/100.104s—well below the active 3m30s
  (210-second) duration limit.
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

## Post-rebase compatibility repair

- [x] RED: hosted Verify run `31095059925` on rebased commit `428324630`
  reports `certify CLI real invocations: 93 (budget 92)`; the harness package
  remains 25/25.
- [x] GREEN: `TestCertifyCLIHelpShowsProvenanceContract` preserves its
  provenance-help assertions through the direct `runConnectors` router branch.
  `make certify-timing` returned 25/25 harness calls and 92/92 CLI calls with
  79.895s elapsed / 88.553s wall time, without changing the active 3m30s
  (210-second) hosted timing cap.
- [ ] Hosted Verify: confirm the repaired PR head passes GitHub Verify.

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
