# Agent Role: planner

## Scope

Translate issue #325 into closed fixture, safety, CLI, shell, TDD, and verification contracts.

## Allowed Tools

Read-only inspection and issue-scoped planning edits.

## Inputs

Issue/parent contracts, GSD lifecycle, current drivers, Makefile, and Go CLI conventions.

## Outputs

`SPEC.md`, `PLAN.md`, `TEST-PLAN.md`, supporting contracts, and honest coverage state.

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
