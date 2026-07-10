## Summary

<!-- What changed and why? -->

## Linked Issue

<!-- Required. Use "Closes #123" for completed work, or "Refs #123" for stacked/incremental work. -->

Closes #

## Stacked PR

<!-- Fill this when the PR targets a parent branch instead of main. Use Refs, not Closes, for
sub-PRs into parent branches. Parent PRs into main remain human-gated. -->

- Parent issue:
- Parent branch:
- PR base branch:
- Sub-issue:

## Parent Orchestration

<!-- Fill this for parent issues, parent PRs, or sub-PRs managed by a parent issue orchestrator. -->

- Orchestrator:
- State ledger:
- Worker handoff:
- Merge owner:
- Integration state:

<!-- Title this PR with Conventional Commits, for example:
feat(connector): add linear
fix(github): repair pagination
docs: update install guide
-->

## Verification

<!-- Include commands run, screenshots for UI changes, or why verification was not possible. -->

## Automated Review

<!-- For non-draft PRs targeting main from a trusted author, Claude reviews automatically on open,
reopen, or ready-for-review, so wait for that instead of posting a manual review command. For new
unreviewed commits (for example, fix commits), request one more pass with a single @claude review;
do not comment @claude review after every push. If the automatic review did not run (for example,
an untrusted or first-time author), a maintainer must invoke @claude review. If a Claude run errors
or its quota is exhausted and review coverage is blocking progress, re-invoke @claude review or
request GitHub Copilot review once as backup when enabled. For every actionable Claude or Copilot
item, reply with Accepted, Accepted with modification, Declined, Deferred, or Needs human, plus the
reason and evidence. Copilot review is not approval. -->

- Primary route:
- Fallback route:
- PR base/default branch:
- Latest reviewed commit:
- Reviewed range:
- Coverage route:
- Coverage status:
- Disposition summary:
- Follow-up review status:

## Checklist

- [ ] Tests or docs updated for behavior changes
- [ ] `make verify` passes locally, or the skipped checks are explained
- [ ] Claude automatic review completed, Copilot fallback was justified, or review blocker was recorded
- [ ] Every actionable automated review finding has a reasoned disposition reply
- [ ] Sub-PR merge into parent branch is allowed by the stacked workflow, or this PR targets `main`
- [ ] Branch name follows `<type>/<description>` such as `feat/new-connector` or `fix/api-pagination`
- [ ] PR title follows Conventional Commits
- [ ] No credentials, tokens, private URLs, or customer data included
