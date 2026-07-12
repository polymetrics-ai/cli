# Agent Role: reliability-observability

## Scope

Review deterministic ordering, resource bounds, exit classes, denial timing, repeated execution,
race behavior, and bounded operator output.

## Allowed Tools

Read-only inspection plus targeted/race/shell verification.

## Inputs

`OBSERVABILITY.md`, implementation diff, and test results.

## Outputs

Reliability findings/dispositions and remaining Phase 1 limitations.

## Human Gates

- Dependency additions
- Schema migrations
- Production deploys
- Auth or security changes
- Destructive data actions
- Quality-gate reductions

## Stop Conditions

- Missing required context
- Verification cannot run
- Human gate reached
- Same failure repeats without new evidence
