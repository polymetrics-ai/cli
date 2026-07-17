# VERIFICATION — Issue 403 progress event bus

## Checklist

Final PR-head finalization gates:

- [x] External strict race on production head `f16207974cc25f6df111fcd2a99c6acec41f3c44`: `go test -race ./... -count=1 -timeout 120m` passed.
- [x] `gofmt -w cmd internal` left no `cmd/**` or `internal/**` diff.
- [x] `go vet ./...`
- [x] `go test ./internal/events/... ./internal/flow/... ./internal/app/... ./internal/connectors/certify/... ./internal/worker/... -count=1`
- [x] `go build ./cmd/pm`
- [x] `make verify`
- [x] `git diff --check origin/feat/cli-architecture-v2...HEAD`
- [x] `git diff -- go.mod go.sum` empty
- [x] No production/dependency diff after final gates.
- [x] `verificationPassed=true` set only after the external strict race plus final local gates passed.

Second targeted review-fix requested local gates:

- [x] `gofmt -w internal/events`
- [x] `go test -race ./internal/events/... -count=1`
- [x] `go vet ./...`
- [x] `go test ./internal/events/... ./internal/flow/... ./internal/app/... ./internal/connectors/certify/... ./internal/worker/... -count=1`
- [x] `go build ./cmd/pm`
- [x] `make verify`
- [x] `git diff --check origin/feat/cli-architecture-v2...HEAD`
- [x] `git diff -- go.mod go.sum` empty
- [x] Strict full-race superseded by final external PR-head pass at `f16207974cc25f6df111fcd2a99c6acec41f3c44`.

Review-fix requested local gates (previous pass at `c9813a788d2bc0ccc29e79920ce6e5e8084e8a8e`):

- [x] `gofmt -w cmd internal`
- [x] `go test -race ./internal/events/... -count=1`
- [x] `go test -race ./internal/flow/... -run 'Test.*Emits' -count=1`
- [x] `go test -race ./internal/app/... -run 'Test.*Emits' -count=1`
- [x] `go test -race ./internal/connectors/certify/... -run 'TestRunBatchEmits' -count=1`
- [x] `go test -race ./internal/worker/... -run 'TestSubmitterEmits' -count=1`
- [x] `go vet ./...`
- [x] `go test ./internal/events/... ./internal/flow/... ./internal/app/... ./internal/connectors/certify/... ./internal/worker/... -count=1`
- [x] `go build ./cmd/pm`
- [x] `git diff --check origin/feat/cli-architecture-v2...HEAD`
- [x] `git diff -- go.mod go.sum` empty
- [x] dependency inspection confirms `internal/events` imports only stdlib + `internal/safety`
- [x] CLI parity marked N/A: no CLI surface, no `--progress` flag in this issue.
- [x] `make verify` (extra feasible gate) passed.

Strict full-race status:

- [x] `go test -race ./... -count=1 -timeout 120m` — external coordinator PR-head run passed on production head `f16207974cc25f6df111fcd2a99c6acec41f3c44` after rebase onto parent `f12d573b6415aed2c47cb3fd346c564d3b752a60`.

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

Prior external evidence at `2c2c16f850484ff5c4c8b99d065f4ef3361dbc61` remains invalidated by accepted review-fix production changes; the valid strict race source is now `f16207974cc25f6df111fcd2a99c6acec41f3c44`. Any later commit is artifacts-only and not a production self-SHA chase.

## Command log

