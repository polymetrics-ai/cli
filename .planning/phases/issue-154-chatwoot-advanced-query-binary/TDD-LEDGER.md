# TDD Ledger — Issue #154 Chatwoot advanced query / binary engine

| Cycle | Red test / evidence | Expected failure | Green change | Status |
|---|---|---|---|---|
| 1 | `go test ./internal/connectors/engine -run TestDirectReadRootRelativeEndpointUsesConfiguredOrigin -count=1` | Captured red failure: root-relative direct read under a scoped base path dispatched under `/api/v1/accounts/{account_id}` instead of the Chatwoot origin root. | Updated direct-read request path resolution for root-relative official paths. | passed |
| 2 | `go test ./cmd/connectorgen -run TestChatwootAPISurfaceOperationLedgerMetrics -count=1` | Captured red failure after metadata generation until counts were updated: remaining official GET operations moved from `operation.model=direct_read` to covered direct-read commands. | Generated bounded JSON direct-read commands and `covered_by.direct_read` rows. | passed |
| 3 | `go run ./cmd/connectorgen validate internal/connectors/defs` | New command surface had to resolve to declared direct-read rows and supported output policy. | Validated all connector defs. | passed |

## Notes

- Fixture/live credentials are not used.
- Write/admin/sensitive/destructive mutations remain blocked for #155.
