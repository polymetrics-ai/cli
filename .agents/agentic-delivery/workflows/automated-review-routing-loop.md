# Automated Review Routing Loop

Use this workflow to choose the automated review route for a PR without wasting the Claude
subscription quota or creating noisy duplicate reviews.

Accessed: 2026-07-07

## Primary Sources

- Claude review Action (this repo): `.github/workflows/claude-review.yml`
- Claude Code GitHub Action: https://github.com/anthropics/claude-code-action
- Claude Code documentation: https://docs.claude.com/en/docs/claude-code
- GitHub Copilot code review: https://docs.github.com/en/copilot/how-tos/use-copilot-agents/request-a-code-review/use-code-review
- GitHub Copilot code review concepts: https://docs.github.com/en/copilot/concepts/agents/code-review
- GitHub Copilot automatic review configuration: https://docs.github.com/en/copilot/how-tos/copilot-on-github/set-up-copilot/configure-automatic-review
- GitHub CLI Copilot reviewer support: https://github.blog/changelog/2026-03-11-request-copilot-code-review-from-github-cli/

## Policy

Claude is the primary automated reviewer, delivered by the `.github/workflows/claude-review.yml`
GitHub Action. GitHub Copilot is a backup review route only when Claude cannot provide timely
coverage. Copilot feedback is review input, not approval.

## Review Order

1. Prefer Claude automatic review for non-draft PRs opened, reopened, or marked ready by a trusted
   author (owner, member, collaborator, or contributor).
2. For a PR whose automatic review did not run (untrusted or first-time external author), have a
   maintainer invoke `@claude review`.
3. For non-default-base stacked PRs, rely on the sub-PR automatic review when the author is trusted;
   otherwise a maintainer invokes `@claude review`, or route coverage through the parent PR
   fallback.
4. Use GitHub Copilot code review as a backup only when Claude is unavailable — a review run failed
   or the subscription quota is exhausted — and the PR still needs automated review coverage.
5. Use human review as the final fallback when both automated routes are unavailable or the change
   crosses a human gate.

## Decision Table

| Condition | Route | Notes |
| --- | --- | --- |
| Non-draft PR opened/reopened/ready by a trusted author | `claude_auto` | Wait for the automatic review. Do not post a manual command. |
| Automatic review did not run (untrusted author) or new unreviewed commits need another pass | `claude_manual` | A maintainer posts one `@claude review` when the fallback condition applies. |
| The claude-review Action cannot run at all (missing/empty `ANTHROPIC_API_KEY` secret, errored workflow runs) but Claude itself is available locally | `claude_local` | Run `scripts/claude-local-review.sh <pr>` — first-party `claude` CLI (default `claude-opus-4-8`, operator subscription, never a gateway). Posts one consolidated review comment with the reviewed head SHA and a durable copy under `.planning/reviews/`. Disposition findings like any Claude review; comment-only, approval stays human. |
| Claude review (Action, manual, and local CLI) failed or subscription quota is exhausted and review coverage blocks progress | `copilot_backup` | Request Copilot once, then disposition Copilot comments like other review comments. |
| Claude and Copilot are both unavailable, or the change crosses a human gate | `human` | Record `needs_human`; do not weaken the gate. |

## Claude Unavailability Handling

When a Claude review run fails or the Claude subscription quota is exhausted:

1. Do not retry immediately.
2. Record the failure and the affected commit SHA.
3. If the work is not blocked, wait and re-invoke `@claude review`, or wait for the next automatic
   trigger (a reopen or ready-for-review event on a trusted-author PR).
4. If the review is blocking progress, request GitHub Copilot review as a backup when it is enabled
   for the repository or organization.
5. If Copilot is unavailable or disabled, record `needs_human` and ask the coordinator for review
   direction.
6. Do not spam `@claude review` to test whether the quota has cleared. Re-invoke deliberately once,
   or wait for the next automatic trigger.

## Copilot Backup Rules

GitHub Copilot review is backup review input, not approval:

- Request it only after a Claude review run failed, the subscription quota is exhausted, or the
  author-trust gate blocked automatic review with no maintainer available to invoke it.
- Do not request it routinely on every PR when Claude is healthy.
- Do not use it to bypass required human approval or parent PR human gates.
- Do not treat it as Claude approval.
- Record that Copilot reviews leave comment reviews and do not satisfy required approvals.
- Record that Copilot review may consume GitHub AI credits or be blocked by organization budget or
  repository policy.
- Triage Copilot comments with the same disposition template used for Claude comments.
- Request Copilot once per blocker window unless new commits materially change the diff.

## Requesting Copilot Review

Use the lowest-noise supported route:

```bash
gh pr edit <number> --add-reviewer @copilot
```

If the GitHub CLI path is unavailable, use the GitHub REST review-request API with reviewer
`copilot-pull-request-reviewer[bot]`, as documented by GitHub. If both routes fail, record the
failure and do not retry in a loop.

Do not combine a Copilot backup request with a new Claude manual review command. Pick one route
for the blocker window and record it.

## Review Coverage Record

Record all automated review routes in the PR body, parent issue state ledger, or phase artifact:

- primary route: `claude_auto`, `claude_manual`, `copilot_backup`, `human`, or `blocked`
- primary route status
- fallback route, if any
- fallback status
- reviewed commit SHA or range
- review URLs
- disposition summary
- blocker and retry action, if applicable
