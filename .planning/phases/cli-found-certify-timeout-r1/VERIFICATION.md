# Verification checklist — certify-harness cost foundation (#3795)

## Final 93-call measurement

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
- [x] `CERTIFY_TIMING_MAX_DURATION` is `3m`. The maximum final-topology wall
  measurement is 161.236s and the observed spread is
  `161.236 - 147.309 = 13.927s`; `161.236 + 13.927 = 175.163s`, rounded up to
  the next 15-second boundary, gives 180s.

The pre-refactor harness made 782 real invocations. The current exact 93-call
cap prevents added real work from being hidden as duplicated coverage, while
the 3-minute wall cap detects broad cost regressions.

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
