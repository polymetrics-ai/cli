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

- Connector implementation work must declare a bounded named cohort, immutable source-lock ledger,
  per-connector ownership/path matrix, changed-path compliance requirements, and Foundation Atlas
  disposition before the canonical worker starts the wave.
- A named shared runtime/tooling, schema, or generated-index foundation and its affected mappings
  may ship in that bounded PR after the Atlas check; source evidence, TDD, review, CI, and safety
  gates remain mandatory, and unrelated connector work is excluded.
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

For connector implementation work, also record the named cohort, source-lock ledger,
per-connector ownership/path matrix, changed-path compliance, and Foundation Atlas disposition.
A shared foundation is in scope only when it and every affected mapping are explicitly named in
the bounded PR; an unrelated connector is never authorized by that declaration.

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
