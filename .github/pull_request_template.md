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

## Local Automated Review

<!-- Record local reviewer/verifier/security coverage for the exact head or diff range. Remote
PR-bot review is not required by default. For every actionable finding, record Accepted, Accepted
with modification, Declined, Deferred, or Needs human, plus the reason and evidence. -->

- Local route:
- Reviewed commit or range:
- Reviewer role/runtime:
- Coverage status:
- Disposition summary:
- Follow-up review status:

## Checklist

- [ ] Tests or docs updated for behavior changes
- [ ] `make verify` passes locally, or the skipped checks are explained
- [ ] Local automated review coverage recorded, or review blocker was recorded
- [ ] Every actionable local review finding has a reasoned disposition
- [ ] Sub-PR merge into parent branch is allowed by the stacked workflow, or this PR targets `main`
- [ ] Branch name follows `<type>/<description>` such as `feat/new-connector` or `fix/api-pagination`
- [ ] PR title follows Conventional Commits
- [ ] No credentials, tokens, private URLs, or customer data included
