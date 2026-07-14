# Code review disposition template

Use this template for local automated review findings and any optional human review comments. Record
the disposition in the phase artifact, PR body, handoff, or comment thread where the finding is
tracked.

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

- Do not silently dismiss a finding as false positive. Explain why.
- Do not accept a suggestion that crosses a hard stop. Mark it `Needs human`.
- Create or reference a follow-up issue for valid work that is outside the current PR.
- Ensure accepted fix commits receive a follow-up local review pass when they materially change the
  reviewed code.
- Do not treat optional remote PR-bot comments as required approval.
