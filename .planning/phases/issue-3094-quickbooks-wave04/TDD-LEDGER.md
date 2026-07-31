# TDD Ledger — QuickBooks Wave 04

## Red checks planned before production edits

- `go test ./cmd/connectorgen -run TestQuickBooksAPISurfaceOperationLedgerMetrics -count=1` should fail before the ledger rewrite because current `api_surface.json` has no `operation_ledger_version`, only 11 coarse rows, and legacy `excluded` classifiers.

## Green checks expected after implementation

- QuickBooks API-surface metrics test passes with 161 operation-ledger rows and lane counts matching #3094.
- `connectorgen validate internal/connectors/defs/quickbooks` passes.
- QuickBooks conformance remains fixture-only with explicit skip reasons and no certified/live claim.
- Connector-local hook tests continue to pass, including realm ID safety.

## Evidence

- Red result: `go test ./cmd/connectorgen -run TestQuickBooksAPISurfaceOperationLedgerMetrics -count=1` failed as expected before production edits with `operation_ledger_version = 0, want 1`.
- Green result: focused ledger metrics, connector-local hook tests, focused temp-parent `connectorgen validate`, and `go test ./internal/connectors/conformance -run 'TestConformance/quickbooks' -count=1` pass after the ledger/docs/metadata changes.
- Refactor result: reviewer findings were addressed with connector-local fixes (realm ID non-secret numeric config plus secret-fallback redaction, Intuit Basic auth token refresh, explicit safety-cap error/docs, no advertised unsupported global flags, attachment upload reclassified as blocked sensitive file-write metadata). Generated connector manual/skill/website data and golden transcripts were refreshed; full definitions validation, focused CLI tests, `GOFLAGS='-p=1' make verify`, and `git diff --check` pass.
