# Plan: Issue-first agentic delivery foundation

## Checklist

- [x] Inspect current repo planning, PR template, and CI workflows.
- [x] Add red tests for an issue-first PR guard.
- [x] Implement minimal Go guard and GitHub Actions workflow.
- [x] Add an agent task issue form.
- [x] Isolate agent-neutral contracts under `.agents/connector-cli-parity/`.
- [x] Isolate repo-local YAML agents under `.codex/agents/connector-cli-parity/`.
- [ ] Run targeted and relevant broad verification.

## Implementation sequence

1. Foundation guard PR slice:
   - `internal/coordination/issueguard`
   - `cmd/prissueguard`
   - `.github/workflows/pr-issue-guard.yml`
   - `.github/pull_request_template.md`
2. Agentic delivery assets:
   - `.agents/connector-cli-parity/issue-agent-contract.md`
   - `.agents/connector-cli-parity/task-skill-matrix.yaml`
   - `.agents/connector-cli-parity/agent-spec.schema.yaml`
   - `.codex/agents/connector-cli-parity/*.agent.yaml`
3. GitHub authoring harness:
   - `.github/ISSUE_TEMPLATE/agent_task.yml`
   - PR body must contain `Closes #43`

## Verification

- `go test ./internal/coordination/issueguard`
- `go run ./cmd/prissueguard --title 'feat(github): add cli surface metadata' --body 'Closes #123'`
- `go test ./cmd/prissueguard ./internal/coordination/issueguard`
- YAML parse check for `.agents/`, `.codex/agents/connector-cli-parity/`, and GitHub harness files
- targeted secret-pattern scan over new agent, workflow, template, and guard files
