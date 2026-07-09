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

Result after test edit: failed because `PUT /api/v1/profile` did not explicitly record the official multipart avatar/profile policy.

```text
PUT /api/v1/profile must record the blocked multipart avatar/profile policy
```

## Green implementation notes

- Added a `notes` field to the `PUT /api/v1/profile` operation row documenting that multipart avatar/profile updates require the later bounded binary/file policy before file-bearing actions are exposed.
- Tightened that row's reason to call out profile/password/avatar mutation risk.
- Re-ran Chatwoot operation-ledger metrics, GitHub/Chatwoot API surface tests, connector definition validation, and Chatwoot conformance.

## Refactor notes

- Reused existing `githubOperation`, `requiresSourceOrNotes`, and `assertStringIntMap` test helpers from the connectorgen package.
- No runtime execution path was added.
