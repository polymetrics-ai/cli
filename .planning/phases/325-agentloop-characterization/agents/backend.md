# Agent Role: backend

## Scope

Implement only the standard-library replay, validation, redaction, safety, and loopctl core after
the strict TDD gate passes.

## Allowed Tools

`apply_patch`, gofmt, targeted Go tests, race tests, and read-only inspection.

## Inputs

Phase contracts and red-confirmed tests.

## Outputs

Small explicit Go types/functions with no external I/O except bounded fixture reads and CLI output.

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
