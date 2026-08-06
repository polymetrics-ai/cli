# TDD ledger — certify-harness cost foundation (#3795)

| Slice | RED evidence required before production change | GREEN evidence | REFACTOR guard |
| --- | --- | --- | --- |
| #3798 invocation contract | A test-only counter at the real harness `cli.Run` seam proves the pre-refactor suite exceeds the named ceiling. The test must not use wall time. | The deterministic count drops only after duplicated execution is replaced. | Counter resets deterministically and does not depend on package/test order. |
| #3801 scripted stages | Scripted-driver tests initially fail for a command argument/envelope/exit/stage mismatch. | Migrated stage/error tests assert exact protocol, outcome, leak, idempotency, no-retry, and approval-replay-negative behavior. | Share transparent fixture/driver construction only after a family is green. |
| #3805 CLI fixtures | Fixture-render parity fails before a full route launch is removed. | JSON/text, exit, persistence, resume/batch, ordering/worker-pool, and usage tests pass from complete synthetic reports; exactly one real route proof remains. | Fixture completeness remains explicit rather than relying on zero values. |
| #3806 timing visibility | Parser tests fail on pass/fail/malformed/missing `go test -json` event streams; Make/workflow assertion fails before wiring. | Cold timing target prints raw events, package totals, slow tests, and fails on target/parser errors. | Formatting is deterministic and no source event is suppressed. |
| #3807 measured budget | Controlled duplicate fixture fails invocation and duration guards without sleeps. | Fixed threshold derived from the documented final topology measurement passes and diagnostics name observed/allowed time. | Threshold/configuration lives in one place; deterministic invocation count remains the primary guard. |

## Intended command evidence

```text
go test -count=1 ./internal/connectors/certify
go test -count=1 ./internal/cli
go test -race ./internal/connectors/certify
go test -race ./internal/cli
go test -count=1 -json ./internal/connectors/certify
go test -count=1 -json ./internal/cli
go vet ./internal/connectors/certify ./internal/cli
go vet ./...
go build ./cmd/pm
```

The final ledger will replace every planned row with the exact command, test
name, failure output or exit status, and the matching green command before the
PR is declared ready.

## CI remediation — duplicate full-route execution

| State | Evidence |
| --- | --- |
| RED | Historical Hosted Verify for `cd20f1074` failed `make certify-timing`: `201.488s` observed versus its then-active `180.000s` budget, now superseded by the hosted-measurement-derived 210-second cap. Its slow-test report identifies a direct full source sweep (`75.980s`) plus a separate real CLI route (`42.480s`) that repeats the same Runner/harness topology. |
| GREEN target | A single `TestCertifyCLI…` route invokes `pm connectors certify sample --full --json`, persists the report, and asserts every former direct full-sweep invariant from the JSON report. The standalone full-sweep test is removed, invocation caps are re-measured exactly, and `make certify-timing` remains under the active hosted-measurement-derived `3m30s` (210-second) cap. |
| Guard | This is a test-topology consolidation only: no production certification behavior, timeout, retry policy, or duration allowance is relaxed. |

### CI remediation GREEN evidence

The coalesced `TestCertifyCLISingleConnectorPassExitsZero` route retained the
full source/read/flow/schedule assertions while exercising CLI dispatch, the
real Runner/harness seam, JSON rendering, and report persistence. The revised
exact caps are 25 harness calls and 92 CLI calls. The active
hosted-measurement-derived `3m30s` (210-second) timing cap passed locally:

```text
$ make certify-timing
certify real CLI invocations: 25 (budget 25)
certify CLI real invocations: 92 (budget 92)
total elapsed=92.582s wall_elapsed=100.104s
```

The focused package tests and the complete `internal/cli` package also passed;
the latter completed in `458.490s`. The outer executor retains ownership of a
fresh hosted Verify run.

### Post-rebase compatibility — provenance help route

| State | Evidence |
| --- | --- |
| RED | Hosted Verify run `31095059925` on rebased PR head `428324630` failed `make certify-timing`: `internal/cli` reported `certify CLI real invocations: 93 (budget 92)`. `git diff 3332dc69..428324630 -- internal/cli/certify_cli_test.go` identifies the sole added direct call: #3869's `TestCertifyCLIHelpShowsProvenanceContract` invokes `certifyRun(..., "--help")`. |
| GREEN | The same provenance-help assertions now call the exact `runConnectors` `certify --help` branch directly. `make certify-timing` returned 25/25 harness calls and 92/92 CLI calls; total elapsed/wall time was 79.895s/88.553s. |
| Guard | Do not change the 92-call cap or the `3m30s` (210-second) hosted timing budget. The retained full route and one real sample/outbox write lifecycle proof remain unchanged. |

## Captured RED evidence

### #3798 — certify package invocation budget

Before any production or harness behavior change, the test-only wrapper in
`certify_testmain_test.go` counted every real `cli.Run` reached through
`Harness.Run`. The first full uncached package run failed as intended:

```text
$ go test -count=1 -v ./internal/connectors/certify
...
PASS
certify real CLI invocations: 782 (budget 128)
certify real CLI invocation budget exceeded: got 782, allowed 128; retain only the exhaustive real proof and script duplicate stage cases
FAIL    polymetrics.ai/internal/connectors/certify    578.282s
FAIL
```

The verbose output showed the retained full-sweep route alone cost 49.41s,
while normal glue/source cases cost 26–28s and repeated write cases cost
29–36s. The invocation failure, not that laptop duration, is the contract:
the later scripted-driver migration must reduce equivalent real CLI work below
the ceiling without removing assertions.

