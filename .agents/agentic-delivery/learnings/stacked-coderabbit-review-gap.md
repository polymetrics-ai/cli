# Learning: Stacked PR CodeRabbit Review Gap

Date: 2026-07-07

## Incident

PR #48 (`feat/34-cli-surface-metadata` -> `feat/44-github-cli-parity`) was treated as ready after
CI passed, but the agentic workflow did not close the program loop correctly:

- The parent roadmap issue #44 required a parent PR from `feat/44-github-cli-parity` to `main`, but
  no parent PR existed.
- CodeRabbit posted a skipped-review status because automatic reviews were disabled for a
  non-default base branch.
- A manually requested CodeRabbit full review also produced one actionable review finding, but the
  review finding was not fixed before handoff.
- The sub-PR was not merged into the parent branch, and the parent branch was not advanced toward a
  complete main-targeted PR.

## Root Causes

- The workflow allowed agents to interpret a green CodeRabbit status as review completion without
  requiring review-record inspection.
- The stacked workflow described the parent PR shape but did not make parent PR creation an
  executable precondition before sub-issues started.
- The CLI-surface agent did not require the `review_disposition` skill group, so CodeRabbit triage
  was not part of the task-specific agent contract.
- The sub-PR completion criteria did not distinguish between review coverage on the stacked sub-PR
  and aggregate review coverage on the parent PR to `main`.

## New Rules

- A CodeRabbit skipped-review status is informational, not approval.
- Agents must inspect CodeRabbit comments and review records before marking review complete.
- Stacked sub-issues require a parent PR to `main` before they are executable.
- If CodeRabbit skips a non-`main` sub-PR, the parent PR must carry CodeRabbit review coverage for
  the integrated commit range.
- CLI-surface work requires review-disposition capability because it changes shared validation and
  docs contracts.
- Sub-issues are not done merely because CI passed; they are done when reviewed, dispositioned,
  integrated into the parent branch, and represented in the parent PR status.

## Recovery Applied

- Added a regression test for the CodeRabbit finding: present `api_surface.json` with
  `endpoints: []` must still fail CLI-surface endpoint references.
- Updated `checkCLISurface` to validate endpoint references whenever `b.Surface != nil`, including
  empty endpoint sets.
- Updated agentic contracts, workflows, and YAML agents to require parent PR creation and explicit
  review coverage for stacked PRs.
