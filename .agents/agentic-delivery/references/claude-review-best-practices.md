# Claude review best practices

Accessed: 2026-07-06

## Primary sources

- Claude review Action (this repo): `.github/workflows/claude-review.yml`
- Claude Code GitHub Action: https://github.com/anthropics/claude-code-action
- Claude Code documentation: https://docs.claude.com/en/docs/claude-code
- GitHub pull-request review comments REST API: https://docs.github.com/en/rest/pulls/comments
- GitHub issue comments REST API: https://docs.github.com/en/rest/issues/comments
- GitHub pull-request GraphQL mutations: https://docs.github.com/en/graphql/reference/pulls
- GitHub Copilot code review: https://docs.github.com/en/copilot/how-tos/use-copilot-agents/request-a-code-review/use-code-review

## Workflow findings

- The `.github/workflows/claude-review.yml` Action reviews a pull request automatically when a
  trusted author (GitHub author association owner, member, collaborator, or contributor) opens,
  reopens, or marks it ready for review. It does not run on every push.
- For such a PR, wait for the automatic review instead of posting `@claude review`.
- Automatic review does not re-run when new commits are pushed. To get another pass over new
  commits, a maintainer posts one `@claude review`.
- A PR from an untrusted or first-time external author does not trigger automatic review. A
  maintainer with write access invokes `@claude review` to start it.
- Use `@claude review` as an on-demand trigger for new, unreviewed changes. It is appropriate when
  automatic review did not run (untrusted author), when fix commits need another pass, or when the
  coordinator approves a fresh review.
- Do not post `@claude review` after every push. If the latest commits have already been reviewed,
  the command only adds noise. Request a fresh review only when there are new unreviewed commits.
- `@claude review` can also target part of a diff (for example, `@claude review the error
  handling`) when a focused pass is wanted.
- If a Claude review run fails or the Claude subscription quota is exhausted, do not retry
  immediately. Record the blocker, wait, and re-invoke `@claude review` deliberately, or fall back
  to Copilot/human review.
- If Claude is unavailable and automated review coverage is blocking progress, route to GitHub
  Copilot backup review using
  `.agents/agentic-delivery/workflows/automated-review-routing-loop.md`.
- Resolve a Claude review thread only after its finding has actually been addressed or explicitly
  dispositioned. Resolving is a GitHub conversation action, not a bot command.
- Reply directly to the specific inline comment when possible. This is the highest-context way to
  communicate corrections and durable guidance.
- Explain why a decision was made, not only what action was taken.
- Distinguish one-off PR exceptions from durable team preferences. One-off exceptions can be
  resolved with an explanation; recurring review guidance should become durable repo guidance.
- GitHub exposes inline pull-request review comments separately from issue comments. A review loop
  must inspect both surfaces: Claude posts inline comments on specific lines plus one top-level
  summary comment.
- Creating comments can trigger notifications and GitHub secondary rate limiting, so batch
  disposition summaries for top-level informational comments instead of posting noisy one-line
  replies.

## Project policy

For Polymetrics PRs, every actionable Claude item must receive one of these dispositions before
merge:

- accepted
- accepted_with_modification
- declined
- deferred
- needs_human

The disposition must include a reason and evidence. Accepted fixes require tests or a documented
docs-only rationale. Deferred work requires a follow-up issue or explicit coordinator approval.
Copilot backup comments use the same disposition template, but Copilot review is not approval.
