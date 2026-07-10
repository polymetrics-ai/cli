# TDD Ledger — Issue #154 Chatwoot advanced query / binary engine

| Cycle | Red test / evidence | Expected failure | Green change | Status |
|---|---|---|---|---|
| 1 | `go test ./internal/connectors/engine -run TestDirectReadRootRelativeEndpointUsesConfiguredOrigin -count=1` | Root-relative direct read under a scoped base path dispatches under `/api/v1/accounts/{account_id}` instead of the Chatwoot origin root. | Update direct-read request path resolution for root-relative official paths. | planned |
| 2 | `go test ./cmd/connectorgen -run TestChatwootAPISurfaceOperationLedgerMetrics -count=1` | Remaining official GET operations stay `operation.model=direct_read` instead of covered direct-read commands. | Generate bounded JSON direct-read commands and `covered_by.direct_read` rows. | planned |
| 3 | `go run ./cmd/connectorgen validate internal/connectors/defs` | New command surface must resolve to declared direct-read rows and supported output policy. | Validate all connector defs. | planned |

## Notes

- Fixture/live credentials are not used.
- Write/admin/sensitive/destructive mutations remain blocked for #155.
