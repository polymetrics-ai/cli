# VERIFICATION — Issue 403 progress event bus

## Checklist

- [x] `gofmt -w cmd internal`
- [x] `go test -race ./internal/events/... -count=1`
- [ ] `go test -race ./internal/flow/... ./internal/app/... ./internal/connectors/certify/... ./internal/worker/... -count=1` — blocked by certify package timeout
- [ ] `go test -race ./...` — not run after repeated focused race timeout
- [ ] `go vet ./...` — not run after repeated focused race timeout
- [ ] `go test ./...` — not run after repeated focused race timeout
- [ ] `go build ./cmd/pm` — not run after repeated focused race timeout
- [ ] `make verify` — not run after repeated focused race timeout
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
| `git diff --check origin/feat/cli-architecture-v2...HEAD && git diff -- go.mod go.sum` | pass | Post-commit rerun clean; go.mod/go.sum diff empty. |

## Gate notes

Runtime-backed checks/services: not run; issue explicitly forbids services/credentials/external writes. Worker polling tests must use local test seam only.

Automated review route: Claude disabled / Copilot quota exhausted per task; no review requests. Human/parent fallback pending after PR open.