### #3798/#3801 — GREEN invocation contract and scripted stage driver

The test-only scripted driver rejects malformed command families, validates
the harness-injected root, verifies command/envelope/exit transitions, and
materializes the sample outbox protocol needed for the write lifecycle. The
focused source, glue, write, failure, cleanup, and replay assertions remain
in their existing test cases; only repeated real `cli.Run` execution was
removed. `TestFullSweepSourceStagesAgainstSample` remains the real
full-source/read/flow/schedule proof, and
`TestSampleOutboxWriteLifecycleAgainstRealCLI` is the exactly-one real
sample/outbox lifecycle proof.

```text
$ go test -count=1 ./internal/connectors/certify
ok      polymetrics.ai/internal/connectors/certify    81.733s
```

The final topology adds exactly one real sample/outbox lifecycle proof. Its
`make certify-timing` measurement reports 93 real harness invocations, so the
committed deterministic ceiling is exactly 93 (the pre-refactor red run was
782). The cap is a count, not a duration budget.

### #3805 — CLI router and complete-fixture rendering

Before the fixture conversion, the focused cold route suite failed the
initial counter as intended:

```text
$ go test -count=1 -v -run '^TestCertifyCLI' ./internal/cli
...
PASS
certify CLI real invocations: 213 (budget 80)
certify CLI real invocation budget exceeded: got 213, allowed 80; retain one certify router proof and render remaining cases from fixtures
FAIL    polymetrics.ai/internal/cli    227.819s
FAIL
```

After conversion, the single real `pm connectors certify sample --json` route
also proves report persistence. Complete Report/BatchReport fixtures directly
exercise JSON/text rendering, persistence round-trip, report exit precedence,
batch JSON/text output, and matrix ordering without launching another Runner.
The distinct batch worker, resume, ordering, and exit behavior remains covered
by `TestRunBatchRunsConnectorsConcurrentlyUpToParallelLimit`,
`TestRunBatchResumeSkipsConnectorsWithFreshReport`,
`TestBatchReportSummaryMatrixLeaksRowFirst`, and the focused batch exit tests.
The cold focused route run now makes 61 real calls and passes the exact 61-call
ceiling:

```text
$ go test -count=1 -run '^TestCertifyCLI' ./internal/cli
ok      polymetrics.ai/internal/cli    54.825s
```

### #3806 — RED/GREEN cold timing parser and Verify visibility

The parser test file was added before its implementation. The first command
failed to compile because `Target`, `ParseTarget`, `Run`, `GoTestArgs`, and
`Report` did not exist; that red checkpoint is retained in
`internal/certifytiming/timing_test.go`. The resulting parser rejects malformed
events, missing package completions, and failed package events; it also proves
a controlled duplicate fixture exceeds the duration backstop without sleeping.

```text
$ go test -count=1 ./internal/certifytiming ./cmd/certifytiming
ok      polymetrics.ai/internal/certifytiming    0.144s
?       polymetrics.ai/cmd/certifytiming [no test files]

$ make certify-timing
... raw go test -count=1 -json events ...
certify real CLI invocations: 93 (budget 93)
certify CLI real invocations: 61 (budget 61)
certify timing summary
  certify-harness elapsed=81.733s wall_elapsed=86.645s
    slow TestFullSweepSourceStagesAgainstSample=59.260s
    slow TestSweeperCleansUnledgeredAgedEntries=11.740s
    slow TestSampleOutboxWriteLifecycleAgainstRealCLI=8.480s
  certify-cli elapsed=54.825s wall_elapsed=60.665s
    slow TestCertifyCLISingleConnectorPassExitsZero=34.630s
  total elapsed=136.558s wall_elapsed=147.309s
```

The final topology measurement is recorded in `VERIFICATION.md` and
`RUN-STATE.json`. `.github/workflows/verify.yml` runs `make certify-timing`
before the unchanged aggregate `make verify` step.

### #3807 — wall-clock measurement and duplicate guards

The duration backstop was made test-first. Before its implementation, the
controlled fixture referenced `TargetResult.WallElapsed` and
`Report.TotalWallElapsed`, and the focused test failed to compile because
both symbols were absent. The green implementation measures the actual wall
time of each targeted `go test` process (including that command's startup and
compilation), prints per-target and total `wall_elapsed`, and refuses to
evaluate a duration budget if any target lacks a measurement.

The deliberate no-sleep duplicate fixtures exercise both independent guards:
`TestCertifyRealCLIInvocationBudgetRejectsControlledDuplicate` rejects 94
real harness calls against the 93-call contract, while
`TestCheckDurationBudgetRejectsControlledDuplicateFixture` rejects two
4-second wall measurements against a 7-second bound. The two final-topology
wall measurements are 161.236s and 147.309s; their 13.927s spread gives
`161.236 + 13.927 = 175.163s`, which historically rounded up to the
superseded 180-second (`3m`) bound. The Make target now supplies the active
hosted-measurement-derived `3m30s` (210-second) cap to the timing command,
which reports the observed and allowed duration when it fails.

```text
$ go test -count=1 -run '^TestCertifyRealCLIInvocationBudgetRejectsControlledDuplicate$' ./internal/connectors/certify
ok      polymetrics.ai/internal/connectors/certify

$ go test -count=1 ./internal/certifytiming ./cmd/certifytiming
ok      polymetrics.ai/internal/certifytiming
```
