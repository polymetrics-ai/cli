# Xero parity TDD ledger

## Red/validation baseline before production edits

- `go run ./cmd/connectorgen validate internal/connectors/defs` filtered for Xero: passed with 0 findings before edits, but existing Xero ledger had 233 rows with stale `/api.xro/2.0` prefixes and no operation-ledger version.
- Official re-audit from Xero Accounting OpenAPI found 235 operations: `GET=126`, `PUT=53`, `POST=46`, `DELETE=10`.
- Baseline gap: `api_surface.json` covered 85 non-attachment write rows but official reverse-ETL lane has 87; missing `POST /BankTransfers` (`deleteBankTransfers`) and `POST /BankTransfers/{BankTransferID}` (`deleteBankTransfer`).
- Baseline fixture gap: 6 stream fixture directories and 0 write fixtures, below the task requirement for every executable operation.

## Planned tests / assertions

- Add a Xero-specific connectorgen test asserting:
  - 235 official rows;
  - method counts match the OpenAPI audit;
  - 78 ETL reads, 87 reverse writes, 11 direct report reads, and 59 binary/file rows (including the four official `/pdf` GET operations);
  - no legacy `excluded` rows;
  - no stale `/api.xro/2.0` path prefixes;
  - 87 write fixtures and 100 stream fixture directories exist.
- Run Xero conformance to exercise generated stream/write fixtures.
- Run focused CLI tests and generated docs validation after regeneration.

## Green evidence

- `go run ./cmd/connectorgen validate internal/connectors/defs`: 549 connectors checked, 0 findings.
- `go test ./cmd/connectorgen -run 'TestXero' -count=1`: passed.
- `go test ./internal/connectors/conformance -run 'TestConformance/xero' -count=1`: passed.
- `go test ./internal/connectors/engine -run 'TestXeroReportOperationsDirectRead' -count=1`: passed.
- `go test ./internal/cli -run 'Connector|Dynamic|Golden' -count=1`: passed after updating golden transcripts for the Xero command-surface tagline.

## Refactor/safety notes

- Keep generated fixtures synthetic and secret-free.
- Do not loosen validation or connector boundary gates.
- Do not edit shared runtime code to add binary transfer execution.
