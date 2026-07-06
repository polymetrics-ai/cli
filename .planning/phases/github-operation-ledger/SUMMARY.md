# GitHub Operation Ledger Summary

Issue: #37
Branch: `feat/37-github-operation-ledger`

## Delivered

- Added `operation_ledger_version: 1` support to `api_surface.json`.
- Added blocked operation metadata for non-executable API surface rows.
- Added semantic validation that operation rows are blocked, explained, mutually exclusive with
  `covered_by`, and source-linked or noted for sensitive/admin/destructive/disallowed decisions.
- Reclassified all 403 legacy GitHub `excluded` rows to `operation` rows while preserving the 100
  existing covered stream/write mappings.
- Added metrics coverage for GitHub endpoint counts by method, operation model, risk, and status.

## GitHub Metrics

- Total tracked REST rows: 503.
- Covered rows: 100.
- Operation-ledger rows: 403.
- Legacy excluded rows: 0.
- Operation models:
  - `direct_read`: 159
  - `binary_read`: 10
  - `sensitive_reverse_etl`: 58
  - `admin_reverse_etl`: 94
  - `destructive_action`: 5
  - `local_workflow`: 0
  - `duplicate`: 67
  - `deprecated`: 1
  - `disallowed`: 9

## Safety

- No newly reclassified row is executable.
- Operation rows require `status: blocked` and `blocked_by_default: true`.
- Direct command routing remains limited to existing stream/write mappings.
- No raw API, GraphQL, generic HTTP, generic SQL, shell, or destructive/admin dispatch was added.