| Command | Result | Notes |
|---|---|---|
| `scripts/gsd doctor` | pass | Adapter health OK. |
| `scripts/gsd prompt plan-phase 403 --skip-research` | pass | Prompt generated. |
| `scripts/gsd prompt programming-loop init --phase 403 --dry-run` | fail | Adapter gap: `scripts/gsd: unknown GSD command: programming-loop`; using `.pi/prompts/pm-gsd-loop.md` fallback inline. |
| `scripts/gsd doctor` | pass | Review-fix rerun 2026-07-17. |
| `scripts/gsd list` | pass | Review-fix rerun; 69 commands. |
| `scripts/gsd prompt plan-phase 403 --skip-research` | pass | Review-fix rerun; prompt generated to `/tmp/gsd-plan-403-reviewfix.txt`. |
| `scripts/gsd prompt programming-loop init --phase 403 --dry-run` | fail | Review-fix rerun; adapter gap remains `scripts/gsd: unknown GSD command: programming-loop`; manual inline loop continues. |
| `scripts/gsd doctor` | pass | Second targeted review-fix rerun 2026-07-17. |
| `scripts/gsd list` | pass | Second targeted review-fix rerun; 69 commands. |
| `scripts/gsd prompt plan-phase 403 --skip-research` | pass | Second targeted review-fix rerun; prompt generated to `/tmp/gsd-plan-403-reviewfix2.txt`. |
| `scripts/gsd prompt programming-loop init --phase 403 --dry-run` | fail | Second targeted review-fix rerun; adapter gap remains `scripts/gsd: unknown GSD command: programming-loop`; manual inline loop continues. |
| `scripts/gsd doctor` | pass | Finalization rerun 2026-07-17; adapter health OK. |
| `scripts/gsd list` | pass | Finalization rerun; 69 commands. |
| `scripts/gsd prompt plan-phase 403 --skip-research` | pass | Finalization rerun; prompt generated to `/tmp/gsd-plan-403-finalize.txt`. |
| `scripts/gsd prompt programming-loop init --phase 403 --dry-run` | fail | Finalization rerun; adapter gap remains `scripts/gsd: unknown GSD command: programming-loop`; manual inline loop continues. |
| `go test -race ./... -count=1 -timeout 120m` | pass | External coordinator PR-head source at production head `f16207974cc25f6df111fcd2a99c6acec41f3c44`: `internal/cli 1842.794s`; `internal/connectors/certify 3802.054s`; `internal/events 2.665s`; `internal/flow 2.590s`; `internal/worker 1.611s`; `real 3809.60`; `user 6223.47`; `sys 77.80`. |
| `gofmt -w cmd internal` | pass | Final PR-head finalization: no `cmd/**` or `internal/**` diff after command. |
| `go vet ./...` | pass | Final PR-head finalization: no output. |
| `go test ./internal/events/... ./internal/flow/... ./internal/app/... ./internal/connectors/certify/... ./internal/worker/... -count=1` | pass | Final PR-head finalization: `events 0.604s`; `flow 0.793s`; `app 17.306s`; `certify 335.483s`; `worker 0.534s`. |
| `go build ./cmd/pm` | pass | Final PR-head finalization: no output. |
| `make verify` | pass | Final PR-head finalization: `go test -timeout 20m ./...` passed; `internal/cli 163.819s`; `internal/connectors/certify 336.906s`; `smoke ok`; `0 issues`; `connectorgen validate: 547 connector(s) checked, 0 findings`. |
| `git diff --check origin/feat/cli-architecture-v2...HEAD` | pass | Final PR-head finalization: no output. |
| `git diff -- go.mod go.sum` | pass | Final PR-head finalization: no output; no dependency delta. |
| production/dependency diff check | pass | `git diff --name-only -- cmd internal go.mod go.sum` empty after final gates. |
| `go test ./internal/events/... -run TestChanCloseWaitsForInFlightEventAccounting -count=1` | fail | Second review-fix red: `DropStats() after Close = {Progress:1 Lifecycle:0}, want {Progress:1 Lifecycle:1}` at `events_test.go:194`; proves `Close` returned before runner accounted in-flight lifecycle drop. |
| `gofmt -w internal/events` | pass | Second review-fix final: no output. |
| `go test ./internal/events/... -run TestChanCloseWaitsForInFlightEventAccounting -count=1` | pass | Second review-fix green: `ok   polymetrics.ai/internal/events 0.430s`. |
| `go test -race ./internal/events/... -count=1` | pass | Second review-fix final: `ok   polymetrics.ai/internal/events 1.279s`. |
| `go vet ./...` | pass | Second review-fix final: no output. |
| `go test ./internal/events/... ./internal/flow/... ./internal/app/... ./internal/connectors/certify/... ./internal/worker/... -count=1` | pass | `events 0.272s`; `flow 0.367s`; `app 17.145s`; `certify 339.825s`; `worker 0.535s`. |
| `go build ./cmd/pm` | pass | Second review-fix final: no output. |
| `make verify` | pass | Second review-fix final: `go test -timeout 20m ./...` passed; `internal/events 4.218s`; `smoke ok`; `0 issues`; `connectorgen validate: 547 connector(s) checked, 0 findings`. |
| `git fetch origin feat/cli-architecture-v2` | pass | Updated `origin/feat/cli-architecture-v2` from `1678f9ab` to `e6faecfb` for diff-check base. |
| `git diff --check origin/feat/cli-architecture-v2...HEAD` | pass | no output. |
| `git diff -- go.mod go.sum` | pass | no output; no dependency delta. |
| `go test ./internal/events/... -run 'TestChan\|TestThrottle' -count=1` | fail | Review-fix red: build failed because `sink.DropStats` was undefined at `internal/events/events_test.go:108:19`, `:132:19`, `:146:16`. |
| `go test ./internal/events/... -run 'TestChan\|TestThrottle' -count=1` | pass | Review-fix green after Chan/Throttle fix: `ok   polymetrics.ai/internal/events 0.385s`. |
| `go test -race ./internal/events/... -count=1` | pass | Review-fix final: `ok   polymetrics.ai/internal/events 1.388s`. Prior pass was `1.178s`. |
| `go test -race ./internal/flow/... -run 'Test.*Emits' -count=1` | pass | Review-fix final: `ok   polymetrics.ai/internal/flow 1.291s`. Prior narrower pass was `1.437s`. |
| `go test -race ./internal/app/... -run 'Test.*Emits' -count=1` | pass | Review-fix final: `ok   polymetrics.ai/internal/app 29.128s`. Prior narrower pass was `18.027s`. |
| `go test -race ./internal/connectors/certify/... -run 'TestRunBatchEmits' -count=1` | pass | Review-fix final: `ok   polymetrics.ai/internal/connectors/certify 1.639s`. Prior pass was `1.632s`. |
| `go test -race ./internal/worker/... -run 'TestSubmitterEmits' -count=1` | pass | Review-fix final: `ok   polymetrics.ai/internal/worker 1.331s`. Prior pass was `1.351s`. |
| `go test ./internal/worker/... -count=1` | pass | `ok   polymetrics.ai/internal/worker 0.541s` |
| `go test ./internal/flow/... -count=1` | pass | `ok   polymetrics.ai/internal/flow 0.401s` |
| `go test ./internal/app/... -run 'TestRunETLEmits|TestRunWarehouseETLEmits|TestRunETLWritesBoundedBatches' -count=1` | pass | `ok   polymetrics.ai/internal/app 2.989s` |
| `go test ./internal/connectors/certify/... -run TestRunBatchEmits -count=1` | pass | `ok   polymetrics.ai/internal/connectors/certify 0.530s` |
| `go list -deps -f '{{if not .Standard}}{{.ImportPath}}{{end}}' ./internal/events \| grep -v '^$'` | pass | Output only `polymetrics.ai/internal/safety` and `polymetrics.ai/internal/events`. |
| `go test -race ./internal/flow/... ./internal/app/... ./internal/connectors/certify/... ./internal/worker/... -count=1` | fail | Go test default timeout: `panic: test timed out after 10m0s` in `internal/connectors/certify` after flow/app passed. |
| `go test -race -timeout 30m ./internal/flow/... ./internal/app/... ./internal/connectors/certify/... ./internal/worker/... -count=1` | fail | `internal/connectors/certify` timed out after 30m in existing source-stage tests after flow/app passed. |
| `go test -race ./... -count=1 -timeout 120m` | invalidated / superseded | External PR-head source at `2c2c16f850484ff5c4c8b99d065f4ef3361dbc61` previously passed, but accepted production review-fix changes invalidate that evidence for `e5404809fc66296f6d02e243b09b431dade921fb` and later. Superseded by final external pass at `f16207974cc25f6df111fcd2a99c6acec41f3c44`. |
| `gofmt -w cmd internal` | pass | Review-fix final: no output. |
| `go vet ./...` | pass | Review-fix final: no output. |
| `go test ./internal/events/... ./internal/flow/... ./internal/app/... ./internal/connectors/certify/... ./internal/worker/... -count=1` | pass | `events 0.488s`; `flow 0.506s`; `app 18.817s`; `certify 340.499s`; `worker 0.439s`. |
| `go build ./cmd/pm` | pass | Review-fix final: no output. |
| `make verify` | pass | Review-fix feasible gate: `go test -timeout 20m ./...` passed; `internal/cli 160.273s`; `internal/connectors/certify 339.605s`; `internal/events 2.131s`; `smoke ok`; `0 issues`; `connectorgen validate: 547 connector(s) checked, 0 findings`. |
| `git diff --check origin/feat/cli-architecture-v2...HEAD` | pass | no output after fetching current `origin/feat/cli-architecture-v2`; `real 0.02`. |
| `git diff -- go.mod go.sum` | pass | no output; no dependency delta; final rerun `real 0.01`. |

## Gate notes

Runtime-backed checks/services: not run; issue explicitly forbids services/credentials/external writes. Worker polling tests must use local test seam only.

Automated review route for this review-fix: no Claude/Copilot per user instruction. Human/parent fallback pending.

Current verification disposition: strict full-race passed externally on production head `f16207974cc25f6df111fcd2a99c6acec41f3c44`; requested local final gates passed; `RUN-STATE.json` records `verificationPassed=true`. Any commit after `f16207974cc25f6df111fcd2a99c6acec41f3c44` is artifacts-only.
