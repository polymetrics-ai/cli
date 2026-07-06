# CodeRabbit review best practices

Accessed: 2026-07-07

## Primary sources

- CodeRabbit review commands: https://docs.coderabbit.ai/reference/review-commands
- CodeRabbit manage reviews guide: https://docs.coderabbit.ai/guides/commands
- CodeRabbit learnings guide: https://docs.coderabbit.ai/knowledge-base/learnings
- CodeRabbit auto-review configuration: https://docs.coderabbit.ai/configuration/auto-review
- GitHub pull-request review comments REST API: https://docs.github.com/en/rest/pulls/comments
- GitHub issue comments REST API: https://docs.github.com/en/rest/issues/comments
- GitHub pull-request GraphQL mutations: https://docs.github.com/en/graphql/reference/pulls

## Workflow findings

- Use `@coderabbitai full review` for the first complete pass over a ready PR.
- CodeRabbit's command reference distinguishes `@coderabbitai review` as an incremental review of
  new changes and `@coderabbitai full review` as a complete review of all files from scratch.
- CodeRabbit automatically reviews new PRs by default and updates its review when new commits are
  pushed.
- CodeRabbit's `reviews.auto_review.auto_incremental_review` setting defaults to enabled. When it
  is enabled, new pushes receive focused incremental review without manually posting a command.
- Use `@coderabbitai review` only as a manual incremental trigger for new, unreviewed changes. It is
  appropriate when automatic review is disabled, paused, skipped, or has reached the configured
  automatic pause threshold.
- Do not post `@coderabbitai review` after every push. If automatic review is active, wait for the
  automatic review result instead. If the latest commit has already been reviewed, the command adds
  noise and does not produce a new review of old commits.
- Use `@coderabbitai full review` for a fresh review from scratch only when the PR has changed
  substantially, the previous review is stale, or the coordinator asks for a complete pass.
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
- Treat skipped-review status comments as review-routing events, not approval. In PR #48,
  CodeRabbit reported that auto reviews were disabled on base/target branches other than the
  default branch, while a manually requested full review still produced an actionable review record.
  Agents must inspect both the comments and the review records before deciding whether CodeRabbit
  coverage exists.
- For stacked sub-PRs against a non-default parent branch, require one of:
  - CodeRabbit review records on the sub-PR that cover the relevant commits.
  - CodeRabbit review records on the parent PR to the default branch after the sub-issue commits are
    integrated into the parent branch.
  - A recorded blocker approved by the human coordinator.

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
