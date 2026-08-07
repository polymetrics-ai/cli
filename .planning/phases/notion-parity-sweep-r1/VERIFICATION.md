# Notion parity verification

| Gate | Result |
| --- | --- |
| `connectorgen validate` (551 connectors) | 0 findings |
| `TestEveryImplementedCommandPassesRuntimePreflight` | pass |
| `TestNotionAPISurfaceOperationLedger` (count lock) | pass, and verified to fail on drift |
| `connectorgen surface-sync --check` | clean |
| `internal/connectors/conformance` | pass |
| `make certify-timing` | **pass** — 89.1s of 3m30s, 92 real CLI invocations at budget |
| `make docs-check` | pass |
| `make connector-boundary` | pass |
| `make agent-contract-check` | pass |
| `make tidy-check` | pass |
| `go build ./cmd/pm` | pass |

## Help / manual / docs / website parity

- `pm notion` bare namespace renders contextual help and exits 0 (16 command groups).
- `pm notion <command> --help` verified across direct_read, reverse_etl, destructive, and blocked.
- `docs/cli/**` and `docs/connectors/**` regenerated; catalog records `write: true`.
- Website catalog regenerated; diff verified by object.

## Counts against the ledger

| | Value |
| --- | --- |
| Ledger target (carried forward, stale) | 50 |
| Re-derived from the official OpenAPI 3.1.0 | **51** |
| Declared before | 6 |
| Declared after | 51 (54 rows) |
| Blocked with a named dependency | 1 |
| Not executable, source-cited | 5 |
| Blank dispositions | **0** |
