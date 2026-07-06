# Automated review routing best practices

Accessed: 2026-07-07

## Primary sources

- CodeRabbit review commands: https://docs.coderabbit.ai/reference/review-commands
- CodeRabbit automatic review controls: https://docs.coderabbit.ai/configuration/auto-review
- CodeRabbit rate limits: https://docs.coderabbit.ai/management/plans#rate-limits
- CodeRabbit usage-based add-on: https://docs.coderabbit.ai/management/usage-based-addon
- GitHub Copilot code review: https://docs.github.com/en/copilot/how-tos/use-copilot-agents/request-a-code-review/use-code-review
- GitHub Copilot code review concepts: https://docs.github.com/en/copilot/concepts/agents/code-review
- GitHub Copilot automatic review configuration: https://docs.github.com/en/copilot/how-tos/copilot-on-github/set-up-copilot/configure-automatic-review
- GitHub CLI Copilot reviewer support: https://github.blog/changelog/2026-03-11-request-copilot-code-review-from-github-cli/

## Findings

- CodeRabbit supports automatic pull request review and manual review commands. Manual commands are
  fallback controls, not a command to repeat after every push.
- CodeRabbit rate limits can delay additional reviews. When a limit is reported, immediate retries
  waste review allowance and add PR noise. Record the wait window and choose one fallback route.
- GitHub Copilot can be requested as a pull request reviewer, including through GitHub CLI
  `gh pr edit <number> --add-reviewer @copilot` when the installed CLI and repository support it.
- GitHub documents the REST reviewer identity as `copilot-pull-request-reviewer[bot]` for Copilot
  pull request review requests.
- Copilot reviews leave comment feedback. They are not approval, do not request changes, and do not
  replace required human approval.
- Copilot review can be limited by organization policy, repository configuration, licensing, or AI
  credit budgets. A failed Copilot request is a review blocker, not permission to bypass review.

## Polymetrics policy

- Use CodeRabbit automatic review as the primary automated route.
- Do not manually trigger CodeRabbit on normal non-draft PRs targeting the default branch.
- Use GitHub Copilot review only when CodeRabbit is rate-limited, skipped, disabled, paused, or
  unavailable and automated review coverage is blocking progress.
- Do not request both CodeRabbit and Copilot repeatedly in the same blocker window.
- Disposition Copilot comments with the same accepted, accepted_with_modification, declined,
  deferred, or needs_human template used for CodeRabbit comments.
- Treat Copilot review as backup review input, never approval.
