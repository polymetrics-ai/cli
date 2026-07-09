# TDD Ledger: Issue #137 HubSpot Operation Ledger

Date: 2026-07-09

## GSD/TDD mode

- Desired command: `scripts/gsd prompt programming-loop init --phase issue-137-hubspot-operation-ledger --dry-run`.
- Result: blocked because the pinned registry does not contain `programming-loop`.
- Active fallback: manual universal programming loop with `scripts/gsd prompt plan-phase issue-137-hubspot-operation-ledger --skip-research` evidence.

## Red-test plan

Before production edits:

1. Add a failing HubSpot ledger test that expects:
   - `operation_ledger_version: 1`.
   - 3,060 unique official method/path operations.
   - method counts GET 1,038 / POST 1,314 / PUT 169 / PATCH 232 / DELETE 307.
   - no duplicate method/path rows.
   - no legacy `excluded` rows.
2. Add a failing validator/schema test for the app-operation ledger models needed by HubSpot classification (`stream_etl`, `query_etl`, `reverse_etl`, `binary_write` if needed) before adding them to the schema/validator vocabulary.

## Red evidence

Pending.

## Green evidence

Pending.

## Refactor evidence

Pending.

## Safety/TDD notes

- No live credentials.
- No live HubSpot API calls.
- Temporary official spec clone/read is public and credential-free.
- Do not create executable generic write actions for mutation endpoints in this ledger-only slice.
- Every mutation classification must remain blocked by default until a named reverse ETL action with fixed schema and plan → preview → approval → execute exists.
- Every binary classification must remain blocked by default until bounded destination/max-bytes policy exists.
