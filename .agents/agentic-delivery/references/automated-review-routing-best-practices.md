# Automated review routing best practices

Accessed: 2026-07-06

## Primary sources

- Claude review Action (this repo): `.github/workflows/claude-review.yml`
- Claude Code GitHub Action: https://github.com/anthropics/claude-code-action
- Claude Code documentation: https://docs.claude.com/en/docs/claude-code
- GitHub Copilot code review: https://docs.github.com/en/copilot/how-tos/use-copilot-agents/request-a-code-review/use-code-review
- GitHub Copilot code review concepts: https://docs.github.com/en/copilot/concepts/agents/code-review
- GitHub Copilot automatic review configuration: https://docs.github.com/en/copilot/how-tos/copilot-on-github/set-up-copilot/configure-automatic-review
- GitHub CLI Copilot reviewer support: https://github.blog/changelog/2026-03-11-request-copilot-code-review-from-github-cli/

## Findings

- The Claude review Action supports automatic pull request review for trusted authors (on open,
  reopen, or ready-for-review) plus on-demand `@claude review` comments. The manual command is an
  on-demand trigger, not a command to repeat after every push.
- A failed Claude review run or an exhausted Claude subscription quota simply errors the job.
  Immediate retries waste the quota and add PR noise. Record the failure and choose one path:
  re-invoke `@claude review` deliberately, wait for the next automatic trigger, or fall back.
- GitHub Copilot can be requested as a pull request reviewer, including through GitHub CLI
  `gh pr edit <number> --add-reviewer @copilot` when the installed CLI and repository support it.
- GitHub documents the REST reviewer identity as `copilot-pull-request-reviewer[bot]` for Copilot
  pull request review requests.
- Copilot reviews leave comment feedback. They are not approval, do not request changes, and do not
  replace required human approval.
- Copilot review can be limited by organization policy, repository configuration, licensing, or AI
  credit budgets. A failed Copilot request is a review blocker, not permission to bypass review.

## Polymetrics policy

- Use Claude automatic review as the primary automated route.
- Do not manually trigger Claude on normal non-draft PRs opened by a trusted author. Wait for the
  automatic review.
- Use GitHub Copilot review only when Claude is unavailable — a review run failed or the
  subscription quota is exhausted — and automated review coverage is blocking progress.
- Do not request both Claude and Copilot repeatedly in the same blocker window.
- Disposition Copilot comments with the same accepted, accepted_with_modification, declined,
  deferred, or needs_human template used for Claude comments.
- Treat Copilot review as backup review input, never approval.
