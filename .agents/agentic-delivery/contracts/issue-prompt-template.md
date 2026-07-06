# Issue prompt template

Add this section to implementation issues.

```markdown
## Agent execution contract

Follow the generic issue-to-PR contract:
`.agents/agentic-delivery/contracts/issue-agent-contract.md`

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
```
