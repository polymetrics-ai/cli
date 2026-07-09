# Issue roadmap best practices

Accessed: 2026-07-06

## Primary sources

- GitHub Projects overview: https://docs.github.com/en/issues/planning-and-tracking-with-projects/learning-about-projects/about-projects
- GitHub Projects best practices: https://docs.github.com/en/issues/planning-and-tracking-with-projects/learning-about-projects/best-practices-for-projects
- GitHub milestones: https://docs.github.com/en/issues/using-labels-and-milestones-to-track-work/about-milestones
- GitHub sub-issues: https://docs.github.com/en/issues/tracking-your-work-with-issues/using-issues/adding-sub-issues
- GitHub issue dependencies: https://docs.github.com/en/issues/tracking-your-work-with-issues/using-issues/creating-issue-dependencies
- GitHub branch creation for issues: https://docs.github.com/en/issues/tracking-your-work-with-issues/using-issues/creating-a-branch-for-an-issue
- GitHub pull request issue linking: https://docs.github.com/en/issues/tracking-your-work-with-issues/using-issues/linking-a-pull-request-to-an-issue
- GitHub pull request creation: https://docs.github.com/en/pull-requests/collaborating-with-pull-requests/proposing-changes-to-your-work-with-pull-requests/creating-a-pull-request
- GitHub comparing branches in pull requests: https://docs.github.com/en/pull-requests/collaborating-with-pull-requests/proposing-changes-to-your-work-with-pull-requests/about-comparing-branches-in-pull-requests
- GitHub CLI issue create: https://cli.github.com/manual/gh_issue_create
- GitHub CLI issue edit: https://cli.github.com/manual/gh_issue_edit
- GitHub CLI PR create: https://cli.github.com/manual/gh_pr_create
- GitHub REST sub-issues API: https://docs.github.com/en/rest/issues/sub-issues
- Atlassian epics, stories, and initiatives: https://www.atlassian.com/agile/project-management/epics-stories-themes
- Atlassian epics guide: https://www.atlassian.com/agile/project-management/epics
- Atlassian user stories guide: https://www.atlassian.com/agile/project-management/user-stories

## Findings

- Use GitHub Projects for the cross-cutting roadmap view when project scope is available. Projects
  can show issues and pull requests as tables, boards, and roadmaps, with custom fields for status,
  priority, target dates, and phase.
- Use repository milestones for bounded delivery phases. Milestones track groups of issues and pull
  requests, due dates, completion percentage, and open/closed counts.
- Use parent issues for epic-sized deliverables that need one parent branch and one parent PR.
  Create the parent PR as a draft early so checks, discussion, and review automation attach to the
  main-targeted branch from the start. GitHub sub-issues support hierarchy and progress tracking in
  Projects.
- GitHub pull requests compare changes in a head branch against a base branch. If the parent branch
  has no diff yet, create a deliberate parent seed commit before opening the draft parent PR. Prefer
  a meaningful roadmap/status scaffold when useful; otherwise use an empty commit to avoid file
  churn.
- Keep sub-issues small enough for one independently reviewable PR. Atlassian frames epics as
  larger bodies of work broken into stories, with stories sized to one sprint or less by default.
- Use issue dependencies for sequencing that is not parent/child ownership. A sub-issue can be part
  of the same parent epic and still be blocked by another sub-issue.
- Use `Refs #N` for stacked or incremental sub-PRs into a non-default parent branch. GitHub only
  interprets closing keywords for PRs targeting the default branch, so `Closes #N` in a sub-PR to a
  parent branch is misleading.
- Use `Closes #N` only on the final parent PR into the default branch. That final PR should close
  the parent issue and, when appropriate, the sub-issues that have been integrated into the parent
  branch.
- Do not create or modify GitHub Projects from an agent unless the required `project` scope is
  already available. Do not run `gh auth refresh -s project` without human approval.
- If local `gh` does not support current sub-issue flags, use the official REST sub-issues API with
  issue write permission, or record the hierarchy in the parent issue body as a fallback.

## Polymetrics hierarchy

- Initiative: connector CLI migration program.
- Parent issue: provider-level epic, for example GitHub CLI feature parity.
- Milestone: bounded phase such as docs surface, operation ledger/direct reads, GraphQL, or
  sensitive/admin.
- Sub-issue: one implementation slice that can be merged into the parent branch after automated
  gates pass.
- Sub-PR: one issue-backed PR from a sub-issue branch into the parent branch.
- Parent PR: one draft-to-final PR from the parent branch into `main`; it starts as the review and
  integration thread, then becomes the human-approved final merge PR when the parent issue is
  complete.
