# TDD Ledger — Issue #155 Chatwoot sensitive/admin/destructive policy

| Cycle | Red test / evidence | Expected failure | Green change | Status |
|---|---|---|---|---|
| 1 | `go test ./internal/connectors/engine -run TestWriteRootRelativeEndpointUsesConfiguredOrigin -count=1` | Root-relative write under a scoped base path dispatches below `/api/v1/accounts/{account_id}` instead of the Chatwoot origin root. | Update write request path resolution for fixed root-relative official paths. | planned |
| 2 | `go test ./cmd/connectorgen -run TestChatwootAPISurfaceOperationLedgerMetrics -count=1` | Admin/sensitive/destructive rows remain blocked operations. | Generate typed write coverage for all non-disallowed/non-duplicate operation rows. | planned |
| 3 | `go run ./cmd/connectorgen validate internal/connectors/defs` | New write commands must resolve to declared write actions and supported schemas. | Validate all connector defs. | planned |

## Notes

- No live Chatwoot credentials are used.
- Destructive commands keep typed confirmation metadata and remain subject to reverse ETL approval flow.
