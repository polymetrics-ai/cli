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
- Added a CodeRabbit post-implementation review loop, disposition reply template, source-backed
  guidance, and a dedicated CodeRabbit disposition agent.
- Added parent issue, sub-issue, milestone roadmap, and stacked PR workflow guidance for large
  implementation efforts.
- Added YAML agent definitions grouped by functional area and type under `.agents/`.
- Converted the pre-existing `.codex/agents` Pass B expander and connector reviewer TOML files into
  `.agents/connector-migration/` YAML specs and removed the obsolete TOML files.
- Tightened the PR guard after automated PR review so its title regex matches the repository
  `pr-title` workflow and ambiguous issue relationships do not pass.
- Updated live issue #43 to use the `.agents/agentic-delivery/` and `.agents/connector-migration/`
  acceptance criteria.
- Linked the PR slice to live issue #43.
- Requested CodeRabbit review on PR #47.
- Accepted and fixed all 7 CodeRabbit findings from the full review, including CLI exit-code test
  coverage and agent permission hardening.
- Updated the CodeRabbit workflow after observing the incremental-review note so agents wait for
  automatic incremental review when active and only request manual review for new unreviewed commits
  when automatic review is paused, disabled, skipped, rate-limit retry is due, or auto-paused.
- Updated root `AGENTS.md` and added `CLAUDE.md` so Codex, Claude Code, and other agents share the
  same issue-first and CodeRabbit review rules without duplicating the full workflow.
- Made `gsd-programming-loop` mandatory for implementation and behavior-changing agent work, with a
  required manual-GSD fallback note only when local GSD scripts are unavailable.
- Updated GitHub CLI feature parity planning so issue #44 is the parent roadmap and issues #34-#42
  are the sub-issue implementation slices.
- Updated live issue #44 and attached issues #34-#42 as GitHub sub-issues through the REST API.

## Out Of Scope

- Complete GitHub CLI migration planning remains in #44.
- Implementing the GitHub CLI parity slices remains in #34-#42.
- Blog work remains in #45.
- GitHub Projects setup is not part of this PR.

## Verification

See `VERIFICATION.md`.

## PR status

PR title: `feat(agentic): add issue-first delivery system`.

PR body is recorded in `PR-BODY.md` and includes `Closes #43`.
