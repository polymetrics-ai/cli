## Summary

Adds the issue-first agentic delivery foundation for connector CLI parity:

- generic issue-to-PR agent contract
- task-to-skill matrix
- YAML agent definitions for implementation, review, operation-ledger, CLI surface, help renderer,
  direct-read, GraphQL, sensitive/admin, and rollout work
- GitHub agent task issue form
- PR issue guard command and GitHub Actions workflow
- isolated `.agents/` and `.codex/agents/` layout for agent conventions and role specs

Closes #43

## Verification

```bash
go test ./internal/coordination/issueguard
go test ./cmd/prissueguard ./internal/coordination/issueguard
ruby -e 'require "yaml"; ARGV.each { |f| YAML.load_file(f) }' .agents/connector-cli-parity/*.yaml .codex/agents/connector-cli-parity/*.agent.yaml .github/ISSUE_TEMPLATE/agent_task.yml .github/workflows/pr-issue-guard.yml
git diff --check
make verify
```

## Copilot review loop

After the branch is pushed and the draft PR exists:

1. Request GitHub Copilot review from the PR UI, or comment `@copilot review` if enabled for the
   repository.
2. Address every Copilot finding with a new commit.
3. Rerun targeted checks and `make verify`.
4. Request another Copilot review.
5. Repeat until Copilot has no actionable findings.
6. Ping the human coordinator for approval before marking the PR ready.

## Checklist

- [x] Tests or docs updated for behavior changes
- [x] `make verify` passes locally
- [x] Branch name follows `codex/<issue>-<description>`
- [x] PR title follows Conventional Commits
- [x] No credentials, tokens, private URLs, or customer data included
