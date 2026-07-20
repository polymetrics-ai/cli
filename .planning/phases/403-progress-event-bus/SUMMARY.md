# SUMMARY — Issue 403 progress event bus

Status: PR #451 finalization complete. Coordinator strict full-race passed on production head `f16207974cc25f6df111fcd2a99c6acec41f3c44`; requested local gates passed; `verificationPassed=true` is valid after final gates. Later commit(s) are artifacts-only.

## Second targeted review findings accepted and fixed

- MEDIUM `Chan.Close` in-flight acknowledgment:
  - added regression forcing one lifecycle event removed from the queue and blocked on `Events()` with no consumer, plus one queued progress event;
  - red: `DropStats() after Close = {Progress:1 Lifecycle:0}, want {Progress:1 Lifecycle:1}`;
  - fixed with runner `stopped` acknowledgment: `Close` closes `done`, waits for runner to account in-flight drop, close `Events()`, then return;
  - green asserts `Close` finite, `Events()` closed immediately after `Close`, exact `DropStats{Progress:1, Lifecycle:1}`.
- LOW `Multi` contract correction:
  - `Multi` remains synchronous and does not make arbitrary custom/writer sinks finite;
  - comments/tests/artifacts/PR claim narrowed to bounded `Chan` sinks; blocking custom sinks must be finite or honor context cancellation;
  - no goroutine-per-sink fanout added.

## Delivered earlier in issue #403

- Dependency-free `internal/events` package with typed events, context emitter, Nop, NDJSON, Chan, Throttle, Multi, Collector.
- Instrumentation in flow, ETL flush, certify batch, and worker polling paths.
- Secret/terminal sanitization for NDJSON via `internal/safety`.

## Verification

Strict full-race passed externally on production head `f16207974cc25f6df111fcd2a99c6acec41f3c44`:

```text
go test -race ./... -count=1 -timeout 120m
PASS
internal/cli 1842.794s
internal/connectors/certify 3802.054s
internal/events 2.665s
internal/flow 2.590s
internal/worker 1.611s
real 3809.60
user 6223.47
sys 77.80
```

Final local gates passed after the strict race:

- `gofmt -w cmd internal` — pass; no `cmd/**` or `internal/**` diff
- `go vet ./...` — pass, no output
- `go test ./internal/events/... ./internal/flow/... ./internal/app/... ./internal/connectors/certify/... ./internal/worker/... -count=1` — pass (`events 0.604s`; `flow 0.793s`; `app 17.306s`; `certify 335.483s`; `worker 0.534s`)
- `go build ./cmd/pm` — pass, no output
- `make verify` — pass; `internal/cli 163.819s`; `internal/connectors/certify 336.906s`; `smoke ok`; `0 issues`; `connectorgen validate: 547 connector(s) checked, 0 findings`
- `git diff --check origin/feat/cli-architecture-v2...HEAD` — pass, no output
- `git diff -- go.mod go.sum` — pass, no output
- production/dependency diff check — pass; no `cmd/**`, `internal/**`, `go.mod`, or `go.sum` diff after final gates

## Pending

- CLI parity: N/A, no CLI command/flag/help/docs/website surface changed; `--progress ndjson` remains #405.
- No Claude/Copilot requested per finalization instruction.
- Parent PR merge to `main` remains human-gated.
