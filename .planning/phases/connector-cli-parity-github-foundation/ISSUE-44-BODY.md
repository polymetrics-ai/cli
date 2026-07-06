## Objective

Deliver GitHub connector CLI feature parity through a parent issue, parent branch, and sub-issues
that can be implemented as stacked PRs.

## Background

GitHub CLI parity is too large for one implementation PR. GitHub's own planning model supports
sub-issues for breaking large issues into smaller tasks, milestones for phase tracking, and Projects
for roadmap views. Atlassian's agile guidance maps this shape to an epic split into stories that are
small enough to deliver incrementally.

This issue is the parent roadmap for the GitHub provider pilot. It turns the earlier GitHub CLI
research into executable issue-backed work while preserving the safety model from the agentic
delivery system.

## Roadmap shape

- Parent issue: #44
- Parent branch: `feat/44-github-cli-parity`
- Parent PR: draft PR from `feat/44-github-cli-parity` into `main`
- Final target branch: `main`
- Project: optional; do not refresh `project` auth scope without human approval
- Parent workflow: `.agents/agentic-delivery/workflows/stacked-parent-subissue-workflow.md`
- Parent issue template: `.agents/agentic-delivery/contracts/parent-issue-roadmap-template.md`

## Sub-issues

| Issue | Milestone | Branch | PR base | Intent | Status |
| --- | --- | --- | --- | --- | --- |
| #34 | GitHub CLI Surface Docs | `feat/34-cli-surface-metadata` | `feat/44-github-cli-parity` | Promote GitHub `cli_surface` metadata into validated production form. | Backlog |
| #35 | GitHub CLI Surface Docs | `feat/35-github-help-renderer` | `feat/44-github-cli-parity` | Render gh-like GitHub connector help from metadata. | Backlog |
| #36 | GitHub Operation Ledger & Direct Reads | `feat/36-github-stream-runner` | `feat/44-github-cli-parity` | Execute stream-backed commands through generic connector runner. | Backlog |
| #37 | GitHub Operation Ledger & Direct Reads | `feat/37-github-operation-ledger` | `feat/44-github-cli-parity` | Reclassify GitHub REST rows into execution models. | Backlog |
| #38 | GitHub Operation Ledger & Direct Reads | `feat/38-github-direct-read` | `feat/44-github-cli-parity` | Add constrained direct-read execution for safe operations. | Backlog |
| #39 | GitHub GraphQL Projects & Discussions | `feat/39-github-graphql-engine` | `feat/44-github-cli-parity` | Add declarative GraphQL support for fixed operations. | Backlog |
| #40 | GitHub GraphQL Projects & Discussions | `feat/40-github-projects-discussions` | `feat/44-github-cli-parity` | Model Projects and Discussions command groups. | Backlog |
| #41 | GitHub Sensitive/Admin Actions | `feat/41-github-sensitive-admin` | `feat/44-github-cli-parity` | Design sensitive/admin reverse ETL policy. | Backlog |
| #42 | Cross-Connector Rollout | `docs/42-cross-connector-rollout` | `feat/44-github-cli-parity` | Generalize GitHub learnings to other connectors. | Backlog |

## Branch and PR policy

- Parent branch starts from `main`.
- Parent PR targets `main` and includes `Closes #44`.
- Sub-issue branches start from `feat/44-github-cli-parity`.
- Sub-PRs target `feat/44-github-cli-parity`.
- Sub-PRs use `Refs #<sub-issue>` and `Refs #44`, not closing keywords.
- Closing keywords are reserved for the final parent PR into `main`.

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
- the diff stays inside the sub-issue scope

## Human gates

- Parent PR merge to `main`.
- GitHub Project creation or `gh auth refresh -s project`.
- Auth scope changes, secrets, dependencies, destructive external actions, production deploys, or
  quality gate reductions.
- Generic shell, unrestricted HTTP write, unrestricted SQL write, unrestricted raw API tooling, or
  reverse ETL execution outside plan, preview, approval, execute.

## Acceptance criteria

- #44 is the parent issue for GitHub CLI feature parity.
- #34-#42 are linked as sub-issues when the GitHub sub-issues API is available.
- The parent branch and stacked sub-PR policy is documented.
- Every sub-issue names its branch, PR base, primary agent, verification, and human gates.
- Sub-PR merge without human approval is limited to parent-branch integration and only after all
  automated gates pass.
- Parent PR merge to `main` remains human-approved.

## Verification

```bash
find .agents .github/ISSUE_TEMPLATE .github/workflows -type f \( -name '*.yaml' -o -name '*.yml' \) -print0 | xargs -0 ruby -e 'require "yaml"; ARGV.each { |f| YAML.load_file(f) }'
jq empty .planning/phases/connector-cli-parity-github-foundation/GITHUB-CLI-PARITY-ISSUE-HIERARCHY.json
git diff --check
make verify
```

## Safety notes

- Do not refresh GitHub auth scopes from an agent.
- Do not create live GitHub Projects without human approval.
- Do not merge parent PRs into `main` without human approval.
- Do not close sub-issues from sub-PRs that target the parent branch.

## Sources

- `.agents/agentic-delivery/references/issue-roadmap-best-practices.md`
- GitHub sub-issues docs: https://docs.github.com/en/issues/tracking-your-work-with-issues/using-issues/adding-sub-issues
- GitHub Projects best practices: https://docs.github.com/en/issues/planning-and-tracking-with-projects/learning-about-projects/best-practices-for-projects
- GitHub PR issue linking docs: https://docs.github.com/en/issues/tracking-your-work-with-issues/using-issues/linking-a-pull-request-to-an-issue
- Atlassian epics guide: https://www.atlassian.com/agile/project-management/epics

## Agent execution contract

- Contract: `.agents/agentic-delivery/contracts/issue-agent-contract.md`
- Parent workflow: `.agents/agentic-delivery/workflows/stacked-parent-subissue-workflow.md`
- Task type: `parent-roadmap`
- Primary agent: `.agents/agentic-delivery/agents/coordination/stacked-roadmap-coordinator.agent.yaml`
- Required skill groups: `github_planning`, `review_disposition`, `docs_and_ux`
