# GitHub Operation Ledger

Issue: #37
Branch: `feat/37-github-operation-ledger`
Parent issue: #44

## Goal

Replace GitHub's broad `excluded` API-surface rows with an explicit operation ledger that explains
which endpoints are direct-read candidates, binary reads, sensitive/admin reverse ETL candidates,
destructive actions, duplicates, deprecated rows, local workflows, or disallowed operations.

## Scope

- Preserve all 100 existing `covered_by` stream/write rows.
- Add an opt-in `operation_ledger_version: 1` mode to `api_surface.json`.
- Convert all 403 GitHub `excluded` rows to `operation` metadata.
- Keep every operation row blocked by default.
- Add validator/conformance checks that ledger rows do not enable execution.
- Add metrics enough to answer coverage by method/model/risk/status.

## Operation Models

- `direct_read`
- `binary_read`
- `sensitive_reverse_etl`
- `admin_reverse_etl`
- `destructive_action`
- `local_workflow`
- `duplicate`
- `deprecated`
- `disallowed`

## Safety Rules

- In ledger mode each endpoint must have exactly one of `covered_by` or `operation`.
- Legacy `excluded` is forbidden in ledger mode but remains valid for other connectors.
- Every `operation` row must declare `blocked_by_default: true`.
- Every `operation` row must have a non-empty reason.
- Sensitive/admin/destructive/disallowed rows must include a `source_url` or `notes` explaining the
  decision.
- `duplicate` rows must include `duplicate_of`.

## Red/Green Plan

1. Add red bundle/schema test for `operation_ledger_version` + `operation`.
2. Add red connectorgen validator tests for exactly-one classification and blocked-by-default rules.
3. Add red GitHub metrics test expecting 503 total, 100 covered, 403 operation rows, and zero legacy
   exclusions in ledger mode.
4. Implement schema/types/validator changes.
5. Convert GitHub `api_surface.json` mechanically from `excluded` to `operation`.
6. Validate JSON, connectorgen, and targeted Go tests.

## Human Gates

- Any execution path for operation-ledger rows.
- Any raw API, GraphQL, generic HTTP, generic SQL, or shell execution path.
- Any destructive/admin action dispatch.
- Parent PR merge into `main`.
