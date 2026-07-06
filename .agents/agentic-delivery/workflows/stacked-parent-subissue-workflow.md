# Stacked parent/subissue PR workflow

Use this workflow for large work that needs one parent issue, multiple sub-issues, one parent branch,
and sub-PRs that merge into the parent branch before the parent PR goes to `main`.

When a parent issue has multiple workers or sub-issues, use
`.agents/agentic-delivery/workflows/parent-issue-orchestration-loop.md`. The parent issue
orchestrator owns shared parent artifacts, parent PR state, sub-PR merge decisions, and CodeRabbit
coverage routing.

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
- The parent PR is created as soon as the parent branch exists. Keep it draft while sub-issues are
  still landing. The parent PR may use `Refs #<parent-issue>` while incomplete; change it to
  `Closes #<parent-issue>` only when the parent issue acceptance criteria are complete and the PR is
  ready for human approval.
- GitHub PRs compare commits on a head branch against a base branch. If the parent branch has no
  diff yet, create a deliberate parent seed commit before opening the draft parent PR:
  - Prefer a small roadmap/status scaffold commit when the repo needs one.
  - Use an empty commit when the roadmap already exists and any file change would be noise.
  - Example: `git commit --allow-empty -m "chore(agentic): open parent orchestration pr"`.
  - The seed commit exists only to create the reviewable parent PR thread, checks surface, and
    CodeRabbit/GitHub review target.
- A missing parent PR is a workflow blocker for stacked sub-issues. Do not treat a sub-PR as
  complete when there is no parent PR to carry the main-branch review, final verification, and human
  approval gate.

## PR body model

Parent PR body:

```markdown
Refs #<parent-issue>

Integrated sub-issues:
- Closes #<sub-issue>
```

Before requesting final human approval, replace `Refs #<parent-issue>` with `Closes
#<parent-issue>` and ensure all integrated sub-issues are listed with closing keywords.

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
3. Confirm the parent PR exists from the parent branch to `main`. Create a draft parent PR if it is
   missing and no human gate blocks creation. If the parent branch has no diff against `main`, add
   the parent seed commit described above first.
4. Create the sub-issue branch from the parent branch.
5. Follow the issue-to-PR contract and test-first loop.
6. Open the sub-PR against the parent branch with `Refs #<sub-issue>` and `Refs #<parent-issue>`.
7. Run targeted verification and broader issue verification.
8. Run the CodeRabbit review loop and reply to every actionable finding with a disposition. Use
   automatic review whenever the PR is non-draft and targets a reviewed base branch. Use manual
   review commands only under the fallback conditions in `workflows/coderabbit-review-loop.md`.
9. Commit and push green sub-issue slices to the sub-issue branch after local green gates. Never
   push to `main`; stop only when a human gate is triggered.
10. If CodeRabbit skips the sub-PR because the base branch is not `main`, record that skip as a
    review-routing event, not as approval. The sub-PR may be integrated into the parent branch only
    when the parent PR exists and the orchestrator observes CodeRabbit review, or records an
    allowed fallback route, on the parent PR commit range that includes the sub-issue.
11. Merge the sub-PR into the parent branch without human approval only if every automated gate is
    green, CodeRabbit coverage is satisfied through the sub-PR or parent-PR fallback, and no human
    gate is triggered.
12. After merging a sub-PR into the parent branch, push the parent branch and update the parent
    PR's integrated-subissue list. If the parent PR is non-draft and targets `main`, wait for
    automatic CodeRabbit review. If the parent PR is draft or automatic review is skipped, record
    coverage as pending or use the fallback rules in `workflows/coderabbit-review-loop.md`.
13. Comment on the sub-issue with the merged sub-PR, commit, verification, CodeRabbit coverage
    route, and parent PR status.
14. Leave the sub-issue open until the parent PR lands on `main`, unless the coordinator explicitly
    decides that parent-branch integration is the definition of done.

## Parent PR execution loop

1. Keep the parent PR draft while sub-issues are still landing into the parent branch.
2. Rebase or merge `main` into the parent branch at planned checkpoints.
3. After all required sub-issues are integrated, run full verification from the parent issue.
4. Observe automatic CodeRabbit review on the parent PR after each integrated sub-issue batch when
   the parent PR is non-draft and targets `main`. Manual review commands are fallback-only.
5. Resolve all automated review comments using the CodeRabbit review loop.
6. Mark the parent PR ready for human review.
7. Human approval is required before merging the parent PR into `main`.

## Merge without human approval

Sub-PR auto-merge into a parent branch is allowed only for integration branches and only when the
sub-PR does not cross a human gate. Agents must stop instead of merging when:

- the sub-PR touches auth, secrets, dependencies, destructive external actions, production deploys,
  quality gates, generic write tools, or reverse ETL execution
- CodeRabbit has unresolved actionable comments
- CodeRabbit review was skipped for the sub-PR and no parent-PR review fallback has been created
- CI is failing or unavailable without a documented infrastructure reason
- the PR changes files outside the sub-issue scope
- the parent branch owner marks the parent issue blocked

The parent PR into `main` always requires human approval.
