# Issue prompt template

Add this section to implementation issues.

```markdown
## Agent execution contract

Follow the generic issue-to-PR contract:
`.agents/agentic-delivery/contracts/issue-agent-contract.md`

Follow the post-implementation CodeRabbit review loop:
`.agents/agentic-delivery/workflows/coderabbit-review-loop.md`

For parent issues, sub-issues, and stacked PRs, follow:
`.agents/agentic-delivery/workflows/stacked-parent-subissue-workflow.md`

For parent issues that spawn or assign multiple workers, follow:
`.agents/agentic-delivery/contracts/parent-orchestrator-contract.md`
`.agents/agentic-delivery/workflows/parent-issue-orchestration-loop.md`

Task type: `<task-type-from-task-skill-matrix>`

Required skills:
- `gsd-programming-loop` for implementation or behavior-changing work
- `<skill capability or local skill name>`

Primary agent:
`.agents/<functional-area>/agents/<type>/<agent>.agent.yaml`

Parent issue:
- `<parent issue URL or "None">`

Orchestration:
- spawned by: `<parent issue orchestrator or "None">`
- state ledger: `<issue comment, PR body section, file path, or "None">`
- worker handoff template: `.agents/agentic-delivery/contracts/worker-handoff-template.md`
- merge owner: `<parent issue orchestrator | assigned coordinator | not applicable>`
- CodeRabbit coverage route: `<sub_pr | parent_pr_fallback | blocked | not applicable>`

Branch policy:
- parent branch: `<type>/<parent-issue>-<slug>` or `None`
- PR base: `main` for parent PRs, parent branch for sub-PRs

Hard stops:
- Do not change auth scopes.
- Do not request or print secrets.
- Do not weaken tests or quality gates.
- Do not expand scope beyond this issue.

PR body must include one of:
- `Closes #<issue-number>` when the PR completes the issue
- `Refs #<issue-number>` when the PR is stacked or incremental

Before merge:
- confirm `gsd-programming-loop` was loaded and followed for implementation or behavior-changing
  work, or record the manual-GSD fallback when local GSD scripts are unavailable
- confirm the GSD plan, TDD ledger, and verification checklist were created or updated before
  production edits
- commit and push coherent green slices to the active issue/PR branch after local green gates;
  never push to `main`
- observe automatic CodeRabbit review after implementation when the PR is non-draft and targets
  `main`; do not post manual review commands unless the documented fallback conditions apply
- confirm CodeRabbit actually reviewed the relevant commits, or record the parent PR fallback route
  for stacked sub-PRs
- reply to every actionable CodeRabbit item with accepted, accepted_with_modification, declined,
  deferred, or needs_human
- rerun verification after accepted fixes
- ensure accepted fix commits are CodeRabbit-reviewed; wait for automatic incremental review when
  active, and use manual `@coderabbitai review` only when automatic review is paused, disabled,
  skipped, rate-limit retry is due, or the automatic pause threshold was reached
- merge sub-PRs into parent branches only when all automated gates pass and no human gate is
  triggered
- require human approval before merging parent PRs into `main`
```
