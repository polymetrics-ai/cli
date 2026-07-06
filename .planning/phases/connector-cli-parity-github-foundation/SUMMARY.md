# Summary: Issue-first agentic delivery foundation

## Completed

- Added a tested issue-first PR guard:
  - `internal/coordination/issueguard`
  - `cmd/prissueguard`
  - `.github/workflows/pr-issue-guard.yml`
  - `.github/pull_request_template.md`
- Added a GitHub issue form for agent implementation tasks:
  - `.github/ISSUE_TEMPLATE/agent_task.yml`
- Added agent-neutral delivery contracts under `.agents/connector-cli-parity/`, including a generic
  issue-to-PR contract, task skill matrix, YAML agent spec guidance, and prompt template.
- Added repo-local YAML agent definitions under `.codex/agents/connector-cli-parity/`.
- Linked the PR slice to live issue #43.

## Out Of Scope

- Complete GitHub CLI migration planning remains in #44.
- Blog work remains in #45.
- GitHub Projects setup is not part of this PR.

## Verification

See `VERIFICATION.md`.

## PR status

Draft PR title: `feat(agentic): add issue-first delivery system`.

PR body is recorded in `PR-BODY.md` and includes `Closes #43`.
