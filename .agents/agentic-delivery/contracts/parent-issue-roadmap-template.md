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
- Orchestrator:
- Orchestration workflow: `.agents/agentic-delivery/workflows/parent-issue-orchestration-loop.md`
- Automated review routing: `.agents/agentic-delivery/workflows/automated-review-routing-loop.md`

## Sub-issues

| Issue | Milestone | Branch | PR base | Intent | Status |
| --- | --- | --- | --- | --- | --- |
| #N | <milestone> | `<type>/<issue>-<slug>` | `<parent-branch>` | <one-slice outcome> | Backlog |

## Orchestration state

| Issue | Worker | Branch | PR | Latest SHA | Verification | Automated review coverage | Merge state | Blocker |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| #N | `<agent>` | `<branch>` | `<url>` | `<sha>` | Pending | `<sub_pr|parent_pr_fallback|copilot_backup|blocked>` | Planned | None |

## Branch and PR policy

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
  a parent PR thread, checks surface, and human integration target.

## Automated sub-PR merge policy

Agents may merge a sub-PR into the parent branch without human approval only when all of these are
true:

- the sub-PR is scoped to exactly one sub-issue
- branch name and PR title checks pass
- PR body references the sub-issue and parent issue
- targeted tests and issue verification pass
- local automated review loop is complete and findings are resolved or dispositioned
- local automated review coverage exists for the sub-issue head or integrated parent-branch batch
- no human gate is triggered
- no requested-changes review is open
- the parent branch is current enough that the sub-PR diff is reviewable
- the parent issue orchestrator records the merge decision

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
