# Phase 402 Verification

## Required gate checklist

Review-fix gates rerun 2026-07-16 against new fixes.

- [x] `gofmt -w cmd internal`
- [x] `go test ./internal/config/... -count=1`
- [x] `go test ./internal/runtimecheck/... -count=1`
- [x] `go test ./internal/perf/... -count=1`
- [x] `go test ./internal/worker/... -count=1`
- [x] `go test ./internal/cli/ -run 'Golden|Config|Runtime|Perf|Worker' -count=1`
- [x] `go test ./internal/cli/ -run Certify -count=1`
- [x] `go vet ./...`
- [x] `go test ./...`
- [x] `go build ./cmd/pm`
- [x] `make verify`
- [x] `git diff --check origin/feat/cli-architecture-v2...HEAD`
- [x] `git diff origin/feat/cli-architecture-v2...HEAD -- go.mod go.sum` (no output)

## Optional / safety-limited

- [x] Runtime-backed integration tests not run; services were not started.
- [x] No credentialed checks.
- [x] No reverse ETL execution outside `make verify` local smoke flow required by project gate.
- [x] No new dependencies.

## CLI parity checklist

- [x] Golden transcripts unchanged (`go test ./internal/cli/ -run Golden -count=1` included in focused CLI gate).
- [x] `./pm help config` checked after docs caveat change in prior slice; review-fix has no intended help/docs behavior change.
- [x] `./pm runtime --help`, `./pm schedule --help`, `./pm rlm --help`, `./pm agent --help` checked in prior slice; hidden `worker --help` remains pre-existing unavailable hidden-topic behavior.
- [x] `docs/cli/config.md` updated in prior slice; no review-fix doc behavior change planned.
- [x] `website/content/docs/cli-reference.mdx` and generated `website/lib/docs.generated.ts` updated in prior slice; no review-fix doc behavior change planned.
- [x] Bare namespace behavior unchanged for visible namespaces (`runtime`, `agent`, `rlm`, `schedule`, `perf` exit 0).

## Results

- Focused packages: pass (`config`, `runtimecheck`, `perf`, `worker`).
- CLI focused/golden/config/runtime/perf/worker: pass.
- Certify: pass.
- Full gates: `go vet ./...`, `go test ./...`, `go build ./cmd/pm`, and `make verify` pass.
- Diff checks: `git diff --check origin/feat/cli-architecture-v2...HEAD` pass; `git diff origin/feat/cli-architecture-v2...HEAD -- go.mod go.sum` no output.
- Global config-reader search: no `pmconfig.Load(Options{})` or `runtimecheck.FromEnv()` remains in CLI/worker/perf path; retained hits are `runtimecheck.FromEnv` compatibility and legacy `schedule.SelectBackend` compatibility. Raw env hits remain documented secret/user/test/container env exclusions.
- Runtime services/credentialed checks not run.
