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

## Required Final Parent Gates

```bash
gofmt -w cmd internal
go vet ./...
go test ./...
go build ./cmd/pm
make verify
go run ./cmd/connectorgen validate internal/connectors/defs
```

## Review Route

- Parent PR route: CodeRabbit automatic review on non-draft parent PR targeting `main`, or pending coverage while draft.
- Stacked sub-PRs: use parent-PR fallback if CodeRabbit skips non-default base PRs.
- Do not post manual CodeRabbit review commands unless automatic review is skipped, disabled, paused, rate-limited past retry window, or otherwise blocked per `.agents/agentic-delivery/workflows/automated-review-routing-loop.md`.
