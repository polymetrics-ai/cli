# SUMMARY — Issue 403 progress event bus

Status: PR #451 review-fix complete locally; strict full-race remains pending with parent orchestrator because accepted production fixes invalidate the prior full-race evidence.

## Review findings accepted and fixed

- HIGH Chan backpressure/close semantics:
  - internal queue remains within configured capacity;
  - lifecycle insertion evicts/coalesces progress first;
  - full-lifecycle stalls use bounded wait with context/close handling;
  - `DropStats` accounts progress and lifecycle drops;
  - `Close` is finite with explicit accounted close-drop semantics;
  - `Multi` is no longer indefinitely stalled by a full/stalled `Chan` sink.
- MEDIUM Throttle stale progress after terminal:
  - pending progress flushes before completed/failed/skipped terminal lifecycle events;
  - terminal lifecycle remains last.
- Residual sequence gaps:
  - added focused tests for flow skip, ETL failure terminal, certify skip/failure, and worker success paths.

## Delivered earlier in issue #403

- Dependency-free `internal/events` package with typed events, context emitter, Nop, NDJSON, Chan, Throttle, Multi, Collector.
- Instrumentation in flow, ETL flush, certify batch, and worker polling paths.
- Secret/terminal sanitization for NDJSON via `internal/safety`.

## Verification

Passed locally on review-fix head:

- `go test -race ./internal/events/... -count=1` — `ok ... 1.388s`
- `go test -race ./internal/flow/... -run 'Test.*Emits' -count=1` — `ok ... 1.291s`
- `go test -race ./internal/app/... -run 'Test.*Emits' -count=1` — `ok ... 29.128s`
- `go test -race ./internal/connectors/certify/... -run 'TestRunBatchEmits' -count=1` — `ok ... 1.639s`
- `go test -race ./internal/worker/... -run 'TestSubmitterEmits' -count=1` — `ok ... 1.331s`
- `go vet ./...` — pass, no output
- `go test ./internal/events/... ./internal/flow/... ./internal/app/... ./internal/connectors/certify/... ./internal/worker/... -count=1` — pass (`certify 340.499s`)
- `go build ./cmd/pm` — pass, no output
- `make verify` — pass; `smoke ok`; `0 issues`; `connectorgen validate: 547 connector(s) checked, 0 findings`
- `git diff --check origin/feat/cli-architecture-v2...HEAD` — pass, no output
- `git diff -- go.mod go.sum` — pass, no output

## Pending

- `go test -race ./... -count=1 -timeout 120m`: **pending parent orchestrator** on final production head. Prior pass at `2c2c16f850484ff5c4c8b99d065f4ef3361dbc61` is invalidated; do not claim it covers `e5404809fc66296f6d02e243b09b431dade921fb` or later.
- CLI parity: N/A, no CLI command/flag/help/docs/website surface changed; `--progress ndjson` remains #405.
- No Claude/Copilot requested per review-fix instruction.
