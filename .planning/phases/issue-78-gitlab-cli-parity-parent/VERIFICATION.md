# Verification: GitLab CLI Parity Parent Orchestration

## Completed Preflight

```bash
scripts/gsd doctor
scripts/gsd verify-pi
scripts/gsd list --json
scripts/gsd prompt plan-phase issue-78-gitlab-cli-parity --skip-research
```

Result: adapter health and prompt generation succeeded. `scripts/gsd prompt programming-loop ...` is unavailable in this registry; manual GSD fallback is recorded.

## Parent Setup Checks

```bash
jq empty .planning/phases/issue-78-gitlab-cli-parity-parent/RUN-STATE.json \
  .planning/phases/issue-78-gitlab-cli-parity-parent/ORCHESTRATION-STATE.json
git diff --check
```

Result: passed before parent PR seed commit.

## #83 Integrated Local Checks

```bash
go test ./internal/connectors/engine -run TestBundleLoadEmbeddedGitLabCLISurface -count=1
go test ./cmd/connectorgen ./internal/connectors/engine -run 'CLISurface|GitLab' -count=1
go test ./cmd/connectorgen ./internal/connectors/engine ./internal/connectors/commandrunner -count=1
go test ./internal/connectors/conformance -run 'TestConformance/gitlab' -count=1
go run ./cmd/connectorgen validate internal/connectors/defs --json
gofmt -w cmd internal
go vet ./...
go test ./...
go build ./cmd/pm
go run ./cmd/connectorgen validate internal/connectors/defs
make verify
cd website && pnpm run gen:website-data
```

Result: passed. Website typecheck/build remain blocked locally because `node_modules` is absent and
`pnpm install --frozen-lockfile` fails with `ERR_PNPM_LOCKFILE_CONFIG_MISMATCH`.

## #84-#89 Focused Local Checks

```bash
go test ./internal/cli -run 'TestGitLabCommandSurfaceHelp' -count=1
go test ./internal/cli -run 'TestGitLabCommandSurfaceRunsStreamBackedIssueList' -count=1
go test ./internal/connectors/engine -run TestBundleLoadGitLabOperationLedgerFromDisk -count=1
go test ./internal/connectors/engine -run TestDirectReadJSONRedactedPolicyRemovesSensitiveFields -count=1
go test ./internal/cli -run 'TestGitLabCommandSurfaceRunsBoundedDirectReadProjectView' -count=1
go test ./cmd/connectorgen ./internal/connectors/engine ./internal/connectors/commandrunner ./internal/cli -count=1
go run ./cmd/connectorgen validate internal/connectors/defs --json
go run ./cmd/pm help gitlab
go run ./cmd/pm gitlab
go run ./cmd/pm --json gitlab --help
cd website && pnpm run gen:website-data
```

Result: passed locally. `connectors_checked=547`, `findings=[]`.

## Final Parent Gates

```bash
gofmt -w cmd internal
go vet ./...
go test ./...
go build ./cmd/pm
make verify
go run ./cmd/connectorgen validate internal/connectors/defs
```

Result: passed on 2026-07-09. `make verify` included `go mod tidy`, `go vet ./...`,
`go test -timeout 20m ./...`, `go build ./cmd/pm`, `./pm docs validate`, smoke test, golangci-lint,
and connectorgen validation.

## Website Blocker

```bash
cd website && pnpm run typecheck
cd website && pnpm install --frozen-lockfile
```

Result: blocked locally. `pnpm run typecheck` fails with `tsc: command not found` because
`node_modules` is absent. `pnpm install --frozen-lockfile` fails with
`ERR_PNPM_LOCKFILE_CONFIG_MISMATCH`. Generated website data was refreshed with
`pnpm run gen:website-data`; no non-frozen install was run.

## Review Route

- Parent PR route: PR #127 was marked ready for review after local verification.
- CodeRabbit: automatic review attempted, then reported `Review limit reached` with next review available in 55 minutes. No manual CodeRabbit review command was posted.
- Fallback: GitHub Copilot review was requested as backup reviewer `Copilot` because CodeRabbit rate limiting blocked automated review coverage.
- Copilot review submitted 3 actionable comments; all were addressed by narrowing `content` redaction to exact-key matching and cleaning GitLab global flag summaries before `maps_to` rendering.
- Review-fix verification: `go test ./internal/connectors/engine -run TestDirectReadJSONRedactedPolicyRemovesSensitiveFields -count=1`, `go run ./cmd/connectorgen validate internal/connectors/defs --json`, `go test -timeout 20m ./...`, `go build ./cmd/pm`, `make verify`, and `go run ./cmd/connectorgen validate internal/connectors/defs` passed.
- Stacked sub-PRs: use parent-PR fallback if CodeRabbit skips non-default base PRs.
- Do not post manual CodeRabbit review commands unless automatic review is skipped, disabled, paused, rate-limited past retry window, or otherwise blocked per `.agents/agentic-delivery/workflows/automated-review-routing-loop.md`.
