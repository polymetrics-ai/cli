# Phase 402 TDD Ledger

Issue: #402 — migrate config-shaped env reads to typed config.

## Skills loaded

`gsd-core`, `caveman`, `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-context`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-concurrency`, `golang-spf13-cobra`, `golang-spf13-viper`, `golang-documentation` (docs caveat only if changed).

## GSD command evidence

```bash
scripts/gsd doctor
scripts/gsd prompt plan-phase 402 --skip-research >/tmp/gsd-plan-phase-402.prompt
scripts/gsd prompt programming-loop init --phase 402 --dry-run >/tmp/gsd-programming-loop-402.prompt
```

Result:

- `doctor`: pass.
- `plan-phase`: prompt written.
- `programming-loop`: blocked by adapter registry (`scripts/gsd: unknown GSD command: programming-loop`); manual GSD fallback active.

## Red / green / refactor log

| Step | Kind | Command / test | Result | Notes |
|---:|---|---|---|---|
| 0 | Planning | Create PLAN/TDD-LEDGER/VERIFICATION/SUMMARY/RUN-STATE/PROMPTS | Green | Pre-production artifact checkpoint. |
| 1 | Red | `go test ./internal/runtimecheck/... -count=1` | Fail | `undefined: FromConfig`; proves runtimecheck needs typed config API. |
| 2 | Red | `go test ./internal/schedule/... -count=1` | Fail | Missing `BackendConfig`, `SelectBackendFromConfig`, `CrontabBackend.File`; proves schedule needs injected config seam. |
| 3 | Red | `go test ./internal/cli/ -run 'Config|Runtime|RLM|Schedule|Worker|AgentImage' -count=1` | Fail | Runtime/worker/RLM config-file tests fail against legacy env readers. |
| 4 | Green | `go test ./internal/config/... -count=1 && go test ./internal/runtimecheck/... -count=1 && go test ./internal/schedule/... -count=1 && go test ./internal/worker/... -count=1 && go test ./internal/rlm/... -count=1` | Pass | Added explicit config metadata, runtimecheck FromConfig, schedule backend config/file seam, worker submitter activity injection, RLM typed non-secret settings. |
| 5 | Green | `go test ./internal/cli/ -run 'Golden|Config|Runtime|RLM|Schedule|Worker|AgentImage' -count=1` | Pass | CLI config-file/runtime/worker/schedule/agent image/RLM coverage green; golden transcripts unchanged. |
| 6 | Green | `go test ./internal/cli/ -run Certify -count=1` | Pass | Certify crontab save/restore behavior preserved. |
| 7 | Refactor | docs/help parity | Pass | Updated config caveat in embedded docs, `docs/cli/config.md`, website source, and generated website data. |
| 8 | Gate | `go vet ./...`; `go build ./cmd/pm`; `go test ./...`; `make verify` | Pass | Full local gate green; no go.mod/go.sum diff. |

## Implemented green slice

- `internal/config`: added `ExplicitKeys`/`IsExplicit` metadata from changed bound flags, env aliases, and config-file keys to preserve opt-in Temporal semantics while still providing defaults to runtime doctor/runtime ETL.
- `internal/runtimecheck`: added `FromConfig(config.Config)` as primary API; compatibility `FromEnv` delegates through typed loader.
- `internal/cli`: resolved config once in `Run` and injects through Cobra legacy handlers into runtime, runtime ETL, worker, schedule, agent image, RLM, extract, and flow RLM paths.
- `internal/schedule`: added narrow `BackendConfig`, `SelectBackendFromConfig`, and `CrontabBackend.File`; legacy `SelectBackend` delegates through typed loader while keeping Temporal opt-in.
- `internal/worker`: added `SubmitterForActivities`/`NewPodmanActivities`; default compatibility path delegates through typed config.
- `internal/rlm`: added `LLMConfigFromSettings` for non-secret typed LLM settings; API keys remain env-only.
- Docs: config caveat updated because runtime/RLM/schedule keys are now active in migrated call sites.

## Planned red tests

- `internal/runtimecheck`: `FromConfig` maps `config.Config.Runtime`; `FromEnv` honors `PM_*` alias through typed loader.
- `internal/schedule`: typed backend config selects Temporal only when explicit addr provided; crontab file injection writes temp file without raw env dependency.
- `internal/cli`: runtime doctor config-file value appears in redacted output; worker status remains byte-compatible when temporal unset but accepts explicit config file; schedule install/remove honors config-file crontab path; agent image uses config-file podman bin; RLM fake runner works from config-file key.
- `internal/worker`: typed submitter activities use injected podman image/bin; cancellation tests remain green.
- Exclusion guard: credential `--from-env` and provider API-key env paths remain raw env only.

## Exact red outputs

### Runtimecheck

```bash
go test ./internal/runtimecheck/... -count=1
```

```text
# polymetrics.ai/internal/runtimecheck [polymetrics.ai/internal/runtimecheck.test]
internal/runtimecheck/runtimecheck_config_test.go:18:9: undefined: FromConfig
FAIL	polymetrics.ai/internal/runtimecheck [build failed]
FAIL
```

### Schedule

```bash
go test ./internal/schedule/... -count=1
```

```text
# polymetrics.ai/internal/schedule [polymetrics.ai/internal/schedule.test]
internal/schedule/config_test.go:12:9: undefined: BackendConfig
internal/schedule/config_test.go:17:13: undefined: SelectBackendFromConfig
internal/schedule/config_test.go:26:28: unknown field File in struct literal of type CrontabBackend
FAIL	polymetrics.ai/internal/schedule [build failed]
FAIL
```

### CLI config migration

```bash
go test ./internal/cli/ -run 'Config|Runtime|RLM|Schedule|Worker|AgentImage' -count=1
```

```text
--- FAIL: TestRuntimeDoctorUsesConfigFile (0.03s)
    config_migration_test.go:32: postgres_url = postgres://***@localhost:15433/polymetrics?sslmode=disable, want config-file value
