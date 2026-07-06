# CodeRabbit review best practices

Accessed: 2026-07-06

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
- Use `@coderabbitai review` for incremental review after follow-up commits.
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
