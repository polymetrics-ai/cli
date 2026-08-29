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

Follow the post-implementation Claude review loop:
`.agents/agentic-delivery/workflows/claude-review-loop.md`

Choose the automated review route before posting review commands:
`.agents/agentic-delivery/workflows/automated-review-routing-loop.md`

For parent issues, sub-issues, and stacked PRs, follow:
`.agents/agentic-delivery/workflows/stacked-parent-subissue-workflow.md`

For parent issues, sub-issues, and stacked PRs, use the single-worker compatibility ownership
contract:
`.agents/agentic-delivery/contracts/parent-orchestrator-contract.md`

This contract supersedes the legacy parent-orchestration entry point. The canonical worker owns
parent state inline and must not spawn or assign another worker or role. Legacy files remain only
for Wave 6 cleanup and are not active instructions for this flow.

Task type: `<task-type-from-task-skill-matrix>`

Connector implementation scope (fill when applicable):
- named connector cohort: `<one or more connector slugs>`
- immutable source-lock ledger and per-connector ownership/path matrix: `<path>`
- changed-path compliance required: `<yes>`
- Foundation Atlas disposition for each shared contract: `<reuse | extension | actual_gap with captain approval>`
- shared foundation/mapping scope in this bounded PR: `<paths and named consumers, or none>`
- unrelated connector work: `<excluded>`

Required skills:
- the installed GSD sequence for implementation or behavior-changing work: `discuss-phase`,
  `plan-phase --tdd`, `execute-phase`, `verify-work` with gap closure, then `code-review`
- `golang-how-to` for Go work, plus task-specific Go skills from `required-skills-routing.md`
- design skills such as `frontend-design`, `web-design-guidelines`, and `vercel-react-best-practices` for website/docs UI work
- `<skill capability or local skill name>`

Canonical worker:
- role: `pm-delivery-worker` or inheriting `pm-connector-worker`
- source: `.agents/agentic-delivery/canonical/delivery-contract.json`
- delegation: none

Parent issue:
- `<parent issue URL or "None">`

Parent job ownership:
- active owner: `<canonical worker or "not applicable">`
- state ledger: `<issue comment, PR body section, file path, or "None">`
- durable handoff: `<issues, branches, PRs, and GSD artifacts>`
- sub-PR integration owner: `<canonical worker | not applicable>`
- parent merge owner: `<captain | not applicable>`
- Automated review coverage route: `<sub_pr | parent_pr_fallback | copilot_backup | blocked | not applicable>`
- Copilot fallback route: `<copilot_backup | human | none | not applicable>`

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
- confirm the installed GSD lifecycle commands resolved through `scripts/gsd sources`, were
  followed through Pi aliases or `scripts/gsd prompt ...`, and any inline/manual fallback was
  recorded
- confirm the GSD plan, TDD ledger, and verification checklist were created or updated before
  production edits
- confirm required Go/design skills from `.agents/agentic-delivery/references/required-skills-routing.md` were loaded and recorded
- for CLI feature work, confirm runtime help, bare namespace behavior, `docs/cli/**`, website docs,
  generated help/manual artifacts, and tests are updated or explicitly marked not applicable
- for connector implementation work, confirm the named cohort, immutable source-lock ledger,
  ownership/path matrix, changed-path compliance, Foundation Atlas disposition, and any in-PR
  shared foundation are recorded before PR review
- commit and push coherent green slices to the active issue/PR branch after local green gates;
  never push to `main`
- observe automatic Claude review after implementation when the PR is non-draft and targets
  `main`; do not post manual review commands unless the documented fallback conditions apply
- confirm Claude actually reviewed the relevant commits, or record the parent PR, Copilot, or
  human fallback route for stacked sub-PRs
- if Claude is rate-limited, skipped, disabled, paused, or unavailable and review coverage is
  blocking progress, request GitHub Copilot review once as backup when it is enabled
- reply to every actionable Claude or Copilot item with accepted, accepted_with_modification,
  declined, deferred, or needs_human
- rerun verification after accepted fixes
- ensure accepted fix commits are Claude-reviewed; wait for automatic incremental review when
  active, and use manual `@claude review` only when automatic review is paused, disabled,
  skipped, rate-limit retry is due, or the automatic pause threshold was reached
- advance sub-PRs to canonical integration only when all automated gates pass and no human gate is
  triggered
- require human approval before merging parent PRs into `main`
```
