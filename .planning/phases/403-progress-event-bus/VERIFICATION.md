# VERIFICATION — Issue 403 progress event bus

## Checklist

- [x] `gofmt -w cmd internal`
- [x] `go test -race ./internal/events/... -count=1`
- [x] `go test -race ./internal/flow/... ./internal/app/... ./internal/connectors/certify/... ./internal/worker/... -count=1` — superseded by strict full-race gate below; prior 10m/30m timeouts matched suite duration
- [x] `go test -race ./... -count=1 -timeout 120m` — external PR-head source on `2c2c16f850484ff5c4c8b99d065f4ef3361dbc61`
- [x] `go vet ./...`
- [x] `go test ./...`
- [x] `go build ./cmd/pm`
- [x] `make verify`
- [x] `git diff --check origin/feat/cli-architecture-v2...HEAD`
- [x] `git diff -- go.mod go.sum` empty
- [x] dependency inspection confirms `internal/events` imports only stdlib + `internal/safety`
- [x] CLI parity marked N/A: no CLI surface, no `--progress` flag in this issue.

## Command log

| Command | Result | Notes |
|---|---|---|
| `scripts/gsd doctor` | pass | Adapter health OK. |
| `scripts/gsd prompt plan-phase 403 --skip-research` | pass | Prompt generated. |
| `scripts/gsd prompt programming-loop init --phase 403 --dry-run` | fail | Adapter gap: `scripts/gsd: unknown GSD command: programming-loop`; using `.pi/prompts/pm-gsd-loop.md` fallback inline. |
| `go test -race ./internal/events/... -count=1` | pass | `ok   polymetrics.ai/internal/events 1.178s` |
| `go test -race ./internal/flow/... -run 'TestEngineEmits' -count=1` | pass | `ok   polymetrics.ai/internal/flow 1.437s` |
| `go test -race ./internal/app/... -run 'TestRunETLEmits|TestRunWarehouseETLEmits' -count=1` | pass | `ok   polymetrics.ai/internal/app 18.027s` |
| `go test -race ./internal/connectors/certify/... -run TestRunBatchEmits -count=1` | pass | `ok   polymetrics.ai/internal/connectors/certify 1.632s` |
| `go test -race ./internal/worker/... -run TestSubmitterEmits -count=1` | pass | `ok   polymetrics.ai/internal/worker 1.351s` |
| `go test ./internal/worker/... -count=1` | pass | `ok   polymetrics.ai/internal/worker 0.541s` |
| `go test ./internal/flow/... -count=1` | pass | `ok   polymetrics.ai/internal/flow 0.401s` |
| `go test ./internal/app/... -run 'TestRunETLEmits|TestRunWarehouseETLEmits|TestRunETLWritesBoundedBatches' -count=1` | pass | `ok   polymetrics.ai/internal/app 2.989s` |
| `go test ./internal/connectors/certify/... -run TestRunBatchEmits -count=1` | pass | `ok   polymetrics.ai/internal/connectors/certify 0.530s` |
| `go list -deps -f '{{if not .Standard}}{{.ImportPath}}{{end}}' ./internal/events \| grep -v '^$'` | pass | Output only `polymetrics.ai/internal/safety` and `polymetrics.ai/internal/events`. |
| `go test -race ./internal/flow/... ./internal/app/... ./internal/connectors/certify/... ./internal/worker/... -count=1` | fail | Go test default timeout: `panic: test timed out after 10m0s` in `internal/connectors/certify` after flow/app passed. |
| `go test -race -timeout 30m ./internal/flow/... ./internal/app/... ./internal/connectors/certify/... ./internal/worker/... -count=1` | fail | `internal/connectors/certify` timed out after 30m in existing source-stage tests after flow/app passed. |
| `go test -race ./... -count=1 -timeout 120m` | pass | External PR-head source at `2c2c16f850484ff5c4c8b99d065f4ef3361dbc61`: `PASS`; `internal/cli 1841.988s`; `internal/connectors/certify 3892.688s`; `internal/events 1.317s`; `real 3898.97`; `user 6294.91`; `sys 84.56`. Baseline/worker suspect certify tests had nearly identical pass times, confirming prior 10m/30m timeout was suite duration, not event regression. |
| `gofmt -w cmd internal` | pass | no output; `real 0.27`, `user 0.45`, `sys 0.32`; no production file diff. |
| `go vet ./...` | pass | no output; `real 2.19`, `user 1.22`, `sys 2.70`. |
| `go test ./...` | pass | `internal/cli 163.985s`; `internal/connectors/certify 343.668s`; `internal/events 2.383s`; `real 348.54`, `user 783.02`, `sys 47.85`. |
| `go build ./cmd/pm` | pass | no output; `real 1.14`, `user 0.56`, `sys 1.21`. |
| `make verify` | pass | includes fmt, tidy-check, vet, `go test -timeout 20m ./...`, build, docs validate, smoke, lint, connectorgen validate. `smoke ok`; `0 issues`; `connectorgen validate: 547 connector(s) checked, 0 findings`; `real 367.33`, `user 787.62`, `sys 68.30`. |
| `git diff --check origin/feat/cli-architecture-v2...HEAD` | pass | no output after fetching current `origin/feat/cli-architecture-v2`; `real 0.02`. |
| `git diff -- go.mod go.sum` | pass | no output; no dependency delta; final rerun `real 0.01`. |

## Gate notes

Runtime-backed checks/services: not run; issue explicitly forbids services/credentials/external writes. Worker polling tests must use local test seam only.

Automated review route: Claude disabled / Copilot quota exhausted per task; no review requests. Human/parent fallback pending after PR open.
