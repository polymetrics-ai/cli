# SUMMARY — Issue 403 progress event bus

Status: final verification passed. Coordinator-provided strict full race gate passed at branch head `2c2c16f850484ff5c4c8b99d065f4ef3361dbc61`; remaining non-race final gates passed in this worker run.

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
- `go vet ./...`
- `go build ./cmd/pm`

## Strict race resolution

- Coordinator external gate at branch head `2c2c16f850484ff5c4c8b99d065f4ef3361dbc61`:

```text
go test -race ./... -count=1 -timeout 120m
PASS
internal/cli 1841.988s
internal/connectors/certify 3892.688s
internal/events 1.317s
real 3898.97
user 6294.91
sys 84.56
```

- Baseline/worker suspect certify tests had nearly identical pass times, confirming prior 10m/30m timeouts were suite duration, not event regression.

## Final verification

- `gofmt -w cmd internal` — pass, no production file diff.
- `go vet ./...` — pass, no output.
- `go test ./...` — pass, `internal/cli 163.985s`, `internal/connectors/certify 343.668s`, `internal/events 2.383s`, `real 348.54`.
- `go build ./cmd/pm` — pass, no output.
- `make verify` — pass, `smoke ok`, `0 issues`, `connectorgen validate: 547 connector(s) checked, 0 findings`, `real 367.33`.
- `git diff --check origin/feat/cli-architecture-v2...HEAD` — pass, no output.
- `git diff -- go.mod go.sum` — pass, no output; no dependency delta.
- CLI parity: N/A; no CLI command/flag/help/docs/website surface changed. `--progress` belongs to #405.
- PR will be opened non-draft to current `feat/cli-architecture-v2`. No Claude/Copilot request per task.
