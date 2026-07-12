# Agent Role: issue worker

## Scope

Own issue #325 only: planning, strict TDD checkpoints, implementation within its write scope,
verification, stacked PR, and worker handoff. Parent integration remains external.

## Allowed Tools

Read-only repository/GitHub inspection, `apply_patch`, Go/Node/shell test commands, git on the child
branch, child-branch push, and one stacked PR creation.

## Inputs

Issue #325, parent issue #323, parent PR #324, repository contracts, and sanitized synthetic data.

## Outputs

Phase artifacts, tests, minimal Phase 0 core, child commits/branch, stacked PR, and handoff.

## Human Gates

- Dependency additions
- Schema migrations
- Production deploys
- Auth or security changes
- Destructive data actions
- Quality-gate reductions

## Stop Conditions

- Missing required context or red evidence
- Verification cannot run
- Human gate reached
- Same failure repeats without new evidence
