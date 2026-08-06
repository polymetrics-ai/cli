# Plan — certify-harness cost foundation (#3795)

## Workflow and skills

This plan follows the installed GSD lifecycle with an inline/manual fallback:
`discuss-phase` → `plan-phase --tdd` → `execute-phase` → `verify-work` →
`code-review`. The adapter prompt expects Pi role workers, while the canonical
single-worker contract and this task forbid spawning them; this worker executes
the generated workflow inline and records the evidence here.

Required skills loaded: `golang-how-to`, `golang-testing`, `golang-benchmark`,
`golang-performance`, `golang-continuous-integration`,
`golang-error-handling`, `golang-safety`, `golang-cli`, `golang-security`,
`golang-lint`, `golang-troubleshooting`, `github-issue-first-delivery`, and
`no-mistakes`.

## Verified premise

- `Makefile:26-27` gives Verify one serial `go test -timeout 20m ./...` gate.
- The previously observed Verify step consumed 21m31s of the 22m41s job
  (GitHub job `92398457040`, retrieved 2026-08-06), before the completion
  foundation in #3740 could be merged.
- `internal/connectors/certify/cliharness.go:36-40,82-112` executes real
  `cli.Run` for each harness command. The source stages rebuild roots and
  repeat source/read/full sweeps; tests currently duplicate those runs.
- Parent #3795 records 18 direct `Runner.Run` calls and five additional
  successful CLI route runs. This is the test-cost defect, not a production
  behavior regression or an interaction with #3739/#3743.

## Ordered slices and owned paths

### Slice 1 — #3798: invocation contract and cold timing protocol

Own `internal/connectors/certify/*_test.go`, including a test-only TestMain
helper if needed. Add a deterministic, order-independent counter at the real
`cli.Run` harness seam; explicitly list retained real proofs and a ceiling
which fails against today’s duplicate pattern. Add an evidence command for
fresh `go test -count=1 -json` package runs, without changing production code,
the report schema, workflow, or Makefile.

Checkpoint: planning seed; red invocation-budget test committed; run both
affected packages and their JSON forms.

### Slice 2 — #3801: scripted stage cases with retained real proof

Own certify tests plus only the minimal behavior-preserving harness seam. Write
driver-protocol red tests first, then migrate error/stage scenarios one family
at a time from repeated `Runner.Run` execution to a deterministic scripted
driver. Retain the exhaustive real proof of full source/read sweep with
flow/schedule. The scripted driver must preserve the destructive plan →
preview → approval → execute lifecycle, no-retry behavior, and
approval-replay-negative assertions; the sample connector has no
definition-owned full write inventory, so `--full --write` is a documented
non-passing combination rather than a second real proof.

Checkpoint: each migrated family has red/green evidence and the contract is
green because equivalent full runner work is no longer duplicated.

### Slice 3 — #3805: one CLI route proof and complete report fixtures

Own `internal/cli/certify_cli_test.go` and related certify CLI tests. Retain
one real `pm connectors certify sample --json` test which reaches CLI dispatch,
real runner, harness `cli.Run`, and persistence. Before replacing each route
launch, establish fixture-render parity from complete synthetic
`certify.Report`/`certify.BatchReport` values for JSON/text rendering, exit
mapping, persistence, resume/batch matrix, ordering, worker-pool, and usage
errors.

Checkpoint: one route proof remains; all migrated rendering and batch cases
remain executable against complete fixtures.

### Slice 4 — #3806: cold timing visibility in Verify

Own a narrow timing runner/parser and its focused Make/Verify wiring. Add parser
tests for pass/fail/malformed/missing Go test events. Invoke real cold package
tests (`-count=1 -json`), preserve raw events in the log, print per-package
elapsed time plus slow tests, and keep the aggregate test gate unchanged.

Checkpoint: run the target twice locally; push after targeted green tests so
GitHub-hosted cold samples are available in Verify logs.

### Slice 5 — #3807: measured two-layer budget

Own only the deterministic assertion and narrow Make/CI invocation needed to
enforce it. First prove a controlled duplicate fixture trips both parser and
invocation guards without sleeps. After the final topology includes exactly one
real sample/outbox lifecycle proof, record its cold timing measurement, derive
a margin from the observed process overhead, centralize the explicit threshold,
and make failure name observed and allowed values. Do not increase the
20-minute timeout, retry tests, skip tests, cache evidence, or move test work
out of Verify.

Checkpoint: measured threshold committed with the evidence, then all targeted
and scoped verification completed.

## TDD and verification

The detailed red/green/refactor ledger is in `TDD-LEDGER.md`. Verify the
changed packages with `-count=1`; run `-race` for certify and CLI as required.
Run `go vet ./...`, `go build ./cmd/pm`, and the individual non-suite Verify
gates. Do not run aggregate `go test ./...` or `make verify` as a single
per-command timeout-bound local call; CI carries those aggregate gates.

### CI remediation — coalesced full-route proof

Hosted Verify on commit `cd20f1074` measured `201.488s` against the then-active
historical `180s` duration guard, now superseded by the
hosted-measurement-derived `3m30s` (210-second) cap. The slow tests show that the direct
`TestFullSweepSourceStagesAgainstSample` proof and
`TestCertifyCLISingleConnectorPassExitsZero` each execute the same expensive
source certification topology. Move the complete full-sweep assertions into
the real `pm connectors certify sample --full --json` route proof and remove
the standalone duplicate. This retains CLI dispatch, the real Runner and
harness seam, report persistence, and all full source/read/flow/schedule
assertions in one executable invocation. Re-measure the exact invocation caps
and preserve the wall-clock evidence after the move without recalibrating the
active hosted cap.
The local remediation measurement held 25/25 harness calls and 92/92 CLI
calls, with `100.104s` total wall time below the active 210-second limit; the
original 180-second calculation remains superseded historical evidence.

### Post-rebase compatibility repair

Current `main` adds the #3869 provenance-help assertion as
`TestCertifyCLIHelpShowsProvenanceContract`. Its direct `certifyRun(...,
"--help")` call is the single new real CLI invocation observed after rebasing:
hosted Verify run `31095059925` reports 93 calls against the measured 92-call
contract. Keep the same help-text assertion, but exercise the exact
`runConnectors(..., []string{"certify", "--help"}, ...)` router branch
directly. This removes duplicate `Run` wrapper work without removing help
coverage, retains the one full route proof, and does not recalibrate the
measured budget.

## Delivery record

Parent issue: #3795. Sub-issues: #3798, #3801, #3805, #3806, #3807. Parent
branch: `fm/cli-found-certify-timeout-r1`. Draft PR must target `main` before
production edits. The final PR body will reference the parent and all
sub-issues, retain GSD/skill/TDD evidence, list CI timing samples, and record
automated-review routing.
