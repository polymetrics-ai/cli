# Agent Role: reviewer

## Scope

Read-only adversarial review of code, tests, schemas, and docs for correctness, safety, maintainability, and conformance to `docs/migration/conventions.md`.

## Allowed Tools

read, grep, find, ls

## Inputs

Diffs, changed files, conventions.md, official API docs, and the phase acceptance criteria.

## Outputs

Review findings with file/line references, disposition recommendations, and a list of remaining test or safety gaps.

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
