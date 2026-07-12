# TDD Ledger — #100 Linear operation ledger

| Step | Evidence | Result |
|---|---|---|
| Plan | Phase artifact created before production edits; manual-GSD fallback active because `programming-loop` is unavailable. | done |
| Red | `TestLinearOperationLedgerInventoriesGraphQLOperations` failed before `api_surface.json` exposed ledger v1 coverage/blocked rows. | done |
| Green | `api_surface.json` now inventories 532 rows: all 531 live Linear GraphQL root fields plus the raw-GraphQL escape hatch; all 514 official non-deprecated fields are covered, with raw arbitrary GraphQL and deprecated `integrationSettingsUpdate` blocked. | done |
| Refactor/verify | `go test ./internal/connectors/engine -run TestLinearOperationLedgerInventoriesGraphQLOperations -count=1`, `go run ./cmd/connectorgen validate internal/connectors/defs --json`. | done |

No credentialed Linear checks or live Linear writes were run.
