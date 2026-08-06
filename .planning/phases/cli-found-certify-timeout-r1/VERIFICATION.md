# Verification checklist — certify-harness cost foundation (#3795)

## Planned local gates

- [x] `go test -count=1 ./internal/connectors/certify` — rebase target run through `make certify-timing`: 57.399s; 85/85 real CLI calls
- [x] `go test -count=1 -run '^TestCertifyCLI' ./internal/cli` — rebase target run through `make certify-timing`: 43.335s; 61/61 real CLI calls
- [x] `go test -count=1 ./internal/cli` — 363.575s after rebase
- [ ] `go test -race ./internal/connectors/certify` — the complete package reached Go's built-in 10-minute test timeout while the retained real CLI proof was running; do not rerun the same broad race command. The changed scripted-stage contract passed its focused race command.
- [ ] `go test -race ./internal/cli` — the changed CLI route/fixture contract passed `go test -race -count=1 -run '^TestCertifyCLI' ./internal/cli`.
- [x] timing parser/runner pass and controlled failure modes — `go test -count=1 ./internal/certifytiming ./cmd/certifytiming`
- [x] enforced timing target — `make certify-timing` reports 85/85 and 61/61 invocation caps, 112.159s measured local wall time, and passes the fixed `4m15s` bound
- [x] two cold local `-count=1 -json` timing samples: (1) harness 57.745s, CLI 42.359s, total 100.104s; (2) harness 56.539s, CLI 43.328s, total 99.867s. These diagnostics do not set the hosted wall-time limit.
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
- [ ] `git diff --check` — repeat after the final delivery-record update

## CI evidence required before measured threshold

- [x] At least two post-refactor GitHub-hosted Verify samples from the timing
  target, each cold (`-count=1`) and carrying raw `go test -json` events.
- [x] Record run/job URL, retrieval date, runner environment if exposed,
  certify elapsed time, CLI elapsed time, slowest tests, and sample maximum.
- [x] Record the arithmetic used to derive the explicit threshold and explain
  its margin. A local laptop measurement is not sufficient for #3807.
- [x] Enforced hosted Verify — [run 31055574285 / job
  92472218411](https://github.com/polymetrics-ai/cli/actions/runs/31055574285/job/92472218411)
  passed with 156.405s measured target wall time below the `4m15s` bound and
  then completed the unchanged aggregate Verify gate.

### Hosted timing observations (retrieved 2026-08-06)

- Diagnostic sample before the wall-clock field: [Verify run 31050792264 / job
  92457226995](https://github.com/polymetrics-ai/cli/actions/runs/31050792264/job/92457226995)
  emitted raw cold events, 85/85 harness calls, 87.680s certify package elapsed,
  66.624s CLI package elapsed, and 154.304s total. The slowest retained proofs
  were `TestFullSweepSourceStagesAgainstSample` (76.700s) and
  `TestCertifyCLISingleConnectorPassExitsZero` (43.190s). This run failed only
  at the subsequently fixed `errcheck` findings, and it does not count toward
  the two wall-clock samples because that field was not yet emitted.

- First measured wall-clock sample: [Verify run 31053373131 / job
  92465443313](https://github.com/polymetrics-ai/cli/actions/runs/31053373131/job/92465443313)
  completed successfully on GitHub-hosted Ubuntu 24.04.4 (`ubuntu-24.04`, image
  version `20260720.247.2`). Its raw `-count=1 -json` stream reports 85/85
  harness calls, certify package elapsed/wall elapsed 87.621s/124.466s, CLI
  package elapsed/wall elapsed 66.064s/68.747s, and aggregate
  elapsed/wall elapsed 153.685s/193.213s. The slowest retained tests were
  `TestFullSweepSourceStagesAgainstSample` and
  `TestCertifyCLISingleConnectorPassExitsZero`. This is sample 1 of 2; no
  threshold was selected at this point.

- Second measured wall-clock sample: [Verify run 31054516637 / job
  92468986349](https://github.com/polymetrics-ai/cli/actions/runs/31054516637/job/92468986349)
  completed successfully on the same GitHub-hosted Ubuntu 24.04.4 image. Its
  raw `-count=1 -json` stream reports 85/85 harness calls, certify package
  elapsed/wall elapsed 63.340s/91.407s, CLI package elapsed/wall elapsed
  47.563s/49.525s, and aggregate elapsed/wall elapsed 110.903s/140.932s.
  The same retained real proof names were the slowest tests.

### Fixed measured budget

The two hosted wall samples are 193.213s and 140.932s. The first is the
maximum and their observed spread is 52.281s, so the derived pre-rounding
bound is `193.213 + 52.281 = 245.494s`. `CERTIFY_TIMING_MAX_DURATION` is
therefore the next 15-second boundary, `4m15s` (255s): 52.281s of measured
variance plus 9.506s of explicit rounding room above the slowest sample. The
Make timing target now passes that fixed value to the command, which enforces
the sum of measured target process wall times and names observed/allowed values
on failure. The deterministic 85/61 real-CLI caps remain the primary guard.

### Enforced hosted confirmation

[Verify run 31055574285 / job
92472218411](https://github.com/polymetrics-ai/cli/actions/runs/31055574285/job/92472218411),
retrieved 2026-08-06, ran `go run ./cmd/certifytiming --max-duration 4m15s`
on the same GitHub-hosted runner. It held the 85/85 and 61/61 deterministic
counts, reported 101.291s harness wall time plus 55.114s CLI wall time
(156.405s total), and passed the aggregate Verify gate. This confirmation
precedes the required rebase-on-current-main validation.

## Rebase validation

The branch was finally cleanly rebased onto `origin/main` at `d215d9636` on
2026-08-06. The final-base timing target passed at 121.617s under the 4m15s
budget; the complete CLI package passed in 364.625s. `go vet ./...` and
`make lint` passed again on that final base; the individual `tidy-check`,
`docs-check`, `smoke-no-build`, `agent-contract-check`,
`connectorgen-validate`, `connectorgen-surface-sync`, `connector-boundary`,
and `release-workflow-check` gates passed during the immediately preceding
clean rebase validation. A fresh hosted Verify run is still required for the
final rebased commit range.

## Deliberate exclusions

No provider/credential/runtime checks are run. Aggregate `go test ./...` and
`make verify` remain CI gates because repository guidance documents that their
single-command duration exceeds the agent timeout. No CLI documentation parity
change is required because production CLI surface remains unchanged.
