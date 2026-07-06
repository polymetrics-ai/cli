# Automated Review Routing Loop

Use this workflow to choose the automated review route for a PR without wasting review allowance or
creating noisy duplicate reviews.

Accessed: 2026-07-07

## Primary Sources

- CodeRabbit review commands: https://docs.coderabbit.ai/reference/review-commands
- CodeRabbit automatic review controls: https://docs.coderabbit.ai/configuration/auto-review
- CodeRabbit rate limits: https://docs.coderabbit.ai/management/plans#rate-limits
- GitHub Copilot code review: https://docs.github.com/en/copilot/how-tos/use-copilot-agents/request-a-code-review/use-code-review
- GitHub Copilot code review concepts: https://docs.github.com/en/copilot/concepts/agents/code-review
- GitHub Copilot automatic review configuration: https://docs.github.com/en/copilot/how-tos/copilot-on-github/set-up-copilot/configure-automatic-review
- GitHub CLI Copilot reviewer support: https://github.blog/changelog/2026-03-11-request-copilot-code-review-from-github-cli/

## Policy

CodeRabbit is the primary automated reviewer. GitHub Copilot is a backup review route only when
CodeRabbit cannot provide timely coverage. Copilot feedback is review input, not approval.

## Review Order

1. Prefer CodeRabbit automatic review for non-draft PRs targeting the default branch.
2. Prefer CodeRabbit automatic incremental review for pushed fix commits when automatic review is
   active.
3. For non-default-base stacked PRs, use CodeRabbit only when repository configuration reviews that
   base branch; otherwise route coverage through the parent PR fallback.
4. Use GitHub Copilot code review as a backup only when CodeRabbit is unavailable, rate-limited,
   skipped, disabled, or paused and the PR still needs automated review coverage.
5. Use human review as the final fallback when both automated routes are unavailable or the change
   crosses a human gate.

## Decision Table

| Condition | Route | Notes |
| --- | --- | --- |
| Non-draft PR targets default branch and CodeRabbit auto-review is enabled | `coderabbit_auto` | Wait for the automatic review. Do not post a manual command. |
| New fix commit on a PR with active CodeRabbit incremental review | `coderabbit_auto_incremental` | Push the fix and wait for the automatic incremental review. |
| CodeRabbit automatic review is paused, disabled, skipped, or past its retry window with new unreviewed commits | `coderabbit_manual_fallback` | Use one manual command only when the documented fallback condition applies. |
| CodeRabbit reports review allowance exhausted or fair-usage delay and review coverage blocks progress | `copilot_backup` | Request Copilot once, then disposition Copilot comments like other review comments. |
| CodeRabbit and Copilot are both unavailable, or the change crosses a human gate | `human` | Record `needs_human`; do not weaken the gate. |

## CodeRabbit Rate-Limit Handling

When CodeRabbit reports a review limit or fair-usage delay:

1. Do not retry immediately.
2. Record the reported wait window and the affected commit SHA.
3. If the wait window is short and the work is not blocked, wait for CodeRabbit or the next
   automatic review trigger.
4. If the review is blocking progress, request GitHub Copilot review as a backup when it is enabled
   for the repository or organization.
5. If Copilot is unavailable or disabled, record `needs_human` and ask the coordinator for review
   direction.
6. Do not post another CodeRabbit review command just to test whether the limit has cleared. Use
   the next automatic review trigger or the recorded retry window.

## Copilot Backup Rules

GitHub Copilot review is backup review input, not approval:

- Request it only after CodeRabbit is rate-limited, skipped, disabled, paused, or unavailable.
- Do not request it routinely on every PR when CodeRabbit is healthy.
- Do not use it to bypass required human approval or parent PR human gates.
- Do not treat it as CodeRabbit approval.
- Record that Copilot reviews leave comment reviews and do not satisfy required approvals.
- Record that Copilot review may consume GitHub AI credits or be blocked by organization budget or
  repository policy.
- Triage Copilot comments with the same disposition template used for CodeRabbit comments.
- Request Copilot once per blocker window unless new commits materially change the diff.

## Requesting Copilot Review

Use the lowest-noise supported route:

```bash
gh pr edit <number> --add-reviewer @copilot
```

If the GitHub CLI path is unavailable, use the GitHub REST review-request API with reviewer
`copilot-pull-request-reviewer[bot]`, as documented by GitHub. If both routes fail, record the
failure and do not retry in a loop.

Do not combine a Copilot backup request with a new CodeRabbit manual review command. Pick one route
for the blocker window and record it.

## Review Coverage Record

Record all automated review routes in the PR body, parent issue state ledger, or phase artifact:

- primary route: `coderabbit_auto`, `coderabbit_auto_incremental`,
  `coderabbit_manual_fallback`, `copilot_backup`, `human`, or `blocked`
- primary route status
- fallback route, if any
- fallback status
- reviewed commit SHA or range
- review URLs
- disposition summary
- blocker and retry window, if applicable
