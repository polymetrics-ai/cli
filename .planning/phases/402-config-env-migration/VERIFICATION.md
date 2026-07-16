# Phase 402 Verification

## Required gate checklist

- [x] `gofmt -w cmd internal`
- [x] `go test ./internal/config/... -count=1`
- [x] `go test ./internal/runtimecheck/... -count=1`
- [x] `go test ./internal/schedule/... -count=1`
- [x] `go test ./internal/worker/... -count=1`
- [x] `go test ./internal/cli/ -run 'Golden|Config|Runtime|RLM|Schedule|Worker|AgentImage' -count=1`
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
- [x] `./pm help config` checked after docs caveat change.
- [x] `./pm runtime --help`, `./pm schedule --help`, `./pm rlm --help`, `./pm agent --help` checked; hidden `worker --help` remains pre-existing unavailable hidden-topic behavior.
- [x] `docs/cli/config.md` updated.
- [x] `website/content/docs/cli-reference.mdx` and generated `website/lib/docs.generated.ts` updated.
- [x] Bare namespace behavior unchanged for visible namespaces (`runtime`, `agent`, `rlm`, `schedule` exit 0).

## Results

- Focused packages: pass.
- CLI focused/golden/config migration: pass.
- Certify: pass.
- Full gates: `go vet ./...`, `go test ./...`, `go build ./cmd/pm`, and `make verify` pass.
- Docs/help parity: pass for changed config docs and visible affected namespaces; `worker --help` hidden-topic failure recorded as pre-existing hidden-command behavior.
- Runtime services/credentialed checks not run.
