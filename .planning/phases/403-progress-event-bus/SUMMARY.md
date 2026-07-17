# SUMMARY — Issue 403 progress event bus

Status: PR #451 second targeted review-fix complete locally; strict full-race remains pending with parent orchestrator because accepted production fixes invalidate prior full-race evidence.

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

Passed locally on second review-fix head:

- `gofmt -w internal/events` — pass, no output
- `go test -race ./internal/events/... -count=1` — `ok ... 1.279s`
- `go vet ./...` — pass, no output
- `go test ./internal/events/... ./internal/flow/... ./internal/app/... ./internal/connectors/certify/... ./internal/worker/... -count=1` — pass (`certify 339.825s`)
- `go build ./cmd/pm` — pass, no output
- `make verify` — pass; `smoke ok`; `0 issues`; `connectorgen validate: 547 connector(s) checked, 0 findings`
- `git diff --check origin/feat/cli-architecture-v2...HEAD` — pass, no output
- `git diff -- go.mod go.sum` — pass, no output

## Pending

- `go test -race ./... -count=1 -timeout 120m`: **pending parent orchestrator** on final production head. Prior pass at `2c2c16f850484ff5c4c8b99d065f4ef3361dbc61` is invalidated; do not claim it covers `c9813a788d2bc0ccc29e79920ce6e5e8084e8a8e` or later.
- CLI parity: N/A, no CLI command/flag/help/docs/website surface changed; `--progress ndjson` remains #405.
- No Claude/Copilot requested per review-fix instruction.
