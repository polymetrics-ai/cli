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

Pending after planning artifact creation:

```bash
jq empty .planning/phases/issue-78-gitlab-cli-parity-parent/RUN-STATE.json \
  .planning/phases/issue-78-gitlab-cli-parity-parent/ORCHESTRATION-STATE.json
git diff --check
```

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
