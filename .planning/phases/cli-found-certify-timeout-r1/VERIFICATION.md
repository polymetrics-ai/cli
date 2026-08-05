# Verification checklist — certify-harness cost foundation (#3795)

## Planned local gates

- [x] `go test -count=1 ./internal/connectors/certify` — 58.972s; 85/85 real CLI calls (cold count reported by verbose precursor)
- [x] `go test -count=1 -run '^TestCertifyCLI' ./internal/cli` — 43.128s; 61/61 real CLI calls
- [x] `go test -count=1 ./internal/cli`
- [ ] `go test -race ./internal/connectors/certify` — the complete package reached Go's built-in 10-minute test timeout while the retained real CLI proof was running; do not rerun the same broad race command. The changed scripted-stage contract passed its focused race command.
- [ ] `go test -race ./internal/cli` — the changed CLI route/fixture contract passed `go test -race -count=1 -run '^TestCertifyCLI' ./internal/cli`.
- [x] timing parser/runner pass and controlled failure modes — `go test -count=1 ./internal/certifytiming ./cmd/certifytiming`
- [x] two cold local `-count=1 -json` timing samples: (1) harness 57.745s, CLI 42.359s, total 100.104s; (2) harness 56.539s, CLI 43.328s, total 99.867s. These diagnostics do not set the hosted wall-time limit.
- [ ] `go vet ./...`
- [ ] `go build ./cmd/pm`
- [ ] `make tidy-check`
- [x] `make lint`
- [ ] `make docs-check`
- [ ] `make smoke-no-build`
- [ ] `make agent-contract-check`
- [ ] `make connectorgen-validate`
- [ ] `make connectorgen-surface-sync`
- [ ] `make connector-boundary`
- [ ] `make release-workflow-check`
- [ ] `git diff --check`

## CI evidence required before measured threshold

- [ ] At least two post-refactor GitHub-hosted Verify samples from the timing
  target, each cold (`-count=1`) and carrying raw `go test -json` events.
- [ ] Record run/job URL, retrieval date, runner environment if exposed,
  certify elapsed time, CLI elapsed time, slowest tests, and sample maximum.
- [ ] Record the arithmetic used to derive the explicit threshold and explain
  its margin. A local laptop measurement is not sufficient for #3807.

### Hosted timing observations (retrieved 2026-08-06)

- Diagnostic sample before the wall-clock field: [Verify run 31050792264 / job
  92457226995](https://github.com/polymetrics-ai/cli/actions/runs/31050792264/job/92457226995)
  emitted raw cold events, 85/85 harness calls, 87.680s certify package elapsed,
  66.624s CLI package elapsed, and 154.304s total. The slowest retained proofs
  were `TestFullSweepSourceStagesAgainstSample` (76.700s) and
  `TestCertifyCLISingleConnectorPassExitsZero` (43.190s). This run failed only
  at the subsequently fixed `errcheck` findings, and it does not count toward
  the two wall-clock samples because that field was not yet emitted.

## Deliberate exclusions

No provider/credential/runtime checks are run. Aggregate `go test ./...` and
`make verify` remain CI gates because repository guidance documents that their
single-command duration exceeds the agent timeout. No CLI documentation parity
change is required because production CLI surface remains unchanged.
