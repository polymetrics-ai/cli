# Issue prompt template

Add this section to implementation issues.

```markdown
## Agent execution contract

Follow the generic issue-to-PR contract:
`.agents/agentic-delivery/contracts/issue-agent-contract.md`

Use the repo-local official GSD Core Pi adapter:
`.agents/agentic-delivery/references/gsd-pi-adapter.md`

Load required Go/design skills:
`.agents/agentic-delivery/references/required-skills-routing.md`

For CLI command, flag, output, connector surface, or help-topic changes, use:
`.agents/agentic-delivery/references/cli-help-docs-website-parity.md`

Follow the post-implementation local review loop:
`.agents/agentic-delivery/workflows/local-review-loop.md`

Choose the local automated review route before handoff:
`.agents/agentic-delivery/workflows/automated-review-routing-loop.md`

For parent issues, sub-issues, and stacked PRs, follow:
`.agents/agentic-delivery/workflows/stacked-parent-subissue-workflow.md`

For parent issues that spawn or assign multiple workers, follow:
`.agents/agentic-delivery/contracts/parent-orchestrator-contract.md`
`.agents/agentic-delivery/workflows/parent-issue-orchestration-loop.md`

Task type: `<task-type-from-task-skill-matrix>`

Required skills:
- `gsd-programming-loop` for implementation or behavior-changing work through `/gsd-programming-loop` in Pi or `scripts/gsd prompt programming-loop ...` from shell
- `golang-how-to` for Go work, plus task-specific Go skills from `required-skills-routing.md`
- design skills such as `frontend-design`, `web-design-guidelines`, and `vercel-react-best-practices` for website/docs UI work
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
- Local automated review coverage route: `<local_reviewer | local_security | local_verifier | human | blocked | not applicable>`
- Local review coverage scope: `<sub_issue_head | parent_batch | not applicable>`

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
- confirm `gsd-programming-loop` was loaded and followed through `/gsd-programming-loop ...` or
  `scripts/gsd prompt programming-loop ...` for implementation or behavior-changing work, or record
  the manual-GSD fallback when the repo-local adapter is unavailable
- confirm the GSD plan, TDD ledger, and verification checklist were created or updated before
  production edits
- confirm required Go/design skills from `.agents/agentic-delivery/references/required-skills-routing.md` were loaded and recorded
- for CLI feature work, confirm runtime help, bare namespace behavior, `docs/cli/**`, website docs,
  generated help/manual artifacts, and tests are updated or explicitly marked not applicable
- commit and push coherent green slices to the active issue/PR branch after local green gates;
  never push to `main`
- run and record local automated review coverage for the exact candidate head or diff range
- reply to or record every actionable local review item with accepted, accepted_with_modification,
  declined, deferred, or needs_human
- rerun verification after accepted fixes
- ensure accepted fix commits receive a follow-up local review pass when they materially change the
  reviewed code
- merge sub-PRs into parent branches only when all automated gates pass and no human gate is
  triggered
- require human approval before merging parent PRs into `main`
```
