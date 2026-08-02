# Code review disposition template

Use this template for Claude, GitHub Copilot, and any other automated PR review comments. Reply
directly to the inline thread when possible. If the review item is a top-level PR comment, use a
top-level disposition summary and identify the original comment.

```markdown
Disposition: Accepted | Accepted with modification | Declined | Deferred | Needs human

Action:
<What changed, what will change, or why no code change is being made.>

Reason:
<Why this disposition is correct after checking the code, issue scope, project rules, and sources.>

Evidence:
<Tests, source URLs, issue scope, commits, or file references used to make the decision.>

Follow-up:
<Follow-up issue, human approval gate, or "None".>
```

## Rules

- Leave a disposition reply on every finding, including the ones you fix. No finding is silently
  actioned or ignored: for each finding a reader must be able to see whether it was fixed and why,
  or why it was not. "I fixed it" still requires a reply stating the fix and the reason.
- Do not resolve a review thread before leaving a disposition reply. Resolve the conversation in
  GitHub after findings are addressed; there is no bot resolve command.
- Do not silently dismiss a finding as false positive. Explain why.
- Update docs as part of the disposition. If an `Accepted` or `Accepted with modification` fix
  changes behavior, CLI surface, flags, output, config, or a shared contract, update the
  corresponding docs and website in the same PR per
  `.agents/agentic-delivery/references/cli-help-docs-website-parity.md`, and record it in the
  Action/Evidence fields. If you `Decline` a finding because the current behavior is intended, cite
  the doc that records that intent in Evidence — and if no such doc exists or it is stale, update it
  so the intended behavior is documented rather than left implicit.
- Do not accept a suggestion that crosses a hard stop. Mark it `Needs human`.
- Create or reference a follow-up issue for valid work that is outside the current PR.
- Ensure accepted fix commits are reviewed. Claude auto-reviews a PR when a trusted author opens,
  reopens, or marks it ready; request a single `@claude review` only for new unreviewed commits,
  such as after pushing fix commits. Do not post `@claude review` after every push.
- Use GitHub Copilot review as fallback input only when Claude is unavailable, for example its run
  failed or its subscription quota is exhausted, and review coverage is blocking progress. Copilot
  comments require dispositions, but Copilot review is not approval.
- Do not ask both Claude and Copilot for repeated fresh reviews in the same blocker window.
