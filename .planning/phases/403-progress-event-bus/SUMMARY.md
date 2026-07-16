# SUMMARY — Issue 403 progress event bus

Status: implementation pushed; verification blocked on repeated focused race timeout in existing `internal/connectors/certify` source-stage tests.

## Delivered

- New dependency-free `internal/events` package:
  - typed `Event`, `Counters`, `Kind`, `Scope`
  - `Emitter` via context with `Nop` default
  - sinks: `Nop`, NDJSON writer sink, bounded/coalescing `Chan`, `Throttle`, `Multi`, and in-memory `Collector`
  - NDJSON sanitizes terminal controls and redacts secret-like values through `internal/safety`
- Instrumentation:
  - `internal/flow/engine.go` emits flow/step lifecycle events
  - `internal/app/app.go` and `internal/app/local_warehouse.go` emit ETL lifecycle and flush progress
  - `internal/connectors/certify/batch.go` emits batch and connector worker-pool lifecycle
  - `internal/worker/submit.go` emits Temporal workflow submit/poll/done/cancel events through a testable client seam
- Tests added beside packages for event sinks, deterministic sequences, race-clean concurrent emit, and worker cancellation without external Temporal service.

## Passing evidence

- `go test -race ./internal/events/... -count=1`
- `go test -race ./internal/flow/... -run 'TestEngineEmits' -count=1`
- `go test -race ./internal/app/... -run 'TestRunETLEmits|TestRunWarehouseETLEmits' -count=1`
- `go test -race ./internal/connectors/certify/... -run TestRunBatchEmits -count=1`
- `go test -race ./internal/worker/... -run TestSubmitterEmits -count=1`
- `go list -deps -f '{{if not .Standard}}{{.ImportPath}}{{end}}' ./internal/events | grep -v '^$'` output only `polymetrics.ai/internal/safety` and `polymetrics.ai/internal/events`.
- `git diff -- go.mod go.sum` empty.

## Blocker

- Exact focused package race gate failed: `go test -race ./internal/flow/... ./internal/app/... ./internal/connectors/certify/... ./internal/worker/... -count=1` hit Go test default `panic: test timed out after 10m0s` in `internal/connectors/certify` after flow/app passed.
- Supplemental longer run also failed: `go test -race -timeout 30m ./internal/flow/... ./internal/app/... ./internal/connectors/certify/... ./internal/worker/... -count=1` timed out in `internal/connectors/certify` source-stage tests after flow/app passed.

## Review / PR

- Branch pushed.
- Sub-PR not opened because local verification gate is blocked.
- Claude disabled / Copilot quota exhausted per task; no review requests made.
