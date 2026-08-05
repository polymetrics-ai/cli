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
