# Plan: Issue-first agentic delivery foundation

## Checklist

- [x] Inspect current repo planning, PR template, and CI workflows.
- [x] Add red tests for an issue-first PR guard.
- [x] Implement minimal Go guard and GitHub Actions workflow.
- [x] Add an agent task issue form.
- [x] Isolate agent-neutral contracts under `.agents/agentic-delivery/`.
- [x] Group YAML agents by functional area and type under `.agents/`.
- [x] Convert pre-existing `.codex/agents` TOML specs into `.agents/connector-migration/`.
- [x] Run targeted and relevant broad verification.

## Implementation sequence

1. Foundation guard PR slice:
   - `internal/coordination/issueguard`
   - `cmd/prissueguard`
   - `.github/workflows/pr-issue-guard.yml`
   - `.github/pull_request_template.md`
2. Agentic delivery assets:
   - `.agents/agentic-delivery/contracts/issue-agent-contract.md`
   - `.agents/agentic-delivery/matrices/task-skill-matrix.yaml`
   - `.agents/agentic-delivery/schemas/agent-spec.schema.yaml`
   - `.agents/agentic-delivery/agents/<type>/*.agent.yaml`
   - `.agents/connector-migration/agents/<type>/*.agent.yaml`
3. GitHub authoring harness:
   - `.github/ISSUE_TEMPLATE/agent_task.yml`
   - PR body must contain `Closes #43`

## Verification

- `go test ./internal/coordination/issueguard`
- `go run ./cmd/prissueguard --title 'feat(github): add cli surface metadata' --body 'Closes #123'`
- `go test ./cmd/prissueguard ./internal/coordination/issueguard`
- YAML parse check for `.agents/` and GitHub harness files
- targeted secret-pattern scan over new agent, workflow, template, and guard files
