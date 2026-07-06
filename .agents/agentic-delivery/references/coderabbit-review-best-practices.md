# CodeRabbit review best practices

Accessed: 2026-07-06

## Primary sources

- CodeRabbit review commands: https://docs.coderabbit.ai/reference/review-commands
- CodeRabbit manage reviews guide: https://docs.coderabbit.ai/guides/commands
- CodeRabbit learnings guide: https://docs.coderabbit.ai/knowledge-base/learnings
- CodeRabbit auto-review configuration: https://docs.coderabbit.ai/configuration/auto-review
- CodeRabbit configuration reference: https://docs.coderabbit.ai/reference/configuration
- CodeRabbit plans and rate limits: https://docs.coderabbit.ai/management/plans
- CodeRabbit usage-based add-on: https://docs.coderabbit.ai/management/usage-based-addon
- GitHub pull-request review comments REST API: https://docs.github.com/en/rest/pulls/comments
- GitHub issue comments REST API: https://docs.github.com/en/rest/issues/comments
- GitHub pull-request GraphQL mutations: https://docs.github.com/en/graphql/reference/pulls
- GitHub Copilot code review: https://docs.github.com/en/copilot/how-tos/use-copilot-agents/request-a-code-review/use-code-review

## Workflow findings

- CodeRabbit automatically reviews pull requests opened against the default branch when automatic
  review is enabled and the PR is not excluded by draft, branch, label, title, or user settings.
- For a non-draft PR targeting the default branch, wait for automatic review instead of posting
  `@coderabbitai full review`.
- CodeRabbit updates its review when new commits are pushed.
- CodeRabbit's `reviews.auto_review.auto_incremental_review` setting defaults to enabled. When it
  is enabled, new pushes receive focused incremental review without manually posting a command.
- Use `@coderabbitai review` only as a manual incremental trigger for new, unreviewed changes. It is
  appropriate when automatic review is disabled, paused, skipped, due after a rate-limit window, or
  has reached the configured automatic pause threshold.
- Do not post `@coderabbitai review` after every push. If automatic review is active, wait for the
  automatic review result instead. If the latest commit has already been reviewed, the command adds
  noise and does not produce a new review of old commits.
- Use `@coderabbitai full review` for a fresh review from scratch only when automatic review did not
  provide coverage and the PR has changed substantially, the previous review is stale, or the
  coordinator explicitly approves a complete manual pass.
- Manual review commands consume PR review allowance when they run. If CodeRabbit reports a review
  limit or fair-usage delay, do not retry immediately. Record the blocker, wait for the reported
  window, and prefer the next automatic review triggered by a pushed commit.
- If CodeRabbit is rate-limited, skipped, disabled, paused, or unavailable and automated review
  coverage is blocking progress, route to GitHub Copilot backup review using
  `.agents/agentic-delivery/workflows/automated-review-routing-loop.md`.
- Use `@coderabbitai resolve` only after CodeRabbit feedback has actually been addressed.
- Use `@coderabbitai approve` only when the repository enables CodeRabbit's request-changes
  workflow and approval is desired.
- Reply directly to the specific inline comment when possible. CodeRabbit documents this as the
  highest-context way to communicate corrections and durable guidance.
- Explain why a decision was made, not only what action was taken.
- Distinguish one-off PR exceptions from durable team preferences. One-off exceptions can be
  resolved with an explanation; recurring review guidance should become durable repo guidance.
- GitHub exposes inline pull-request review comments separately from issue comments. A review loop
  must inspect both surfaces.
- Creating comments can trigger notifications and secondary rate limiting, so batch disposition
  summaries for top-level informational comments instead of posting noisy one-line replies.
- Treat CodeRabbit's incremental-review note as informational. Update local process state instead
  of replying with another review command unless there are new unreviewed commits and automatic
  review is paused or unavailable.

## Project policy

For Polymetrics PRs, every actionable CodeRabbit item must receive one of these dispositions before
merge:

- accepted
- accepted_with_modification
- declined
- deferred
- needs_human

The disposition must include a reason and evidence. Accepted fixes require tests or a documented
docs-only rationale. Deferred work requires a follow-up issue or explicit coordinator approval.
Copilot backup comments use the same disposition template, but Copilot review is not approval.
