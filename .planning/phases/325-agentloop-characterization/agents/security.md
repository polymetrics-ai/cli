# Agent Role: security

## Scope

Review fail-closed ordering, no-enable invariant, bounded fixture ingestion, synthetic identity,
secret/path rejection, and stdout/stderr leakage risks.

## Allowed Tools

Read-only inspection, targeted negative tests, and issue-scoped review-fix patches.

## Inputs

`THREAT-MODEL.md`, production diff, and test results.

## Outputs

Security findings/dispositions and no-sensitive-content confirmation.

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
