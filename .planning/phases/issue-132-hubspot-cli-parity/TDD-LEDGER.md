# TDD Ledger: Issue #132 HubSpot CLI Feature Parity Parent

Date: 2026-07-10

## Parent planning gate

No production code changes before these parent artifacts were created:

- `PLAN.md`
- `TDD-LEDGER.md`
- `VERIFICATION.md`
- `RUN-STATE.json`
- `ORCHESTRATION-STATE.json`

## GSD/TDD mode

- Desired command: `scripts/gsd prompt programming-loop init --phase issue-132-hubspot-cli-parity --dry-run`
- Result: blocked because the pinned command registry does not contain `programming-loop`.
- Active fallback: manual universal programming loop using `scripts/gsd prompt plan-phase issue-132-hubspot-cli-parity --skip-research` and `scripts/gsd prompt execute-phase issue-132-hubspot-cli-parity --dry-run` prompts as the GSD adapter path.

## Red-test plan for first implementation lane (#134)

Before production edits, add failing tests for:

1. `cli_surface.json` accepts safe `binary` intent only when backed by typed operation metadata.
2. Implemented binary commands without typed operations are rejected.
3. HubSpot CLI surface metadata exists and has no `raw_api` or `direct_write` commands.
4. HubSpot API inventory metrics match the official baseline once the ledger is introduced: 3,060 unique operations; method counts GET 1,038, POST 1,314, PUT 169, PATCH 232, DELETE 307.

## Red evidence

Pending. No production edits yet.

## Green evidence

Pending.

## Refactor evidence

Pending.

## Safety/TDD notes

- Do not use credentials.
- Do not run live connector checks.
- Do not expose generic raw HTTP write or direct-write execution.
- Keep writes as reverse ETL plan → preview → approval → execute.
- If a safe engine shape is missing, add a typed test and implementation or record an issue-linked blocker with exact evidence.
