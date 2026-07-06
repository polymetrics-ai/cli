# Code review disposition template

Use this template for CodeRabbit, GitHub Copilot, and any other automated PR review comments. Reply
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

- Do not resolve a review thread before leaving a disposition reply.
- Do not silently dismiss a finding as false positive. Explain why.
- Do not accept a suggestion that crosses a hard stop. Mark it `Needs human`.
- Create or reference a follow-up issue for valid work that is outside the current PR.
- Ensure accepted fix commits are reviewed. Prefer automatic CodeRabbit incremental review when it
  is active; request manual `@coderabbitai review` only for new unreviewed commits when automatic
  review is paused, disabled, skipped, or due after a rate-limit or auto-pause window. Do not post
  `@coderabbitai full review` for a normal non-draft PR targeting the default branch.
- Use GitHub Copilot review as fallback input only when CodeRabbit is rate-limited, skipped,
  disabled, paused, or unavailable and review coverage is blocking progress. Copilot comments
  require dispositions, but Copilot review is not approval.
- Do not ask both CodeRabbit and Copilot for repeated fresh reviews in the same blocker window.
