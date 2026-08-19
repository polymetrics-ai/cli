---
status: clean
files_reviewed: 2
findings:
  critical: 0
  warning: 0
  info: 0
  total: 0
depth: standard
mode: manual-inline
---

# Code review — Issue 3980

## Scope

- `internal/connectors/database/parquet_delivery_workset.go`
- `internal/connectors/database/parquet_delivery_workset_test.go`

## Review result

Clean after a standard inline review. The review specifically traced the
identity binding back to `ManagedTargetDeliveryLedgerKey`, checked path/SQL
construction for caller-controlled data, verified defensive tombstone copies,
and compared every refusal with its observable no-publish/no-baseline-mutation
test.

Two issues found during review were fixed and covered before this final pass:

1. The workset request had no finite artifact ceiling; it now requires and
   enforces `MaxArtifactBytes` (up to 1 GiB).
2. Zero-byte Parquet is the warehouse's defined empty-table representation, so
   derivation/counting now recognizes it without sending it to DuckDB.

No unaddressed critical, warning, or informational finding remains.
