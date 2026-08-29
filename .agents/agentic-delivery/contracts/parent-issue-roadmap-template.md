# Parent issue roadmap template

Use this template for epic-sized work that is intentionally split into sub-issues and stacked PRs.

```markdown
## Objective

## Background

## Roadmap shape

- Parent issue:
- Parent branch:
- Parent PR:
- Final target branch:
- Milestones:
- Project:
- Canonical worker: `pm-delivery-worker` or `pm-connector-worker`
- Parent ownership contract: `.agents/agentic-delivery/contracts/parent-orchestrator-contract.md`
- Automated review routing: `.agents/agentic-delivery/workflows/automated-review-routing-loop.md`

## Sub-issues

| Issue | Milestone | Branch | PR base | Intent | Status |
| --- | --- | --- | --- | --- | --- |
| #N | <milestone> | `<type>/<issue>-<slug>` | `<parent-branch>` | <one-slice outcome> | Backlog |

## Parent job state

| Issue | Active owner | Branch | PR | Latest SHA | Verification | Automated review coverage | Canonical state | Blocker |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| #N | `<canonical worker>` | `<branch>` | `<url>` | `<sha>` | Pending | `<sub_pr|parent_pr_fallback|copilot_backup|blocked>` | `<state_machine step ID>` | None |

## Branch and PR policy

- Connector implementation sub-issues must declare exactly one target connector, ownership guard
  evidence, changed-path compliance requirements, and any foundation issue/PR path before the
  canonical worker starts the wave.
- Shared runtime/tooling, schema, generated-index, or unrelated connector work discovered by a connector lane is split to a separate foundation issue/PR before the connector implementation proceeds.
- Parent branch starts from `main`.
- Parent PR targets `main` and is created as soon as the parent branch exists.
- Parent PR stays draft and may use `Refs #<parent-issue>` until all required sub-issues are
  integrated and final verification is ready.
- Parent PR contains `Closes #<parent-issue>` only when it is ready for human approval.
- Sub-issue branches start from the parent branch.
- Sub-PRs target the parent branch and use `Refs #<sub-issue>` and `Refs #<parent-issue>`.
- Sub-PRs do not use closing keywords because they do not target the default branch.
- The final parent PR closes integrated sub-issues when the parent branch lands on `main`.
- If the parent branch has no useful diff yet, create a deliberate parent seed commit so GitHub has
  a parent PR thread, checks surface, and review target.

## Child integration evidence

Do not copy a gate list into the issue. The authoritative integration criteria are
`tracker.integrate_when` in `.agents/agentic-delivery/canonical/delivery-contract.json`; PR-shape
and review-routing mechanics live in
`.agents/agentic-delivery/workflows/stacked-parent-subissue-workflow.md`. Record evidence against
every current criterion in the parent job state before advancing `integrate_sub_pr`.

For a connector implementation child, also record the exact target connector, ownership guard,
changed-path compliance, and separate foundation issue/PR for any shared runtime/tooling or
unrelated connector work. Naming or linking a foundation issue does not authorize those paths in
the connector PR.

## Human gates

- parent PR merge to `main`
- auth scope changes or `gh auth refresh`
- secret handling changes
- new dependencies
- destructive external actions
- production deploys
- quality gate reductions
- generic shell, unrestricted HTTP write, unrestricted SQL write, or unrestricted raw API tooling
- reverse ETL execution outside plan, preview, approval, execute

## Acceptance criteria

## Verification

## Sources
```
