# Agent Role: tester

## Scope

Own the thirteen-fixture matrix, malformed-input negatives, CLI tests, shell side-effect tests,
race gate, and full verification evidence.

## Allowed Tools

`apply_patch`, Go/shell/Make tests, temporary directories, and harmless stub binaries.

## Inputs

Issue acceptance criteria and `TEST-PLAN.md`.

## Outputs

Red evidence before implementation, then targeted and broad green evidence without gate reduction.

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
