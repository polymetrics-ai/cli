# TDD Ledger: Chatwoot Operation Ledger

## Setup

- Issue: #152
- Parent: #148
- Branch: `feat/152-chatwoot-operation-ledger`
- GSD: `scripts/gsd prompt quick --validate ...` succeeded; programming-loop prompt is unavailable and recorded as manual fallback.

## Red / green ledger

### Red 1 — missing Chatwoot operation-ledger metrics test

Add `cmd/connectorgen/chatwoot_api_surface_test.go` with `TestChatwootAPISurfaceOperationLedgerMetrics`.

Expected checks:

- `operation_ledger_version: 1`
- 144 official operations across 89 paths
- method counts: DELETE 18, GET 62, PATCH 21, POST 41, PUT 2
- 13 covered rows, 131 blocked operation rows, 0 legacy excluded rows
- model counts: direct_read 53, admin_reverse_etl 35, sensitive_reverse_etl 19, destructive_action 19, disallowed 4, duplicate 1
- risk counts: low 5, medium 60, high 61, critical 5
- status counts: blocked 131
- every operation row is blocked, explained, and source-linked/noted when required

Command:

```bash
go test ./cmd/connectorgen -run ChatwootAPISurfaceOperationLedgerMetrics -count=1
```

Expected after test edit: pass if #149 accounting is already complete; fail if a row is unclassified or unsafe.

## Green implementation notes

Pending.

## Refactor notes

Pending.