--- FAIL: TestWorkerStatusUsesExplicitConfigFileTemporalAddr (0.00s)
    config_migration_test.go:54: addr = , want explicit config-file temporal addr
--- FAIL: TestRLMAgentFakeRunnerUsesConfigFile (0.00s)
    config_migration_test.go:79: exit = 1, stderr = error: rlm: remote agent backend unavailable (set POLYMETRICS_TEMPORAL_ADDR and install podman, or use --mode deterministic)
redis: 2026/07/16 21:23:12 pool.go:724: redis: connection pool: failed to dial after 5 attempts: dial tcp 127.0.0.1:1: connect: connection refused
redis: 2026/07/16 21:23:12 pool.go:724: redis: connection pool: failed to dial after 5 attempts: dial tcp 127.0.0.1:1: connect: connection refused
redis: 2026/07/16 21:23:13 pool.go:724: redis: connection pool: failed to dial after 5 attempts: dial tcp 127.0.0.1:1: connect: connection refused
redis: 2026/07/16 21:23:13 pool.go:724: redis: connection pool: failed to dial after 5 attempts: dial tcp 127.0.0.1:1: connect: connection refused
FAIL
FAIL	polymetrics.ai/internal/cli	3.817s
FAIL
```

## Exact green outputs

### Focused packages

```bash
go test ./internal/config/... -count=1 && go test ./internal/runtimecheck/... -count=1 && go test ./internal/schedule/... -count=1 && go test ./internal/worker/... -count=1 && go test ./internal/rlm/... -count=1
```

```text
ok  	polymetrics.ai/internal/config	0.397s
ok  	polymetrics.ai/internal/runtimecheck	0.366s
ok  	polymetrics.ai/internal/schedule	0.347s
ok  	polymetrics.ai/internal/worker	0.538s
ok  	polymetrics.ai/internal/rlm	0.672s
ok  	polymetrics.ai/internal/rlm/router	0.330s
```

### CLI focused/golden

```bash
go test ./internal/cli/ -run 'Golden|Config|Runtime|RLM|Schedule|Worker|AgentImage' -count=1
```

```text
ok  	polymetrics.ai/internal/cli	11.286s
```

### Certify crontab preservation

```bash
go test ./internal/cli/ -run Certify -count=1
```

```text
ok  	polymetrics.ai/internal/cli	92.245s
```

### Full gates

```bash
go vet ./...
go build ./cmd/pm
go test ./...
make verify
```

```text
go vet ./...: pass (no output)
go build ./cmd/pm: pass (no output)
go test ./...: pass
make verify: pass; smoke ok; golangci-lint 0 issues; connectorgen validate 547 connector(s) checked, 0 findings
```
