# Summary: Issue-first agentic delivery foundation

## Completed

- Added a tested issue-first PR guard:
  - `internal/coordination/issueguard`
  - `cmd/prissueguard`
  - `.github/workflows/pr-issue-guard.yml`
  - `.github/pull_request_template.md`
- Added a GitHub issue form for agent implementation tasks:
  - `.github/ISSUE_TEMPLATE/agent_task.yml`
- Added agent-neutral delivery contracts under `.agents/agentic-delivery/`, including a generic
  issue-to-PR contract, task skill matrix, YAML agent spec guidance, and prompt template.
- Added YAML agent definitions grouped by functional area and type under `.agents/`.
- Converted the pre-existing `.codex/agents` Pass B expander and connector reviewer TOML files into
  `.agents/connector-migration/` YAML specs and removed the obsolete TOML files.
- Tightened the PR guard after automated PR review so its title regex matches the repository
  `pr-title` workflow and ambiguous issue relationships do not pass.
- Updated live issue #43 to use the `.agents/agentic-delivery/` and `.agents/connector-migration/`
  acceptance criteria.
- Linked the PR slice to live issue #43.

## Out Of Scope

- Complete GitHub CLI migration planning remains in #44.
- Blog work remains in #45.
- GitHub Projects setup is not part of this PR.

## Verification

See `VERIFICATION.md`.

## PR status

PR title: `feat(agentic): add issue-first delivery system`.

PR body is recorded in `PR-BODY.md` and includes `Closes #43`.
