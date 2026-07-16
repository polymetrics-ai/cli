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
| 4 | Green | pending | pending | Implement minimal config injection. |
| 3 | Refactor | pending | pending | Docs caveat and cleanup while tests green. |

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
