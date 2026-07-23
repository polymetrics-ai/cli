# Stacked parent/subissue PR workflow

Use this workflow for large work that needs one parent issue, multiple sub-issues, one parent branch,
and sub-PRs that merge into the parent branch before the parent PR goes to `main`.

When a parent issue has multiple workers or sub-issues, use
`.agents/agentic-delivery/workflows/parent-issue-orchestration-loop.md`. The parent issue
orchestrator owns shared parent artifacts, parent PR state, sub-PR merge decisions, exact-head local
Codex review, and independent Shepherd validation.
Use `.agents/agentic-delivery/workflows/local-codex-review-loop.md` followed by
`.agents/agentic-delivery/workflows/shepherd-validator.md`. Do not request Claude or GitHub Copilot
as required or fallback PM coverage.

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
  - The seed commit exists only to create the reviewable parent PR thread, checks surface, exact-head
    local review target, and final human gate.
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
8. Run exact-head verification, then dispatch fresh-context read-only local Codex review and
   disposition every actionable finding. Any fix creates a new head and requires affected gates and
   fresh-context re-review.
9. Run independent Shepherd validation against the same exact evidence after review is clean.
   Require `PROCEED` before integration.
10. Commit and push green sub-issue slices to the sub-issue branch after local green gates. Never
    push to `main`; stop only when a human gate is triggered.
11. Merge the sub-PR into the parent branch without human approval only if every gate is green,
    exact-head local Codex review is clean, Shepherd passes, and no human gate is triggered.
12. After merging a sub-PR into the parent branch, push the parent branch, update the parent PR's
    integrated-subissue list, and record exact integrated-range verification/review/trajectory
    evidence. Do not assume pre-integration identities describe the resulting parent head.
13. Comment on the sub-issue with the merged sub-PR, commit, verification, local Codex disposition,
    Shepherd evidence, and parent PR status.
14. Leave the sub-issue open until the parent PR lands on `main`, unless the coordinator explicitly
    decides that parent-branch integration is the definition of done.

## Parent PR execution loop

1. Keep the parent PR draft while sub-issues are still landing into the parent branch.
2. Merge `main` into a published parent branch at planned checkpoints without rewriting history.
3. After all required sub-issues are integrated, run full verification from the parent issue.
4. Run fresh-context local Codex review at the exact parent/main identities and disposition every
   finding; re-verify and re-review changed heads.
5. Run independent Shepherd trajectory validation for the final parent range.
6. Mark the parent PR ready for human review only after both gates and CI pass.
7. Human approval is required before merging the parent PR into `main`.

## Merge without human approval

Sub-PR auto-merge into a parent branch is allowed only for integration branches and only when the
sub-PR does not cross a human gate. Agents must stop instead of merging when:

- the sub-PR touches auth, secrets, dependencies, destructive external actions, production deploys,
  quality gates, generic write tools, or reverse ETL execution
- local Codex review has unresolved actionable findings
- exact base/head identities drifted after review
- Shepherd did not return `PROCEED` for the clean review transition
- CI is failing or unavailable without a documented infrastructure reason
- the PR changes files outside the sub-issue scope
- the parent branch owner marks the parent issue blocked

The parent PR into `main` always requires human approval.
