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

## Sub-issues

| Issue | Milestone | Branch | PR base | Intent | Status |
| --- | --- | --- | --- | --- | --- |
| #N | <milestone> | `<type>/<issue>-<slug>` | `<parent-branch>` | <one-slice outcome> | Backlog |

## Branch and PR policy

- Parent branch starts from `main`.
- Parent PR targets `main` and contains `Closes #<parent-issue>`.
- Sub-issue branches start from the parent branch.
- Sub-PRs target the parent branch and use `Refs #<sub-issue>` and `Refs #<parent-issue>`.
- Sub-PRs do not use closing keywords because they do not target the default branch.
- The final parent PR closes integrated sub-issues when the parent branch lands on `main`.

## Automated sub-PR merge policy

Agents may merge a sub-PR into the parent branch without human approval only when all of these are
true:

- the sub-PR is scoped to exactly one sub-issue
- branch name and PR title checks pass
- PR body references the sub-issue and parent issue
- targeted tests and issue verification pass
- CodeRabbit review loop is complete and comments are resolved
- no human gate is triggered
- no requested-changes review is open
- the parent branch is current enough that the sub-PR diff is reviewable

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
