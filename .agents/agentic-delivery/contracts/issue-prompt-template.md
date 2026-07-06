# Issue prompt template

Add this section to implementation issues.

```markdown
## Agent execution contract

Follow the generic issue-to-PR contract:
`.agents/agentic-delivery/contracts/issue-agent-contract.md`

Follow the post-implementation CodeRabbit review loop:
`.agents/agentic-delivery/workflows/coderabbit-review-loop.md`

Task type: `<task-type-from-task-skill-matrix>`

Required skills:
- `<skill capability or local skill name>`

Primary agent:
`.agents/<functional-area>/agents/<type>/<agent>.agent.yaml`

Hard stops:
- Do not change auth scopes.
- Do not request or print secrets.
- Do not weaken tests or quality gates.
- Do not expand scope beyond this issue.

PR body must include:
`Closes #<issue-number>`

Before merge:
- request CodeRabbit review after implementation
- reply to every actionable CodeRabbit item with accepted, accepted_with_modification, declined,
  deferred, or needs_human
- rerun verification after accepted fixes
- request incremental CodeRabbit review until no actionable findings remain
```
