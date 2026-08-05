# TDD ledger — certify-harness cost foundation (#3795)

| Slice | RED evidence required before production change | GREEN evidence | REFACTOR guard |
| --- | --- | --- | --- |
| #3798 invocation contract | A test-only counter at the real harness `cli.Run` seam proves the pre-refactor suite exceeds the named ceiling. The test must not use wall time. | The deterministic count drops only after duplicated execution is replaced. | Counter resets deterministically and does not depend on package/test order. |
| #3801 scripted stages | Scripted-driver tests initially fail for a command argument/envelope/exit/stage mismatch. | Migrated stage/error tests assert exact protocol, outcome, leak, idempotency, no-retry, and approval-replay-negative behavior. | Share transparent fixture/driver construction only after a family is green. |
| #3805 CLI fixtures | Fixture-render parity fails before a full route launch is removed. | JSON/text, exit, persistence, resume/batch, ordering/worker-pool, and usage tests pass from complete synthetic reports; exactly one real route proof remains. | Fixture completeness remains explicit rather than relying on zero values. |
| #3806 timing visibility | Parser tests fail on pass/fail/malformed/missing `go test -json` event streams; Make/workflow assertion fails before wiring. | Cold timing target prints raw events, package totals, slow tests, and fails on target/parser errors. | Formatting is deterministic and no source event is suppressed. |
| #3807 measured budget | Controlled duplicate fixture fails invocation and duration guards without sleeps. | Fixed threshold derived from documented GitHub-hosted cold samples passes and diagnostics name observed/allowed time. | Threshold/configuration lives in one place; deterministic invocation count remains the primary guard. |

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
removed. `TestFullSweepSourceStagesAgainstSample` remains the one real
full-source/read/flow/schedule proof.

```text
$ go test -count=1 ./internal/connectors/certify
ok      polymetrics.ai/internal/connectors/certify    58.972s
```

The TestMain output from the preceding verbose cold run measured 85 real CLI
invocations, so the committed deterministic ceiling is exactly 85 (the
pre-refactor red run was 782). The cap is a count, not a duration budget.

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
also proves report persistence. Text rendering, persistence round-trip, and
JSON/text batch output use complete Report/BatchReport fixtures; batch worker,
ordering, resume, and exit behavior remain covered by their existing focused
tests. The cold focused route run now makes 61 real calls and passes the exact
61-call ceiling:

```text
$ go test -count=1 -run '^TestCertifyCLI' ./internal/cli
ok      polymetrics.ai/internal/cli    43.128s
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
certify timing summary
  certify-harness package=polymetrics.ai/internal/connectors/certify elapsed=57.745s
    slow_test=TestFullSweepSourceStagesAgainstSample elapsed=49.160s
  certify-cli package=polymetrics.ai/internal/cli elapsed=42.359s
    slow_test=TestCertifyCLISingleConnectorPassExitsZero elapsed=27.150s
  total elapsed=100.104s
```

The local sample is diagnostic only; it cannot establish #3807's fixed
threshold. `.github/workflows/verify.yml` now runs `make certify-timing`
before the unchanged aggregate `make verify` step, so the next two
GitHub-hosted runs provide the required evidence.
