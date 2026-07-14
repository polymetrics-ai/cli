# Learning: Stacked PR Remote Review Gap

Date: 2026-07-07

## Incident

PR #48 (`feat/34-cli-surface-metadata` -> `feat/44-github-cli-parity`) was treated as ready after
CI passed, but the agentic workflow did not close the program loop correctly:

- The parent roadmap issue #44 required a parent PR from `feat/44-github-cli-parity` to `main`, but
  no parent PR existed.
- Remote PR-bot review status was interpreted as sufficient review signal even though the stacked
  workflow needed a parent integration target and actionable review disposition.
- One actionable review finding was not fixed before handoff.
- The sub-PR was not merged into the parent branch, and the parent branch was not advanced toward a
  complete main-targeted PR.

## Root Causes

- The workflow allowed agents to treat remote status output as review completion without local
  disposition evidence.
- The stacked workflow described the parent PR shape but did not make parent PR creation an
  executable precondition before sub-issues started.
- The CLI-surface agent did not require the `review_disposition` skill group.
- The sub-PR completion criteria did not distinguish between local review coverage on the stacked
  sub-PR and aggregate local review coverage on the parent branch.

## New Rules

- Remote PR-bot status is optional information, not required approval.
- Agents must record local review coverage and dispositions before marking review complete.
- Stacked sub-issues require a parent PR to `main` before they are executable.
- If the parent branch has no diff, create a deliberate parent seed commit so GitHub can open the
  draft parent PR. Prefer a real roadmap/status scaffold only when it adds signal; otherwise use an
  empty commit to avoid file churn.
- CLI-surface work requires review-disposition capability because it changes shared validation and
  docs contracts.
- Sub-issues are not done merely because CI passed; they are done when reviewed, dispositioned,
  integrated into the parent branch, and represented in the parent PR status.

## Recovery Applied

- Added a regression test for the finding: present `api_surface.json` with `endpoints: []` must still
  fail CLI-surface endpoint references.
- Updated `checkCLISurface` to validate endpoint references whenever `b.Surface != nil`, including
  empty endpoint sets.
- Updated agentic contracts, workflows, and YAML agents to require parent PR creation and explicit
  local review coverage for stacked PRs.
