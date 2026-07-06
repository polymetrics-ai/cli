# Stacked parent/subissue PR workflow

Use this workflow for large work that needs one parent issue, multiple sub-issues, one parent branch,
and sub-PRs that merge into the parent branch before the parent PR goes to `main`.

## Create the roadmap

1. Write or update the parent issue with `.agents/agentic-delivery/contracts/parent-issue-roadmap-template.md`.
2. Map the parent issue to milestones and one or more sub-issues.
3. Keep each sub-issue small enough for one independently testable PR.
4. Add explicit dependencies when one sub-issue blocks another.
5. Add the parent issue and sub-issues to a GitHub Project only when project scope is already
   available. Do not refresh auth scopes without human approval.

## Branch model

- Parent branch: `<type>/<parent-issue>-<slug>`, for example `feat/44-github-cli-parity`.
- Sub-issue branch: `<type>/<sub-issue>-<slug>`, for example `feat/34-cli-surface-metadata`.
- Parent branch starts from `main`.
- Sub-issue branches start from the parent branch.
- Parent PR base is `main`.
- Sub-PR base is the parent branch.

## PR body model

Parent PR body:

```markdown
Closes #<parent-issue>

Integrated sub-issues:
- Closes #<sub-issue>
```

Sub-PR body:

```markdown
Refs #<sub-issue>
Refs #<parent-issue>

Parent branch: `<parent-branch>`
Sub-PR base: `<parent-branch>`
```

Do not use closing keywords in sub-PRs that target a parent branch. GitHub only interprets closing
keywords for PRs targeting the default branch.

## Sub-PR execution loop

1. Read the sub-issue and parent issue.
2. Confirm the sub-issue is still in scope and not blocked.
3. Create the sub-issue branch from the parent branch.
4. Follow the issue-to-PR contract and test-first loop.
5. Open the sub-PR against the parent branch with `Refs #<sub-issue>` and `Refs #<parent-issue>`.
6. Run targeted verification and broader issue verification.
7. Run the CodeRabbit review loop and reply to every actionable finding with a disposition. For
   fix commits, wait for automatic incremental review when active and use manual review commands
   only under the conditions in `workflows/coderabbit-review-loop.md`.
8. Merge the sub-PR into the parent branch without human approval only if every automated gate is
   green and no human gate is triggered.
9. Comment on the sub-issue with the merged sub-PR, commit, verification, and parent PR status.
10. Leave the sub-issue open until the parent PR lands on `main`, unless the coordinator explicitly
    decides that parent-branch integration is the definition of done.

## Parent PR execution loop

1. Keep the parent PR draft while sub-issues are still landing into the parent branch.
2. Rebase or merge `main` into the parent branch at planned checkpoints.
3. After all required sub-issues are integrated, run full verification from the parent issue.
4. Request CodeRabbit full review on the parent PR when it is ready for the complete parent pass.
5. Resolve all automated review comments using the CodeRabbit review loop.
6. Mark the parent PR ready for human review.
7. Human approval is required before merging the parent PR into `main`.

## Merge without human approval

Sub-PR auto-merge into a parent branch is allowed only for integration branches and only when the
sub-PR does not cross a human gate. Agents must stop instead of merging when:

- the sub-PR touches auth, secrets, dependencies, destructive external actions, production deploys,
  quality gates, generic write tools, or reverse ETL execution
- CodeRabbit has unresolved actionable comments
- CI is failing or unavailable without a documented infrastructure reason
- the PR changes files outside the sub-issue scope
- the parent branch owner marks the parent issue blocked

The parent PR into `main` always requires human approval.
