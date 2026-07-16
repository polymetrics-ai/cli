# Phase 402 Verification

## Required gate checklist

- [ ] `gofmt -w cmd internal`
- [ ] `go test ./internal/config/... -count=1`
- [ ] `go test ./internal/runtimecheck/... -count=1`
- [ ] `go test ./internal/schedule/... -count=1`
- [ ] `go test ./internal/worker/... -count=1`
- [ ] `go test ./internal/cli/ -run 'Golden|Config|Runtime|RLM|Schedule|Worker|AgentImage' -count=1`
- [ ] `go test ./internal/cli/ -run Certify -count=1`
- [ ] `go vet ./...`
- [ ] `go test ./...`
- [ ] `go build ./cmd/pm`
- [ ] `make verify`
- [ ] `git diff --check origin/feat/cli-architecture-v2...HEAD`
- [ ] `git diff origin/feat/cli-architecture-v2...HEAD -- go.mod go.sum`

## Optional / safety-limited

- [ ] Runtime-backed integration tests: not planned; do not start services. Run only if already available and safe.
- [ ] No credentialed checks.
- [ ] No reverse ETL execution.
- [ ] No new dependencies.

## CLI parity checklist

- [ ] Golden transcripts unchanged (`go test ./internal/cli/ -run Golden -count=1`).
- [ ] `pm help config` checked if config docs change.
- [ ] `pm runtime --help`, `pm worker --help`, `pm schedule --help`, `pm rlm --help`, `pm agent --help` unchanged or documented.
- [ ] `docs/cli/config.md` updated if caveat changed.
- [ ] `website/content/docs/cli-reference.mdx` and generated `website/lib/docs.generated.ts` updated if caveat changed.
- [ ] Bare namespace behavior unchanged.

## Results

Pending.
